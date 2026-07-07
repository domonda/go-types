package money

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"slices"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/nullable"
)

// Implemented interfaces
var (
	_ fmt.Stringer               = DecimalAmount{}
	_ fmt.GoStringer             = DecimalAmount{}
	_ fmt.Formatter              = DecimalAmount{}
	_ driver.Valuer              = DecimalAmount{}
	_ sql.Scanner                = (*DecimalAmount)(nil)
	_ json.Marshaler             = DecimalAmount{}
	_ json.Unmarshaler           = (*DecimalAmount)(nil)
	_ encoding.TextMarshaler     = DecimalAmount{}
	_ encoding.TextUnmarshaler   = (*DecimalAmount)(nil)
	_ encoding.BinaryMarshaler   = DecimalAmount{}
	_ encoding.BinaryUnmarshaler = (*DecimalAmount)(nil)
)

// RoundingMode selects how a value that cannot be represented exactly at the
// target scale is rounded. It is used by every DecimalAmount operation that may
// round: RoundToDecimals, RoundToCents, RoundToInt, Mul, Div, MultipliedByRate,
// DividedByRate, Percentage and the float conversions.
type RoundingMode uint8

const (
	// RoundHalfAwayFromZero rounds to nearest, ties away from zero (0.5 → 1,
	// -0.5 → -1). As the zero value of RoundingMode it is the conventional
	// default for commercial rounding.
	RoundHalfAwayFromZero RoundingMode = iota
	// RoundHalfToEven rounds to nearest, ties to the even neighbor (banker's
	// rounding: 0.5 → 0, 1.5 → 2).
	RoundHalfToEven
	// RoundHalfUp rounds to nearest, ties toward positive infinity.
	RoundHalfUp
	// RoundHalfDown rounds to nearest, ties toward negative infinity.
	RoundHalfDown
	// RoundDown rounds toward zero (truncates).
	RoundDown
	// RoundUp rounds away from zero.
	RoundUp
	// RoundFloor rounds toward negative infinity.
	RoundFloor
	// RoundCeil rounds toward positive infinity.
	RoundCeil
)

// String implements fmt.Stringer.
func (m RoundingMode) String() string {
	switch m {
	case RoundHalfAwayFromZero:
		return "RoundHalfAwayFromZero"
	case RoundHalfToEven:
		return "RoundHalfToEven"
	case RoundHalfUp:
		return "RoundHalfUp"
	case RoundHalfDown:
		return "RoundHalfDown"
	case RoundDown:
		return "RoundDown"
	case RoundUp:
		return "RoundUp"
	case RoundFloor:
		return "RoundFloor"
	case RoundCeil:
		return "RoundCeil"
	default:
		return fmt.Sprintf("Unknown rounding mode: %d", uint8(m))
	}
}

// DecimalAmount is an exact fixed-point monetary amount stored in a single int64.
//
// The value equals Coefficient × 10^-Scale. Both are packed into one int64:
// the low 5 bits hold the scale (number of decimal places, 0..MaxDecimalAmountScale)
// and the high 59 bits hold the two's-complement signed coefficient
// (range ±(2^58-1), roughly 18 significant decimal digits).
//
// Finite values are exact: unlike Amount, which is a float64, values like 0.10
// carry no rounding error. The per-value scale is preserved, so 1.50 (scale 2)
// and 1.5 (scale 1) are distinct representations of the same value; use Equal or
// Cmp for value comparison rather than the == operator, which compares the
// packed bits.
//
// DecimalAmount also has the non-finite states NaN, +Inf and -Inf, encoded with
// the otherwise-unused scale value 31. Arithmetic never panics on overflow or
// division by zero: overflow yields ±Inf, x/0 yields ±Inf (or NaN for 0/0), and
// NaN propagates like IEEE floating point. Use Valid/IsFinite/IsNaN/IsInf to
// test, mirroring Amount. (Explicit misuse — an out-of-range scale or decimals
// argument — still panics.)
//
// The zero value is a valid finite amount of 0 with scale 0.
type DecimalAmount struct {
	packed int64
}

const (
	// MaxDecimalAmountScale is the maximum number of decimal places a
	// DecimalAmount can represent. It is bounded by the largest power of ten
	// that fits in an int64 (10^18, used to rescale coefficients) and by the
	// ~18 significant digits of the 59-bit coefficient. The 5 scale bits could
	// encode up to 31, but scales above 18 add no usable precision, so raising
	// this limit later would be a backward-compatible change.
	MaxDecimalAmountScale = 18

	scaleBits = 5
	scaleMask = 1<<scaleBits - 1 // low 5 bits

	// nonFiniteScale is the reserved scale value (31, the top of the 5-bit
	// field) that tags NaN and ±Inf. The coefficient sign then selects the
	// kind: +1 → +Inf, -1 → -Inf, 0 → NaN.
	nonFiniteScale = scaleMask

	// The coefficient range is kept symmetric (excluding the extreme -2^58) so
	// that Neg and Abs never overflow a valid value.
	maxDecimalAmountCoefficient = 1<<(63-scaleBits) - 1 // 2^58 - 1
	minDecimalAmountCoefficient = -maxDecimalAmountCoefficient
)

// pow5[i] == 5^i for i in 0..MaxDecimalAmountScale (5^18 < 2^42), used by the
// int128 float-to-decimal conversion.
var pow5 = [MaxDecimalAmountScale + 1]int64{
	1,
	5,
	25,
	125,
	625,
	3125,
	15625,
	78125,
	390625,
	1953125,
	9765625,
	48828125,
	244140625,
	1220703125,
	6103515625,
	30517578125,
	152587890625,
	762939453125,
	3814697265625,
}

// decimalPow10[i] == 10^i for i in 0..MaxDecimalAmountScale.
// 10^18 is the largest power of ten that fits in an int64.
var decimalPow10 = [MaxDecimalAmountScale + 1]int64{
	1,
	10,
	100,
	1000,
	10000,
	100000,
	1000000,
	10000000,
	100000000,
	1000000000,
	10000000000,
	100000000000,
	1000000000000,
	10000000000000,
	100000000000000,
	1000000000000000,
	10000000000000000,
	100000000000000000,
	1000000000000000000,
}

// packDecimalAmount builds a DecimalAmount from a coefficient and an
// already-validated scale (0..MaxDecimalAmountScale), panicking if the
// coefficient exceeds the representable 59-bit range.
func packDecimalAmount(coefficient int64, scale int) DecimalAmount {
	if coefficient < minDecimalAmountCoefficient || coefficient > maxDecimalAmountCoefficient {
		panic(fmt.Sprintf("money.DecimalAmount coefficient %d overflows the representable range [%d, %d]", coefficient, minDecimalAmountCoefficient, maxDecimalAmountCoefficient))
	}
	return DecimalAmount{packed: coefficient<<scaleBits | int64(scale)}
}

