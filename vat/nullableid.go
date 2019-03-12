package vat

import (
	"database/sql/driver"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/country"
)

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
		return errors.Errorf("VAT ID is valid but not normalized: '%s'", n)
	}
	return nil
}

// ValidateIsNormalizedAndNotNull returns an error if id is not a valid and normalized VAT ID.
func (n NullableID) ValidateIsNormalizedAndNotNull() error {
	return ID(n).ValidateIsNormalized()
}

// ID returns the NullableID as ID without any validation.
func (n NullableID) ID() ID {
	return ID(n)
}

// CountryCode returns the country.Code of the VAT ID,
// or country.Null if the id is not valid.
func (n NullableID) CountryCode() country.Code {
	return ID(n).CountryCode()
}

// Number returns the number part after the country code of the VAT ID,
// or and empty string if the id is not valid.
func (n NullableID) Number() string {
	return ID(n).Number()
}

// AssignString tries to parse and assign the passed
// source string as value of the implementing object.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for normalized and nil for err.
// AssignString implements strfmt.StringAssignable
func (n *NullableID) AssignString(source string) (normalized bool, err error) {
	newID, err := NullableID(source).Normalized()
	if err != nil {
		return false, err
	}
	*n = newID
	return string(newID) == source, nil
}

// Scan implements the database/sql.Scanner interface.
func (n *NullableID) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*n = NullableID(x).NormalizedUnchecked()
	case []byte:
		*n = NullableID(x).NormalizedUnchecked()
	case nil:
		*n = Null
	default:
		return errors.Errorf("can't scan SQL value of type %T as vat.NullableID", value)
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
