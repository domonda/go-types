package email

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/xml"
	"mime"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
)

// Compile-time check that Address implements types.NormalizableValidator[Address]
var _ types.NormalizableValidator[Address] = Address("")

// Address represents a non-normalized email address string that may contain
// an optional display name part before the mandatory email address part.
// Examples: "user@example.com", "John Doe <john@example.com>"
//
// Scan, UnmarshalJSON and UnmarshalXML store a valid address normalized,
// a blank one as "" and an invalid one unchanged. Value writes the
// normalized address, SQL NULL for a blank one, and returns an error for
// an invalid one. MarshalJSON and MarshalText write the normalized address
// or an invalid one unchanged, and fmt prints the address as stored.
// Address has no UnmarshalText, so encoding/json reads and writes the keys
// of a map[Address]V as given. See "Reading and writing values" in the
// README of the module.
type Address string

// AddressFrom creates an Address from a parsed mail.Address.
// If the address has no name part, returns just the address.
// If the address has a name part, returns the name as quoted string
// followed by the address within angle brackets like mail.Address.String,
// but a non-ASCII name is kept as UTF-8 instead of being RFC 2047 encoded,
// so that the Address stays readable in JSON and databases,
// and quotes, backslashes and runes that are not graphic are removed
// from the name instead of being escaped, because ParseAddress reads
// no escapes in names. A local part is quoted if net/mail or
// ParseAddress can't read it unquoted, like "a b"@example.com.
// A name containing "=?" is RFC 2047 encoded, see encodedName.
// Use mail.Address.String to format an address for a message header.
func AddressFrom(addr *mail.Address) Address {
	if addr == nil {
		return ""
	}
	// The name is cleaned like by ParseAddress,
	// so that the result parses back to the same address
	name := cleanName(addr.Name)
	switch {
	case name == "":
		// Without name just return the address part.
		// mail.Address.String always puts the address part
		// within angle brackets which is only needed
		// if there is also a name part.
		return Address(addrSpec(addr))
	case strings.Contains(name, "=?"):
		return Address(encodedName(name) + " <" + addrSpec(addr) + ">")
	}
	return Address(`"` + name + `" <` + addrSpec(addr) + ">")
}

// encodedName returns name as RFC 2047 encoded-word, also if it is
// ASCII, which mime.BEncoding.Encode would return unchanged.
// It is used for names containing "=?", which ParseAddress and
// mail clients would decode as encoded-word, also within quotes,
// so that a name like "=?utf-8?q?x?=" reads back unchanged
// instead of as "x", and normalizing it again gives the same address.
func encodedName(name string) string {
	return "=?utf-8?b?" + base64.StdEncoding.EncodeToString([]byte(name)) + "?="
}

// addrSpec returns the address part of addr (local-part@domain)
// formatted like mail.Address.String, which quotes a local part
// that needs quoting, like "a b"@example.com, but without angle brackets.
// A local part that the parser would not read back unquoted, like one with
// a leading apostrophe or letters it doesn't accept unquoted, is quoted too.
func addrSpec(addr *mail.Address) string {
	at := strings.LastIndexByte(addr.Address, '@')
	if at < 0 {
		// Not an address that mail.Address.String could format,
		// it would append an @ to a hand-built value without domain
		return addr.Address
	}
	s := (&mail.Address{Address: addr.Address}).String()
	s = strings.TrimSuffix(strings.TrimPrefix(s, "<"), ">")
	if strings.HasPrefix(s, `"`) || isUnquotedLocalPart(addr.Address[:at]) {
		return s
	}
	return quoteLocalPart(addr.Address[:at]) + addr.Address[at:]
}

