package types

import (
	"encoding/json"
	"reflect"
	"slices"
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

func TestSet_GetOne(t *testing.T) {
	if got := NewSet(7).GetOne(); got != 7 {
		t.Errorf("GetOne() of a one element set = %d, want 7", got)
	}
	if got := NewSet(1, 2).GetOne(); got != 1 && got != 2 {
		t.Errorf("GetOne() = %d, want 1 or 2", got)
	}
	// The zero value is the documented result for an empty set,
	// so callers can use it without a separate emptiness check.
	if got := Set[int](nil).GetOne(); got != 0 {
		t.Errorf("GetOne() of a nil set = %d, want 0", got)
	}
	if got := NewSet[string]().GetOne(); got != "" {
		t.Errorf("GetOne() of an empty set = %q, want empty string", got)
	}
}

func TestSet_ContainsAny(t *testing.T) {
	set := NewSet(1, 2, 3)
	if !set.ContainsAny(4, 2) {
		t.Error("ContainsAny(4, 2) = false, want true")
	}
	if set.ContainsAny(4, 5) {
		t.Error("ContainsAny(4, 5) = true, want false")
	}
	if set.ContainsAny() {
		t.Error("ContainsAny() without values = true, want false")
	}
}

func TestSet_Map(t *testing.T) {
	set := NewSet(1, 2, 3)
	// The bool result of mapFunc filters, so Map is both a map and a filter:
	// odd values are dropped, even ones are negated.
	got := set.Map(func(val int) (int, bool) {
		if val%2 != 0 {
			return 0, false
		}
		return -val, true
	})
	if want := NewSet(-2); !got.Equal(want) {
		t.Errorf("Map() = %v, want %v", got, want)
	}
	// The receiver must not be modified.
	if want := NewSet(1, 2, 3); !set.Equal(want) {
		t.Errorf("Map modified its receiver to %v", set)
	}
	// Mapping several values onto the same result deduplicates them.
	got = set.Map(func(int) (int, bool) { return 0, true })
	if want := NewSet(0); !got.Equal(want) {
		t.Errorf("Map() to a constant = %v, want %v", got, want)
	}
}

func TestSet_IsEmptyAndIsNull(t *testing.T) {
	// IsEmpty and IsNull differ only for the allocated empty set:
	// a nil set is both, an allocated empty set is only empty.
	// nullable.Nullable uses IsNull to decide SQL/JSON null.
	tests := []struct {
		name       string
		set        Set[int]
		wantEmpty  bool
		wantIsNull bool
	}{
		{name: "nil", set: nil, wantEmpty: true, wantIsNull: true},
		{name: "allocated empty", set: NewSet[int](), wantEmpty: true, wantIsNull: false},
		{name: "non empty", set: NewSet(1), wantEmpty: false, wantIsNull: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.set.IsEmpty(); got != tt.wantEmpty {
				t.Errorf("IsEmpty() = %t, want %t", got, tt.wantEmpty)
			}
			if got := tt.set.IsNull(); got != tt.wantIsNull {
				t.Errorf("IsNull() = %t, want %t", got, tt.wantIsNull)
			}
		})
	}
}

func TestSet_MarshalJSON(t *testing.T) {
	// The elements are sorted so that the JSON of a set is deterministic
	// and can be compared or used as a cache key.
	got, err := json.Marshal(NewSet(3, 1, 2))
	if err != nil {
		t.Fatalf("Marshal returned %v", err)
	}
	if want := `[1,2,3]`; string(got) != want {
		t.Errorf("Marshal() = %s, want %s", got, want)
	}

	got, err = json.Marshal(Set[int](nil))
	if err != nil {
		t.Fatalf("Marshal of a nil set returned %v", err)
	}
	if want := `null`; string(got) != want {
		t.Errorf("Marshal of a nil set = %s, want %s", got, want)
	}

	// null and the empty array are distinct: only a nil set is null.
	// Sorted returns a nil slice for an empty set, so MarshalJSON has to
	// special-case it or an allocated empty set would also emit null.
	got, err = json.Marshal(NewSet[int]())
	if err != nil {
		t.Fatalf("Marshal of an empty set returned %v", err)
	}
	if want := `[]`; string(got) != want {
		t.Errorf("Marshal of an empty set = %s, want %s", got, want)
	}
}

