package country

import (
	"encoding/json"
	"regexp"
	"testing"
)

func TestNullableCode_Normalized(t *testing.T) {
	tests := []struct {
		n       NullableCode
		want    NullableCode
		wantErr bool
	}{
		{n: Null, want: Null},
		// Whitespace is trimmed before the NULL check.
		{n: "  ", want: Null},
		{n: "DE", want: "DE"},
		{n: " de ", want: "DE"},
		{n: "A", want: "AT"},
		{n: "Österreich", want: "AT"},
		// Errors: the input is returned unchanged.
		{n: "xx", want: "xx", wantErr: true},
		{n: "Atlantis", want: "Atlantis", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.n), func(t *testing.T) {
			got, err := tt.n.Normalized()
			if (err != nil) != tt.wantErr {
				t.Fatalf("NullableCode(%q).Normalized() error = %v, wantErr %t", string(tt.n), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NullableCode(%q).Normalized() = %q, want %q", string(tt.n), string(got), string(tt.want))
			}
		})
	}
}

// NormalizedWithAltCodes must resolve the same alternative codes as Normalized
// and keep treating the empty string as Null.
func TestNullableCode_NormalizedWithAltCodes(t *testing.T) {
	tests := []struct {
		n       NullableCode
		want    NullableCode
		wantErr bool
	}{
		{n: Null, want: Null},
		{n: "SUI", want: "CH"},
		{n: " deutschland ", want: "DE"},
		{n: "fr", want: "FR"},
		{n: "xx", want: "xx", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.n), func(t *testing.T) {
			got, err := tt.n.NormalizedWithAltCodes()
			if (err != nil) != tt.wantErr {
				t.Fatalf("NullableCode(%q).NormalizedWithAltCodes() error = %v, wantErr %t", string(tt.n), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NullableCode(%q).NormalizedWithAltCodes() = %q, want %q", string(tt.n), string(got), string(tt.want))
			}
		})
	}
}

func TestNullableCode_NormalizedOrNull(t *testing.T) {
	tests := []struct {
		n    NullableCode
		want NullableCode
	}{
		{n: Null, want: Null},
		{n: "at", want: "AT"},
		// Unresolvable input becomes Null instead of being stored as garbage.
		{n: "Atlantis", want: Null},
	}
	for _, tt := range tests {
		if got := tt.n.NormalizedOrNull(); got != tt.want {
			t.Errorf("NullableCode(%q).NormalizedOrNull() = %q, want %q", string(tt.n), string(got), string(tt.want))
		}
	}
}

// The validation methods differ only in how they treat Null and
// non-normalized input. This table pins that matrix.
func TestNullableCode_Validation(t *testing.T) {
	tests := []struct {
		n                  NullableCode
		valid              bool // Valid() and Validate() == nil
		validAndNotNull    bool
		validAndNormalized bool
	}{
		{n: Null, valid: true, validAndNotNull: false, validAndNormalized: true},
		{n: "DE", valid: true, validAndNotNull: true, validAndNormalized: true},
		{n: "de", valid: true, validAndNotNull: true, validAndNormalized: false},
		{n: "Deutschland", valid: true, validAndNotNull: true, validAndNormalized: false},
		{n: "xx", valid: false, validAndNotNull: false, validAndNormalized: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.n), func(t *testing.T) {
			if got := tt.n.Valid(); got != tt.valid {
				t.Errorf("Valid() = %t, want %t", got, tt.valid)
			}
			if err := tt.n.Validate(); (err == nil) != tt.valid {
				t.Errorf("Validate() = %v, want valid %t", err, tt.valid)
			}
			if got := tt.n.ValidAndNotNull(); got != tt.validAndNotNull {
				t.Errorf("ValidAndNotNull() = %t, want %t", got, tt.validAndNotNull)
			}
			if got := tt.n.ValidAndNormalized(); got != tt.validAndNormalized {
				t.Errorf("ValidAndNormalized() = %t, want %t", got, tt.validAndNormalized)
			}
		})
	}
}

// IsEU and EnglishName delegate to Code, so Null has to behave like an
// invalid code there: not in the EU and without a name.
func TestNullableCode_IsEUAndEnglishName(t *testing.T) {
	if Null.IsEU() {
		t.Error("Null.IsEU() = true, want false")
	}
	if got := Null.EnglishName(); got != "" {
		t.Errorf("Null.EnglishName() = %q, want empty", got)
	}
	if !NullableCode("at").IsEU() {
		t.Error(`NullableCode("at").IsEU() = false, want true`)
	}
	if got := NullableCode("at").EnglishName(); got != "Austria" {
		t.Errorf(`NullableCode("at").EnglishName() = %q, want "Austria"`, got)
	}
}