// NewDecimalAmount returns a DecimalAmount equal to coefficient × 10^-scale.
// It panics if scale is not in [0, MaxDecimalAmountScale] or if coefficient is
// outside the representable range of roughly ±2.88×10^17.
func NewDecimalAmount(coefficient int64, scale int) DecimalAmount {
	if scale < 0 || scale > MaxDecimalAmountScale {
		panic(fmt.Sprintf("money.NewDecimalAmount scale %d out of range [0, %d]", scale, MaxDecimalAmountScale))
	}
	return packDecimalAmount(coefficient, scale)
}

// DecimalAmountFromInt returns i as a DecimalAmount with scale 0.
// It panics if i is outside the representable coefficient range.
func DecimalAmountFromInt(i int64) DecimalAmount {
	return NewDecimalAmount(i, 0)
}

// DecimalAmountNaN returns the not-a-number DecimalAmount.
func DecimalAmountNaN() DecimalAmount {
	return DecimalAmount{packed: nonFiniteScale}
}

// DecimalAmountInf returns positive infinity if sign >= 0, negative infinity
// otherwise.
func DecimalAmountInf(sign int) DecimalAmount {
	return decimalInf(sign < 0)
}

// decimalInf returns -Inf if negative, +Inf otherwise.
func decimalInf(negative bool) DecimalAmount {
	coeff := int64(1)
	if negative {
		coeff = -1
	}
	return DecimalAmount{packed: coeff<<scaleBits | nonFiniteScale}
}

// IsFinite reports whether the amount is a finite number (not NaN or ±Inf).
func (a DecimalAmount) IsFinite() bool {
	return int(a.packed&scaleMask) <= MaxDecimalAmountScale
}

// Valid reports whether the amount is finite. It mirrors Amount.Valid.
func (a DecimalAmount) Valid() bool {
	return a.IsFinite()
}

// IsNaN reports whether the amount is the not-a-number value.
func (a DecimalAmount) IsNaN() bool {
	return !a.IsFinite() && a.Coefficient() == 0
}

// IsInf reports whether the amount is an infinity of the given sign:
// sign > 0 tests +Inf, sign < 0 tests -Inf, sign == 0 tests either.
func (a DecimalAmount) IsInf(sign int) bool {
	if a.IsFinite() {
		return false
	}
	c := a.Coefficient()
	switch {
	case sign > 0:
		return c > 0
	case sign < 0:
		return c < 0
	default:
		return c != 0
	}
}

// DecimalAmountFromFloat converts f to a DecimalAmount at the given scale,
// rounding the exact value of the binary float64 with the given rounding mode.
// A NaN or infinite f maps to the corresponding non-finite DecimalAmount and an
// out-of-range value maps to ±Inf; an error is returned only for an out-of-range
// scale.
func DecimalAmountFromFloat(f float64, scale int, rounding RoundingMode) (DecimalAmount, error) {
	return decimalFromFloatRounded(f, scale, rounding)
}

// DecimalAmountFromAmount converts the float64-based Amount to a DecimalAmount at
// the given scale using the given rounding mode. It is the inverse of
// DecimalAmount.Amount and shares the precision caveat of DecimalAmountFromFloat.
func DecimalAmountFromAmount(amount Amount, scale int, rounding RoundingMode) (DecimalAmount, error) {
	return decimalFromFloatRounded(float64(amount), scale, rounding)
}

// ParseDecimalAmount parses an exact fixed-point amount from str using the same
// locale-aware separator detection as float.ParseDetails, so both "1,234.56"
// and "1.234,56" parse to 1234.56. Unlike ParseAmount it is exact: the
// coefficient is taken directly from the digits of str without a float64
// round-trip, so values like 99999999999999.99 keep their last cent.
//
// Like ParseAmount, a lone separator group is treated as a decimal separator,
// not a thousands separator: "1,000" and "1'000" parse to 1.000 (== 1), not
// 1000. Pass acceptedDecimals to reject such ambiguous inputs.
//
// The tokens "NaN", "Inf"/"+Inf"/"Infinity" and "-Inf"/"-Infinity" parse to the
// corresponding non-finite values. Scientific notation is rejected. If
// acceptedDecimals is non-empty, the number of decimal places must be one of the
// listed values. An error is returned if the value needs more than
// MaxDecimalAmountScale decimal places or does not fit the coefficient range.
func ParseDecimalAmount(str string, acceptedDecimals ...int) (DecimalAmount, error) {
	// The long "Infinity" spellings are not recognized by float.ParseDetails but
	// are the PostgreSQL numeric literals emitted by Value.
	switch strings.TrimSpace(str) {
	case "Infinity", "+Infinity":
		return decimalInf(false), nil
	case "-Infinity":
		return decimalInf(true), nil
	}
	f, _, _, decimals, err := float.ParseDetails(str)
	if err != nil {
		return DecimalAmount{}, err
	}
	if math.IsNaN(f) {
		return DecimalAmountNaN(), nil
	}
	if math.IsInf(f, 0) {
		return decimalInf(math.Signbit(f)), nil
	}
	if strings.ContainsAny(str, "eE") {
		return DecimalAmount{}, fmt.Errorf("scientific notation is not supported for money.DecimalAmount: %q", str)
	}
	if decimals > MaxDecimalAmountScale {
		return DecimalAmount{}, fmt.Errorf("money.DecimalAmount supports at most %d decimal places but %q has %d", MaxDecimalAmountScale, str, decimals)
	}
	if len(acceptedDecimals) > 0 && !slices.Contains(acceptedDecimals, decimals) {
		return DecimalAmount{}, fmt.Errorf("parsing %q returned %d decimals which is not in the accepted list %v", str, decimals, acceptedDecimals)
	}
	// The coefficient is exactly the sequence of all digit runes of str:
	// float.ParseDetails only accepts digits, sign, separators and the
	// exponent runes 'eE' (rejected above). Separators carry no digits, so
	// concatenating every digit reconstructs the coefficient, and decimals
	// (from ParseDetails) is the matching scale.
	var b strings.Builder
	b.Grow(len(str))
	for _, r := range str {
		if r >= '0' && r <= '9' {
			b.WriteByte(byte(r))
		}
	}
	coefficient, err := strconv.ParseInt(b.String(), 10, 64)
	if err != nil {
		return DecimalAmount{}, fmt.Errorf("money.DecimalAmount value %q is too large: %w", str, err)
	}
	if strings.ContainsRune(str, '-') {
		coefficient = -coefficient
	}
	if coefficient < minDecimalAmountCoefficient || coefficient > maxDecimalAmountCoefficient {
		return DecimalAmount{}, fmt.Errorf("money.DecimalAmount value %q does not fit the representable range", str)
	}
	return packDecimalAmount(coefficient, decimals), nil
}

