// Package account provides account number handling with validation,
// parsing, and database integration for Go applications.
//
// The package includes:
// - Account number validation with regex patterns
// - Support for alphanumeric account numbers
// - Numeric conversion utilities
// - Database integration (Scanner/Valuer interfaces)
// - JSON and XML marshalling/unmarshalling
// - Nullable account number support
package account

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-errs"
)

// Error constants for account number validation
const (
	ErrInvalidNumber      errs.Sentinel = "invalid account number"
	ErrAlphanumericNumber errs.Sentinel = "account number is alphanumeric"
)

// NumberRegex defines the regular expression pattern for valid account numbers.
// Allows alphanumeric characters, underscores, hyphens, forward slashes, colons,
// periods, semicolons, and commas. Must start with alphanumeric character.
const NumberRegex = `^[0-9A-Za-z][0-9A-Za-z_\-\/:.;,]*$`

var numberRegexp = regexp.MustCompile(NumberRegex)

// Compile time check if types implement interfaces
var (
	_ fmt.Stringer     = Number("")
	_ driver.Valuer    = Number("")
	_ sql.Scanner      = new(Number)
	_ json.Marshaler   = Number("")
	_ json.Unmarshaler = new(Number)
	// _ xml.Marshaler    = Number("")
	_ xml.Unmarshaler = new(Number)
)

// Number represents an account number that can contain alphanumeric characters
// and special characters like underscores, hyphens, forward slashes, colons,
// periods, semicolons, and commas. Must start with an alphanumeric character.
type Number string

// NumberFrom creates a Number from a string with validation.
// Trims leading and trailing whitespace and validates the format.
// Returns an error if the string is not a valid account number.
func NumberFrom(str string) (Number, error) {
	str = strings.TrimSpace(str)
	if err := Number(str).Validate(); err != nil {
		return "", err
	}
	return Number(str), nil
}

// NumberFromUint creates a Number from a uint64 value.
func NumberFromUint(u uint64) Number {
	return Number(strconv.FormatUint(u, 10))
}

// Valid returns true if the Number matches the regular expression pattern.
// The pattern allows alphanumeric characters and special characters like
// underscores, hyphens, forward slashes, colons, periods, semicolons, and commas.
func (n Number) Valid() bool {
	return numberRegexp.MatchString(string(n))
}

// Validate returns an error if the Number does not match the regular expression pattern.
// Returns ErrInvalidNumber wrapped with the invalid number if validation fails.
func (n Number) Validate() error {
	if !n.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidNumber, n)
	}
	return nil
}

// HasPrefix checks if the Number starts with the specified prefix.
func (n Number) HasPrefix(prefix string) bool {
	return strings.HasPrefix(string(n), prefix)
}

// HasSuffix checks if the Number ends with the specified suffix.
func (n Number) HasSuffix(suffix string) bool {
	return strings.HasSuffix(string(n), suffix)
}

