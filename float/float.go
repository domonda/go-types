package float

import "math"

// RoundToDecimals returns the float64 rounded
// to the passed number of decimal places.
func RoundToDecimals(f float64, decimals int) float64 {
	pow := math.Pow10(decimals)
	return math.Round(f*pow) / pow
}

// DerefOr dereferences ptr or returns defaultVal if ptr is nil
func DerefOr(ptr *float64, defaultVal float64) float64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Valid returns if the passed float64 is neither infinite nor NaN
func Valid(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

// ValidAndHasSign returns if Valid(f) and
// if it has the same sign than the passed int argument
// or any sign if 0 is passed.
func ValidAndHasSign(f float64, sign int) bool {
	if !Valid(f) {
		return false
	}
	switch {
	case sign > 0:
		return f > 0
	case sign < 0:
		return f < 0
	default:
		return true
	}
}
