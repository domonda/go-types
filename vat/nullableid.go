package vat

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/nullable"
)

// Ensure that NullableID implements StringGetter
var _ nullable.StringGetter = NullableID("")

// Null is an empty string and will be treatet as SQL NULL.
var Null NullableID

// NullableID is identical to ID, except that the Null value (empty string)
// is considered valid by the Valid() and Validate() methods.
type NullableID string

// NormalizedUnchecked returns a generic normalized version of ID without performing any format checks.
func (n NullableID) NormalizedUnchecked() NullableID {
	return NullableID(ID(n).NormalizedUnchecked())
}

// Normalized returns the id in normalized form,
// or an error if the VAT ID is not Null and not valid.
func (n NullableID) Normalized() (NullableID, error) {
	if n == Null {
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

// IsNull returns true if the NullableID is null.
// IsNull implements the nullable.Nullable interface.
func (n NullableID) IsNull() bool {
	return n == Null
}

// IsNotNull returns true if the NullableID is not null.
func (n NullableID) IsNotNull() bool {
	return n != Null
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
		panic("NULL vat.ID")
	}
	return ID(n)
}

// StringOr returns the NullableID as string
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

// ValidAndNotNull returns if this is a valid not Null VAIT ID.
func (n NullableID) ValidAndNotNull() bool {
	return n != Null && n.Valid()
}

// Validate returns an error if id is not a valid VAT ID or Null,
// ignoring normalization.
func (n NullableID) Validate() error {
	if n == Null {
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

// String returns the normalized ID if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (n NullableID) String() string {
	norm, err := n.Normalized()
	if err != nil {
		return string(n)
	}
	return string(norm)
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (n *NullableID) ScanString(source string) (normalized bool, err error) {
	newID, err := NullableID(source).Normalized()
	if err != nil {
		return false, err
	}
	*n = newID
	return string(newID) == source, nil
}

// Scan implements the database/sql.Scanner interface.
func (n *NullableID) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*n = NullableID(x).NormalizedUnchecked()
	case []byte:
		*n = NullableID(x).NormalizedUnchecked()
	case nil:
		*n = Null
	default:
		return fmt.Errorf("can't scan SQL value of type %T as vat.NullableID", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n NullableID) Value() (driver.Value, error) {
	normalized := n.NormalizedUnchecked()
	if normalized == Null {
		return nil, nil
	}
	return string(normalized), nil
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value for an empty (null) string.
func (n NullableID) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(string(n))
}
