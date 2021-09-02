package money

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/domonda/go-types/float"
)

// Amount adds money related methods to float64
type Amount float64

// ParseAmount parses an amount from str accepting only certain decimal digit counts.
// If no acceptedDecimals are passed, then any decimal digit count is accepted.
func ParseAmount(str string, acceptedDecimals ...int) (Amount, error) {
	f, _, _, decimals, err := float.ParseDetails(str)
	if err != nil {
		return 0, err
	}
	if len(acceptedDecimals) == 0 {
		return Amount(f), nil
	}
	for _, accepted := range acceptedDecimals {
		if decimals == accepted {
			return Amount(f), nil
		}
	}
	return 0, fmt.Errorf("parsing %q returned %d decimals wich is not in accepted list of %v", str, decimals, acceptedDecimals)
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
func (a Amount) Format(thousandsSep, decimalSep byte, precision int) string {
	return float.Format(float64(a), thousandsSep, decimalSep, precision, true)
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
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func (a Amount) Abs() Amount {
	return Amount(math.Abs(float64(a)))
}

// Inverted returns the amount with inverted sign.
func (a Amount) Inverted() Amount {
	return -a
}

// Invert inverts the sign of the amount.
func (a *Amount) Invert() {
	*a = -*a
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

// ScaleAmountsToSumRoundToCents returns the passed amounts equally scaled
// in a way that their sum is equal to the passed sum.
// The scaled amounts are rounded to cents and it's assured,
// that the sum of cents equals the passed sum in cents.
func ScaleAmountsToSumRoundToCents(amounts []Amount, sum Amount) []Amount {
	numAmounts := len(amounts)
	if numAmounts == 0 {
		return nil
	}

	compareSum := Amount(0)
	for _, amount := range amounts {
		compareSum += amount.Copysign(sum)
	}
	scaleFactor := sum / compareSum

	compareSum = 0
	scaled := make([]Amount, numAmounts)
	for i := 0; i < numAmounts-1; i++ {
		scaled[i] = (amounts[i].Copysign(sum) * scaleFactor).RoundToCents()
		compareSum += scaled[i]
	}
	scaled[numAmounts-1] = (sum.RoundToCents() - compareSum).RoundToCents()

	return scaled
}

// Valid returns if a is neither infinite nor NaN
func (a Amount) Valid() bool {
	return float.Valid(float64(a))
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
	return float.ValidAndHasSign(float64(a), sign)
}

// UnmarshalJSON implements encoding/json.Unmarshaler
// and accepts numbers, strings, and null.
// JSON null will set the amout to zero.
func (a *Amount) UnmarshalJSON(j []byte) error {
	if len(j) == 0 {
		return errors.New("can't unmarshal empty JSON")
	}

	if bytes.Equal(j, []byte("null")) {
		*a = 0
		return nil
	}

	s := string(j)
	if len(j) > 2 && j[0] == '"' && j[len(j)-1] == '"' {
		s = s[1 : len(j)-1]
	}

	amount, err := ParseAmount(s)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON(%s) as money.Amount because of: %w", j, err)
	}

	*a = amount
	return nil
}
