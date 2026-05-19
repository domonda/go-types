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

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/strutil"
)

// Code represents a language code in its normalized form as an ISO 639-1 two-character language code.
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and treats an empty Code string as SQL NULL value.
type Code string

// Compile-time check that Code implements types.NormalizableValidator[Code]
var _ types.NormalizableValidator[Code] = Code("")

// Valid returns true if the Code is a valid ISO 639-1 language code.
func (c Code) Valid() bool {
	_, ok := codeNames[c]
	return ok
}

// ValidAndNormalized returns true if the Code is valid and already normalized.
func (c Code) ValidAndNormalized() bool {
	norm, err := c.Normalized()
	return err == nil && c == norm
}

// Normalized returns the normalized ISO 639-1 two-letter language code.
//
// Accepted input shapes:
//   - ISO 639-1 two-letter codes, case-insensitive, with surrounding
//     whitespace trimmed ("en", " EN ", "En").
//   - ISO 639-2 three-letter codes, both the bibliographic (B) and
//     terminologic (T) variants where they differ ("eng", "deu", "ger").
//   - ISO 639-3 three-letter codes, where a 639-1 equivalent exists
//     ("eng", "deu", "fra"). 639-3 codes for languages without a 639-1
//     assignment are rejected.
//   - BCP-47 language tags: only the leading language subtag is
//     consulted; script, region, variant, and extension subtags are
//     dropped ("en-US" → "en", "zh-Hant-CN" → "zh", "sr-Latn" → "sr").
//     POSIX-style underscores ("en_US") are accepted as separators.
//
// Returns an error wrapping the original input if no 639-1 code can be
// derived. The error preserves the unnormalized value for diagnostics.
//
// See:
//   - https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
//   - https://www.rfc-editor.org/info/bcp47
//   - https://www.unicode.org/reports/tr35/#Unicode_locale_identifier
func (c Code) Normalized() (Code, error) {
	s := strings.ToLower(strutil.TrimSpace(string(c)))
	if s == "" {
		return c, fmt.Errorf("invalid language.Code: %q", string(c))
	}
	// Strip BCP-47 region/script/variant subtags by keeping only the
	// primary language subtag. POSIX locale strings ("en_US") use '_'.
	if i := strings.IndexAny(s, "-_"); i >= 0 {
		s = s[:i]
	}
	switch len(s) {
	case 2:
		if _, ok := codeNames[Code(s)]; ok {
			return Code(s), nil
		}
	case 3:
		if code, ok := iso6393To1[s]; ok {
			return code, nil
		}
	}
	return c, fmt.Errorf("invalid language.Code: %q", string(c))
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
