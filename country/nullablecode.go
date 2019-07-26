package country

import (
	"database/sql/driver"
	"strings"

	"github.com/domonda/errors"
)

const Null NullableCode = ""

// NullableCode for a country according ISO 3166-1 alpha 2.
// NullableCode implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty NullableCode string as SQL NULL.
// Null.Valid() or NullableCode("").Valid() will return true.
type NullableCode string

func (c NullableCode) Valid() bool {
	return c == Null || Code(c).Valid()
}

func (c NullableCode) ValidAndNotNull() bool {
	return Code(c).Valid()
}

func (c NullableCode) Validate() error {
	if !c.Valid() {
		return errors.Errorf("invalid country-code: %q", c)
	}
	return nil
}

func (c NullableCode) Normalized() (NullableCode, error) {
	normalized := NullableCode(strings.ToUpper(string(c)))
	err := normalized.Validate()
	if err != nil {
		return Null, err
	}
	return normalized, nil
}

func (c NullableCode) CountryName() string {
	return Code(c).CountryName()
}

func (c NullableCode) Code() Code {
	return Code(c)
}

// Scan implements the database/sql.Scanner interface.
func (c *NullableCode) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*c = NullableCode(x)
	case []byte:
		*c = NullableCode(x)
	case nil:
		*c = Null
	default:
		return errors.Errorf("can't scan SQL value of type %T as country.NullableCode", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (c NullableCode) Value() (driver.Value, error) {
	if c == Null {
		return nil, nil
	}
	return string(c), nil
}

// MarshalJSON implements encoding/json.Marshaler
func (c NullableCode) MarshalJSON() ([]byte, error) {
	if c == Null {
		return []byte("null"), nil
	}
	return []byte(c), nil
}

// AssignString tries to parse and assign the passed
// source string as value of the implementing object.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for normalized and nil for err.
// AssignString implements strfmt.StringAssignable
func (c *NullableCode) AssignString(source string) (normalized bool, err error) {
	newNullableCode := NullableCode(strings.ToUpper(source))
	if !newNullableCode.Valid() {
		return false, errors.Errorf("invalid country-code: '%s'", source)
	}
	*c = newNullableCode
	return newNullableCode == NullableCode(source), nil
}
