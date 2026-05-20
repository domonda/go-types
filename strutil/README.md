# strutil

String manipulation utilities used across `go-types`: aggressive whitespace handling (incl. invalid UTF-8 and zero-width chars), rune-predicate splitting/filtering, transliteration, file-name sanitization, snake-case conversion, string sets, per-key mutexes, HTML escaping, domain-name extraction, and crypto-random strings.

```
import "github.com/domonda/go-types/strutil"
```

## Trim & truncate

The package's `TrimSpace` is more aggressive than `strings.TrimSpace`: it removes any Unicode whitespace, any non-printable rune (control / format / surrogate / private-use / unassigned, including zero-width chars like `​‌‍⁠` and the BOM `﻿`), and any byte sequence that doesn't decode as valid UTF-8.

| Function                            | Description                                        |
|-------------------------------------|----------------------------------------------------|
| `TrimSpace[S ~string](s)`           | Generic trim using the predicate above.            |
| `TrimSpaceBytes(b []byte)`          | Same for `[]byte`.                                 |
| `CutTrimSpace(s, sep)`              | `strings.Cut`-style split with both halves trimmed. |
| `Truncate[S ~string](s, maxRunes)`  | UTF-8-aware truncation.                            |
| `TruncateTrimSpace(s, maxRunes)`    | Truncate then trim.                                |
| `TruncateWithEllipsis(s, maxRunes)` | Append `…` if truncated; total length stays within bound. |
| `IsSpace(r rune)`                   | Unicode space OR zero-width space `​`.              |

## Rune predicates

```go
type IsRuneFunc func(rune) bool

IsRune(runes...)             // any of the given runes
IsRuneInverse(f)             // NOT f
IsRuneAll(f1, f2, ...)       // AND
IsRuneAny(f1, f2, ...)       // OR
IsWordSeparator(r)           // whitespace + punctuation + symbols
IsNorLetterOrDigit(r)        // !letter && !digit
```

| Helper                                  | Description                                        |
|-----------------------------------------|----------------------------------------------------|
| `RemoveRunes(str, removeRunes...)`      | Filter out matching runes; returns `[]byte`.       |
| `KeepRunes(str, keepRunes...)`          | Keep only matching runes; returns `[]byte`.        |
| `RemoveRunesString` / `KeepRunesString` | Same, returning `string`.                          |
| `MapRuneIsAfterWordSeparator(str)`      | `[]bool` aligned with runes — true after a separator. |

## Splitting

```go
strutil.SplitAndTrimIndex(str, isSplit, isTrim)
// Returns [][]int word-boundary indices, splitting on isSplit
// and trimming with isTrim. Powers the date / VAT-ID finders.
```

## String helpers

`Ptr(s)`, `DerefPtr(p)`, `EmptyStringToNil(s)`, `StringToPtrEmptyToNil(s)`, `IndexInStrings(s, slice)`, `EqualPtrOrString(a, b)`, generic `Join[T ~string](elems, sep)`, `CompareStringsShorterFirst[T ~string](a, b)`, `ConvertSlice[T, S ~string](s)`.

## Transliteration & sanitization

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `TransliterateSpecialCharacters(s)`                | Replace umlauts/diacritics with ASCII equivalents. |
| `TransliterateSpecialCharactersMaxLen(s, n)`       | Same, bounded.                                     |
| `MakeValidFileName(name)` / `SanitizeFileName(n)`  | Replace forbidden filename characters.             |
| `NormalizeExt(ext)`                                | Canonicalize a file extension.                     |
| `SanitizeLineEndings(s)` / `...Bytes(b)`           | Normalize CRLF / LF / CR.                          |
| `ToSnakeCase(s)`                                   | `MyHTTPRequest` → `my_http_request`.               |
| `StringContainsAny(s, subs)` / `SubStringIn(sub, strs)` | Membership probes.                                 |
| `EqualJSON(a, b any)`                              | Compare two values for JSON-equivalent content.    |

## DomainName

```go
strutil.ParseDomainName("https://example.com/x") // "example.com"
strutil.ParseDomainName("user@example.de")       // "example.de"
strutil.ParseDomainName("www.example.org/y")     // "www.example.org"
```

Recognizes HTTP(S) URLs, `www.`-prefixed hosts, and email addresses, but only matches a curated list of TLDs (`com net org edu gov mil de at eu ch es nl uk cz dk fi`). `ParseDomainNameIndex` also returns the match indices.

## HTML

```go
strutil.HTMLEscapeSpecialRunes(str)
// 'ä' → "&auml;", '&' → "&amp;", etc.
```

Backed by an extensive rune-to-entity map (Greek letters, math symbols, currency signs, accented Latin, ...). Best for content destined for very old browsers; modern HTML can usually rely on UTF-8.

## StringSet

```go
s := strutil.NewStringSet("a", "b", "c")
s.AddSlice([]string{"d", "e"})
s.Sorted()                 // []string sorted asc
s.String()                 // `["a", "b", "c", "d", "e"]`

strutil.NewStringSetMergeSlices(s1, s2, s3) // union from multiple slices
```

## StrMutex

Per-key mutex pool. Useful for serializing concurrent work on the same logical key without pre-declaring per-key mutexes.

```go
locks := strutil.NewStrMutex()
locks.Lock("user-42")
defer locks.Unlock("user-42")
```

Each key allocates a mutex on first lock and reclaims it once its lock count returns to zero, so the map size scales with concurrent contention, not total key count.

## Random

```go
strutil.RandomString(32)      // URL-safe base64 string of len 32
strutil.RandomStringBytes(32) // same, returned as []byte to save a copy
```

Uses `crypto/rand`; suitable for tokens. Panics if the OS RNG fails (extremely rare) or if `length < 0`.

## Related

- `types.Finder` — interface implemented by `date.Finder`, `vat.IDFinder`, `bank.BICFinder`, etc., on top of `SplitAndTrimIndex`.
- `nullable.TrimmedString` / `notnull.TrimmedString` — use this package's `TrimSpace` predicate.
