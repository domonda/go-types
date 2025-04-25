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

	"github.com/domonda/go-errs"
	"github.com/invopop/jsonschema"
)

// Errors
const (
	ErrInvalidNumber      errs.Sentinel = "invalid account number"
	ErrAlphanumericNumber errs.Sentinel = "account number is alphanumeric"
)

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

// Number represents an account number with the option for alphanumerical characters.
type Number string

// NumberFrom returns an Number for the passed string
// or an error if the string is not a valid account number.
// It trims leading and trailing whitespace from the passed string.
func NumberFrom(str string) (Number, error) {
	str = strings.TrimSpace(str)
	if err := Number(str).Validate(); err != nil {
		return "", err
	}
	return Number(str), nil
}

func NumberFromUint(u uint64) Number {
	return Number(strconv.FormatUint(u, 10))
}

// Valid returns true if the Number matches
// the regular expression `^[0-9A-Za-z][0-9A-Za-z_\-\/:.;,]*$`
func (n Number) Valid() bool {
	return numberRegexp.MatchString(string(n))
}

// Validate returns a wrapped ErrInvalidNumber
// error if the Number does not match
// the regular expression `^[0-9A-Za-z][0-9A-Za-z_\-\/:.;,]*$`
func (n Number) Validate() error {
	if !n.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidNumber, n)
	}
	return nil
}

func (n Number) HasPrefix(prefix string) bool {
	return strings.HasPrefix(string(n), prefix)
}

func (n Number) HasSuffix(suffix string) bool {
	return strings.HasSuffix(string(n), suffix)
}

// IsNumeric indicates if the Number only contains digits.
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

// ValidateNumeric returns a wrapped ErrAlphanumericNumber
// error if the Number is not numeric.
func (n Number) ValidateNumeric() error {
	if !n.IsNumeric() {
		return fmt.Errorf("%w: %q", ErrAlphanumericNumber, n)
	}
	return nil
}

func (n Number) Uint() (uint64, error) {
	// Check upfront to prevent non digit strings like hex numbers
	// to be parsed by strconv.ParseUint
	if err := n.ValidateNumeric(); err != nil {
		return 0, err
	}
	return strconv.ParseUint(string(n), 10, 64)
}

func (n Number) UintPtr() (*uint64, error) {
	u, err := n.Uint()
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (n Number) Int() (int64, error) {
	// Check upfront to prevent non digit strings like hex numbers
	// to be parsed by strconv.ParseInt
	if !n.IsNumeric() {
		return 0, fmt.Errorf("%w: %q", ErrAlphanumericNumber, n)
	}
	return strconv.ParseInt(string(n), 10, 64)
}

func (n Number) Cut(sep string) (before, after Number, found bool) {
	left, right, found := strings.Cut(string(n), sep)
	return Number(left), Number(right), found
}

func (n Number) TrimLeadingZeros() Number {
	for i, r := range n {
		if r != '0' {
			return n[i:]
		}
	}
	return ""
}

func (n Number) Nullable() NullableNumber {
	return NullableNumber(n)
}

func (n Number) String() string {
	return string(n)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// It trims leading and trailing whitespace from the text.
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
