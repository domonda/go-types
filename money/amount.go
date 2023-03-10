package money

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/domonda/go-types/float"
)

// Amount adds money related methods to float64
type Amount float64

// ParseAmount parses an amount from str accepting only certain decimal digit counts.
// If no acceptedDecimals are passed, then any decimal digit count is accepted.
// Infinity and NaN are parsed and returned without error.
// The Amount.Valid method can be aused to check for infinity and NaN.
func ParseAmount(str string, acceptedDecimals ...int) (Amount, error) {
	f, _, _, decimals, err := float.ParseDetails(str)
	if err != nil {
		return 0, err
	}
	if len(acceptedDecimals) == 0 || math.IsNaN(f) || math.IsInf(f, 0) {
		return Amount(f), nil
	}
	for _, accepted := range acceptedDecimals {
		if decimals == accepted {
			return Amount(f), nil
		}
	}
	return 0, fmt.Errorf("parsing %q returned %d decimals wich is not in accepted list of %v", str, decimals, acceptedDecimals)
}

// NewAmount returns a pointer to an Amount
// with the passed value.
func NewAmount(value float64) *Amount {
	a := new(Amount)
	*a = Amount(value)
	return a
}

// AmountFromPtr dereferences ptr or returns defaultVal if it is nil
func AmountFromPtr(ptr *Amount, defaultVal Amount) Amount {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (a *Amount) ScanString(source string) (sourceWasNormalized bool, err error) {
	f, err := float.Parse(source)
	if err != nil {
		return false, err
	}
	*a = Amount(f)
	return true, nil
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
	return Amount(math.Round(float64(a)*100) / 100)
}

// RoundToDecimals returns the amount rounded
// to the passed number of decimal places.
func (a Amount) RoundToDecimals(decimals int) Amount {
	pow := math.Pow10(decimals)
	return Amount(math.Round(float64(a)*pow) / pow)
}

// String returns the amount rounded to two decimal places
// formatted with a dot as decimal separator.
// String implements the fmt.Stringer interface.
func (a Amount) String() string {
	return a.RoundToCents().Format(0, '.', 2)

	// neg := a < 0
	// s := strconv.FormatInt(a.Abs().Cents(), 10)

	// l := len(s) + 1
	// if l < 4 {
	// 	l = 4
	// }
	// if neg {
	// 	l++
	// }
	// var b strings.Builder
	// b.Grow(l)

	// if neg {
	// 	b.WriteByte('-')
	// }
	// switch len(s) {
	// case 1:
	// 	b.WriteString("0.0")
	// 	b.WriteString(s)
	// case 2:
	// 	b.WriteString("0.")
	// 	b.WriteString(s)
	// default:
	// 	b.WriteString(s[:len(s)-2])
	// 	b.WriteByte('.')
	// 	b.WriteString(s[len(s)-2:])
	// }

	// return b.String()
}

// GoString returns the amount as string
// in full float64 precision for debugging
func (a Amount) GoString() string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.200f", float64(a)), "0"), ".")
}

// StringOr returns ptr.String() or defaultVal if ptr is nil.
func (ptr *Amount) StringOr(defaultVal string) string {
	if ptr == nil {
		return defaultVal
	}
	return ptr.String()
}

// FloatOr returns the pointed to amount as float64 or defaultVal if ptr is nil.
func (ptr *Amount) FloatOr(defaultVal float64) float64 {
	if ptr == nil {
		return defaultVal
	}
	return float64(*ptr)
}

