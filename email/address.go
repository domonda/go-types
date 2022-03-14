package email

import (
	"net/mail"
	"strings"
)

// Address is a string containing a non-normalized email-address
// with an optional name part before the mandatory address part.
type Address string

func AddressFrom(a *mail.Address) Address {
	if a == nil {
		return ""
	}
	return Address(a.String())
}

// NormalizedAddress parses an email address less strict
// than the standard net/mail.ParseAddress function
// fixing malformed addresses and lower cases the address part.
func NormalizedAddress(addr string) (normalized Address, err error) {
	return Address(addr).Normalized()
}

// Normalized parses an email address less strict
// than the standard net/mail.ParseAddress function
// fixing malformed addresses and lower cases the address part.
func (a Address) Normalized() (Address, error) {
	parsed, err := a.Parse()
	if err != nil {
		return "", err
	}
	if parsed.Name == "" {
		// Without name just return the address part.
		// parsed.String() always puts the address part
		// within angle brackets which is only needed
		// if there is also a name part.
		return Address(parsed.Address), nil
	}
	return Address(parsed.String()), nil
}

// Parse the Address as *mail.Address less strict than
// the standard net/mail.ParseAddress function
// fixing malformed addresses and lower cases the address part.
func (a Address) Parse() (*mail.Address, error) {
	return ParseAddress(string(a))
}

func (a Address) Validate() error {
	_, err := a.Parse()
	return err
}

func (a Address) Nullable() NullableAddress {
	return NullableAddress(a)
}

func (a Address) NamePart() (string, error) {
	parsed, err := a.Parse()
	if err != nil {
		return "", err
	}
	return parsed.Name, nil
}

// AddressPartString returns the normalized lower case address part
// from an email address with an optional name part.
func (a Address) AddressPartString() (string, error) {
	parsed, err := a.Parse()
	if err != nil {
		return "", err
	}
	return parsed.Address, nil
}

// AddressPart returns the normalized lower case address part
// from an email address with an optional name part.
func (a Address) AddressPart() (Address, error) {
	addr, err := a.AddressPartString()
	return Address(addr), err
}

// LocalPart returns the sub-part of the address part before the @ character
func (a Address) LocalPart() (string, error) {
	parsed, err := a.Parse()
	if err != nil {
		return "", err
	}
	return parsed.Address[:strings.IndexByte(parsed.Address, '@')], nil
}

// DomainPart returns the part of the address after the @ character
// or an empty string in case it can't be parsed.
func (a Address) DomainPart() string {
	s := string(a)
	s = s[strings.LastIndexByte(s, '@')+1:]
	return strings.TrimRight(s, "> \t\r\n")
}

func (a Address) IsFromKnownEmailProvider() bool {
	_, is := emailProviderDomains[a.DomainPart()]
	return is
}

func (a Address) AsList() AddressList {
	return AddressList(a)
}
