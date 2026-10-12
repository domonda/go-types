package vat

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/country"
)

// Null must be a typed constant like date.Null and country.Null,
// so that no code can reassign the null value of the package.
// This declaration does not compile if Null is a variable.
const _ NullableID = Null

func TestNullableID(t *testing.T) {
	var n NullableID
	assert.True(t, n.Valid(), "empty NullableID is valid")
	assert.NoError(t, n.Validate(), "empty NullableID is valid")

	assert.Empty(t, n.NormalizedOrNull())
}

func TestNullableID_Normalized(t *testing.T) {
	t.Run("Null is valid and stays Null", func(t *testing.T) {
		norm, err := Null.Normalized()
		require.NoError(t, err)
		assert.Equal(t, Null, norm)
	})
	t.Run("whitespace only is Null", func(t *testing.T) {
		// Like country.NullableCode, a whitespace only value is null,
		// so it must be valid and normalize to Null instead of
		// failing as a too short VAT ID.
		norm, err := NullableID(" \t\n").Normalized()
		require.NoError(t, err)
		assert.Equal(t, Null, norm)
	})
	t.Run("valid ID is normalized", func(t *testing.T) {
		norm, err := NullableID("at u 1022 3006").Normalized()
		require.NoError(t, err)
		assert.Equal(t, NullableID("ATU10223006"), norm)
	})
	t.Run("invalid ID wraps ErrInvalidID", func(t *testing.T) {
		_, err := NullableID("ATU12345678").Normalized()
		assert.ErrorIs(t, err, ErrInvalidID)
	})
}

func TestNullableID_NormalizedOrNull(t *testing.T) {
	assert.Equal(t, NullableID("ATU10223006"), NullableID("atu 10223006").NormalizedOrNull())
	assert.Equal(t, Null, NullableID("ATU12345678").NormalizedOrNull(), "invalid becomes Null")
	assert.Equal(t, Null, Null.NormalizedOrNull())
}

// NormalizedNotNull is for callers that require a VAT ID,
// so Null must be rejected like any other invalid ID.
func TestNullableID_NormalizedNotNull(t *testing.T) {
	id, err := NullableID("atu 10223006").NormalizedNotNull()
	require.NoError(t, err)
	assert.Equal(t, ID("ATU10223006"), id)

	_, err = Null.NormalizedNotNull()
	assert.ErrorIs(t, err, ErrInvalidID)

	_, err = NullableID("ATU12345678").NormalizedNotNull()
	assert.ErrorIs(t, err, ErrInvalidID)
}

func TestNullableID_NullHandling(t *testing.T) {
	assert.True(t, Null.IsNull())
	assert.False(t, Null.IsNotNull())
	assert.False(t, NullableID("ATU10223006").IsNull())
	assert.True(t, NullableID("ATU10223006").IsNotNull())
	// Normalized returns Null for whitespace only, so IsNull must agree
	// like for country.NullableCode, else a value that normalizes to Null
	// is handled as a real VAT ID.
	for _, n := range []NullableID{" ", "  ", "\t\n", "\u00a0"} {
		assert.True(t, n.IsNull(), "NullableID(%q).IsNull()", string(n))
		assert.False(t, n.IsNotNull(), "NullableID(%q).IsNotNull()", string(n))
		assert.Equal(t, "none", n.StringOr("none"), "NullableID(%q).StringOr", string(n))
		assert.Equal(t, ID("DE167015661"), n.GetOr("DE167015661"), "NullableID(%q).GetOr", string(n))
		assert.Panics(t, func() { n.Get() }, "NullableID(%q).Get", string(n))
	}

	var n NullableID
	n.Set("ATU10223006")
	assert.Equal(t, NullableID("ATU10223006"), n)
	assert.Equal(t, ID("ATU10223006"), n.Get())
	assert.Equal(t, ID("ATU10223006"), n.GetOr("DE167015661"))
	assert.Equal(t, "ATU10223006", n.StringOr("none"))

	n.SetNull()
	assert.True(t, n.IsNull())
	assert.Equal(t, ID("DE167015661"), n.GetOr("DE167015661"))
	assert.Equal(t, "none", n.StringOr("none"))
	// Get on Null must panic instead of returning an empty ID
	// that could be mistaken for a real value.
	assert.Panics(t, func() { n.Get() })
}