// NullableDecimalAmount is a nullable DecimalAmount whose zero value is NULL.
type NullableDecimalAmount = nullable.Type[DecimalAmount]

// NullableDecimalAmountFrom returns a non-null NullableDecimalAmount wrapping value.
func NullableDecimalAmountFrom(value DecimalAmount) NullableDecimalAmount {
	return nullable.TypeFrom(value)
}

// NullableDecimalAmountFromPtr returns a NullableDecimalAmount from a pointer,
// using nil as the null value.
func NullableDecimalAmountFromPtr(ptr *DecimalAmount) NullableDecimalAmount {
	return nullable.TypeFromPtr(ptr)
}

// Coefficient returns the unscaled integer value; the finite amount equals
// Coefficient × 10^-Scale. It is only meaningful when IsFinite is true.
func (a DecimalAmount) Coefficient() int64 {
	return a.packed >> scaleBits // arithmetic shift preserves the sign
}

// Scale returns the number of decimal places (0..MaxDecimalAmountScale).
// It is only meaningful when IsFinite is true.
func (a DecimalAmount) Scale() int {
	return int(a.packed & scaleMask)
}

// Ptr returns a pointer to a copy of the amount.
func (a DecimalAmount) Ptr() *DecimalAmount {
	return &a
}

// Float returns the amount as a float64, which may lose precision.
// The non-finite states map to their float64 equivalents (NaN, ±Inf).
func (a DecimalAmount) Float() float64 {
	if !a.IsFinite() {
		switch {
		case a.IsNaN():
			return math.NaN()
		case a.Signbit():
			return math.Inf(-1)
		default:
			return math.Inf(1)
		}
	}
	return float64(a.Coefficient()) / float64(decimalPow10[a.Scale()])
}

// Amount returns the amount as a float64-based money.Amount,
// which may lose precision. See Float.
func (a DecimalAmount) Amount() Amount {
	return Amount(a.Float())
}

// Sign returns -1 if the amount is negative, +1 if positive and 0 if zero.
// -Inf returns -1, +Inf returns +1 and NaN returns 0.
func (a DecimalAmount) Sign() int {
	switch c := a.Coefficient(); {
	case c < 0:
		return -1
	case c > 0:
		return +1
	default:
		return 0
	}
}

// IsZero returns whether the amount is exactly finite zero.
func (a DecimalAmount) IsZero() bool {
	return a.IsFinite() && a.Coefficient() == 0
}

// Signbit reports whether the amount is negative (including -Inf).
func (a DecimalAmount) Signbit() bool {
	return a.Coefficient() < 0
}

// Abs returns the absolute value of the amount, keeping its scale.
// Abs(±Inf) is +Inf and Abs(NaN) is NaN.
func (a DecimalAmount) Abs() DecimalAmount {
	if !a.IsFinite() {
		if a.IsNaN() {
			return a
		}
		return decimalInf(false)
	}
	if c := a.Coefficient(); c < 0 {
		return packDecimalAmount(-c, a.Scale())
	}
	return a
}

// Neg returns the amount with its sign inverted, keeping its scale.
// Neg(±Inf) flips the infinity and Neg(NaN) is NaN.
func (a DecimalAmount) Neg() DecimalAmount {
	if !a.IsFinite() {
		if a.IsNaN() {
			return a
		}
		return decimalInf(!a.Signbit())
	}
	return packDecimalAmount(-a.Coefficient(), a.Scale())
}

// Cmp compares a and b by value, ignoring their scales, and returns
// -1 if a < b, +1 if a > b and 0 if they are equal, so 1.50 and 1.5 compare as
// equal. Non-finite values form a total order -Inf < finite < +Inf < NaN; in
// particular Cmp treats two NaNs as equal (unlike IEEE floating point) so that
// DecimalAmount values sort deterministically.
func (a DecimalAmount) Cmp(b DecimalAmount) int {
	if a.IsFinite() && b.IsFinite() {
		as, bs := a.Scale(), b.Scale()
		if as == bs {
			switch ac, bc := a.Coefficient(), b.Coefficient(); {
			case ac < bc:
				return -1
			case ac > bc:
				return +1
			default:
				return 0
			}
		}
		// Different scales: compare a.coeff·10^b.scale vs b.coeff·10^a.scale.
		// Both cross-products fit in 128 bits (coeff ≤ 2^58, 10^scale ≤ 10^18),
		// so compare signs first, then magnitudes with 128-bit integers.
		ac, bc := a.Coefficient(), b.Coefficient()
		if sa, sb := signOf(ac), signOf(bc); sa != sb {
			if sa < sb {
				return -1
			}
			return +1
		} else if sa == 0 {
			return 0 // both zero
		}
		aHi, aLo := bits.Mul64(abs64(ac), pow10u(bs))
		bHi, bLo := bits.Mul64(abs64(bc), pow10u(as))
		cmp := cmpUint128(aHi, aLo, bHi, bLo)
		if ac < 0 {
			cmp = -cmp // both negative: larger magnitude is the smaller value
		}
		return cmp
	}
	switch ra, rb := a.orderRank(), b.orderRank(); {
	case ra < rb:
		return -1
	case ra > rb:
		return +1
	default:
		return 0
	}
}

// orderRank maps a value to its position in the total order used by Cmp when at
// least one operand is non-finite: -Inf < finite < +Inf < NaN.
func (a DecimalAmount) orderRank() int {
	switch {
	case a.IsNaN():
		return 2
	case a.IsInf(1):
		return 1
	case a.IsInf(-1):
		return -1
	default:
		return 0
	}
}

// Equal reports whether a and b represent the same value, ignoring scale.
// It differs from the == operator, which compares the packed representation:
// NewDecimalAmount(150, 2) == NewDecimalAmount(15, 1) is false, while their
// Equal is true.
func (a DecimalAmount) Equal(b DecimalAmount) bool {
	return a.Cmp(b) == 0
}

// Add returns a + b. The result scale is the larger of the two scales.
// Overflow yields ±Inf; +Inf + -Inf and any NaN operand yield NaN.
func (a DecimalAmount) Add(b DecimalAmount) DecimalAmount {
	if r, ok := addNonFinite(a, b); ok {
		return r
	}
	scale := a.Scale()
	if bs := b.Scale(); bs > scale {
		scale = bs
	}
	// Aligning to the larger scale only ever increases scale (pads with zeros),
	// so no rounding happens and the mode is irrelevant. Alignment overflow
	// means the dominant operand is unrepresentable at the target scale, so the
	// result is that operand's infinity.
	ac, aOver := scaleCoefficient(a.Coefficient(), a.Scale(), scale, RoundHalfAwayFromZero)
	if aOver {
		return decimalInf(a.Signbit())
	}
	bc, bOver := scaleCoefficient(b.Coefficient(), b.Scale(), scale, RoundHalfAwayFromZero)
	if bOver {
		return decimalInf(b.Signbit())
	}
	sum, over := addOverflow(ac, bc)
	if over || sum < minDecimalAmountCoefficient || sum > maxDecimalAmountCoefficient {
		return decimalInf(ac < 0)
	}
	return packDecimalAmount(sum, scale)
}

