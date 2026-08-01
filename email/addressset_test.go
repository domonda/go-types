package email

import (
	"reflect"
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
