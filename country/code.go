// Package country provides comprehensive country code handling and validation
// based on ISO 3166-1 alpha-2 standards for Go applications.
//
// The package includes:
// - ISO 3166-1 alpha-2 country code validation and normalization
// - Alternative country code mappings (ITU codes, German names, etc.)
// - European Union membership checking
// - Database integration (Scanner/Valuer interfaces)
// - JSON marshalling/unmarshalling
// - Nullable country code support
package country

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/strutil"
)

// Invalid represents an invalid country code.
const Invalid Code = ""

// Compile-time check that Code implements types.NormalizableValidator[Code]
var _ types.NormalizableValidator[Code] = Code("")

// Code represents a country code according to ISO 3166-1 alpha-2 standard.
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and treats an empty Code string as SQL NULL.
// See NullableCode for a nullable version of this type.
type Code string

// Valid returns true if the normalized Code is valid.
// See Normalized for the normalization process.
func (c Code) Valid() bool {
	_, err := c.Normalized()
	return err == nil
}

// ValidAndNormalized returns true if the Code is valid and already normalized.
func (c Code) ValidAndNormalized() bool {
	norm, err := c.Normalized()
	return err == nil && c == norm
}

// Validate returns an error if the normalized Code is not valid.
// See Normalized for the normalization process.
func (c Code) Validate() error {
	_, err := c.Normalized()
	return err
}

// Normalized uses the whitespace-trimmed uppercase string of the code to look up
// and return the standard ISO 3166-1 alpha-2 code.
// If not found, then AltCodes is used to look up alternative code and name mappings
// using the whitespace-trimmed uppercase code.
// If no mapping exists, then the original Code is returned unchanged together with an error.
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

// ParseCode parses and normalizes a country identifier given in any of
// the formats this package understands, returning the canonical ISO
// 3166-1 alpha-2 Code.
//
// It tries the cheapest and most common interpretations first and only
// falls back to the more expensive ones:
//
//  1. An input that is already a canonical ISO 3166-1 alpha-2 code
//     ("DE", "FR") is returned unchanged without any further work.
//  2. Otherwise the input is run through Normalized, which accepts case
//     and whitespace variants of alpha-2 codes plus the alternative
//     codes and German country names of AltCodes ("d", " AUT ", "SUI",
//     "Deutschland", "Österreich").
//  3. As a last resort the input is matched case-insensitively against
//     the English country names ("Germany", "united kingdom").
//
// ParseCode returns an error wrapping the original input if no ISO
// 3166-1 alpha-2 code can be derived. Like Normalized, the error
// preserves the unnormalized value for diagnostics.
func ParseCode(str string) (Code, error) {
	// 1. Short path: the input is already a canonical alpha-2 code.
	if _, ok := countryMap[Code(str)]; ok {
		return Code(str), nil
	}
	// 2. Codes and names handled by Normalized: case and whitespace
	//    variants, plus the alternative codes and German country names
	//    of AltCodes.
	c, err := Code(str).Normalized()
	if err == nil {
		return c, nil
	}
	// 3. English country name.
	if named, ok := codeForName(str); ok {
		return named, nil
	}
	// Reuse the input-preserving Code and error returned by Normalized.
	return c, err
}

// nameToCode is the lazily-built reverse index from a lower-cased English
// country name to its ISO 3166-1 alpha-2 Code. It is consulted only by
// ParseCode when the input is not a recognizable code, so it is built on
// first use rather than at package initialization.
var nameToCode = sync.OnceValue(func() map[string]Code {
	m := make(map[string]Code, len(countryMap))
	for code, name := range countryMap {
		m[strings.ToLower(name)] = code
	}
	return m
})

// codeForName resolves an English country name to its ISO 3166-1 alpha-2
// Code, matching case-insensitively. The bool result reports whether a
// match was found.
func codeForName(name string) (Code, bool) {
	key := strings.ToLower(strutil.TrimSpace(name))
	if key == "" {
		return Invalid, false
	}
	c, ok := nameToCode()[key]
	return c, ok
}

// IsEU indicates if a country is a member of the European Union.
func (c Code) IsEU() bool {
	norm, err := c.Normalized()
	if err != nil {
		return false
	}
	_, ok := euCountries[norm]
	return ok
}

// EnglishName returns the English name of the country.
// Returns an empty string if the code is invalid.
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
// Returns nil for SQL NULL if the Code is Invalid.
func (c Code) Value() (driver.Value, error) {
	if c == Invalid {
		return nil, nil
	}
	norm, _ := c.Normalized()
	return string(norm), nil
}

// MarshalJSON implements encoding/json.Marshaler.
// Returns the normalized code as a JSON string.
func (c Code) MarshalJSON() ([]byte, error) {
	norm, _ := c.Normalized()
	return json.Marshal(string(norm))
}

// JSONSchema returns the JSON schema definition for the Code type.
func (Code) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:   "ISO 3166-1 alpha 2 Country Code",
		Type:    "string",
		Pattern: "^[A-Z]{2}$",
	}
}

// ScanString tries to parse and assign the passed source string as value of the implementing type.
// If validate is true, the source string is checked for validity before assignment.
// If validate is false and the source string can still be assigned in some non-normalized way,
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
