package email

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/domonda/go-types/nullable"
)

// NormalizeAddressList parses a comma-separated email address list using lenient validation
// that fixes common malformations and normalizes address parts to lowercase.
// Duplicate addresses (based on normalized address parts) are automatically removed.
// Returns an error if the list does not contain at least one valid address.
func NormalizeAddressList(list string) (normalized []string, err error) {
	addrs, err := ParseAddressList(list)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, nil
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

// AddressList represents a comma-separated list of email addresses.
// The list must contain at least one valid email address.
// Use NullableAddressList for lists that can be empty.
type AddressList string

// AddressListJoin creates an AddressList by joining multiple Address values
// with comma separators. Empty addresses are skipped.
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

// AddressListJoinStrings creates an AddressList by joining multiple string values
// with comma separators. Empty strings are skipped.
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

// Append adds additional addresses to the AddressList.
// Empty addresses are skipped.
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

// Parse converts the AddressList to a slice of *mail.Address using lenient validation.
func (l AddressList) Parse() ([]*mail.Address, error) {
	return ParseAddressList(string(l))
}

// Split converts the AddressList to a slice of Address values.
// Returns nil if the list is empty or invalid.
func (l AddressList) Split() ([]Address, error) {
	parsed, err := l.Parse()
	if err != nil {
		return nil, err
	}
	if len(parsed) == 0 {
		return nil, nil
	}
	a := make([]Address, len(parsed))
	for i, p := range parsed {
		a[i] = AddressFrom(p)
	}
	return a, nil
}

// UniqueAddressParts returns an AddressSet containing the unique normalized
// address parts from the AddressList. This removes duplicates based on
// the actual email addresses (ignoring display names).
func (l AddressList) UniqueAddressParts() (AddressSet, error) {
	parsed, err := l.Parse()
	if err != nil {
		return nil, err
	}
	if len(parsed) == 0 {
		return nil, nil
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

// ValidAndNormalized returns true if the AddressList is valid and already normalized.
func (l AddressList) ValidAndNormalized() bool {
	norm, err := l.Normalized()
	return err == nil && l == norm
}

func (l AddressList) Normalized() (AddressList, error) {
	parsed, err := l.Parse()
	if err != nil {
		return l, err
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
