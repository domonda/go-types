package types

import "golang.org/x/exp/constraints"

// Min returns the minimum of the passed values
// or the zero value of T if no values are passed.
func Min[T constraints.Ordered](vals ...T) T {
	l := len(vals)
	if l == 0 {
		var zero T
		return zero
	}
	min := vals[0]
	for i := 1; i < l; i++ {
		if val := vals[i]; val < min {
			min = val
		}
	}
	return min
}

// Max returns the maximum of the passed values
// or the zero value of T if no values are passed.
func Max[T constraints.Ordered](vals ...T) T {
	l := len(vals)
	if l == 0 {
		var zero T
		return zero
	}
	max := vals[0]
	for i := 1; i < l; i++ {
		if val := vals[i]; val > max {
			max = val
		}
	}
	return max
}
