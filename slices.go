package types

import "slices"

// SliceContainsAll returns true if outer contains all elements of inner.
func SliceContainsAll[T comparable](outer []T, inner ...T) bool {
	if len(outer) < len(inner) {
		return false
	}
	for _, innerElem := range inner {
		if !slices.Contains(outer, innerElem) {
			return false
		}
	}
	return true
}

// SliceContainsAny returns true if outer contains any element of inner.
func SliceContainsAny[T comparable](outer []T, inner ...T) bool {
	for _, innerElem := range inner {
		if slices.Contains(outer, innerElem) {
			return true
		}
	}
	return false
}
