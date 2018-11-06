package money

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/domonda/errors"
	"github.com/guregu/null"

	"github.com/domonda/go-types/strutil"
)

// Amount adds money related methods to float64
type Amount float64

const (
	intAmountR         = `^\-?\d+$`
	commaAmountR       = `^\-?\d+,\d{2}$`
	commaPointsAmountR = `^\-?\d{1,3}(?:\.\d{3})*(?:,\d{2})$`
	pointAmountR       = `^\-?\d+\.\d{2}$`
	pointCommasAmountR = `^\-?\d{1,3}(?:,\d{3})*(?:\.\d{2})$`
)

var (
	amountRegex = regexp.MustCompile(
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

// StringIsAmount returns if str can be parsed as Amount.
func StringIsAmount(str string, acceptInt bool) bool {
	return amountRegex.MatchString(str) || (acceptInt && intAmountRegex.MatchString(str))
}

// ParseAmount tries to parse an Amount from str.
func ParseAmount(str string, acceptInt bool) (Amount, error) {
	switch {
	case commaAmountRegex.MatchString(str):
		// fmt.Println("commaAmountRegex:", str)
		str = strings.Replace(str, ",", ".", 1)

	case commaPointsAmountRegex.MatchString(str):
		// fmt.Println("commaPointsAmountRegex:", str)
		str = strings.Replace(str, ".", "", -1)
		str = strings.Replace(str, ",", ".", 1)

	case pointAmountRegex.MatchString(str):
		// fmt.Println("pointAmountRegex:", str)
		// no changes needed

	case pointCommasAmountRegex.MatchString(str):
		// fmt.Println("pointCommasAmountRegex:", str)
		str = strings.Replace(str, ",", "", -1)

	case acceptInt && intAmountRegex.MatchString(str):
		// no changes needed

	default:
		return 0, errors.Errorf("Can't parse as money.Amount: '%s'", str)
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, errors.Errorf("Can't parse as money.Amount: '%s'", str)
	}
	return Amount(val), nil
}

// AssignString implements strfmt.StringAssignable
func (a *Amount) AssignString(str string) error {
	parsed, err := ParseAmount(str, true)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}

// Cents returns the amount rounded to cents
func (a Amount) Cents() int64 {
	return int64(math.Round(float64(a) * 100))
}

// RoundToInt returns the amount rounded to an integer number
func (a Amount) RoundToInt() Amount {
	return Amount(math.Round(float64(a)))
}

// RoundToCents returns the amount rounded to cents
func (a Amount) RoundToCents() Amount {
	s := strconv.FormatFloat(float64(a), 'f', -1, 64)

	p := strings.IndexByte(s, '.')
	if p == -1 {
		return a
	}
	if decim := len(s) - p - 1; decim <= 2 {
		return a
	}

	if s[p+3] < '5' {
		// If third decimal is smaller than 5 cut off rest
		s = s[:p+3]
	} else {
		// If third decimal is equal or larger than 5,
		// increase second decimal by one
		s = s[:p+2] + string(s[p+2]+1)
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return Amount(f)
}

// String returns the amount formatted to two decimal places
func (a Amount) String() string {
	neg := a < 0
	s := strconv.FormatInt(a.Abs().Cents(), 10)

	l := len(s) + 1
	if l < 4 {
		l = 4
	}
	if neg {
		l++
	}
	var b strings.Builder
	b.Grow(l)

	if neg {
		b.WriteByte('-')
	}
	switch len(s) {
	case 1:
		b.WriteString("0.0")
		b.WriteString(s)
	case 2:
		b.WriteString("0.")
		b.WriteString(s)
	default:
		b.WriteString(s[:len(s)-2])
		b.WriteByte('.')
		b.WriteString(s[len(s)-2:])
	}

	return b.String()
}

// StringPrecision formats the amount with decimals precision.
// -1 for decimals returns the minimum number of decimals
func (a Amount) StringPrecision(decimals int) string {
	return strconv.FormatFloat(float64(a), 'f', decimals, 64)
}

// GermanString returns the amount formatted with German comma
func (a Amount) GermanString() string {
	return strings.Replace(a.String(), ".", ",", 1)
}

// GermanStringPrecision formats the amount with decimals precision and a German comma.
// -1 for decimals returns the minimum number of decimals
func (a Amount) GermanStringPrecision(decimals int) string {
	return strings.Replace(a.StringPrecision(decimals), ".", ",", 1)
}

// GermanGroupedString returns the amount formatted
func (a Amount) GermanGroupedString() string {
	s := a.String()
	lenS := len(s)
	numPoints := ((lenS - 1) / 3) - 1
	// fmt.Println(s, numPoints)
	b := make([]byte, lenS+numPoints)
	for is, ib := 0, 0; is < lenS; is++ {
		if is == lenS-3 {
			b[ib] = ','
		} else {
			b[ib] = s[is]
		}
		// fmt.Println(string(b[ib]))
		ib++
		if is < lenS-6 && (lenS-is)%3 == 1 {
			b[ib] = '.'
			// fmt.Println(string(b[ib]))
			ib++
		}
	}
	return string(b)
}

// BigFloat returns m as a new big.Float
func (a Amount) BigFloat() *big.Float {
	return big.NewFloat(float64(a))
}

func (a *Amount) Equal(b *Amount) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func (a *Amount) NullFloat() null.Float {
	if a == nil {
		return null.Float{}
	}
	return null.FloatFrom(float64(*a))
}

// Copysign returns an Amount with the magnitude
// of a and with the sign of the sign argument.
func (a Amount) Copysign(sign Amount) Amount {
	return Amount(math.Copysign(float64(a), float64(sign)))
}

// Abs returns the absolute value of a.
//
// Special cases are:
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func (a Amount) Abs() Amount {
	return Amount(math.Abs(float64(a)))
}

// Valid returns if a is not infinit or NaN
func (a Amount) Valid() bool {
	return !math.IsNaN(float64(a)) && !math.IsInf(float64(a), 0)
}

func (a Amount) ValidAndGreaterZero() bool {
	return a.Valid() && a > 0
}

func (a Amount) ValidAndSmallerZero() bool {
	return a.Valid() && a < 0
}

// ValidAndHasSign returns if a.Valid() and
// if it has the same sign than the passed int argument or any sign if 0 is passed.
func (a Amount) ValidAndHasSign(sign int) bool {
	if !a.Valid() {
		return false
	}
	switch {
	case sign > 0:
		return a > 0
	case sign < 0:
		return a < 0
	}
	return true
}

// AmountFromPtr returns an Amount pointed by the pointer
func AmountFromPtr(a *Amount) Amount {
	if a != nil {
		return *a
	}
	return 0
}
