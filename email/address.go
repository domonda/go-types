// Package email provides comprehensive email address handling, parsing, validation,
// and message processing utilities for Go applications.
//
// The package includes:
// - Email address parsing with lenient validation
// - Address normalization and validation
// - Email message parsing (MIME, TNEF)
// - Address lists and sets management
// - Database integration (Scanner/Valuer interfaces)
// - JSON schema generation
package email

import (
	"net/mail"
	"strings"

	"github.com/invopop/jsonschema"
)

// Address represents a non-normalized email address string that may contain
// an optional display name part before the mandatory email address part.
// Examples: "user@example.com", "John Doe <john@example.com>"
type Address string

// AddressFrom creates an Address from a parsed mail.Address.
// If the address has no name part, returns just the address.
// If the address has a name part, returns the full formatted string.
func AddressFrom(addr *mail.Address) Address {
	if addr == nil {
		return ""
	}
	if addr.Name == "" {
		// Without name just return the address part.
		// parsed.String() always puts the address part
		// within angle brackets which is only needed
		// if there is also a name part.
		return Address(addr.Address)
	}
	return Address(addr.String())
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

// Validate checks if the Address is a valid email address format.
// Returns an error if the address cannot be parsed.
func (a Address) Validate() error {
	_, err := a.Parse()
	return err
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
// (the part after the @ symbol) from an email address.
// This strips any display name and returns just the email address.
func (a Address) AddressPartString() (string, error) {
	parsed, err := a.Parse()
	if err != nil {
		return "", err
	}
	return parsed.Address, nil
}

// AddressPart returns the normalized lowercase address part
// (the part after the @ symbol) as an Address type.
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
	s = s[strings.LastIndexByte(s, '@')+1:]
	return strings.TrimRight(s, "> \t\r\n")
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
