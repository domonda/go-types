package float

import "math"

// RoundToDecimals returns the float rounded
// to the passed number of decimal places.
func RoundToDecimals[T float32 | float64](f T, decimals int) T {
	pow := math.Pow10(decimals)
	return T(math.Round(float64(f)*pow) / pow)
}

// DerefOr dereferences ptr or returns defaultVal if ptr is nil
func DerefOr[T float32 | float64](ptr *T, defaultVal T) T {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Valid returns if the passed float is neither infinite nor NaN
func Valid[T float32 | float64](f T) bool {
	return !math.IsNaN(float64(f)) && !math.IsInf(float64(f), 0)
}

// ValidAndHasSign returns true if Valid(f) and
// if it has the same sign than the passed non zero int argument.
// If 0 is passed as sign then the sign check always returns true.
func ValidAndHasSign[T float32 | float64](f T, sign int) bool {
	if !Valid(f) {
		return false
	}
	switch {
	case sign > 0:
		return f > 0
	case sign < 0:
		return f < 0
	}
	return true
}
