package vat

import "github.com/domonda/go-types/strutil"

// IDFinder is a package-level value that implements FindAllIndex for locating
// VAT IDs within arbitrary byte slices. It splits the input on whitespace and
// colon runes, then tests every sequence of one to three consecutive words for
// a valid VAT ID, returning the byte-index spans of all matches.
var IDFinder idFinder

type idFinder struct{}

func (idFinder) FindAllIndex(str []byte, n int) (indices [][]int) {
	l := len(str)
	if l < IDMinLength {
		return nil
	}

	wordIndices := strutil.SplitAndTrimIndex(str, isVATIDSplitRune, isVATIDTrimRune)
	// fmt.Println("STRING", string(str), wordIndices)

	for begSpace := range wordIndices {
		for endSpace := begSpace; endSpace < begSpace+3 && endSpace < len(wordIndices); endSpace++ {
			beg := wordIndices[begSpace][0]
			end := wordIndices[endSpace][1]
			// fmt.Println("TEST", str[beg:end])
			if BytesAreVATID(str[beg:end]) {
				indices = append(indices, []int{beg, end})
				break
			}
		}
	}

	return indices
}
