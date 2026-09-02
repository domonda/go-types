// Package mapset defines operations on sets represented either as
// map[K]struct{}, or as map[K]bool where the boolean value is ignored.
//
// It mirrors the API of the container/mapset package proposed for Go 1.28
// (https://go.dev/issue/77052, CL 724420) function for function, so that
// once that package ships this one can be replaced by changing the import
// path. The set types of this module (types.Set, uu.IDSet,
// strutil.StringSet and email.AddressSet) implement the methods of the
// abstract set interface of https://go.dev/issue/80590 by delegating here.
//
// # Nil sets
//
// A nil set is a valid empty set for every operation that does not have to
// store an element: Contains, ContainsAll, All, Equal, Intersects,
// String, Delete, DeleteAll, DeleteFunc, IntersectionWith and DifferenceWith
// all accept a nil set. Union, Intersection, Difference and
// SymmetricDifference accept nil operands and always return a newly
// allocated set.
//
// Insert, InsertAll, UnionWith and SymmetricDifferenceWith panic on a nil
// set if they actually have to store an element, exactly like an assignment
// to a nil Go map.
package mapset

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

// present returns the value stored for a key that is a member of the set:
// struct{}{} for map[K]struct{} sets and true for map[K]bool sets.
func present[V bool | struct{}]() V {
	var v V
	if b, ok := any(&v).(*bool); ok {
		*b = true
	}
	return v
}

// Collect returns a new set containing the elements of seq.
func Collect[K comparable](seq iter.Seq[K]) map[K]struct{} {
	s := make(map[K]struct{})
	for k := range seq {
		s[k] = struct{}{}
	}
	return s
}

// CollectBool returns a new map[K]bool set containing the elements of seq.
func CollectBool[K comparable](seq iter.Seq[K]) map[K]bool {
	s := make(map[K]bool)
	for k := range seq {
		s[k] = true
	}
	return s
}

// Of returns a new set containing elems.
func Of[K comparable](elems ...K) map[K]struct{} {
	s := make(map[K]struct{}, len(elems))
	for _, k := range elems {
		s[k] = struct{}{}
	}
	return s
}

// OfBool returns a new map[K]bool set containing elems.
func OfBool[K comparable](elems ...K) map[K]bool {
	s := make(map[K]bool, len(elems))
	for _, k := range elems {
		s[k] = true
	}
	return s
}

// String formats x as {elem1 elem2 ...} with the elements formatted
// using fmt and sorted by their formatted representation.
//
// The Go proposal does not pin this format down, so it is the one
// function of this package whose output may differ from the standard
// library version. None of the set types of this module use it; they all
// implement their own String method with a format of their own.
func String[M ~map[K]V, K comparable, V bool | struct{}](x M) string {
	elems := make([]string, 0, len(x))
	for k := range x {
		elems = append(elems, fmt.Sprint(k))
	}
	slices.Sort(elems)
	return "{" + strings.Join(elems, " ") + "}"
}

// Equal reports whether x and y contain the same elements.
func Equal[M ~map[K]V, K comparable, V bool | struct{}](x, y M) bool {
	if len(x) != len(y) {
		return false
	}
	for k := range x {
		if _, ok := y[k]; !ok {
			return false
		}
	}
	return true
}

// Contains reports whether k is an element of x.
func Contains[M ~map[K]V, K comparable, V bool | struct{}](x M, k K) bool {
	_, ok := x[k]
	return ok
}

// ContainsAll reports whether every element of keys is an element of x.
// It returns true for an empty sequence.
func ContainsAll[M ~map[K]V, K comparable, V bool | struct{}](x M, keys iter.Seq[K]) bool {
	for k := range keys {
		if _, ok := x[k]; !ok {
			return false
		}
	}
	return true
}

// All returns an iterator over the elements of x in undefined order.
func All[M ~map[K]V, K comparable, V bool | struct{}](x M) iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range x {
			if !yield(k) {
				return
			}
		}
	}
}

// Intersects reports whether x and y have at least one element in common.
func Intersects[MX, MY ~map[K]V, K comparable, V bool | struct{}](x MX, y MY) bool {
	// Iterate the smaller set and look up in the larger one.
	if len(x) > len(y) {
		for k := range y {
			if _, ok := x[k]; ok {
				return true
			}
		}
		return false
	}
	for k := range x {
		if _, ok := y[k]; ok {
			return true
		}
	}
	return false
}

