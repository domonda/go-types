// Package language provides comprehensive language code handling and validation
// based on ISO 639-1 standards for Go applications.
//
// The package includes:
// - ISO 639-1 two-character language code validation and normalization
// - Language name mapping and retrieval
// - Database integration (Scanner/Valuer interfaces)
// - JSON marshalling/unmarshalling
// - Support for common language codes and names
package language

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
)

// Code represents a language code in its normalized form as an ISO 639-1 two-character language code.
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and treats an empty Code string as SQL NULL value.
type Code string

// Valid returns true if the Code is a valid ISO 639-1 language code.
func (c Code) Valid() bool {
	_, ok := codeNames[c]
	return ok
}

// Normalized returns the normalized language code or an error if invalid.
// TODO: normalize 3 letter codes https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
// TODO: normalize BCP-47 language codes, such as "en-US" or "sr-Latn"
// http://www.unicode.org/reports/tr35/#Unicode_locale_identifier.
func (c Code) Normalized() (Code, error) {
	normalized := Code(strings.ToLower(string(c)))
	if _, ok := codeNames[normalized]; !ok {
		return c, fmt.Errorf("invalid language.Code: %q", string(c))
	}
	return normalized, nil
}

// LanguageName returns the English name of the language for the code.
func (c Code) LanguageName() string {
	return codeNames[c]
}

// String returns the normalized code if possible, else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (c Code) String() string {
	norm, err := c.Normalized()
	if err != nil {
		return string(c)
	}
	return string(norm)
}

// Scan implements the database/sql.Scanner interface.
func (c *Code) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*c = Code(x)
	case []byte:
		*c = Code(x)
	case nil:
		*c = Null
	default:
		return fmt.Errorf("can't scan SQL value of type %T as language.Code", value)
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

// JSONSchema returns the JSON schema definition for the Code type.
func (Code) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:   "Language Code",
		Type:    "string",
		Pattern: `^[a-z]{2}$`,
	}
}
