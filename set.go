package types

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"

	"github.com/domonda/go-types/mapset"
)

// Set represents a collection of unique ordered values.
// It is implemented as a map[T]struct{} for efficient lookups and memory usage.
// The type parameter T must be comparable and ordered (implements cmp.Ordered).
type Set[T cmp.Ordered] map[T]struct{}

// NewSet creates a new Set containing the provided values.
// Duplicate values are automatically deduplicated.
func NewSet[T cmp.Ordered](vals ...T) Set[T] {
	set := make(Set[T], len(vals))
	for _, val := range vals {
		set[val] = struct{}{}
	}
	return set
}

// Sorted returns a slice containing all values in the set, sorted in ascending order.
func (set Set[T]) Sorted() []T {
	return SetToSortedSlice(set)
}

// GetOne returns one element of the set or the default value for T if the set is empty.
// The element returned is not guaranteed to be any specific element from the set.
func (set Set[T]) GetOne() T {
	for val := range set {
		return val
	}
	return *new(T)
}

// Add adds a value to the set.
//
// Deprecated: use [Set.Insert] which additionally reports
// whether the set was changed.
func (set Set[T]) Add(val T) {
	set[val] = struct{}{}
}

// AddSlice adds all values from the slice to the set.
//
// Deprecated: use set.InsertAll(slices.Values(vals)).
func (set Set[T]) AddSlice(vals []T) {
	for _, val := range vals {
		set[val] = struct{}{}
	}
}

// AddSet adds all values from another set to this set.
//
// Deprecated: use [Set.UnionWith].
func (set Set[T]) AddSet(other Set[T]) {
	for val := range other {
		set[val] = struct{}{}
	}
}

// Insert adds val to the set and reports whether the set was changed.
// It panics if the set is nil and val is not already an element.
func (set Set[T]) Insert(val T) bool {
	return mapset.Insert(set, val)
}

// InsertAll adds all values yielded by seq to the set
// and reports whether the set was changed.
// It panics if the set is nil and seq yields a value
// that is not already an element.
func (set Set[T]) InsertAll(seq iter.Seq[T]) bool {
	return mapset.InsertAll(set, seq)
}

// All returns an iterator over the values of the set in undefined order.
// It is valid to call this method on a nil set.
func (set Set[T]) All() iter.Seq[T] {
	return mapset.All(set)
}

// Contains returns true if the set contains the specified value.
// It is valid to call this method on a nil set.
func (set Set[T]) Contains(val T) bool {
	return mapset.Contains(set, val)
}

// ContainsAny returns true if the set contains any of the specified values.
func (set Set[T]) ContainsAny(vals ...T) bool {
	return slices.ContainsFunc(vals, set.Contains)
}

// ContainsAll returns true if the set contains all values yielded by seq.
// It returns true for an empty sequence and is valid to call on a nil set.
func (set Set[T]) ContainsAll(seq iter.Seq[T]) bool {
	return mapset.ContainsAll(set, seq)
}

// ContainsSet returns true if this set contains all values from the other set.
func (set Set[T]) ContainsSet(other Set[T]) bool {
	return mapset.ContainsAll(set, other.All())
}

// Delete removes a value from the set and reports whether the set was changed.
// It is valid to call this method on a nil set.
func (set Set[T]) Delete(val T) bool {
	return mapset.Delete(set, val)
}

// DeleteAll removes all values yielded by seq from the set
// and reports whether the set was changed.
// It is valid to call this method on a nil set.
func (set Set[T]) DeleteAll(seq iter.Seq[T]) bool {
	return mapset.DeleteAll(set, seq)
}

// DeleteFunc removes all values for which del returns true
// and reports whether the set was changed.
// It is valid to call this method on a nil set.
func (set Set[T]) DeleteFunc(del func(T) bool) bool {
	return mapset.DeleteFunc(set, del)
}

// DeleteSlice removes all values from the slice from the set.
//
// Deprecated: use set.DeleteAll(slices.Values(vals)).
func (set Set[T]) DeleteSlice(vals []T) {
	for _, val := range vals {
		delete(set, val)
	}
}

// DeleteSet removes all values from the other set from this set.
//
// Deprecated: use [Set.DifferenceWith].
func (set Set[T]) DeleteSet(other Set[T]) {
	for str := range other {
		delete(set, str)
	}
}

// Clear removes all values from the set.
// It is valid to call this method on a nil set.
func (set Set[T]) Clear() {
	clear(set)
}

// Clone creates a deep copy of the set.
func (set Set[T]) Clone() Set[T] {
	if set == nil {
		return nil
	}
	return maps.Clone(set)
}

// Union returns a new set containing all values from both this set and the other set.
func (set Set[T]) Union(other Set[T]) Set[T] {
	return mapset.Union(set, other)
}

// UnionWith adds all values of the other set to this set.
// It panics if this set is nil and the other set has a value
// that is not already an element.
func (set Set[T]) UnionWith(other Set[T]) {
	mapset.UnionWith(set, other)
}

// Intersection returns a new set containing only values that exist in both sets.
func (set Set[T]) Intersection(other Set[T]) Set[T] {
	return mapset.Intersection(set, other)
}

// IntersectionWith removes all values from this set
// that are not also in the other set.
// It is valid to call this method on a nil set.
func (set Set[T]) IntersectionWith(other Set[T]) {
	mapset.IntersectionWith(set, other)
}

