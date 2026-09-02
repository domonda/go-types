package strutil

import (
	"iter"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/domonda/go-types/mapset"
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

// Len returns the number of strings in the set.
// It is valid to call this method on a nil StringSet.
func (set StringSet) Len() int {
	return len(set)
}

// IsEmpty returns true if the set is empty or nil.
func (set StringSet) IsEmpty() bool {
	return len(set) == 0
}

// AddSlice adds all strings from the slice s to the set.
//
// Deprecated: use set.InsertAll(slices.Values(s)).
func (set StringSet) AddSlice(s []string) {
	for _, str := range s {
		set[str] = struct{}{}
	}
}

// AddSet adds all strings from the other set to this set.
//
// Deprecated: use [StringSet.UnionWith].
func (set StringSet) AddSet(other StringSet) {
	for str := range other {
		set[str] = struct{}{}
	}
}

// Add adds str to the set.
//
// Deprecated: use [StringSet.Insert] which additionally reports
// whether the set was changed.
func (set StringSet) Add(str string) {
	set[str] = struct{}{}
}

// Insert adds str to the set and reports whether the set was changed.
// It panics if the set is nil and str is not already an element.
func (set StringSet) Insert(str string) bool {
	return mapset.Insert(set, str)
}

// InsertAll adds all strings yielded by seq to the set
// and reports whether the set was changed.
// It panics if the set is nil and seq yields a string
// that is not already an element.
func (set StringSet) InsertAll(seq iter.Seq[string]) bool {
	return mapset.InsertAll(set, seq)
}

// All returns an iterator over the strings of the set in undefined order.
// It is valid to call this method on a nil StringSet.
func (set StringSet) All() iter.Seq[string] {
	return mapset.All(set)
}

// Contains returns true if str is in the set.
// It is valid to call this method on a nil StringSet.
func (set StringSet) Contains(str string) bool {
	return mapset.Contains(set, str)
}

// ContainsAny returns true if any of the provided strings are in the set.
func (set StringSet) ContainsAny(strs ...string) bool {
	return slices.ContainsFunc(strs, set.Contains)
}

// ContainsAll reports whether the set contains all strings yielded by seq.
// It returns true for an empty sequence and is valid to call on a nil StringSet.
func (set StringSet) ContainsAll(seq iter.Seq[string]) bool {
	return mapset.ContainsAll(set, seq)
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

// Delete removes str from the set and reports whether the set was changed.
// It is valid to call this method on a nil StringSet.
func (set StringSet) Delete(str string) bool {
	return mapset.Delete(set, str)
}

// DeleteAll removes all strings yielded by seq from the set
// and reports whether the set was changed.
// It is valid to call this method on a nil StringSet.
func (set StringSet) DeleteAll(seq iter.Seq[string]) bool {
	return mapset.DeleteAll(set, seq)
}

// DeleteFunc removes every string for which del returns true
// and reports whether the set was changed.
// It is valid to call this method on a nil StringSet.
func (set StringSet) DeleteFunc(del func(string) bool) bool {
	return mapset.DeleteFunc(set, del)
}

// Clear removes all strings from the set.
// It is valid to call this method on a nil StringSet.
func (set StringSet) Clear() {
	clear(set)
}

// DeleteSlice removes all strings in the slice s from the set.
//
// Deprecated: use set.DeleteAll(slices.Values(s)).
func (set StringSet) DeleteSlice(s []string) {
	for _, str := range s {
		delete(set, str)
	}
}

// DeleteSet removes all strings in the other set from this set.
//
// Deprecated: use [StringSet.DifferenceWith].
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

// Union returns a new StringSet with all strings of set and other.
func (set StringSet) Union(other StringSet) StringSet {
	return mapset.Union(set, other)
}

// UnionWith adds all strings of other to set.
// It panics if set is nil and other has a string
// that is not already an element of set.
func (set StringSet) UnionWith(other StringSet) {
	mapset.UnionWith(set, other)
}

// Intersection returns a new StringSet with the strings
// that are in both set and other.
func (set StringSet) Intersection(other StringSet) StringSet {
	return mapset.Intersection(set, other)
}

// IntersectionWith removes every string from set that is not also in other.
// It is valid to call this method on a nil StringSet.
func (set StringSet) IntersectionWith(other StringSet) {
	mapset.IntersectionWith(set, other)
}

// Intersects reports whether set and other have at least one string in common.
func (set StringSet) Intersects(other StringSet) bool {
	return mapset.Intersects(set, other)
}

// Difference returns a new StringSet with the strings of set that are not in other.
//
// This replaces the former Diff method, which returned the symmetric
// difference and is now [StringSet.SymmetricDifference].
func (set StringSet) Difference(other StringSet) StringSet {
	return mapset.Difference(set, other)
}

// DifferenceWith removes every string of other from set.
// It is valid to call this method on a nil StringSet.
func (set StringSet) DifferenceWith(other StringSet) {
	mapset.DifferenceWith(set, other)
}

// SymmetricDifference returns a new StringSet containing strings
// that are in either set but not in both.
func (set StringSet) SymmetricDifference(other StringSet) StringSet {
	return mapset.SymmetricDifference(set, other)
}

// SymmetricDifferenceWith replaces the strings of set with the strings
// that are in exactly one of set and other.
// It panics if set is nil and other has a string
// that is not already an element of set.
func (set StringSet) SymmetricDifferenceWith(other StringSet) {
	mapset.SymmetricDifferenceWith(set, other)
}

// Equal returns true if set and other contain exactly the same strings.
func (set StringSet) Equal(other StringSet) bool {
	return mapset.Equal(set, other)
}