// The validation methods differ only in how they treat Null and
// non-normalized input. This table pins that matrix.
func TestNullableID_Validation(t *testing.T) {
	tests := []struct {
		name                   string
		n                      NullableID
		valid                  bool
		validAndNormalized     bool
		validAndNotNull        bool
		isNormalized           bool // ValidateIsNormalized() == nil
		isNormalizedAndNotNull bool // ValidateIsNormalizedAndNotNull() == nil
		wrapsErrInvalidID      bool // ValidateIsNormalized wraps ErrInvalidID
	}{
		{
			name:                   "Null",
			n:                      Null,
			valid:                  true,
			validAndNormalized:     true,
			validAndNotNull:        false,
			isNormalized:           true,
			isNormalizedAndNotNull: false,
		},
		{
			// Whitespace only is null, but not the normalized Null
			name:                   "whitespace only",
			n:                      "  ",
			valid:                  true,
			validAndNormalized:     false,
			validAndNotNull:        false,
			isNormalized:           false,
			isNormalizedAndNotNull: false,
		},
		{
			name:                   "valid normalized",
			n:                      "ATU10223006",
			valid:                  true,
			validAndNormalized:     true,
			validAndNotNull:        true,
			isNormalized:           true,
			isNormalizedAndNotNull: true,
		},
		{
			name:                   "valid not normalized",
			n:                      "atu 10223006",
			valid:                  true,
			validAndNormalized:     false,
			validAndNotNull:        true,
			isNormalized:           false,
			isNormalizedAndNotNull: false,
		},
		{
			name:                   "invalid",
			n:                      "ATU12345678",
			valid:                  false,
			validAndNormalized:     false,
			validAndNotNull:        false,
			isNormalized:           false,
			isNormalizedAndNotNull: false,
			wrapsErrInvalidID:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.n.Valid(), "Valid")
			assert.Equal(t, tt.valid, tt.n.Validate() == nil, "Validate")
			assert.Equal(t, tt.validAndNormalized, tt.n.ValidAndNormalized(), "ValidAndNormalized")
			assert.Equal(t, tt.validAndNotNull, tt.n.ValidAndNotNull(), "ValidAndNotNull")

			err := tt.n.ValidateIsNormalized()
			assert.Equal(t, tt.isNormalized, err == nil, "ValidateIsNormalized: %v", err)
			if err != nil {
				// Like ID.ValidateIsNormalized: only invalid IDs wrap ErrInvalidID.
				assert.Equal(t, tt.wrapsErrInvalidID, errors.Is(err, ErrInvalidID), "ValidateIsNormalized error: %v", err)
			}
			err = tt.n.ValidateIsNormalizedAndNotNull()
			assert.Equal(t, tt.isNormalizedAndNotNull, err == nil, "ValidateIsNormalizedAndNotNull: %v", err)
		})
	}
}

func TestNullableID_CountryCode(t *testing.T) {
	// Null and invalid IDs both map to the NULL country code,
	// so the result can be written to a nullable country column.
	assert.Equal(t, country.Null, Null.CountryCode())
	assert.Equal(t, country.Null, NullableID("ATU12345678").CountryCode())
	assert.Equal(t, country.NullableCode("AT"), NullableID("atu 10223006").CountryCode())
	assert.Equal(t, country.NullableCode("BE"), NullableID("EU372008134").CountryCode(), "MOSS maps to Belgium")
}

func TestNullableID_IsMOSS(t *testing.T) {
	assert.False(t, Null.IsMOSS())
	assert.True(t, NullableID("EU372008134").IsMOSS())
	assert.False(t, NullableID("ATU10223006").IsMOSS())
}

func TestNullableID_Number(t *testing.T) {
	assert.Equal(t, "", Null.Number())
	assert.Equal(t, "167015661", NullableID("DE 167 015 661").Number())
	assert.Equal(t, "", NullableID("ATU12345678").Number())
}

func TestNullableID_String(t *testing.T) {
	assert.Equal(t, "", Null.String())
	assert.Equal(t, "ATU10223006", NullableID("at u 1022 3006").String())
	assert.Equal(t, "atu 1234 5678", NullableID("atu 1234 5678").String(), "invalid input is returned unchanged")
}

func TestNullableID_ScanString(t *testing.T) {
	// Text sources such as CSV files spell NULL in several ways,
	// all of them must reset a previously set value to Null.
	for _, source := range []string{"", "NULL", "null", "nil"} {
		for _, validate := range []bool{true, false} {
			n := NullableID("ATU10223006")
			require.NoError(t, n.ScanString(source, validate), "ScanString(%q, %t)", source, validate)
			assert.Equal(t, Null, n, "ScanString(%q, %t)", source, validate)
		}
	}

	var n NullableID
	require.NoError(t, n.ScanString("de 167-015-661", true))
	assert.Equal(t, NullableID("DE167015661"), n, "valid input is stored normalized")

	err := n.ScanString("ATU12345678", true)
	assert.ErrorIs(t, err, ErrInvalidID)
	assert.Equal(t, NullableID("DE167015661"), n, "failed validation keeps the old value")

	require.NoError(t, n.ScanString("atu 1234 5678", false))
	assert.Equal(t, NullableID("atu 1234 5678"), n, "without validate invalid input is stored as is")
}

