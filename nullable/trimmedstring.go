package nullable

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"unsafe"

	"github.com/invopop/jsonschema"

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

// TrimmedStringNull is the NULL value "" for TrimmedString
const TrimmedStringNull TrimmedString = ""

// TrimmedString is a string type where the empty trimmed string value
// is interpreted as SQL NULL and JSON null by
// implementing the sql.Scanner and driver.Valuer interfaces
// and also json.Marshaler and json.Unmarshaler.
//
// Note that this type can't hold a not null empty string,
// because it will interpret it as null SQL or JSON value.
type TrimmedString string

// TrimmedStringf formats a string using fmt.Sprintf
// and returns it as TrimmedString.
// An empty trimmed string will be interpreted as null value.
func TrimmedStringf(format string, a ...any) TrimmedString {
	return TrimmedString(strutil.TrimSpace(fmt.Sprintf(format, a...)))
}

// TrimmedStringFrom trims the passed str and returns it as TrimmedString
// An empty trimmed string will be interpreted as null value.
func TrimmedStringFrom(str string) TrimmedString {
	return TrimmedString(strutil.TrimSpace(str))
}

// TrimmedStringFromPtr converts a string pointer to a TrimmedString
// interpreting nil as null value "".
// An empty trimmed string will be interpreted as null value.
func TrimmedStringFromPtr(ptr *string) TrimmedString {
	if ptr == nil {
		return ""
	}
	return TrimmedString(strutil.TrimSpace(*ptr))
}

// TrimmedStringFromError converts an error to a TrimmedString
// interpreting a nil error as null value ""
// or else using err.Error() as value.
func TrimmedStringFromError(err error) TrimmedString {
	if err == nil {
		return ""
	}
	return TrimmedString(strutil.TrimSpace(err.Error()))
}

// TrimmedStringJoin joins only those strings that are
// not empty after trimming with the passed separator between them.
func TrimmedStringJoin[S ~string](separator string, strs ...S) TrimmedString {
	var b strings.Builder
	for _, s := range strs {
		s := strutil.TrimSpace(string(s))
		if s == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(separator)
		}
		b.WriteString(s)
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
//
// IsNull implements the Nullable interface.
func (s TrimmedString) IsNull() bool {
	for _, r := range s {
		if !strutil.IsSpace(r) {
			return false
		}
	}
	return true
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

// Get returns the non nullable string value
// or panics if the TrimmedString is null.
// Note: check with IsNull before using Get!
func (s TrimmedString) Get() string {
	if s.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", s))
	}
	return s.String()
}

// Set the passed string as TrimmedString.
// Passing an empty trimmed string will be interpreted as setting NULL.
func (s *TrimmedString) Set(str string) {
	*s = TrimmedString(strutil.TrimSpace(str))
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
		*s = TrimmedStringFrom(x)
		return nil

	case []byte:
		*s = TrimmedStringFrom(string(x))
		return nil

	default:
		return fmt.Errorf("can not scan %T as nullable.TrimmedString", value)
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
	*s = TrimmedString(strutil.TrimSpaceBytes(text))
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

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (s *TrimmedString) UnmarshalJSON(j []byte) error {
	if bytes.Equal(j, []byte(`null`)) {
		s.SetNull()
		return nil
	}
	var str string
	err := json.Unmarshal(j, &str)
	if err != nil {
		return fmt.Errorf("can not unmarshal JSON (%s) as nullable.TrimmedString because: %w", j, err)
	}
	s.Set(str)
	return nil
}

func (TrimmedString) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable Trimmed String",
		OneOf: []*jsonschema.Schema{
			{Type: "string"},
			{Type: "null"},
		},
		Default: TrimmedStringNull,
	}
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
