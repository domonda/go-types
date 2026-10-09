package float

import (
	"math"
	"testing"
)

func TestTolerant_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		json         []byte
		wantTolerant Tolerant
		wantErr      bool
	}{
		{name: "nil", json: nil, wantErr: true},
		{name: "JSON boolean", json: []byte(`true`), wantErr: true},
		{name: "JSON non number string", json: []byte(`"true"`), wantErr: true},

		{name: "JSON null", json: []byte(`null`), wantTolerant: 0},
		{name: "empty JSON string", json: []byte(`""`), wantTolerant: 0},
		{name: "0", json: []byte(`0`), wantTolerant: 0},
		{name: "-0.12345", json: []byte(`-0.12345`), wantTolerant: -0.12345},
		{name: "naked NaN", json: []byte(`NaN`), wantTolerant: Tolerant(math.NaN())},
		{name: "quoted NaN", json: []byte(`"NaN"`), wantTolerant: Tolerant(math.NaN())},
		{name: "naked Inf", json: []byte(`Inf`), wantTolerant: Tolerant(math.Inf(1))},
		{name: "quoted Inf", json: []byte(`"Inf"`), wantTolerant: Tolerant(math.Inf(1))},
		{name: "naked +Inf", json: []byte(`+Inf`), wantTolerant: Tolerant(math.Inf(1))},
		{name: "quoted +Inf", json: []byte(`"+Inf"`), wantTolerant: Tolerant(math.Inf(1))},
		{name: "naked -Inf", json: []byte(`-Inf`), wantTolerant: Tolerant(math.Inf(-1))},
		{name: "quoted -Inf", json: []byte(`"-Inf"`), wantTolerant: Tolerant(math.Inf(-1))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var amount Tolerant = 666 // Init with value different from default 0
			err := amount.UnmarshalJSON(tt.json)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tolerant.UnmarshalJSON(%s) error = %v, wantErr %v", tt.json, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalInclNaN(float64(amount), float64(tt.wantTolerant)) {
				t.Errorf("Tolerant.UnmarshalJSON(%s) got %f, want %f", tt.json, amount, tt.wantTolerant)
			}
		})
	}
}

func equalInclNaN(a, b float64) bool {
	if math.IsNaN(a) {
		return math.IsNaN(b)
	}
	if math.IsNaN(b) {
		return false
	}
	return a == b
}

// TestTolerant_UnmarshalJSON_escapedString protects the reason JSON strings
// are unquoted with strconv.Unquote instead of just trimming the quotes:
// escaped characters inside the string must be decoded before parsing.
func TestTolerant_UnmarshalJSON_escapedString(t *testing.T) {
	var f Tolerant
	err := f.UnmarshalJSON([]byte(`"1,5"`)) // "1,5" with an escaped comma
	if err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if f != 1.5 {
		t.Errorf("UnmarshalJSON(`\"1\\u002c5\"`) = %v, want 1.5", f)
	}

	// German decimal comma in a quoted string, as found in non standard JSON
	err = f.UnmarshalJSON([]byte(`"1.234,56"`))
	if err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if f != 1234.56 {
		t.Errorf("UnmarshalJSON(`\"1.234,56\"`) = %v, want 1234.56", f)
	}
}

func TestTolerant_UnmarshalJSON_invalidQuotedString(t *testing.T) {
	for _, j := range []string{`"1.5`, `"\x"`} {
		var f Tolerant = 666
		err := f.UnmarshalJSON([]byte(j))
		if err == nil {
			t.Errorf("UnmarshalJSON(%s) expected error for malformed JSON string", j)
		}
		if f != 666 {
			t.Errorf("UnmarshalJSON(%s) must not modify the value on error, got %v", j, f)
		}
	}
}

// TestTolerant_AsFloatPtr protects that the returned pointer aliases the
// Tolerant value, so code writing through it modifies the original.
func TestTolerant_AsFloatPtr(t *testing.T) {
	f := Tolerant(1.5)
	ptr := f.AsFloatPtr()
	if *ptr != 1.5 {
		t.Errorf("*AsFloatPtr() = %v, want 1.5", *ptr)
	}
	*ptr = 2.5
	if f != 2.5 {
		t.Errorf("writing through AsFloatPtr() must modify the Tolerant, got %v", f)
	}
}

// TestTolerant_validity protects that NaN and ±Inf (which UnmarshalJSON
// accepts) are never reported as valid, and that the sign checks require
// validity first so that +Inf is not "greater zero".
func TestTolerant_validity(t *testing.T) {
	tests := []struct {
		name         string
		f            Tolerant
		valid        bool
		greaterZero  bool
		smallerZero  bool
		hasPosSign   bool
		hasNegSign   bool
		hasAnySign   bool
		isNaN, isInf bool
	}{
		{name: "zero", f: 0, valid: true, hasPosSign: true, hasAnySign: true},
		{name: "positive", f: 1.5, valid: true, greaterZero: true, hasPosSign: true, hasAnySign: true},
		{name: "negative", f: -1.5, valid: true, smallerZero: true, hasNegSign: true, hasAnySign: true},
		{name: "NaN", f: Tolerant(math.NaN()), isNaN: true},
		{name: "+Inf", f: Tolerant(math.Inf(1)), isInf: true},
		{name: "-Inf", f: Tolerant(math.Inf(-1)), isInf: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Valid(); got != tt.valid {
				t.Errorf("Valid() = %v, want %v", got, tt.valid)
			}
			if got := tt.f.ValidAndGreaterZero(); got != tt.greaterZero {
				t.Errorf("ValidAndGreaterZero() = %v, want %v", got, tt.greaterZero)
			}
			if got := tt.f.ValidAndSmallerZero(); got != tt.smallerZero {
				t.Errorf("ValidAndSmallerZero() = %v, want %v", got, tt.smallerZero)
			}
			if got := tt.f.ValidAndHasSign(+1); got != tt.hasPosSign {
				t.Errorf("ValidAndHasSign(+1) = %v, want %v", got, tt.hasPosSign)
			}
			if got := tt.f.ValidAndHasSign(-1); got != tt.hasNegSign {
				t.Errorf("ValidAndHasSign(-1) = %v, want %v", got, tt.hasNegSign)
			}
			if got := tt.f.ValidAndHasSign(0); got != tt.hasAnySign {
				t.Errorf("ValidAndHasSign(0) = %v, want %v", got, tt.hasAnySign)
			}
			if got := tt.f.IsNaN(); got != tt.isNaN {
				t.Errorf("IsNaN() = %v, want %v", got, tt.isNaN)
			}
			if got := tt.f.IsInf(); got != tt.isInf {
				t.Errorf("IsInf() = %v, want %v", got, tt.isInf)
			}
		})
	}
}
