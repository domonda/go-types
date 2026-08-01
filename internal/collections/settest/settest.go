// Package settest provides a reusable conformance test for the abstract
// set interface of the Go 1.28 collections proposal, https://go.dev/issue/80590.
//
// The set types of this module implement that interface by delegating to
// github.com/domonda/go-types/mapset, which is written to be replaceable by
// the standard library container/mapset package once it ships. Running this
// specification against every set type checks the behaviour of the methods
// themselves rather than the fact that they delegate, so that a behaviour
// change in the underlying package is caught at the type it affects.
//
// It is only imported from _test.go files.
package settest

import (
	"slices"
	"testing"

	"github.com/domonda/go-types/internal/collections"
)

// Run checks that S implements the abstract set specification for elements E.
//
// makeSet must return a new non-nil set containing exactly the passed elements,
// nilSet must be the nil value of S, and a, b and c must be three distinct
// elements of E.
//
// Run never inserts into nilSet: storing into a nil set is documented to panic
// like an assignment to a nil Go map.
func Run[E comparable, S collections.Set[E, S]](t *testing.T, makeSet func(elems ...E) S, nilSet S, a, b, c E) {
	t.Helper()

	// contains reports set membership through All, independently of Contains,
	// so that neither method is checked only against itself.
	contains := func(set S, elem E) bool {
		return slices.Contains(slices.Collect(set.All()), elem)
	}
	elems := func(set S) []E {
		return slices.Collect(set.All())
	}

	t.Run("Len", func(t *testing.T) {
		if got := makeSet(a, b).Len(); got != 2 {
			t.Errorf("Len() = %d, want 2", got)
		}
		if got := makeSet().Len(); got != 0 {
			t.Errorf("Len() of empty set = %d, want 0", got)
		}
		if got := nilSet.Len(); got != 0 {
			t.Errorf("Len() of nil set = %d, want 0", got)
		}
	})

	t.Run("Contains", func(t *testing.T) {
		set := makeSet(a, b)
		if !set.Contains(a) {
			t.Error("Contains(a) = false, want true")
		}
		if set.Contains(c) {
			t.Error("Contains(c) = true, want false")
		}
		// A nil set is a valid empty set for every read operation.
		if nilSet.Contains(a) {
			t.Error("Contains(a) on nil set = true, want false")
		}
	})

	t.Run("ContainsAll", func(t *testing.T) {
		set := makeSet(a, b)
		if !set.ContainsAll(slices.Values([]E{a, b})) {
			t.Error("ContainsAll([a b]) = false, want true")
		}
		if set.ContainsAll(slices.Values([]E{a, c})) {
			t.Error("ContainsAll([a c]) = true, want false")
		}
		if !set.ContainsAll(slices.Values([]E(nil))) {
			t.Error("ContainsAll([]) = false, want true")
		}
		if !nilSet.ContainsAll(slices.Values([]E(nil))) {
			t.Error("ContainsAll([]) on nil set = false, want true")
		}
		if nilSet.ContainsAll(slices.Values([]E{a})) {
			t.Error("ContainsAll([a]) on nil set = true, want false")
		}
	})

	t.Run("All", func(t *testing.T) {
		got := elems(makeSet(a, b))
		if len(got) != 2 || !slices.Contains(got, a) || !slices.Contains(got, b) {
			t.Errorf("All() = %v, want a and b in some order", got)
		}
		if got := elems(nilSet); len(got) != 0 {
			t.Errorf("All() on nil set = %v, want empty", got)
		}
		// A range loop over All must be able to stop early.
		count := 0
		for range makeSet(a, b, c).All() {
			count++
			break
		}
		if count != 1 {
			t.Errorf("All() yielded %d elements after break, want 1", count)
		}
	})

	t.Run("Insert", func(t *testing.T) {
		set := makeSet(a)
		// The bool result is what distinguishes Insert from the older Add:
		// it lets callers tell "added" from "already present".
		if !set.Insert(b) {
			t.Error("Insert(b) = false, want true")
		}
		if set.Insert(b) {
			t.Error("Insert(b) again = true, want false")
		}
		if !contains(set, b) || set.Len() != 2 {
			t.Errorf("after Insert(b) set = %v, want a and b", elems(set))
		}
	})

	t.Run("InsertAll", func(t *testing.T) {
		set := makeSet(a)
		if !set.InsertAll(slices.Values([]E{a, b})) {
			t.Error("InsertAll([a b]) = false, want true")
		}
		if set.InsertAll(slices.Values([]E{a, b})) {
			t.Error("InsertAll([a b]) again = true, want false")
		}
		if set.InsertAll(slices.Values([]E(nil))) {
			t.Error("InsertAll([]) = true, want false")
		}
		if set.Len() != 2 {
			t.Errorf("after InsertAll set = %v, want a and b", elems(set))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		set := makeSet(a, b)
		if !set.Delete(a) {
			t.Error("Delete(a) = false, want true")
		}
		if set.Delete(a) {
			t.Error("Delete(a) again = true, want false")
		}
		if contains(set, a) || set.Len() != 1 {
			t.Errorf("after Delete(a) set = %v, want b only", elems(set))
		}
		// Deleting from a nil map is legal in Go, so it must be here too.
		if nilSet.Delete(a) {
			t.Error("Delete(a) on nil set = true, want false")
		}
	})

	t.Run("DeleteAll", func(t *testing.T) {
		set := makeSet(a, b)
		if !set.DeleteAll(slices.Values([]E{a, c})) {
			t.Error("DeleteAll([a c]) = false, want true")
		}
		if set.DeleteAll(slices.Values([]E{a, c})) {
			t.Error("DeleteAll([a c]) again = true, want false")
		}
		if set.Len() != 1 || !contains(set, b) {
			t.Errorf("after DeleteAll set = %v, want b only", elems(set))
		}
		if nilSet.DeleteAll(slices.Values([]E{a})) {
			t.Error("DeleteAll([a]) on nil set = true, want false")
		}
	})

	t.Run("DeleteFunc", func(t *testing.T) {
		set := makeSet(a, b)
		if !set.DeleteFunc(func(e E) bool { return e == a }) {
			t.Error("DeleteFunc(== a) = false, want true")
		}
		if set.DeleteFunc(func(E) bool { return false }) {
			t.Error("DeleteFunc(never) = true, want false")
		}
		if set.Len() != 1 || !contains(set, b) {
			t.Errorf("after DeleteFunc set = %v, want b only", elems(set))
		}
		if nilSet.DeleteFunc(func(E) bool { return true }) {
			t.Error("DeleteFunc(always) on nil set = true, want false")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		set := makeSet(a, b)
		set.Clear()
		if set.Len() != 0 {
			t.Errorf("after Clear set = %v, want empty", elems(set))
		}
		// Clearing a nil set must be a no-op, not a panic.
		nilSet.Clear()
	})

	t.Run("Clone", func(t *testing.T) {
		set := makeSet(a, b)
		clone := set.Clone()
		if !clone.Equal(set) {
			t.Errorf("Clone() = %v, want %v", elems(clone), elems(set))
		}
		// The copy must be independent, otherwise callers that clone
		// before mutating would corrupt the original.
		clone.Delete(a)
		if !set.Contains(a) {
			t.Error("deleting from the clone also changed the original")
		}
		if got := nilSet.Clone().Len(); got != 0 {
			t.Errorf("Clone() of nil set has Len %d, want 0", got)
		}
	})

	t.Run("Equal", func(t *testing.T) {
		if !makeSet(a, b).Equal(makeSet(b, a)) {
			t.Error("Equal of the same elements in another order = false, want true")
		}
		if makeSet(a).Equal(makeSet(a, b)) {
			t.Error("Equal of different lengths = true, want false")
		}
		if makeSet(a).Equal(makeSet(b)) {
			t.Error("Equal of different elements = true, want false")
		}
		if !nilSet.Equal(makeSet()) {
			t.Error("nil set is not Equal to the empty set, want true")
		}
	})

	t.Run("String", func(t *testing.T) {
		// The format is type specific, but it must at least be callable
		// on a nil set, since String is part of the abstract interface.
		_ = makeSet(a, b).String()
		_ = nilSet.String()
	})

	t.Run("Union", func(t *testing.T) {
		x, y := makeSet(a, b), makeSet(b, c)
		if got := x.Union(y); !got.Equal(makeSet(a, b, c)) {
			t.Errorf("Union() = %v, want a, b and c", elems(got))
		}
		// Neither operand may be modified by the functional form.
		if !x.Equal(makeSet(a, b)) || !y.Equal(makeSet(b, c)) {
			t.Error("Union modified one of its operands")
		}
		if got := nilSet.Union(makeSet(a)); !got.Equal(makeSet(a)) {
			t.Errorf("Union() on nil set = %v, want a", elems(got))
		}
	})

	t.Run("UnionWith", func(t *testing.T) {
		x, y := makeSet(a, b), makeSet(b, c)
		x.UnionWith(y)
		if !x.Equal(makeSet(a, b, c)) {
			t.Errorf("after UnionWith x = %v, want a, b and c", elems(x))
		}
		if !y.Equal(makeSet(b, c)) {
			t.Error("UnionWith modified its argument")
		}
		// Nothing to store, so this must not panic.
		nilSet.UnionWith(makeSet())
	})

	t.Run("Intersection", func(t *testing.T) {
		x, y := makeSet(a, b), makeSet(b, c)
		if got := x.Intersection(y); !got.Equal(makeSet(b)) {
			t.Errorf("Intersection() = %v, want b", elems(got))
		}
		if !x.Equal(makeSet(a, b)) || !y.Equal(makeSet(b, c)) {
			t.Error("Intersection modified one of its operands")
		}
		if got := nilSet.Intersection(makeSet(a)); got.Len() != 0 {
			t.Errorf("Intersection() on nil set = %v, want empty", elems(got))
		}
	})

	t.Run("IntersectionWith", func(t *testing.T) {
		x, y := makeSet(a, b), makeSet(b, c)
		x.IntersectionWith(y)
		if !x.Equal(makeSet(b)) {
			t.Errorf("after IntersectionWith x = %v, want b", elems(x))
		}
		if !y.Equal(makeSet(b, c)) {
			t.Error("IntersectionWith modified its argument")
		}
		// Only removes elements, so a nil receiver is valid.
		nilSet.IntersectionWith(makeSet(a))
	})

	t.Run("Intersects", func(t *testing.T) {
		if !makeSet(a, b).Intersects(makeSet(b, c)) {
			t.Error("Intersects of overlapping sets = false, want true")
		}
		if makeSet(a).Intersects(makeSet(b)) {
			t.Error("Intersects of disjoint sets = true, want false")
		}
		// The implementation iterates the smaller set, so cover both sizes.
		if !makeSet(a, b, c).Intersects(makeSet(c)) {
			t.Error("Intersects with a smaller argument = false, want true")
		}
		if !makeSet(c).Intersects(makeSet(a, b, c)) {
			t.Error("Intersects with a larger argument = false, want true")
		}
		if nilSet.Intersects(makeSet(a)) {
			t.Error("Intersects on nil set = true, want false")
		}
	})

	t.Run("Difference", func(t *testing.T) {
		x, y := makeSet(a, b), makeSet(b, c)
		// Asymmetric, as specified by the Go collections proposal:
		// only the elements of the receiver that are not in the argument.
		// The symmetric variant is SymmetricDifference.
		if got := x.Difference(y); !got.Equal(makeSet(a)) {
			t.Errorf("Difference() = %v, want a only", elems(got))
		}
		if got := y.Difference(x); !got.Equal(makeSet(c)) {
			t.Errorf("reversed Difference() = %v, want c only", elems(got))
		}
		if !x.Equal(makeSet(a, b)) || !y.Equal(makeSet(b, c)) {
			t.Error("Difference modified one of its operands")
		}
		if got := nilSet.Difference(makeSet(a)); got.Len() != 0 {
			t.Errorf("Difference() on nil set = %v, want empty", elems(got))
		}
	})

	t.Run("DifferenceWith", func(t *testing.T) {
		x, y := makeSet(a, b), makeSet(b, c)
		x.DifferenceWith(y)
		if !x.Equal(makeSet(a)) {
			t.Errorf("after DifferenceWith x = %v, want a only", elems(x))
		}
		if !y.Equal(makeSet(b, c)) {
			t.Error("DifferenceWith modified its argument")
		}
		// Only removes elements, so a nil receiver is valid.
		nilSet.DifferenceWith(makeSet(a))
	})

	t.Run("SymmetricDifference", func(t *testing.T) {
		x, y := makeSet(a, b), makeSet(b, c)
		if got := x.SymmetricDifference(y); !got.Equal(makeSet(a, c)) {
			t.Errorf("SymmetricDifference() = %v, want a and c", elems(got))
		}
		// Symmetric, unlike Difference.
		if got := y.SymmetricDifference(x); !got.Equal(makeSet(a, c)) {
			t.Errorf("reversed SymmetricDifference() = %v, want a and c", elems(got))
		}
		if !x.Equal(makeSet(a, b)) || !y.Equal(makeSet(b, c)) {
			t.Error("SymmetricDifference modified one of its operands")
		}
		if got := nilSet.SymmetricDifference(makeSet(a)); !got.Equal(makeSet(a)) {
			t.Errorf("SymmetricDifference() on nil set = %v, want a", elems(got))
		}
	})

	t.Run("SymmetricDifferenceWith", func(t *testing.T) {
		x, y := makeSet(a, b), makeSet(b, c)
		x.SymmetricDifferenceWith(y)
		if !x.Equal(makeSet(a, c)) {
			t.Errorf("after SymmetricDifferenceWith x = %v, want a and c", elems(x))
		}
		if !y.Equal(makeSet(b, c)) {
			t.Error("SymmetricDifferenceWith modified its argument")
		}
		// Nothing to store, so this must not panic.
		nilSet.SymmetricDifferenceWith(makeSet())
	})
}
