package strutil

import (
	"slices"
	"testing"

	"github.com/domonda/go-types/internal/collections"
	"github.com/domonda/go-types/internal/collections/settest"
)

// StringSet must implement the abstract set interface of the Go 1.28
// collections proposal, https://go.dev/issue/80590.
var _ collections.Set[string, StringSet] = StringSet(nil)

func TestStringSet_AbstractSetConformance(t *testing.T) {
	settest.Run(t, NewStringSet, StringSet(nil), "a", "b", "c")
}

func TestStringSet_Difference(t *testing.T) {
	// Difference is asymmetric as of the Go 1.28 collections proposal.
	// It replaces the former Diff method, which returned the symmetric
	// difference and is now SymmetricDifference. This test exists to pin
	// the new meaning down against a regression to the old one.
	x := NewStringSet("a", "b")
	y := NewStringSet("b", "c")

	if got, want := x.Difference(y), NewStringSet("a"); !got.Equal(want) {
		t.Errorf("Difference() = %s, want %s", got, want)
	}
	if got, want := y.Difference(x), NewStringSet("c"); !got.Equal(want) {
		t.Errorf("reversed Difference() = %s, want %s", got, want)
	}
	if got, want := x.SymmetricDifference(y), NewStringSet("a", "c"); !got.Equal(want) {
		t.Errorf("SymmetricDifference() = %s, want %s", got, want)
	}
}

func TestStringSet_NilInsertPanics(t *testing.T) {
	// Documented behaviour: a nil StringSet is valid for every read and for
	// removals, but storing into one panics like a nil map assignment.
	defer func() {
		if recover() == nil {
			t.Error("Insert on a nil StringSet did not panic")
		}
	}()
	StringSet(nil).Insert("a")
}

func TestStringSet_AllMatchesSorted(t *testing.T) {
	// All is the iterator form of the pre-existing Sorted accessor.
	// They must agree, otherwise iterating a set would silently skip strings.
	set := NewStringSet("c", "a", "b")
	if got, want := slices.Sorted(set.All()), set.Sorted(); !slices.Equal(got, want) {
		t.Errorf("All() = %v, Sorted() = %v", got, want)
	}
}

func TestStringSet_Len(t *testing.T) {
	if got := NewStringSet("a", "b").Len(); got != 2 {
		t.Errorf("Len() = %d, want 2", got)
	}
	if got := StringSet(nil).Len(); got != 0 {
		t.Errorf("Len() of nil set = %d, want 0", got)
	}
}
