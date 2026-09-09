package money

import (
	"cmp"
	"fmt"
	"math"
	"math/bits"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/domonda/go-types/nullable"
)

// CentAmount adds money related methods to int64
// where the value counts whole cents,
// meaning hundredths of a currency unit.
//
// It is the exact integer counterpart of the float64 based Amount
// for the common case of two decimal places. Every value in the range
// from MinCentAmount to MaxCentAmount is a valid amount,
// there are no NaN or infinite states.
type CentAmount int64

const (
	// MinCentAmount is the smallest valid CentAmount.
	// It is one above math.MinInt64 so that negating a valid amount
	// can never overflow, mirroring how the DecimalAmount coefficient
	// range excludes its own most negative value.
	MinCentAmount CentAmount = math.MinInt64 + 1

	// MaxCentAmount is the largest valid CentAmount.
	MaxCentAmount CentAmount = math.MaxInt64
)

// NullableCentAmount is a nullable wrapper around CentAmount,
// where the zero value of the underlying nullable.Type represents NULL.
type NullableCentAmount = nullable.Type[CentAmount]

// NullableCentAmountFrom returns a not null nullable cent amount from a value
// using the value as the non-null value.
func NullableCentAmountFrom(value int64) NullableCentAmount {
	return NullableCentAmount(nullable.TypeFrom(CentAmount(value)))
}

// NullableCentAmountFromPtr returns a nullable cent amount from a pointer
// using nil as the null value.
func NullableCentAmountFromPtr(ptr *CentAmount) NullableCentAmount {
	return NullableCentAmount(nullable.TypeFromPtr(ptr))
}

// ParseCentAmount parses a decimal amount from str and returns it as whole
// cents, using the passed rounding mode if str has more than two decimal places.
// If no acceptedDecimals are passed, then any decimal digit count up to
// the maximum DecimalAmount scale is accepted.
// Unlike ParseAmount the parsing is exact without a float64 round-trip,
// and NaN or infinity are returned as error because CentAmount
// has no non-finite states.
func ParseCentAmount(str string, rounding RoundingMode, acceptedDecimals ...int) (CentAmount, error) {
	amount, err := ParseDecimalAmount(str, acceptedDecimals...)
	if err != nil {
		return 0, err
	}
	return amount.CentAmount(rounding)
}

// NewCentAmount returns a pointer to a CentAmount
// with the passed value.
func NewCentAmount(value int64) *CentAmount {
	c := new(CentAmount)
	*c = CentAmount(value)
	return c
}

