package country

import (
	"encoding/json"
	"maps"
	"regexp"
	"slices"
	"testing"

	"github.com/domonda/go-types/internal/normtest"
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
		{in: "Greece", want: GR}, // also the name of EL, see TestNameToCode_OrderIndependent

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

// TestNameToCode_OrderIndependent pins that ParseCode resolves an English
// country name to the same code in every process. GR and EL (the European
// Union's code for Greece) share the name "Greece", and the name index is
// built by ranging over countryMap whose iteration order is random, so a
// "last one wins" index returned GR or EL at random. The ISO 3166-1 code GR
// must win independent of the order, and any new duplicate name would make
// the forward and backward builds differ.
func TestNameToCode_OrderIndependent(t *testing.T) {
	codes := slices.Sorted(maps.Keys(countryMap))
	reversed := slices.Clone(codes)
	slices.Reverse(reversed)
	forward := newNameToCode(slices.Values(codes))
	backward := newNameToCode(slices.Values(reversed))
	if !maps.Equal(forward, backward) {
		t.Errorf("name index depends on the order of the country codes")
	}
	if got := forward["greece"]; got != GR {
		t.Errorf(`name index["greece"] = %q, want %q`, got, GR)
	}

	// Each build ranges over countryMap in a new random order
	for range 50 {
		if got := newNameToCode(maps.Keys(countryMap))["greece"]; got != GR {
			t.Fatalf(`name index["greece"] = %q, want %q`, got, GR)
		}
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
// Valid and Validate carry a #custom marker because they resolve through
// Normalized and therefore accept non-normalized input (lower case,
// surrounding whitespace, AltCodes and German names). The exact-match switch
// over the enum constants that go-enum generates instead would reject all of
// it. String is the one go-enum generates: it returns the code as stored,
// so fmt shows what a Code holds instead of a normalized form that hides
// a non-canonical value. Invalid is marked //#exclude so that go-enum keeps
// it out of the enum value lists while it stays a typed Code constant with
// working methods.
//
// Value: protects=country.Code's #custom Valid/Validate keep accepting
// non-normalized input, String returns the stored code, and Invalid stays
// a typed sentinel outside Enums and EnumStrings;
// fails_when=a go-enum run replaces Valid or Validate with the generated
// exact-match switch, String normalizes again, or Invalid loses its
// //#exclude marker;
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
				if got, err := tt.in.Normalized(); err != nil || string(got) != tt.want {
					t.Errorf("Code(%q).Normalized() = %q, %v; want %q", string(tt.in), got, err, tt.want)
				}
				if got := tt.in.String(); got != string(tt.in) {
					t.Errorf("Code(%q).String() = %q, want the stored code", string(tt.in), got)
				}
			})
		}
	})

	t.Run("rejected input is returned unchanged", func(t *testing.T) {
		// Normalized returns an unresolvable code unchanged with its error,
		// and String returns every code as stored.
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

func TestCode_ValidAndNormalized(t *testing.T) {
	tests := []struct {
		c    Code
		want bool
	}{
		{c: "DE", want: true},
		{c: "de", want: false},   // valid, but lower case
		{c: " DE ", want: false}, // valid, but surrounded by spaces
		{c: "D", want: false},    // valid AltCode, but not ISO 3166-1
		{c: "xx", want: false},
		{c: Invalid, want: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			if got := tt.c.ValidAndNormalized(); got != tt.want {
				t.Errorf("Code(%q).ValidAndNormalized() = %t, want %t", string(tt.c), got, tt.want)
			}
		})
	}

	// Every canonical code must report itself as already normalized,
	// otherwise code that only normalizes when !ValidAndNormalized()
	// would never settle on a stable value.
	for _, c := range Code("").Enums() {
		if !c.ValidAndNormalized() {
			t.Errorf("enum value %q is not ValidAndNormalized", string(c))
		}
	}
}

func TestCode_IsEU(t *testing.T) {
	tests := []struct {
		c    Code
		want bool
	}{
		{c: AT, want: true},
		{c: DE, want: true},
		{c: GR, want: true},
		// EL is the EU's own code for Greece, used by Greek VAT IDs
		// (vat.ID.CountryCode), so it must be an EU member like GR.
		{c: EL, want: true},
		{c: "el", want: true},
		{c: "de", want: true},          // normalized before the lookup
		{c: "Deutschland", want: true}, // AltCodes are resolved
		{c: CH, want: false},           // EFTA, not EU
		{c: NO, want: false},           // EEA, not EU
		{c: GB, want: false},           // left the EU with Brexit
		{c: US, want: false},
		{c: "xx", want: false},
		{c: Invalid, want: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			if got := tt.c.IsEU(); got != tt.want {
				t.Errorf("Code(%q).IsEU() = %t, want %t", string(tt.c), got, tt.want)
			}
		})
	}
}

