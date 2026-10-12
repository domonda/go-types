// Package float provides utilities for working with floating-point numbers in Go.
// It includes functions for rounding, validation, and safe pointer dereferencing
// for both float32 and float64 types using Go generics.
//
// The package includes:
//   - Decimal rounding utilities
//   - Float validation (NaN and infinity checks)
//   - Sign validation
//   - Safe pointer dereferencing with default values
package float

import (
	"math"
	"reflect"
	"strconv"
	"strings"
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
	intDigits, fracDigits, _ := strings.Cut(strconv.FormatFloat(math.Abs(float64(f)), 'f', -1, bitSize), ".")
	if len(fracDigits) <= decimals {
		return f // nothing to round
	}
	digits := roundDigitsHalfAwayFromZero(intDigits, fracDigits, decimals)
	if digits == "" {
		digits = "0"
	}
	str := digits + "e" + strconv.Itoa(-decimals)
	if math.Signbit(float64(f)) {
		str = "-" + str
	}
	// A result beyond the range of T parses as ±Inf with a range error,
	// which is the infinity that an overflowing float operation yields
	rounded, _ := strconv.ParseFloat(str, bitSize)
	return T(rounded)
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
