package bank

import (
	"regexp"
	"unicode/utf8"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/strutil"
)

var (
	bicFindRegex  = regexp.MustCompile(`[A-Z]{4}([A-Z]{2})[A-Z2-9][A-NP-Z0-9](?:XXX|[A-WY-Z0-9][A-Z0-9]{2})?`)
	bicExactRegex = regexp.MustCompile(`^([A-Z]{4})([A-Z]{2})([A-Z2-9][A-NP-Z0-9])(XXX|[A-WY-Z0-9][A-Z0-9]{2})?$`)
)

const (
	BICMinLength = 8
	BICMaxLength = 11
)

var BICFinder bicFinder

type bicFinder struct{}

func (bicFinder) FindAllIndex(str []byte, n int) [][]int {
	// fmt.Println(string(str))
	indices := bicFindRegex.FindAllSubmatchIndex(str, n)
	if len(indices) == 0 {
		return nil
	}
	result := make([][]int, 0, len(indices))
	for _, matchIndices := range indices {
		if len(matchIndices) != 2*2 {
			panic(errors.Errorf("Expected 4 match indices but len(matchIndices) = %d", len(matchIndices)))
		}
		// for _, i := range matchIndices {
		// 	if i < 0 || i > len(str) {
		// 		fmt.Println("bicFinder invalid index", i, len(str))
		// 		continue
		// 	}
		// }
		bic := str[matchIndices[0]:matchIndices[1]]
		countryCode := country.Code(str[matchIndices[2]:matchIndices[3]])
		_, isValidCountry := ibanCountryLengthMap[countryCode]
		_, isFalse := falseBICs[BIC(bic)]
		if isValidCountry && !isFalse && bicExactRegex.Match(bic) {
			// BIC must also be surrounded by line bounds,
			// or a separator rune
			if matchIndices[0] > 0 {
				r, _ := utf8.DecodeLastRune(str[:matchIndices[0]])
				if !strutil.IsWordSeparator(r) {
					continue
				}
			}
			if matchIndices[1] < len(str) {
				r, _ := utf8.DecodeRune(str[matchIndices[1]:])
				if !strutil.IsWordSeparator(r) {
					continue
				}
			}

			result = append(result, matchIndices[:2])
		}
	}
	return result
}