func TestCode_EnglishName(t *testing.T) {
	tests := []struct {
		c    Code
		want string
	}{
		{c: DE, want: "Germany"},
		{c: " at ", want: "Austria"}, // normalized before the lookup
		{c: "Schweiz", want: "Switzerland"},
		{c: "xx", want: ""},
		{c: Invalid, want: ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			if got := tt.c.EnglishName(); got != tt.want {
				t.Errorf("Code(%q).EnglishName() = %q, want %q", string(tt.c), got, tt.want)
			}
		})
	}

	// ParseCode accepts English names, so every code must have one.
	for _, c := range Code("").Enums() {
		if c.EnglishName() == "" {
			t.Errorf("enum value %q has no English name", string(c))
		}
	}
}

func TestCode_Scan(t *testing.T) {
	t.Run("string is stored normalized", func(t *testing.T) {
		// Scan normalizes, so a code read from the database
		// compares equal to the constant of the same country.
		var c Code
		if err := c.Scan("de"); err != nil {
			t.Fatal(err)
		}
		if c != DE {
			t.Errorf("Scan(\"de\") = %q, want %q", string(c), "DE")
		}
	})
	t.Run("bytes are copied", func(t *testing.T) {
		// database/sql drivers may reuse the []byte buffer after Scan,
		// so the Code must not alias it.
		buf := []byte("DE")
		var c Code
		if err := c.Scan(buf); err != nil {
			t.Fatal(err)
		}
		copy(buf, "XX")
		if c != DE {
			t.Errorf("Scan([]byte) = %q, want %q after the buffer was reused", string(c), "DE")
		}
	})
	t.Run("NULL is Invalid", func(t *testing.T) {
		c := DE
		if err := c.Scan(nil); err != nil {
			t.Fatal(err)
		}
		if c != Invalid {
			t.Errorf("Scan(nil) = %q, want Invalid", string(c))
		}
	})
	t.Run("unsupported type", func(t *testing.T) {
		var c Code
		if err := c.Scan(int64(49)); err == nil {
			t.Error("Scan(int64) = nil, want an error")
		}
	})
}

func TestCode_Value(t *testing.T) {
	tests := []struct {
		c    Code
		want any
	}{
		// Invalid is the empty string and must be written as SQL NULL.
		{c: Invalid, want: nil},
		{c: DE, want: "DE"},
		// Codes are written normalized so equal countries compare equal in SQL.
		{c: "de", want: "DE"},
		{c: "Österreich", want: "AT"},
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			got, err := tt.c.Value()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Code(%q).Value() = %#v, want %#v", string(tt.c), got, tt.want)
			}
		})
	}
}

// Scan(Value()) must give back the same Code, including Invalid as NULL.
func TestCode_SQLRoundTrip(t *testing.T) {
	for _, c := range []Code{Invalid, DE, XK} {
		v, err := c.Value()
		if err != nil {
			t.Fatal(err)
		}
		var got Code
		if err := got.Scan(v); err != nil {
			t.Fatal(err)
		}
		if got != c {
			t.Errorf("Scan(Value()) of %q = %q", string(c), string(got))
		}
	}
}

