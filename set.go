package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Set[T constraints.Ordered] map[T]struct{}

func NewSet[T constraints.Ordered](vals ...T) Set[T] {
	set := make(Set[T], len(vals))
	for _, val := range vals {
		set[val] = struct{}{}
	}
	return set
}

func (set Set[T]) Len() int {
	return len(set)
}

// IsNull implements the nullable.Nullable interface
// by returning true if the set is nil.
func (set Set[T]) IsNull() bool {
	return set == nil
}

// String implements the fmt.Stringer interface.
func (set Set[T]) String() string { //#nosec
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

func (set Set[T]) Sorted() []T {
	return SetToSortedSlice(set)
}

func (set Set[T]) Add(val T) {
	set[val] = struct{}{}
}

func (set Set[T]) AddSlice(vals []T) {
	for _, val := range vals {
		set[val] = struct{}{}
	}
}

func (set Set[T]) AddSet(other Set[T]) {
	for val := range other {
		set[val] = struct{}{}
	}
}

func (set Set[T]) Contains(val T) bool {
	_, ok := set[val]
	return ok
}

func (set Set[T]) ContainsAny(vals ...T) bool {
	for _, val := range vals {
		if set.Contains(val) {
			return true
		}
	}
	return false
}

func (set Set[T]) ContainsAll(vals ...T) bool {
	for _, val := range vals {
		if !set.Contains(val) {
			return false
		}
	}
	return true
}

func (set Set[T]) ContainsSet(other Set[T]) bool {
	for val := range other {
		if !set.Contains(val) {
			return false
		}
	}
	return true
}

func (set Set[T]) Delete(val T) {
	delete(set, val)
}

func (set Set[T]) DeleteSlice(vals []T) {
	for _, val := range vals {
		delete(set, val)
	}
}

func (set Set[T]) DeleteSet(other Set[T]) {
	for str := range other {
		delete(set, str)
	}
}

func (set Set[T]) Clear() {
	maps.Clear(set)
}

func (set Set[T]) Clone() Set[T] {
	return maps.Clone(set)
}

func (set Set[T]) Union(other Set[T]) Set[T] {
	union := make(Set[T], (len(set)+len(other))/2)
	for val := range set {
		union.Add(val)
	}
	for val := range other {
		union.Add(val)
	}
	return union
}

func (set Set[T]) Intersection(other Set[T]) Set[T] {
	inter := make(Set[T], (len(set)+len(other))/2)
	for val := range set {
		if other.Contains(val) {
			inter.Add(val)
		}
	}
	return inter
}

func (set Set[T]) Difference(other Set[T]) Set[T] {
	diff := make(Set[T])
	for val := range set {
		if !other.Contains(val) {
			diff.Add(val)
		}
	}
	for val := range other {
		if !set.Contains(val) {
			diff.Add(val)
		}
	}
	return diff
}

func (set Set[T]) Equal(other Set[T]) bool {
	return maps.Equal(set, other)
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value for an empty (null) string.
func (set Set[T]) MarshalJSON() ([]byte, error) {
	if set.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(set.Sorted())
}

// UnmarshalJSON implements encoding/json.Unmarshaler
func (set *Set[T]) UnmarshalJSON(j []byte) error {
	if bytes.Equal(j, []byte(`null`)) {
		*set = nil
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
		set.AddSlice(slice)
	}
	return nil
}

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
	slices.Sort(slice)
	return slice
}
