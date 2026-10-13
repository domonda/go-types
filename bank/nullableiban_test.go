package bank

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/country"
)

func TestNullableIBAN_Null(t *testing.T) {
	// NULL is a valid value for a NullableIBAN and must pass validation
	// and normalization unchanged without errors
	var iban NullableIBAN
	assert.True(t, iban.IsNull(), "zero value must be NULL")
	assert.False(t, iban.IsNotNull())
	assert.True(t, iban.Valid())
	assert.NoError(t, iban.Validate())
	assert.False(t, iban.ValidAndNotNull())
	assert.True(t, iban.ValidAndNormalized())
	assert.Equal(t, country.Code(""), iban.CountryCode())

	norm, err := iban.Normalized()
	require.NoError(t, err)
	assert.Equal(t, IBANNull, norm)
	spaced, err := iban.NormalizedWithSpaces()
	require.NoError(t, err)
	assert.Equal(t, IBANNull, spaced)
	assert.Equal(t, IBANNull, iban.NormalizedOrNull())

	assert.Equal(t, "", iban.String())
	assert.Equal(t, "default", iban.StringOr("default"))
	assert.Equal(t, IBAN("DE89370400440532013000"), iban.GetOr("DE89370400440532013000"))
	assert.Panics(t, func() { iban.Get() }, "Get on NULL must panic instead of returning an invalid IBAN")
}

func TestNullableIBAN_Valid(t *testing.T) {
	const spacedIBAN NullableIBAN = "DE89 3704 0044 0532 0130 00"
	const normIBAN NullableIBAN = "DE89370400440532013000"

	assert.True(t, spacedIBAN.Valid())
	assert.True(t, spacedIBAN.ValidAndNotNull())
	assert.NoError(t, spacedIBAN.Validate())
	assert.False(t, spacedIBAN.ValidAndNormalized())
	assert.True(t, normIBAN.ValidAndNormalized())
	assert.Equal(t, country.DE, spacedIBAN.CountryCode())
	// Valid accepts leading whitespace, so CountryCode must read the
	// country of the normalized IBAN like IBAN.CountryCode,
	// not the first two raw characters (" D").
	const leadingSpaceIBAN NullableIBAN = " DE89 3704 0044 0532 0130 00"
	require.True(t, leadingSpaceIBAN.Valid())
	assert.Equal(t, IBAN(leadingSpaceIBAN).CountryCode(), leadingSpaceIBAN.CountryCode())
	assert.Equal(t, country.DE, leadingSpaceIBAN.CountryCode())

	norm, err := spacedIBAN.Normalized()
	require.NoError(t, err)
	assert.Equal(t, normIBAN, norm)
	assert.Equal(t, normIBAN, spacedIBAN.NormalizedOrNull())
	assert.Equal(t, string(normIBAN), spacedIBAN.String())

	spaced, err := normIBAN.NormalizedWithSpaces()
	require.NoError(t, err)
	assert.Equal(t, spacedIBAN, spaced)
}

func TestNullableIBAN_WhitespaceIsNull(t *testing.T) {
	// Like country.NullableCode and vat.NullableID, a whitespace only
	// value is null, else a value that carries no IBAN would be handled
	// as an invalid IBAN and written to SQL and JSON as whitespace.
	for _, iban := range []NullableIBAN{" ", "  ", "\t\n", "\u00a0"} {
		t.Run(fmt.Sprintf("%q", string(iban)), func(t *testing.T) {
			assert.True(t, iban.IsNull())
			assert.False(t, iban.IsNotNull())
			assert.True(t, iban.Valid())
			assert.NoError(t, iban.Validate())
			assert.False(t, iban.ValidAndNotNull())
			assert.False(t, iban.ValidAndNormalized(), "whitespace is null, but not the normalized IBANNull")
			assert.Equal(t, country.Code(""), iban.CountryCode())

			norm, err := iban.Normalized()
			require.NoError(t, err)
			assert.Equal(t, IBANNull, norm)
			spaced, err := iban.NormalizedWithSpaces()
			require.NoError(t, err)
			assert.Equal(t, IBANNull, spaced)
			assert.Equal(t, IBANNull, iban.NormalizedOrNull())

			assert.Equal(t, "", iban.String())
			assert.Equal(t, "default", iban.StringOr("default"))
			assert.Equal(t, IBAN("DE89370400440532013000"), iban.GetOr("DE89370400440532013000"))
			assert.Panics(t, func() { iban.Get() })

			v, err := iban.Value()
			require.NoError(t, err)
			assert.Nil(t, v, "must be written as SQL NULL instead of the whitespace")
			j, err := json.Marshal(iban)
			require.NoError(t, err)
			assert.Equal(t, `null`, string(j))
		})
	}
}