// AmountOr returns the pointed to amount or defaultVal if ptr is nil.
func (ptr *Amount) AmountOr(defaultVal Amount) Amount {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Format formats the Amount similar to strconv.FormatFloat with the 'f' format option,
// but with decimalSep as decimal separator instead of a point
// and optional grouping of the integer part.
// Valid values for decimalSep are '.' and ','.
// If thousandsSep is not zero, then the integer part of the number is grouped
// with thousandsSep between every group of 3 digits.
// Valid values for thousandsSep are [0, ',', '.']
// and thousandsSep must be different from decimalSep.
// The precision argument controls the number of digits (excluding the exponent).
// Note that the last digit is not rounded!
// The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return f exactly.
func (a Amount) Format(thousandsSep, decimalSep rune, precision int) string {
	return float.Format(float64(a), thousandsSep, decimalSep, precision, true)
}

// BigFloat returns m as a new big.Float
func (a Amount) BigFloat() *big.Float {
	return big.NewFloat(float64(a))
}

// Equal returns if two Amount pointers
// point to Amounts with equal values
// or equal addresses.
func (a *Amount) Equal(b *Amount) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// Signbit reports whether a is negative or negative zero.
func (a Amount) Signbit() bool {
	return math.Signbit(float64(a))
}

// Copysign returns an Amount with the magnitude
// of a and with the sign of the sign argument.
func (a Amount) Copysign(sign Amount) Amount {
	return Amount(math.Copysign(float64(a), float64(sign)))
}

// Abs returns the absolute value of a.
//
// Special cases are:
//
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func (a Amount) Abs() Amount {
	return Amount(math.Abs(float64(a)))
}

// AbsFloat returns the absolute value of a as float64.
//
// Special cases are:
//
//	AbsFloat(±Inf) = +Inf
//	AbsFloat(NaN) = NaN
func (a Amount) AbsFloat() float64 {
	return math.Abs(float64(a))
}

// Invert inverts the sign of the amount.
func (a *Amount) Invert() {
	*a = -*a
}

// Inverted returns the amount with inverted sign.
func (a Amount) Inverted() Amount {
	return -a
}

// WithPosSign returns the amount with a positive sign (abs) if true is passed,
// or with a negative sign if false is passed.
func (a Amount) WithPosSign(positive bool) Amount {
	if positive {
		return a.Copysign(+1)
	} else {
		return a.Copysign(-1)
	}
}

// WithNegSign returns the amount with a negative sign if true is passed,
// or with a positive sign (abs) if false is passed.
func (a Amount) WithNegSign(negative bool) Amount {
	if negative {
		return a.Copysign(-1)
	} else {
		return a.Copysign(+1)
	}
}

// MultipliedByRate returns the amount multiplied by a rate.
func (a Amount) MultipliedByRate(rate Rate) Amount {
	return a * Amount(rate)
}

// DividedByRate returns the amount divided by a rate.
func (a Amount) DividedByRate(rate Rate) Amount {
	return a / Amount(rate)
}

// Percentage returns the amount multiplied by (percent / 100).
func (a Amount) Percentage(percent float64) Amount {
	return a * Amount(percent) / 100
}

// SplitEquallyRoundToCents divides the amount equally into numAmounts amounts
// that are rounded to cents and that sum up to the initial amount rounded to cents.
// The last amount may slightly differ from the others amounts to guarantee
// that the sum of the rounded cents equals the rounded cents of the initial amount.
func (a Amount) SplitEquallyRoundToCents(numAmounts int) []Amount {
	if numAmounts < 1 {
		return nil
	}
	splitted := make([]Amount, numAmounts)
	part := (a / Amount(numAmounts)).RoundToCents()
	for i := 0; i < numAmounts-1; i++ {
		splitted[i] = part
	}
	splitted[numAmounts-1] = (a.RoundToCents() - (part * Amount(numAmounts-1))).RoundToCents()
	return splitted
}

// SplitProportionaly splits an amount proportianly
// to the passed weights into the same number of amounts
// and makes sure that the sum the amounts rounded to cents
// is identical to the amount rounded to cents.
// The passed weights can be positive, negative, or zero.
func (a Amount) SplitProportionaly(weights []Amount) []Amount {
	numWeights := len(weights)
	if numWeights == 0 {
		return nil
	}

	compareSum := Amount(0)
	for _, amount := range weights {
		compareSum += amount.Copysign(a)
	}
	scaleFactor := a / compareSum

	compareSum = 0
	split := make([]Amount, numWeights)
	for i := 0; i < numWeights-1; i++ {
		split[i] = (weights[i].Copysign(a) * scaleFactor).RoundToCents()
		compareSum += split[i]
	}
	split[numWeights-1] = (a.RoundToCents() - compareSum).RoundToCents()

	return split
}

// Valid returns if the amount is neither infinite nor NaN
func (a Amount) Valid() bool {
	return !a.IsInf() && !a.IsNaN()
}

// ValidAndGreaterZero returns if the amount is neither infinite nor NaN
// and greater than zero.
func (a Amount) ValidAndGreaterZero() bool {
	return a.Valid() && a > 0
}

// ValidAndSmallerZero returns if the amount is neither infinite nor NaN
// and smaller than zero.
func (a Amount) ValidAndSmallerZero() bool {
	return a.Valid() && a < 0
}

// IsNaN returns if the amount is not a number (NaN)
func (a Amount) IsNaN() bool {
	return math.IsNaN(float64(a))
}

// IsNaN returns if the amount is positive or negative infinity
func (a Amount) IsInf() bool {
	return math.IsInf(float64(a), 0)
}

// ValidAndHasSign returns if a.Valid() and
// if it has the same sign than the passed int argument or any sign if 0 is passed.
func (a Amount) ValidAndHasSign(sign int) bool {
	return float.ValidAndHasSign(float64(a), sign)
}

// UnmarshalJSON implements encoding/json.Unmarshaler
// and accepts numbers, strings, and null.
// JSON null and "" will set the amout to zero.
func (a *Amount) UnmarshalJSON(j []byte) error {
	s := string(j)

	if s == `null` || s == `""` {
		*a = 0
		return nil
	}

	// Strip quotes
	if l := len(s); l > 2 && s[0] == '"' && s[l-1] == '"' {
		s = s[1 : l-1]
	}

	amount, err := ParseAmount(s)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON(%s) as money.Amount because of: %w", j, err)
	}

	*a = amount
	return nil
}
