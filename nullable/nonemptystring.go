package nullable

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// NonEmptyString is a string type where the empty string value
// is interpreted as SQL NULL and JSON null by
// implementing the sql.Scanner and driver.Valuer interfaces
// and also json.Marshaler and json.Unmarshaler.
// Note that this type can't hold an empty string without
// interpreting it as not null SQL or JSON value.
type NonEmptyString string

// NonEmptyStringFromPtr converts a string pointer to a NonEmptyString
// interpreting nil as null value "".
func NonEmptyStringFromPtr(ptr *string) NonEmptyString {
	if ptr == nil {
		return ""
	}
	return NonEmptyString(*ptr)
}

// NonEmptyStringFromError converts an error to a NonEmptyString
// interpreting a nil error as null value ""
// or else using err.Error() as value.
func NonEmptyStringFromError(err error) NonEmptyString {
	if err == nil {
		return ""
	}
	return NonEmptyString(err.Error())
}

// Ptr returns the address of the string value or nil if n.IsNull()
func (n NonEmptyString) Ptr() *string {
	if n.IsNull() {
		return nil
	}
	return (*string)(&n)
}

// IsNull returns true if the string n is empty.
func (n NonEmptyString) IsNull() bool {
	return n == ""
}

// StringOr returns the string value of n or the passed nullString if n.IsNull()
func (n NonEmptyString) StringOr(nullString string) string {
	if n.IsNull() {
		return nullString
	}
	return string(n)
}

// Scan implements the database/sql.Scanner interface.
func (n *NonEmptyString) Scan(value interface{}) error {
	switch s := value.(type) {
	case nil:
		*n = ""
		return nil

	case string:
		*n = NonEmptyString(s)
		return nil

	default:
		return fmt.Errorf("can't scan %T as nullable.NonEmptyString", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n NonEmptyString) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return string(n), nil
}

// UnarshalJSON implements encoding/json.Unmarshaler.
// Interprets []byte(nil), []byte(""), []byte("null") as null.
func (n *NonEmptyString) UnmarshalJSON(sourceJSON []byte) error {
	if len(sourceJSON) == 0 || bytes.Equal(sourceJSON, []byte("null")) {
		*n = ""
		return nil
	}
	return json.Unmarshal(sourceJSON, n)
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value if n is an empty string.
func (n NonEmptyString) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte("null"), nil
	}
	return json.Marshal(string(n))
}
