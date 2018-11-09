package vat

import "github.com/domonda/go-types/strutil"

var IDFinder idFinder

type idFinder struct{}

func (idFinder) FindAllIndex(str []byte, n int) (indices [][]int) {
	l := len(str)
	if l < IDMinLength {
		return nil
	}

	wordIndices := strutil.SplitAndTrimIndex(str, isVATIDSplitRune, isVATIDTrimRune)
	// fmt.Println("STRING", string(str), wordIndices)

	for begSpace := 0; begSpace < len(wordIndices); begSpace++ {
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
