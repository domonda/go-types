package types

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/domonda/errors"
	"github.com/guregu/null"

	"github.com/domonda/go-types/strutil"
)

// Number is a 64 bit floating point number
type Number float64

// Round returns the amount rounded to an integer number
func (n Number) Round() Number {
	return Number(math.Round(float64(n)))
}

const (
	intNumberR         = `^\-?\d+$`
	commaNumberR       = `^\-?\d+,\d+$`
	commaPointsNumberR = `^\-?\d{1,3}(?:\.\d{3})*(?:,\d+)?$`
	pointNumberR       = `^\-?\d+\.\d+$`
	pointCommasNumberR = `^\-?\d{1,3}(?:,\d{3})*(?:\.\d+)?$`
)

var (
	numberRegex = regexp.MustCompile(
		intNumberR +
			`|` +
			commaNumberR +
			`|` +
			commaPointsNumberR +
			`|` +
			pointNumberR +
			`|` +
			pointCommasNumberR)

	intNumberRegex         = regexp.MustCompile(intNumberR)
	pointNumberRegex       = regexp.MustCompile(pointNumberR)
	pointCommasNumberRegex = regexp.MustCompile(pointCommasNumberR)
	commaNumberRegex       = regexp.MustCompile(commaNumberR)
	commaPointsNumberRegex = regexp.MustCompile(commaPointsNumberR)
)

func isNumberSplitRune(r rune) bool {
	return unicode.IsSpace(r) || r == ':'
}

var isNumberTrimRune = strutil.IsRune('.', ',', ';')

var NumberFinder numberFinder

type numberFinder struct{}

func (numberFinder) FindAllIndex(str []byte, n int) (indices [][]int) {
	for _, pos := range strutil.SplitAndTrimIndex(str, isNumberSplitRune, isNumberTrimRune) {
		if numberRegex.Match(str[pos[0]:pos[1]]) {
			indices = append(indices, pos)
		}
	}
	return indices
}

// StringIsNumber returns if str can be parsed as Number.
// The first given lang argument is used as language hint.
func StringIsNumber(str string) bool {
	return numberRegex.MatchString(str)
}

// ParseNumber tries to parse an Number from str.
// The first given lang argument is used as language hint.
func ParseNumber(str string) (Number, error) {
	switch {
	case intNumberRegex.MatchString(str):
		// fmt.Println("intNumberRegex:", str)
		// no changes needed

	case commaNumberRegex.MatchString(str):
		// fmt.Println("commaNumberRegex:", str)
		str = strings.Replace(str, ",", ".", 1)

	case commaPointsNumberRegex.MatchString(str):
		// fmt.Println("commaPointsNumberRegex:", str)
		str = strings.Replace(str, ".", "", -1)
		str = strings.Replace(str, ",", ".", 1)

	case pointNumberRegex.MatchString(str):
		// fmt.Println("pointNumberRegex:", str)
		// no changes needed

	case pointCommasNumberRegex.MatchString(str):
		// fmt.Println("pointCommasNumberRegex:", str)
		str = strings.Replace(str, ",", "", -1)

	default:
		return 0, errors.New("not an number")
	}

	val, err := strconv.ParseFloat(str, 64)
	return Number(val), err
}

func (n *Number) NullFloat() null.Float {
	if n == nil {
		return null.Float{}
	}
	return null.FloatFrom(float64(*n))
}
