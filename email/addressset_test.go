package email

import (
	"reflect"
	"slices"
	"testing"

	"github.com/domonda/go-types/internal/collections"
	"github.com/domonda/go-types/internal/collections/settest"
)

func TestAddressSet_Scan(t *testing.T) {
	tests := []struct {
		name    string
		set     AddressSet
		value   any
		want    AddressSet
		wantErr bool
	}{
		{
			name:  "SplitArray bug",
			value: "{some@example.com,if.need.of.a.''declaration.of.compliance''.please.contact.us@example.com}",
			want:  MakeAddressSet("some@example.com", "if.need.of.a.''declaration.of.compliance''.please.contact.us@example.com"),
		},
		{
			name:  "{single_quote'@example.com}",
			value: "{single_quote'@example.com}",
			want:  MakeAddressSet("single_quote'@example.com"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.set.Scan(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("AddressSet.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.set, tt.want) {
				t.Errorf("AddressSet.Scan() = %v, want %v", tt.set, tt.want)
			}
		})
	}
}

// AddressSet must implement the abstract set interface of the Go 1.28
// collections proposal, https://go.dev/issue/80590.
var _ collections.Set[Address, AddressSet] = AddressSet(nil)

func TestAddressSet_AbstractSetConformance(t *testing.T) {
	settest.Run(t, MakeAddressSet, AddressSet(nil), "a@example.com", "b@example.com", "c@example.com")
}

func TestAddressSet_Difference(t *testing.T) {
	// Difference is asymmetric as specified by the Go 1.28 collections proposal.
	x := MakeAddressSet("a@example.com", "b@example.com")
	y := MakeAddressSet("b@example.com", "c@example.com")

	if got, want := x.Difference(y), MakeAddressSet("a@example.com"); !got.Equal(want) {
		t.Errorf("Difference() = %s, want %s", got, want)
	}
	if got, want := y.Difference(x), MakeAddressSet("c@example.com"); !got.Equal(want) {
		t.Errorf("reversed Difference() = %s, want %s", got, want)
	}
	if got, want := x.SymmetricDifference(y), MakeAddressSet("a@example.com", "c@example.com"); !got.Equal(want) {
		t.Errorf("SymmetricDifference() = %s, want %s", got, want)
	}
}

// TestAddressSet_NilReceivers covers the receiver split that is specific to
// AddressSet: Add and AddSet keep pointer receivers so that they can allocate
// into a nil variable or struct field, while the value-receiver Insert and
// UnionWith of the abstract set interface panic on a nil set like a nil map
// assignment. Existing code stores into nil AddressSet struct fields through
// Add, so that behaviour must not change.
func TestAddressSet_NilReceivers(t *testing.T) {
	t.Run("Add allocates", func(t *testing.T) {
		var set AddressSet
		set.Add("a@example.com")
		if !set.Contains("a@example.com") {
			t.Error("Add on a nil AddressSet did not allocate")
		}
	})

	t.Run("Add on nil struct field allocates", func(t *testing.T) {
		// The pattern that pointer receivers exist for.
		var data struct{ Addrs AddressSet }
		data.Addrs.Add("a@example.com")
		if data.Addrs.Len() != 1 {
			t.Error("Add on a nil AddressSet struct field did not allocate")
		}
	})

	t.Run("AddSet allocates", func(t *testing.T) {
		var set AddressSet
		set.AddSet(MakeAddressSet("a@example.com"))
		if !set.Contains("a@example.com") {
			t.Error("AddSet on a nil AddressSet did not allocate")
		}
	})

	t.Run("AddSet of empty set stays nil", func(t *testing.T) {
		var set AddressSet
		set.AddSet(nil)
		if set != nil {
			t.Error("AddSet of an empty set allocated a map, want nil")
		}
	})

	t.Run("AddNormalized allocates", func(t *testing.T) {
		var set AddressSet
		if err := set.AddNormalized("Some Name <A@Example.COM>"); err != nil {
			t.Fatalf("AddNormalized returned %v", err)
		}
		if set.Len() != 1 {
			t.Error("AddNormalized on a nil AddressSet did not allocate")
		}
	})

	t.Run("AddAddressPart allocates", func(t *testing.T) {
		var set AddressSet
		if err := set.AddAddressPart("Some Name <a@example.com>"); err != nil {
			t.Fatalf("AddAddressPart returned %v", err)
		}
		if !set.Contains("a@example.com") {
			t.Errorf("AddAddressPart on a nil AddressSet = %v, want a@example.com", set.Sorted())
		}
	})

	t.Run("Insert panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("Insert on a nil AddressSet did not panic")
			}
		}()
		AddressSet(nil).Insert("a@example.com")
	})

	t.Run("UnionWith panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("UnionWith on a nil AddressSet did not panic")
			}
		}()
		AddressSet(nil).UnionWith(MakeAddressSet("a@example.com"))
	})
}

