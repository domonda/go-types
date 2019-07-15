package language

import (
	"database/sql/driver"
	"strings"

	"github.com/domonda/errors"
)

// Code in its normalized form a ISO 639-1 two character language code.
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty Code string as SQL NULL value.
type Code string

func (c Code) Valid() bool {
	_, ok := codeNames[c]
	return ok
}

func (c Code) Normalized() Code {
	// TODO normalize 3 letter codes https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
	// TODO normalize BCP-47 language codes, such as "en-US" or "sr-Latn"
	// http://www.unicode.org/reports/tr35/#Unicode_locale_identifier.
	normalized := Code(strings.ToLower(string(c)))
	if _, ok := codeNames[normalized]; !ok {
		return ""
	}
	return normalized
}

func (c Code) LanguageName() string {
	return codeNames[c]
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
		return errors.Errorf("can't scan SQL value of type %T as language.Code", value)
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
