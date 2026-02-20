package date

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/domonda/go-types/language"
)

func NewFinder(lang ...language.Code) *Finder {
	return &Finder{LangHint: getLangHint(lang)}
}

type Finder struct {
	LangHint language.Code
}

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
