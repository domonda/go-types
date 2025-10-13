// Package float provides utilities for working with floating-point numbers in Go.
// It includes functions for rounding, validation, and safe pointer dereferencing
// for both float32 and float64 types using Go generics.
//
// The package includes:
// - Decimal rounding utilities
// - Float validation (NaN and infinity checks)
// - Sign validation
// - Safe pointer dereferencing with default values
package float

import "math"

// RoundToDecimals returns the float rounded to the passed number of decimal places.
// Uses math.Round with power-of-10 multiplication for precise decimal rounding.
func RoundToDecimals[T ~float32 | ~float64](f T, decimals int) T {
	pow := math.Pow10(decimals)
	return T(math.Round(float64(f)*pow) / pow)
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
