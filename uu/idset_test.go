package uu

import (
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
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

func TestMakeIDSetFromStrings(t *testing.T) {
	got, err := MakeIDSetFromStrings([]string{testIDA.String(), testIDB.String(), testIDA.String()})
	if err != nil {
		t.Fatalf("MakeIDSetFromStrings returned %v", err)
	}
	if want := MakeIDSet(testIDA, testIDB); !got.Equal(want) {
		t.Errorf("MakeIDSetFromStrings() = %s, want %s", got, want)
	}

	// A single invalid string must fail the whole call instead of
	// silently producing a set that is missing IDs.
	got, err = MakeIDSetFromStrings([]string{testIDA.String(), "not-an-uuid"})
	if err == nil {
		t.Errorf("MakeIDSetFromStrings() of an invalid string = %s, want an error", got)
	}
	if got != nil {
		t.Errorf("MakeIDSetFromStrings() of an invalid string = %s, want nil", got)
	}
}

func TestMakeIDSetMustFromStrings(t *testing.T) {
	got := MakeIDSetMustFromStrings(testIDA.String(), testIDB.String())
	if want := MakeIDSet(testIDA, testIDB); !got.Equal(want) {
		t.Errorf("MakeIDSetMustFromStrings() = %s, want %s", got, want)
	}

	t.Run("panics on an invalid string", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("MakeIDSetMustFromStrings of an invalid string did not panic")
			}
		}()
		MakeIDSetMustFromStrings("not-an-uuid")
	})
}

func TestIDSetFromString(t *testing.T) {
	// IDSetFromString has to parse back what IDSet.String() writes,
	// which is the "set" prefix followed by the slice format, but it
	// also accepts the bare formats used in logs and SQL dumps.
	tests := []struct {
		str  string
		want IDSet
	}{
		{str: MakeIDSet(testIDA, testIDB).String(), want: MakeIDSet(testIDA, testIDB)},
		{str: "[" + testIDA.String() + "," + testIDB.String() + "]", want: MakeIDSet(testIDA, testIDB)},
		{str: testIDA.String() + ", " + testIDB.String(), want: MakeIDSet(testIDA, testIDB)},
		{str: "null", want: nil},
		{str: "NULL", want: nil},
		{str: "", want: nil},
		{str: "[]", want: nil},
		{str: "set[]", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := IDSetFromString(tt.str)
			if err != nil {
				t.Fatalf("IDSetFromString(%q) returned %v", tt.str, err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("IDSetFromString(%q) = %s, want %s", tt.str, got, tt.want)
			}
			if tt.want == nil && got != nil {
				t.Errorf("IDSetFromString(%q) = %s, want nil", tt.str, got)
			}
		})
	}

	if got, err := IDSetFromString("[not-an-uuid]"); err == nil {
		t.Errorf("IDSetFromString of an invalid ID = %s, want an error", got)
	}
}

func TestIDSetMust(t *testing.T) {
	// Without values the result is nil, not an allocated empty set,
	// so that it is written as SQL NULL and JSON null.
	if got := IDSetMust[string](); got != nil {
		t.Errorf("IDSetMust() = %s, want nil", got)
	}
	got := IDSetMust(testIDA.String(), testIDB.String())
	if want := MakeIDSet(testIDA, testIDB); !got.Equal(want) {
		t.Errorf("IDSetMust() = %s, want %s", got, want)
	}
	if got := IDSetMust(testIDA, testIDB); !got.Equal(MakeIDSet(testIDA, testIDB)) {
		t.Errorf("IDSetMust() of IDs = %s, want %s", got, MakeIDSet(testIDA, testIDB))
	}

	t.Run("panics on an invalid ID", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("IDSetMust of an invalid ID did not panic")
			}
		}()
		IDSetMust("not-an-uuid")
	})
}

func TestIDSet_Strings(t *testing.T) {
	// Sorted, because the map iteration order of a set is undefined
	// and callers compare or log the result.
	got := MakeIDSet(testIDA, testIDB, testIDC).Strings()
	want := []string{testIDA.String(), testIDB.String(), testIDC.String()}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Strings() = %v, want %v", got, want)
	}
	if got := IDSet(nil).Strings(); got != nil {
		t.Errorf("Strings() of a nil set = %v, want nil", got)
	}
}

func TestIDSet_StringAndPrettyString(t *testing.T) {
	// Built from the sorted ID strings rather than from AsSortedSlice(), so
	// the assertion does not restate the implementation: String is literally
	// "set" + AsSortedSlice().String(), and comparing against that would pass
	// even if both regressed together.
	ids := []string{testIDA.String(), testIDB.String()}
	slices.Sort(ids)
	want := "set[" + strings.Join(ids, ",") + "]"

	set := MakeIDSet(testIDA, testIDB)
	if got := set.String(); got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}
	// PrettyString deliberately drops the "set" prefix that String carries:
	// it is s.AsSortedSlice().PrettyString(), so an IDSet and an IDSlice with
	// the same IDs pretty-print identically. Only String round-trips through
	// IDSetFromString, which strips that prefix.
	if got, want := set.PrettyString(), "["+strings.Join(ids, ",")+"]"; got != want {
		t.Errorf("PrettyString() = %s, want %s", got, want)
	}
	if got, want := IDSet(nil).String(), "set[]"; got != want {
		t.Errorf("String() of a nil set = %s, want %s", got, want)
	}
}

