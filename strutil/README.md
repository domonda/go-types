# strutil

String helpers used across `go-types`: trimming that also removes zero-width characters and invalid UTF-8, rune-predicate filtering and splitting, transliteration and file-name sanitization, snake case, line endings, domain extraction, HTML entities, a string set, a per-key mutex and random strings. Use it when the standard `strings` package leaves invisible runes in user input or you need one of the small helpers the other packages of this module are built on.

```
import "github.com/domonda/go-types/strutil"
```

The subpackage [`stringfrom`](stringfrom/README.md) formats pointers and other values as strings with a default for nil.

## Trim and truncate

`TrimSpace` removes more than `strings.TrimSpace`: any Unicode whitespace, every non-printable rune (control, format, surrogate, private use and unassigned runes, including the zero-width characters `\u200b`, `\u200c`, `\u200d`, `\u2060` and the BOM `\ufeff`), every byte that is not valid UTF-8, and the replacement character `\ufffd`.

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `TrimSpace[S ~string](s S) S`                      | Remove leading and trailing trimmable runes (see above). ASCII-only input takes a fast path. |
| `TrimSpaceBytes(s []byte) []byte`                  | Same for `[]byte`, returns a subslice of `s`.      |
| `CutTrimSpace[S ~string](s, sep S) (before, after S, found bool)` | `strings.Cut` with both halves trimmed. If `sep` is missing it returns `s` untrimmed, `""`, `false`. |
| `Truncate[S ~string](s S, maxLen int) S`           | First `maxLen` runes of `s`, never splitting a rune. `maxLen <= 0` gives `""`. |
| `TruncateTrimSpace[S ~string](s S, maxLen int) S`  | `Truncate`, then `TrimSpace`.                      |
| `TruncateWithEllipsis[S ~string](s S, maxLenInclEllipsis int) S` | Truncate and append `…` so the result has at most `maxLenInclEllipsis` runes. `<= 0` gives `""`. |
| `IsSpace(r rune) bool`                             | `unicode.IsSpace(r)` or `\u200b`. Narrower than the `TrimSpace` predicate: `\ufeff` and `\u200c` are not spaces. |

## Rune predicates and filters

| Identifier                                         | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `IsRuneFunc`                                       | `func(rune) bool`, the predicate type used by all functions below. |
| `IsRune(runes ...rune) IsRuneFunc`                 | True for any of `runes`. Without arguments always false. |
| `IsRuneInverse(isRune IsRuneFunc) IsRuneFunc`      | Negates `isRune`.                                  |
| `IsRuneAll(isRune ...IsRuneFunc) IsRuneFunc`       | True if all functions are true. Without arguments always true. |
| `IsRuneAny(isRune ...IsRuneFunc) IsRuneFunc`       | True if any function is true. Without arguments always false. |
| `IsWordSeparator(r rune) bool`                     | Unicode space, punctuation, mark or symbol. `'-'`, `'_'`, `'€'` and combining marks are separators, `'ä'` is not. |
| `IsNorLetterOrDigit(r rune) bool`                  | Neither a letter nor a digit.                      |
| `RemoveRunes(str []byte, removeRunes ...IsRuneFunc) []byte` | New slice without the runes where any function is true. |
| `KeepRunes(str []byte, keepRunes ...IsRuneFunc) []byte` | New slice with only the runes where any function is true. |
| `RemoveRunesString(str string, removeRunes ...IsRuneFunc) string` | String variant of `RemoveRunes`. Returns `str` itself if no rune was removed and `str` is valid UTF-8. |
| `KeepRunesString(str string, keepRunes ...IsRuneFunc) string` | String variant of `KeepRunes`.                     |
| `MapRuneIsAfterWordSeparator(str string) []bool`   | One entry per **byte** of `str`: true if the rune at that byte follows an `IsWordSeparator` rune. All bytes of a multi-byte rune share the value, the first rune counts as following a separator. |

The four filter functions pass every invalid UTF-8 byte to the predicates as `utf8.RuneError` and write it as `\ufffd` if it is kept, so the byte and string variants return the same result.

## Splitting

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `SplitAndTrimIndex(str []byte, isSplitRune, isTrimRune IsRuneFunc) [][]int` | Split `str` at runes where `isSplitRune` is true, trim `isTrimRune` runes from both ends of each word, and return the `[start, end)` byte offsets of the words that are not empty after trimming. |