func TestNullableIBAN_Invalid(t *testing.T) {
	// Only the empty or whitespace only string is NULL,
	// everything else must be a valid IBAN
	for _, iban := range []NullableIBAN{"NULL", "DE89370400440532013001", "XX89370400440532013000"} {
		t.Run(string(iban), func(t *testing.T) {
			assert.False(t, iban.IsNull())
			assert.False(t, iban.Valid())
			assert.False(t, iban.ValidAndNotNull())
			assert.Error(t, iban.Validate())
			assert.False(t, iban.ValidAndNormalized())
			assert.Equal(t, country.Code(""), iban.CountryCode())

			norm, err := iban.Normalized()
			assert.Error(t, err)
			assert.Equal(t, iban, norm, "invalid IBAN must be returned unchanged")
			spaced, err := iban.NormalizedWithSpaces()
			assert.Error(t, err)
			assert.Equal(t, iban, spaced, "invalid IBAN must be returned unchanged")
			assert.Equal(t, IBANNull, iban.NormalizedOrNull())
			assert.Equal(t, string(iban), iban.String())
		})
	}
}

func TestNullableIBAN_SetGet(t *testing.T) {
	var iban NullableIBAN
	iban.Set("DE89370400440532013000")
	assert.True(t, iban.IsNotNull())
	assert.Equal(t, IBAN("DE89370400440532013000"), iban.Get())
	assert.Equal(t, IBAN("DE89370400440532013000"), iban.GetOr("AT611904300234573201"))
	assert.Equal(t, "DE89370400440532013000", iban.StringOr("default"))

	iban.SetNull()
	assert.True(t, iban.IsNull())

	// Conversions between IBAN and NullableIBAN
	assert.Equal(t, NullableIBAN("DE89370400440532013000"), IBAN("DE89370400440532013000").Nullable())
	assert.True(t, IBAN("").Nullable().IsNull())
}

func TestNullableIBAN_ScanString(t *testing.T) {
	// Textual NULL representations from CSV or form input must become NULL
	for _, source := range []string{"", "NULL", "null", "nil"} {
		t.Run(source, func(t *testing.T) {
			iban := NullableIBAN("DE89370400440532013000")
			require.NoError(t, iban.ScanString(source, true))
			assert.True(t, iban.IsNull())
		})
	}

	t.Run("normalizes valid IBAN", func(t *testing.T) {
		for _, validate := range []bool{true, false} {
			var iban NullableIBAN
			require.NoError(t, iban.ScanString("DE89 3704 0044 0532 0130 00", validate))
			assert.Equal(t, NullableIBAN("DE89370400440532013000"), iban)
		}
	})
	t.Run("validate rejects invalid IBAN", func(t *testing.T) {
		iban := NullableIBAN("unchanged")
		require.Error(t, iban.ScanString("DE89370400440532013001", true))
		assert.Equal(t, NullableIBAN("unchanged"), iban)
	})
	t.Run("without validate assigns invalid IBAN as is", func(t *testing.T) {
		var iban NullableIBAN
		require.NoError(t, iban.ScanString("DE89 3704 0044 0532 0130 01", false))
		assert.Equal(t, NullableIBAN("DE89 3704 0044 0532 0130 01"), iban)
	})
}

// StringOr returns the normalized value like String, but the alternative
// instead of a value that can't be normalized, so a caller printing
// StringOr never shows a misspelled or invalid value as if it were valid.
func TestNullableIBAN_StringOr(t *testing.T) {
	assert.Equal(t, "none", IBANNull.StringOr("none"))
	assert.Equal(t, "DE89370400440532013000", NullableIBAN("DE89 3704 0044 0532 0130 00").StringOr("none"))
	assert.Equal(t, "none", NullableIBAN("DE00 1234").StringOr("none"))
}

func TestNullableIBAN_SQL(t *testing.T) {
	t.Run("NULL is written as nil", func(t *testing.T) {
		// NULL must be written to SQL as nil, not as an empty string
		v, err := IBANNull.Value()
		require.NoError(t, err)
		assert.Nil(t, v)
	})
	t.Run("value is written as string", func(t *testing.T) {
		v, err := NullableIBAN("DE89370400440532013000").Value()
		require.NoError(t, err)
		assert.Equal(t, "DE89370400440532013000", v)
	})

	tests := []struct {
		name    string
		value   any
		want    NullableIBAN
		wantErr bool
	}{
		{name: "SQL NULL", value: nil, want: IBANNull},
		{name: "string", value: "DE89370400440532013000", want: "DE89370400440532013000"},
		{name: "bytes", value: []byte("DE89370400440532013000"), want: "DE89370400440532013000"},
		{name: "unsupported type", value: int64(1), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iban := NullableIBAN("unchanged")
			err := iban.Scan(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, NullableIBAN("unchanged"), iban)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, iban)
		})
	}

	t.Run("Scan copies bytes", func(t *testing.T) {
		// database/sql may reuse the []byte buffer after Scan returns
		buf := []byte("DE89370400440532013000")
		var iban NullableIBAN
		require.NoError(t, iban.Scan(buf))
		copy(buf, "XXXXXXXXXXXXXXXXXXXXXX")
		assert.Equal(t, NullableIBAN("DE89370400440532013000"), iban)
	})
}

