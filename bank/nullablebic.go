package bank

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"

	"github.com/invopop/jsonschema"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// BICNull is an empty string and will be treated as SQL NULL.
const BICNull NullableBIC = ""

// NullableBIC is a BIC value which can hold an empty string ("") as the null value.
// A whitespace only NullableBIC is also null, like for country.NullableCode.
//
// NullableBIC reads and writes like BIC, except that null is SQL NULL
// and JSON null. See "Reading and writing values" in the README of the module.
type NullableBIC string

// Compile-time check that NullableBIC implements nullable.NullSetable[BIC]
var _ nullable.NullSetable[BIC] = (*NullableBIC)(nil)

// Compile-time check that NullableBIC implements types.NormalizableValidator[NullableBIC]
var _ types.NormalizableValidator[NullableBIC] = NullableBIC("")

// ScanString implements the strfmt.Scannable interface.
// The strings "NULL", "null" and "nil" set the NullableBIC to null.
// Without validate it stores source like Scan, with validate it stores
// the normalized BIC or returns the error of Normalized.
func (bic *NullableBIC) ScanString(source string, validate bool) error {
	return internal.ScanStringNormalized(bic, source, validate, true, NullableBIC.Normalized, nil)
}

// Valid returns true if bic is null or a valid SWIFT Business Identifier Code
func (bic NullableBIC) Valid() bool {
	return bic.Validate() == nil
}

// ValidAndNotNull returns true if bic is not null and a valid SWIFT Business Identifier Code
func (bic NullableBIC) ValidAndNotNull() bool {
	return bic.IsNotNull() && bic.Valid()
}

// Validate returns an error if this is not null
// and not a valid SWIFT Business Identifier Code
func (bic NullableBIC) Validate() error {
	if bic.IsNull() {
		return nil
	}
	return BIC(bic).Validate()
}

// ValidAndNormalized returns true if the BIC is valid and already normalized,
// or if it is the empty BICNull. A whitespace-only value is null and valid,
// but not normalized.
func (bic NullableBIC) ValidAndNormalized() bool {
	norm, err := bic.Normalized()
	return err == nil && bic == norm
}

// Normalized returns the BIC normalized to 11 characters,
// or BICNull if the value is null (empty or whitespace only), see IsNull.
// Returns an error if the BIC is non-null but invalid.
func (bic NullableBIC) Normalized() (NullableBIC, error) {
	if bic.IsNull() {
		return BICNull, nil
	}
	normalized, err := BIC(bic).Normalized()
	if err != nil {
		return bic, err
	}
	return NullableBIC(normalized), nil
}

// NormalizedOrNull returns the NullableBIC in normalized form,
// or BICNull if the value is null or normalization fails.
func (bic NullableBIC) NormalizedOrNull() NullableBIC {
	normalized, err := bic.Normalized()
	if err != nil {
		return BICNull
	}
	return normalized
}

// Scan implements the database/sql.Scanner interface.
// A valid BIC is stored normalized, SQL NULL and a blank string as BICNull,
// and a BIC that can't be normalized unchanged without an error,
// so that one invalid value doesn't fail a whole query; Validate reports it.
func (bic *NullableBIC) Scan(value any) error {
	return internal.ScanNormalized(bic, value, NullableBIC.Normalized, nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing the JSON null
// value if the NullableBIC is null, else the normalized BIC,
// or the BIC unchanged if it can't be normalized.
func (bic NullableBIC) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(bic, true, NullableBIC.Normalized, nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null sets the
// NullableBIC to null, so that decoding null into a field that holds
// a value resets it, and any other JSON value is an error.
func (bic *NullableBIC) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(bic, data, true, NullableBIC.Normalized, nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// but with "" for null, so that encoding/xml and strfmt write the
// normalized BIC. encoding/json writes map keys of string types
// as they are and doesn't call MarshalText for them.
func (bic NullableBIC) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(bic, NullableBIC.Normalized, nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (bic *NullableBIC) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(bic, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (bic *NullableBIC) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(bic, d, start, NullableBIC.Normalized, nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (bic *NullableBIC) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(bic, attr, NullableBIC.Normalized, nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the normalized BIC, SQL NULL if the NullableBIC is null,
// or the error of Normalized for a BIC that can't be normalized.
func (bic NullableBIC) Value() (driver.Value, error) {
	return internal.ValueNormalized(bic, NullableBIC.Normalized, nil)
}

// Set sets an BIC for this NullableBIC
func (bic *NullableBIC) Set(id BIC) {
	*bic = NullableBIC(id)
}

// SetNull sets the NullableBIC to null
func (bic *NullableBIC) SetNull() {
	*bic = BICNull
}

// Get returns the non nullable BIC value
// or panics if the NullableBIC is null.
// Note: check with IsNull before using Get!
func (bic NullableBIC) Get() BIC {
	if bic.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", bic))
	}
	return BIC(bic)
}

// GetOr returns the non nullable BIC value
// or the passed defaultBIC if the NullableBIC is null.
func (bic NullableBIC) GetOr(defaultBIC BIC) BIC {
	if bic.IsNull() {
		return defaultBIC
	}
	return BIC(bic)
}

// StringOr returns the NullableBIC as stored,
// or the passed defaultString if the NullableBIC is null.
func (bic NullableBIC) StringOr(defaultString string) string {
	if bic.IsNull() {
		return defaultString
	}
	return string(bic)
}

// IsNull returns true if the NullableBIC is null,
// which includes a string of only the runes strutil.TrimSpace removes:
// whitespace, control and other non-printable runes, and invalid UTF-8,
// like for Normalized.
// IsNull implements the nullable.Nullable interface.
func (bic NullableBIC) IsNull() bool {
	return strutil.TrimSpace(bic) == BICNull
}

// IsNotNull returns true if the NullableBIC is not null,
// see IsNull.
func (bic NullableBIC) IsNotNull() bool {
	return !bic.IsNull()
}

// JSONSchema returns the JSON schema definition for the NullableBIC type,
// accepting either a BIC string matching BICRegex or JSON null.
func (NullableBIC) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable BIC/SWIFT-Code",
		OneOf: []*jsonschema.Schema{
			{
				Type:    "string",
				Pattern: BICRegex,
			},
			{Type: "null"},
		},
		Default: BICNull,
	}
}
