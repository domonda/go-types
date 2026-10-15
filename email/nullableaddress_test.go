package email

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/mail"
	"reflect"
	"testing"

	"github.com/domonda/go-types/internal/normtest"
	"github.com/domonda/go-types/strfmt"
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

// TestNullableAddress_WhitespaceIsNull protects that a whitespace only
// value is null like for country.NullableCode and vat.NullableID, else
// a value that carries no address would be handled as an invalid address
// and written to SQL and JSON as whitespace.
func TestNullableAddress_WhitespaceIsNull(t *testing.T) {
	for _, n := range []NullableAddress{" ", "  ", "\t\n", "\u00a0"} {
		t.Run(fmt.Sprintf("%q", string(n)), func(t *testing.T) {
			assert.True(t, n.IsNull())
			assert.False(t, n.IsNotNull())
			assert.NoError(t, n.Validate())
			assert.False(t, n.ValidAndNormalized(), "whitespace is null, but not the normalized AddressNull")

			normalized, err := n.Normalized()
			require.NoError(t, err)
			assert.Equal(t, AddressNull, normalized)
			normalized, err = NormalizedNullableAddress(string(n))
			require.NoError(t, err)
			assert.Equal(t, AddressNull, normalized)
			parsed, err := n.Parse()
			require.NoError(t, err)
			assert.Nil(t, parsed)

			assert.Equal(t, "<none>", n.StringOr("<none>"))
			assert.Equal(t, Address("default@example.com"), n.GetOr("default@example.com"))
			assert.Panics(t, func() { n.Get() })
			assert.Equal(t, NullableAddressList(""), n.AsList(), "a null address must become the NULL list")

			v, err := n.Value()
			require.NoError(t, err)
			assert.Nil(t, v, "must be written as SQL NULL instead of the whitespace")
			j, err := json.Marshal(n)
			require.NoError(t, err)
			assert.Equal(t, `null`, string(j))
		})
	}
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
	assert.Equal(t, string(n), n.StringOr("<none>"), "StringOr returns the address as stored")
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

	// Without a name part the result must also match the non-nullable type:
	// the bare address without angle brackets, which is the normalized form.
	// Otherwise the same address converted via both types would not
	// compare equal and the nullable value would not be normalized.
	addr = &mail.Address{Address: "john@example.com"}
	assert.Equal(t, NullableAddress("john@example.com"), NullableAddressFrom(addr))
	assert.Equal(t, AddressFrom(addr).Nullable(), NullableAddressFrom(addr))
	assert.True(t, NullableAddressFrom(addr).ValidAndNormalized())
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
		// An empty or whitespace only string is null, see IsNull
		{name: "empty string", value: "", want: AddressNull},
		{name: "whitespace only string", value: " \t ", want: AddressNull},
		// Drivers may return text columns as []byte, which must scan like
		// the same string, as for AddressList and NullableAddressList.
		{name: "[]byte", value: []byte("a@example.com"), want: "a@example.com"},
		{name: "empty []byte", value: []byte{}, want: AddressNull},
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
// as nil (not as an empty string) and that every
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

// A non-null NullableAddress is written normalized like Address, see
// TestAddress_ValueAndMarshalJSONNormalize, and null stays SQL NULL and
// JSON null, including the whitespace-only null. Value refuses an invalid
// address, JSON writes it unchanged.
func TestNullableAddress_ValueAndMarshalJSONNormalize(t *testing.T) {
	tests := []struct {
		n        NullableAddress
		wantSQL  driver.Value
		wantJSON string
	}{
		{n: AddressNull, wantSQL: nil, wantJSON: `null`},
		{n: "  ", wantSQL: nil, wantJSON: `null`},
		{n: `Zoë Bäcker <Zoe@Example.com>`, wantSQL: `"Zoë Bäcker" <zoe@example.com>`, wantJSON: `"\"Zoë Bäcker\" \u003czoe@example.com\u003e"`},
	}
	for _, tt := range tests {
		v, err := tt.n.Value()
		require.NoError(t, err)
		assert.Equal(t, tt.wantSQL, v, "NullableAddress(%q).Value()", string(tt.n))

		j, err := json.Marshal(tt.n)
		require.NoError(t, err)
		assert.Equal(t, tt.wantJSON, string(j), "json.Marshal(NullableAddress(%q))", string(tt.n))
	}

	_, err := NullableAddress(`not an address`).Value()
	assert.Error(t, err, "Value() of an invalid address")
	j, err := json.Marshal(NullableAddress(`not an address`))
	require.NoError(t, err)
	assert.Equal(t, `"not an address"`, string(j))
}

// MarshalText and ASCII work like those of Address, see
// TestAddress_MarshalTextAndASCII, and return "" for null. StringOr
// returns the alternative only for null and any other value as stored,
// so an invalid address is shown as what it is.
func TestNullableAddress_MarshalTextASCIIStringOr(t *testing.T) {
	tests := []struct {
		n            NullableAddress
		wantText     string
		wantASCII    string
		wantStringOr string
	}{
		{n: AddressNull, wantText: ``, wantASCII: ``, wantStringOr: `none`},
		{n: `  `, wantText: ``, wantASCII: ``, wantStringOr: `none`},
		{n: `Zoë <Zoe@Example.com>`, wantText: `"Zoë" <zoe@example.com>`, wantASCII: `=?utf-8?q?Zo=C3=AB?= <zoe@example.com>`, wantStringOr: `Zoë <Zoe@Example.com>`},
		{n: `not an address`, wantText: `not an address`, wantASCII: `not an address`, wantStringOr: `not an address`},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.wantText, marshalText(t, tt.n), "NullableAddress(%q).MarshalText()", string(tt.n))
		assert.Equal(t, tt.wantASCII, tt.n.ASCII(), "NullableAddress(%q).ASCII()", string(tt.n))
		assert.Equal(t, tt.wantStringOr, tt.n.StringOr("none"), "NullableAddress(%q).StringOr()", string(tt.n))
	}
}

