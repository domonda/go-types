package account

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

var (
	_ fmt.Stringer                 = NullableNumber("")
	_ driver.Valuer                = NullableNumber("")
	_ sql.Scanner                  = new(NullableNumber)
	_ json.Marshaler               = NullableNumber("")
	_ json.Unmarshaler             = new(NullableNumber)
	_ xml.Marshaler                = NullableNumber("")
	_ xml.Unmarshaler              = new(NullableNumber)
	_ nullable.NullSetable[Number] = (*NullableNumber)(nil)
)

// NumberNull is the null value for NullableNumber, represented as an empty string.
const NumberNull NullableNumber = ""

// NullableNumber represents an account number with the option for alphanumerical characters
// where an empty string represents NULL.
type NullableNumber string

// NullableNumberFrom returns an NullableNumber for the passed string
// or an error if the string is not a valid account number.
// It trims leading and trailing whitespace from the passed string
// and interprets an empty string as null.
func NullableNumberFrom(str string) (NullableNumber, error) {
	str = strutil.TrimSpace(str)
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

// Valid returns true if the NullableNumber is null (see IsNull)
// or matches NumberRegexp (DefaultNumberRegex unless reassigned).
func (n NullableNumber) Valid() bool {
	return n.IsNull() || Number(n).Valid()
}

// Validate returns a wrapped ErrInvalidNumber
// error if the NullableNumber is not null and does not match
// NumberRegexp (DefaultNumberRegex unless reassigned).
func (n NullableNumber) Validate() error {
	if !n.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidNumber, n)
	}
	return nil
}

// HasPrefix reports whether the NullableNumber begins with the given prefix string.
func (n NullableNumber) HasPrefix(prefix string) bool {
	return strings.HasPrefix(string(n), prefix)
}

// HasSuffix reports whether the NullableNumber ends with the given suffix string.
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

// IsNull returns true if the string is empty or consists only of
// the runes strutil.TrimSpace removes: whitespace, control and other
// non-printable runes, and invalid UTF-8,
// which Value writes as SQL NULL and MarshalJSON as null.
// IsNull implements the Nullable interface.
func (n NullableNumber) IsNull() bool {
	return strutil.TrimSpace(n) == NumberNull
}

// IsNotNull returns true if the string is not empty or whitespace only.
func (n NullableNumber) IsNotNull() bool {
	return !n.IsNull()
}

// SetNull sets and empty string representing null
func (n *NullableNumber) SetNull() {
	*n = NumberNull
}

// Set sets a Number for this NullableNumber
func (n *NullableNumber) Set(num Number) {
	*n = NullableNumber(num)
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

// GetOr returns the non nullable Number value
// or the passed defaultNumber if the NullableNumber is null.
func (n NullableNumber) GetOr(defaultNumber Number) Number {
	if n.IsNull() {
		return defaultNumber
	}
	return Number(n)
}

// Uint converts the NullableNumber to a uint64.
// Returns 0 and no error if the NullableNumber is null.
// Returns an error if the value is not purely numeric.
func (n NullableNumber) Uint() (uint64, error) {
	if n.IsNull() {
		return 0, nil
	}
	return Number(n).Uint()
}

// UintPtr converts the NullableNumber to a *uint64.
// Returns nil and no error if the NullableNumber is null.
// Returns an error if the value is not purely numeric.
func (n NullableNumber) UintPtr() (*uint64, error) {
	if n.IsNull() {
		return nil, nil
	}
	u, err := n.Uint()
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Int converts the NullableNumber to an int64.
// Returns 0 and no error if the NullableNumber is null.
// Returns an error if the value is not purely numeric.
func (n NullableNumber) Int() (int64, error) {
	if n.IsNull() {
		return 0, nil
	}
	return Number(n).Int()

}

// Cut splits the NullableNumber at the first occurrence of sep.
// Returns the text before and after the separator, and a boolean indicating
// whether the separator was found.
func (n NullableNumber) Cut(sep string) (before, after NullableNumber, found bool) {
	left, right, found := strings.Cut(string(n), sep)
	return NullableNumber(left), NullableNumber(right), found
}

// TrimLeadingZeros removes leading zero characters from the NullableNumber.
// Returns an empty NullableNumber (null) if the value consists only of zeros.
func (n NullableNumber) TrimLeadingZeros() NullableNumber {
	for i, r := range n {
		if r != '0' {
			return n[i:]
		}
	}
	return ""
}

// String returns the string representation of the NullableNumber,
// returning an empty string if the NullableNumber is null (see IsNull).
func (n NullableNumber) String() string {
	if n.IsNull() {
		return ""
	}
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
	str := strutil.TrimSpace(string(n))
	if str == "" {
		return nil, nil
	}
	return str, nil
}

// Scan implements the database/sql.Scanner interface.
// A float64 value must be a non-negative integer,
// NaN, infinities, and fractions return an error.
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
		if math.IsNaN(x) || math.IsInf(x, 0) || x < 0 || x != math.Trunc(x) {
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
	str := strutil.TrimSpace(string(n))
	if str == "" {
		return []byte(`null`), nil
	}
	return json.Marshal(str)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// JSON null, an empty or whitespace only JSON string,
// and the JSON number 0 give NumberNull like Scan and UnmarshalText.
func (n *NullableNumber) UnmarshalJSON(j []byte) error {
	if bytes.Equal(j, []byte(`null`)) || bytes.Equal(bytes.TrimSpace(j), []byte(`0`)) {
		*n = NumberNull
		return nil
	}
	var str string
	if json.Unmarshal(j, &str) == nil {
		// Trimmed like by Scan and UnmarshalText
		no, err := NullableNumberFrom(str)
		if err != nil {
			return fmt.Errorf("%w from JSON: %s", err, j)
		}
		*n = no
		return nil
	}
	return (*Number)(n).UnmarshalJSON(j)
}

// JSONSchema returns a JSON schema for NullableNumber, allowing either
// a string matching the account number pattern of NumberRegexp
// (DefaultNumberRegex unless reassigned) or null.
func (NullableNumber) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable Account Number",
		OneOf: []*jsonschema.Schema{
			{
				Type:    "string",
				Pattern: NumberRegexp.String(),
			},
			{Type: "null"},
		},
		Default: NumberNull,
	}
}

// MarshalXML implements encoding/xml.Marshaler.
// Encodes the NullableNumber as an XML element containing its
// whitespace trimmed string value, which is empty for null.
func (n NullableNumber) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(strutil.TrimSpace(string(n)), start)
}

// UnmarshalXML implements encoding/xml.Unmarshaler.
// Decodes an XML element into a NullableNumber, validating the account number format.
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
