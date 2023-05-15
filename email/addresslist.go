package email

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/domonda/go-types/nullable"
)

// NormalizeAddressList parses an email address list less strict
// than the standard net/mail.ParseAddressList function
// fixing malformed addresses and lower cases the address part.
// Duplicates with the same normalized address part
// will be removed from the result.
// NormalizeAddressList returns an error if list does not contain
// at least one address.
func NormalizeAddressList(list string) (normalized []string, err error) {
	addrs, err := ParseAddressList(list)
	if err != nil {
		return nil, err
	}
	appended := make(map[string]bool, len(addrs))
	normalized = make([]string, 0, len(addrs))
	for _, a := range addrs {
		if appended[a.Address] {
			continue
		}
		normalized = append(normalized, string(AddressFrom(a)))
		appended[a.Address] = true
	}
	return normalized, nil
}

// AddressList is a comma separated list
// of at least one email address.
//
// Use NullableAddressList for a list
// that can contain zero addresses.
type AddressList string

func AddressListJoin(addrs ...Address) AddressList {
	var b strings.Builder
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(string(addr))
	}
	return AddressList(b.String())
}

func AddressListJoinStrings(addrs ...string) AddressList {
	var b strings.Builder
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(addr)
	}
	return AddressList(b.String())
}

func (l AddressList) Append(addrs ...Address) AddressList {
	var b strings.Builder
	b.WriteString(string(l))
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(string(addr))
	}
	return AddressList(b.String())
}

func (l AddressList) Parse() ([]*mail.Address, error) {
	return ParseAddressList(string(l))
}

func (l AddressList) Split() ([]Address, error) {
	parsed, err := l.Parse()
	if err != nil {
		return nil, err
	}
	a := make([]Address, len(parsed))
	for i, p := range parsed {
		a[i] = AddressFrom(p)
	}
	return a, nil
}

func (l AddressList) UniqueAddressParts() (AddressSet, error) {
	parsed, err := l.Parse()
	if err != nil {
		return nil, err
	}
	set := make(AddressSet, len(parsed))
	for _, addr := range parsed {
		set.Add(Address(addr.Address))
	}
	return set, nil
}

func (l AddressList) Validate() error {
	_, err := l.Parse()
	return err
}

func (l AddressList) Normalized() (AddressList, error) {
	parsed, err := l.Parse()
	if err != nil {
		return "", err
	}
	b := strings.Builder{}
	b.Grow(len(l))
	for i, p := range parsed {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(string(AddressFrom(p)))
	}
	return AddressList(b.String()), nil
}

func (l AddressList) Nullable() NullableAddressList {
	return NullableAddressList(l)
}

// Scan implements the database/sql.Scanner interface.
// Supports scanning SQL strings and string arrays.
func (l *AddressList) Scan(value any) error {
	switch s := value.(type) {
	case string:
		if s == "" {
			return errors.New("can't scan empty string as email.AddressList")
		}
		if s[0] == '{' && s[len(s)-1] == '}' {
			stringArray, err := nullable.SplitArray(s)
			if err != nil {
				return fmt.Errorf("can't scan SQL array string %q as email.AddressList because of: %w", s, err)
			}
			*l = AddressListJoinStrings(stringArray...)
			return nil
		}
		*l = AddressList(s)
		return nil

	case []byte:
		return l.Scan(string(s))

	default:
		return fmt.Errorf("can't scan %T as email.AddressList", value)
	}
}
