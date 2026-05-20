package bank

import (
	"regexp"
	"unicode/utf8"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/strutil"
)

const (
	// BICRegex is the regular expression pattern for validating a BIC/SWIFT-Code.
	BICRegex = `^([A-Z]{4})([A-Z]{2})([A-Z2-9][A-NP-Z0-9])(XXX|[A-WY-Z0-9][A-Z0-9]{2})?$`
	// BICMinLength is the minimum length of a valid BIC (8-character short form without branch code).
	BICMinLength = 8
	// BICMaxLength is the maximum length of a valid BIC (11-character full form with branch code).
	BICMaxLength = 11
)

var (
	bicExactRegexp = regexp.MustCompile(BICRegex)
	bicFindRegexp  = regexp.MustCompile(`[A-Z]{4}([A-Z]{2})[A-Z2-9][A-NP-Z0-9](?:XXX|[A-WY-Z0-9][A-Z0-9]{2})?`)
)

// BICFinder finds BIC/SWIFT-Codes within a byte slice.
// It validates each candidate against the BIC regex, a valid country code,
// and requires word-separator characters on both sides of the match.
var BICFinder bicFinder

type bicFinder struct{}

func (bicFinder) FindAllIndex(str []byte, n int) [][]int {
	if n == 0 {
		return nil
	}
	// Pass -1 to match all candidates: n must limit the number of valid
	// BICs returned, not the number of raw regex matches before filtering.
	indices := bicFindRegexp.FindAllSubmatchIndex(str, -1)
	if len(indices) == 0 {
		return nil
	}
	result := make([][]int, 0, len(indices))
	for _, matchIndices := range indices {
		bic := str[matchIndices[0]:matchIndices[1]]
		countryCode := country.Code(str[matchIndices[2]:matchIndices[3]])
		_, isFalse := falseBICs[BIC(bic)]
		if countryCode.Valid() && !isFalse && bicExactRegexp.Match(bic) {
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
			if n > 0 && len(result) == n {
				return result
			}
		}
	}
	return result
}
