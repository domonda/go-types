package money

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/nullable"
)

// Amount adds money related methods to float64
type Amount float64

// NullableAmount is a nullable wrapper around Amount,
// where the zero value of the underlying nullable.Type represents NULL.
type NullableAmount = nullable.Type[Amount]

// NullableAmountFrom returns a not null nullable amount from a value
// using the value as the non-null value.
func NullableAmountFrom(value float64) NullableAmount {
	return NullableAmount(nullable.TypeFrom(Amount(value)))
}

// NullableAmountFromPtr returns a nullable amount from a pointer
// using nil as the null value.
func NullableAmountFromPtr(ptr *Amount) NullableAmount {
	return NullableAmount(nullable.TypeFromPtr(ptr))
}

// ParseAmount parses an amount from str accepting only certain decimal digit counts.
// If no acceptedDecimals are passed, then any decimal digit count is accepted
// and infinity and NaN are parsed and returned without error.
// The Amount.Valid method can be used to check for infinity and NaN.
// If acceptedDecimals are passed, then infinity and NaN are errors
// because they have no decimals that could be accepted.
func ParseAmount(str string, acceptedDecimals ...int) (Amount, error) {
	f, _, _, decimals, err := float.ParseDetails(str)
	if err != nil {
		return 0, err
	}
	if len(acceptedDecimals) == 0 {
		return Amount(f), nil
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, errNonFiniteNotAccepted(str, acceptedDecimals)
	}
	if slices.Contains(acceptedDecimals, decimals) {
		return Amount(f), nil
	}
	return 0, fmt.Errorf("parsing %q returned %d decimals which is not in accepted list of %v", str, decimals, acceptedDecimals)
}

// errNonFiniteNotAccepted returns the error of the parse functions for
// a NaN or infinite str parsed with acceptedDecimals.
func errNonFiniteNotAccepted(str string, acceptedDecimals []int) error {
	return fmt.Errorf("parsing %q returned a non-finite value which has no decimals of the accepted list %v", str, acceptedDecimals)
}