// Sub returns a - b. The result scale is the larger of the two scales.
// Overflow yields ±Inf and NaN operands propagate; see Add.
func (a DecimalAmount) Sub(b DecimalAmount) DecimalAmount {
	return a.Add(b.Neg())
}

// MulInt returns a multiplied by the integer n, keeping a's scale.
// Overflow yields ±Inf; a NaN receiver yields NaN and ±Inf × 0 yields NaN.
func (a DecimalAmount) MulInt(n int) DecimalAmount {
	return a.MulInt64(int64(n))
}

// MulInt64 returns a multiplied by the integer n, keeping a's scale.
// Overflow yields ±Inf; a NaN receiver yields NaN and ±Inf × 0 yields NaN.
func (a DecimalAmount) MulInt64(n int64) DecimalAmount {
	if !a.IsFinite() {
		switch {
		case a.IsNaN():
			return a
		case n == 0:
			return DecimalAmountNaN()
		default:
			return decimalInf(a.Signbit() != (n < 0))
		}
	}
	product, over := mulOverflow(a.Coefficient(), n)
	if over || product < minDecimalAmountCoefficient || product > maxDecimalAmountCoefficient {
		return decimalInf((a.Coefficient() < 0) != (n < 0))
	}
	return packDecimalAmount(product, a.Scale())
}

// MultipliedByRate returns the amount multiplied by the float64 Rate, rounded to
// the given number of decimal places with the given rounding mode. Because Rate
// is a float64 the product is computed in floating point, so the result is exact
// only to the requested scale. A non-finite product yields the corresponding
// NaN or ±Inf. It panics only if scale is out of range.
func (a DecimalAmount) MultipliedByRate(rate Rate, scale int, rounding RoundingMode) DecimalAmount {
	return mustDecimalFromFloat(a.Float()*float64(rate), scale, rounding)
}

// DividedByRate returns the amount divided by the float64 Rate, rounded to the
// given number of decimal places with the given rounding mode. See
// MultipliedByRate for the precision caveat. Division by a zero rate yields
// ±Inf (or NaN). It panics only if scale is out of range.
func (a DecimalAmount) DividedByRate(rate Rate, scale int, rounding RoundingMode) DecimalAmount {
	return mustDecimalFromFloat(a.Float()/float64(rate), scale, rounding)
}

// Percentage returns the amount multiplied by (percent / 100), rounded to the
// given number of decimal places with the given rounding mode. See
// MultipliedByRate for the precision caveat.
func (a DecimalAmount) Percentage(percent float64, scale int, rounding RoundingMode) DecimalAmount {
	return mustDecimalFromFloat(a.Float()*percent/100, scale, rounding)
}

// Mul returns a × b rounded to the scale of a using the given rounding mode.
// The exact product is computed in 128-bit integer arithmetic (no heap
// allocation), so no intermediate precision is lost before rounding. Overflow
// yields ±Inf; NaN propagates and ±Inf × 0 yields NaN.
func (a DecimalAmount) Mul(b DecimalAmount, rounding RoundingMode) DecimalAmount {
	if r, ok := mulNonFinite(a, b); ok {
		return r
	}
	ca, cb := a.Coefficient(), b.Coefficient()
	negative := (ca < 0) != (cb < 0)
	// Exact product coefficient ca·cb has scale a.Scale()+b.Scale(); dividing
	// by 10^b.Scale() brings it to the target scale a.Scale().
	hi, lo := bits.Mul64(abs64(ca), abs64(cb))
	return decimalDivRound(hi, lo, pow10u(b.Scale()), negative, a.Scale(), rounding)
}

// Div returns a / b rounded to the scale of a using the given rounding mode.
// The quotient is computed in 128-bit integer arithmetic (no heap allocation).
// Division by zero yields ±Inf (or NaN for 0/0), overflow yields ±Inf, and NaN
// propagates.
func (a DecimalAmount) Div(b DecimalAmount, rounding RoundingMode) DecimalAmount {
	if r, ok := divNonFinite(a, b); ok {
		return r
	}
	ca, cb := a.Coefficient(), b.Coefficient()
	if cb == 0 {
		if ca == 0 {
			return DecimalAmountNaN()
		}
		return decimalInf(ca < 0)
	}
	negative := (ca < 0) != (cb < 0)
	// result_coeff = round( ca·10^b.Scale() / cb ) at scale a.Scale().
	hi, lo := bits.Mul64(abs64(ca), pow10u(b.Scale()))
	return decimalDivRound(hi, lo, abs64(cb), negative, a.Scale(), rounding)
}

// RoundToDecimals returns the amount rounded to the given number of decimal
// places with the given rounding mode. The result always has scale == decimals;
// increasing the scale pads with zeros. A non-finite amount is returned
// unchanged and padding overflow yields ±Inf. It panics if decimals is out of
// range [0, MaxDecimalAmountScale].
func (a DecimalAmount) RoundToDecimals(decimals int, rounding RoundingMode) DecimalAmount {
	if decimals < 0 || decimals > MaxDecimalAmountScale {
		panic(fmt.Sprintf("money.DecimalAmount.RoundToDecimals decimals %d out of range [0, %d]", decimals, MaxDecimalAmountScale))
	}
	if !a.IsFinite() {
		return a
	}
	coeff, over := scaleCoefficient(a.Coefficient(), a.Scale(), decimals, rounding)
	if over {
		return decimalInf(a.Signbit())
	}
	return packDecimalAmount(coeff, decimals)
}

// RoundToCents returns the amount rounded to 2 decimal places using the given
// rounding mode.
func (a DecimalAmount) RoundToCents(rounding RoundingMode) DecimalAmount {
	return a.RoundToDecimals(2, rounding)
}

// RoundToInt returns the amount rounded to an integer (scale 0) using the given
// rounding mode.
func (a DecimalAmount) RoundToInt(rounding RoundingMode) DecimalAmount {
	return a.RoundToDecimals(0, rounding)
}

