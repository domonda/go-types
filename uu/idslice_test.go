package uu

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDSlice(t *testing.T) {
	expected := IDSliceMustFromStrings(
		"ec449f0f-e10c-4edb-8b59-0e6c896fdca5",
		"2d6a2c10-e4a6-45a3-a705-8115214a3778",
		"f3e52e97-e976-4a4c-a602-294310bcf935",
		"cc5873e6-286d-48cd-ae88-bda3a1e986e3",
	)

	jsonStr := `["ec449f0f-e10c-4edb-8b59-0e6c896fdca5","2d6a2c10-e4a6-45a3-a705-8115214a3778","f3e52e97-e976-4a4c-a602-294310bcf935","cc5873e6-286d-48cd-ae88-bda3a1e986e3"]`

	j, err := json.Marshal(expected)
	assert.NoError(t, err)
	assert.Equal(t, jsonStr, string(j))

	var parsed IDSlice
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed)

	err = json.Unmarshal([]byte(`null`), &parsed)
	assert.NoError(t, err)
	assert.Nil(t, parsed)

	j, err = json.Marshal(nil)
	assert.NoError(t, err)
	assert.Equal(t, `null`, string(j))

	parsed = nil
	err = json.Unmarshal([]byte(`[]`), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, IDSlice{}, parsed)

	j, err = json.Marshal(make(IDSlice, 0))
	assert.NoError(t, err)
	assert.Equal(t, `[]`, string(j))

	tests := []string{
		jsonStr,
		`"ec449f0f-e10c-4edb-8b59-0e6c896fdca5","2d6a2c10-e4a6-45a3-a705-8115214a3778","f3e52e97-e976-4a4c-a602-294310bcf935","cc5873e6-286d-48cd-ae88-bda3a1e986e3"`,
		`ec449f0f-e10c-4edb-8b59-0e6c896fdca5,2d6a2c10-e4a6-45a3-a705-8115214a3778,f3e52e97-e976-4a4c-a602-294310bcf935,cc5873e6-286d-48cd-ae88-bda3a1e986e3`,
		`[ec449f0f-e10c-4edb-8b59-0e6c896fdca5,2d6a2c10-e4a6-45a3-a705-8115214a3778,f3e52e97-e976-4a4c-a602-294310bcf935,cc5873e6-286d-48cd-ae88-bda3a1e986e3]`,
	}
	for _, str := range tests {
		t.Run(str, func(t *testing.T) {
			got, err := IDSliceFromString(str)
			assert.NoError(t, err)
			assert.Equal(t, expected, got)
		})
	}

	got, err := IDSliceFromString("")
	assert.NoError(t, err)
	assert.Nil(t, got)

	got, err = IDSliceFromStrings(nil)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestIDSlice_Value(t *testing.T) {
	tests := []struct {
		name    string
		s       IDSlice
		want    driver.Value
		wantErr bool
	}{
		{name: "nil", s: nil, want: nil},
		{name: "len0", s: IDSlice{}, want: driver.Value(`{}`)},
		{name: "len1", s: IDSliceMustFromStrings("3004417b-25ca-441c-924f-102e571e5b5b"), want: driver.Value(`{"3004417b-25ca-441c-924f-102e571e5b5b"}`)},
		{name: "len2", s: IDSliceMustFromStrings("4a6ae04c-8718-4cea-929e-0d8071d328c7", "52d75836-03e0-4b38-8405-bbaa0f392d12"), want: driver.Value(`{"4a6ae04c-8718-4cea-929e-0d8071d328c7","52d75836-03e0-4b38-8405-bbaa0f392d12"}`)},
		{name: "len3", s: IDSliceMustFromStrings("2de0c5e3-d660-4ced-b929-f3a28a42849c", "b445f20b-d964-4fde-a360-39dc60b2157d", "2646cba5-4bc6-454f-bdd0-b869a2650f7e"), want: driver.Value(`{"2de0c5e3-d660-4ced-b929-f3a28a42849c","b445f20b-d964-4fde-a360-39dc60b2157d","2646cba5-4bc6-454f-bdd0-b869a2650f7e"}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("IDSlice.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IDSlice.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIDSlice_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    IDSlice
		wantErr bool
	}{
		{name: "nil", value: nil, want: nil, wantErr: false},
		{name: "empty", value: ``, want: nil, wantErr: false},
		{name: "xxx", value: `xxx`, want: nil, wantErr: true},
		{
			name:  "21e70a0c-7f0e-44d8-8ae5-48992b89d0c5",
			value: `{"21e70a0c-7f0e-44d8-8ae5-48992b89d0c5"}`,
			want:  IDSlice{IDMust("21e70a0c-7f0e-44d8-8ae5-48992b89d0c5")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s IDSlice
			err := s.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("IDSlice.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("IDSlice.Scan() = %v, want %v", s, tt.want)
			}
		})
	}
}

func TestIDSliceFromString_Null(t *testing.T) {
	// The unquoted SQL and JSON null spellings map to the nil slice,
	// which Value and MarshalJSON write back as NULL and null.
	for _, str := range []string{"null", "NULL"} {
		got, err := IDSliceFromString(str)
		if err != nil {
			t.Fatalf("IDSliceFromString(%q) returned %v", str, err)
		}
		if got != nil {
			t.Errorf("IDSliceFromString(%q) = %s, want nil", str, got)
		}
	}
}

func TestIDSliceFromStrings_Error(t *testing.T) {
	// A single invalid string must fail the whole call instead of silently
	// leaving an IDNil element in the middle of the slice.
	got, err := IDSliceFromStrings([]string{testIDA.String(), "not-an-uuid"})
	if err == nil {
		t.Errorf("IDSliceFromStrings() of an invalid string = %s, want an error", got)
	}
	if got != nil {
		t.Errorf("IDSliceFromStrings() of an invalid string = %s, want nil", got)
	}
}

func TestIDSliceMustFromStrings_Panics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("IDSliceMustFromStrings of an invalid string did not panic")
		}
	}()
	IDSliceMustFromStrings("not-an-uuid")
}

func TestIDSliceMust(t *testing.T) {
	// Without values the result is nil, not an allocated empty slice,
	// so that it is written as SQL NULL and JSON null.
	if got := IDSliceMust[string](); got != nil {
		t.Errorf("IDSliceMust() = %s, want nil", got)
	}
	got := IDSliceMust(testIDA.String(), testIDB.String())
	if want := (IDSlice{testIDA, testIDB}); !got.Equal(want) {
		t.Errorf("IDSliceMust() = %s, want %s", got, want)
	}
	// Unlike IDSliceMustFromStrings this keeps the argument order.
	if got := IDSliceMust(testIDB, testIDA); !got.Equal(IDSlice{testIDB, testIDA}) {
		t.Errorf("IDSliceMust() of IDs = %s, want %s", got, IDSlice{testIDB, testIDA})
	}

	t.Run("panics on an invalid ID", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("IDSliceMust of an invalid ID did not panic")
			}
		}()
		IDSliceMust("not-an-uuid")
	})
}

func TestIDSlice_AsSlice(t *testing.T) {
	// AsSlice exists only to implement the IDs interface, so it must
	// return the receiver itself rather than a copy.
	s := IDSlice{testIDA}
	if got := s.AsSlice(); !slices.Equal(got, s) {
		t.Errorf("AsSlice() = %s, want %s", got, s)
	}
}

func TestIDSlice_Raw(t *testing.T) {
	// Raw reinterprets the slice without copying, so that it can be passed
	// to APIs taking [][16]byte. A copy would silently drop writes.
	s := IDSlice{testIDA, testIDB}
	raw := s.Raw()
	if len(raw) != len(s) {
		t.Fatalf("Raw() has length %d, want %d", len(raw), len(s))
	}
	for i := range s {
		if raw[i] != [16]byte(s[i]) {
			t.Errorf("Raw()[%d] = %x, want %x", i, raw[i], s[i])
		}
	}
	raw[0] = [16]byte(testIDC)
	if s[0] != testIDC {
		t.Error("Raw() returned a copy, want a view of the same memory")
	}
}

func TestIDSlice_ForEach(t *testing.T) {
	s := IDSlice{testIDA, testIDB, testIDC}
	var visited IDSlice
	err := s.ForEach(func(id ID) error {
		visited = append(visited, id)
		return nil
	})
	if err != nil {
		t.Fatalf("ForEach returned %v", err)
	}
	if !visited.Equal(s) {
		t.Errorf("ForEach visited %s, want %s", visited, s)
	}

	// A callback error stops the loop immediately, which is what makes
	// a sentinel error usable as a break.
	stop := errors.New("stop")
	var count int
	err = s.ForEach(func(ID) error {
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

func TestIDSlice_PrettyString(t *testing.T) {
	s := IDSlice{testIDA, testIDB}
	if got, want := s.PrettyString(), s.String(); got != want {
		t.Errorf("PrettyString() = %s, want %s", got, want)
	}
}

func TestIDSlice_SortedCloneAndClone(t *testing.T) {
	s := IDSlice{testIDB, testIDA, testIDC}
	sorted := s.SortedClone()
	// Sorting a clone must not disturb the order of the original,
	// which is the only difference to Sort.
	if !s.Equal(IDSlice{testIDB, testIDA, testIDC}) {
		t.Errorf("SortedClone modified the receiver to %s", s)
	}
	want := IDSlice{testIDA, testIDB, testIDC}
	want.Sort()
	if !sorted.Equal(want) {
		t.Errorf("SortedClone() = %s, want %s", sorted, want)
	}

	clone := s.Clone()
	if !clone.Equal(s) {
		t.Errorf("Clone() = %s, want %s", clone, s)
	}
	clone[0] = testIDC
	if s[0] == testIDC {
		t.Error("writing to the clone also changed the original")
	}
	if got := IDSlice(nil).Clone(); got != nil {
		t.Errorf("Clone() of a nil slice = %s, want nil", got)
	}
}

func TestIDSlice_IndexOfAndContains(t *testing.T) {
	s := IDSlice{testIDA, testIDB, testIDA}
	if got := s.IndexOf(testIDA); got != 0 {
		t.Errorf("IndexOf() = %d, want the first occurrence at 0", got)
	}
	if got := s.IndexOf(testIDB); got != 1 {
		t.Errorf("IndexOf() = %d, want 1", got)
	}
	if got := s.IndexOf(testIDC); got != -1 {
		t.Errorf("IndexOf() of a missing ID = %d, want -1", got)
	}
	if !s.Contains(testIDB) {
		t.Error("Contains() of a present ID = false, want true")
	}
	if s.Contains(testIDC) {
		t.Error("Contains() of a missing ID = true, want false")
	}
}

func TestIDSlice_ContainsAny(t *testing.T) {
	s := IDSlice{testIDA, testIDB}
	if !s.ContainsAny(IDSlice{testIDC, testIDB}) {
		t.Error("ContainsAny() of an overlapping slice = false, want true")
	}
	if s.ContainsAny(IDSlice{testIDC}) {
		t.Error("ContainsAny() of a disjoint slice = true, want false")
	}
	if s.ContainsAny(nil) {
		t.Error("ContainsAny(nil) = true, want false")
	}
}

func TestIDSlice_ContainsAnyFromSet(t *testing.T) {
	s := IDSlice{testIDA, testIDB}
	if !s.ContainsAnyFromSet(MakeIDSet(testIDC, testIDB)) {
		t.Error("ContainsAnyFromSet() of an overlapping set = false, want true")
	}
	if s.ContainsAnyFromSet(MakeIDSet(testIDC)) {
		t.Error("ContainsAnyFromSet() of a disjoint set = true, want false")
	}
	if s.ContainsAnyFromSet(nil) {
		t.Error("ContainsAnyFromSet(nil) = true, want false")
	}
}

func TestIDSlice_Equal(t *testing.T) {
	// Unlike IDSet.Equal this compares order, because a slice is ordered.
	s := IDSlice{testIDA, testIDB}
	if !s.Equal(IDSlice{testIDA, testIDB}) {
		t.Error("Equal() of identical slices = false, want true")
	}
	if s.Equal(IDSlice{testIDB, testIDA}) {
		t.Error("Equal() of a reordered slice = true, want false")
	}
	if s.Equal(IDSlice{testIDA}) {
		t.Error("Equal() of a shorter slice = true, want false")
	}
	if !IDSlice(nil).Equal(IDSlice{}) {
		t.Error("Equal() of a nil and an empty slice = false, want true")
	}
}

func TestIDSlice_Remove(t *testing.T) {
	t.Run("RemoveFirst", func(t *testing.T) {
		s := IDSlice{testIDA, testIDB, testIDA}
		if got := s.RemoveFirst(testIDA); got != 0 {
			t.Errorf("RemoveFirst() = %d, want 0", got)
		}
		// Only the first occurrence is removed, the second one stays.
		if want := (IDSlice{testIDB, testIDA}); !s.Equal(want) {
			t.Errorf("after RemoveFirst slice = %s, want %s", s, want)
		}
		if got := s.RemoveFirst(testIDC); got != -1 {
			t.Errorf("RemoveFirst() of a missing ID = %d, want -1", got)
		}
		if want := (IDSlice{testIDB, testIDA}); !s.Equal(want) {
			t.Errorf("RemoveFirst of a missing ID changed the slice to %s", s)
		}
	})

	t.Run("RemoveAll", func(t *testing.T) {
		// Adjacent duplicates check that the loop index is corrected
		// after a removal, otherwise every second one would survive.
		s := IDSlice{testIDA, testIDA, testIDB, testIDA}
		if got := s.RemoveAll(testIDA); got != 3 {
			t.Errorf("RemoveAll() = %d, want 3", got)
		}
		if want := (IDSlice{testIDB}); !s.Equal(want) {
			t.Errorf("after RemoveAll slice = %s, want %s", s, want)
		}
		if got := s.RemoveAll(testIDC); got != 0 {
			t.Errorf("RemoveAll() of a missing ID = %d, want 0", got)
		}
	})

	t.Run("RemoveAt", func(t *testing.T) {
		s := IDSlice{testIDA, testIDB, testIDC}
		s.RemoveAt(1)
		if want := (IDSlice{testIDA, testIDC}); !s.Equal(want) {
			t.Errorf("after RemoveAt slice = %s, want %s", s, want)
		}
	})
}

func TestIDSlice_MarshalAndUnmarshalText(t *testing.T) {
	s := IDSlice{testIDA, testIDB}
	text, err := s.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText returned %v", err)
	}
	if want := s.String(); string(text) != want {
		t.Errorf("MarshalText() = %s, want %s", text, want)
	}

	var parsed IDSlice
	if err := parsed.UnmarshalText(text); err != nil {
		t.Fatalf("UnmarshalText returned %v", err)
	}
	if !parsed.Equal(s) {
		t.Errorf("UnmarshalText() = %s, want %s", parsed, s)
	}

	if err := parsed.UnmarshalText([]byte("[not-an-uuid]")); err == nil {
		t.Error("UnmarshalText of an invalid ID returned no error")
	}
}

func TestIDSlice_ScanBytes(t *testing.T) {
	// Database drivers return the PostgreSQL array literal as []byte,
	// so that case must behave exactly like the string case.
	want := IDSlice{testIDA}
	value, err := want.Value()
	if err != nil {
		t.Fatalf("Value returned %v", err)
	}
	var got IDSlice
	if err := got.Scan([]byte(value.(string))); err != nil {
		t.Fatalf("Scan returned %v", err)
	}
	if !got.Equal(want) {
		t.Errorf("Scan() = %s, want %s", got, want)
	}

	if err := got.Scan([]byte(`{not-an-uuid}`)); err == nil {
		t.Error("Scan of an invalid ID returned no error")
	}
}

func TestIDSlice_MarshalJSON_Nil(t *testing.T) {
	// The nil slice is JSON null, unlike the empty slice which is [].
	got, err := IDSlice(nil).MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned %v", err)
	}
	if string(got) != "null" {
		t.Errorf("MarshalJSON() of a nil slice = %s, want null", got)
	}
}

func TestIDSlice_UnmarshalJSON_Errors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "not an array", data: `{"a":1}`},
		{name: "not a string array", data: `[1,2]`},
		{name: "invalid ID", data: `["not-an-uuid"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s IDSlice
			if err := s.UnmarshalJSON([]byte(tt.data)); err == nil {
				t.Errorf("UnmarshalJSON(%s) returned no error", tt.data)
			}
		})
	}
}
