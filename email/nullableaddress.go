package email

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"

	"github.com/invopop/jsonschema"
)

// NullableAddress is a string containing a non-normalized email-address
// with an optional name part before the mandatory address part.
// An empty string represents the SQL/JSON null value.
type NullableAddress string

const AddressNull NullableAddress = ""

func NullableAddressFrom(a *mail.Address) NullableAddress {
	if a == nil {
		return ""
	}
	return NullableAddress(a.String())
}

// NormalizedNullableAddress parses an email address less strict
// than the standard net/mail.ParseAddress function
// fixing malformed addresses and lower cases the address part.
// An empty string is interpreted as valid NULL address.
func NormalizedNullableAddress(addr string) (normalized NullableAddress, err error) {
	return NullableAddress(addr).Normalized()
}

func (n NullableAddress) Parse() (*mail.Address, error) {
	if n.IsNull() {
		return nil, nil
	}
	return Address(n).Parse()
}

func (n NullableAddress) Validate() error {
	if n.IsNull() {
		return nil
	}
	return Address(n).Validate()
}

// Normalized parses an email address less strict
// than the standard net/mail.ParseAddress function
// fixing malformed addresses and lower cases the address part.
// An empty string is interpreted as valid NULL address.
func (n NullableAddress) Normalized() (NullableAddress, error) {
	if n.IsNull() {
		return n, nil
	}
	norm, err := Address(n).Normalized()
	return NullableAddress(norm), err
}

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

// IsNull returns true if the string n is empty.
func (n NullableAddress) IsNull() bool {
	return n == ""
}

// IsNotNull returns true if the string n is not empty.
func (n NullableAddress) IsNotNull() bool {
	return n != ""
}

// StringOr returns the string value of n or the passed nullString if n.IsNull()
func (n NullableAddress) StringOr(nullString string) string {
	if n.IsNull() {
		return nullString
	}
	return string(n)
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
func (n *NullableAddress) Scan(value any) error {
	switch s := value.(type) {
	case nil:
		n.SetNull()
		return nil

	case string:
		if s == "" {
			return errors.New("can't scan empty string as email.NullableAddress")
		}
		*n = NullableAddress(s)
		return nil

	default:
		return fmt.Errorf("can't scan %T as email.NullableAddress", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n NullableAddress) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return string(n), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (n *NullableAddress) UnmarshalText(text []byte) error {
	*n = NullableAddress(text)
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
// returning the JSON null value for and empty/NULL address.
func (n *NullableAddress) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(`null`)) {
		n.SetNull()
		return nil
	}
	return json.Unmarshal(data, (*string)(n))
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullableAddress) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(string(n))
}

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

func (n NullableAddress) AsList() NullableAddressList {
	return NullableAddressList(n)
}
