package money

import (
	"regexp"
	"unicode"

	"github.com/domonda/go-types/strutil"
)

// StringIsAmount returns if str is an amount with exactly two decimal places,
// a point or comma as decimal separator and optional thousands grouping
// with the other one, like "1.234,56" or "-1234.56", or one of the
// non-finite values NaN, Inf, +Inf and -Inf.
// If acceptInt is true, then an integer without decimal places is accepted too.
// Other numbers like "12.5" are not detected even though ParseAmount parses them.
func StringIsAmount(str string, acceptInt bool) bool {
	return amountRegex.MatchString(str) || (acceptInt && intAmountRegex.MatchString(str))
}

const (
	intAmountR         = `^\-?\d+$`
	commaAmountR       = `^\-?\d+,\d{2}$`
	commaPointsAmountR = `^\-?\d{1,3}(?:\.\d{3})*(?:,\d{2})$`
	pointAmountR       = `^\-?\d+\.\d{2}$`
	pointCommasAmountR = `^\-?\d{1,3}(?:,\d{3})*(?:\.\d{2})$`
)

var (
	amountRegex = regexp.MustCompile(
		`^(?:NaN|[+-]?Inf)$|` +
			commaAmountR +
			`|` +
			commaPointsAmountR +
			`|` +
			pointAmountR +
			`|` +
			pointCommasAmountR)

	intAmountRegex = regexp.MustCompile(intAmountR)
	// pointAmountRegex       = regexp.MustCompile(pointAmountR)
	// pointCommasAmountRegex = regexp.MustCompile(pointCommasAmountR)
	// commaAmountRegex       = regexp.MustCompile(commaAmountR)
	// commaPointsAmountRegex = regexp.MustCompile(commaPointsAmountR)
)

// const (
// 	intNumberR         = `^\-?\d+$`
// 	commaNumberR       = `^\-?\d+,\d+$`
// 	commaPointsNumberR = `^\-?\d{1,3}(?:\.\d{3})*(?:,\d+)?$`
// 	pointNumberR       = `^\-?\d+\.\d+$`
// 	pointCommasNumberR = `^\-?\d{1,3}(?:,\d{3})*(?:\.\d+)?$`
// )

// var (
// 	numberRegex = regexp.MustCompile(
// 		intNumberR +
// 			`|` +
// 			commaNumberR +
// 			`|` +
// 			commaPointsNumberR +
// 			`|` +
// 			pointNumberR +
// 			`|` +
// 			pointCommasNumberR)

// 	intNumberRegex         = regexp.MustCompile(intNumberR)
// 	pointNumberRegex       = regexp.MustCompile(pointNumberR)
// 	pointCommasNumberRegex = regexp.MustCompile(pointCommasNumberR)
// 	commaNumberRegex       = regexp.MustCompile(commaNumberR)
// 	commaPointsNumberRegex = regexp.MustCompile(commaPointsNumberR)
// )

func isAmountSplitRune(r rune) bool {
	return unicode.IsSpace(r) || r == ':'
}

var isAmountTrimRune = strutil.IsRune('.', ',', ';')

// AmountFinder is a singleton that implements the types.Finder interface
// to locate all valid amount substrings within a byte slice.
var AmountFinder amountFinder

type amountFinder struct{}

// FindAllIndex returns the [start, end) byte index pairs of the amounts in str.
// At most n results are returned; pass n < 0 to return all of them.
func (amountFinder) FindAllIndex(str []byte, n int) (indices [][]int) {
	if n == 0 {
		return nil
	}
	for _, pos := range strutil.SplitAndTrimIndex(str, isAmountSplitRune, isAmountTrimRune) {
		if amountRegex.Match(str[pos[0]:pos[1]]) {
			indices = append(indices, pos)
			if n > 0 && len(indices) == n {
				return indices
			}
		}
	}
	return indices
}
