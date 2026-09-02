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

func TestNewStringSetMergeSlices(t *testing.T) {
	got := NewStringSetMergeSlices([]string{"a", "b"}, []string{"b", "c"}, nil)
	if want := NewStringSet("a", "b", "c"); !got.Equal(want) {
		t.Errorf("NewStringSetMergeSlices() = %s, want %s", got, want)
	}
	if got := NewStringSetMergeSlices(); got.Len() != 0 {
		t.Errorf("NewStringSetMergeSlices() without slices = %s, want empty", got)
	}
}

func TestStringSet_String(t *testing.T) {
	// The format is quoted and sorted so that it is stable across runs,
	// which matters because it ends up in log lines and test failures.
	if got, want := NewStringSet("b", "a").String(), `["a", "b"]`; got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}
	if got, want := NewStringSet().String(), `[]`; got != want {
		t.Errorf("String() of an empty set = %s, want %s", got, want)
	}
	if got, want := StringSet(nil).String(), `[]`; got != want {
		t.Errorf("String() of a nil set = %s, want %s", got, want)
	}
}

func TestStringSet_Sorted(t *testing.T) {
	if got, want := NewStringSet("c", "a", "b").Sorted(), []string{"a", "b", "c"}; !slices.Equal(got, want) {
		t.Errorf("Sorted() = %v, want %v", got, want)
	}
	if got := StringSet(nil).Sorted(); got != nil {
		t.Errorf("Sorted() of a nil set = %v, want nil", got)
	}
}

func TestStringSet_ContainsAny(t *testing.T) {
	set := NewStringSet("a", "b")
	if !set.ContainsAny("x", "b") {
		t.Error(`ContainsAny("x", "b") = false, want true`)
	}
	if set.ContainsAny("x", "y") {
		t.Error(`ContainsAny("x", "y") = true, want false`)
	}
	if set.ContainsAny() {
		t.Error("ContainsAny() without strings = true, want false")
	}
}

func TestStringSet_StringContainsAnyOfSet(t *testing.T) {
	// The reverse of Contains: the set holds the needles, not the haystack.
	set := NewStringSet("foo", "bar")
	if !set.StringContainsAnyOfSet("a bar in a string") {
		t.Error("StringContainsAnyOfSet of a string with a substring = false, want true")
	}
	if set.StringContainsAnyOfSet("nothing here") {
		t.Error("StringContainsAnyOfSet of an unrelated string = true, want false")
	}
	if StringSet(nil).StringContainsAnyOfSet("foo") {
		t.Error("StringContainsAnyOfSet on a nil set = true, want false")
	}
}

// TestStringSet_DeprecatedMutators covers the pre-Insert API that is kept
// for compatibility. It must stay behaviour compatible with the new methods,
// because callers mix both while migrating.
func TestStringSet_DeprecatedMutators(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		set := NewStringSet()
		set.Add("a")
		set.Add("a")
		if want := NewStringSet("a"); !set.Equal(want) {
			t.Errorf("after Add set = %s, want %s", set, want)
		}
	})

	t.Run("AddSlice", func(t *testing.T) {
		set := NewStringSet("a")
		set.AddSlice([]string{"b", "c", "b"})
		if want := NewStringSet("a", "b", "c"); !set.Equal(want) {
			t.Errorf("after AddSlice set = %s, want %s", set, want)
		}
		set.AddSlice(nil)
		if want := NewStringSet("a", "b", "c"); !set.Equal(want) {
			t.Errorf("AddSlice(nil) changed the set to %s", set)
		}
	})

	t.Run("AddSet", func(t *testing.T) {
		set := NewStringSet("a")
		set.AddSet(NewStringSet("a", "b"))
		if want := NewStringSet("a", "b"); !set.Equal(want) {
			t.Errorf("after AddSet set = %s, want %s", set, want)
		}
		set.AddSet(nil)
		if want := NewStringSet("a", "b"); !set.Equal(want) {
			t.Errorf("AddSet(nil) changed the set to %s", set)
		}
	})

	t.Run("DeleteSlice", func(t *testing.T) {
		set := NewStringSet("a", "b", "c")
		set.DeleteSlice([]string{"b", "x"})
		if want := NewStringSet("a", "c"); !set.Equal(want) {
			t.Errorf("after DeleteSlice set = %s, want %s", set, want)
		}
	})

	t.Run("DeleteSet", func(t *testing.T) {
		set := NewStringSet("a", "b", "c")
		set.DeleteSet(NewStringSet("b", "x"))
		if want := NewStringSet("a", "c"); !set.Equal(want) {
			t.Errorf("after DeleteSet set = %s, want %s", set, want)
		}
	})
}

func TestStringSet_IsEmpty(t *testing.T) {
	// StringSet has no IsNull, so unlike the other three set types it cannot
	// distinguish a nil from an allocated empty set. IsEmpty is true for both.
	if !StringSet(nil).IsEmpty() {
		t.Error("IsEmpty() of a nil set = false, want true")
	}
	if !NewStringSet().IsEmpty() {
		t.Error("IsEmpty() of an allocated empty set = false, want true")
	}
	if NewStringSet("a").IsEmpty() {
		t.Error("IsEmpty() of a non-empty set = true, want false")
	}
}
