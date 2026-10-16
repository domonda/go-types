package email

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"
	"net/mail"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// NullableAddress is a string containing a non-normalized email-address
// with an optional name part before the mandatory address part.
// An empty or whitespace only string represents the SQL/JSON null value,
// like for country.NullableCode.
//
// NullableAddress reads and writes like Address, except that null is
// SQL NULL and JSON null. See "Reading and writing values" in the README
// of the module.
type NullableAddress string

// Compile-time check that NullableAddress implements nullable.NullSetable[Address]
var _ nullable.NullSetable[Address] = (*NullableAddress)(nil)

// Compile-time check that NullableAddress implements types.NormalizableValidator[NullableAddress]
var _ types.NormalizableValidator[NullableAddress] = NullableAddress("")

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

// Valid returns true if the NullableAddress is null or a valid address.
func (n NullableAddress) Valid() bool {
	return n.Validate() == nil
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

// IsNull returns true if the string n is empty or consists only of
// the runes strutil.TrimSpace removes: whitespace, control and other
// non-printable runes, and invalid UTF-8, like for Normalized.
// IsNull implements the nullable.Nullable interface.
func (n NullableAddress) IsNull() bool {
	return strutil.TrimSpace(n) == AddressNull
}

// IsNotNull returns true if the NullableAddress is not null,
// see IsNull.
func (n NullableAddress) IsNotNull() bool {
	return !n.IsNull()
}

// ASCII returns the address like Address.ASCII, or "" for null.
func (n NullableAddress) ASCII() string {
	if n.IsNull() {
		return ""
	}
	return Address(n).ASCII()
}

// StringOr returns the NullableAddress as stored,
// or the passed nullString if the NullableAddress is null.
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
// Surrounding whitespace is trimmed. SQL NULL and a blank string are
// AddressNull. An invalid address is kept after that trim, without an error.
func (n *NullableAddress) Scan(value any) error {
	return internal.ScanNormalized(n, value, internal.Trim[NullableAddress], nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the trimmed address, or SQL NULL if the NullableAddress is
// null. An invalid address is written after that trim, not rejected.
func (n NullableAddress) Value() (driver.Value, error) {
	return Address(n).Value()
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null sets the
// NullableAddress to null, so that decoding null into a field that holds
// a value resets it, and any other JSON value is an error.
func (n *NullableAddress) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(n, data, true, internal.Trim[NullableAddress], nil)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (n *NullableAddress) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(n, d, start, internal.Trim[NullableAddress], nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (n *NullableAddress) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(n, attr, internal.Trim[NullableAddress], nil)
}

// ScanString implements the strfmt.Scannable interface.
// The strings "NULL", "null" and "nil" set the NullableAddress to null.
// Without validate it stores source like Scan, with validate it stores
// the normalized address or returns the error of Normalized.
func (n *NullableAddress) ScanString(source string, validate bool) error {
	if !validate {
		return internal.ScanStringNormalized(n, source, false, true, internal.Trim[NullableAddress], nil)
	}
	return internal.ScanStringNormalized(n, source, true, true, NullableAddress.Normalized, nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing JSON null
// if the NullableAddress is null, else the trimmed address.
func (n NullableAddress) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte("null"), nil
	}
	return Address(n).MarshalJSON()
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// but with "" for null, so that encoding/xml and strfmt write the
// trimmed address. encoding/json writes map keys of string types as
// they are and doesn't call MarshalText for them.
func (n NullableAddress) MarshalText() ([]byte, error) {
	if n.IsNull() {
		return []byte{}, nil
	}
	return Address(n).MarshalText()
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (n *NullableAddress) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(n, name)
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