// String returns the exact decimal representation of the amount using a dot as
// decimal separator and no thousands separator, with exactly Scale digits after
// the point. Non-finite values return "NaN", "Inf" or "-Inf".
// String implements fmt.Stringer.
func (a DecimalAmount) String() string {
	if !a.IsFinite() {
		return nonFiniteString(a)
	}
	intPart, fracPart, neg := a.parts()
	var b strings.Builder
	b.Grow(len(intPart) + len(fracPart) + 2)
	if neg {
		b.WriteByte('-')
	}
	b.WriteString(intPart)
	if fracPart != "" {
		b.WriteByte('.')
		b.WriteString(fracPart)
	}
	return b.String()
}

// nonFiniteString returns the token for a non-finite value. These tokens
// round-trip through ParseDecimalAmount.
func nonFiniteString(a DecimalAmount) string {
	switch {
	case a.IsNaN():
		return "NaN"
	case a.Signbit():
		return "-Inf"
	default:
		return "Inf"
	}
}

// GoString returns a Go source representation of the amount for debugging.
// GoString implements fmt.GoStringer.
func (a DecimalAmount) GoString() string {
	switch {
	case a.IsNaN():
		return "money.DecimalAmountNaN()"
	case a.IsInf(1):
		return "money.DecimalAmountInf(1)"
	case a.IsInf(-1):
		return "money.DecimalAmountInf(-1)"
	default:
		return fmt.Sprintf("money.NewDecimalAmount(%d, %d)", a.Coefficient(), a.Scale())
	}
}

// FormatSep returns the exact decimal representation of the amount with the
// given decimal separator and optional thousands separator grouping the integer
// part. A zero thousandsSep disables grouping; a zero decimalSep defaults to
// '.'. The number of fractional digits is the amount's Scale.
func (a DecimalAmount) FormatSep(thousandsSep, decimalSep rune) string {
	if !a.IsFinite() {
		return nonFiniteString(a)
	}
	if decimalSep == 0 {
		decimalSep = '.'
	}
	intPart, fracPart, neg := a.parts()
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	writeGrouped(&b, intPart, thousandsSep)
	if fracPart != "" {
		b.WriteRune(decimalSep)
		b.WriteString(fracPart)
	}
	return b.String()
}

// Format implements fmt.Formatter, giving the amount format-string aware output:
//
//	%s, %v  exact decimal string (like String)
//	%#v     Go source representation (like GoString)
//	%q      double-quoted exact decimal string
//	%f, %F  fixed-point with the precision as decimal places
//	        (default: the amount's own Scale)
//	%d      integer part, rounded to zero decimals
//
// The '+' and ' ' flags control the sign of positive values, the '#' flag
// groups the integer part with ',' thousands separators (for %v, %s, %f and
// %d), and the width with the '-' and '0' flags controls padding. Formatting
// never panics.
func (a DecimalAmount) Format(f fmt.State, verb rune) {
	if !a.IsFinite() {
		if verb == 'v' && f.Flag('#') {
			writeStr(f, a.GoString())
			return
		}
		s := a.String() // "NaN", "Inf" or "-Inf"
		if a.IsInf(1) {
			if f.Flag('+') {
				s = "+" + s
			} else if f.Flag(' ') {
				s = " " + s
			}
		}
		if verb == 'q' {
			s = strconv.Quote(s)
		}
		// Non-finite tokens are never zero-padded, matching the standard
		// library (e.g. "%08.2f" of +Inf is "     Inf", not "00000Inf").
		writeSpacePadded(f, s)
		return
	}
	switch verb {
	case 'v':
		if f.Flag('#') {
			writeStr(f, a.GoString())
			return
		}
		writePadded(f, a.formatBody(f, a.Scale()))
	case 's':
		writePadded(f, a.formatBody(f, a.Scale()))
	case 'q':
		writePadded(f, strconv.Quote(a.formatBody(f, a.Scale())))
	case 'f', 'F':
		prec := a.Scale()
		if p, ok := f.Precision(); ok {
			prec = min(p, MaxDecimalAmountScale)
		}
		writePadded(f, a.formatBody(f, prec))
	case 'd':
		writePadded(f, a.formatBody(f, 0))
	default:
		fmt.Fprintf(f, "%%!%c(money.DecimalAmount=%s)", verb, a.String())
	}
}

// MarshalJSON implements json.Marshaler, encoding a finite amount as an unquoted
// JSON number that preserves the exact value and scale, e.g. 1234.56. Non-finite
// values are encoded as the quoted strings "NaN", "Inf" or "-Inf" (JSON has no
// number literal for them). UnmarshalJSON accepts a quoted string for any value.
func (a DecimalAmount) MarshalJSON() ([]byte, error) {
	if !a.IsFinite() {
		return []byte(strconv.Quote(a.String())), nil
	}
	return []byte(a.String()), nil
}

// UnmarshalJSON implements json.Unmarshaler and accepts a JSON number, a quoted
// decimal string, or null. null and "" are decoded as zero.
func (a *DecimalAmount) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		*a = DecimalAmount{}
		return nil
	}
	if l := len(s); l >= 2 && s[0] == '"' && s[l-1] == '"' {
		s = s[1 : l-1]
	}
	parsed, err := ParseDecimalAmount(s)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON %s as money.DecimalAmount: %w", data, err)
	}
	*a = parsed
	return nil
}

// MarshalText implements encoding.TextMarshaler, returning the exact decimal
// string (like String).
func (a DecimalAmount) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. An empty text is decoded
// as zero.
func (a *DecimalAmount) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*a = DecimalAmount{}
		return nil
	}
	parsed, err := ParseDecimalAmount(string(text))
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler, encoding the packed int64
// as 8 big-endian bytes.
func (a DecimalAmount) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(a.packed)) //#nosec G115 -- packed is a bit pattern
	return buf, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler, decoding 8 big-endian
// bytes produced by MarshalBinary.
func (a *DecimalAmount) UnmarshalBinary(data []byte) error {
	if len(data) != 8 {
		return fmt.Errorf("money.DecimalAmount.UnmarshalBinary needs 8 bytes but got %d", len(data))
	}
	packed := int64(binary.BigEndian.Uint64(data)) //#nosec G115 -- restoring a bit pattern
	if int(packed&scaleMask) > MaxDecimalAmountScale {
		// Any scale above the maximum is non-finite; canonicalize it to scale
		// nonFiniteScale from its coefficient sign.
		switch coeff := packed >> scaleBits; {
		case coeff > 0:
			*a = decimalInf(false)
		case coeff < 0:
			*a = decimalInf(true)
		default:
			*a = DecimalAmountNaN()
		}
		return nil
	}
	a.packed = packed
	return nil
}