func TestNullableID_Scan(t *testing.T) {
	t.Run("NULL resets to Null", func(t *testing.T) {
		n := NullableID("ATU10223006")
		require.NoError(t, n.Scan(nil))
		assert.Equal(t, Null, n)
	})
	t.Run("string", func(t *testing.T) {
		var n NullableID
		require.NoError(t, n.Scan("ATU10223006"))
		assert.Equal(t, NullableID("ATU10223006"), n)
	})
	t.Run("empty string is Null", func(t *testing.T) {
		n := NullableID("ATU10223006")
		require.NoError(t, n.Scan(""))
		assert.True(t, n.IsNull())
	})
	t.Run("bytes are copied", func(t *testing.T) {
		// database/sql drivers may reuse the []byte buffer after Scan.
		buf := []byte("ATU10223006")
		var n NullableID
		require.NoError(t, n.Scan(buf))
		copy(buf, "XXXXXXXXXXX")
		assert.Equal(t, NullableID("ATU10223006"), n)
	})
	t.Run("unsupported type", func(t *testing.T) {
		var n NullableID
		assert.Error(t, n.Scan(int64(10223006)))
	})
}

func TestNullableID_Value(t *testing.T) {
	// Null must be written as SQL NULL, not as an empty string,
	// so that IS NULL queries and unique constraints work.
	v, err := Null.Value()
	require.NoError(t, err)
	assert.Nil(t, v)

	// Whitespace only is null, so it must also be written as SQL NULL
	// instead of storing the whitespace as VAT ID.
	v, err = NullableID(" \t ").Value()
	require.NoError(t, err)
	assert.Nil(t, v)

	v, err = NullableID("at u 1022 3006").Value()
	require.NoError(t, err)
	assert.Equal(t, "ATU10223006", v, "valid IDs are written normalized")
}

// Scan(Value()) must give back the same value, including NULL.
func TestNullableID_SQLRoundTrip(t *testing.T) {
	for _, n := range []NullableID{Null, "ATU10223006", "EU372008134"} {
		v, err := n.Value()
		require.NoError(t, err)
		var got NullableID
		require.NoError(t, got.Scan(v))
		assert.Equal(t, n, got)
	}
}

func TestNullableID_JSON(t *testing.T) {
	type doc struct {
		VAT NullableID `json:"vat"`
	}
	t.Run("Null is JSON null, not an empty string", func(t *testing.T) {
		data, err := json.Marshal(doc{VAT: Null})
		require.NoError(t, err)
		assert.JSONEq(t, `{"vat":null}`, string(data))

		var got doc
		require.NoError(t, json.Unmarshal(data, &got))
		assert.True(t, got.VAT.IsNull())
	})
	t.Run("whitespace only is JSON null", func(t *testing.T) {
		data, err := json.Marshal(doc{VAT: "  "})
		require.NoError(t, err)
		assert.JSONEq(t, `{"vat":null}`, string(data))
	})
	t.Run("value round-trip", func(t *testing.T) {
		data, err := json.Marshal(doc{VAT: "ATU10223006"})
		require.NoError(t, err)
		assert.JSONEq(t, `{"vat":"ATU10223006"}`, string(data))

		var got doc
		require.NoError(t, json.Unmarshal(data, &got))
		assert.Equal(t, NullableID("ATU10223006"), got.VAT)
	})
}

func TestNullableID_UnmarshalJSON(t *testing.T) {
	type doc struct {
		VAT NullableID `json:"vat"`
	}
	// JSON null must reset a field that holds a value, like for
	// uu.NullableID and email.NullableAddress, else decoding a partial
	// update with null into an existing struct keeps the old VAT ID.
	d := doc{VAT: "ATU10223006"}
	require.NoError(t, json.Unmarshal([]byte(`{"vat":null}`), &d))
	assert.Equal(t, Null, d.VAT)

	// A JSON string is assigned unchanged, without normalization or
	// validation, like encoding/json did before NullableID had
	// UnmarshalJSON.
	for _, str := range []string{"ATU10223006", "atu 1022 3006", "bad", ""} {
		d := doc{VAT: "DE167015661"}
		data, err := json.Marshal(map[string]string{"vat": str})
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, &d))
		assert.Equal(t, NullableID(str), d.VAT, "json.Unmarshal(%s)", data)
	}

	// Any other JSON value is an error and leaves the value unchanged.
	for _, data := range []string{`123`, `true`, `{}`, `["ATU10223006"]`} {
		n := NullableID("ATU10223006")
		assert.Error(t, json.Unmarshal([]byte(data), &n), "json.Unmarshal(%s)", data)
		assert.Equal(t, NullableID("ATU10223006"), n, "json.Unmarshal(%s)", data)
	}
}

func TestNullableID_JSONSchema(t *testing.T) {
	s := NullableID("").JSONSchema()
	require.Len(t, s.OneOf, 2)
	str, null := s.OneOf[0], s.OneOf[1]
	assert.Equal(t, "string", str.Type)
	require.NotNil(t, str.MinLength)
	require.NotNil(t, str.MaxLength)
	assert.Equal(t, uint64(IDMinLength), *str.MinLength, "must match the ID length limits")
	assert.Equal(t, uint64(IDMaxLength), *str.MaxLength, "must match the ID length limits")
	assert.Equal(t, "null", null.Type)

	// The default is Null and must appear as JSON null in the schema,
	// an empty string would not be a valid VAT ID.
	data, err := json.Marshal(s)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	def, ok := decoded["default"]
	assert.True(t, ok, "schema has a default")
	assert.Nil(t, def, "schema default is JSON null")
}