func TestSet_UnmarshalJSON(t *testing.T) {
	t.Run("into nil set", func(t *testing.T) {
		var set Set[int]
		if err := json.Unmarshal([]byte(`[1,2,3]`), &set); err != nil {
			t.Fatalf("Unmarshal returned %v", err)
		}
		if want := NewSet(1, 2, 3); !set.Equal(want) {
			t.Errorf("Unmarshal() = %v, want %v", set, want)
		}
	})

	t.Run("duplicates are deduplicated", func(t *testing.T) {
		var set Set[string]
		if err := json.Unmarshal([]byte(`["a","b","a"]`), &set); err != nil {
			t.Fatalf("Unmarshal returned %v", err)
		}
		if set.Len() != 2 {
			t.Errorf("Unmarshal() = %v, want 2 elements", set)
		}
	})

	t.Run("replaces the elements of an existing set", func(t *testing.T) {
		// The existing map is reused, so the previous elements have to be
		// cleared instead of merged with the unmarshalled ones.
		set := NewSet(1, 2, 3)
		if err := json.Unmarshal([]byte(`[9]`), &set); err != nil {
			t.Fatalf("Unmarshal returned %v", err)
		}
		if want := NewSet(9); !set.Equal(want) {
			t.Errorf("Unmarshal() = %v, want %v", set, want)
		}
	})

	t.Run("null empties the set instead of nilling it", func(t *testing.T) {
		// A nil map panics when a key is set, so unmarshalling null must
		// never leave a nil map behind: the next Insert would crash. This
		// is the opposite of a slice, where nil is the right answer for
		// null because a nil slice can be appended to.
		//
		// Asserted on the IsNull axis, not IsEmpty: IsEmpty is true for
		// both a nil and an allocated empty set and would not catch a
		// regression to nil.
		for _, name := range []string{"nil receiver", "allocated receiver"} {
			t.Run(name, func(t *testing.T) {
				set := Set[int](nil)
				if name == "allocated receiver" {
					set = NewSet(1, 2)
				}
				if err := json.Unmarshal([]byte(`null`), &set); err != nil {
					t.Fatalf("Unmarshal of null returned %v", err)
				}
				if set.IsNull() {
					t.Fatal("Unmarshal of null produced a nil set, want an allocated empty one")
				}
				if !set.IsEmpty() {
					t.Errorf("Unmarshal of null = %v, want empty", set)
				}
				// The property the allocation exists for.
				set.Insert(1)
			})
		}
	})

	t.Run("null empties an allocated set in place", func(t *testing.T) {
		// Emptied, not replaced, so another holder of the same map sees it.
		set := NewSet(1, 2)
		alias := set
		if err := json.Unmarshal([]byte(`null`), &set); err != nil {
			t.Fatalf("Unmarshal of null returned %v", err)
		}
		if alias.Len() != 0 {
			t.Errorf("Unmarshal of null replaced the map instead of clearing it: alias still holds %v", alias)
		}
	})

	t.Run("not an array", func(t *testing.T) {
		var set Set[int]
		if err := json.Unmarshal([]byte(`{"a":1}`), &set); err == nil {
			t.Error("Unmarshal of a JSON object returned no error")
		}
	})
}

func TestReduceSet(t *testing.T) {
	sum := ReduceSet(NewSet(1, 2, 3), func(last, val int) int { return last + val })
	if sum != 6 {
		t.Errorf("ReduceSet() = %d, want 6", sum)
	}
	// The zero value of the result type is the accumulator seed,
	// which is what an empty set has to return.
	if got := ReduceSet(Set[int](nil), func(last, val int) int { return last + val }); got != 0 {
		t.Errorf("ReduceSet() of a nil set = %d, want 0", got)
	}
}

