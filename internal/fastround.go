package internal

import "math"

// FastRound returns x rounded half away from zero to an integer and whether
// that provably equals rounding the exact value that x approximates, given
// that x differs from the exact value by less than |x|·2^-51. Callers get
// that bound from a few float64 roundings by a factor of at most 1±2^-53
// each, and fall back to an order of magnitude slower exact computation,
// like rounding decimal digits or big.Rat arithmetic, if ok is false.
// The relative bound doesn't hold for subnormal floats, so a caller has to
// show that they round the same way or never make ok true.
//
// If the distance of x to the nearest half-integer exceeds the four times
// larger |x|·2^-49, x and the exact value lie strictly between the same two
// half-integers and round to the same integer. The distance x-rounded is
// exact (Sterbenz lemma), and 0.5 minus it is either exact or above 1/4,
// which exceeds the error below 2^48, the only range where the condition
// can hold. It never holds for NaN and ±Inf. If Go fuses the final
// multiplication that computed x into the distance, the distance is
// computed from the exact product and rounded once by at most 2^-54
// instead, which is far below the bound for a nonzero rounded and changes
// nothing for a zero one.
func FastRound(x float64) (rounded float64, ok bool) {
	rounded = math.Round(x)
	return rounded, 0.5-math.Abs(x-rounded) > math.Abs(x)*0x1p-49
}