func TestNullableAddress_JSONSchema(t *testing.T) {
	// The schema must allow both an email string and null.
	schema := AddressNull.JSONSchema()
	require.Len(t, schema.OneOf, 2)
	assert.Equal(t, "string", schema.OneOf[0].Type)
	assert.Equal(t, "email", schema.OneOf[0].Format)
	assert.Equal(t, "null", schema.OneOf[1].Type)
}

// TestNullableAddress_ReadingAndWriting checks the reading and writing
// contract shared by the normalizing types, with a blank NullableAddress
// as null.
func TestNullableAddress_ReadingAndWriting(t *testing.T) {
	normtest.Run(t, normtest.Spec[NullableAddress]{
		Canonical: `"John Doe" <john@example.com>`,
		Variants:  []NullableAddress{`John Doe <John@Example.com>`, ` "John Doe" <JOHN@example.com> `},
		Invalid:   []NullableAddress{"not an address", "john@"},
		Nullable:  true,
	})
}

// strfmt.Scan stores a NullableAddress normalized like an Address,
// see TestAddress_StrfmtScan, and a null word or an empty cell as null.
func TestNullableAddress_StrfmtScan(t *testing.T) {
	config := strfmt.NewScanConfig()
	var addr NullableAddress
	require.NoError(t, strfmt.Scan(reflect.ValueOf(&addr).Elem(), "John Doe <John@Example.com>", config))
	assert.Equal(t, NullableAddress(`"John Doe" <john@example.com>`), addr)

	for _, null := range []string{"", "NULL", "null", "nil"} {
		addr := NullableAddress("a@example.com")
		require.NoError(t, strfmt.Scan(reflect.ValueOf(&addr).Elem(), null, config))
		assert.True(t, addr.IsNull(), "strfmt.Scan of %q", null)
	}
}
