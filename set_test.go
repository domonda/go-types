package types

import (
	"reflect"
	"testing"

	"github.com/domonda/go-types/internal/collections"
	"github.com/domonda/go-types/internal/collections/settest"
)

func TestSetToSortedSlice(t *testing.T) {
	{
		set := map[int]struct{}{
			3: {},
			2: {},
			1: {},
		}
		want := []int{1, 2, 3}
		got := SetToSortedSlice(set)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SetToSortedSlice() = %#v, want %#v", got, want)
		}
	}
	{
		set := map[string]struct{}{
			"3": {},
			"2": {},
			"1": {},
		}
		want := []string{"1", "2", "3"}
		got := SetToSortedSlice(set)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SetToSortedSlice() = %#v, want %#v", got, want)
		}
	}
}

// Set[T] must implement the abstract set interface of the Go 1.28
// collections proposal, https://go.dev/issue/80590.
var _ collections.Set[int, Set[int]] = Set[int](nil)

func TestSet_AbstractSetConformance(t *testing.T) {
	settest.Run(t, NewSet[int], Set[int](nil), 1, 2, 3)
	settest.Run(t, NewSet[string], Set[string](nil), "a", "b", "c")
}

func TestSet_Difference(t *testing.T) {
	// Difference is asymmetric as of the Go 1.28 collections proposal.
	// It returned the symmetric difference before, so this test exists
	// to pin the new meaning down against a regression to the old one.
	x := NewSet(1, 2, 3)
	y := NewSet(3, 4)

	if got, want := x.Difference(y), NewSet(1, 2); !got.Equal(want) {
		t.Errorf("Difference() = %v, want %v", got, want)
	}
	if got, want := y.Difference(x), NewSet(4); !got.Equal(want) {
		t.Errorf("reversed Difference() = %v, want %v", got, want)
	}
	if got, want := x.SymmetricDifference(y), NewSet(1, 2, 4); !got.Equal(want) {
		t.Errorf("SymmetricDifference() = %v, want %v", got, want)
	}
}

func TestSet_ContainsSet(t *testing.T) {
	set := NewSet(1, 2, 3)
	if !set.ContainsSet(NewSet(1, 3)) {
		t.Error("ContainsSet of a subset = false, want true")
	}
	if set.ContainsSet(NewSet(1, 4)) {
		t.Error("ContainsSet of a non-subset = true, want false")
	}
	if !set.ContainsSet(nil) {
		t.Error("ContainsSet of the empty set = false, want true")
	}
}

func TestSet_NilInsertPanics(t *testing.T) {
	// Documented behaviour: a nil set is valid for every read and for
	// removals, but storing into one panics like a nil map assignment.
	defer func() {
		if recover() == nil {
			t.Error("Insert on a nil Set did not panic")
		}
	}()
	Set[int](nil).Insert(1)
}
