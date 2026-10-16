package vat

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"
	"strings"
	"unicode"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/strutil"
)

// Compile-time check that ID implements types.NormalizableValidator[ID]
var _ types.NormalizableValidator[ID] = ID("")

// MOSSSchemaVATCountryCode or the VAT Mini One Stop Shop (MOSS) is an optional scheme that allows you
// to account for VAT - normally due in multiple EU countries – in just one EU country. Check out:
// https://europa.eu/youreurope/business/taxation/vat/vat-digital-services-moss-scheme/index_en.htm
const MOSSSchemaVATCountryCode = "EU"

// NorthernIrelandVATCountryCode is the post-Brexit VAT prefix for Northern Ireland businesses
// that remain inside the EU VAT regime for goods. "XI" is not an ISO 3166-1 country code, so
// it is allow-listed alongside MOSSSchemaVATCountryCode.
const NorthernIrelandVATCountryCode = "XI"

// ErrInvalidID is returned (wrapped) by Normalized and Validate when a VAT ID
// string does not conform to the expected format or fails the country-specific
// check-sum verification.
const ErrInvalidID errs.Sentinel = "invalid VAT ID"

// ID is a european VAT ID.
// See NullableID for a nullable version of this type.
//
// Scan, UnmarshalJSON, UnmarshalXML and ScanString without validate trim
// surrounding whitespace and uppercase the letters. Punctuation stays.
// A blank ID is "". An invalid ID is kept after that prettification,
// without an error. Value, MarshalJSON and MarshalText write the same
// text, and Value writes SQL NULL for a blank ID. They do not reject an
// invalid ID. ScanString with validate stores the normalized ID or
// returns the error of Normalized. fmt prints the ID as stored. ID has
// no UnmarshalText, so encoding/json reads and writes the keys of a
// map[ID]V as given. See "Reading and writing values" in the README
// of the module.
type ID string

// NormalizeVATID returns str as normalized VAT ID or an error.
//
// Returns a wrapped ErrInvalidID error if the VAT ID is not valid.
func NormalizeVATID(str string) (ID, error) {
	return ID(str).Normalized()
}

// StringIsVATID returns if a string can be parsed as VATID.
func StringIsVATID(str string) bool {
	return ID(str).Valid()
}

// BytesAreVATID returns if a byte string is a valid VAT ID
func BytesAreVATID(str []byte) bool {
	return ID(str).Valid()
}

func isVATIDSplitRune(r rune) bool {
	return unicode.IsSpace(r) || r == ':'
}

func isVATIDTrimRune(r rune) bool {
	return unicode.IsPunct(r)
}

// Normalized returns the id in normalized form,
// or a wrapped ErrInvalidID error if the VAT ID is not valid.
func (id ID) Normalized() (ID, error) {
	normalized := ID(strings.ToUpper(strutil.RemoveRunesString(string(id), strutil.IsSpace, unicode.IsPunct)))

	// Check length
	if len(normalized) < IDMinLength {
		return id, fmt.Errorf("%w: %q is too short", ErrInvalidID, string(id))
	}
	if len(normalized) > IDMaxLength {
		return id, fmt.Errorf("%w: %q is too long", ErrInvalidID, string(id))
	}

	// Check country code
	countryCode := country.Code(normalized[:2])
	if countryCode != MOSSSchemaVATCountryCode && countryCode != NorthernIrelandVATCountryCode && !countryCode.Valid() {
		return id, fmt.Errorf("%w: %q has an invalid country code: %q", ErrInvalidID, string(id), string(countryCode))
	}

	// Check format with country specific regex
	regex, ok := idRegex[countryCode]
	if !ok {
		return id, fmt.Errorf("%w: %q has an unsupported country code: %q", ErrInvalidID, string(id), string(countryCode))
	}
	if !regex.MatchString(string(normalized)) {
		return id, fmt.Errorf("%w: %q has an invalid format", ErrInvalidID, string(id))
	}

	// Test checkFunc-sum if a function is available for the country
	checkFunc, ok := checkSumFuncs[countryCode]
	if ok && !checkFunc(id, normalized) {
		return id, fmt.Errorf("%w: %q has an invalid check-sum", ErrInvalidID, string(id))
	}

	return normalized, nil
}

// NormalizedOrNull returns the id in normalized form
// or Null if the VAT ID is not valid.
func (id ID) NormalizedOrNull() NullableID {
	normalized, err := id.Normalized()
	if err != nil {
		return Null
	}
	return NullableID(normalized)
}

// Valid returns if id is a valid VAT ID,
// ignoring normalization.
func (id ID) Valid() bool {
	_, err := id.Normalized()
	return err == nil
}