`vat.IDFinder` and `money.AmountFinder` use `SplitAndTrimIndex` to cut text into candidate words. The `bank` and `date` finders scan the text with their own logic.

## Strings and pointers

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `Ptr(str string) *string`                          | Pointer to a copy of `str`.                        |
| `DerefPtr(ptr *string) string`                     | `*ptr`, or `""` for nil.                           |
| `StringToPtrEmptyToNil(str string) *string`        | nil for `""`, else a pointer to a copy of `str`.   |
| `EmptyStringToNil(str string) any`                 | Untyped nil for `""`, else `str`. Useful as SQL argument when `""` should be stored as NULL. |
| `EqualPtrOrString(a, b *string) bool`              | Same pointer, or both non-nil and pointing to equal strings. nil and a pointer to `""` are not equal. |
| `IndexInStrings(str string, slice []string) int`   | Index of the first element equal to `str`, or -1.  |
| `Join[T ~string](elems []T, sep string) string`    | `strings.Join` for any string type.                |
| `CompareStringsShorterFirst[T ~string](a, b T) int` | Compare by byte length first, then lexicographically. Usable with `slices.SortFunc`. |
| `ConvertSlice[T, S ~string](s []S) []T`            | Reinterpret `[]S` as `[]T` through `unsafe.Pointer`. No copy: the result shares its elements with `s`. |
| `StringContainsAny(str string, subStrings []string) bool` | `str` contains any of `subStrings`. An empty element always matches. |
| `SubStringIn(subString string, strs []string) bool` | Any of `strs` contains `subString`.                |
| `EqualJSON(a, b any) bool`                         | `a == b` for comparable values, otherwise equal `json.Marshal` output. `1` and `1.0` are equal, `1` and `"1"` are not. False if marshalling fails, true for two nils. |

## Case and line endings

| Function                                       | Description                                        |
|------------------------------------------------|----------------------------------------------------|
| `ToSnakeCase(s string) string`                 | Lower-case `s`, insert `_` before an upper-case rune that follows a rune that is neither upper case nor replaced, and replace space, punctuation and symbol runes with `_`. A run of capitals stays one word: `"MyHTTPRequest"` gives `"my_httprequest"`, `"getURL"` gives `"get_url"`. |
| `SanitizeLineEndings(text string) string`      | Collapse `"\r\n"`, `"\n\r"` and `"\r"` to `"\n"` in one pass, so `"\r\n\r"` becomes one `"\n"` and `"\r\n\r\n"` two. Returns `text` unchanged if it has no `'\r'`. |
| `SanitizeLineEndingsBytes(text []byte) []byte` | Same for `[]byte`. Returns `text` itself if it has no `'\r'`, otherwise a new slice. |

## Transliteration and file names

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `TransliterateSpecialCharacters(str string) string` | Replace 71 special Latin letters and ligatures with ASCII, like `"ö"` to `"oe"`, `"Æ"` to `"AE"`, `"ß"` to `"ss"`. Drops invalid UTF-8 and `\ufffd`. Other runes pass through, so `"Łódź"` gives `"Lodź"`. |
| `TransliterateSpecialCharactersMaxLen(str string, maxLen int) string` | Same, stopping at `maxLen` output runes. A transliteration that starts below the limit is written completely, so `("Ä", 1)` gives `"Ae"`. Negative `maxLen` means no limit, `0` gives `""`. |
| `SanitizeFileName(name string) string`             | ASCII file name from arbitrary input, see below. Never returns `""`. |
| `MakeValidFileName(name string) string`            | Only trims and replaces `\ / : * " < > \|`, newline, carriage return and tab with `_`; at the start and end of the name they are removed. Keeps spaces and non-ASCII runes. Returns `"_"` for an empty result. |
| `NormalizeExt(ext string) string`                  | Lower-case the extension and map `.jpg` to `.jpeg`, `.tif` to `.tiff`, `"."` to `""`. Other runes are replaced like in `SanitizeFileName`. |

`SanitizeFileName` applies these steps:

1. Trim with `TrimSpace`.
2. URL-unescape with `url.QueryUnescape` if the name contains `%` or `+`, so `+` becomes a space.
3. Split off the extension (`path.Ext`) and normalize it with `NormalizeExt`.
4. Transliterate with the `TransliterateSpecialCharacters` table, keep ASCII letters, digits, `-` and `_`.
5. Replace whitespace with `_` and every other rune with `-`, but not at the start and only once per run of replaced runes.

