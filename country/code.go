package country

import (
	"database/sql/driver"
	"strings"

	"github.com/domonda/errors"
)

// Code for a country according ISO 3166-1 alpha 2.
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty Code string as SQL NULL.
type Code string

func (c Code) Valid() bool {
	_, ok := countryMap[c]
	return ok
}

func (c Code) CountryName() string {
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
		*c = Null
	default:
		return errors.Errorf("can't scan SQL value of type %T as country.Code", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (c Code) Value() (driver.Value, error) {
	if c == Null {
		return nil, nil
	}
	return string(c), nil
}

// AssignString tries to parse and assign the passed
// source string as value of the implementing object.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for normalized and nil for err.
// AssignString implements strfmt.StringAssignable
func (c *Code) AssignString(source string) (normalized bool, err error) {
	newCode := Code(strings.ToUpper(source))
	if !newCode.Valid() {
		return false, errors.Errorf("invalid country code: %#v", source)
	}
	*c = newCode
	return newCode == Code(source), nil
}