// CentAmountFromPtr dereferences ptr or returns defaultVal if it is nil.
// See also CentAmount.Ptr.
func CentAmountFromPtr(ptr *CentAmount, defaultVal CentAmount) CentAmount {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Ptr returns a new pointer to a copy of the cent amount value.
// See also CentAmountFromPtr.
func (c CentAmount) Ptr() *CentAmount {
	return &c
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type
// rounding half away from zero beyond two decimal places.
//
// The validate argument has no effect because every string
// that parses as a CentAmount is a valid CentAmount.
func (c *CentAmount) ScanString(source string, validate bool) error {
	parsed, err := ParseCentAmount(source, RoundHalfAwayFromZero)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// Cents returns the amount as whole cents.
func (c CentAmount) Cents() int64 {
	return int64(c)
}

// Amount returns the cent amount as float64 based Amount,
// which may lose precision above 2^53 cents.
func (c CentAmount) Amount() Amount {
	return Amount(c) / 100
}

// DecimalAmount returns the cent amount as exact fixed-point
// DecimalAmount with scale 2. A magnitude beyond the representable
// DecimalAmount coefficient range of roughly ±2.88×10^17 cents
// maps to ±Inf like every other DecimalAmount overflow.
func (c CentAmount) DecimalAmount() DecimalAmount {
	if c < minDecimalAmountCoefficient || c > maxDecimalAmountCoefficient {
		return decimalInf(c < 0)
	}
	return packDecimalAmount(int64(c), 2)
}

// WithinOneCent returns true if c and b are equal
// within a one cent tolerance.
func (c CentAmount) WithinOneCent(b CentAmount) bool {
	if c < b {
		c, b = b, c
	}
	// The difference of two int64 can overflow int64 but never uint64,
	// so a plain c-b would wrap and report the most distant amounts as equal.
	return uint64(c)-uint64(b) <= 1
}

// RoundToInt returns the amount rounded to whole currency units,
// so the result is always a multiple of 100 cents.
func (c CentAmount) RoundToInt() CentAmount {
	return CentAmount(divRoundHalfAwayFromZero(int64(c), 100) * 100)
}

// String returns the amount formatted with a dot as decimal separator
// and always exactly two decimal places.
// String implements the fmt.Stringer interface.
func (c CentAmount) String() string {
	return c.FormatSep(0, '.')
}

// GoString returns the Go source representation
// of the cent amount for debugging.
func (c CentAmount) GoString() string {
	return fmt.Sprintf("money.CentAmount(%d)", int64(c))
}

// StringOr returns ptr.String() or defaultVal if ptr is nil.
func (ptr *CentAmount) StringOr(defaultVal string) string {
	if ptr == nil {
		return defaultVal
	}
	return ptr.String()
}

// CentsOr returns the pointed to amount as whole cents or defaultVal if ptr is nil.
func (ptr *CentAmount) CentsOr(defaultVal int64) int64 {
	if ptr == nil {
		return defaultVal
	}
	return int64(*ptr)
}

// CentAmountOr returns the pointed to cent amount or defaultVal if ptr is nil.
func (ptr *CentAmount) CentAmountOr(defaultVal CentAmount) CentAmount {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// FormatSep formats the amount with decimalSep as decimal separator
// and always exactly two decimal places.
// A zero decimalSep defaults to '.'.
// If thousandsSep is not zero, then the integer part of the number is grouped
// with thousandsSep between every group of 3 digits.
//
// It deliberately does not delegate to DecimalAmount.FormatSep, which would
// saturate to Inf beyond its coefficient range and mishandle the sign of
// math.MinInt64 cents.
func (c CentAmount) FormatSep(thousandsSep, decimalSep rune) string {
	if decimalSep == 0 {
		decimalSep = '.'
	}
	cents := abs64(int64(c))
	intPart := strconv.FormatUint(cents/100, 10)
	frac := cents % 100
	var b strings.Builder
	// Sign, decimal separator, two fraction digits and the group separators
	b.Grow(len(intPart) + len(intPart)/3*utf8.RuneLen(thousandsSep) + utf8.RuneLen(decimalSep) + 3)
	if c < 0 {
		b.WriteByte('-')
	}
	writeGrouped(&b, intPart, thousandsSep)
	b.WriteRune(decimalSep)
	b.WriteByte(byte('0' + frac/10))
	b.WriteByte(byte('0' + frac%10))
	return b.String()
}

// Equal returns if two CentAmount pointers
// point to CentAmounts with equal values
// or equal addresses.
func (c *CentAmount) Equal(b *CentAmount) bool {
	if c == b {
		return true
	}
	if c == nil || b == nil {
		return false
	}
	return *c == *b
}

// Sign returns -1 if the amount is negative, +1 if positive and 0 if zero.
func (c CentAmount) Sign() int {
	return signOf(int64(c))
}

// IsZero returns if the amount is zero.
func (c CentAmount) IsZero() bool {
	return c == 0
}

// Signbit reports whether c is negative.
func (c CentAmount) Signbit() bool {
	return c < 0
}

// Copysign returns a CentAmount with the magnitude
// of c and with the sign of the sign argument.
func (c CentAmount) Copysign(sign CentAmount) CentAmount {
	if (c < 0) != (sign < 0) {
		return -c
	}
	return c
}

// Abs returns the absolute value of c.
// A c below MinCentAmount is out of range and stays negative
// because negating math.MinInt64 overflows int64.
func (c CentAmount) Abs() CentAmount {
	if c < 0 {
		return -c
	}
	return c
}

// Invert inverts the sign of the amount.
func (c *CentAmount) Invert() {
	*c = -*c
}

// Inverted returns the amount with inverted sign.
func (c CentAmount) Inverted() CentAmount {
	return -c
}

// WithPosSign returns the amount with a positive sign (abs) if true is passed,
// or with a negative sign if false is passed.
func (c CentAmount) WithPosSign(positive bool) CentAmount {
	if positive {
		return c.Copysign(+1)
	} else {
		return c.Copysign(-1)
	}
}

// WithNegSign returns the amount with a negative sign if true is passed,
// or with a positive sign (abs) if false is passed.
func (c CentAmount) WithNegSign(negative bool) CentAmount {
	if negative {
		return c.Copysign(-1)
	} else {
		return c.Copysign(+1)
	}
}

// MultipliedByRate returns the amount multiplied by a rate,
// rounded back to whole cents half away from zero.
// The multiplication is done in float64 precision.
// A NaN result yields zero and an out of range result is clamped
// to MinCentAmount or MaxCentAmount.
func (c CentAmount) MultipliedByRate(rate Rate) CentAmount {
	return centAmountFromFloat(float64(c) * float64(rate))
}

// DividedByRate returns the amount divided by a rate,
// rounded back to whole cents half away from zero.
// The division is done in float64 precision.
// A zero rate yields MinCentAmount or MaxCentAmount
// and a NaN result yields zero.
func (c CentAmount) DividedByRate(rate Rate) CentAmount {
	return centAmountFromFloat(float64(c) / float64(rate))
}

// Percentage returns the amount multiplied by (percent / 100),
// rounded back to whole cents half away from zero.
// The calculation is done in float64 precision.
// A NaN result yields zero and an out of range result is clamped
// to MinCentAmount or MaxCentAmount.
func (c CentAmount) Percentage(percent float64) CentAmount {
	return centAmountFromFloat(float64(c) * percent / 100)
}

// SplitEqually divides the amount equally into count amounts
// that sum up exactly to the initial amount.
// The cents that can't be divided evenly are handed to the last amounts,
// so no two amounts ever differ by more than one cent.
func (c CentAmount) SplitEqually(count int) []CentAmount {
	if count < 1 {
		return nil
	}
	if count == 1 {
		return []CentAmount{c}
	}
	// Integer division truncates towards zero, so the remainder has the sign
	// of c and a magnitude below count: it is exactly the number of amounts
	// that get one cent more. Handing the rest to a single amount instead
	// would make that amount differ by up to count cents and even flip its
	// sign, while the sum stayed correct.
	part := c / CentAmount(count)
	rest := c % CentAmount(count)
	extra := int(rest.Abs())
	one := CentAmount(rest.Sign())
	result := make([]CentAmount, count)
	for i := range count {
		result[i] = part
		if i >= count-extra {
			result[i] += one
		}
	}
	return result
}

// SplitProportionally splits an amount proportionally
// to the passed weights into the same number of amounts
// that sum up exactly to the initial amount.
// Only the magnitudes of the weights are used and every result amount
// gets the sign of the split amount, so the passed weights can be
// positive, negative, or zero.
// If all weights are zero then the whole amount ends up
// in the last result element.
//
// The cents lost to truncation are handed to the amounts with the largest
// truncated fraction (the largest remainder method), so every amount is
// within one cent of its exact share.
func (c CentAmount) SplitProportionally(weights []CentAmount) []CentAmount {
	count := len(weights)
	if count == 0 {
		return nil
	}
	if count == 1 {
		return []CentAmount{c} // Don't introduce rounding errors
	}
	totalWeight, shift := sumWeightMagnitudes(weights)
	result := make([]CentAmount, count)
	if totalWeight == 0 {
		result[count-1] = c
		return result
	}
	magnitude := abs64(int64(c))
	parts := make([]uint64, count)
	remainders := make([]uint64, count)
	var allocated uint64
	for i, weight := range weights {
		// Exact 128 bit magnitude*weight/totalWeight.
		// bits.Div64 can't overflow because every shifted weight magnitude is
		// less or equal totalWeight, so the quotient is less or equal magnitude.
		hi, lo := bits.Mul64(magnitude, abs64(int64(weight))>>shift)
		parts[i], remainders[i] = bits.Div64(hi, lo, totalWeight)
		allocated += parts[i]
	}
	// Every truncated quotient lost less than one cent, so the undistributed
	// rest is smaller than count and goes to the largest remainders.
	if rest := int(magnitude - allocated); rest > 0 {
		order := make([]int, count)
		for i := range order {
			order[i] = i
		}
		slices.SortStableFunc(order, func(x, y int) int {
			return cmp.Compare(remainders[y], remainders[x])
		})
		for _, i := range order[:rest] {
			parts[i]++
		}
	}
	for i, part := range parts {
		result[i] = CentAmount(part).Copysign(c)
	}
	return result
}

// sumWeightMagnitudes returns the sum of the weight magnitudes together with
// the number of bits every magnitude has to be shifted right for the sum to
// fit into an uint64. It only happens for weight magnitudes summing up to
// more than 2^64, which is far beyond any real money amount, and it truncates
// the smallest weights towards zero rather than scaling them exactly. Without it the sum would wrap silently:
// a wrap to zero would route the whole amount to the last result element and
// a small wrapped sum would make bits.Div64 panic with an integer overflow.
func sumWeightMagnitudes(weights []CentAmount) (total uint64, shift uint) {
	for {
		var hi, lo uint64
		for _, weight := range weights {
			var carry uint64
			lo, carry = bits.Add64(lo, abs64(int64(weight))>>shift, 0)
			hi += carry
		}
		if hi == 0 {
			return lo, shift
		}
		shift++
	}
}

// centAmountFromFloat returns f rounded to whole cents half away from zero.
// NaN maps to zero and a value outside of the valid CentAmount range is
// clamped to it, so a non-finite intermediate result never yields a
// platform dependent value and the result can always be negated.
func centAmountFromFloat(f float64) CentAmount {
	rounded := math.Round(f)
	switch {
	case math.IsNaN(rounded):
		return 0
	case rounded >= float64(MaxCentAmount):
		return MaxCentAmount
	case rounded <= float64(MinCentAmount):
		return MinCentAmount
	}
	return CentAmount(rounded)
}

// divRoundHalfAwayFromZero returns a/b rounded half away from zero.
// It panics if b is zero.
func divRoundHalfAwayFromZero(a, b int64) int64 {
	quo, rem := a/b, a%b
	if rem == 0 {
		return quo
	}
	// rem magnitude doubled can't overflow uint64 because |rem| < |b| <= 2^63
	if abs64(rem)*2 >= abs64(b) {
		if (a < 0) != (b < 0) {
			return quo - 1
		}
		return quo + 1
	}
	return quo
}
