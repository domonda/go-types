package email

import (
	"database/sql/driver"
	"encoding/json"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNullableAddress_Null protects that the NULL value (empty string)
// is a valid, already normalized value for which every accessor returns
// an empty result without error, so optional addresses like Reply-To
// can be handled without special casing.
func TestNullableAddress_Null(t *testing.T) {
	var n NullableAddress

	assert.True(t, n.IsNull())
	assert.False(t, n.IsNotNull())
	assert.Equal(t, AddressNull, n)
	assert.NoError(t, n.Validate())
	assert.True(t, n.ValidAndNormalized())

	parsed, err := n.Parse()
	assert.NoError(t, err)
	assert.Nil(t, parsed)

	normalized, err := n.Normalized()
	assert.NoError(t, err)
	assert.Equal(t, AddressNull, normalized)

	name, err := n.NamePart()
	assert.NoError(t, err)
	assert.Empty(t, name)

	addressPart, err := n.AddressPart()
	assert.NoError(t, err)
	assert.Equal(t, AddressNull, addressPart)

	localPart, err := n.LocalPart()
	assert.NoError(t, err)
	assert.Empty(t, localPart)

	assert.Empty(t, n.DomainPart())
	assert.Equal(t, "<none>", n.StringOr("<none>"))
	assert.Equal(t, Address("default@example.com"), n.GetOr("default@example.com"))
	assert.Equal(t, NullableAddressList(""), n.AsList(), "NULL address must become the NULL list")
	assert.Panics(t, func() { n.Get() }, "Get on NULL must panic instead of returning an invalid Address")
}

// TestNullableAddress_NotNull protects that a non-NULL value behaves
// exactly like the corresponding Address.
func TestNullableAddress_NotNull(t *testing.T) {
	const n NullableAddress = `"Doe, John" <John.Doe@Example.com>`

	assert.False(t, n.IsNull())
	assert.True(t, n.IsNotNull())
	assert.NoError(t, n.Validate())
	assert.False(t, n.ValidAndNormalized(), "upper case address part is not normalized")

	parsed, err := n.Parse()
	require.NoError(t, err)
	assert.Equal(t, &mail.Address{Name: "Doe, John", Address: "john.doe@example.com"}, parsed)

	normalized, err := n.Normalized()
	require.NoError(t, err)
	assert.Equal(t, NullableAddress(`"Doe, John" <john.doe@example.com>`), normalized)
	assert.True(t, normalized.ValidAndNormalized())

	name, err := n.NamePart()
	require.NoError(t, err)
	assert.Equal(t, "Doe, John", name)

	addressPart, err := n.AddressPart()
	require.NoError(t, err)
	assert.Equal(t, NullableAddress("john.doe@example.com"), addressPart)

	localPart, err := n.LocalPart()
	require.NoError(t, err)
	assert.Equal(t, "john.doe", localPart)

	assert.Equal(t, "Example.com", n.DomainPart())
	assert.Equal(t, string(n), n.StringOr("<none>"))
	assert.Equal(t, Address(n), n.Get())
	assert.Equal(t, Address(n), n.GetOr("default@example.com"))
	assert.Equal(t, NullableAddressList(n), n.AsList())
}

func TestNullableAddress_Invalid(t *testing.T) {
	// A non-NULL invalid value must not be mistaken for NULL.
	const n NullableAddress = "not an address"

	assert.Error(t, n.Validate())
	assert.False(t, n.ValidAndNormalized())

	_, err := n.Parse()
	assert.Error(t, err, "Parse")
	normalized, err := n.Normalized()
	assert.Error(t, err, "Normalized")
	assert.Equal(t, n, normalized, "Normalized must return the input unchanged on error")
	_, err = n.NamePart()
	assert.Error(t, err, "NamePart")
	_, err = n.AddressPart()
	assert.Error(t, err, "AddressPart")
	_, err = n.LocalPart()
	assert.Error(t, err, "LocalPart")

	_, err = NormalizedNullableAddress(string(n))
	assert.Error(t, err, "NormalizedNullableAddress")
}

func TestNullableAddressFrom(t *testing.T) {
	// A missing parsed address is NULL, not some placeholder string.
	assert.Equal(t, AddressNull, NullableAddressFrom(nil))

	// With a name part the result must be the same as for the non-nullable type.
	addr := &mail.Address{Name: "Doe, John", Address: "john@example.com"}
	assert.Equal(t, AddressFrom(addr).Nullable(), NullableAddressFrom(addr))
}

func TestNullableAddress_SetAndSetNull(t *testing.T) {
	var n NullableAddress
	n.Set("a@example.com")
	assert.Equal(t, NullableAddress("a@example.com"), n)

	n.SetNull()
	assert.True(t, n.IsNull())

	n.Set("a@example.com")
	n.Set("") // documented: setting the empty Address means NULL
	assert.True(t, n.IsNull())
}

// TestNullableAddress_Scan protects the SQL NULL mapping: NULL scans as the
// NULL value, while an empty string can only come from a non-NULL column
// holding no address, which is a data error that must be reported.
func TestNullableAddress_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    NullableAddress
		wantErr bool
	}{
		{name: "NULL", value: nil, want: AddressNull},
		{name: "string", value: `"Doe, John" <john@example.com>`, want: `"Doe, John" <john@example.com>`},
		{name: "empty string", value: "", wantErr: true},
		{name: "unsupported type", value: 123, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NullableAddress("previous@example.com")
			err := n.Scan(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, n)
		})
	}
}