func TestNullableCode_NullHandling(t *testing.T) {
	if !Null.IsNull() || Null.IsNotNull() {
		t.Error("Null must be IsNull and not IsNotNull")
	}
	if NullableCode("DE").IsNull() || !NullableCode("DE").IsNotNull() {
		t.Error("DE must be IsNotNull and not IsNull")
	}
	// Normalized returns Null for whitespace only, so IsNull must agree,
	// else a value that normalizes to Null is handled as a real code.
	for _, n := range []NullableCode{" ", "  ", "\t\n", " "} {
		if !n.IsNull() || n.IsNotNull() {
			t.Errorf("NullableCode(%q) must be IsNull and not IsNotNull like its Normalized value", string(n))
		}
	}

	var n NullableCode
	n.Set(DE)
	if n != "DE" {
		t.Errorf("after Set(DE) = %q", string(n))
	}
	if got := n.Get(); got != DE {
		t.Errorf("Get() = %q, want DE", string(got))
	}
	if got := n.GetOr(AT); got != DE {
		t.Errorf("GetOr(AT) = %q, want DE", string(got))
	}
	if got := n.StringOr("none"); got != "DE" {
		t.Errorf(`StringOr("none") = %q, want "DE"`, got)
	}

	n.SetNull()
	if !n.IsNull() {
		t.Errorf("after SetNull() = %q, want Null", string(n))
	}
	if got := n.GetOr(AT); got != AT {
		t.Errorf("Null.GetOr(AT) = %q, want AT", string(got))
	}
	if got := n.StringOr("none"); got != "none" {
		t.Errorf(`Null.StringOr("none") = %q, want "none"`, got)
	}

	// Get on Null must panic instead of returning the Invalid code
	// that could be mistaken for a real value.
	defer func() {
		if recover() == nil {
			t.Error("Null.Get() did not panic")
		}
	}()
	n.Get()
}

func TestNullableCode_String(t *testing.T) {
	tests := []struct {
		n    NullableCode
		want string
	}{
		{n: Null, want: ""},
		{n: "de", want: "DE"},
		{n: "Österreich", want: "AT"},
		// String has no error result, so unresolvable input
		// comes back unchanged instead of being lost.
		{n: "Atlantis", want: "Atlantis"},
	}
	for _, tt := range tests {
		if got := tt.n.String(); got != tt.want {
			t.Errorf("NullableCode(%q).String() = %q, want %q", string(tt.n), got, tt.want)
		}
	}
}

func TestNullableCode_Scan(t *testing.T) {
	t.Run("NULL resets to Null", func(t *testing.T) {
		n := NullableCode("DE")
		if err := n.Scan(nil); err != nil {
			t.Fatal(err)
		}
		if n != Null {
			t.Errorf("Scan(nil) = %q, want Null", string(n))
		}
	})
	t.Run("empty string is Null", func(t *testing.T) {
		n := NullableCode("DE")
		if err := n.Scan(""); err != nil {
			t.Fatal(err)
		}
		if !n.IsNull() {
			t.Errorf(`Scan("") = %q, want Null`, string(n))
		}
	})
	t.Run("string is stored raw", func(t *testing.T) {
		var n NullableCode
		if err := n.Scan("de"); err != nil {
			t.Fatal(err)
		}
		if n != "de" {
			t.Errorf(`Scan("de") = %q, want "de"`, string(n))
		}
	})
	t.Run("bytes are copied", func(t *testing.T) {
		// database/sql drivers may reuse the []byte buffer after Scan.
		buf := []byte("DE")
		var n NullableCode
		if err := n.Scan(buf); err != nil {
			t.Fatal(err)
		}
		copy(buf, "XX")
		if n != "DE" {
			t.Errorf("Scan([]byte) = %q, want DE after the buffer was reused", string(n))
		}
	})
	t.Run("unsupported type", func(t *testing.T) {
		var n NullableCode
		if err := n.Scan(int64(49)); err == nil {
			t.Error("Scan(int64) = nil, want an error")
		}
	})
}

func TestNullableCode_Value(t *testing.T) {
	tests := []struct {
		n    NullableCode
		want any
	}{
		// Null must be written as SQL NULL, not as an empty string.
		{n: Null, want: nil},
		// Whitespace only normalizes to Null, so it must also be written
		// as SQL NULL instead of storing the whitespace as country code.
		{n: "  ", want: nil},
		{n: "\t\n", want: nil},
		{n: "DE", want: "DE"},
		// Codes are written normalized so equal countries compare equal in SQL.
		{n: "de", want: "DE"},
		{n: "Schweiz", want: "CH"},
	}
	for _, tt := range tests {
		got, err := tt.n.Value()
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("NullableCode(%q).Value() = %#v, want %#v", string(tt.n), got, tt.want)
		}
	}
}

// Scan(Value()) must give back the same value, including NULL.
func TestNullableCode_SQLRoundTrip(t *testing.T) {
	for _, n := range []NullableCode{Null, "DE", "XK"} {
		v, err := n.Value()
		if err != nil {
			t.Fatal(err)
		}
		var got NullableCode
		if err := got.Scan(v); err != nil {
			t.Fatal(err)
		}
		if got != n {
			t.Errorf("Scan(Value()) of %q = %q", string(n), string(got))
		}
	}
}

