package country

import (
	"database/sql/driver"
	"encoding/json"
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
	_, ok := countryMap[c.normalized()]
	return ok
}

func (c Code) Validate() error {
	if c.Valid() {
		return nil
	}
	return fmt.Errorf("invalid country.Code: %q", string(c))
}

func (c Code) Normalized() (Code, error) {
	norm := c.normalized()
	if _, ok := countryMap[norm]; !ok {
		return "", fmt.Errorf("invalid country.Code: %q", string(c))
	}
	return norm, nil
}

func (c Code) normalized() Code {
	return Code(strings.ToUpper(strings.TrimSpace(string(c))))
}

// NormalizedWithAltCodes uses AltCodes to map
// to ISO 3166-1 alpha 2 codes or return the
// result of Normalized() if no mapping exists.
func (c Code) NormalizedWithAltCodes() (Code, error) {
	if norm, ok := AltCodes[strings.ToUpper(strings.TrimSpace(string(c)))]; ok {
		return norm, nil
	}
	return c.Normalized()
}

// IsEU indicates if a country is member of the European Union
func (c Code) IsEU() bool {
	_, ok := euCountries[c.normalized()]
	return ok
}

func (c Code) EnglishName() string {
	return countryMap[c.normalized()]
}

// Scan implements the database/sql.Scanner interface.
func (c *Code) Scan(value any) error {
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
	return string(c.normalized()), nil
}

// MarshalJSON implements encoding/json.Marshaler
func (c Code) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c.normalized()))
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
	return string(c.normalized())
}

// Nullable returns the Code as NullableCode.
// Country code Invalid is returned as Null.
func (c Code) Nullable() NullableCode {
	return NullableCode(c)
}