| Input                   | `SanitizeFileName`      |
|-------------------------|-------------------------|
| `"Mein Foto (1).JPG"`   | `"Mein_Foto_1-.jpeg"`   |
| `"  Über/Größe:*.tif "` | `"Ueber-Groesse-.tiff"` |
| `"a+b.pdf"`             | `"a_b.pdf"`             |
| `"100%25.txt"`          | `"100-.txt"`            |
| `"x.Tar.GZ"`            | `"x-Tar.gz"`            |
| `"..."`                 | `"_"`                   |

## Domain names

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseDomainName(word string) string`              | Lower-cased domain of an `http://` or `https://` URL, a `www.` host, or an email address. `""` if `word` is none of these. |
| `ParseDomainNameIndex(word string) (domain string, indices []int)` | Same, plus the `[start, end)` byte offsets of the domain in `word` itself, not in its lower-cased copy. nil offsets if nothing matched. |

The patterns are anchored, so `word` must be exactly one URL, host or address. Only these top-level domains are recognized: `com net org edu gov mil de at eu ch es nl uk cz dk fi`.

| Input                            | `ParseDomainName`   | Why                             |
|----------------------------------|---------------------|---------------------------------|
| `"HTTPS://Sub.Example.COM/Path"` | `"sub.example.com"` | URL hosts may contain more dots |
| `"www.example.org/y"`            | `"www.example.org"` |                                 |
| `"user@example.de"`              | `"example.de"`      |                                 |
| `"example.com"`                  | `""`                | no scheme, `www.` or `@`        |
| `"https://example.com:8080/"`    | `""`                | ports are not matched           |
| `"www.sub.example.com"`          | `""`                | no more dots after `www.`       |
| `"user@mail.example.com"`        | `""`                | no more dots after `@`          |
| `"https://example.io"`           | `""`                | TLD not in the list             |
| `"see https://example.com"`      | `""`                | not a single word               |

## HTML entities

| Function                                    | Description                                        |
|---------------------------------------------|----------------------------------------------------|
| `HTMLEscapeSpecialRunes(str string) string` | Replace `&` and 238 other runes (accented letters, Greek letters, typographic quotes and dashes, math symbols, arrows, currency signs, `\u00a0`) with named entities, like `ä` to `&auml;` and `€` to `&euro;`. |

The result is not safe for untrusted input: `<`, `>`, `"` and `'` are not escaped. `html.UnescapeString` restores the input, except that `\u2006` (six-per-em space) is written as `&nbsp;` and comes back as `\u00a0`.

## StringSet

