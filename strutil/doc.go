// Package strutil provides string helpers used across the go-types module.
//
// The main groups are:
//   - [TrimSpace], [TrimSpaceBytes], [CutTrimSpace] and the Truncate
//     functions: trimming that also removes zero-width characters,
//     the BOM, other non-printable runes and invalid UTF-8.
//   - [IsRuneFunc] predicates with [RemoveRunes], [KeepRunes] and
//     [SplitAndTrimIndex], which vat.IDFinder and money.AmountFinder
//     use to split text into candidate words.
//   - [TransliterateSpecialCharacters], [SanitizeFileName],
//     [MakeValidFileName] and [NormalizeExt] for ASCII-friendly
//     names and file names.
//   - [ToSnakeCase], [SanitizeLineEndings], [EqualJSON] and small
//     string and pointer helpers.
//   - [ParseDomainName] for domains in URLs, www hosts and email
//     addresses with a fixed list of top level domains.
//   - [HTMLEscapeSpecialRunes], which replaces '&' and special runes
//     like 'ä' or '€' with named HTML entities. It does not escape
//     <, >, " or the apostrophe, so it is not a replacement
//     for [html.EscapeString].
//   - [StringSet], a map based string set with the methods of the
//     Go collections proposal, and [StrMutex], a per-key mutex.
//   - [RandomString] for URL-safe random strings from crypto/rand.
//
// The subpackage [github.com/domonda/go-types/strutil/stringfrom]
// formats pointers and other values as strings with a default for nil.
package strutil
