package types

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestLenString_Validate checks that Validate rejects both inconsistent
// constraints and strings outside of the inclusive [min, max] length range,
// because a LenString is only meaningful if its bounds make sense and hold.
func TestLenString_Validate(t *testing.T) {
	tests := []struct {
		name    string
		s       *LenString
		wantErr string // substring of the expected error, empty for valid
	}{
		{name: "nil", s: nil, wantErr: "nil LenString"},
		{name: "empty string with zero bounds", s: NewLenString("", 0, 0)},
		{name: "exactly min length", s: NewLenString("ab", 2, 4)},
		{name: "exactly max length", s: NewLenString("abcd", 2, 4)},
		{name: "between min and max", s: NewLenString("abc", 2, 4)},
		{name: "min equals max", s: NewLenString("abc", 3, 3)},
		{name: "one shorter than min", s: NewLenString("a", 2, 4), wantErr: "shorter than minimum of 2"},
		{name: "one longer than max", s: NewLenString("abcde", 2, 4), wantErr: "longer than maximum of 4"},
		{name: "negative min", s: NewLenString("a", -1, 4), wantErr: "negative minimum length -1"},
		{name: "negative max", s: NewLenString("a", 0, -1), wantErr: "negative maximum length -1"},
		{name: "min greater than max", s: NewLenString("abc", 4, 2), wantErr: "minimum length 4 is greater than maximum length 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() = %q, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestLenString_ImplementsValidatErr checks that a LenString pointer is
// picked up by the generic Validate function, so it gets validated
// together with other values, e.g. by DeepValidate.
func TestLenString_ImplementsValidatErr(t *testing.T) {
	var _ ValidatErr = (*LenString)(nil)
	if err := Validate(NewLenString("too long", 0, 3)); err == nil {
		t.Error("Validate(*LenString) with too long string = nil, want error")
	}
}

// TestNewLenString_DoesNotValidate checks the documented difference to
// MustLenString: NewLenString must not reject out-of-bounds input,
// so callers can construct first and validate separately.
func TestNewLenString_DoesNotValidate(t *testing.T) {
	s := NewLenString("too long", 0, 3)
	if s == nil {
		t.Fatal("NewLenString() = nil")
	}
	if s.String() != "too long" || s.MinLen() != 0 || s.MaxLen() != 3 {
		t.Errorf("NewLenString() = {%q, %d, %d}, want {%q, 0, 3}", s.String(), s.MinLen(), s.MaxLen(), "too long")
	}
	if s.Validate() == nil {
		t.Error("Validate() of out-of-bounds NewLenString = nil, want error")
	}
}

// TestMustLenString checks that MustLenString only returns values that
// satisfy their constraints and panics otherwise.
func TestMustLenString(t *testing.T) {
	s := MustLenString("abc", 1, 3)
	if s.String() != "abc" || s.MinLen() != 1 || s.MaxLen() != 3 {
		t.Errorf("MustLenString() = {%q, %d, %d}, want {%q, 1, 3}", s.String(), s.MinLen(), s.MaxLen(), "abc")
	}

	for _, tt := range []struct {
		name     string
		str      string
		min, max int
	}{
		{name: "too long", str: "abcd", min: 1, max: 3},
		{name: "too short", str: "", min: 1, max: 3},
		{name: "invalid bounds", str: "abc", min: 3, max: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("MustLenString(%q, %d, %d) did not panic", tt.str, tt.min, tt.max)
				}
			}()
			MustLenString(tt.str, tt.min, tt.max)
		})
	}
}

// TestLenString_String checks that a nil *LenString can be printed
// without panicking, as fmt.Stringer implementations should.
func TestLenString_String(t *testing.T) {
	var s *LenString
	if got := s.String(); got != "<nil>" {
		t.Errorf("nil String() = %q, want %q", got, "<nil>")
	}
	if got := NewLenString("abc", 0, 5).String(); got != "abc" {
		t.Errorf("String() = %q, want %q", got, "abc")
	}
}

// TestLenString_SetString checks that SetString enforces the length bounds
// and leaves the previous value untouched on rejection,
// so a failed assignment can't leave an invalid value behind.
func TestLenString_SetString(t *testing.T) {
	s := MustLenString("abc", 2, 4)

	for _, invalid := range []string{"a", "abcde"} {
		if err := s.SetString(invalid); err == nil {
			t.Errorf("SetString(%q) = nil, want error", invalid)
		}
		if s.String() != "abc" {
			t.Errorf("after rejected SetString(%q) String() = %q, want unchanged %q", invalid, s.String(), "abc")
		}
	}

	for _, valid := range []string{"ab", "abcd"} {
		if err := s.SetString(valid); err != nil {
			t.Errorf("SetString(%q) = %v, want nil", valid, err)
		}
		if s.String() != valid {
			t.Errorf("after SetString(%q) String() = %q", valid, s.String())
		}
	}
}

// TestLenString_ScanString checks the documented validate flag:
// with validate=true out-of-bounds input is rejected and the value kept,
// with validate=false the input is assigned as is.
func TestLenString_ScanString(t *testing.T) {
	s := MustLenString("abc", 2, 4)

	if err := s.ScanString("abcde", true); err == nil {
		t.Error("ScanString(too long, validate=true) = nil, want error")
	}
	if s.String() != "abc" {
		t.Errorf("after rejected ScanString String() = %q, want unchanged %q", s.String(), "abc")
	}

	if err := s.ScanString("abcd", true); err != nil {
		t.Errorf("ScanString(valid, validate=true) = %v, want nil", err)
	}
	if s.String() != "abcd" {
		t.Errorf("after ScanString(valid) String() = %q, want %q", s.String(), "abcd")
	}

	if err := s.ScanString("abcde", false); err != nil {
		t.Errorf("ScanString(too long, validate=false) = %v, want nil", err)
	}
	if s.String() != "abcde" {
		t.Errorf("after ScanString(validate=false) String() = %q, want %q", s.String(), "abcde")
	}
	if s.Validate() == nil {
		t.Error("Validate() after unvalidated ScanString of too long string = nil, want error")
	}
}