// IsNumeric returns true if the Number contains only digits (0-9).
// Returns false for empty strings.
func (n Number) IsNumeric() bool {
	if n == "" {
		return false
	}
	for _, r := range n {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// ValidateNumeric returns an error if the Number is not purely numeric.
// Returns ErrAlphanumericNumber wrapped with the invalid number if validation fails.
func (n Number) ValidateNumeric() error {
	if !n.IsNumeric() {
		return fmt.Errorf("%w: %q", ErrAlphanumericNumber, n)
	}
	return nil
}

// Uint converts the Number to a uint64.
// Returns an error if the Number is not purely numeric.
// Prevents parsing of non-digit strings like hex numbers.
func (n Number) Uint() (uint64, error) {
	// Check upfront to prevent non digit strings like hex numbers
	// to be parsed by strconv.ParseUint
	if err := n.ValidateNumeric(); err != nil {
		return 0, err
	}
	return strconv.ParseUint(string(n), 10, 64)
}

// UintPtr converts the Number to a *uint64.
// Returns nil if conversion fails.
func (n Number) UintPtr() (*uint64, error) {
	u, err := n.Uint()
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Int converts the Number to an int64.
// Returns an error if the Number is not purely numeric.
// Prevents parsing of non-digit strings like hex numbers.
func (n Number) Int() (int64, error) {
	// Check upfront to prevent non digit strings like hex numbers
	// to be parsed by strconv.ParseInt
	if !n.IsNumeric() {
		return 0, fmt.Errorf("%w: %q", ErrAlphanumericNumber, n)
	}
	return strconv.ParseInt(string(n), 10, 64)
}

// Cut splits the Number at the first occurrence of sep.
// Returns the text before and after the separator, and a boolean indicating
// whether the separator was found.
func (n Number) Cut(sep string) (before, after Number, found bool) {
	left, right, found := strings.Cut(string(n), sep)
	return Number(left), Number(right), found
}

// TrimLeadingZeros removes leading zero characters from the Number.
// Returns an empty string if the Number consists only of zeros.
func (n Number) TrimLeadingZeros() Number {
	for i, r := range n {
		if r != '0' {
			return n[i:]
		}
	}
	return ""
}

// Nullable converts the Number to a NullableNumber type.
func (n Number) Nullable() NullableNumber {
	return NullableNumber(n)
}

// String returns the string representation of the Number.
func (n Number) String() string {
	return string(n)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// Trims leading and trailing whitespace from the text before validation.
func (n *Number) UnmarshalText(text []byte) error {
	no, err := NumberFrom(string(text))
	if err != nil {
		return err
	}
	*n = no
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
// Returns nil for SQL NULL if the Number is empty.
func (n Number) Value() (driver.Value, error) {
	return NullableNumber(n).Value()
}

// Scan implements the database/sql.Scanner interface
func (n *Number) Scan(value any) error {
	switch x := value.(type) {
	case string:
		no, err := NumberFrom(x)
		if err != nil {
			return err
		}
		*n = no
		return nil

	case []byte:
		no, err := NumberFrom(string(x))
		if err != nil {
			return err
		}
		*n = no
		return nil

	case int64:
		if x < 0 {
			return fmt.Errorf("%w: %v", ErrInvalidNumber, x)
		}
		*n = Number(strconv.FormatInt(x, 10))
		return nil

	case float64:
		if x < 0 {
			return fmt.Errorf("%w: %v", ErrInvalidNumber, x)
		}
		*n = Number(strconv.FormatFloat(x, 'f', -1, 64))
		return nil

	default:
		return fmt.Errorf("can't scan %T as Number", value)
	}
}

// MarshalJSON implements encoding/json.Marshaler.
// Returns the JSON null value if the Number is empty.
func (n Number) MarshalJSON() ([]byte, error) {
	return NullableNumber(n).MarshalJSON()
}

// UnmarshalJSON implements encoding/json.Unmarshaler
func (n *Number) UnmarshalJSON(j []byte) error {
	var str string
	// First try unmarshalling to string
	err := json.Unmarshal(j, &str)
	if err != nil {
		// Try unmarshalling to uint64
		var u uint64
		err = json.Unmarshal(j, &u)
		if err != nil {
			return fmt.Errorf("%w from JSON: %s", ErrInvalidNumber, j)
		}
		*n = Number(strconv.FormatUint(u, 10))
		return nil
	}
	no, err := NumberFrom(str)
	if err != nil {
		return fmt.Errorf("%w from JSON: %s", err, j)
	}
	*n = no
	return nil
}

func (Number) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:   "Account Number",
		Type:    "string",
		Pattern: NumberRegex,
	}
}

// func (n Number) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
// 	return e.EncodeElement(string(n), start)
// }

func (n *Number) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string
	err := d.DecodeElement(&str, &start)
	if err != nil {
		return err
	}
	no, err := NumberFrom(str)
	if err != nil {
		return err
	}
	*n = no
	return nil
}