func TestIDSet_GetOne(t *testing.T) {
	if got := MakeIDSet(testIDA).GetOne(); got != testIDA {
		t.Errorf("GetOne() of a one element set = %s, want %s", got, testIDA)
	}
	// IDNil is the documented result for an empty set, so callers can use
	// it without a separate emptiness check.
	if got := IDSet(nil).GetOne(); got != IDNil {
		t.Errorf("GetOne() of a nil set = %s, want %s", got, IDNil)
	}
}

func TestIDSet_AsSet(t *testing.T) {
	// AsSet exists only to implement the IDs interface, so it must return the
	// receiver itself rather than a copy. Assert aliasing, not equality:
	// maps.Equal would also hold for maps.Clone(s), and callers like AddIDs
	// store through the returned map.
	set := MakeIDSet(testIDA)
	got := set.AsSet()
	if !maps.Equal(got, set) {
		t.Errorf("AsSet() = %s, want %s", got, set)
	}
	got.Insert(testIDB)
	if !set.Contains(testIDB) {
		t.Error("AsSet() returned a copy, want a view of the receiver")
	}
}

func TestIDSet_ForEach(t *testing.T) {
	set := MakeIDSet(testIDA, testIDB, testIDC)
	var visited IDSlice
	err := set.ForEach(func(id ID) error {
		visited = append(visited, id)
		return nil
	})
	if err != nil {
		t.Fatalf("ForEach returned %v", err)
	}
	visited.Sort()
	if want := set.AsSortedSlice(); !visited.Equal(want) {
		t.Errorf("ForEach visited %s, want %s", visited, want)
	}

	// A callback error stops the loop immediately, which is what makes
	// a sentinel error usable as a break.
	stop := errors.New("stop")
	count := 0
	err = set.ForEach(func(ID) error {
		count++
		return stop
	})
	if !errors.Is(err, stop) {
		t.Errorf("ForEach returned %v, want %v", err, stop)
	}
	if count != 1 {
		t.Errorf("ForEach called the callback %d times after an error, want 1", count)
	}
}

func TestIDSet_IsEmptyAndIsNull(t *testing.T) {
	// IsEmpty and IsNull differ only for the allocated empty set:
	// a nil set is both, an allocated empty set is only empty.
	// Value and MarshalJSON use nil to write SQL NULL and JSON null.
	tests := []struct {
		name       string
		set        IDSet
		wantEmpty  bool
		wantIsNull bool
	}{
		{name: "nil", set: nil, wantEmpty: true, wantIsNull: true},
		{name: "allocated empty", set: make(IDSet), wantEmpty: true, wantIsNull: false},
		{name: "non empty", set: MakeIDSet(testIDA), wantEmpty: false, wantIsNull: false},
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

func TestIDSet_MarshalAndUnmarshalText(t *testing.T) {
	set := MakeIDSet(testIDA, testIDB)
	text, err := set.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText returned %v", err)
	}
	// Compared against the literal, not against set.String(), which is what
	// MarshalText returns verbatim.
	ids := []string{testIDA.String(), testIDB.String()}
	slices.Sort(ids)
	if want := "set[" + strings.Join(ids, ",") + "]"; string(text) != want {
		t.Errorf("MarshalText() = %s, want %s", text, want)
	}

	var parsed IDSet
	if err := parsed.UnmarshalText(text); err != nil {
		t.Fatalf("UnmarshalText returned %v", err)
	}
	if !parsed.Equal(set) {
		t.Errorf("UnmarshalText() = %s, want %s", parsed, set)
	}

	if err := parsed.UnmarshalText([]byte("[not-an-uuid]")); err == nil {
		t.Error("UnmarshalText of an invalid ID returned no error")
	}
}

func TestIDSet_ScanAndValue(t *testing.T) {
	set := MakeIDSet(testIDA, testIDB)

	value, err := set.Value()
	if err != nil {
		t.Fatalf("Value returned %v", err)
	}
	// Value writes the sorted slice, so that the SQL literal of a set
	// is deterministic.
	wantValue, err := set.AsSortedSlice().Value()
	if err != nil {
		t.Fatalf("IDSlice.Value returned %v", err)
	}
	if value != wantValue {
		t.Errorf("Value() = %v, want %v", value, wantValue)
	}

	// Scan assigns a new map, so it works on an uninitialized variable.
	var scanned IDSet
	if err := scanned.Scan(value); err != nil {
		t.Fatalf("Scan returned %v", err)
	}
	if !scanned.Equal(set) {
		t.Errorf("Scan() = %s, want %s", scanned, set)
	}

	// The nil map is SQL NULL in both directions.
	if value, err := IDSet(nil).Value(); err != nil || value != nil {
		t.Errorf("Value() of a nil set = %v, %v, want nil, nil", value, err)
	}
	if err := scanned.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) returned %v", err)
	}
	if scanned != nil {
		t.Errorf("Scan(nil) = %s, want nil", scanned)
	}

	if err := scanned.Scan(123); err == nil {
		t.Error("Scan of an unsupported type returned no error")
	}
}

