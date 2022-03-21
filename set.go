package types

import (
	"sort"

	"golang.org/x/exp/constraints"
)

func SetToRandomizedSlice[S ~map[T]struct{}, T constraints.Ordered](set S) []T {
	l := len(set)
	if l == 0 {
		return nil
	}
	slice := make([]T, l)
	i := 0
	for val := range set {
		slice[i] = val
		i++
	}
	return slice
}

func SetToSortedSlice[S ~map[T]struct{}, T constraints.Ordered](set S) []T {
	l := len(set)
	switch l {
	case 0:
		return nil
	case 1:
		for val := range set {
			return []T{val}
		}
	}
	slice := make([]T, l)
	i := 0
	for val := range set {
		slice[i] = val
		i++
	}
	sort.Slice(slice, func(i, j int) bool { return slice[i] < slice[j] })
	return slice
}
