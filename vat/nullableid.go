package vat

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"

	"github.com/invopop/jsonschema"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// Null is an empty string and will be treated as SQL NULL.
const Null NullableID = ""

// NullableID is identical to ID, except that the Null value (empty string)
// is considered valid by the Valid() and Validate() methods.
// A whitespace only NullableID is also null, like for country.NullableCode.
//
// NullableID reads and writes like ID, except that null is SQL NULL and
// JSON null. See "Reading and writing values" in the README of the module.
type NullableID string

// Compile-time check that NullableID implements nullable.NullSetable[ID]
var _ nullable.NullSetable[ID] = (*NullableID)(nil)

// Compile-time check that NullableID implements types.NormalizableValidator[NullableID]
var _ types.NormalizableValidator[NullableID] = NullableID("")

// Normalized returns the id in normalized form,
// or an error if the VAT ID is not null and not valid.
// Returns Null for a null (empty or whitespace only) id, see IsNull.
func (n NullableID) Normalized() (NullableID, error) {
	if n.IsNull() {
		return Null, nil
	}
	id, err := ID(n).Normalized()
	return NullableID(id), err
}

// NormalizedOrNull returns n in normalized form
// or Null if id is not valid.
func (n NullableID) NormalizedOrNull() NullableID {
	normalized, err := n.Normalized()
	if err != nil {
		return Null
	}
	return normalized
}

// NormalizedNotNull returns the id in normalized form,
// or an error if the VAT ID is not valid or Null.
func (n NullableID) NormalizedNotNull() (ID, error) {
	return ID(n).Normalized()
}

// IsNull returns true if the NullableID is null,
// which includes a string of only the runes strutil.TrimSpace removes:
// whitespace, control and other non-printable runes, and invalid UTF-8,
// like for Normalized.
// IsNull implements the nullable.Nullable interface.
func (n NullableID) IsNull() bool {
	return strutil.TrimSpace(n) == Null
}

// IsNotNull returns true if the NullableID is not null,
// see IsNull.
func (n NullableID) IsNotNull() bool {
	return !n.IsNull()
}

// Set sets an ID for this NullableID
func (n *NullableID) Set(id ID) {
	*n = NullableID(id)
}

// SetNull sets the NullableID to null
func (n *NullableID) SetNull() {
	*n = Null
}

// Get returns the non nullable ID value
// or panics if the NullableID is null.
// Note: check with IsNull before using Get!
func (n NullableID) Get() ID {
	if n.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", n))
	}
	return ID(n)
}

// GetOr returns the non nullable ID value
// or the passed defaultID if the NullableID is null.
func (n NullableID) GetOr(defaultID ID) ID {
	if n.IsNull() {
		return defaultID
	}
	return ID(n)
}

// StringOr returns the NullableID as stored,
// or the passed nullString if the NullableID is null.
func (n NullableID) StringOr(nullString string) string {
	if n.IsNull() {
		return nullString
	}
	return string(n)
}

// Valid returns if id is a valid VAT ID or Null,
// ignoring normalization.
func (n NullableID) Valid() bool {
	return n.Validate() == nil
}

// ValidAndNormalized returns if id is Null or a valid and normalized VAT ID.
func (n NullableID) ValidAndNormalized() bool {
	norm, err := n.Normalized()
	return err == nil && n == norm
}

// ValidAndNotNull returns if this is a valid not Null VAT ID.
func (n NullableID) ValidAndNotNull() bool {
	return n.IsNotNull() && n.Valid()
}

// Validate returns an error if id is not a valid VAT ID or null,
// ignoring normalization.
func (n NullableID) Validate() error {
	if n.IsNull() {
		return nil
	}
	return ID(n).Validate()
}

// ValidateIsNormalized returns an error if id is not Null or a valid and normalized VAT ID.
func (n NullableID) ValidateIsNormalized() error {
	norm, err := n.Normalized()
	if err != nil {
		return err
	}
	if n != norm {
		return fmt.Errorf("VAT ID is valid but not normalized: %q", string(n))
	}
	return nil
}

// ValidateIsNormalizedAndNotNull returns an error if id is not a valid and normalized VAT ID.
func (n NullableID) ValidateIsNormalizedAndNotNull() error {
	return ID(n).ValidateIsNormalized()
}

// CountryCode returns the country.NullableCode of the VAT ID,
// or ccountry.Null if the id is null or not valid.
// For a VAT Mini One Stop Shop (MOSS) ID that begins with "EU"
// the EU's capital Brussels' country Belgum's
// code country.BE will be returned.
// See also NullableID.IsMOSS.
func (n NullableID) CountryCode() country.NullableCode {
	if n.IsNull() {
		return country.Null
	}
	return country.NullableCode(ID(n).CountryCode())
}

// IsMOSS returns true if the ID follows the
// VAT Mini One Stop Shop (MOSS) schema beginning with "EU".
func (n NullableID) IsMOSS() bool {
	if n.IsNull() {
		return false
	}
	return ID(n).IsMOSS()
}

// Number returns the number part after the country code of the VAT ID,
// or and empty string if the id is not valid.
func (n NullableID) Number() string {
	return ID(n).Number()
}

// ScanString implements the strfmt.Scannable interface.
// The strings "NULL", "null" and "nil" set the NullableID to null.
// Without validate it stores source like Scan, with validate it stores
// the normalized ID or returns the error of Normalized.
func (n *NullableID) ScanString(source string, validate bool) error {
	return internal.ScanStringNormalized(n, source, validate, true, NullableID.Normalized, nil)
}

// Scan implements the database/sql.Scanner interface.
// A valid ID is stored normalized, SQL NULL and a blank string as Null,
// and an ID that can't be normalized unchanged without an error,
// so that one invalid value doesn't fail a whole query; Validate reports it.
func (n *NullableID) Scan(value any) error {
	return internal.ScanNormalized(n, value, NullableID.Normalized, nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the normalized ID, SQL NULL if the NullableID is null,
// or the error of Normalized wrapping ErrInvalidID
// for an ID that can't be normalized.
func (n NullableID) Value() (driver.Value, error) {
	return internal.ValueNormalized(n, NullableID.Normalized, nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing the JSON null
// value if the NullableID is null, else the normalized ID,
// or the ID unchanged if it can't be normalized.
func (n NullableID) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(n, true, NullableID.Normalized, nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null sets the
// NullableID to null, so that decoding null into a field that holds
// a value resets it, and any other JSON value is an error.
func (n *NullableID) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(n, data, true, NullableID.Normalized, nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// but with "" for null, so that encoding/xml and strfmt write the
// normalized ID. encoding/json writes map keys of string types
// as they are and doesn't call MarshalText for them.
func (n NullableID) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(n, NullableID.Normalized, nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (n *NullableID) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(n, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (n *NullableID) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(n, d, start, NullableID.Normalized, nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (n *NullableID) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(n, attr, NullableID.Normalized, nil)
}

// JSONSchema returns the JSON schema for NullableID, describing it with the
// title "Nullable Value Added Tax ID" as a oneOf that accepts either a string
// with the package-defined minimum length or a JSON null value.
func (NullableID) JSONSchema() *jsonschema.Schema {
	// No MaxLength like for ID.JSONSchema
	minLength := uint64(IDMinLength)
	return &jsonschema.Schema{
		Title: "Nullable Value Added Tax ID",
		OneOf: []*jsonschema.Schema{
			{
				Type:      "string",
				MinLength: &minLength,
			},
			{Type: "null"},
		},
		Default: Null,
	}
}