func TestCode_MarshalJSON(t *testing.T) {
	tests := []struct {
		c    Code
		want string
	}{
		{c: DE, want: `"DE"`},
		// JSON output is normalized so API consumers always see ISO codes.
		{c: "de", want: `"DE"`},
		{c: "Deutschland", want: `"DE"`},
		{c: Invalid, want: `""`},
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			got, err := json.Marshal(tt.c)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("json.Marshal(Code(%q)) = %s, want %s", string(tt.c), got, tt.want)
			}
		})
	}

	t.Run("round-trip", func(t *testing.T) {
		type doc struct {
			Country Code `json:"country"`
		}
		data, err := json.Marshal(doc{Country: AT})
		if err != nil {
			t.Fatal(err)
		}
		var got doc
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if got.Country != AT {
			t.Errorf("JSON round-trip of AT = %q", string(got.Country))
		}
	})
}

// The published JSON schema must accept every code this package produces,
// otherwise API clients validating against the schema would reject
// our own output.
func TestCode_JSONSchema(t *testing.T) {
	s := Code("").JSONSchema()
	if s.Type != "string" {
		t.Errorf("JSONSchema().Type = %q, want string", s.Type)
	}
	pattern := regexp.MustCompile(s.Pattern)
	for _, c := range Code("").Enums() {
		if !pattern.MatchString(string(c)) {
			t.Errorf("enum value %q does not match the schema pattern %s", string(c), s.Pattern)
		}
	}
	if pattern.MatchString("de") {
		t.Errorf("schema pattern %s accepts the non-normalized code \"de\"", s.Pattern)
	}
}

func TestCode_ScanString(t *testing.T) {
	t.Run("valid input is stored normalized", func(t *testing.T) {
		for _, validate := range []bool{true, false} {
			var c Code
			if err := c.ScanString(" österreich ", validate); err != nil {
				t.Fatal(err)
			}
			if c != AT {
				t.Errorf("ScanString(validate=%t) = %q, want AT", validate, string(c))
			}
		}
	})
	t.Run("validate rejects invalid input and keeps the old value", func(t *testing.T) {
		c := DE
		if err := c.ScanString("Atlantis", true); err == nil {
			t.Error("ScanString(\"Atlantis\", true) = nil, want an error")
		}
		if c != DE {
			t.Errorf("failed ScanString changed the value to %q", string(c))
		}
	})
	t.Run("without validate invalid input is stored as is", func(t *testing.T) {
		var c Code
		if err := c.ScanString("Atlantis", false); err != nil {
			t.Fatal(err)
		}
		if c != "Atlantis" {
			t.Errorf("ScanString(\"Atlantis\", false) = %q, want the input unchanged", string(c))
		}
	})
}

// Nullable must map the Invalid sentinel to Null, as documented,
// and keep every other value unchanged.
func TestCode_Nullable(t *testing.T) {
	if got := Invalid.Nullable(); got != Null {
		t.Errorf("Invalid.Nullable() = %q, want Null", string(got))
	}
	if got := DE.Nullable(); got != NullableCode("DE") {
		t.Errorf("DE.Nullable() = %q, want DE", string(got))
	}
	if got := Code("de").Nullable(); got != NullableCode("de") {
		t.Errorf("Code(\"de\").Nullable() = %q, want the input unchanged", string(got))
	}
}

// TestCode_ReadingAndWriting checks the reading and writing contract shared
// by the normalizing types: every reading path stores a valid code
// normalized and an invalid one unchanged, Value refuses an invalid code,
// and fmt prints the stored code.
func TestCode_ReadingAndWriting(t *testing.T) {
	normtest.Run(t, normtest.Spec[Code]{
		Canonical:    DE,
		Variants:     []Code{"de", " DE ", "d", "Deutschland"},
		Invalid:      []Code{"XX", "Atlantis"},
		GoEnumString: true,
	})
}