// Value implements the database/sql/driver.Valuer interface, returning the
// exact decimal string so it can be stored in a SQL numeric/decimal column
// without precision loss. Non-finite values use the PostgreSQL numeric literals
// "NaN", "Infinity" and "-Infinity"; Scan and ParseDecimalAmount read them back.
func (a DecimalAmount) Value() (driver.Value, error) {
	if !a.IsFinite() {
		switch {
		case a.IsNaN():
			return "NaN", nil
		case a.Signbit():
			return "-Infinity", nil
		default:
			return "Infinity", nil
		}
	}
	return a.String(), nil
}

// Scan implements the database/sql.Scanner interface and accepts string,
// []byte, int64 and float64. A float64 is converted via its shortest exact
// decimal representation, or rounded to MaxDecimalAmountScale if that needs
// more fractional digits.
func (a *DecimalAmount) Scan(value any) error {
	switch x := value.(type) {
	case string:
		return a.scanParse(x)
	case []byte:
		return a.scanParse(string(x))
	case int64:
		if x < minDecimalAmountCoefficient || x > maxDecimalAmountCoefficient {
			return fmt.Errorf("int64 %d is out of range for money.DecimalAmount", x)
		}
		*a = packDecimalAmount(x, 0)
		return nil
	case float64:
		// Use the shortest exact decimal when it fits; a value needing more
		// than MaxDecimalAmountScale fractional digits is rounded to that scale
		// rather than rejected.
		if parsed, err := ParseDecimalAmount(strconv.FormatFloat(x, 'f', -1, 64)); err == nil {
			*a = parsed
			return nil
		}
		parsed, err := DecimalAmountFromFloat(x, MaxDecimalAmountScale, RoundHalfAwayFromZero)
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	default:
		return fmt.Errorf("can't scan value of type %T as money.DecimalAmount", value)
	}
}

// ScanString implements the strfmt scanning convention by parsing source as a
// DecimalAmount. Parsing already rejects invalid input, so the validate flag
// has no additional effect.
func (a *DecimalAmount) ScanString(source string, validate bool) error {
	return a.scanParse(source)
}

// JSONSchema returns the JSON Schema for a DecimalAmount as a numeric type,
// matching the JSON number produced by MarshalJSON.
func (DecimalAmount) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "number",
		Title:       "Decimal Amount",
		Description: "Exact fixed-point monetary amount",
	}
}

func (a *DecimalAmount) scanParse(s string) error {
	parsed, err := ParseDecimalAmount(s)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}

// parts splits the absolute value into its integer and fractional digit strings
// (fracPart has exactly Scale digits, or is empty when Scale is 0) and reports
// whether the amount is negative.
func (a DecimalAmount) parts() (intPart, fracPart string, neg bool) {
	return decimalDigits(a.Coefficient(), a.Scale())
}

// formatBody renders the amount to prec decimal places (rounding half away from
// zero, padding with zeros when prec exceeds the scale) and applies the sign
// ('+'/' ') and grouping ('#') flags of f, but not width/padding. It never
// panics: reducing precision cannot overflow and padding is done with string
// zeros rather than by rescaling the coefficient.
func (a DecimalAmount) formatBody(f fmt.State, prec int) string {
	var intPart, fracPart string
	var neg bool
	if scale := a.Scale(); prec >= scale {
		intPart, fracPart, neg = a.parts()
		if pad := prec - scale; pad > 0 {
			fracPart += strings.Repeat("0", pad)
		}
	} else {
		// Display rounding uses half away from zero; a format verb cannot carry
		// a RoundingMode. Reducing precision never overflows.
		coeff, _ := scaleCoefficient(a.Coefficient(), scale, prec, RoundHalfAwayFromZero)
		intPart, fracPart, neg = decimalDigits(coeff, prec)
	}

	var b strings.Builder
	switch {
	case neg:
		b.WriteByte('-')
	case f.Flag('+'):
		b.WriteByte('+')
	case f.Flag(' '):
		b.WriteByte(' ')
	}
	if f.Flag('#') {
		writeGrouped(&b, intPart, ',')
	} else {
		b.WriteString(intPart)
	}
	if fracPart != "" {
		b.WriteByte('.')
		b.WriteString(fracPart)
	}
	return b.String()
}

// mustDecimalFromFloat converts f to a DecimalAmount rounded to scale with the
// given rounding mode. NaN, infinity and overflow yield the corresponding
// non-finite value; it panics only on an out-of-range scale (a programmer
// error).
func mustDecimalFromFloat(f float64, scale int, rounding RoundingMode) DecimalAmount {
	d, err := decimalFromFloatRounded(f, scale, rounding)
	if err != nil {
		panic(fmt.Sprintf("money.DecimalAmount from float %v: %s", f, err))
	}
	return d
}

// decimalDigits splits abs(coeff) into integer and fractional digit strings for
// the given scale and reports whether coeff is negative.
func decimalDigits(coeff int64, scale int) (intPart, fracPart string, neg bool) {
	neg = coeff < 0
	if neg {
		coeff = -coeff
	}
	digits := strconv.FormatInt(coeff, 10)
	switch {
	case scale == 0:
		return digits, "", neg
	case len(digits) <= scale:
		return "0", strings.Repeat("0", scale-len(digits)) + digits, neg
	default:
		return digits[:len(digits)-scale], digits[len(digits)-scale:], neg
	}
}

// writeGrouped writes the ASCII digit string intPart to b, inserting sep between
// every group of three digits from the right. A zero sep disables grouping.
func writeGrouped(b *strings.Builder, intPart string, sep rune) {
	if sep == 0 {
		b.WriteString(intPart)
		return
	}
	n := len(intPart)
	for i := range n {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteRune(sep)
		}
		b.WriteByte(intPart[i])
	}
}

// writeSpacePadded writes s to f padded with spaces to the field width,
// honoring the '-' (left-align) flag but never the '0' flag. Used for
// non-finite tokens, which the standard library never zero-pads.
func writeSpacePadded(f fmt.State, s string) {
	width, ok := f.Width()
	if !ok || len(s) >= width {
		writeStr(f, s)
		return
	}
	pad := strings.Repeat(" ", width-len(s))
	if f.Flag('-') {
		writeStr(f, s+pad)
	} else {
		writeStr(f, pad+s)
	}
}

// writePadded writes s to f, applying the width and the '-' and '0' padding
// flags. s is expected to be ASCII (digits, sign, '.', ',', '"').
func writePadded(f fmt.State, s string) {
	width, ok := f.Width()
	if !ok || len(s) >= width {
		writeStr(f, s)
		return
	}
	pad := strings.Repeat(padByte(f), width-len(s))
	switch {
	case f.Flag('-'):
		writeStr(f, s)
		writeStr(f, pad)
	case f.Flag('0') && len(s) > 0 && (s[0] == '-' || s[0] == '+' || s[0] == ' '):
		// Zero-pad after the sign so it stays leftmost.
		writeStr(f, s[:1])
		writeStr(f, pad)
		writeStr(f, s[1:])
	default:
		writeStr(f, pad)
		writeStr(f, s)
	}
}