func TestNormalizedAddressSet(t *testing.T) {
	// Normalizing on construction is what makes the set deduplicate
	// addresses that only differ in case or display name spelling.
	got, err := NormalizedAddressSet("A@Example.COM", "a@example.com")
	if err != nil {
		t.Fatalf("NormalizedAddressSet returned %v", err)
	}
	if want := MakeAddressSet("a@example.com"); !got.Equal(want) {
		t.Errorf("NormalizedAddressSet() = %s, want %s", got, want)
	}

	got, err = NormalizedAddressSet("a@example.com", "Hello World!")
	if err == nil {
		t.Errorf("NormalizedAddressSet() of an invalid address = %s, want an error", got)
	}
	if got != nil {
		t.Errorf("NormalizedAddressSet() of an invalid address = %s, want nil", got)
	}
}

func TestNormalizedAddressPartSet(t *testing.T) {
	// Unlike NormalizedAddressSet this drops the display name, so two
	// addresses of the same mailbox with different names collapse into one.
	got, err := NormalizedAddressPartSet("Some Name <A@Example.COM>", "Other Name <a@example.com>")
	if err != nil {
		t.Fatalf("NormalizedAddressPartSet returned %v", err)
	}
	if want := MakeAddressSet("a@example.com"); !got.Equal(want) {
		t.Errorf("NormalizedAddressPartSet() = %s, want %s", got, want)
	}

	got, err = NormalizedAddressPartSet("Hello World!")
	if err == nil {
		t.Errorf("NormalizedAddressPartSet() of an invalid address = %s, want an error", got)
	}
	if got != nil {
		t.Errorf("NormalizedAddressPartSet() of an invalid address = %s, want nil", got)
	}
}

func TestAddressSet_AddNormalizedAndAddAddressPart_Errors(t *testing.T) {
	// An invalid address must not end up in the set unnormalized.
	var set AddressSet
	if err := set.AddNormalized("Hello World!"); err == nil {
		t.Error("AddNormalized of an invalid address returned no error")
	}
	if err := set.AddAddressPart("Hello World!"); err == nil {
		t.Error("AddAddressPart of an invalid address returned no error")
	}
	if !set.IsEmpty() {
		t.Errorf("set = %s after failed inserts, want empty", set)
	}
}

