package country

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// Null represents a null country code.
const Null NullableCode = ""

// Compile-time check that NullableCode implements nullable.NullSetable[Code]
var _ nullable.NullSetable[Code] = (*NullableCode)(nil)

// Compile-time check that NullableCode implements types.NormalizableValidator[NullableCode]
var _ types.NormalizableValidator[NullableCode] = NullableCode("")

// NullableCode represents a country code according to ISO 3166-1 alpha-2 standard
// where an empty or whitespace only string represents NULL.
// Null.Valid() or NullableCode("").Valid() will return true.
//
// NullableCode reads and writes like Code, except that null is
// SQL NULL and JSON null. See "Reading and writing values"
// in the README of the module.
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

// NormalizedWithAltCodes returns the same result as Normalized,
// which already maps the AltCodes to ISO 3166-1 alpha-2 codes.
// It is kept for compatibility.
func (n NullableCode) NormalizedWithAltCodes() (NullableCode, error) {
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

// IsNull returns true if the NullableCode is null,
// which includes a string of only the runes strutil.TrimSpace removes:
// whitespace, control and other non-printable runes, and invalid UTF-8,
// like for Normalized.
// IsNull implements the nullable.Nullable interface.
func (n NullableCode) IsNull() bool {
	return strutil.TrimSpace(n) == Null
}

// IsNotNull returns true if the NullableCode is not null,
// see IsNull.
func (n NullableCode) IsNotNull() bool {
	return !n.IsNull()
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

// StringOr returns the NullableCode as stored,
// or the passed defaultString if the NullableCode is null.
func (n NullableCode) StringOr(defaultString string) string {
	if n.IsNull() {
		return defaultString
	}
	return string(n)
}

// Scan implements the database/sql.Scanner interface.
// A valid code is stored normalized, SQL NULL and a blank string as Null,
// and a code that can't be normalized unchanged without an error,
// so that one invalid value doesn't fail a whole query; Validate reports it.
func (n *NullableCode) Scan(value any) error {
	return internal.ScanNormalized(n, value, NullableCode.Normalized, nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the normalized code, SQL NULL if the NullableCode is null,
// or the error of Normalized for a code that can't be normalized.
func (n NullableCode) Value() (driver.Value, error) {
	return internal.ValueNormalized(n, NullableCode.Normalized, nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing the JSON null
// value if the NullableCode is null, else the normalized code,
// or the code unchanged if it can't be normalized.
func (n NullableCode) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(n, true, NullableCode.Normalized, nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null sets the
// NullableCode to null, so that decoding null into a field that holds
// a value resets it, and any other JSON value is an error.
func (n *NullableCode) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(n, data, true, NullableCode.Normalized, nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// but with "" for null, so that encoding/xml and strfmt write the
// normalized code. encoding/json writes map keys of string types
// as they are and doesn't call MarshalText for them.
func (n NullableCode) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(n, NullableCode.Normalized, nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (n *NullableCode) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(n, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (n *NullableCode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(n, d, start, NullableCode.Normalized, nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (n *NullableCode) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(n, attr, NullableCode.Normalized, nil)
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

// ScanString implements the strfmt.Scannable interface.
// The strings "NULL", "null" and "nil" set the NullableCode to null.
// Without validate it stores source like Scan, with validate it stores
// the normalized code or returns the error of Normalized.
func (n *NullableCode) ScanString(source string, validate bool) error {
	return internal.ScanStringNormalized(n, source, validate, true, NullableCode.Normalized, nil)
}
