package strutil

import (
	"maps"
	"sort"
	"strings"
)

// StringSet is a set of unique strings implemented as a map.
type StringSet map[string]struct{}

// NewStringSet creates and returns a new StringSet containing the provided strings.
func NewStringSet(strings ...string) StringSet {
	set := make(StringSet, len(strings))
	for _, s := range strings {
		set[s] = struct{}{}
	}
	return set
}

// NewStringSetMergeSlices creates and returns a new StringSet containing all unique strings
// from the provided slices.
func NewStringSetMergeSlices(slices ...[]string) StringSet {
	cap := 0
	for _, strings := range slices {
		cap += len(strings)
	}
	set := make(StringSet, cap)
	for _, strings := range slices {
		for _, s := range strings {
			set[s] = struct{}{}
		}
	}
	return set
}

// Sorted returns all strings in the set as a sorted slice.
func (set StringSet) Sorted() (s []string) {
	if count := len(set); count > 0 {
		s = make([]string, 0, count)
		for str := range set {
			s = append(s, str)
		}
		sort.Strings(s)
	}
	return s
}

// String implements the fmt.Stringer interface.
func (set StringSet) String() string {
	if len(set) == 0 {
		return `[]`
	}
	return `["` + strings.Join(set.Sorted(), `", "`) + `"]`
}

// AddSlice adds all strings from the slice s to the set.
func (set StringSet) AddSlice(s []string) {
	for _, str := range s {
		set[str] = struct{}{}
	}
}

// AddSet adds all strings from the other set to this set.
func (set StringSet) AddSet(other StringSet) {
	for str := range other {
		set[str] = struct{}{}
	}
}

// Add adds str to the set.
func (set StringSet) Add(str string) {
	set[str] = struct{}{}
}

// Contains returns true if str is in the set.
func (set StringSet) Contains(str string) bool {
	_, has := set[str]
	return has
}

// ContainsAny returns true if any of the provided strings are in the set.
func (set StringSet) ContainsAny(strs ...string) bool {
	for _, str := range strs {
		if set.Contains(str) {
			return true
		}
	}
	return false
}

// StringContainsAnyOfSet returns true if the passed string
// contains any of the strings of the StringSet.
func (set StringSet) StringContainsAnyOfSet(str string) bool {
	for s := range set {
		if strings.Contains(str, s) {
			return true
		}
	}
	return false
}

// Delete removes str from the set.
func (set StringSet) Delete(str string) {
	delete(set, str)
}

// Clear removes all strings from the set.
func (set StringSet) Clear() {
	clear(set)
}

// DeleteSlice removes all strings in the slice s from the set.
func (set StringSet) DeleteSlice(s []string) {
	for _, str := range s {
		delete(set, str)
	}
}

// DeleteSet removes all strings in the other set from this set.
func (set StringSet) DeleteSet(other StringSet) {
	for str := range other {
		delete(set, str)
	}
}

// Clone returns a deep copy of the set. Returns nil if the set is nil.
func (set StringSet) Clone() StringSet {
	if set == nil {
		return nil
	}
	return maps.Clone(set)
}

// Diff returns a new StringSet containing strings that are in either set
// but not in both (symmetric difference).
func (set StringSet) Diff(other StringSet) StringSet {
	diff := make(StringSet, len(set))
	for str := range set {
		if !other.Contains(str) {
			diff.Add(str)
		}
	}
	for str := range other {
		if !set.Contains(str) {
			diff.Add(str)
		}
	}
	return diff
}

// Equal returns true if set and other contain exactly the same strings.
func (set StringSet) Equal(other StringSet) bool {
	if len(set) != len(other) {
		return false
	}
	for str := range set {
		if !other.Contains(str) {
			return false
		}
	}
	return true
}
