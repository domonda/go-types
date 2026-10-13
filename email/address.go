package email

import (
	"database/sql/driver"
	"encoding/json"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types"
)

// Compile-time check that Address implements types.NormalizableValidator[Address]
var _ types.NormalizableValidator[Address] = Address("")

// Address represents a non-normalized email address string that may contain
// an optional display name part before the mandatory email address part.
// Examples: "user@example.com", "John Doe <john@example.com>"
type Address string

// AddressFrom creates an Address from a parsed mail.Address.
// If the address has no name part, returns just the address.
// If the address has a name part, returns the name as quoted string
// followed by the address within angle brackets like mail.Address.String,
// but a non-ASCII name is kept as UTF-8 instead of being RFC 2047 encoded,
// so that the Address stays readable in JSON and databases.
// Use mail.Address.String to format an address for a message header.
func AddressFrom(addr *mail.Address) Address {
	if addr == nil {
		return ""
	}
	if addr.Name == "" {
		// Without name just return the address part.
		// mail.Address.String always puts the address part
		// within angle brackets which is only needed
		// if there is also a name part.
		return Address(addr.Address)
	}
	// Without name String formats only the address part,
	// quoting the local part where needed
	addrPart := (&mail.Address{Address: addr.Address}).String()
	return Address(quoteName(addr.Name) + " " + addrPart)
}

// quoteName returns name as quoted string like mail.Address.String
// does for an ASCII name, also for a non-ASCII name, which is allowed
// as UTF-8 by RFC 6532. Like mail.Address.String, it escapes quotes and
// backslashes and drops ASCII control characters other than tab.
func quoteName(name string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range name {
		switch {
		case r == '"' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == ' ' || r == '\t' || ('!' <= r && r <= '~') || r >= utf8.RuneSelf:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// asciiAddress formats addr like AddressFrom, but RFC 2047 encodes
// a non-ASCII name like mail.Address.String does.
func asciiAddress(addr *mail.Address) string {
	if addr.Name == "" {
		return addr.Address
	}
	return addr.String()
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

// String returns the normalized address,
// or the unchanged address if it can't be normalized.
// A non-ASCII name stays UTF-8, see ASCII.
// String implements the fmt.Stringer interface.
func (a Address) String() string {
	norm, _ := a.Normalized()
	return string(norm)
}

// ASCII returns the normalized address like String, but with
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

// Value implements the driver database/sql/driver.Valuer interface.
// Returns the normalized address, or the unchanged string
// if it can't be normalized, so an empty Address is written as empty string.
func (a Address) Value() (driver.Value, error) {
	norm, _ := a.Normalized()
	return string(norm), nil
}

// MarshalJSON implements encoding/json.Marshaler.
// Returns the normalized address as JSON string,
// or the unchanged string if it can't be normalized.
func (a Address) MarshalJSON() ([]byte, error) {
	norm, _ := a.Normalized()
	return json.Marshal(string(norm))
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
func (a Address) AddressPartString() (string, error) {
	parsed, err := a.Parse()
	if err != nil {
		return "", err
	}
	return parsed.Address, nil
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
func (a Address) LocalPart() (string, error) {
	parsed, err := a.Parse()
	if err != nil {
		return "", err
	}
	return parsed.Address[:strings.IndexByte(parsed.Address, '@')], nil
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