func padByte(f fmt.State) string {
	if f.Flag('0') && !f.Flag('-') {
		return "0"
	}
	return " "
}

// mulOverflow multiplies a and b and reports overflow. The first argument must
// be a value within the representable coefficient range so that the division
// check can never hit the MinInt64/-1 division trap.
func mulOverflow(a, b int64) (int64, bool) {
	if b == 0 {
		return 0, false
	}
	p := a * b
	if p/b != a {
		return 0, true
	}
	return p, false
}

// addOverflow adds a and b and reports signed overflow.
func addOverflow(a, b int64) (int64, bool) {
	s := a + b
	if (a > 0 && b > 0 && s < 0) || (a < 0 && b < 0 && s >= 0) {
		return 0, true
	}
	return s, false
}

// pow10u returns 10^scale as a uint64 for 0 ≤ scale ≤ MaxDecimalAmountScale.
func pow10u(scale int) uint64 {
	return uint64(decimalPow10[scale]) //#nosec G115 -- 10^scale is a positive int64
}

// writeStr writes s to the fmt.State f, discarding the write error, which a
// fmt.State never produces meaningfully.
func writeStr(f fmt.State, s string) {
	_, _ = io.WriteString(f, s) //#nosec G104 -- writing to a fmt.State cannot fail meaningfully
}

// abs64 returns the magnitude of x as an unsigned integer. It is safe for every
// value in the coefficient range, which never reaches math.MinInt64.
func abs64(x int64) uint64 {
	if x < 0 {
		return uint64(-x)
	}
	return uint64(x)
}

// signOf returns -1, 0 or +1 for the sign of x.
func signOf(x int64) int {
	switch {
	case x < 0:
		return -1
	case x > 0:
		return +1
	default:
		return 0
	}
}

// cmpUint128 compares two 128-bit unsigned values (hi, lo), returning -1, 0 or +1.
func cmpUint128(aHi, aLo, bHi, bLo uint64) int {
	switch {
	case aHi < bHi:
		return -1
	case aHi > bHi:
		return +1
	case aLo < bLo:
		return -1
	case aLo > bLo:
		return +1
	default:
		return 0
	}
}

// decimalDivRound divides the 128-bit magnitude (numHi, numLo) by divisor,
// rounds the quotient to an integer coefficient with the given rounding mode and
// sign, and packs it at scale. It returns ±Inf if the result does not fit the
// coefficient range. divisor must be non-zero.
func decimalDivRound(numHi, numLo, divisor uint64, negative bool, scale int, rounding RoundingMode) DecimalAmount {
	quoHi, quoLo, rem := div128by64(numHi, numLo, divisor)
	if roundUpMagnitude(rounding, quoLo, rem, divisor, negative) {
		quoLo++
		if quoLo == 0 {
			quoHi++
		}
	}
	if quoHi != 0 || quoLo > uint64(maxDecimalAmountCoefficient) {
		return decimalInf(negative)
	}
	coeff := int64(quoLo)
	if negative {
		coeff = -coeff
	}
	return packDecimalAmount(coeff, scale)
}

// addNonFinite applies the non-finite rules of Add/Sub. It returns (result,
// true) when at least one operand is non-finite, otherwise (_, false).
func addNonFinite(a, b DecimalAmount) (DecimalAmount, bool) {
	if a.IsFinite() && b.IsFinite() {
		return DecimalAmount{}, false
	}
	if a.IsNaN() || b.IsNaN() {
		return DecimalAmountNaN(), true
	}
	// At least one is ±Inf, neither is NaN.
	if !a.IsFinite() && !b.IsFinite() {
		if a.Signbit() != b.Signbit() {
			return DecimalAmountNaN(), true // +Inf + -Inf
		}
		return a, true
	}
	if !a.IsFinite() {
		return a, true
	}
	return b, true
}

// mulNonFinite applies the non-finite rules of Mul.
func mulNonFinite(a, b DecimalAmount) (DecimalAmount, bool) {
	if a.IsFinite() && b.IsFinite() {
		return DecimalAmount{}, false
	}
	if a.IsNaN() || b.IsNaN() {
		return DecimalAmountNaN(), true
	}
	if a.IsZero() || b.IsZero() { // ±Inf × 0
		return DecimalAmountNaN(), true
	}
	return decimalInf(a.Signbit() != b.Signbit()), true
}

// divNonFinite applies the non-finite rules of Div. Division by a finite zero is
// left to the finite path.
func divNonFinite(a, b DecimalAmount) (DecimalAmount, bool) {
	if a.IsFinite() && b.IsFinite() {
		return DecimalAmount{}, false
	}
	if a.IsNaN() || b.IsNaN() {
		return DecimalAmountNaN(), true
	}
	if !a.IsFinite() && !b.IsFinite() {
		return DecimalAmountNaN(), true // Inf / Inf
	}
	if !a.IsFinite() { // Inf / finite
		return decimalInf(a.Signbit() != b.Signbit()), true
	}
	// finite / Inf = 0 at the receiver's scale.
	return packDecimalAmount(0, a.Scale()), true
}

// div128by64 divides the 128-bit unsigned value (hi, lo) by d (which must be
// non-zero), returning the 128-bit quotient (quoHi, quoLo) and the 64-bit
// remainder.
func div128by64(hi, lo, d uint64) (quoHi, quoLo, rem uint64) {
	if hi >= d {
		quoHi, hi = bits.Div64(0, hi, d) // hi/d and hi%d; safe because 0 < d
	}
	quoLo, rem = bits.Div64(hi, lo, d) // hi < d now, so no quotient overflow
	return quoHi, quoLo, rem
}

// roundUpMagnitude reports whether the truncated quotient magnitude quo (with
// remainder rem over divisor d and the given sign) should be incremented by one
// under the rounding mode. rem must be less than d.
func roundUpMagnitude(mode RoundingMode, quo, rem, d uint64, negative bool) bool {
	if rem == 0 {
		return false
	}
	// Compare 2·rem to d to locate rem relative to the halfway point.
	// rem < d ≤ 10^18 < 2^60, so doubling cannot overflow.
	var halfCmp int
	switch double := rem * 2; {
	case double > d:
		halfCmp = +1
	case double < d:
		halfCmp = -1
	}
	return roundAwayFromZero(mode, negative, quo&1 == 1, halfCmp)
}

