package notnull

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

// TrimmedString is a string type where
// all marshaller and unmarshaller will trim
// whitespace first before returning or using a value.
type TrimmedString string

// TrimmedStringf formats a string using fmt.Sprintf
// and returns it as TrimmedString.
func TrimmedStringf(format string, a ...any) TrimmedString {
	return TrimmedString(strings.TrimSpace(fmt.Sprintf(format, a...)))
}

// TrimmedStringFrom trims the passed str and returns it as TrimmedString
func TrimmedStringFrom(str string) TrimmedString {
	return TrimmedString(strings.TrimSpace(str))
}

// JoinTrimmedStrings joins trimmed strings with the passed separator between them
func JoinTrimmedStrings(separator string, strs ...TrimmedString) TrimmedString {
	var b strings.Builder
	for _, s := range strs {
		if b.Len() > 0 {
			b.WriteString(separator)
		}
		b.WriteString(strings.TrimSpace(string(s)))
	}
	return TrimmedString(b.String())
}

// String implements the fmt.Stringer interface
// by returning a trimmed string that might be empty
// if the underlying string consisting only of whitespace.
func (s TrimmedString) String() string {
	return strings.TrimSpace(string(s))
}

// Set the passed string as TrimmedString
func (s *TrimmedString) Set(str string) {
	*s = TrimmedString(strings.TrimSpace(str))
}

// Value implements the driver database/sql/driver.Valuer interface
func (s TrimmedString) Value() (driver.Value, error) {
	return s.String(), nil
}

// Scan implements the database/sql.Scanner interface
func (s *TrimmedString) Scan(value any) error {
	switch x := value.(type) {
	case string:
		x = strings.TrimSpace(x)
		if x == "" {
			return errors.New("can't scan empty trimmed string as notnull.TrimmedString")
		}
		*s = TrimmedString(x)
		return nil

	default:
		return fmt.Errorf("can't scan %T as notnull.TrimmedString", value)
	}
}

// UnmarshalText implements the encoding.TextMarshaler interface
func (s TrimmedString) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (s *TrimmedString) UnmarshalText(text []byte) error {
	*s = TrimmedString(bytes.TrimSpace(text))
	return nil
}

// MarshalJSON implements encoding/json.Marshaler
func (s TrimmedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// MarshalJSON implements encoding/json.Unmarshaler
func (s *TrimmedString) UnmarshalJSON(j []byte) error {
	var str string
	err := json.Unmarshal(j, &str)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON (%s) as notnull.TrimmedString because: %w", j, err)
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
