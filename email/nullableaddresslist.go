package email

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/domonda/go-types/nullable"
)

// NullableAddressList is a comma separated list
// of email addresses.
// The empty string default value represents
// an empty list and maps to SQL NULL and JSON null.
type NullableAddressList string

func NullableAddressListJoin(addrs ...Address) NullableAddressList {
	return NullableAddressList(AddressListJoin(addrs...))
}

func NullableAddressListJoinStrings(addrs ...string) NullableAddressList {
	return NullableAddressList(AddressListJoinStrings(addrs...))
}

func (n NullableAddressList) Append(addrs ...Address) NullableAddressList {
	var b strings.Builder
	b.WriteString(string(n))
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(string(addr))
	}
	return NullableAddressList(b.String())
}

func (n NullableAddressList) Parse() ([]*mail.Address, error) {
	if n.IsNull() {
		return nil, nil
	}
	return AddressList(n).Parse()
}

func (n NullableAddressList) Split() ([]Address, error) {
	if n.IsNull() {
		return nil, nil
	}
	return AddressList(n).Split()
}

func (n NullableAddressList) Validate() error {
	if n.IsNull() {
		return nil
	}
	return AddressList(n).Validate()
}

func (n NullableAddressList) Normalized() (NullableAddressList, error) {
	if n.IsNull() {
		return n, nil
	}
	norm, err := AddressList(n).Normalized()
	return NullableAddressList(norm), err
}

// IsNull returns true if the string n is empty.
func (n NullableAddressList) IsNull() bool {
	return n == ""
}

// IsNotNull returns true if the string n is not empty.
func (n NullableAddressList) IsNotNull() bool {
	return n != ""
}

// StringOr returns the string value of n or the passed nullString if n.IsNull()
func (n NullableAddressList) StringOr(nullString string) string {
	if n.IsNull() {
		return nullString
	}
	return string(n)
}

// Get returns the non nullable string value
// or panics if the NullableAddressList is null.
// Note: check with IsNull before using Get!
func (n NullableAddressList) Get() string {
	if n.IsNull() {
		panic("NULL email.NullableAddressList")
	}
	return string(n)
}

// Set the passed string as NullableAddressList.
// Passing an empty string will be interpreted as setting NULL.
func (n *NullableAddressList) Set(s string) {
	*n = NullableAddressList(s)
}

// SetNull sets the string to its null value
func (n *NullableAddressList) SetNull() {
	*n = ""
}

// Scan implements the database/sql.Scanner interface.
// Supports scanning SQL strings and string arrays.
func (n *NullableAddressList) Scan(value any) error {
	switch s := value.(type) {
	case nil:
		n.SetNull()
		return nil

	case string:
		if s == "" {
			return errors.New("can't scan empty string as email.NullableAddressList")
		}
		if s[0] == '{' && s[len(s)-1] == '}' {
			stringArray, err := nullable.SplitArray(s)
			if err != nil {
				return fmt.Errorf("can't scan SQL array string %q as email.NullableAddressList because of: %w", s, err)
			}
			*n = NullableAddressListJoinStrings(stringArray...)
			return nil
		}
		*n = NullableAddressList(s)
		return nil

	case []byte:
		return n.Scan(string(s))

	default:
		return fmt.Errorf("can't scan %T as email.NullableAddressList", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n NullableAddressList) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return string(n), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (n *NullableAddressList) UnmarshalText(text []byte) error {
	*n = NullableAddressList(text)
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
// returning the JSON null value for and empty/NULL address list.
func (n *NullableAddressList) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		n.SetNull()
		return nil
	}
	return json.Unmarshal(data, (*string)(n))
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullableAddressList) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte("null"), nil
	}
	return json.Marshal(string(n))
}
