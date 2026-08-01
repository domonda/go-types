package uu

import (
	"slices"
	"testing"

	"github.com/domonda/go-types/internal/collections"
	"github.com/domonda/go-types/internal/collections/settest"
)

// IDSet must implement the abstract set interface of the Go 1.28
// collections proposal, https://go.dev/issue/80590.
var _ collections.Set[ID, IDSet] = IDSet(nil)

var (
	testIDA = IDMustFromString("9256978d-18e6-4435-ad16-d7046d41b71a")
	testIDB = IDMustFromString("a92bb308-f0f9-43d2-b31d-c0962198c31c")
	testIDC = IDMustFromString("59a268e5-d820-4884-b120-10d1a9b0dd00")
)

func TestIDSet_AbstractSetConformance(t *testing.T) {
	settest.Run(t, MakeIDSet, IDSet(nil), testIDA, testIDB, testIDC)
}

func TestIDSet_Difference(t *testing.T) {
	// Difference is asymmetric as of the Go 1.28 collections proposal.
	// It replaces the former Diff method, which returned the symmetric
	// difference and is now SymmetricDifference. This test exists to pin
	// the new meaning down against a regression to the old one.
	x := MakeIDSet(testIDA, testIDB)
	y := MakeIDSet(testIDB, testIDC)

	if got, want := x.Difference(y), MakeIDSet(testIDA); !got.Equal(want) {
		t.Errorf("Difference() = %s, want %s", got, want)
	}
	if got, want := y.Difference(x), MakeIDSet(testIDC); !got.Equal(want) {
		t.Errorf("reversed Difference() = %s, want %s", got, want)
	}
	if got, want := x.SymmetricDifference(y), MakeIDSet(testIDA, testIDC); !got.Equal(want) {
		t.Errorf("SymmetricDifference() = %s, want %s", got, want)
	}
}

func TestIDSet_NilInsertPanics(t *testing.T) {
	// Documented behaviour: a nil IDSet is valid for every read and for
	// removals, but storing into one panics like a nil map assignment.
	defer func() {
		if recover() == nil {
			t.Error("Insert on a nil IDSet did not panic")
		}
	}()
	IDSet(nil).Insert(testIDA)
}

func TestIDSet_AllMatchesAsSlice(t *testing.T) {
	// All is the iterator form of the pre-existing AsSlice accessor.
	// They must agree, otherwise iterating a set would silently skip IDs.
	set := MakeIDSet(testIDA, testIDB, testIDC)
	fromAll := IDSlice(slices.Collect(set.All()))
	fromAll.Sort()
	fromSlice := set.AsSortedSlice()
	if !slices.Equal(fromAll, fromSlice) {
		t.Errorf("All() = %v, AsSortedSlice() = %v", fromAll, fromSlice)
	}
}