// asciiAddress formats addr like AddressFrom, but RFC 2047 encodes
// a non-ASCII name like mail.Address.String does.
func asciiAddress(addr *mail.Address) string {
	name := cleanName(addr.Name)
	switch {
	case name == "":
		return addrSpec(addr)
	case strings.Contains(name, "=?"):
		name = encodedName(name)
	case !isASCII(name) && strings.ContainsAny(name, "\"#$%&'(),.:;<>@[]^`{|}~"):
		// Like mail.Address.String, because an encoded-word
		// in a name must not hold these characters unencoded
		name = mime.BEncoding.Encode("utf-8", name)
	case !isASCII(name):
		name = mime.QEncoding.Encode("utf-8", name)
	default:
		name = `"` + name + `"`
	}
	return name + " <" + addrSpec(addr) + ">"
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// NormalizedAddress parses an email address using lenient validation
// that fixes common malformations and normalizes the address part to lowercase.
// This is more permissive than the standard net/mail.ParseAddress function.
func NormalizedAddress(addr string) (normalized Address, err error) {
	return Address(addr).Normalized()
}

// Normalized parses the email address using lenient validation
// that fixes common malformations and normalizes the address part to lowercase.
// This is more permissive than the standard net/mail.ParseAddress function.
func (a Address) Normalized() (Address, error) {
	parsed, err := a.Parse()
	if err != nil {
		return a, err
	}
	return AddressFrom(parsed), nil
}

// Parse converts the Address to a *mail.Address using lenient validation
// that fixes common malformations and normalizes the address part to lowercase.
// This is more permissive than the standard net/mail.ParseAddress function.
func (a Address) Parse() (*mail.Address, error) {
	return ParseAddress(string(a))
}

// ASCII returns the normalized address, but with
// a non-ASCII name RFC 2047 encoded like mail.Address.String does,
// as required for message headers, or the unchanged address
// if it can't be normalized. The address part is not encoded,
// so an address part with non-ASCII characters stays non-ASCII.
func (a Address) ASCII() string {
	parsed, err := a.Parse()
	if err != nil {
		return string(a)
	}
	return asciiAddress(parsed)
}

// Scan implements the database/sql.Scanner interface.
// A valid address is stored normalized, SQL NULL and a blank string as "",
// and an address that can't be normalized unchanged without an error,
// so that one invalid value doesn't fail a whole query; Validate reports it.
func (a *Address) Scan(value any) error {
	return internal.ScanNormalized(a, value, Address.Normalized, nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the normalized address, SQL NULL for a blank Address,
// or the error of Normalized for an address that can't be normalized.
func (a Address) Value() (driver.Value, error) {
	return internal.ValueNormalized(a, Address.Normalized, nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing the
// normalized address, "" for a blank Address, or the address unchanged
// if it can't be normalized.
func (a Address) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(a, Address.Normalized, nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null leaves the
// Address unchanged like encoding/json does for a string type,
// and any other JSON value is an error.
func (a *Address) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(a, data, false, Address.Normalized, nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// so that encoding/xml and strfmt write the normalized address.
// encoding/json writes map keys of string types as they are
// and doesn't call MarshalText for them.
func (a Address) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(a, Address.Normalized, nil)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (a *Address) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(a, d, start, Address.Normalized, nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (a *Address) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(a, attr, Address.Normalized, nil)
}

// Validate checks if the Address is a valid email address format.
// Returns an error if the address cannot be parsed.
func (a Address) Validate() error {
	_, err := a.Parse()
	return err
}

// Valid returns true if the Address is a valid email address format.
func (a Address) Valid() bool {
	return a.Validate() == nil
}

// ValidAndNormalized returns true if the Address is valid and already normalized.
func (a Address) ValidAndNormalized() bool {
	norm, err := a.Normalized()
	return err == nil && a == norm
}

// Nullable converts the Address to a NullableAddress type.
func (a Address) Nullable() NullableAddress {
	return NullableAddress(a)
}

// NamePart extracts the display name part from the email address.
// Returns an empty string if no name part is present.
func (a Address) NamePart() (string, error) {
	parsed, err := a.Parse()
	if err != nil {
		return "", err
	}
	return parsed.Name, nil
}

// AddressPartString returns the normalized lowercase address part
// (local-part@domain without the display name) from an email address.
// This strips any display name and returns just the email address.
// A local part that needs quotes is returned quoted,
// like "a b"@example.com.
func (a Address) AddressPartString() (string, error) {
	parsed, err := a.Parse()
	if err != nil {
		return "", err
	}
	return addrSpec(parsed), nil
}

// AddressPart returns the normalized lowercase address part
// (local-part@domain without the display name) as an Address type.
// This strips any display name and returns just the email address.
func (a Address) AddressPart() (Address, error) {
	addr, err := a.AddressPartString()
	return Address(addr), err
}

// LocalPart returns the part of the email address before the @ symbol.
// This is the local part of the email address (e.g., "user" from "user@example.com").
// A local part that needs quotes is returned quoted, like "a b".
func (a Address) LocalPart() (string, error) {
	addr, err := a.AddressPartString()
	if err != nil {
		return "", err
	}
	// The last @, because a quoted local part can contain one
	return addr[:strings.LastIndexByte(addr, '@')], nil
}

// DomainPart returns the domain part of the email address (after the @ symbol).
// This is a simple string extraction that doesn't require parsing.
// Returns an empty string if no @ symbol is found.
func (a Address) DomainPart() string {
	s := string(a)
	i := strings.LastIndexByte(s, '@')
	if i < 0 {
		return ""
	}
	return strings.TrimRight(s[i+1:], "> \t\r\n")
}

// AsList converts the Address to an AddressList containing this single address.
func (a Address) AsList() AddressList {
	return AddressList(a)
}

// JSONSchema returns the JSON schema definition for the Address type.
// This is used for API documentation and validation.
func (Address) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:  "Email Address",
		Type:   "string",
		Format: "email",
	}
}
