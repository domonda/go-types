package date

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/domonda/go-types/language"
)

// NewFinder returns a Finder that uses the first element of lang as a language
// hint when parsing date candidates. The hint is used to disambiguate ambiguous
// dd/mm vs mm/dd orderings (e.g. pass language.EN to prefer US ordering).
func NewFinder(lang ...language.Code) *Finder {
	return &Finder{LangHint: getLangHint(lang)}
}

// Finder scans arbitrary text for substrings that can be parsed as dates.
// Use NewFinder to create one, optionally with a language hint to resolve
// ambiguous day/month orderings.
type Finder struct {
	LangHint language.Code
}

// FindAllIndex returns the byte-index pairs [start, end] of every date found in
// str, trying spans of up to three whitespace-delimited words at a time.
// At most n results are returned; pass n < 0 to return all matches.
// The returned slice is nil when no dates are found.
func (df *Finder) FindAllIndex(str []byte, n int) (indices [][]int) {
	if len(str) < MinLength {
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
			for r, n := utf8.DecodeRune(str[beg:end]); r != utf8.RuneError && isDateTrimRune(r); {
				beg += n
				r, n = utf8.DecodeRune(str[beg:end])
			}
			for r, n := utf8.DecodeLastRune(str[beg:end]); r != utf8.RuneError && isDateTrimRune(r); {
				end -= n
				r, n = utf8.DecodeLastRune(str[beg:end])
			}
			_, _, err := normalizeAndCheckDate(strings.ToLower(s[beg:end]), df.LangHint)
			if err == nil {
				indices = append(indices, []int{beg, end})
				break
			}
		}
	}

	return indices
}