// ValidAndNormalized returns if id is a valid and normalized VAT ID.
func (id ID) ValidAndNormalized() bool {
	norm, err := id.Normalized()
	return err == nil && id == norm
}

// Validate returns an error if id is not a valid VAT ID,
// ignoring normalization.
//
// Returns a wrapped ErrInvalidID error if the VAT ID is not valid.
func (id ID) Validate() error {
	_, err := id.Normalized()
	return err
}

// ValidateIsNormalized returns an error if id is not a valid and normalized VAT ID.
//
// Will return ErrInvalidID if the id is not valid,
// but another error if the id is valid but not normalized.
func (id ID) ValidateIsNormalized() error {
	norm, err := id.Normalized()
	if err != nil {
		return err
	}
	if id != norm {
		return fmt.Errorf("VAT ID is valid but not normalized: %q", string(id))
	}
	return nil
}

// Nullable returns the id as NullableID
func (id ID) Nullable() NullableID {
	return NullableID(id)
}

// CountryCode returns the country.Code of the VAT ID,
// or country.Invalid if the id is not valid.
// For a VAT Mini One Stop Shop (MOSS) ID that begins with "EU"
// the EU's capital Brussels' country Belgum's
// code country.BE will be returned.
// For a Northern Ireland VAT ID that begins with "XI" the UK code country.GB
// will be returned, since "XI" is not a real ISO 3166-1 country code.
// See also ID.IsMOSS.
func (id ID) CountryCode() country.Code {
	norm, err := id.Normalized()
	if err != nil {
		return country.Invalid
	}
	code := country.Code(norm[:2])
	if code == MOSSSchemaVATCountryCode {
		// MOSS VAT begins with "EU" - Europe is not a country
		return country.BE
	}
	if code == NorthernIrelandVATCountryCode {
		return country.GB
	}
	return code
}

// IsMOSS returns true if the ID follows the
// VAT Mini One Stop Shop (MOSS) schema beginning with "EU".
func (id ID) IsMOSS() bool {
	norm, err := id.Normalized()
	if err != nil {
		return false
	}
	return norm[:2] == MOSSSchemaVATCountryCode
}

// Number returns the number part after the country code of the VAT ID,
// or and empty string if the id is not valid.
func (id ID) Number() string {
	norm, err := id.Normalized()
	if err != nil {
		return ""
	}
	return string(norm[2:])
}

// ScanString implements the strfmt.Scannable interface.
// Without validate it stores source like Scan, with validate it stores
// the normalized ID or returns the error of Normalized,
// also for a blank source.
func (id *ID) ScanString(source string, validate bool) error {
	if validate {
		return internal.ScanStringNormalized(id, source, true, false, ID.Normalized, nil)
	}
	return internal.ScanStringNormalized(id, source, false, false, internal.TrimUpper[ID], nil)
}

// Scan implements the database/sql.Scanner interface.
// Surrounding whitespace is trimmed and letters are uppercased.
// SQL NULL and a blank string are stored as "". An invalid ID is
// kept after that prettification, without an error.
func (id *ID) Scan(value any) error {
	return internal.ScanNormalized(id, value, internal.TrimUpper[ID], nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the trimmed and uppercased ID, or SQL NULL for a blank one.
// An invalid ID is written after that prettification, not rejected.
func (id ID) Value() (driver.Value, error) {
	return internal.ValueNormalized(id, internal.TrimUpper[ID], nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing the
// trimmed and uppercased ID, or "" for a blank ID.
func (id ID) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(id, false, internal.TrimUpper[ID], nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null leaves the ID
// unchanged like encoding/json does for a string type, and any other
// JSON value is an error.
func (id *ID) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(id, data, false, internal.TrimUpper[ID], nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// so that encoding/xml and strfmt write the trimmed and uppercased ID.
// encoding/json writes map keys of string types as they are
// and doesn't call MarshalText for them.
func (id ID) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(id, internal.TrimUpper[ID], nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (id *ID) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(id, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (id *ID) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(id, d, start, internal.TrimUpper[ID], nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (id *ID) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(id, attr, internal.TrimUpper[ID], nil)
}

// JSONSchema returns the JSON schema for ID, describing it as a string with
// the title "Value Added Tax ID" and the package-defined minimum length.
// It has no maximum length, see IDMaxLength.
func (ID) JSONSchema() *jsonschema.Schema {
	// No MaxLength, because IDMaxLength limits the normalized ID while
	// JSON input can be formatted with whitespace and punctuation
	minLength := uint64(IDMinLength)
	return &jsonschema.Schema{
		Title:     "Value Added Tax ID",
		Type:      "string",
		MinLength: &minLength,
	}
}