// AmountFromPtr dereferences ptr or returns defaultVal if it is nil.
// See also Amount.Ptr.
func AmountFromPtr(ptr *Amount, defaultVal Amount) Amount {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Ptr returns a new pointer to a copy of the amount value.
// See also AmountFromPtr.
func (a Amount) Ptr() *Amount {
	return &a
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// If validate is true, then the Amount.Valid method is checked
// to return an error if the floating point value is infinite or NaN.
func (a *Amount) ScanString(source string, validate bool) error {
	f, err := float.Parse(source)
	if err != nil {
		return err
	}
	if validate && !Amount(f).Valid() {
		return fmt.Errorf("invalid amount: %q", source)
	}
	*a = Amount(f)
	return nil
}

// Cents returns the amount rounded half away from zero to whole cents.
// Like Amount.CentAmount it rounds the shortest decimal representation
// of the float64 value, so Amount(0.145) has 15 cents.
// The result is undefined for infinite and NaN amounts,
// see Amount.CentAmount for a conversion that handles them.
func (a Amount) Cents() int64 {
	if !a.Valid() {
		// Keep the platform dependent conversion of non-finite values
		return int64(math.Round(float64(a) * 100))
	}
	return int64(a.CentAmount(RoundHalfAwayFromZero))
}

// CentAmount returns the amount rounded to whole cents
// with the passed rounding mode as CentAmount.
//
// The rounding is applied to the shortest decimal representation
// of the float64 value instead of multiplying by 100 first, so
// Amount(0.145) rounds to the 15 cents of the decimal literal
// rather than to the 14 cents of its binary representation.
//
// A NaN amount maps to zero and an amount outside of the
// CentAmount range is clamped to MinCentAmount or MaxCentAmount,
// because CentAmount has no non-finite states.
func (a Amount) CentAmount(rounding RoundingMode) CentAmount {
	f := float64(a)
	if math.IsNaN(f) {
		return 0
	}
	if math.IsInf(f, 0) {
		if math.Signbit(f) {
			return MinCentAmount
		}
		return MaxCentAmount
	}
	// FormatFloat with 'f' and precision -1 is the shortest decimal that
	// round-trips to f, written without an exponent, so its digits are the
	// decimal representation to round. Going through a DecimalAmount instead
	// would cap the result at its narrower coefficient range.
	digits := strconv.FormatFloat(math.Abs(f), 'f', -1, 64)
	intDigits, fracDigits, _ := strings.Cut(digits, ".")
	cents, ok := centsFromDecimalDigits(intDigits, fracDigits, math.Signbit(f), rounding)
	if !ok {
		if math.Signbit(f) {
			return MinCentAmount
		}
		return MaxCentAmount
	}
	return cents
}

// DecimalAmount converts the float64 Amount to an exact fixed-point
// DecimalAmount rounded to the given scale with the given rounding mode.
// Like Amount.CentAmount it rounds the shortest decimal representation
// of the float64 value, so Amount(0.145) rounds half away from zero
// to 0.15 and not to the 0.14 of its binary value.
// A NaN or infinite Amount maps to the corresponding non-finite DecimalAmount
// and an out-of-range value maps to ±Inf; it panics only if scale is out of
// range. See DecimalAmountFrom for a conversion that keeps the full float64
// precision instead of rounding to a fixed scale.
func (a Amount) DecimalAmount(scale int, rounding RoundingMode) DecimalAmount {
	return decimalFromFloatRounded(float64(a), 64, scale, rounding)
}

// WithinOneCent returns true if the whole cents of a and b,
// as returned by Amount.Cents, differ by at most one cent.
// Because both amounts are rounded to cents first,
// the difference of the unrounded amounts can be up to
// just under two cents, like for 9.995 and 10.0149.
func (a Amount) WithinOneCent(b Amount) bool {
	diff := a.Cents() - b.Cents()
	return diff >= -1 && diff <= 1
}

// RoundToInt returns the amount rounded to an integer number
func (a Amount) RoundToInt() Amount {
	return Amount(math.Round(float64(a)))
}

// RoundToCents returns the amount rounded half away from zero
// to cents (2 decimal places), see Amount.RoundToDecimals.
func (a Amount) RoundToCents() Amount {
	return a.RoundToDecimals(2)
}

// RoundToDecimals returns the amount rounded half away from zero
// to the passed number of decimal places.
// Negative decimals round to tens, hundreds, and so on.
// Like Amount.CentAmount it rounds the shortest decimal representation
// of the float64 value, so Amount(0.145).RoundToDecimals(2) is 0.15
// and not the 0.14 of its binary value.
// NaN and infinite amounts are returned unchanged.
// See float.RoundToDecimals.
func (a Amount) RoundToDecimals(decimals int) Amount {
	return float.RoundToDecimals(a, decimals)
}

// String returns the amount rounded to two decimal places
// formatted with a dot as decimal separator.
// String implements the fmt.Stringer interface.
func (a Amount) String() string {
	return a.FormatSep(0, '.', 2)

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

// GoString returns the Go source representation of the amount for debugging.
// The float literal is the exact decimal expansion of the float64 value,
// so it always parses back to the identical amount.
// GoString implements the fmt.GoStringer interface.
func (a Amount) GoString() string {
	switch {
	case a.IsNaN():
		return "money.Amount(math.NaN())"
	case a.IsInf():
		if a.Signbit() {
			return "money.Amount(math.Inf(-1))"
		}
		return "money.Amount(math.Inf(1))"
	case a == 0 && a.Signbit():
		// The Go literal -0 is the untyped constant zero, so negative zero
		// needs a constructor to survive the round-trip through Go source.
		return "money.Amount(math.Copysign(0, -1))"
	}
	// 1074 fractional digits are always enough for the exact decimal
	// expansion of a float64 because the smallest subnormal is 2^-1074.
	return "money.Amount(" + strings.TrimRight(
		strings.TrimRight(
			fmt.Sprintf("%.1074f", float64(a)),
			"0", // remove trailing zeros
		),
		".", // remove trailing dot
	) + ")"
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

// FormatSep formats the Amount similar to strconv.FormatFloat with the 'f' format option,
// but with decimalSep as decimal separator instead of a point
// and optional grouping of the integer part.
// Valid values for decimalSep are '.' and ','.
// If thousandsSep is not zero, then the integer part of the number is grouped
// with thousandsSep between every group of 3 digits.
// Valid values for thousandsSep are 0 (no grouping), ',', '.', ' '
// and the apostrophe,
// and thousandsSep must be different from decimalSep.
// The precision argument is the number of decimal places, padded with zeros.
// The amount is rounded to them half away from zero like Amount.RoundToDecimals
// does, so Amount(0.125) and Amount(0.145) format as "0.13" and "0.15"
// with precision 2, see float.Format.
// The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return the amount exactly.
func (a Amount) FormatSep(thousandsSep, decimalSep rune, precision int) string {
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

// SplitEqually divides the amount equally into count amounts
// that are rounded to cents and sum up to the initial amount rounded
// half away from zero to cents.
// Like CentAmount.SplitEqually the cents that can't be divided evenly
// are handed to the last amounts, so no two amounts ever differ by more
// than one cent and none has the opposite sign of the initial amount.
// A count of one returns the amount unrounded.
// A NaN or infinite amount, or one beyond the CentAmount range,
// is divided without rounding. Beyond about ±90 trillion (2^53 cents)
// float64 can't represent every cent anymore, so the parts
// may not sum up to the amount exactly.
func (a Amount) SplitEqually(count int) []Amount {
	if count < 1 {
		return nil
	}
	if count == 1 {
		return []Amount{a} // Don't introduce rounding errors
	}
	result := make([]Amount, count)
	cents, ok := a.splitCents()
	if !ok {
		for i := range result {
			result[i] = a / Amount(count)
		}
		return result
	}
	for i, part := range cents.SplitEqually(count) {
		result[i] = part.Amount()
	}
	return result
}

// SplitProportionally splits an amount proportionally
// to the passed weights into the same number of amounts
// that are rounded to cents and sum up to the initial amount rounded
// half away from zero to cents.
// Only the magnitudes of the weights are used and every result amount
// gets the sign of the split amount, so the passed weights can be
// positive, negative, or zero.
// If all weights are zero then the whole amount ends up
// in the last result element.
// A single weight returns the amount unrounded.
//
// Like CentAmount.SplitProportionally the cents lost to truncation are
// handed to the amounts with the largest truncated fraction (the largest
// remainder method), so every amount is within one cent of its exact share
// up to the float64 precision of the weight ratios.
// A NaN or infinite amount or weight, or an amount beyond the CentAmount
// range, is split without rounding. Beyond about ±90 trillion (2^53 cents)
// float64 can't represent every cent anymore, so the parts
// may not sum up to the amount exactly.
func (a Amount) SplitProportionally(weights []Amount) []Amount {
	count := len(weights)
	if count == 0 {
		return nil
	}
	if count == 1 {
		return []Amount{a} // Don't introduce rounding errors
	}
	cents, ok := a.splitCents()
	maxWeight := 0.0
	for _, weight := range weights {
		ok = ok && weight.Valid()
		maxWeight = max(maxWeight, weight.AbsFloat())
	}
	result := make([]Amount, count)
	if !ok {
		totalWeight := Amount(0)
		for _, weight := range weights {
			totalWeight += weight.Abs()
		}
		for i, weight := range weights {
			result[i] = a * (weight.Abs() / totalWeight)
		}
		return result
	}
	// CentAmount.SplitProportionally apportions the cents exactly for integer
	// weights. Scaling the weight magnitudes so that the largest one is 2^53
	// keeps the full float64 precision of their ratios as integers.
	intWeights := make([]CentAmount, count)
	if maxWeight > 0 {
		for i, weight := range weights {
			intWeights[i] = CentAmount(math.Round(weight.AbsFloat() / maxWeight * (1 << 53)))
		}
	}
	for i, part := range cents.SplitProportionally(intWeights) {
		result[i] = part.Amount()
	}
	return result
}

// splitCents returns the amount rounded half away from zero to cents
// and whether the amount is finite and within the CentAmount range,
// so that its cents can be apportioned exactly by the CentAmount splits.
func (a Amount) splitCents() (CentAmount, bool) {
	cents := a.CentAmount(RoundHalfAwayFromZero)
	return cents, a.Valid() && cents != MinCentAmount && cents != MaxCentAmount
}

// Valid returns if the amount is neither infinite nor NaN
func (a Amount) Valid() bool {
	return !a.IsInf() && !a.IsNaN()
}

// ValidAndPositive returns if the amount is neither infinite nor NaN
// and greater than or equal zero.
func (a Amount) ValidAndPositive() bool {
	return a.Valid() && a >= 0
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

// IsInf returns if the amount is positive or negative infinity
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

	// Decode a JSON string through the standard library so \uXXXX and other
	// escape sequences are resolved, not just the surrounding quotes stripped.
	if len(s) > 0 && s[0] == '"' {
		if err := json.Unmarshal(j, &s); err != nil {
			return fmt.Errorf("can't unmarshal JSON(%s) as money.Amount because of: %w", j, err)
		}
	}

	amount, err := ParseAmount(s)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON(%s) as money.Amount because of: %w", j, err)
	}

	*a = amount
	return nil
}
