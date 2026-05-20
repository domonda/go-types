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
	"sync"

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

// ParseCode parses and normalizes a language identifier given in any of
// the formats this package understands, returning the canonical ISO
// 639-1 two-letter Code.
//
// It tries the cheapest and most common interpretations first and only
// falls back to the more expensive ones:
//
//  1. An input that is already a canonical ISO 639-1 code ("en", "de")
//     is returned unchanged without any further work.
//  2. Otherwise the input is run through Normalized, which accepts case
//     and whitespace variants, ISO 639-2/T and 639-2/B codes, ISO 639-3
//     codes with a 639-1 equivalent, BCP-47 tags, and POSIX locale
//     strings ("EN", " eng ", "ger", "en-US", "zh_Hant").
//  3. As a last resort the input is matched case-insensitively against
//     the English language names of the ISO 639-1 and ISO 639-3 tables
//     ("German", "english", "Catalan").
//
// ParseCode returns an error wrapping the original input if no ISO 639-1
// code can be derived. Like Normalized, the error preserves the
// unnormalized value for diagnostics.
func ParseCode(str string) (Code, error) {
	// 1. Short path: the input is already a canonical ISO 639-1 code.
	if c := Code(str); c.Valid() {
		return c, nil
	}
	// 2. Codes and tags: case and whitespace variants, ISO 639-2/3,
	//    639-2/B bibliographic variants, BCP-47 tags, POSIX locales.
	c, err := Code(str).Normalized()
	if err == nil {
		return c, nil
	}
	// 3. English language name.
	if named, ok := codeForName(str); ok {
		return named, nil
	}
	// Reuse the input-preserving Code and error returned by Normalized.
	return c, err
}

// nameToCode is the lazily-built reverse index from a lower-cased English
// language name to its ISO 639-1 Code. It is consulted only by ParseCode
// when the input is not a recognizable code or tag, so it is built on
// first use rather than at package initialization.
var nameToCode = sync.OnceValue(func() map[string]Code {
	m := make(map[string]Code, len(codeNames)+len(iso6393Names))
	// Index the ISO 639-3 names first so the curated ISO 639-1 names
	// added below take precedence on any name collision.
	for code3, name := range iso6393Names {
		if code1, ok := iso6393To1[code3]; ok {
			m[strings.ToLower(name)] = code1
		}
	}
	for code, names := range codeNames {
		// A codeNames value may list synonyms separated by "; ".
		for name := range strings.SplitSeq(names, ";") {
			if name = strings.ToLower(strings.TrimSpace(name)); name != "" {
				m[name] = code
			}
		}
	}
	return m
})

// codeForName resolves an English language name to its ISO 639-1 Code,
// matching case-insensitively. The bool result reports whether a match
// was found.
func codeForName(name string) (Code, bool) {
	key := strings.ToLower(strutil.TrimSpace(name))
	if key == "" {
		return Null, false
	}
	c, ok := nameToCode()[key]
	return c, ok
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
