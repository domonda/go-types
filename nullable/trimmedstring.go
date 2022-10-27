package nullable

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	_ StringGetter             = TrimmedString("")
	_ fmt.Stringer             = TrimmedString("")
	_ driver.Valuer            = TrimmedString("")
	_ sql.Scanner              = new(TrimmedString)
	_ encoding.TextMarshaler   = TrimmedString("")
	_ encoding.TextUnmarshaler = new(TrimmedString)
	_ json.Marshaler           = TrimmedString("")
	_ json.Unmarshaler         = new(TrimmedString)
)

// TrimmedString is a string type where the empty trimmed string value
// is interpreted as SQL NULL and JSON null by
// implementing the sql.Scanner and driver.Valuer interfaces
// and also json.Marshaler and json.Unmarshaler.
// Note that this type can't hold an empty string without
// interpreting it as not null SQL or JSON value.
type TrimmedString string

// TrimmedStringf formats a string using fmt.Sprintf
// and returns it as TrimmedString.
// An empty trimmed string will be interpreted as null value.
func TrimmedStringf(format string, a ...any) TrimmedString {
	return TrimmedString(strings.TrimSpace(fmt.Sprintf(format, a...)))
}

// TrimmedStringFromPtr converts a string pointer to a TrimmedString
// interpreting nil as null value "".
func TrimmedStringFromPtr(ptr *string) TrimmedString {
	if ptr == nil {
		return ""
	}
	return TrimmedString(strings.TrimSpace(*ptr))
}

// TrimmedStringFromError converts an error to a TrimmedString
// interpreting a nil error as null value ""
// or else using err.Error() as value.
func TrimmedStringFromError(err error) TrimmedString {
	if err == nil {
		return ""
	}
	return TrimmedString(strings.TrimSpace(err.Error()))
}

// JoinTrimmedStrings joins only those strings that are
// not empty/null with the passed separator between them.
func JoinTrimmedStrings(separator string, strs ...TrimmedString) TrimmedString {
	var b strings.Builder
	for _, s := range strs {
		if s.IsNull() {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(separator)
		}
		b.WriteString(strings.TrimSpace(string(s)))
	}
	return TrimmedString(b.String())
}

// Ptr returns the address of the string value or nil if n.IsNull()
func (n TrimmedString) Ptr() *string {
	if n.IsNull() {
		return nil
	}
	return (*string)(&n)
}

// IsNull returns true if the string n is empty.
// IsNull implements the Nullable interface.
func (n TrimmedString) IsNull() bool {
	return n == "" || strings.TrimSpace(string(n)) == ""
}

// IsNotNull returns true if the string n is not empty.
func (n TrimmedString) IsNotNull() bool {
	return !n.IsNull()
}

// StringOr returns the trimmed string value of n
// or the passed nullString if n.IsNull()
func (n TrimmedString) StringOr(nullString string) string {
	if n.IsNull() {
		return nullString
	}
	return n.String()
}

// String implements the fmt.Stringer interface
// by returning a trimmed string that might be empty
// in case of the NULL value or an underlying string
// consisting only of whitespace.
func (n TrimmedString) String() string {
	return strings.TrimSpace(string(n))
}

// Get returns the non nullable string value
// or panics if the TrimmedString is null.
// Note: check with IsNull before using Get!
func (n TrimmedString) Get() string {
	if n.IsNull() {
		panic("NULL nullable.TrimmedString")
	}
	return n.String()
}

// Set the passed string as TrimmedString.
// Passing an empty trimmed string will be interpreted as setting NULL.
func (n *TrimmedString) Set(s string) {
	*n = TrimmedString(strings.TrimSpace(s))
}

// SetNull sets the string to its null value
func (n *TrimmedString) SetNull() {
	*n = ""
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n TrimmedString) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.String(), nil
}

// Scan implements the database/sql.Scanner interface.
func (n *TrimmedString) Scan(value any) error {
	switch s := value.(type) {
	case nil:
		n.SetNull()
		return nil

	case string:
		s = strings.TrimSpace(s)
		if s == "" {
			return errors.New("can't scan empty trimmed string as nullable.TrimmedString")
		}
		*n = TrimmedString(s)
		return nil

	default:
		return fmt.Errorf("can't scan %T as nullable.TrimmedString", value)
	}
}

// UnmarshalText implements the encoding.TextMarshaler interface
func (n TrimmedString) MarshalText() ([]byte, error) {
	if n.IsNull() {
		return nil, nil
	}
	return []byte(n.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (n *TrimmedString) UnmarshalText(text []byte) error {
	*n = TrimmedString(bytes.TrimSpace(text))
	return nil
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value for an empty (null) string.
func (n TrimmedString) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(n.String())
}

// MarshalJSON implements encoding/json.Unmarshaler.
func (n *TrimmedString) UnmarshalJSON(j []byte) error {
	if bytes.Equal(j, []byte(`null`)) {
		n.SetNull()
		return nil
	}
	var str string
	err := json.Unmarshal(j, &str)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON (%s) as nullable.TrimmedString because: %w", j, err)
	}
	n.Set(str)
	return nil
}
