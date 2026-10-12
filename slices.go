package types

import "slices"

// SliceContainsAll returns true if outer contains all elements of inner.
// It returns false if any element in inner is not found in outer.
// Duplicate elements in inner don't have to occur multiple times in outer,
// so inner can have more elements than outer.
func SliceContainsAll[T comparable](outer []T, inner ...T) bool {
	for _, innerElem := range inner {
		if !slices.Contains(outer, innerElem) {
			return false
		}
	}
	return true
}

// SliceContainsAny returns true if outer contains any element of inner.
// It returns true as soon as the first element from inner is found in outer,
// false if none of the elements in inner are found in outer.
func SliceContainsAny[T comparable](outer []T, inner ...T) bool {
	for _, innerElem := range inner {
		if slices.Contains(outer, innerElem) {
			return true
		}
	}
	return false
}