// Intersects reports whether this set and the other set
// have at least one value in common.
func (set Set[T]) Intersects(other Set[T]) bool {
	return mapset.Intersects(set, other)
}

// Difference returns a new set containing the values of this set
// that are not in the other set.
//
// Note that this is the asymmetric difference as defined by the Go
// collections proposal. Up to and including go-types v0 pseudo-versions
// before 2026-08 this method returned the symmetric difference, which is
// now [Set.SymmetricDifference].
func (set Set[T]) Difference(other Set[T]) Set[T] {
	return mapset.Difference(set, other)
}

// DifferenceWith removes all values of the other set from this set.
// It is valid to call this method on a nil set.
func (set Set[T]) DifferenceWith(other Set[T]) {
	mapset.DifferenceWith(set, other)
}

// SymmetricDifference returns a new set containing the values
// that exist in either set but not in both.
func (set Set[T]) SymmetricDifference(other Set[T]) Set[T] {
	return mapset.SymmetricDifference(set, other)
}

// SymmetricDifferenceWith replaces the values of this set with the values
// that exist in either set but not in both.
// It panics if this set is nil and the other set has a value
// that is not already an element.
func (set Set[T]) SymmetricDifferenceWith(other Set[T]) {
	mapset.SymmetricDifferenceWith(set, other)
}

// Map applies a transformation function to each value in the set and returns a new set
// containing the results. The mapFunc should return the transformed value and a boolean
// indicating whether to include the result in the new set.
func (set Set[T]) Map(mapFunc func(T) (T, bool)) Set[T] {
	result := make(Set[T], len(set))
	for val := range set {
		if mappedVal, ok := mapFunc(val); ok {
			result.Insert(mappedVal)
		}
	}
	return result
}

// Equal returns true if both sets contain exactly the same values.
func (set Set[T]) Equal(other Set[T]) bool {
	return mapset.Equal(set, other)
}

// Len returns the number of values in the set.
// Returns zero for a nil set.
func (set Set[T]) Len() int {
	return len(set)
}

// IsEmpty returns true if the set is empty or nil.
func (set Set[T]) IsEmpty() bool {
	return len(set) == 0
}

// IsNull implements the nullable.Nullable interface by returning true if the set is nil.
func (set Set[T]) IsNull() bool {
	return set == nil
}

// String implements the fmt.Stringer interface.
// Returns a string representation of the set with values in sorted order.
func (set Set[T]) String() string { //#nosec
	if set == nil {
		return "<nil>"
	}
	var b strings.Builder
	b.WriteByte('[')
	for i, val := range set.Sorted() {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%#v", val)
	}
	b.WriteByte(']')
	return b.String()
}

// MarshalJSON implements encoding/json.Marshaler by returning the JSON null value for an empty (null) string.
func (set Set[T]) MarshalJSON() ([]byte, error) {
	if set.IsNull() {
		return []byte(`null`), nil
	}
	// Sorted returns a nil slice for an empty set, which marshals as null,
	// so the empty array has to be written here to keep null and the empty
	// array distinct.
	if len(set) == 0 {
		return []byte(`[]`), nil
	}
	return json.Marshal(set.Sorted())
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// It unmarshals a JSON array into the set, removing any duplicates.
// JSON null assigns a nil set and the empty array an allocated empty
// one, mirroring [Set.MarshalJSON]. A nil set panics when inserted
// into, exactly like a nil Go map.
func (set *Set[T]) UnmarshalJSON(j []byte) error {
	if bytes.Equal(j, []byte(`null`)) {
		// JSON null is the nil set, symmetric with MarshalJSON, and the
		// empty array is the allocated empty set. Like the nil map that
		// a SQL NULL scans to, the result must not be inserted into
		// without a nil check.
		*set = nil
		return nil
	}
	var slice []T
	err := json.Unmarshal(j, &slice)
	if err != nil {
		return fmt.Errorf("can't unmarshall %T from JSON: %w", *set, err)
	}
	if *set == nil {
		*set = NewSet(slice...)
	} else {
		set.Clear()
		set.InsertAll(slices.Values(slice))
	}
	return nil
}

// ReduceSet applies a reduction function to all values in the set and returns the result.
// The reduceFunc is called for each value in the set, with the accumulated result and the current value.
func ReduceSet[S ~map[T]struct{}, T cmp.Ordered, R any](set S, reduceFunc func(last R, val T) R) (result R) {
	for val := range set {
		result = reduceFunc(result, val)
	}
	return result
}

// ReduceSlice applies a reduction function to all values in the slice and returns the result.
// The reduceFunc is called for each value in the slice, with the accumulated result and the current value.
func ReduceSlice[S ~[]T, T cmp.Ordered, R any](slice S, reduceFunc func(last R, val T) R) (result R) {
	for _, val := range slice {
		result = reduceFunc(result, val)
	}
	return result
}

// SetToRandomizedSlice converts a set to a slice with values in random order.
func SetToRandomizedSlice[S ~map[T]struct{}, T cmp.Ordered](set S) []T {
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

// SetToSortedSlice converts a set to a slice with values in sorted order.
func SetToSortedSlice[S ~map[T]struct{}, T cmp.Ordered](set S) []T {
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
	slices.Sort(slice)
	return slice
}