func TestReduceSlice(t *testing.T) {
	// Unlike ReduceSet the slice order is defined, so a non-commutative
	// reduce function has a single correct result.
	got := ReduceSlice([]string{"a", "b", "c"}, func(last string, val string) string { return last + val })
	if want := "abc"; got != want {
		t.Errorf("ReduceSlice() = %q, want %q", got, want)
	}
	if got := ReduceSlice([]int(nil), func(last, val int) int { return last + val }); got != 0 {
		t.Errorf("ReduceSlice() of a nil slice = %d, want 0", got)
	}
}

func TestSetToRandomizedSlice(t *testing.T) {
	got := SetToRandomizedSlice(NewSet(1, 2, 3))
	slices.Sort(got)
	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("SetToRandomizedSlice() = %v, want %v as a set", got, want)
	}
	// nil instead of an empty slice, so that the result marshals as JSON null.
	if got := SetToRandomizedSlice(Set[int](nil)); got != nil {
		t.Errorf("SetToRandomizedSlice() of a nil set = %v, want nil", got)
	}
}

func TestSetToSortedSlice_LenCases(t *testing.T) {
	// The implementation special cases the empty and the one element set
	// to avoid allocating and sorting, so every case needs to be checked.
	if got := SetToSortedSlice(Set[int](nil)); got != nil {
		t.Errorf("SetToSortedSlice() of a nil set = %v, want nil", got)
	}
	if got, want := SetToSortedSlice(NewSet(1)), []int{1}; !slices.Equal(got, want) {
		t.Errorf("SetToSortedSlice() of a one element set = %v, want %v", got, want)
	}
	if got, want := SetToSortedSlice(NewSet(2, 1)), []int{1, 2}; !slices.Equal(got, want) {
		t.Errorf("SetToSortedSlice() = %v, want %v", got, want)
	}
}

// TestSet_DeprecatedMutators covers the pre-Insert API that is kept for
// compatibility. It must stay behaviour compatible with the new methods,
// because callers mix both while migrating.
func TestSet_DeprecatedMutators(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		set := NewSet[int]()
		set.Add(1)
		set.Add(1)
		if want := NewSet(1); !set.Equal(want) {
			t.Errorf("after Add set = %v, want %v", set, want)
		}
	})

	t.Run("AddSlice", func(t *testing.T) {
		set := NewSet(1)
		set.AddSlice([]int{2, 3, 2})
		if want := NewSet(1, 2, 3); !set.Equal(want) {
			t.Errorf("after AddSlice set = %v, want %v", set, want)
		}
		set.AddSlice(nil)
		if want := NewSet(1, 2, 3); !set.Equal(want) {
			t.Errorf("AddSlice(nil) changed the set to %v", set)
		}
	})

	t.Run("AddSet", func(t *testing.T) {
		set := NewSet(1)
		set.AddSet(NewSet(1, 2))
		if want := NewSet(1, 2); !set.Equal(want) {
			t.Errorf("after AddSet set = %v, want %v", set, want)
		}
		set.AddSet(nil)
		if want := NewSet(1, 2); !set.Equal(want) {
			t.Errorf("AddSet(nil) changed the set to %v", set)
		}
	})

	t.Run("DeleteSlice", func(t *testing.T) {
		set := NewSet(1, 2, 3)
		set.DeleteSlice([]int{2, 4})
		if want := NewSet(1, 3); !set.Equal(want) {
			t.Errorf("after DeleteSlice set = %v, want %v", set, want)
		}
	})

	t.Run("DeleteSet", func(t *testing.T) {
		set := NewSet(1, 2, 3)
		set.DeleteSet(NewSet(2, 4))
		if want := NewSet(1, 3); !set.Equal(want) {
			t.Errorf("after DeleteSet set = %v, want %v", set, want)
		}
	})
}