// TestNullableAddress_ValueScanRoundTrip protects that NULL is written to SQL
// as nil (not as an empty string, which Scan would reject) and that every
// written value scans back unchanged.
func TestNullableAddress_ValueScanRoundTrip(t *testing.T) {
	tests := []struct {
		n    NullableAddress
		want driver.Value
	}{
		{n: AddressNull, want: nil},
		{n: "a@example.com", want: "a@example.com"},
		{n: `"Doe, John" <john@example.com>`, want: `"Doe, John" <john@example.com>`},
	}
	for _, tt := range tests {
		t.Run(string(tt.n), func(t *testing.T) {
			value, err := tt.n.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, value)

			var scanned NullableAddress
			require.NoError(t, scanned.Scan(value))
			assert.Equal(t, tt.n, scanned)
		})
	}
}

// TestNullableAddress_JSON protects the JSON NULL mapping: NULL is written
// as JSON null and both null and "" are read as NULL.
func TestNullableAddress_JSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		data, err := json.Marshal(AddressNull)
		require.NoError(t, err)
		assert.Equal(t, `null`, string(data))

		data, err = json.Marshal(NullableAddress(`"Doe, John" <john@example.com>`))
		require.NoError(t, err)
		// A non-NULL value is a plain JSON string holding the exact address.
		var s string
		require.NoError(t, json.Unmarshal(data, &s))
		assert.Equal(t, `"Doe, John" <john@example.com>`, s)
	})

	t.Run("unmarshal", func(t *testing.T) {
		tests := []struct {
			json    string
			want    NullableAddress
			wantErr bool
		}{
			{json: `null`, want: AddressNull},
			{json: `""`, want: AddressNull},
			{json: `"a@example.com"`, want: "a@example.com"},
			{json: `123`, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.json, func(t *testing.T) {
				n := NullableAddress("previous@example.com")
				err := json.Unmarshal([]byte(tt.json), &n)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.want, n)
			})
		}
	})

	t.Run("struct round-trip", func(t *testing.T) {
		type wrapper struct {
			A NullableAddress `json:"a"`
			B NullableAddress `json:"b"`
		}
		in := wrapper{A: `"Doe, John" <john@example.com>`, B: AddressNull}
		data, err := json.Marshal(in)
		require.NoError(t, err)
		var out wrapper
		require.NoError(t, json.Unmarshal(data, &out))
		assert.Equal(t, in, out)
	})
}

func TestNullableAddress_UnmarshalText(t *testing.T) {
	// Text unmarshalling (used for example by URL query and form decoders)
	// maps empty text to NULL and keeps everything else verbatim.
	n := NullableAddress("previous@example.com")
	require.NoError(t, n.UnmarshalText([]byte("")))
	assert.True(t, n.IsNull())

	require.NoError(t, n.UnmarshalText([]byte("John <john@example.com>")))
	assert.Equal(t, NullableAddress("John <john@example.com>"), n)
}

func TestNullableAddress_JSONSchema(t *testing.T) {
	// The schema must allow both an email string and null.
	schema := AddressNull.JSONSchema()
	require.Len(t, schema.OneOf, 2)
	assert.Equal(t, "string", schema.OneOf[0].Type)
	assert.Equal(t, "email", schema.OneOf[0].Format)
	assert.Equal(t, "null", schema.OneOf[1].Type)
}