func TestIDSet_MarshalAndUnmarshalJSON(t *testing.T) {
	set := MakeIDSet(testIDA, testIDB)

	data, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("Marshal returned %v", err)
	}
	// Sorted, so that the JSON of a set is deterministic and comparable.
	wantData, err := set.AsSortedSlice().MarshalJSON()
	if err != nil {
		t.Fatalf("IDSlice.MarshalJSON returned %v", err)
	}
	if string(data) != string(wantData) {
		t.Errorf("Marshal() = %s, want %s", data, wantData)
	}

	var parsed IDSet
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal returned %v", err)
	}
	if !parsed.Equal(set) {
		t.Errorf("Unmarshal() = %s, want %s", parsed, set)
	}

	// The nil map is JSON null in both directions.
	data, err = json.Marshal(IDSet(nil))
	if err != nil {
		t.Fatalf("Marshal of a nil set returned %v", err)
	}
	if string(data) != "null" {
		t.Errorf("Marshal of a nil set = %s, want null", data)
	}
	if err := json.Unmarshal([]byte("null"), &parsed); err != nil {
		t.Fatalf("Unmarshal of null returned %v", err)
	}
	if parsed != nil {
		t.Errorf("Unmarshal of null = %s, want nil", parsed)
	}

	if err := parsed.UnmarshalJSON([]byte(`["not-an-uuid"]`)); err == nil {
		t.Error("Unmarshal of an invalid ID returned no error")
	}
}

// TestIDSet_DeprecatedMutators covers the pre-Insert API that is kept for
// compatibility. It must stay behaviour compatible with the new methods,
// because callers mix both while migrating.
func TestIDSet_DeprecatedMutators(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		set := make(IDSet)
		set.Add(testIDA)
		set.Add(testIDA)
		if want := MakeIDSet(testIDA); !set.Equal(want) {
			t.Errorf("after Add set = %s, want %s", set, want)
		}
	})

	t.Run("AddSlice", func(t *testing.T) {
		set := MakeIDSet(testIDA)
		set.AddSlice(IDSlice{testIDB, testIDB})
		if want := MakeIDSet(testIDA, testIDB); !set.Equal(want) {
			t.Errorf("after AddSlice set = %s, want %s", set, want)
		}
		set.AddSlice(nil)
		if want := MakeIDSet(testIDA, testIDB); !set.Equal(want) {
			t.Errorf("AddSlice(nil) changed the set to %s", set)
		}
	})

	t.Run("AddSet", func(t *testing.T) {
		set := MakeIDSet(testIDA)
		set.AddSet(MakeIDSet(testIDA, testIDB))
		if want := MakeIDSet(testIDA, testIDB); !set.Equal(want) {
			t.Errorf("after AddSet set = %s, want %s", set, want)
		}
		set.AddSet(nil)
		if want := MakeIDSet(testIDA, testIDB); !set.Equal(want) {
			t.Errorf("AddSet(nil) changed the set to %s", set)
		}
	})

	t.Run("DeleteSlice", func(t *testing.T) {
		set := MakeIDSet(testIDA, testIDB)
		set.DeleteSlice(IDSlice{testIDB, testIDC})
		if want := MakeIDSet(testIDA); !set.Equal(want) {
			t.Errorf("after DeleteSlice set = %s, want %s", set, want)
		}
	})

	t.Run("DeleteSet", func(t *testing.T) {
		set := MakeIDSet(testIDA, testIDB)
		set.DeleteSet(MakeIDSet(testIDB, testIDC))
		if want := MakeIDSet(testIDA); !set.Equal(want) {
			t.Errorf("after DeleteSet set = %s, want %s", set, want)
		}
	})
}

func TestIDSet_AddIDs(t *testing.T) {
	// AddIDs takes the IDs interface, so both an IDSlice and an IDSet
	// have to work as the source.
	set := MakeIDSet(testIDA)
	set.AddIDs(IDSlice{testIDB, testIDB})
	set.AddIDs(MakeIDSet(testIDC))
	if want := MakeIDSet(testIDA, testIDB, testIDC); !set.Equal(want) {
		t.Errorf("after AddIDs set = %s, want %s", set, want)
	}
}
