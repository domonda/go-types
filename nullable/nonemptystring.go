package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Ensure that NonEmptyString implements StringGetter
var _ StringGetter = NonEmptyString("")

// NullNonEmptyString is the SQL NULL and JSON null value for NonEmptyString.
const NullNonEmptyString NonEmptyString = ""

// NonEmptyString is a string type where the empty string value
// is interpreted as SQL NULL and JSON null by
// implementing the sql.Scanner and driver.Valuer interfaces
// and also json.Marshaler and json.Unmarshaler.
// Note that this type can't hold an empty string without
// interpreting it as not null SQL or JSON value.
type NonEmptyString string

// NonEmptyStringf formats a string using fmt.Sprintf
// and returns it as NonEmptyString.
// An empty string will be interpreted as null value.
func NonEmptyStringf(format string, a ...interface{}) NonEmptyString {
	return NonEmptyString(fmt.Sprintf(format, a...))
}

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

// NonEmptyStringTrimSpace returns a NonEmptyString
// by trimming space from the passed string.
// If the passed string with trimmed space is an empty string
// then the NonEmptyString will represent null.
func NonEmptyStringTrimSpace(str string) NonEmptyString {
	return NonEmptyString(str).TrimSpace()
}

// JoinNonEmptyStrings joins only those strings that are
// not empty/null with the passed separator between them.
func JoinNonEmptyStrings(separator string, strs ...NonEmptyString) NonEmptyString {
	var b strings.Builder
	for _, s := range strs {
		if s.IsNull() {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(separator)
		}
		b.WriteString(string(s))
	}
	return NonEmptyString(b.String())
}

// Ptr returns the address of the string value or nil if n.IsNull()
func (n NonEmptyString) Ptr() *string {
	if n.IsNull() {
		return nil
	}
	return (*string)(&n)
}

// IsNull returns true if the string n is empty.
// IsNull implements the Nullable interface.
func (n NonEmptyString) IsNull() bool {
	return n == ""
}

// IsNotNull returns true if the string n is not empty.
func (n NonEmptyString) IsNotNull() bool {
	return n != ""
}

// TrimSpace returns the string with all white-space
// characters trimmed from beginning and end.
// A potentially resulting empty string will be interpreted as null.
func (n NonEmptyString) TrimSpace() NonEmptyString {
	return NonEmptyString(strings.TrimSpace(string(n)))
}

// StringOr returns the string value of n or the passed nullString if n.IsNull()
func (n NonEmptyString) StringOr(nullString string) string {
	if n.IsNull() {
		return nullString
	}
	return string(n)
}

// Get returns the non nullable string value
// or panics if the NonEmptyString is null.
// Note: check with IsNull before using Get!
func (n NonEmptyString) Get() string {
	if n.IsNull() {
		panic("NULL nullable.NonEmptyString")
	}
	return string(n)
}

// Set the passed string as NonEmptyString.
// Passing an empty string will be interpreted as setting NULL.
func (n *NonEmptyString) Set(s string) {
	*n = NonEmptyString(s)
}

// SetNull sets the string to its null value
func (n *NonEmptyString) SetNull() {
	*n = ""
}

// Scan implements the database/sql.Scanner interface.
func (n *NonEmptyString) Scan(value interface{}) error {
	switch s := value.(type) {
	case nil:
		n.SetNull()
		return nil

	case string:
		if s == "" {
			return errors.New("can't scan empty string as nullable.NonEmptyString")
		}
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

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (n *NonEmptyString) UnmarshalText(text []byte) error {
	*n = NonEmptyString(text)
	return nil
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null for an empty/null string.
func (n NonEmptyString) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(string(n))
}