// Union returns a new set with the elements of both x and y.
func Union[MX ~map[K]VX, MY ~map[K]VY, K comparable, VX, VY bool | struct{}](x MX, y MY) MX {
	union := make(MX, len(x)+len(y))
	for k := range x {
		union[k] = present[VX]()
	}
	for k := range y {
		union[k] = present[VX]()
	}
	return union
}

// Intersection returns a new set with the elements that are in both x and y.
func Intersection[MX ~map[K]VX, MY ~map[K]VY, K comparable, VX, VY bool | struct{}](x MX, y MY) MX {
	intersection := make(MX)
	for k := range x {
		if _, ok := y[k]; ok {
			intersection[k] = present[VX]()
		}
	}
	return intersection
}

// Difference returns a new set with the elements of x that are not in y.
func Difference[MX ~map[K]VX, MY ~map[K]VY, K comparable, VX, VY bool | struct{}](x MX, y MY) MX {
	difference := make(MX)
	for k := range x {
		if _, ok := y[k]; !ok {
			difference[k] = present[VX]()
		}
	}
	return difference
}

// SymmetricDifference returns a new set with the elements
// that are in exactly one of x and y.
func SymmetricDifference[MX ~map[K]VX, MY ~map[K]VY, K comparable, VX, VY bool | struct{}](x MX, y MY) MX {
	difference := make(MX)
	for k := range x {
		if _, ok := y[k]; !ok {
			difference[k] = present[VX]()
		}
	}
	for k := range y {
		if _, ok := x[k]; !ok {
			difference[k] = present[VX]()
		}
	}
	return difference
}

// Insert adds elem to x and reports whether the set changed.
// It panics if x is nil and elem is not already an element.
func Insert[M ~map[K]V, K comparable, V bool | struct{}](x M, elem K) bool {
	if _, ok := x[elem]; ok {
		return false
	}
	x[elem] = present[V]()
	return true
}

// InsertAll adds all elements of addenda to x and reports whether the set changed.
// It panics if x is nil and addenda yields an element that is not already an element.
func InsertAll[M ~map[K]V, K comparable, V bool | struct{}](x M, addenda iter.Seq[K]) bool {
	changed := false
	for k := range addenda {
		if _, ok := x[k]; !ok {
			x[k] = present[V]()
			changed = true
		}
	}
	return changed
}

// Delete removes k from x and reports whether the set changed.
func Delete[M ~map[K]V, K comparable, V bool | struct{}](x M, k K) bool {
	if _, ok := x[k]; !ok {
		return false
	}
	delete(x, k)
	return true
}

// DeleteAll removes all elements of delenda from x and reports whether the set changed.
func DeleteAll[M ~map[K]V, K comparable, V bool | struct{}](x M, delenda iter.Seq[K]) bool {
	changed := false
	for k := range delenda {
		if _, ok := x[k]; ok {
			delete(x, k)
			changed = true
		}
	}
	return changed
}

// DeleteFunc removes every element of x for which f returns true
// and reports whether the set changed.
func DeleteFunc[M ~map[K]V, K comparable, V bool | struct{}](x M, f func(K) bool) bool {
	changed := false
	for k := range x {
		if f(k) {
			delete(x, k)
			changed = true
		}
	}
	return changed
}

// UnionWith adds all elements of y to x.
// It panics if x is nil and y has an element that is not already an element of x.
func UnionWith[MX ~map[K]VX, MY ~map[K]VY, K comparable, VX, VY bool | struct{}](x MX, y MY) {
	for k := range y {
		if _, ok := x[k]; !ok {
			x[k] = present[VX]()
		}
	}
}

// IntersectionWith removes every element of x that is not an element of y.
func IntersectionWith[M ~map[K]V, K comparable, V bool | struct{}](x, y M) {
	for k := range x {
		if _, ok := y[k]; !ok {
			delete(x, k)
		}
	}
}

// DifferenceWith removes every element of y from x.
func DifferenceWith[M ~map[K]V, K comparable, V bool | struct{}](x, y M) {
	for k := range y {
		delete(x, k)
	}
}

// SymmetricDifferenceWith replaces the elements of x with the elements
// that are in exactly one of x and y.
// It panics if x is nil and y has an element that is not already an element of x.
func SymmetricDifferenceWith[M ~map[K]V, K comparable, V bool | struct{}](x, y M) {
	for k := range y {
		if _, ok := x[k]; ok {
			delete(x, k)
		} else {
			x[k] = present[V]()
		}
	}
}