func TestNullableCode_JSON(t *testing.T) {
	type doc struct {
		Country NullableCode `json:"country"`
	}
	tests := []struct {
		n        NullableCode
		wantJSON string
		wantBack NullableCode
	}{
		// Null is JSON null, not an empty string.
		{n: Null, wantJSON: `{"country":null}`, wantBack: Null},
		// Whitespace only normalizes to Null, so it is also JSON null.
		{n: "  ", wantJSON: `{"country":null}`, wantBack: Null},
		{n: "DE", wantJSON: `{"country":"DE"}`, wantBack: "DE"},
		// JSON output is normalized like Code.
		{n: "de", wantJSON: `{"country":"DE"}`, wantBack: "DE"},
	}
	for _, tt := range tests {
		t.Run(string(tt.n), func(t *testing.T) {
			data, err := json.Marshal(doc{Country: tt.n})
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("json.Marshal = %s, want %s", data, tt.wantJSON)
			}
			var got doc
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if got.Country != tt.wantBack {
				t.Errorf("json.Unmarshal(%s) = %q, want %q", data, string(got.Country), string(tt.wantBack))
			}
		})
	}
}

func TestNullableCode_UnmarshalJSON(t *testing.T) {
	type doc struct {
		Country NullableCode `json:"country"`
	}
	// JSON null must reset a field that holds a value, like for
	// uu.NullableID and email.NullableAddress, else decoding a partial
	// update with null into an existing struct keeps the old country.
	d := doc{Country: "DE"}
	if err := json.Unmarshal([]byte(`{"country":null}`), &d); err != nil {
		t.Fatal(err)
	}
	if d.Country != Null {
		t.Errorf("json.Unmarshal of null kept %q, want Null", string(d.Country))
	}

	// A JSON string is assigned unchanged, without normalization or
	// validation, like encoding/json did before NullableCode had
	// UnmarshalJSON, and the round trip of a code keeps it.
	for _, str := range []string{"AT", "at", "XX", ""} {
		d := doc{Country: "DE"}
		data, _ := json.Marshal(map[string]string{"country": str})
		if err := json.Unmarshal(data, &d); err != nil {
			t.Fatalf("json.Unmarshal(%s): %s", data, err)
		}
		if d.Country != NullableCode(str) {
			t.Errorf("json.Unmarshal(%s) = %q, want %q", data, string(d.Country), str)
		}
	}

	// Any other JSON value is an error and leaves the value unchanged.
	for _, data := range []string{`123`, `true`, `{}`, `["DE"]`} {
		n := NullableCode("DE")
		if err := json.Unmarshal([]byte(data), &n); err == nil {
			t.Errorf("json.Unmarshal(%s) returned no error", data)
		}
		if n != "DE" {
			t.Errorf("json.Unmarshal(%s) changed the value to %q", data, string(n))
		}
	}
}

func TestNullableCode_JSONSchema(t *testing.T) {
	s := NullableCode("").JSONSchema()
	if len(s.OneOf) != 2 {
		t.Fatalf("JSONSchema().OneOf has %d entries, want 2", len(s.OneOf))
	}
	str, null := s.OneOf[0], s.OneOf[1]
	if str.Type != "string" || null.Type != "null" {
		t.Errorf("JSONSchema().OneOf types = %q, %q, want string, null", str.Type, null.Type)
	}
	// The schema must accept every code this package produces.
	pattern := regexp.MustCompile(str.Pattern)
	for _, c := range Code("").Enums() {
		if !pattern.MatchString(string(c)) {
			t.Errorf("enum value %q does not match the schema pattern %s", string(c), str.Pattern)
		}
	}

	// The default is Null and must appear as JSON null in the schema,
	// an empty string would not match the code pattern.
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	def, ok := decoded["default"]
	if !ok || def != nil {
		t.Errorf("schema default = %#v (present: %t), want JSON null", def, ok)
	}
}

func TestNullableCode_ScanString(t *testing.T) {
	// Text sources such as CSV files spell NULL in several ways,
	// all of them must reset a previously set value to Null.
	for _, source := range []string{"", "NULL", "null", "nil"} {
		for _, validate := range []bool{true, false} {
			n := NullableCode("DE")
			if err := n.ScanString(source, validate); err != nil {
				t.Fatalf("ScanString(%q, %t) error: %v", source, validate, err)
			}
			if n != Null {
				t.Errorf("ScanString(%q, %t) = %q, want Null", source, validate, string(n))
			}
		}
	}

	var n NullableCode
	if err := n.ScanString(" österreich ", true); err != nil {
		t.Fatal(err)
	}
	if n != "AT" {
		t.Errorf("valid input must be stored normalized, got %q", string(n))
	}

	if err := n.ScanString("Atlantis", true); err == nil {
		t.Error(`ScanString("Atlantis", true) = nil, want an error`)
	}
	if n != "AT" {
		t.Errorf("failed ScanString changed the value to %q", string(n))
	}

	if err := n.ScanString("Atlantis", false); err != nil {
		t.Fatal(err)
	}
	if n != "Atlantis" {
		t.Errorf(`ScanString("Atlantis", false) = %q, want the input unchanged`, string(n))
	}
}
