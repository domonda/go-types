package set

import "cmp"

// New creates a new set with the passed values.
func New[T cmp.Ordered](values ...T) map[T]struct{} {
	set := make(map[T]struct{}, len(values))
	for _, val := range values {
		set[val] = struct{}{}
	}
	return set
}

// Add adds the passed values to the set and returns the set.
// If the passed set is nil, a new set is created.
func Add[T cmp.Ordered](set map[T]struct{}, values ...T) map[T]struct{} {
	if set == nil {
		set = make(map[T]struct{}, len(values))
	}
	for _, val := range values {
		set[val] = struct{}{}
	}
	return set
}

// Contains returns true if the set contains the passed value.
func Contains[T cmp.Ordered](set map[T]struct{}, value T) bool {
	_, ok := set[value]
	return ok
}

// ContainsAll returns true if the set contains all of the passed values.
func ContainsAll[T cmp.Ordered](set map[T]struct{}, values ...T) bool {
	for _, val := range values {
		if !Contains(set, val) {
			return false
		}
	}
	return true
}

// ContainsAny returns true if the set contains any of the passed values.
func ContainsAny[T cmp.Ordered](set map[T]struct{}, values ...T) bool {
	for _, val := range values {
		if Contains(set, val) {
			return true
		}
	}
	return false
}

// ContainsAllOther returns true if the set contains all of the values in the other set.
func ContainsAllOther[T cmp.Ordered](set map[T]struct{}, other map[T]struct{}) bool {
	for val := range other {
		if !Contains(set, val) {
			return false
		}
	}
	return true
}

// ContainsAnyOther returns true if the set contains any of the values in the other set.
func ContainsAnyOther[T cmp.Ordered](set map[T]struct{}, other map[T]struct{}) bool {
	for val := range other {
		if Contains(set, val) {
			return true
		}
	}
	return false
}