func TestAddressSet_IsEmptyAndIsNull(t *testing.T) {
	// IsEmpty and IsNull differ only for the allocated empty set:
	// a nil set is both, an allocated empty set is only empty.
	// Value uses nil to write SQL NULL.
	tests := []struct {
		name       string
		set        AddressSet
		wantEmpty  bool
		wantIsNull bool
	}{
		{name: "nil", set: nil, wantEmpty: true, wantIsNull: true},
		{name: "allocated empty", set: MakeAddressSet(), wantEmpty: true, wantIsNull: false},
		{name: "non empty", set: MakeAddressSet("a@example.com"), wantEmpty: false, wantIsNull: false},
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

func TestAddressSet_GetOne(t *testing.T) {
	if got := MakeAddressSet("a@example.com").GetOne(); got != "a@example.com" {
		t.Errorf("GetOne() of a one element set = %s, want a@example.com", got)
	}
	// The empty Address is the documented result for an empty set,
	// so callers can use it without a separate emptiness check.
	if got := AddressSet(nil).GetOne(); got != "" {
		t.Errorf("GetOne() of a nil set = %s, want the empty address", got)
	}
}

func TestAddressSet_Strings(t *testing.T) {
	// The one element case skips the sort, so all three sizes need checking.
	if got := AddressSet(nil).Strings(); got != nil {
		t.Errorf("Strings() of a nil set = %v, want nil", got)
	}
	if got, want := MakeAddressSet("a@example.com").Strings(), []string{"a@example.com"}; !slices.Equal(got, want) {
		t.Errorf("Strings() of a one element set = %v, want %v", got, want)
	}
	got := MakeAddressSet("c@example.com", "a@example.com", "b@example.com").Strings()
	want := []string{"a@example.com", "b@example.com", "c@example.com"}
	if !slices.Equal(got, want) {
		t.Errorf("Strings() = %v, want %v", got, want)
	}
}

func TestAddressSet_Normalized(t *testing.T) {
	got, err := MakeAddressSet("A@Example.COM", "Some Name <B@example.com>").Normalized()
	if err != nil {
		t.Fatalf("Normalized returned %v", err)
	}
	if want := MakeAddressSet("a@example.com", `"Some Name" <b@example.com>`); !got.Equal(want) {
		t.Errorf("Normalized() = %s, want %s", got, want)
	}

	// An empty set is returned unchanged without allocating a new map.
	empty := AddressSet(nil)
	if got, err := empty.Normalized(); err != nil || got != nil {
		t.Errorf("Normalized() of a nil set = %s, %v, want nil, nil", got, err)
	}

	// On error the original set is returned, so a caller that ignores
	// the error keeps working with the unnormalized addresses.
	set := MakeAddressSet("Hello World!")
	if got, err := set.Normalized(); err == nil || !got.Equal(set) {
		t.Errorf("Normalized() of an invalid address = %s, %v, want the original set and an error", got, err)
	}
}

func TestAddressSet_ValidateAndValid(t *testing.T) {
	valid := MakeAddressSet("a@example.com", "b@example.com")
	if err := valid.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
	if !valid.Valid() {
		t.Error("Valid() = false, want true")
	}

	invalid := MakeAddressSet("Hello World!")
	if err := invalid.Validate(); err == nil {
		t.Error("Validate() of an invalid address = nil, want an error")
	}
	if invalid.Valid() {
		t.Error("Valid() of an invalid address = true, want false")
	}

	if err := AddressSet(nil).Validate(); err != nil {
		t.Errorf("Validate() of a nil set = %v, want nil", err)
	}
}

func TestAddressSet_ValidAndNormalized(t *testing.T) {
	tests := []struct {
		name string
		set  AddressSet
		want bool
	}{
		{name: "normalized", set: MakeAddressSet("a@example.com", "b@example.com"), want: true},
		{name: "nil", set: nil, want: true},
		{name: "invalid", set: MakeAddressSet("Hello World!"), want: false},
		// Normalizing lowercases the address part, so the set changes.
		{name: "not normalized", set: MakeAddressSet("A@Example.COM"), want: false},
		// Two addresses that normalize to the same one collapse, which
		// changes the length instead of the elements.
		{name: "collapsing to fewer elements", set: MakeAddressSet("A@example.com", "a@example.com"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.set.ValidAndNormalized(); got != tt.want {
				t.Errorf("ValidAndNormalized() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestAddressSet_ScanNonArray(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    AddressSet
		wantErr bool
	}{
		// A plain string is a single address, not an array literal,
		// so that a non-array SQL column can be scanned too.
		{name: "plain string", value: "a@example.com", want: MakeAddressSet("a@example.com")},
		{name: "bytes", value: []byte(`{a@example.com,b@example.com}`), want: MakeAddressSet("a@example.com", "b@example.com")},
		{name: "empty array", value: "{}", want: AddressSet{}},
		// A hand written literal, unlike PostgreSQL output, may pad its
		// elements. The padding is not part of the address.
		{name: "padded elements", value: "{ a@example.com , b@example.com }", want: MakeAddressSet("a@example.com", "b@example.com")},
		{name: "whitespace only array", value: "{ }", want: AddressSet{}},
		// Regression cases for the array codec swap (nullable.SplitArray ->
		// notnull.StringArray). Each behaves differently than before, so each
		// is pinned rather than left to be discovered in production.
		// A SQL NULL element used to become the address literally spelled NULL.
		{name: "NULL element", value: "{NULL}", wantErr: true},
		// Quoted, it is the ordinary string "NULL", not a SQL NULL.
		{name: "quoted NULL element", value: `{"NULL"}`, want: MakeAddressSet("NULL")},
		// A trailing comma used to be accepted and yield one element.
		{name: "trailing comma", value: "{a@example.com,}", wantErr: true},
		// Backslash escapes are unescaped now; before, both backslashes stayed.
		{name: "escaped backslash", value: `{"a\\b@example.com"}`, want: MakeAddressSet(`a\b@example.com`)},
		// A nested literal used to yield the inner braces as a single address.
		{name: "nested array", value: "{{a@example.com}}", wantErr: true},
		{name: "null", value: nil, want: nil},
		{name: "empty string", value: "", wantErr: true},
		{name: "invalid array", value: "{,}", wantErr: true},
		{name: "unsupported type", value: 123, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var set AddressSet
			err := set.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Scan() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(set, tt.want) {
				t.Errorf("Scan() = %s, want %s", set, tt.want)
			}
		})
	}
}

func TestAddressSet_Value(t *testing.T) {
	// The nil map is SQL NULL, the empty set the empty array literal.
	if got, err := AddressSet(nil).Value(); err != nil || got != nil {
		t.Errorf("Value() of a nil set = %v, %v, want nil, nil", got, err)
	}
	if got, err := MakeAddressSet().Value(); err != nil || got != "{}" {
		t.Errorf("Value() of an empty set = %v, %v, want {}, nil", got, err)
	}

	// Sorted, so that the SQL literal of a set is deterministic.
	got, err := MakeAddressSet("b@example.com", "a@example.com").Value()
	if err != nil {
		t.Fatalf("Value returned %v", err)
	}
	if want := `{"a@example.com","b@example.com"}`; got != want {
		t.Errorf("Value() = %v, want %s", got, want)
	}
}

// TestAddressSet_ValueScanRoundTrip pins down that Scan is the inverse of
// Value. Both use the same PostgreSQL array codec, so the quotes and the
// backslash escapes that Value writes have to be consumed again by Scan
// instead of ending up inside the addresses.
func TestAddressSet_ValueScanRoundTrip(t *testing.T) {
	sets := []AddressSet{
		MakeAddressSet(),
		MakeAddressSet("a@example.com"),
		MakeAddressSet("b@example.com", "a@example.com"),
		// A display name is quoted by Value and contains the comma and
		// the double quotes that the array format has to escape.
		MakeAddressSet(`"Doe, John" <john@example.com>`, "a@example.com"),
		MakeAddressSet("single_quote'@example.com"),
	}
	for _, want := range sets {
		t.Run(want.String(), func(t *testing.T) {
			value, err := want.Value()
			if err != nil {
				t.Fatalf("Value returned %v", err)
			}
			var got AddressSet
			if err := got.Scan(value); err != nil {
				t.Fatalf("Scan(%v) returned %v", value, err)
			}
			if !got.Equal(want) {
				t.Errorf("Scan(Value()) = %s, want %s", got, want)
			}
		})
	}
}

func TestAddressSet_ScanQuotedArrayElements(t *testing.T) {
	// The PostgreSQL array literal that Value writes quotes every element.
	// Before the quotes were kept, so the addresses came back as
	// `"a@example.com"` including the quote characters.
	var set AddressSet
	if err := set.Scan(`{"a@example.com","b@example.com"}`); err != nil {
		t.Fatalf("Scan returned %v", err)
	}
	if want := MakeAddressSet("a@example.com", "b@example.com"); !set.Equal(want) {
		t.Errorf("Scan() = %s, want %s", set, want)
	}

	// A quoted element may contain the separator and escaped quotes.
	if err := set.Scan(`{"\"Doe, John\" <john@example.com>"}`); err != nil {
		t.Fatalf("Scan returned %v", err)
	}
	if want := MakeAddressSet(`"Doe, John" <john@example.com>`); !set.Equal(want) {
		t.Errorf("Scan() of an escaped element = %s, want %s", set, want)
	}
}

func TestAddressSet_AddressListAndString(t *testing.T) {
	set := MakeAddressSet("b@example.com", "a@example.com")
	if got, want := set.AddressList(), AddressList("a@example.com, b@example.com"); got != want {
		t.Errorf("AddressList() = %s, want %s", got, want)
	}
	// The literal, not string(set.AddressList()), which is what String
	// returns verbatim; comparing the two checks the code against itself.
	if got, want := set.String(), "a@example.com, b@example.com"; got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}
}

// TestAddressSet_DeprecatedMutators covers the pre-DeleteAll API that is kept
// for compatibility. It must stay behaviour compatible with the new methods,
// because callers mix both while migrating.
func TestAddressSet_DeprecatedMutators(t *testing.T) {
	t.Run("DeleteSlice", func(t *testing.T) {
		set := MakeAddressSet("a@example.com", "b@example.com")
		set.DeleteSlice([]Address{"b@example.com", "x@example.com"})
		if want := MakeAddressSet("a@example.com"); !set.Equal(want) {
			t.Errorf("after DeleteSlice set = %s, want %s", set, want)
		}
	})

	t.Run("DeleteSet", func(t *testing.T) {
		set := MakeAddressSet("a@example.com", "b@example.com")
		set.DeleteSet(MakeAddressSet("b@example.com", "x@example.com"))
		if want := MakeAddressSet("a@example.com"); !set.Equal(want) {
			t.Errorf("after DeleteSet set = %s, want %s", set, want)
		}
	})
}
