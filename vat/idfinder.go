package vat

import "github.com/domonda/go-types/strutil"

// IDFinder is a package-level value that implements FindAllIndex for locating
// VAT IDs within arbitrary byte slices. It splits the input on whitespace and
// colon runes, then tests every sequence of one to three consecutive words for
// a valid VAT ID, returning the byte-index spans of all matches.
var IDFinder idFinder

type idFinder struct{}

// FindAllIndex returns the byte-index pairs [start, end) of the VAT IDs in str.
// At most n results are returned; pass n < 0 to return all matches.
func (idFinder) FindAllIndex(str []byte, n int) (indices [][]int) {
	l := len(str)
	if n == 0 || l < IDMinLength {
		return nil
	}

	wordIndices := strutil.SplitAndTrimIndex(str, isVATIDSplitRune, isVATIDTrimRune)

	for begSpace := range wordIndices {
		for endSpace := begSpace; endSpace < begSpace+3 && endSpace < len(wordIndices); endSpace++ {
			beg := wordIndices[begSpace][0]
			end := wordIndices[endSpace][1]
			if BytesAreVATID(str[beg:end]) {
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
