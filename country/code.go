package country

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

const Invalid Code = ""

// Code for a country according ISO 3166-1 alpha 2.
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty Code string as SQL NULL.
// See NullableCode
type Code string

func (c Code) Valid() bool {
	_, ok := countryMap[c]
	return ok
}

func (c Code) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid country.Code: %q", string(c))
	}
	return nil
}

func (c Code) Normalized() (Code, error) {
	normalized := Code(strings.ToUpper(string(c)))
	err := normalized.Validate()
	if err != nil {
		return Invalid, err
	}
	return normalized, nil
}

func (c Code) EnglishName() string {
	return countryMap[c]
}

// Scan implements the database/sql.Scanner interface.
func (c *Code) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*c = Code(x)
	case []byte:
		*c = Code(x)
	case nil:
		*c = Invalid
	default:
		return fmt.Errorf("can't scan SQL value of type %T as country.Code", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (c Code) Value() (driver.Value, error) {
	if c == Invalid {
		return nil, nil
	}
	return string(c), nil
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (c *Code) ScanString(source string) (normalized bool, err error) {
	newCode := Code(strings.ToUpper(source))
	if !newCode.Valid() {
		return false, fmt.Errorf("invalid country.Code: '%s'", source)
	}
	*c = newCode
	return newCode == Code(source), nil
}

// String returns the normalized code if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (c Code) String() string {
	norm, err := c.Normalized()
	if err != nil {
		return string(c)
	}
	return string(norm)
}

// Nullable returns the Code as NullableCode.
// Country code Invalid is returned as Null.
func (c Code) Nullable() NullableCode {
	return NullableCode(c)
}
