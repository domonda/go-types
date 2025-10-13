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
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// Null represents a null country code.
const Null NullableCode = ""

// Compile-time check that NullableCode implements nullable.NullSetable[Code]
var _ nullable.NullSetable[Code] = (*NullableCode)(nil)

// NullableCode represents a country code according to ISO 3166-1 alpha-2 standard
// where an empty string represents NULL.
// NullableCode implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and treats an empty NullableCode string as SQL NULL.
// Null.Valid() or NullableCode("").Valid() will return true.
type NullableCode string

// Valid returns true if the NullableCode is valid (including null values).
func (n NullableCode) Valid() bool {
	_, err := n.Normalized()
	return err == nil
}

// ValidAndNotNull returns true if the NullableCode is valid and not null.
func (n NullableCode) ValidAndNotNull() bool {
	return Code(n).Valid()
}

// Validate returns an error if the NullableCode is not valid.
func (n NullableCode) Validate() error {
	_, err := n.Normalized()
	return err
}

// ValidAndNormalized returns true if the NullableCode is valid and already normalized.
func (n NullableCode) ValidAndNormalized() bool {
	norm, err := n.Normalized()
	return err == nil && n == norm
}

// Normalized uses the whitespace-trimmed uppercase string of the code to look up
// and return the standard ISO 3166-1 alpha-2 code, or return Null and no error
// in case of an empty string representing a null value.
// If not found, then AltCodes is used to look up alternative code and name mappings
// using the whitespace-trimmed uppercase code.
// If no mapping exists, then the original Code is returned unchanged together with an error.
func (n NullableCode) Normalized() (NullableCode, error) {
	norm := strings.ToUpper(strutil.TrimSpace(string(n)))
	if norm == "" {
		return Null, nil
	}
	if _, ok := countryMap[Code(norm)]; ok {
		return NullableCode(norm), nil
	}
	if norm, ok := AltCodes[string(norm)]; ok {
		return NullableCode(norm), nil
	}
	return n, fmt.Errorf("invalid country.NullableCode: '%s'", string(n))
}

// NormalizedWithAltCodes uses AltCodes to map to ISO 3166-1 alpha-2 codes
// or returns the result of Normalized() if no mapping exists.
func (n NullableCode) NormalizedWithAltCodes() (NullableCode, error) {
	if norm, ok := AltCodes[strings.ToUpper(strutil.TrimSpace(string(n)))]; ok {
		return norm.Nullable(), nil
	}
	return n.Normalized()
}

// NormalizedOrNull returns the normalized NullableCode or Null if normalization fails.
func (n NullableCode) NormalizedOrNull() NullableCode {
	norm, err := n.Normalized()
	if err != nil {
		return Null
	}
	return norm
}

// IsEU indicates if a country is a member of the European Union.
func (n NullableCode) IsEU() bool {
	return Code(n).IsEU()
}

// EnglishName returns the English name of the country.
func (n NullableCode) EnglishName() string {
	return Code(n).EnglishName()
}

// IsNull returns true if the NullableCode is null.
// IsNull implements the nullable.Nullable interface.
func (n NullableCode) IsNull() bool {
	return n == Null
}

// IsNotNull returns true if the NullableCode is not null.
func (n NullableCode) IsNotNull() bool {
	return n != Null
}

// Set sets a Code for this NullableCode.
func (n *NullableCode) Set(code Code) {
	*n = NullableCode(code)
}

// SetNull sets the NullableCode to null.
func (n *NullableCode) SetNull() {
	*n = Null
}

// Get returns the non-nullable Code value or panics if the NullableCode is null.
// Note: check with IsNull before using Get!
func (n NullableCode) Get() Code {
	if n.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", n))
	}
	return Code(n)
}

// GetOr returns the non-nullable Code value or the passed defaultCode if the NullableCode is null.
func (n NullableCode) GetOr(defaultCode Code) Code {
	if n.IsNull() {
		return defaultCode
	}
	return Code(n)
}

// StringOr returns the NullableCode as string or the passed defaultString if the NullableCode is null.
func (n NullableCode) StringOr(defaultString string) string {
	if n.IsNull() {
		return defaultString
	}
	return string(n)
}

// String returns the normalized code if possible, else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (n NullableCode) String() string {
	if n.IsNull() {
		return ""
	}
	norm, err := n.Normalized()
	if err != nil {
		return string(n)
	}
	return string(norm)
}

// Scan implements the database/sql.Scanner interface.
func (n *NullableCode) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*n = NullableCode(x)
	case []byte:
		*n = NullableCode(x)
	case nil:
		*n = Null
	default:
		return fmt.Errorf("can't scan SQL value of type %T as country.NullableCode", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
// Returns nil for SQL NULL if the NullableCode is Null.
func (n NullableCode) Value() (driver.Value, error) {
	if n == Null {
		return nil, nil
	}
	return Code(n).Value()
}

// MarshalJSON implements encoding/json.Marshaler by returning the JSON null value
// for an empty (null) string.
func (n NullableCode) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return Code(n).MarshalJSON()
}

// JSONSchema returns the JSON schema definition for the NullableCode type.
func (NullableCode) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable ISO 3166-1 alpha 2 Country Code",
		OneOf: []*jsonschema.Schema{
			{
				Type:    "string",
				Pattern: "^[A-Z]{2}$",
			},
			{Type: "null"},
		},
		Default: Null,
	}
}

// ScanString tries to parse and assign the passed source string as value of the implementing type.
// If validate is true, the source string is checked for validity before assignment.
// If validate is false and the source string can still be assigned in some non-normalized way,
// it will be assigned without returning an error.
func (n *NullableCode) ScanString(source string, validate bool) error {
	switch source {
	case "", "NULL", "null", "nil":
		n.SetNull()
		return nil
	}
	code, err := NullableCode(source).Normalized()
	if err != nil {
		if validate {
			return err
		}
		code = NullableCode(source)
	}
	*n = code
	return nil
}
