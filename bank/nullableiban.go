package bank

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// IBANNull is an empty string and will be treated as SQL NULL.
const IBANNull NullableIBAN = ""

// NullableIBAN is a IBAN value which can hold an empty string ("") as the null value.
// A whitespace only NullableIBAN is also null, like for country.NullableCode.
//
// NullableIBAN reads and writes like IBAN, except that null is SQL NULL
// and JSON null. See "Reading and writing values" in the README of the module.
type NullableIBAN string

// Compile-time check that NullableIBAN implements nullable.NullSetable[IBAN]
var _ nullable.NullSetable[IBAN] = (*NullableIBAN)(nil)

// Compile-time check that NullableIBAN implements types.NormalizableValidator[NullableIBAN]
var _ types.NormalizableValidator[NullableIBAN] = NullableIBAN("")

// ScanString implements the strfmt.Scannable interface.
// The strings "NULL", "null" and "nil" set the NullableIBAN to null.
// Without validate it stores source like Scan, with validate it stores
// the normalized IBAN or returns the error of Normalized.
func (iban *NullableIBAN) ScanString(source string, validate bool) error {
	return internal.ScanStringNormalized(iban, source, validate, true, NullableIBAN.Normalized, nil)
}

// Valid returns true if iban is null or a valid International Bank Account Number
func (iban NullableIBAN) Valid() bool {
	return iban.Validate() == nil
}

// ValidAndNotNull returns true if iban is not null and a valid International Bank Account Number
func (iban NullableIBAN) ValidAndNotNull() bool {
	return iban.IsNotNull() && iban.Valid()
}

// Validate returns an error if this is not null and not a valid International Bank Account Number
func (iban NullableIBAN) Validate() error {
	_, err := iban.Normalized()
	return err
}

// ValidAndNormalized returns true if the IBAN is valid and already in normalized form,
// or if it is the empty IBANNull. A whitespace-only value is null and valid,
// but not normalized.
func (iban NullableIBAN) ValidAndNormalized() bool {
	norm, err := iban.Normalized()
	return err == nil && iban == norm
}

// CountryCode returns the country code of the normalized IBAN
// like IBAN.CountryCode, or country.Invalid ("")
// if the IBAN is null or not valid.
func (iban NullableIBAN) CountryCode() country.Code {
	if iban.IsNull() {
		return country.Invalid
	}
	return IBAN(iban).CountryCode()
}

// Normalized returns the iban in normalized form,
// or an error if the format can't be detected.
// Returns IBANNull for a null (empty or whitespace only) iban, see IsNull.
// Returns the NullableIBAN unchanged in case of an error.
func (iban NullableIBAN) Normalized() (NullableIBAN, error) {
	if iban.IsNull() {
		return IBANNull, nil
	}
	normalized, err := IBAN(iban).Normalized()
	if err != nil {
		return iban, err
	}
	return NullableIBAN(normalized), nil
}

// NormalizedOrNull returns the NullableIBAN in normalized form,
// or IBANNull if the value is null or normalization fails.
func (iban NullableIBAN) NormalizedOrNull() NullableIBAN {
	normalized, err := iban.Normalized()
	if err != nil {
		return IBANNull
	}
	return normalized
}

// NormalizedWithSpaces returns the iban in normalized form with spaces every 4 characters,
// or an error if the format can't be detected.
// Returns IBANNull for a null (empty or whitespace only) iban, see IsNull.
// Returns the NullableIBAN unchanged in case of an error.
func (iban NullableIBAN) NormalizedWithSpaces() (NullableIBAN, error) {
	if iban.IsNull() {
		return IBANNull, nil
	}
	normalized, err := IBAN(iban).NormalizedWithSpaces()
	if err != nil {
		return iban, err
	}
	return NullableIBAN(normalized), nil
}

// Scan implements the database/sql.Scanner interface.
// A valid IBAN is stored normalized, SQL NULL and a blank string as IBANNull,
// and a IBAN that can't be normalized unchanged without an error,
// so that one invalid value doesn't fail a whole query; Validate reports it.
func (iban *NullableIBAN) Scan(value any) error {
	return internal.ScanNormalized(iban, value, NullableIBAN.Normalized, nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing the JSON null
// value if the NullableIBAN is null, else the normalized IBAN,
// or the IBAN unchanged if it can't be normalized.
func (iban NullableIBAN) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(iban, true, NullableIBAN.Normalized, nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null sets the
// NullableIBAN to null, so that decoding null into a field that holds
// a value resets it, and any other JSON value is an error.
func (iban *NullableIBAN) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(iban, data, true, NullableIBAN.Normalized, nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// but with "" for null, so that encoding/xml and strfmt write the
// normalized IBAN. encoding/json writes map keys of string types
// as they are and doesn't call MarshalText for them.
func (iban NullableIBAN) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(iban, NullableIBAN.Normalized, nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (iban *NullableIBAN) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(iban, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (iban *NullableIBAN) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(iban, d, start, NullableIBAN.Normalized, nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (iban *NullableIBAN) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(iban, attr, NullableIBAN.Normalized, nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the normalized IBAN, SQL NULL if the NullableIBAN is null,
// or the error of Normalized for a IBAN that can't be normalized.
func (iban NullableIBAN) Value() (driver.Value, error) {
	return internal.ValueNormalized(iban, NullableIBAN.Normalized, nil)
}

// Set sets an IBAN for this NullableIBAN
func (iban *NullableIBAN) Set(id IBAN) {
	*iban = NullableIBAN(id)
}

// SetNull sets the NullableIBAN to null
func (iban *NullableIBAN) SetNull() {
	*iban = IBANNull
}

// Get returns the non nullable IBAN value
// or panics if the NullableIBAN is null.
// Note: check with IsNull before using Get!
func (iban NullableIBAN) Get() IBAN {
	if iban.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", iban))
	}
	return IBAN(iban)
}

// GetOr returns the non nullable IBAN value
// or the passed defaultIBAN if the NullableIBAN is null.
func (iban NullableIBAN) GetOr(defaultIBAN IBAN) IBAN {
	if iban.IsNull() {
		return defaultIBAN
	}
	return IBAN(iban)
}

// StringOr returns the NullableIBAN as stored,
// or the passed defaultString if the NullableIBAN is null.
func (iban NullableIBAN) StringOr(defaultString string) string {
	if iban.IsNull() {
		return defaultString
	}
	return string(iban)
}

// IsNull returns true if the NullableIBAN is null,
// which includes a string of only the runes strutil.TrimSpace removes:
// whitespace, control and other non-printable runes, and invalid UTF-8,
// like for Normalized.
// IsNull implements the nullable.Nullable interface.
func (iban NullableIBAN) IsNull() bool {
	return strutil.TrimSpace(iban) == IBANNull
}

// IsNotNull returns true if the NullableIBAN is not null,
// see IsNull.
func (iban NullableIBAN) IsNotNull() bool {
	return !iban.IsNull()
}

// JSONSchema returns the JSON schema definition for the NullableIBAN type,
// accepting either an IBAN string matching IBANRegex or JSON null.
func (NullableIBAN) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable IBAN",
		OneOf: []*jsonschema.Schema{
			{
				Type:    "string",
				Pattern: IBANRegex,
			},
			{Type: "null"},
		},
		Default: IBANNull,
	}
}
