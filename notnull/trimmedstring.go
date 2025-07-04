package notnull

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"unsafe"

	"github.com/domonda/go-types/strutil"
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
// TrimmedString can hold an empty string.
type TrimmedString string

// TrimmedStringf formats a string using fmt.Sprintf
// and returns it as TrimmedString.
func TrimmedStringf(format string, a ...any) TrimmedString {
	return TrimmedString(strutil.TrimSpace(fmt.Sprintf(format, a...)))
}

// TrimmedStringFrom trims the passed str and returns it as TrimmedString
func TrimmedStringFrom(str string) TrimmedString {
	return TrimmedString(strutil.TrimSpace(str))
}

// JoinTrimmedStrings joins trimmed strings with the passed separator between them
func TrimmedStringJoin[S ~string](separator string, strs ...S) TrimmedString {
	var b strings.Builder
	for i, s := range strs {
		if i > 0 {
			b.WriteString(separator)
		}
		b.WriteString(strutil.TrimSpace(string(s)))
	}
	return TrimmedString(b.String())
}

// IsEmpty indicates if the trimmed string is empty
// which is also the case when the underlying
// string consists only of whitespace.
func (s TrimmedString) IsEmpty() bool {
	for _, r := range s {
		if !strutil.IsSpace(r) {
			return false
		}
	}
	return true
}

// String implements the fmt.Stringer interface
// by returning a trimmed string that might be empty
// if the underlying string consisting only of whitespace.
func (s TrimmedString) String() string {
	return strutil.TrimSpace(string(s))
}

// ToValidUTF8 returns a copy of the TrimmedString with each run of invalid UTF-8 byte sequences
// replaced by the replacement string, which may be empty.
func (s TrimmedString) ToValidUTF8(replacement string) TrimmedString {
	return TrimmedStringFrom(strings.ToValidUTF8(s.String(), replacement))
}

// ToUpper returns s with all Unicode letters mapped to their upper case.
func (s TrimmedString) ToUpper() TrimmedString {
	return TrimmedString(strings.ToUpper(s.String()))
}

// ToLower returns s with all Unicode letters mapped to their lower case.
func (s TrimmedString) ToLower() TrimmedString {
	return TrimmedString(strings.ToLower(s.String()))
}

// Contains reports whether substr is within s.
func (s TrimmedString) Contains(substr string) bool {
	return strings.Contains(s.String(), substr)
}

// ContainsAny reports whether any Unicode code points in chars are within s.
func (s TrimmedString) ContainsAny(chars string) bool {
	return strings.ContainsAny(s.String(), chars)
}

// ContainsRune reports whether the Unicode code point r is within s.
func (s TrimmedString) ContainsRune(r rune) bool {
	return strings.ContainsRune(s.String(), r)
}

// HasPrefix tests whether the TrimmedString begins with prefix.
func (s TrimmedString) HasPrefix(prefix string) bool {
	return strings.HasPrefix(s.String(), prefix)
}

// HasSuffix tests whether the TrimmedString ends with suffix.
func (s TrimmedString) HasSuffix(suffix string) bool {
	return strings.HasSuffix(s.String(), suffix)
}

// TrimPrefix returns s without the provided leading prefix string.
// If the TrimmedString doesn't start with prefix, s is returned unchanged.
func (s TrimmedString) TrimPrefix(prefix string) TrimmedString {
	return TrimmedStringFrom(strings.TrimPrefix(s.String(), prefix))
}

// TrimSuffix returns s without the provided trailing suffix string.
// If the TrimmedString doesn't end with suffix, s is returned unchanged.
func (s TrimmedString) TrimSuffix(suffix string) TrimmedString {
	return TrimmedStringFrom(strings.TrimSuffix(s.String(), suffix))
}

// ReplaceAll returns a copy of the TrimmedString with all
// non-overlapping instances of old replaced by new.
// If old is empty, it matches at the beginning of the string
// and after each UTF-8 sequence, yielding up to k+1 replacements
// for a k-rune string.
func (s TrimmedString) ReplaceAll(old, new string) TrimmedString {
	return TrimmedString(strutil.TrimSpace(strings.ReplaceAll(s.String(), old, new)))
}

// Split slices s into all substrings separated by sep and returns a slice of
// the substrings between those separators.
//
// If s does not contain sep and sep is not empty, Split returns a
// slice of length 1 whose only element is s.
//
// If sep is empty, Split splits after each UTF-8 sequence. If both s
// and sep are empty, Split returns an empty slice.
//
// It is equivalent to SplitN with a count of -1.
//
// To split around the first instance of a separator, see Cut.
func (s TrimmedString) Split(sep string) []TrimmedString {
	substrings := strings.Split(s.String(), sep)
	for i, substring := range substrings {
		substrings[i] = strutil.TrimSpace(substring)
	}
	return *(*[]TrimmedString)(unsafe.Pointer(&substrings)) //#nosec G103 -- unsafe OK
}

// Set the passed string as TrimmedString
func (s *TrimmedString) Set(str string) {
	*s = TrimmedString(strutil.TrimSpace(str))
}

// Value implements the driver database/sql/driver.Valuer interface
func (s TrimmedString) Value() (driver.Value, error) {
	return s.String(), nil
}

// Scan implements the database/sql.Scanner interface
func (s *TrimmedString) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*s = TrimmedString(strutil.TrimSpace(x))
		return nil

	case []byte:
		*s = TrimmedString(strutil.TrimSpaceBytes(x))
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
	*s = TrimmedString(strutil.TrimSpaceBytes(text))
	return nil
}

// MarshalJSON implements encoding/json.Marshaler
func (s TrimmedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements encoding/json.Unmarshaler
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
