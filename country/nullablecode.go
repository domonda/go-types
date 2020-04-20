package country

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

const Null NullableCode = ""

// NullableCode for a country according ISO 3166-1 alpha 2.
// NullableCode implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty NullableCode string as SQL NULL.
// Null.Valid() or NullableCode("").Valid() will return true.
type NullableCode string

// IsNull returns true if the NullableCode is null
func (n NullableCode) IsNull() bool {
	return n == Null
}

func (n NullableCode) Valid() bool {
	return n == Null || Code(n).Valid()
}

func (n NullableCode) ValidAndNotNull() bool {
	return Code(n).Valid()
}

func (n NullableCode) Validate() error {
	if !n.Valid() {
		return fmt.Errorf("invalid country.Code: %q", string(n))
	}
	return nil
}

func (n NullableCode) Normalized() (NullableCode, error) {
	normalized := NullableCode(strings.ToUpper(string(n)))
	err := normalized.Validate()
	if err != nil {
		return Null, err
	}
	return normalized, nil
}

func (n NullableCode) CountryName() string {
	return Code(n).CountryName()
}

func (n NullableCode) Code() Code {
	return Code(n)
}

// Scan implements the database/sql.Scanner interface.
func (n *NullableCode) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*n = NullableCode(x)
	case []byte:
		*n = NullableCode(x)
	case nil:
		*n = Null
	default:
		return fmt.Errorf("can't scan SQL value of type %T as country.NullableCode", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n NullableCode) Value() (driver.Value, error) {
	if n == Null {
		return nil, nil
	}
	return string(n), nil
}

// MarshalJSON implements encoding/json.Marshaler
func (n NullableCode) MarshalJSON() ([]byte, error) {
	if n == Null {
		return []byte("null"), nil
	}
	return []byte(`"` + n + `"`), nil
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (n *NullableCode) ScanString(source string) (normalized bool, err error) {
	newNullableCode := NullableCode(strings.ToUpper(source))
	if !newNullableCode.Valid() {
		return false, fmt.Errorf("invalid country.Code: '%s'", source)
	}
	*n = newNullableCode
	return newNullableCode == NullableCode(source), nil
}

// String returns the normalized code if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (n NullableCode) String() string {
	norm, err := n.Normalized()
	if err != nil {
		return string(n)
	}
	return string(norm)
}
