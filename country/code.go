package country

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/domonda/go-types/strutil"
	"github.com/invopop/jsonschema"
)

const Invalid Code = ""

// Code for a country according ISO 3166-1 alpha 2.
//
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty Code string as SQL NULL.
//
// See NullableCode for a nullable version of this type.
type Code string

// Valid returns true if the normalized Code is valid.
//
// See Normalized for the normalization process.
func (c Code) Valid() bool {
	_, err := c.Normalized()
	return err == nil
}

// Validate returns an error if the normalized Code is not valid.
//
// See Normalized for the normalization process.
func (c Code) Validate() error {
	_, err := c.Normalized()
	return err
}

// Normalized uses the whitespace-trimmed uppercase
// string of the code to look up and return the
// standard ISO 3166-1 alpha 2 code.
//
// If not found then AltCodes is used to look
// up alternative code and name mappings using
// the whitespace-trimmed uppercase code.
//
// If no mapping exists then the original Code
// is returned unchanged together with an error.
func (c Code) Normalized() (Code, error) {
	norm := Code(strings.ToUpper(strutil.TrimSpace(string(c))))
	if _, ok := countryMap[norm]; ok {
		return norm, nil
	}
	if norm, ok := AltCodes[string(norm)]; ok {
		return norm, nil
	}
	return c, fmt.Errorf("invalid country.Code: '%s'", string(c))
}

// IsEU indicates if a country is member of the European Union
func (c Code) IsEU() bool {
	norm, err := c.Normalized()
	if err != nil {
		return false
	}
	_, ok := euCountries[norm]
	return ok
}

func (c Code) EnglishName() string {
	norm, err := c.Normalized()
	if err != nil {
		return ""
	}
	return countryMap[norm]
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
	norm, _ := c.Normalized()
	return string(norm), nil
}

// MarshalJSON implements encoding/json.Marshaler
func (c Code) MarshalJSON() ([]byte, error) {
	norm, _ := c.Normalized()
	return json.Marshal(string(norm))
}

func (Code) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:   "Country Code",
		Type:    "string",
		Pattern: "^[A-Z]{2}$",
	}
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// If validate is true, the source string is checked
// for validity before it is assigned to the type.
//
// If validate is false and the source string
// can still be assigned in some non-normalized way
// it will be assigned without returning an error.
func (c *Code) ScanString(source string, validate bool) error {
	code, err := Code(source).Normalized()
	if err != nil {
		if validate {
			return err
		}
		code = Code(source)
	}
	*c = code
	return nil
}

// String returns the normalized code if possible,
// else it will be returned unchanged as string.
//
// String implements the fmt.Stringer interface.
func (c Code) String() string {
	norm, _ := c.Normalized()
	return string(norm)
}

// Nullable returns the Code as NullableCode.
// Country code Invalid is returned as Null.
func (c Code) Nullable() NullableCode {
	return NullableCode(c)
}
