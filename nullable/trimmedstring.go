package nullable

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"encoding/xml"
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
	_ xml.Marshaler            = TrimmedString("")
	_ xml.Unmarshaler          = new(TrimmedString)
)

// NullTrimmedString is the NULL value "" for TrimmedString
const NullTrimmedString TrimmedString = ""

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

// TrimmedStringFrom trims the passed str and returns it as TrimmedString
// An empty trimmed string will be interpreted as null value.
func TrimmedStringFrom(str string) TrimmedString {
	return TrimmedString(strings.TrimSpace(str))
}

// TrimmedStringFromPtr converts a string pointer to a TrimmedString
// interpreting nil as null value "".
// An empty trimmed string will be interpreted as null value.
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
func (s TrimmedString) Ptr() *string {
	if s.IsNull() {
		return nil
	}
	return (*string)(&s)
}

// IsNull returns true if the string is empty.
// IsNull implements the Nullable interface.
func (s TrimmedString) IsNull() bool {
	return s == "" || strings.TrimSpace(string(s)) == ""
}

// IsNotNull returns true if the string is not empty.
func (s TrimmedString) IsNotNull() bool {
	return !s.IsNull()
}

// StringOr returns the trimmed string value of n
// or the passed nullString if n.IsNull()
func (s TrimmedString) StringOr(nullString string) string {
	if s.IsNull() {
		return nullString
	}
	return s.String()
}

// String implements the fmt.Stringer interface
// by returning a trimmed string that might be empty
// in case of the NULL value or an underlying string
// consisting only of whitespace.
func (s TrimmedString) String() string {
	return strings.TrimSpace(string(s))
}

// Get returns the non nullable string value
// or panics if the TrimmedString is null.
// Note: check with IsNull before using Get!
func (s TrimmedString) Get() string {
	if s.IsNull() {
		panic("NULL nullable.TrimmedString")
	}
	return s.String()
}

// Set the passed string as TrimmedString.
// Passing an empty trimmed string will be interpreted as setting NULL.
func (s *TrimmedString) Set(str string) {
	*s = TrimmedString(strings.TrimSpace(str))
}

// SetNull sets the string to its null value
func (s *TrimmedString) SetNull() {
	*s = ""
}

// Value implements the driver database/sql/driver.Valuer interface.
func (s TrimmedString) Value() (driver.Value, error) {
	if s.IsNull() {
		return nil, nil
	}
	return s.String(), nil
}

// Scan implements the database/sql.Scanner interface.
func (s *TrimmedString) Scan(value any) error {
	switch x := value.(type) {
	case nil:
		s.SetNull()
		return nil

	case string:
		x = strings.TrimSpace(x)
		if x == "" {
			return errors.New("can't scan empty trimmed string as nullable.TrimmedString")
		}
		*s = TrimmedString(x)
		return nil

	default:
		return fmt.Errorf("can't scan %T as nullable.TrimmedString", value)
	}
}

// UnmarshalText implements the encoding.TextMarshaler interface
func (s TrimmedString) MarshalText() ([]byte, error) {
	if s.IsNull() {
		return nil, nil
	}
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (s *TrimmedString) UnmarshalText(text []byte) error {
	*s = TrimmedString(bytes.TrimSpace(text))
	return nil
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value for an empty (null) string.
func (s TrimmedString) MarshalJSON() ([]byte, error) {
	if s.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(s.String())
}

// MarshalJSON implements encoding/json.Unmarshaler.
func (s *TrimmedString) UnmarshalJSON(j []byte) error {
	if bytes.Equal(j, []byte(`null`)) {
		s.SetNull()
		return nil
	}
	var str string
	err := json.Unmarshal(j, &str)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON (%s) as nullable.TrimmedString because: %w", j, err)
	}
	s.Set(str)
	return nil
}

func (s TrimmedString) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(s.String(), start)
}

func (s *TrimmedString) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string
	err := d.DecodeElement(&str, &start)
	if err != nil {
		return err
	}
	s.Set(str)
	return nil
}