`StringSet` is `map[string]struct{}`. Its methods mirror `types.Set[T]` and the abstract set interface of the Go collections proposal ([go.dev/issue/80590](https://go.dev/issue/80590)) and delegate to [`mapset`](../mapset/README.md).

| Identifier                                         | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NewStringSet(strings ...string) StringSet`        | New set with the passed strings.                   |
| `NewStringSetMergeSlices(slices ...[]string) StringSet` | New set with the strings of all slices.            |
| `Len() int`                                        | Number of strings.                                 |
| `IsEmpty() bool`                                   | True for an empty or nil set.                      |
| `Contains(str string) bool`                        | `str` is in the set.                               |
| `ContainsAny(strs ...string) bool`                 | Any of `strs` is in the set. False without arguments. |
| `ContainsAll(seq iter.Seq[string]) bool`           | Every string of `seq` is in the set. True for an empty `seq`. |
| `StringContainsAnyOfSet(str string) bool`          | `str` contains any element as a substring. An element `""` always matches. |
| `All() iter.Seq[string]`                           | Iterate in undefined order.                        |
| `Sorted() []string`                                | Sorted slice, nil for an empty or nil set.         |
| `String() string`                                  | Sorted and quoted, like `["a", "b"]`. `[]` for an empty or nil set. |
| `Equal(other StringSet) bool`                      | Same strings. A nil set equals an empty one.       |
| `Intersects(other StringSet) bool`                 | At least one string in common.                     |
| `Clone() StringSet`                                | Copy of the map, nil for nil.                      |
| `Insert(str string) bool`                          | Add `str`, report whether the set changed.         |
| `InsertAll(seq iter.Seq[string]) bool`             | Add all strings of `seq`, report whether the set changed. |
| `Delete(str string) bool`                          | Remove `str`, report whether the set changed.      |
| `DeleteAll(seq iter.Seq[string]) bool`             | Remove all strings of `seq`, report whether the set changed. |
| `DeleteFunc(del func(string) bool) bool`           | Remove the strings where `del` is true, report whether the set changed. |
| `Clear()`                                          | Remove all strings.                                |
| `Union(other StringSet) StringSet`                 | New set with the strings of both.                  |
| `Intersection(other StringSet) StringSet`          | New set with the strings in both.                  |
| `Difference(other StringSet) StringSet`            | New set with the strings of `set` that are not in `other`. |
| `SymmetricDifference(other StringSet) StringSet`   | New set with the strings in exactly one of the two. This was called `Diff` before. |
| `UnionWith(other StringSet)`                       | In-place `Union`.                                  |
| `IntersectionWith(other StringSet)`                | In-place `Intersection`.                           |
| `DifferenceWith(other StringSet)`                  | In-place `Difference`.                             |
| `SymmetricDifferenceWith(other StringSet)`         | In-place `SymmetricDifference`.                    |

Implements: `fmt.Stringer`. A nil `StringSet` is a valid empty set for every method except `Insert`, `InsertAll`, `UnionWith` and `SymmetricDifferenceWith`, which panic like an assignment to a nil map when they have to add a string. The methods returning a new set never return nil.

## StrMutex

`StrMutex` holds one mutex per string key. A key's mutex is created by the first `Lock` and deleted when no goroutine holds or waits for it, so memory grows with the number of keys in use at the same time, not with all keys ever locked.

| Identifier                | Description                                        |
|---------------------------|----------------------------------------------------|
| `NewStrMutex() *StrMutex` | New mutex. The zero value is usable too.           |
| `Lock(key string)`        | Block until `key` is free, then lock it.           |
| `Unlock(key string)`      | Unlock `key`. Panics with `Unlock called for non locked key: "k"` if `key` is not locked. |

`types.KeyMutex[T]` implements the same algorithm for any comparable key and also has `IsLocked`. `uu.IDMutex` does the same for `uu.ID` keys and adds optional lock callbacks. New code can use `types.KeyMutex[string]`.

## Random strings

| Function                               | Description                                        |
|----------------------------------------|----------------------------------------------------|
| `RandomString(length int) string`      | `length` characters from the URL-safe base64 alphabet (`A-Z a-z 0-9 - _`), 6 random bits each, read from `crypto/rand`. Panics if `length` is negative. |
| `RandomStringBytes(length int) []byte` | Same as `[]byte`, saving the string copy.          |

## Serialization

`StringSet` has no JSON, text or SQL methods, so `encoding/json` treats it as a plain map:

| Go value                 | `json.Marshal`    | `String()`   |
|--------------------------|-------------------|--------------|
| `StringSet(nil)`         | `null`            | `[]`         |
| `StringSet{}`            | `{}`              | `[]`         |
| `NewStringSet("b", "a")` | `{"a":{},"b":{}}` | `["a", "b"]` |

`json.Unmarshal` only accepts the object form and adds its keys to an existing map. A JSON array is an error. There is no `sql.Scanner` or `driver.Valuer`. Use `types.Set[string]` if you need a JSON array with `null` for nil.

## Gotchas

- `IsSpace` and `TrimSpace` disagree: `IsSpace('\ufeff')` is false, but `TrimSpace` removes the BOM. `TrimSpace` also removes a valid `\ufffd`, and `TransliterateSpecialCharacters` drops every `\ufffd`, not only leading and trailing ones.
- `CutTrimSpace(" a ", ":")` returns `" a "`, `""`, `false`. The input is only trimmed when `sep` is found.
- `Truncate` and `TruncateWithEllipsis` count runes, not bytes: `Truncate("Größe", 3)` is `"Grö"`, which is 4 bytes.
- `ToSnakeCase` does not split acronyms (`"HTMLHandler"` gives `"htmlhandler"`) and keeps repeated separators (`"hello  world"` gives `"hello__world"`).
- `TransliterateSpecialCharactersMaxLen` can return one rune more than `maxLen`, and neither transliteration function guarantees ASCII output.
- `SanitizeFileName` URL-unescapes its input, so a literal `+` becomes `_` and `%25` becomes `%` and then `-`. It also renames `.JPG` to `.jpeg`.
- `HTMLEscapeSpecialRunes` is not an XSS filter. Do not chain it with `html.EscapeString` either: running `html.EscapeString` on its output escapes the `&` of every entity again, so `ä` ends up as `&amp;auml;`.
- `ParseDomainName("example.com")` is `""`: a bare domain without scheme, `www.` or `@` is not recognized.
- `MapRuneIsAfterWordSeparator` returns one value per byte, so index it with byte offsets, not rune counts.
- `StringContainsAny(s, []string{""})` and `NewStringSet("").StringContainsAnyOfSet(s)` are always true.
- `var s strutil.StringSet; s.Insert("a")` panics. Create sets with `NewStringSet` or `make`.
- `ConvertSlice` does not copy: the result and the input share their elements.
- `StrMutex` is not reentrant. Locking a key again from the goroutine that holds it deadlocks.

## Examples

Trimming and truncation:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/strutil"
)

func main() {
	fmt.Printf("%q\n", strutil.TrimSpace("\u200b\ufeff Hello \u00a0"))

	before, after, found := strutil.CutTrimSpace(" key : value ", ":")
	fmt.Printf("%q %q %v\n", before, after, found)
	before, after, found = strutil.CutTrimSpace(" no separator ", ":")
	fmt.Printf("%q %q %v\n", before, after, found)

	fmt.Println(strutil.Truncate("Größe", 3))
	fmt.Println(strutil.TruncateWithEllipsis("Hello World", 8))
	// Output:
	// "Hello"
	// "key" "value" true
	// " no separator " "" false
	// Grö
	// Hello W…
}
```

Names, file names and HTML entities:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/strutil"
)

func main() {
	fmt.Println(strutil.ToSnakeCase("MyHTTPRequest"))
	fmt.Println(strutil.ToSnakeCase("getURL"))
	fmt.Println(strutil.TransliterateSpecialCharacters("Größe Łódź"))
	fmt.Println(strutil.TransliterateSpecialCharactersMaxLen("Äpfel", 1))
	fmt.Println(strutil.SanitizeFileName("  Über/Größe:*.tif "))
	fmt.Println(strutil.SanitizeFileName("a+b.pdf"))
	fmt.Println(strutil.MakeValidFileName("Über Größe/2024.txt"))
	fmt.Println(strutil.HTMLEscapeSpecialRunes(`<b>Größe & "Maß"</b>`))
	// Output:
	// my_httprequest
	// get_url
	// Groesse Lodź
	// Ae
	// Ueber-Groesse-.tiff
	// a_b.pdf
	// Über Größe_2024.txt
	// <b>Gr&ouml;&szlig;e &amp; "Ma&szlig;"</b>
}
```

Splitting, domains, sets and the zero `StrMutex`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"unicode"

	"github.com/domonda/go-types/strutil"
)

func main() {
	text := []byte("Total: 1.234,56 EUR.")
	var words []string
	for _, pos := range strutil.SplitAndTrimIndex(text, unicode.IsSpace, strutil.IsRune('.', ',', ':')) {
		words = append(words, string(text[pos[0]:pos[1]]))
	}
	fmt.Printf("%q\n", words)

	domain, indices := strutil.ParseDomainNameIndex("Info@Domonda.COM")
	fmt.Println(domain, indices)
	fmt.Printf("%q\n", strutil.ParseDomainName("example.com"))

	set := strutil.NewStringSet("b", "a")
	fmt.Println(set.Insert("c"), set.Insert("c"))
	fmt.Println(set, set.Difference(strutil.NewStringSet("a")))
	j, _ := json.Marshal(set)
	fmt.Println(string(j))

	var locks strutil.StrMutex // usable without NewStrMutex
	locks.Lock("user-42")
	locks.Unlock("user-42")
	// Output:
	// ["Total" "1.234,56" "EUR"]
	// domonda.com [5 16]
	// ""
	// true false
	// ["a", "b", "c"] ["b", "c"]
	// {"a":{},"b":{},"c":{}}
}
```

## Related

- [`stringfrom`](stringfrom/README.md): format pointers and values as strings with a default for nil.
- [`mapset`](../mapset/README.md): the set functions behind `StringSet`.
- [Root package reference](../docs/reference-types.md): `types.Set[T]` with a JSON array codec, `types.KeyMutex[T]`, and the `types.Finder` interface.
- [`nullable`](../nullable/README.md) and [`notnull`](../notnull/README.md): their `TrimmedString` types trim with `TrimSpace`.
- [`vat`](../vat/README.md) and [`money`](../money/README.md): `IDFinder` and `AmountFinder` are built on `SplitAndTrimIndex`.
- [How to extract identifiers from text](../docs/howto-extract-identifiers-from-text.md)
- [Sets: nil, empty and the collections proposal](../docs/explanation-sets.md)