func TestNullableIBAN_JSON(t *testing.T) {
	type obj struct {
		IBAN NullableIBAN `json:"iban"`
	}

	j, err := json.Marshal(obj{IBAN: IBANNull})
	require.NoError(t, err)
	assert.Equal(t, `{"iban":null}`, string(j), "NULL must be marshalled as JSON null, not as empty string")
	var got obj
	require.NoError(t, json.Unmarshal(j, &got))
	assert.True(t, got.IBAN.IsNull())

	j, err = json.Marshal(obj{IBAN: "DE89370400440532013000"})
	require.NoError(t, err)
	assert.Equal(t, `{"iban":"DE89370400440532013000"}`, string(j))
	require.NoError(t, json.Unmarshal(j, &got))
	assert.Equal(t, NullableIBAN("DE89370400440532013000"), got.IBAN)
}

// A non-null NullableIBAN is written normalized like IBAN, see
// TestIBAN_ValueAndMarshalJSONNormalize, and null stays SQL NULL and
// JSON null, including the whitespace-only null.
func TestNullableIBAN_ValueAndMarshalJSONNormalize(t *testing.T) {
	tests := []struct {
		iban     NullableIBAN
		wantSQL  any
		wantJSON string
	}{
		{iban: IBANNull, wantSQL: nil, wantJSON: `null`},
		{iban: "  ", wantSQL: nil, wantJSON: `null`},
		{iban: "DE89 3704 0044 0532 0130 00", wantSQL: "DE89370400440532013000", wantJSON: `"DE89370400440532013000"`},
		{iban: "DE00 1234", wantSQL: "DE00 1234", wantJSON: `"DE00 1234"`},
	}
	for _, tt := range tests {
		v, err := tt.iban.Value()
		require.NoError(t, err)
		assert.Equal(t, tt.wantSQL, v, "NullableIBAN(%q).Value()", string(tt.iban))

		j, err := json.Marshal(tt.iban)
		require.NoError(t, err)
		assert.Equal(t, tt.wantJSON, string(j), "json.Marshal(NullableIBAN(%q))", string(tt.iban))
	}
}

func TestNullableIBAN_UnmarshalJSON(t *testing.T) {
	type obj struct {
		IBAN NullableIBAN `json:"iban"`
	}
	// JSON null must reset a field that holds a value, like for
	// uu.NullableID and email.NullableAddress, else decoding a partial
	// update with null into an existing struct keeps the old IBAN.
	o := obj{IBAN: "DE89370400440532013000"}
	require.NoError(t, json.Unmarshal([]byte(`{"iban":null}`), &o))
	assert.Equal(t, IBANNull, o.IBAN)

	// A JSON string is assigned unchanged, without normalization or
	// validation, like encoding/json did before NullableIBAN had
	// UnmarshalJSON.
	for _, str := range []string{"DE89370400440532013000", "DE89 3704 0044 0532 0130 00", "invalid", ""} {
		o := obj{IBAN: "AT611904300234573201"}
		data, err := json.Marshal(map[string]string{"iban": str})
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, &o))
		assert.Equal(t, NullableIBAN(str), o.IBAN, "json.Unmarshal(%s)", data)
	}

	// Any other JSON value is an error and leaves the value unchanged.
	for _, data := range []string{`123`, `true`, `{}`, `["DE89370400440532013000"]`} {
		iban := NullableIBAN("DE89370400440532013000")
		assert.Error(t, json.Unmarshal([]byte(data), &iban), "json.Unmarshal(%s)", data)
		assert.Equal(t, NullableIBAN("DE89370400440532013000"), iban, "json.Unmarshal(%s)", data)
	}
}

func TestNullableIBAN_JSONSchema(t *testing.T) {
	// The schema must accept JSON null because NULL is marshalled as null
	s := IBANNull.JSONSchema()
	require.Len(t, s.OneOf, 2)
	assert.Equal(t, "string", s.OneOf[0].Type)
	assert.Equal(t, IBANRegex, s.OneOf[0].Pattern)
	assert.Equal(t, "null", s.OneOf[1].Type)
}
