package money

import (
	"regexp"
	"unicode"

	"github.com/domonda/go-types/strutil"
)

// StringIsAmount returns if str can be parsed as Amount.
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
		`NaN|Inf|\+Inf|-Inf|` +
			commaAmountR +
			`|` +
			commaPointsAmountR +
			`|` +
			pointAmountR +
			`|` +
			pointCommasAmountR)

	intAmountRegex         = regexp.MustCompile(intAmountR)
	pointAmountRegex       = regexp.MustCompile(pointAmountR)
	pointCommasAmountRegex = regexp.MustCompile(pointCommasAmountR)
	commaAmountRegex       = regexp.MustCompile(commaAmountR)
	commaPointsAmountRegex = regexp.MustCompile(commaPointsAmountR)
)

const (
	intNumberR         = `^\-?\d+$`
	commaNumberR       = `^\-?\d+,\d+$`
	commaPointsNumberR = `^\-?\d{1,3}(?:\.\d{3})*(?:,\d+)?$`
	pointNumberR       = `^\-?\d+\.\d+$`
	pointCommasNumberR = `^\-?\d{1,3}(?:,\d{3})*(?:\.\d+)?$`
)

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

var AmountFinder amountFinder

type amountFinder struct{}

func (amountFinder) FindAllIndex(str []byte, n int) (indices [][]int) {
	for _, pos := range strutil.SplitAndTrimIndex(str, isAmountSplitRune, isAmountTrimRune) {
		if amountRegex.Match(str[pos[0]:pos[1]]) {
			indices = append(indices, pos)
		}
	}
	return indices
}
