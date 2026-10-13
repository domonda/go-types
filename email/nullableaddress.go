package email

import (
	"database/sql/driver"
	"fmt"
	"net/mail"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// NullableAddress is a string containing a non-normalized email-address
// with an optional name part before the mandatory address part.
// An empty or whitespace only string represents the SQL/JSON null value,
// like for country.NullableCode.
type NullableAddress string

// Compile-time check that NullableAddress implements nullable.NullSetable[Address]
var _ nullable.NullSetable[Address] = (*NullableAddress)(nil)

// AddressNull is the NULL value for NullableAddress.
const AddressNull NullableAddress = ""

// NullableAddressFrom creates a NullableAddress from a parsed mail.Address.
// A nil address results in the NULL value.
// Like AddressFrom, an address without a name part is returned
// as just the address part without angle brackets.
func NullableAddressFrom(a *mail.Address) NullableAddress {
	return AddressFrom(a).Nullable()
}

// NormalizedNullableAddress parses an email address less strict
// than the standard net/mail.ParseAddress function
// fixing malformed addresses and lower cases the address part.
// An empty or whitespace only string is interpreted as valid NULL address.
func NormalizedNullableAddress(addr string) (normalized NullableAddress, err error) {
	return NullableAddress(addr).Normalized()
}

// Parse converts the NullableAddress to a *mail.Address using lenient
// validation, or returns nil if the address is null.
func (n NullableAddress) Parse() (*mail.Address, error) {
	if n.IsNull() {
		return nil, nil
	}
	return Address(n).Parse()
}

// Validate returns an error if the NullableAddress is not null
// and not a valid email address.
func (n NullableAddress) Validate() error {
	if n.IsNull() {
		return nil
	}
	return Address(n).Validate()
}

// ValidAndNormalized returns true if the NullableAddress is valid and already normalized.
func (n NullableAddress) ValidAndNormalized() bool {
	norm, err := n.Normalized()
	return err == nil && n == norm
}

// Normalized parses an email address less strict
// than the standard net/mail.ParseAddress function
// fixing malformed addresses and lower cases the address part.
// An empty or whitespace only string is interpreted as valid NULL address
// and returned as AddressNull, see IsNull.
func (n NullableAddress) Normalized() (NullableAddress, error) {
	if n.IsNull() {
		return AddressNull, nil
	}
	norm, err := Address(n).Normalized()
	return NullableAddress(norm), err
}

// NamePart extracts the display name part from the email address,
// or returns an empty string if the address is null or has no name part.
func (n NullableAddress) NamePart() (string, error) {
	if n.IsNull() {
		return "", nil
	}
	return Address(n).NamePart()
}

// AddressPart returns the normalized lower case address part
// of an email address that may also contain a name part.
func (n NullableAddress) AddressPart() (NullableAddress, error) {
	if n.IsNull() {
		return "", nil
	}
	a, err := Address(n).AddressPart()
	return NullableAddress(a), err
}

// LocalPart returns the sub-part of the address part before the @ character
// or an empty string in case of a null address.
func (n NullableAddress) LocalPart() (string, error) {
	if n.IsNull() {
		return "", nil
	}
	return Address(n).LocalPart()
}

// DomainPart returns the part of the address after the @ character
// or an empty string in case of a null address.
func (n NullableAddress) DomainPart() string {
	if n.IsNull() {
		return ""
	}
	return Address(n).DomainPart()
}

// IsNull returns true if the string n is empty or whitespace only
// like for Normalized.
// IsNull implements the nullable.Nullable interface.
func (n NullableAddress) IsNull() bool {
	return strutil.TrimSpace(n) == AddressNull
}

// IsNotNull returns true if the NullableAddress is not null,
// see IsNull.
func (n NullableAddress) IsNotNull() bool {
	return !n.IsNull()
}

// String returns the normalized address like Address.String,
// "" for null, or the unchanged address if it can't be normalized.
// String implements the fmt.Stringer interface.
func (n NullableAddress) String() string {
	norm, _ := n.Normalized()
	return string(norm)
}

// ASCII returns the address like Address.ASCII, or "" for null.
func (n NullableAddress) ASCII() string {
	if n.IsNull() {
		return ""
	}
	return Address(n).ASCII()
}

// StringOr returns the normalized NullableAddress as string,
// or the passed nullString if the NullableAddress is null
// or can't be normalized.
func (n NullableAddress) StringOr(nullString string) string {
	norm, err := n.Normalized()
	if err != nil || norm.IsNull() {
		return nullString
	}
	return string(norm)
}

// Get returns the non nullable Address
// or panics if the NullableAddress is null.
// Note: check with IsNull before using Get!
func (n NullableAddress) Get() Address {
	if n.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", n))
	}
	return Address(n)
}

// GetOr returns the non nullable Address value
// or the passed defaultAddress if the NullableAddress is null.
func (n NullableAddress) GetOr(defaultAddress Address) Address {
	if n.IsNull() {
		return defaultAddress
	}
	return Address(n)
}

// Set the passed Address as NullableAddress.
// Passing an empty string will be interpreted as setting NULL.
func (n *NullableAddress) Set(s Address) {
	*n = NullableAddress(s)
}

// SetNull sets the string to its null value
func (n *NullableAddress) SetNull() {
	*n = ""
}

// Scan implements the database/sql.Scanner interface.
// An empty or whitespace only string scans as null, see IsNull.
func (n *NullableAddress) Scan(value any) error {
	switch s := value.(type) {
	case nil:
		n.SetNull()
		return nil

	case string:
		if NullableAddress(s).IsNull() {
			n.SetNull()
			return nil
		}
		*n = NullableAddress(s)
		return nil

	case []byte:
		return n.Scan(string(s))

	default:
		return fmt.Errorf("can't scan %T as email.NullableAddress", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
// Returns nil for SQL NULL if the NullableAddress is null, see IsNull,
// else the normalized address like Address.Value.
func (n NullableAddress) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return Address(n).Value()
}

// UnmarshalText implements encoding.TextUnmarshaler
func (n *NullableAddress) UnmarshalText(text []byte) error {
	*n = NullableAddress(text)
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
// setting the NullableAddress to null for the JSON null value
// or an empty JSON string.
func (n *NullableAddress) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONStringOrNull(n, data)
}

// MarshalJSON implements the json.Marshaler interface
// by returning the JSON null value if the NullableAddress is null, see IsNull,
// else the normalized address like Address.MarshalJSON.
func (n NullableAddress) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return Address(n).MarshalJSON()
}

// JSONSchema returns the JSON schema definition for the NullableAddress type.
func (NullableAddress) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Email Address",
		OneOf: []*jsonschema.Schema{
			{
				Type:   "string",
				Format: "email",
			},
			{Type: "null"},
		},
		Default: AddressNull,
	}
}

// AsList converts the NullableAddress to a NullableAddressList
// containing this single address,
// or the null NullableAddressList if n is null, see IsNull.
func (n NullableAddress) AsList() NullableAddressList {
	if n.IsNull() {
		return ""
	}
	return NullableAddressList(n)
}
