package float

import (
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/domonda/go-types/internal"
)

// RoundToDecimals returns f rounded half away from zero
// to the passed number of decimal places.
// Negative decimals round to tens, hundreds, and so on.
//
// It rounds the shortest decimal representation of f, which is the decimal
// literal that parses back to f, instead of its binary value, so
// RoundToDecimals(1.005, 2) is 1.01 and not the 1 of the binary value
// 1.00499999999999989342. Format rounds the same way.
//
// NaN and ±Inf are returned unchanged, and so is a value that has no more
// decimal places than requested. Only a rounded value beyond the range of T,
// like math.MaxFloat64 rounded to a multiple of 10^307, is ±Inf.
func RoundToDecimals[T ~float32 | ~float64](f T, decimals int) T {
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return f
	}
	bitSize := reflect.TypeFor[T]().Bits()
	if bitSize == 64 {
		if scaled, ok := roundScaledFast(float64(f), decimals); ok {
			// A value with no more decimals than requested scales to an
			// integer, so this is f unchanged like in roundToDecimalsExact
			return T(unscale(scaled, decimals))
		}
	}
	return T(roundToDecimalsExact(float64(f), bitSize, decimals))
}

// roundToDecimalsExact is RoundToDecimals for a finite f of the passed
// bitSize, rounding the digits of its shortest decimal representation.
func roundToDecimalsExact(f float64, bitSize, decimals int) float64 {
	intDigits, fracDigits, _ := strings.Cut(strconv.FormatFloat(math.Abs(f), 'f', -1, bitSize), ".")
	if len(fracDigits) <= decimals {
		return f // nothing to round
	}
	digits := roundDigitsHalfAwayFromZero(intDigits, fracDigits, decimals)
	if digits == "" {
		digits = "0"
	}
	str := digits + "e" + strconv.Itoa(-decimals)
	if math.Signbit(f) {
		str = "-" + str
	}
	// A result beyond the range of T parses as ±Inf with a range error,
	// which is the infinity that an overflowing float operation yields
	rounded, _ := strconv.ParseFloat(str, bitSize)
	return rounded
}

// roundScaledFast returns f·10^decimals rounded half away from zero in
// float64 arithmetic and whether that provably equals rounding the shortest
// decimal representation of f, which is what the digit rounding of
// roundToDecimalsExact and roundFormatted does. The float64 operations are an
// order of magnitude faster than formatting and parsing the digits.
//
// The shortest decimal of f lies within half an ulp of f, so within a factor
// of 1±2^-53, and scaling by the exactly representable 10^|decimals| ≤ 10^22
// rounds by at most another such factor. The scaled float64 therefore differs
// from the scaled shortest decimal by less than |scaled|·2^-51, the bound
// under which internal.FastRound proves its rounding. The relative bounds
// don't hold for subnormal floats, but those scale to far below 1/2 and
// round to zero either way.
func roundScaledFast(f float64, decimals int) (rounded float64, ok bool) {
	var scaled float64
	switch {
	case decimals >= 0 && decimals <= 22:
		scaled = f * math.Pow10(decimals)
	case decimals < 0 && decimals >= -22:
		scaled = f / math.Pow10(-decimals)
	default:
		return 0, false
	}
	return internal.FastRound(scaled)
}

// unscale returns the integer rounded from roundScaledFast divided by
// 10^decimals. Both operands are exact float64 values, so the IEEE 754
// division or multiplication is correctly rounded and yields the same
// float64 as parsing the decimal digits of rounded with exponent -decimals.
func unscale(rounded float64, decimals int) float64 {
	if decimals >= 0 {
		return rounded / math.Pow10(decimals)
	}
	return rounded * math.Pow10(-decimals)
}

// roundDigitsHalfAwayFromZero rounds the non-negative decimal number with the
// integer digits intDigits and the fractional digits fracDigits half away from
// zero to the passed number of decimal places, where negative decimals round
// to tens, hundreds, and so on. It returns the digits of the rounded number
// times 10^decimals, which may have leading zeros and are empty for zero.
// Rounding the digits instead of a binary float keeps ties like 1.005 exact.
func roundDigitsHalfAwayFromZero(intDigits, fracDigits string, decimals int) string {
	digits := intDigits + fracDigits
	keep := len(intDigits) + decimals // number of leading digits to keep
	if keep >= len(digits) {
		return digits + strings.Repeat("0", keep-len(digits))
	}
	if keep < 0 {
		// The rounding unit is more than ten times the number
		return ""
	}
	// Half away from zero only depends on the first cut digit
	kept := digits[:keep]
	if digits[keep] < '5' {
		return kept
	}
	// Add one to the last kept digit, carrying over nines
	b := []byte(kept)
	i := len(b) - 1
	for ; i >= 0 && b[i] == '9'; i-- {
		b[i] = '0'
	}
	if i < 0 {
		return "1" + string(b)
	}
	b[i]++
	return string(b)
}

// DerefOr dereferences ptr or returns defaultVal if ptr is nil.
// Provides safe pointer dereferencing for float types.
func DerefOr[T ~float32 | ~float64](ptr *T, defaultVal T) T {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Valid returns if the passed float is neither infinite nor NaN.
// Useful for validating float values before performing calculations.
func Valid[T ~float32 | ~float64](f T) bool {
	return !math.IsNaN(float64(f)) && !math.IsInf(float64(f), 0)
}

// ValidAndHasSign returns true if Valid(f) and if it has the same sign
// as the passed non-zero int argument.
// If 0 is passed as sign then the sign check always returns true.
func ValidAndHasSign[T ~float32 | ~float64](f T, sign int) bool {
	return Valid(f) && (sign == 0 || (f < 0) == (sign < 0))
}
