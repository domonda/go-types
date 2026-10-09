package email

import (
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emailAddressNormalized = map[Address]string{
	"test@test.com":                  "test@test.com",
	"te+st@test.com":                 "te+st@test.com",
	"test888@test.com":               "test888@test.com",
	"<noname@test.com>":              "noname@test.com",
	"With Name <With.Name@test.com>": `"With Name" <with.name@test.com>`,
}

func TestAddressNormalized(t *testing.T) {
	for addr, expected := range emailAddressNormalized {
		t.Run(string(addr), func(t *testing.T) {
			result, err := addr.Normalized()
			assert.NoError(t, err)
			assert.Equal(t, expected, string(result))
		})
	}
}

func TestAddressFrom(t *testing.T) {
	tests := []struct {
		name string
		addr *mail.Address
		want Address
	}{
		{name: `nil`, addr: nil, want: ``},
		{name: `empty`, addr: &mail.Address{Name: "", Address: ""}, want: ``},
		{name: `erik@domonda.com`, addr: &mail.Address{Name: "", Address: "erik@domonda.com"}, want: `erik@domonda.com`},
		{name: `"Erik Unger" <erik@domonda.com>`, addr: &mail.Address{Name: "Erik Unger", Address: "erik@domonda.com"}, want: `"Erik Unger" <erik@domonda.com>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddressFrom(tt.addr); got != tt.want {
				t.Errorf("AddressFrom() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestAddress_ValidAndValidAndNormalized protects the difference between
// "parsable" and "already in canonical form": Valid accepts the lenient
// formats seen in the wild, while ValidAndNormalized is only true for the
// exact string Normalized would produce, so callers can use it to detect
// values that still need to be rewritten before storing or comparing them.
func TestAddress_ValidAndValidAndNormalized(t *testing.T) {
	tests := []struct {
		addr               Address
		valid              bool
		validAndNormalized bool
	}{
		{addr: `user@example.com`, valid: true, validAndNormalized: true},
		{addr: `"John Doe" <john@example.com>`, valid: true, validAndNormalized: true},
		// Upper case address parts are valid but not normalized because
		// normalized addresses are compared as lower case strings.
		{addr: `User@Example.com`, valid: true, validAndNormalized: false},
		// An unquoted name is valid but the normalized form quotes it.
		{addr: `John Doe <john@example.com>`, valid: true, validAndNormalized: false},
		// Angle brackets without a name are dropped by normalization.
		{addr: `<john@example.com>`, valid: true, validAndNormalized: false},
		{addr: ``, valid: false},
		{addr: `not an address`, valid: false},
		{addr: `@example.com`, valid: false},
		{addr: `user@`, valid: false},
		{addr: `user@example`, valid: false}, // a top level domain is required
	}
	for _, tt := range tests {
		t.Run(string(tt.addr), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.addr.Valid(), "Valid")
			assert.Equal(t, tt.valid, tt.addr.Validate() == nil, "Validate must agree with Valid")
			assert.Equal(t, tt.validAndNormalized, tt.addr.ValidAndNormalized(), "ValidAndNormalized")
		})
	}
}

// TestAddress_NormalizedIsIdempotent protects the string round-trip:
// a normalized address must parse back to itself, otherwise values stored
// in normalized form would keep changing every time they are normalized.
func TestAddress_NormalizedIsIdempotent(t *testing.T) {
	for _, addr := range []Address{
		`John Doe <John.Doe@Example.com>`,
		`"Doe, John" <john@example.com>`,
		`<Noname@Example.com>`,
		`=?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <senor@hola.com>`,
	} {
		t.Run(string(addr), func(t *testing.T) {
			once, err := addr.Normalized()
			require.NoError(t, err)
			twice, err := once.Normalized()
			require.NoError(t, err)
			assert.Equal(t, once, twice)
			assert.True(t, once.ValidAndNormalized())
		})
	}
}

func TestAddress_Normalized_InvalidReturnsInput(t *testing.T) {
	// On error the original value is returned unchanged so that callers
	// that ignore the error don't lose the data they passed in.
	const addr Address = "not an address"
	norm, err := addr.Normalized()
	assert.Error(t, err)
	assert.Equal(t, addr, norm)
}

// TestAddress_Parts protects the accessors used to compare and route
// addresses: the name part is returned verbatim (decoded and unquoted),
// while the address and local parts are lower cased so that differently
// cased spellings of the same mailbox compare equal.
func TestAddress_Parts(t *testing.T) {
	tests := []struct {
		addr        Address
		name        string
		addressPart Address
		localPart   string
	}{
		{addr: `user@example.com`, name: ``, addressPart: `user@example.com`, localPart: `user`},
		{addr: `<User@Example.com>`, name: ``, addressPart: `user@example.com`, localPart: `user`},
		{addr: `Jane <Jane+Invoices@Example.com>`, name: `Jane`, addressPart: `jane+invoices@example.com`, localPart: `jane+invoices`},
		// The comma in the quoted name must not split the address.
		{addr: `"Doe, John" <John.Doe@Example.com>`, name: `Doe, John`, addressPart: `john.doe@example.com`, localPart: `john.doe`},
		// RFC 2047 encoded names are decoded.
		{addr: `=?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <senor@hola.com>`, name: `¡Hola, señor!`, addressPart: `senor@hola.com`, localPart: `senor`},
	}
	for _, tt := range tests {
		t.Run(string(tt.addr), func(t *testing.T) {
			name, err := tt.addr.NamePart()
			require.NoError(t, err)
			assert.Equal(t, tt.name, name, "NamePart")

			addressPart, err := tt.addr.AddressPart()
			require.NoError(t, err)
			assert.Equal(t, tt.addressPart, addressPart, "AddressPart")

			addressPartString, err := tt.addr.AddressPartString()
			require.NoError(t, err)
			assert.Equal(t, string(tt.addressPart), addressPartString, "AddressPartString")

			localPart, err := tt.addr.LocalPart()
			require.NoError(t, err)
			assert.Equal(t, tt.localPart, localPart, "LocalPart")
		})
	}
}

func TestAddress_Parts_InvalidAddress(t *testing.T) {
	// Invalid addresses must report an error instead of
	// returning a partial or made up part.
	const addr Address = "not an address"

	name, err := addr.NamePart()
	assert.Error(t, err, "NamePart")
	assert.Empty(t, name)

	addressPart, err := addr.AddressPart()
	assert.Error(t, err, "AddressPart")
	assert.Empty(t, addressPart)

	localPart, err := addr.LocalPart()
	assert.Error(t, err, "LocalPart")
	assert.Empty(t, localPart)
}

// TestAddress_DomainPart protects that the domain is taken from the
// actual address and not from an address-like display name,
// because the domain is used to identify the sender's organization.
func TestAddress_DomainPart(t *testing.T) {
	tests := []struct {
		addr Address
		want string
	}{
		{addr: `user@example.com`, want: `example.com`},
		{addr: `John <john@example.com>`, want: `example.com`},
		{addr: `John <john@example.com >`, want: `example.com`},
		{addr: "john@example.com\r\n", want: `example.com`},
		// The last @ wins, so an address in the name part is ignored.
		{addr: `"john@old.example.org" <john@new.example.com>`, want: `new.example.com`},
	}
	for _, tt := range tests {
		t.Run(string(tt.addr), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.addr.DomainPart())
		})
	}
}

// TestAddress_Conversions protects that converting to the nullable
// and list types keeps the exact string so no data is lost or reformatted.
func TestAddress_Conversions(t *testing.T) {
	const addr Address = `"Doe, John" <john@example.com>`

	assert.Equal(t, NullableAddress(addr), addr.Nullable())

	list := addr.AsList()
	assert.Equal(t, AddressList(addr), list)
	split, err := list.Split()
	require.NoError(t, err)
	assert.Equal(t, []Address{addr}, split, "the comma in the quoted name must not create a second list entry")
}

func TestAddress_JSONSchema(t *testing.T) {
	// The email format lets API clients validate addresses before sending them.
	schema := Address("").JSONSchema()
	assert.Equal(t, "string", schema.Type)
	assert.Equal(t, "email", schema.Format)
}
