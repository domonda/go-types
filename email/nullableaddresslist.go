package email

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"

	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// NullableAddressList is a comma separated list
// of email addresses.
// The empty string default value represents
// an empty list and maps to SQL NULL and JSON null.
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

// Split converts the NullableAddressList to a slice of Address values,
// or returns nil if the list is null.
//
// Like AddressList.Split it returns all successfully parsed addresses
// even when other addresses in the list can't be parsed, with the
// parsing errors collected and returned joined.
func (n NullableAddressList) Split() ([]Address, error) {
	if n.IsNull() {
		return nil, nil
	}
	return AddressList(n).Split()
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
// returning the null value unchanged.
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

// String returns the normalized list like AddressList.String,
// "" for null, or the unchanged list if it can't be normalized.
// String implements the fmt.Stringer interface.
func (n NullableAddressList) String() string {
	norm, _ := n.Normalized()
	return string(norm)
}

// ASCII returns the list like AddressList.ASCII, or "" for null.
func (n NullableAddressList) ASCII() string {
	if n.IsNull() {
		return ""
	}
	return AddressList(n).ASCII()
}

// StringOr returns the normalized NullableAddressList as string,
// or the passed nullString if the NullableAddressList is null
// or can't be normalized.
func (n NullableAddressList) StringOr(nullString string) string {
	norm, err := n.Normalized()
	if err != nil || norm.IsNull() {
		return nullString
	}
	return string(norm)
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

// Scan implements the database/sql.Scanner interface.
// It accepts a plain string (an already joined address list) or a
// PostgreSQL array text literal ({"a@x.com","b@y.com"}, see
// https://www.postgresql.org/docs/current/arrays.html) whose elements
// are joined into the list. The array form requires PostgreSQL or an
// array-compatible database such as CockroachDB or YugabyteDB; Value
// always writes a plain string and works with any database.
// Whitespace around the array elements or a plain string is trimmed
// like by AddressSet.Scan. An empty or whitespace only string scans
// as null like the empty array {}.
func (n *NullableAddressList) Scan(value any) error {
	switch s := value.(type) {
	case nil:
		n.SetNull()
		return nil

	case string:
		// Trimmed like the array elements by scanAddressArray
		list := strutil.TrimSpace(s)
		if list == "" {
			n.SetNull()
			return nil
		}
		// Detected on the trimmed list, so a padded array is one too
		if list[0] == '{' && list[len(list)-1] == '}' {
			// Elements that are empty after the helper's trim are
			// skipped by the join.
			stringArray, err := scanAddressArray(list, "NullableAddressList")
			if err != nil {
				return err
			}
			*n = NullableAddressListJoinStrings(stringArray...)
			return nil
		}
		*n = NullableAddressList(list)
		return nil

	case []byte:
		return n.Scan(string(s))

	default:
		return fmt.Errorf("can't scan %T as email.NullableAddressList", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
// Returns nil for SQL NULL if the normalized list is null,
// which includes a whitespace only list or one without addresses
// like "undisclosed-recipients:;", else the normalized list,
// or the unchanged string if it can't be normalized.
func (n NullableAddressList) Value() (driver.Value, error) {
	norm, _ := n.Normalized()
	if norm.IsNull() {
		return nil, nil
	}
	return string(norm), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (n *NullableAddressList) UnmarshalText(text []byte) error {
	*n = NullableAddressList(text)
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
// setting the NullableAddressList to null for the JSON null value
// or an empty JSON string.
func (n *NullableAddressList) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONStringOrNull(n, data)
}

// MarshalJSON implements the json.Marshaler interface
// by returning the JSON null value if the normalized list is null like for Value,
// else the normalized list, or the unchanged string if it can't be normalized.
func (n NullableAddressList) MarshalJSON() ([]byte, error) {
	norm, _ := n.Normalized()
	if norm.IsNull() {
		return []byte("null"), nil
	}
	return json.Marshal(string(norm))
}

// MarshalText implements encoding.TextMarshaler by returning String,
// the normalized value, "" for null, or the unchanged string if it can't be
// normalized, so that strfmt and encoding/xml write the canonical form
// like Value and MarshalJSON. encoding/json writes map keys of string
// types as they are and doesn't call MarshalText for them.
func (n NullableAddressList) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}
