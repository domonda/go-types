package date

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/domonda/go-types/language"
)

// NewFinder returns a Finder that uses the first element of lang as a language
// hint when parsing date candidates. The hint is used to disambiguate ambiguous
// dd/mm vs mm/dd orderings with slashes like Normalize: language.EN reads
// "05/06/2024" month first, every other language day first, and without
// a hint it is read month first. "05.06.2024" with dots or any other
// separator is read day first with every hint.
func NewFinder(lang ...language.Code) *Finder {
	return &Finder{LangHint: getLangHint(lang)}
}

// Finder scans arbitrary text for substrings that can be parsed as dates.
// Use NewFinder to create one, optionally with a language hint to resolve
// ambiguous day/month orderings.
type Finder struct {
	LangHint language.Code
}

// FindAllIndex returns the byte-index pairs [start, end) of every date found in
// str, trying spans of up to three whitespace-delimited words at a time.
// A span is a date if Normalize accepts it, so an ISO 8601 date-time
// like "2023-12-25T10:00:00Z" is found as a whole. Only the Italian and
// Spanish month abbreviations that are common words in other languages,
// like "set" and "ago", need language.IT or language.ES as LangHint,
// so that text like "Game set 10 2020" has no date.
// At most n results are returned; pass n < 0 to return all matches.
// The returned slice is nil when no dates are found.
func (df *Finder) FindAllIndex(str []byte, n int) (indices [][]int) {
	if n == 0 || len(str) < MinLength {
		return nil
	}
	s := string(str)

	// Find all spaces, also treat string bounds as spaces
	spacePos := make([]int, 1, 16)
	spacePos[0] = -1 //#nosec G602 -- index 0 is valid, slice created with length 1 above
	for i, r := range s {
		if unicode.IsSpace(r) {
			spacePos = append(spacePos, i)
		}
	}
	spacePos = append(spacePos, len(str))

	for begSpace := 0; begSpace < len(spacePos)-1; begSpace++ {
		for endSpace := begSpace + 1; endSpace < begSpace+4 && endSpace < len(spacePos); endSpace++ {
			beg := spacePos[begSpace] + 1
			end := spacePos[endSpace]
			for r, size := utf8.DecodeRune(str[beg:end]); r != utf8.RuneError && isDateTrimRune(r); {
				beg += size
				r, size = utf8.DecodeRune(str[beg:end])
			}
			for r, size := utf8.DecodeLastRune(str[beg:end]); r != utf8.RuneError && isDateTrimRune(r); {
				end -= size
				r, size = utf8.DecodeLastRune(str[beg:end])
			}
			// Not lower cased here, normalizeDate does that itself
			// after it found the upper case T of an ISO 8601 date-time,
			// so a candidate is a date exactly if Normalize accepts it.
			_, _, err := normalizeAndCheckDate(s[beg:end], df.LangHint)
			if err == nil && !hasCommonWordMonthName(s[beg:end], df.LangHint) {
				indices = append(indices, []int{beg, end})
				if n > 0 && len(indices) == n {
					return indices
				}
				break
			}
		}
	}

	return indices
}

// hasCommonWordMonthName reports if the date candidate has a month name
// of commonWordMonthNames that is no month name of the language lang.
func hasCommonWordMonthName(candidate string, lang language.Code) bool {
	lang = langHint(lang)
	for _, part := range strings.FieldsFunc(strings.ToLower(candidate), isDateSeparatorRune) {
		if langs, ok := commonWordMonthNames[part]; ok && !slices.Contains(langs, lang) {
			return true
		}
	}
	return false
}
