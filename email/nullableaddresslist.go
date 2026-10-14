package email

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/mail"
	"strings"

	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// NullableAddressList is a comma separated list
// of email addresses.
// The empty or whitespace only string represents
// an empty list and maps to SQL NULL and JSON null.
//
// NullableAddressList reads and writes like AddressList, except that null
// is SQL NULL and JSON null. See "Reading and writing values" in the
// README of the module.
type NullableAddressList string

var _ nullable.NullSetable[AddressList] = (*NullableAddressList)(nil)

// NullableAddressListJoin creates a NullableAddressList by joining
// multiple Address values with comma separators. Empty addresses are skipped.
func NullableAddressListJoin(addrs ...Address) NullableAddressList {
	return NullableAddressList(AddressListJoin(addrs...))
}

// NullableAddressListJoinStrings creates a NullableAddressList by joining
// multiple string values with comma separators. Empty strings are skipped.
func NullableAddressListJoinStrings(addrs ...string) NullableAddressList {
	return NullableAddressList(AddressListJoinStrings(addrs...))
}

// Append adds additional addresses to the NullableAddressList.
// Empty addresses are skipped.
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

// Parse converts the NullableAddressList to a slice of *mail.Address
// using lenient validation, or returns nil if the list is null.
func (n NullableAddressList) Parse() ([]*mail.Address, error) {
	if n.IsNull() {
		return nil, nil
	}
	return AddressList(n).Parse()
}

// SplitNormalized parses the NullableAddressList and returns its addresses
// normalized like AddressList.SplitNormalized, or nil if the list is null.
//
// Like AddressList.SplitNormalized it returns all successfully parsed
// addresses even when other addresses in the list can't be parsed, with
// the parsing errors collected and returned joined.
func (n NullableAddressList) SplitNormalized() ([]Address, error) {
	if n.IsNull() {
		return nil, nil
	}
	return AddressList(n).SplitNormalized()
}

// SplitRaw returns the entries of the NullableAddressList as stored,
// only trimmed, see SplitAddressList.
func (n NullableAddressList) SplitRaw() []Address {
	return SplitAddressList(string(n))
}

// Validate returns an error if the NullableAddressList is not null
// and not a valid list of email addresses.
func (n NullableAddressList) Validate() error {
	if n.IsNull() {
		return nil
	}
	return AddressList(n).Validate()
}

// ValidAndNormalized returns true if the NullableAddressList is valid and already normalized.
func (n NullableAddressList) ValidAndNormalized() bool {
	norm, err := n.Normalized()
	return err == nil && n == norm
}

// Normalized parses and normalizes the NullableAddressList,
// returning "" for null.
func (n NullableAddressList) Normalized() (NullableAddressList, error) {
	if n.IsNull() {
		return "", nil
	}
	norm, err := AddressList(n).Normalized()
	return NullableAddressList(norm), err
}

// IsNull returns true if the string n is empty or consists only of
// the runes strutil.TrimSpace removes, like for NullableAddress.
// IsNull implements the nullable.Nullable interface.
func (n NullableAddressList) IsNull() bool {
	return strutil.TrimSpace(n) == ""
}

// IsNotNull returns true if the NullableAddressList is not null,
// see IsNull.
func (n NullableAddressList) IsNotNull() bool {
	return !n.IsNull()
}

// ASCII returns the list like AddressList.ASCII, or "" for null.
func (n NullableAddressList) ASCII() string {
	if n.IsNull() {
		return ""
	}
	return AddressList(n).ASCII()
}

// StringOr returns the NullableAddressList as stored,
// or the passed nullString if the NullableAddressList is null.
func (n NullableAddressList) StringOr(nullString string) string {
	if n.IsNull() {
		return nullString
	}
	return string(n)
}

// Get returns the non nullable string value
// or panics if the NullableAddressList is null.
// Note: check with IsNull before using Get!
func (n NullableAddressList) Get() AddressList {
	if n.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", n))
	}
	return AddressList(n)
}

// GetOr returns the non nullable AddressList value
// or the passed defaultValue if the NullableAddressList is null.
func (n NullableAddressList) GetOr(defaultValue AddressList) AddressList {
	if n.IsNull() {
		return defaultValue
	}
	return AddressList(n)
}

// Set the passed AddressList as NullableAddressList.
// Passing an empty string will be interpreted as setting NULL.
func (n *NullableAddressList) Set(s AddressList) {
	*n = NullableAddressList(s)
}

// SetNull sets the string to its null value
func (n *NullableAddressList) SetNull() {
	*n = ""
}

// Scan implements the database/sql.Scanner interface like AddressList.Scan,
// with SQL NULL, a blank string, "null" and an empty array as null.
// A PostgreSQL array requires PostgreSQL or an array-compatible database
// such as CockroachDB or YugabyteDB; Value always writes a plain string
// and works with any database.
func (n *NullableAddressList) Scan(value any) error {
	var list AddressList
	if err := list.Scan(value); err != nil {
		return err
	}
	*n = NullableAddressList(list)
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns SQL NULL if the NullableAddressList is null, else the list as
// stored with the surrounding whitespace trimmed, or the error of Validate
// for an invalid list.
func (n NullableAddressList) Value() (driver.Value, error) {
	return AddressList(n).Value()
}

// UnmarshalText implements encoding.TextUnmarshaler
// by storing the text unchanged.
func (n *NullableAddressList) UnmarshalText(text []byte) error {
	*n = NullableAddressList(text)
	return nil
}

// UnmarshalJSON implements encoding/json.Unmarshaler like
// AddressList.UnmarshalJSON, but setting the NullableAddressList to null
// for the JSON null value, so that decoding null into a field that holds
// a value resets it.
func (n *NullableAddressList) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(`null`)) {
		n.SetNull()
		return nil
	}
	var list AddressList
	if err := list.UnmarshalJSON(data); err != nil {
		return err
	}
	*n = NullableAddressList(list)
	return nil
}

// MarshalJSON implements encoding/json.Marshaler by writing the JSON null
// value if the NullableAddressList is null, else the list as stored
// as JSON string.
func (n NullableAddressList) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(string(n))
}

// MarshalText implements encoding.TextMarshaler by returning the list
// as stored, or "" for null, so that encoding/xml and strfmt write it
// unchanged.
func (n NullableAddressList) MarshalText() ([]byte, error) {
	if n.IsNull() {
		return nil, nil
	}
	return []byte(n), nil
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (n *NullableAddressList) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(n, name)
}
