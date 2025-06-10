package account

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"
)

var (
	_ fmt.Stringer     = NullableNumber("")
	_ driver.Valuer    = NullableNumber("")
	_ sql.Scanner      = new(NullableNumber)
	_ json.Marshaler   = NullableNumber("")
	_ json.Unmarshaler = new(NullableNumber)
	// _ xml.Marshaler    = NullableNumber("")
	_ xml.Unmarshaler = new(NullableNumber)
)

const NumberNull NullableNumber = ""

// NullableNumber represents an account number with the option for alphanumerical characters
// where an empty string represents NULL.
type NullableNumber string

// NullableNumberFrom returns an NullableNumber for the passed string
// or an error if the string is not a valid account number.
// It trims leading and trailing whitespace from the passed string
// and interprets an empty string as null.
func NullableNumberFrom(str string) (NullableNumber, error) {
	str = strings.TrimSpace(str)
	if err := NullableNumber(str).Validate(); err != nil {
		return "", err
	}
	return NullableNumber(str), nil
}

// NullableNumberFromUint returns an NullableNumber
// for the passed uint64 interpreting 0 as null.
func NullableNumberFromUint(u uint64) NullableNumber {
	if u == 0 {
		return NumberNull
	}
	return NullableNumber(strconv.FormatUint(u, 10))
}

// Valid returns true if the Number is null
// or matches the regular expression `^[0-9A-Za-z_]+$`
func (n NullableNumber) Valid() bool {
	return n == NumberNull || Number(n).Valid()
}

// Validate returns a wrapped ErrInvalidNumber
// error if the NullableNumber is not null and
// does not match the regular expression `^[0-9A-Za-z_]+$`
func (n NullableNumber) Validate() error {
	if !n.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidNumber, n)
	}
	return nil
}

func (n NullableNumber) HasPrefix(prefix string) bool {
	return strings.HasPrefix(string(n), prefix)
}

func (n NullableNumber) HasSuffix(suffix string) bool {
	return strings.HasSuffix(string(n), suffix)
}

// IsNumeric indicates if the Number only contains digits.
func (n NullableNumber) IsNumeric() bool {
	return Number(n).IsNumeric()
}

// ValidateNumeric returns a wrapped ErrAlphanumericNumber
// error if the Number is not numeric.
func (n NullableNumber) ValidateNumeric() error {
	if !n.IsNumeric() {
		return fmt.Errorf("%w: %q", ErrAlphanumericNumber, n)
	}
	return nil
}

// IsNull returns true if the string is empty.
// IsNull implements the Nullable interface.
func (n NullableNumber) IsNull() bool {
	return n == NumberNull
}

// IsNotNull returns true if the string is not empty.
func (n NullableNumber) IsNotNull() bool {
	return n != NumberNull
}

// SetNull sets and empty string representing null
func (n *NullableNumber) SetNull() {
	*n = NumberNull
}

// Get returns the non nullable Number
// or panics if the NullableNumber is null.
// Note: check with IsNull before using Get!
func (n NullableNumber) Get() Number {
	if n.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", n))
	}
	return Number(n)
}

func (n NullableNumber) Uint() (uint64, error) {
	if n == NumberNull {
		return 0, nil
	}
	return Number(n).Uint()
}

func (n NullableNumber) UintPtr() (*uint64, error) {
	if n == NumberNull {
		return nil, nil
	}
	u, err := n.Uint()
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (n NullableNumber) Int() (int64, error) {
	if n == NumberNull {
		return 0, nil
	}
	return Number(n).Int()

}

func (n NullableNumber) Cut(sep string) (before, after NullableNumber, found bool) {
	left, right, found := strings.Cut(string(n), sep)
	return NullableNumber(left), NullableNumber(right), found
}

func (n NullableNumber) TrimLeadingZeros() NullableNumber {
	for i, r := range n {
		if r != '0' {
			return n[i:]
		}
	}
	return ""
}

func (n NullableNumber) String() string {
	return string(n)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// It trims leading and trailing whitespace from the text.
func (n *NullableNumber) UnmarshalText(text []byte) error {
	no, err := NullableNumberFrom(string(text))
	if err != nil {
		return err
	}
	*n = no
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
// Returns nil for SQL NULL if the Number is empty.
func (n NullableNumber) Value() (driver.Value, error) {
	str := strings.TrimSpace(string(n))
	if str == "" {
		return nil, nil
	}
	return str, nil
}

// Scan implements the database/sql.Scanner interface
func (n *NullableNumber) Scan(value any) error {
	switch x := value.(type) {
	case nil:
		*n = NumberNull
		return nil

	case string:
		no, err := NullableNumberFrom(x)
		if err != nil {
			return err
		}
		*n = no
		return nil

	case []byte:
		no, err := NullableNumberFrom(string(x))
		if err != nil {
			return err
		}
		*n = no
		return nil

	case int64:
		if x < 0 {
			return fmt.Errorf("%w: %v", ErrInvalidNumber, x)
		}
		if x == 0 {
			*n = NumberNull
		} else {
			*n = NullableNumber(strconv.FormatInt(x, 10))
		}
		return nil

	case float64:
		if x < 0 {
			return fmt.Errorf("%w: %v", ErrInvalidNumber, x)
		}
		if x == 0 {
			*n = NumberNull
		} else {
			*n = NullableNumber(strconv.FormatFloat(x, 'f', -1, 64))
		}
		return nil

	default:
		return fmt.Errorf("can't scan %T as NullableNumber", value)
	}
}

// MarshalJSON implements encoding/json.Marshaler
func (n NullableNumber) MarshalJSON() ([]byte, error) {
	str := strings.TrimSpace(string(n))
	if str == "" {
		return []byte(`null`), nil
	}
	return json.Marshal(str)
}

// UnmarshalJSON implements encoding/json.Unmarshaler
func (n *NullableNumber) UnmarshalJSON(j []byte) error {
	if bytes.Equal(j, []byte(`null`)) {
		*n = NumberNull
		return nil
	}
	return (*Number)(n).UnmarshalJSON(j)
}

func (NullableNumber) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable Account Number",
		OneOf: []*jsonschema.Schema{
			{
				Type:    "string",
				Pattern: NumberRegex,
			},
			{Type: "null"},
		},
		Default: NumberNull,
	}
}

// func (n NullableNumber) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
// 	return e.EncodeElement(string(n), start)
// }

func (n *NullableNumber) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string
	err := d.DecodeElement(&str, &start)
	if err != nil {
		return err
	}
	no, err := NullableNumberFrom(str)
	if err != nil {
		return err
	}
	*n = no
	return nil
}