// roundAwayFromZero is the shared rounding decision for a value truncated toward
// zero that has a non-zero remainder. quotientOdd is the parity of the truncated
// magnitude and halfCmp compares the remainder to the halfway point
// (-1 below, 0 exactly half, +1 above). It reports whether the magnitude should
// be incremented (i.e. rounded away from zero).
func roundAwayFromZero(mode RoundingMode, negative, quotientOdd bool, halfCmp int) bool {
	switch mode {
	case RoundDown:
		return false
	case RoundUp:
		return true
	case RoundFloor:
		return negative // toward -Inf: away from zero only when negative
	case RoundCeil:
		return !negative // toward +Inf: away from zero only when positive
	}
	// Ties-to-nearest variants.
	if halfCmp > 0 {
		return true
	}
	if halfCmp < 0 {
		return false
	}
	switch mode {
	case RoundHalfToEven:
		return quotientOdd
	case RoundHalfUp:
		return !negative
	case RoundHalfDown:
		return negative
	default: // RoundHalfAwayFromZero
		return true
	}
}

// decimalFromFloatRounded converts f to a DecimalAmount at the given scale using
// the exact value of the binary float64 and the given rounding mode, computed in
// 128-bit integer arithmetic without heap allocation. A NaN or infinite f maps
// to the corresponding non-finite DecimalAmount and an out-of-range value maps to
// ±Inf. It returns an error only for an out-of-range scale (a programmer error).
func decimalFromFloatRounded(f float64, scale int, rounding RoundingMode) (DecimalAmount, error) {
	if scale < 0 || scale > MaxDecimalAmountScale {
		return DecimalAmount{}, fmt.Errorf("money.DecimalAmount scale %d out of range [0, %d]", scale, MaxDecimalAmountScale)
	}
	switch {
	case math.IsNaN(f):
		return DecimalAmountNaN(), nil
	case math.IsInf(f, 0):
		return decimalInf(math.Signbit(f)), nil
	case f == 0:
		return packDecimalAmount(0, scale), nil
	}
	// Decompose |f| exactly as m·2^e with m a 53-bit integer mantissa. Then
	// value·10^scale = m·2^e·10^scale = (m·5^scale)·2^(e+scale). Since m ≤ 2^53
	// and 5^scale < 2^42, the product m·5^scale fits in 128 bits, and the
	// remaining factor is a pure power of two, i.e. a shift.
	negative := math.Signbit(f)
	frac, exp := math.Frexp(f) // f = frac·2^exp, 0.5 ≤ |frac| < 1
	m := uint64(math.Ldexp(math.Abs(frac), 53))
	hi, lo := bits.Mul64(m, uint64(pow5[scale])) //#nosec G115 -- 5^scale is a positive int64
	var mag uint64
	var over bool
	if s := (exp - 53) + scale; s >= 0 {
		mag, over = shiftLeft128(hi, lo, s) // exact, no rounding
	} else {
		mag, over = shiftRightRound128(hi, lo, -s, negative, rounding)
	}
	if over || mag > uint64(maxDecimalAmountCoefficient) {
		return decimalInf(negative), nil
	}
	coeff := int64(mag)
	if negative {
		coeff = -coeff
	}
	return packDecimalAmount(coeff, scale), nil
}

// shiftLeft128 returns ((hi,lo) << n) as a uint64, reporting overflow if the
// result does not fit in 64 bits (the coefficient range is below 2^64).
func shiftLeft128(hi, lo uint64, n int) (uint64, bool) {
	if hi != 0 {
		return 0, true // value ≥ 2^64, shifting left only grows it
	}
	if n >= 64 {
		return 0, lo != 0
	}
	r := lo << n
	if r>>n != lo {
		return 0, true // high bits shifted out
	}
	return r, false
}

// shiftRightRound128 returns round((hi,lo) / 2^k) as a uint64 using the rounding
// mode and sign, reporting overflow if the quotient does not fit in 64 bits.
func shiftRightRound128(hi, lo uint64, k int, negative bool, rounding RoundingMode) (uint64, bool) {
	var qHi, qLo uint64
	switch {
	case k == 0:
		qHi, qLo = hi, lo
	case k < 64:
		qLo = lo>>k | hi<<(64-k)
		qHi = hi >> k
	case k < 128:
		qLo = hi >> (k - 64)
	}
	if qHi != 0 {
		return 0, true
	}
	// The bits shifted out determine rounding: the round bit is bit k-1 and the
	// sticky bit is any set bit below it.
	roundBit := bit128(hi, lo, k-1)
	sticky := lowBitsNonzero128(hi, lo, k-1)
	if roundBit == 0 && !sticky {
		return qLo, false // exact
	}
	halfCmp := -1 // remainder below the halfway point
	if roundBit != 0 {
		if sticky {
			halfCmp = +1 // above halfway
		} else {
			halfCmp = 0 // exactly halfway
		}
	}
	if roundAwayFromZero(rounding, negative, qLo&1 == 1, halfCmp) {
		qLo++
		if qLo == 0 { // wrapped past uint64
			return 0, true
		}
	}
	return qLo, false
}

// bit128 returns bit i (0-based) of the 128-bit value (hi,lo), or 0 out of range.
func bit128(hi, lo uint64, i int) uint64 {
	switch {
	case i < 0 || i >= 128:
		return 0
	case i < 64:
		return lo >> i & 1
	default:
		return hi >> (i - 64) & 1
	}
}

// lowBitsNonzero128 reports whether any of bits [0, n) of (hi,lo) are set.
func lowBitsNonzero128(hi, lo uint64, n int) bool {
	switch {
	case n <= 0:
		return false
	case n >= 128:
		return hi != 0 || lo != 0
	case n <= 64:
		return lo&(1<<n-1) != 0
	default:
		return lo != 0 || hi&(1<<(n-64)-1) != 0
	}
}

// scaleCoefficient returns coeff represented at toScale instead of fromScale and
// reports overflow. Increasing the scale multiplies by a power of ten and
// overflows when the coefficient no longer fits the representable range;
// decreasing the scale rounds according to rounding and never overflows.
func scaleCoefficient(coeff int64, fromScale, toScale int, rounding RoundingMode) (int64, bool) {
	switch {
	case toScale == fromScale:
		return coeff, false
	case toScale > fromScale:
		product, over := mulOverflow(coeff, decimalPow10[toScale-fromScale])
		if over || product < minDecimalAmountCoefficient || product > maxDecimalAmountCoefficient {
			return 0, true
		}
		return product, false
	default:
		negative := coeff < 0
		u := abs64(coeff)
		d := pow10u(fromScale - toScale)
		q := u / d
		if roundUpMagnitude(rounding, q, u%d, d, negative) {
			q++
		}
		// q is a bounded coefficient magnitude ≤ maxDecimalAmountCoefficient.
		if negative {
			return -int64(q), false //#nosec G115 -- q is a bounded coefficient magnitude
		}
		return int64(q), false //#nosec G115 -- q is a bounded coefficient magnitude
	}
}
