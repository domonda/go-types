package country

import (
	"maps"
	"slices"
	"testing"
)

func TestParseCode(t *testing.T) {
	tests := []struct {
		in      string
		want    Code
		wantErr bool
	}{
		// 1. Short path: already a canonical alpha-2 code.
		{in: "DE", want: DE},
		{in: "FR", want: FR},

		// 2. Codes and names resolved via Normalized.
		{in: "de", want: DE},
		{in: " AUT ", want: AT},
		{in: "A", want: AT},
		{in: "SUI", want: CH},
		{in: "Deutschland", want: DE},
		{in: "Österreich", want: AT},

		// 3. English country name.
		{in: "Germany", want: DE},
		{in: "germany", want: DE},
		{in: "  United Kingdom  ", want: GB},
		{in: "United States", want: US},

		// Errors: input is returned unchanged.
		{in: "", want: "", wantErr: true},
		{in: "xxx", want: "xxx", wantErr: true},
		{in: "Atlantis", want: "Atlantis", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseCode(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCode(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseCode(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCode_Normalized(t *testing.T) {
	tests := []struct {
		c       Code
		want    Code
		wantErr bool
	}{
		{c: "A", want: AT},
		{c: "AUT", want: AT},
		{c: "d", want: DE},
		{c: "Deu", want: DE},
		{c: "Deutschland", want: DE},
		{c: "Österreich", want: AT},
		{c: "Oesterreich", want: AT},
		// Errors
		{c: "", want: "", wantErr: true},
		{c: "xxx", want: "xxx", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			got, err := tt.c.Normalized()
			if (err != nil) != tt.wantErr {
				t.Errorf("Code.Normalized() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Code.Normalized() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCode_EnumsMatchCountryMap ties the two lists of country codes together:
// Enums() is generated from the constants in code.go, but Valid() resolves
// through countryMap in data.go. A code added to only one of them would make
// Enums() list a code that Valid() rejects, or miss one that it accepts.
func TestCode_EnumsMatchCountryMap(t *testing.T) {
	got := slices.Sorted(slices.Values(Code("").Enums()))
	want := slices.Sorted(maps.Keys(countryMap))
	if !slices.Equal(got, want) {
		t.Errorf("Enums() = %v\nwant the countryMap keys %v", got, want)
	}
}

// TestCode_CustomMethodsAndInvalidSentinel pins the two properties the
// //#enum migration had to preserve, neither of which any other test touches.
//
// Valid, Validate and String carry a #custom marker because they resolve
// through Normalized and therefore accept non-normalized input (lower case,
// surrounding whitespace, AltCodes and German names). The exact-match switch
// over the enum constants that go-enum generates instead would reject all of
// it, and String would stop normalizing. Invalid is declared as the
// conversion Code("") rather than `Invalid Code = ""` so that go-enum keeps
// it out of the enum value lists while it stays a typed Code constant with
// working methods.
//
// Value: protects=country.Code's #custom Valid/Validate/String keep accepting
// non-normalized input, and Invalid stays a typed sentinel outside Enums and
// EnumStrings;
// fails_when=a go-enum run replaces them with the generated exact-match
// switch and plain string conversion, or Invalid is redeclared as
// `Invalid Code = ""`;
// why_new=no country test calls Valid, Validate, String or Invalid at all,
// and TestCode_EnumsMatchCountryMap never looks at EnumStrings;
// seam=none
func TestCode_CustomMethodsAndInvalidSentinel(t *testing.T) {
	t.Run("accepted non-normalized input", func(t *testing.T) {
		tests := []struct {
			in   Code
			want string
		}{
			{in: "DE", want: "DE"}, // already normalized
			{in: "de", want: "DE"}, // lower case
			{in: "  AT  ", want: "AT"},
			{in: "d", want: "DE"},   // AltCodes: ITU code
			{in: "SUI", want: "CH"}, // AltCodes: alpha-3
			{in: "Deutschland", want: "DE"},
		}
		for _, tt := range tests {
			t.Run(string(tt.in), func(t *testing.T) {
				if !tt.in.Valid() {
					t.Errorf("Code(%q).Valid() = false, want true", string(tt.in))
				}
				if err := tt.in.Validate(); err != nil {
					t.Errorf("Code(%q).Validate() = %v, want nil", string(tt.in), err)
				}
				if got := tt.in.String(); got != tt.want {
					t.Errorf("Code(%q).String() = %q, want the normalized %q", string(tt.in), got, tt.want)
				}
			})
		}
	})

	t.Run("rejected input is returned unchanged", func(t *testing.T) {
		// String has no error result, so an unresolvable code has to come
		// back unchanged instead of being replaced by the empty string.
		for _, in := range []Code{"", "xxx", "Atlantis"} {
			t.Run(string(in), func(t *testing.T) {
				if in.Valid() {
					t.Errorf("Code(%q).Valid() = true, want false", string(in))
				}
				if err := in.Validate(); err == nil {
					t.Errorf("Code(%q).Validate() = nil, want an error", string(in))
				}
				if got := in.String(); got != string(in) {
					t.Errorf("Code(%q).String() = %q, want the input unchanged", string(in), got)
				}
			})
		}
	})

	t.Run("Invalid sentinel", func(t *testing.T) {
		// Compile-time: Invalid stays a typed Code constant, not an untyped
		// string, so it keeps its methods and can be compared to a Code.
		var _ Code = Invalid

		if Invalid.Valid() {
			t.Error("Invalid.Valid() = true, want false")
		}
		if slices.Contains(Code("").Enums(), Invalid) {
			t.Error("Enums() contains Invalid, which is not a valid country code")
		}
		if slices.Contains(Code("").EnumStrings(), string(Invalid)) {
			t.Error("EnumStrings() contains the empty string of Invalid")
		}
	})

	t.Run("EnumStrings matches Enums", func(t *testing.T) {
		// Both lists are generated independently from the same constants, so
		// nothing but this assertion keeps them from drifting apart.
		enums := Code("").Enums()
		want := make([]string, len(enums))
		for i, c := range enums {
			want[i] = string(c)
		}
		if got := Code("").EnumStrings(); !slices.Equal(got, want) {
			t.Errorf("EnumStrings() = %v\nwant the Enums() values as strings %v", got, want)
		}
	})
}