// TestLenString_MarshalText checks that a nil *LenString marshals to no text
// without error and that text round-trips through UnmarshalText.
func TestLenString_MarshalText(t *testing.T) {
	var nilStr *LenString
	text, err := nilStr.MarshalText()
	if err != nil || text != nil {
		t.Errorf("nil MarshalText() = %q, %v, want nil, nil", text, err)
	}

	src := MustLenString("abc", 1, 5)
	text, err = src.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error: %v", err)
	}
	if string(text) != "abc" {
		t.Errorf("MarshalText() = %q, want %q", text, "abc")
	}

	dst := NewLenString("", 1, 5)
	if err := dst.UnmarshalText(text); err != nil {
		t.Fatalf("UnmarshalText(%q) error: %v", text, err)
	}
	if dst.String() != src.String() {
		t.Errorf("text round-trip = %q, want %q", dst.String(), src.String())
	}
}

// TestLenString_UnmarshalText checks that text input is length validated,
// since UnmarshalText is how untrusted input (e.g. form values) gets in.
func TestLenString_UnmarshalText(t *testing.T) {
	s := MustLenString("abc", 2, 4)
	for _, invalid := range []string{"", "a", "abcde"} {
		if err := s.UnmarshalText([]byte(invalid)); err == nil {
			t.Errorf("UnmarshalText(%q) = nil, want error", invalid)
		}
		if s.String() != "abc" {
			t.Errorf("after rejected UnmarshalText(%q) String() = %q, want unchanged %q", invalid, s.String(), "abc")
		}
	}
}

// TestLenString_MarshalJSON checks that a nil *LenString marshals as JSON null
// and a value as a properly escaped JSON string, also as a struct field.
func TestLenString_MarshalJSON(t *testing.T) {
	var nilStr *LenString
	j, err := nilStr.MarshalJSON()
	if err != nil || string(j) != "null" {
		t.Errorf("nil MarshalJSON() = %s, %v, want null, nil", j, err)
	}

	s := MustLenString(`a"b`, 0, 5)
	j, err = s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}
	if want := `"a\"b"`; string(j) != want {
		t.Errorf("MarshalJSON() = %s, want %s", j, want)
	}

	j, err = json.Marshal(&struct{ S *LenString }{S: &s})
	if err != nil {
		t.Fatalf("json.Marshal(struct) error: %v", err)
	}
	if want := `{"S":"a\"b"}`; string(j) != want {
		t.Errorf("json.Marshal(struct) = %s, want %s", j, want)
	}
}

// TestLenString_UnmarshalJSON_NullAndInvalid checks that JSON null is a no-op
// keeping the existing value and that non-string JSON is rejected.
func TestLenString_UnmarshalJSON_NullAndInvalid(t *testing.T) {
	s := MustLenString("abc", 0, 5)
	if err := s.UnmarshalJSON([]byte("null")); err != nil {
		t.Errorf("UnmarshalJSON(null) = %v, want nil", err)
	}
	if s.String() != "abc" {
		t.Errorf("after UnmarshalJSON(null) String() = %q, want unchanged %q", s.String(), "abc")
	}

	for _, invalid := range []string{`123`, `{}`, `"unterminated`} {
		if err := s.UnmarshalJSON([]byte(invalid)); err == nil {
			t.Errorf("UnmarshalJSON(%s) = nil, want error", invalid)
		}
	}
}

// TestLenString_UnmarshalJSON checks that a JSON string is assigned when its
// length is valid and rejected otherwise, leaving the previous value intact,
// because JSON is how untrusted input (e.g. API requests) gets in and a
// LenString must never hold a string outside of its length bounds.
func TestLenString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		min, max int
		valid    string
		invalid  []string
	}{
		{name: "min>0", min: 2, max: 4, valid: "xyz", invalid: []string{"", "a", "abcde"}},
		{name: "min==0", min: 0, max: 4, valid: "xyz", invalid: []string{"abcde"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MustLenString("abc", tt.min, tt.max)
			if err := s.UnmarshalJSON([]byte(`"` + tt.valid + `"`)); err != nil {
				t.Fatalf("UnmarshalJSON(%q) = %v, want nil", tt.valid, err)
			}
			if s.String() != tt.valid {
				t.Errorf("after UnmarshalJSON(%q) String() = %q, want %q", tt.valid, s.String(), tt.valid)
			}

			for _, invalid := range tt.invalid {
				if err := s.UnmarshalJSON([]byte(`"` + invalid + `"`)); err == nil {
					t.Errorf("UnmarshalJSON(%q) = nil, want error", invalid)
				}
				if s.String() != tt.valid {
					t.Errorf("after rejected UnmarshalJSON(%q) String() = %q, want unchanged %q", invalid, s.String(), tt.valid)
				}
			}

			// Same through encoding/json as struct field
			v := struct{ S *LenString }{S: &s}
			if err := json.Unmarshal([]byte(`{"S":"ab"}`), &v); err != nil {
				t.Fatalf("json.Unmarshal(struct) error: %v", err)
			}
			if s.String() != "ab" {
				t.Errorf("after json.Unmarshal(struct) String() = %q, want %q", s.String(), "ab")
			}
		})
	}
}
