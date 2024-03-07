package vat

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/strutil"
)

// MOSSSchemaVATCountryCode or the VAT Mini One Stop Shop (MOSS) is an optional scheme that allows you
// to account for VAT - normally due in multiple EU countries – in just one EU country. Check out:
// https://europa.eu/youreurope/business/taxation/vat/vat-digital-services-moss-scheme/index_en.htm
const MOSSSchemaVATCountryCode = "EU"

// ID is a european VAT ID.
// ID implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// returning errors when the ID is not valid and can't be normalized.
// Use NullableID to read and write SQL NULL values.
type ID string

// NormalizeVATID returns str as normalized VAT ID or an error.
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

// NormalizedUnchecked returns a generic normalized version of ID without performing any format checks.
// func (id ID) NormalizedUnchecked() ID {
// 	return ID(strings.ToUpper(strutil.RemoveRunesString(string(id), unicode.IsSpace, unicode.IsPunct)))
// }

// Normalized returns the id in normalized form,
// or an error if the VAT ID is not valid.
func (id ID) Normalized() (ID, error) {
	normalized := ID(strings.ToUpper(strutil.RemoveRunesString(string(id), unicode.IsSpace, unicode.IsPunct)))

	// Check length
	if len(normalized) < IDMinLength {
		return "", fmt.Errorf("VAT ID %q is too short", string(id))
	}
	if len(normalized) > IDMaxLength {
		return "", fmt.Errorf("VAT ID %q is too long", string(id))
	}

	// Check country code
	countryCode := country.Code(normalized[:2])
	if countryCode != MOSSSchemaVATCountryCode && !countryCode.Valid() {
		return "", fmt.Errorf("VAT ID %q has an invalid country code: %q", string(id), string(countryCode))
	}

	// Check format with country specific regex
	regex, ok := idRegex[countryCode]
	if !ok {
		return "", fmt.Errorf("VAT ID %q has an unsupported country code: %q", string(id), string(countryCode))
	}
	if !regex.MatchString(string(normalized)) {
		return "", fmt.Errorf("VAT ID %q has an invalid format", string(id))
	}

	// Test checkFunc-sum if a function is available for the country
	checkFunc, ok := checkSumFuncs[countryCode]
	if ok && !checkFunc(id, normalized) {
		return "", fmt.Errorf("VAT ID %q has an invalid check-sum", string(id))
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
func (id ID) Validate() error {
	_, err := id.Normalized()
	return err
}

// ValidateIsNormalized returns an error if id is not a valid and normalized VAT ID.
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

// String returns the normalized ID if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (id ID) String() string {
	norm, err := id.Normalized()
	if err != nil {
		return string(id)
	}
	return string(norm)
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for wasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (id *ID) ScanString(source string) (wasNormalized bool, err error) {
	newID, err := ID(source).Normalized()
	if err != nil {
		return false, err
	}
	*id = newID
	return string(newID) == source, nil
}

// Scan implements the database/sql.Scanner interface.
func (id *ID) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*id = ID(x)
	case []byte:
		*id = ID(x)
	case nil:
		return errors.New("can't scan SQL NULL as vat.ID")
	default:
		return fmt.Errorf("can't scan SQL value of type %T as vat.ID", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (id ID) Value() (driver.Value, error) {
	normalized, err := id.Normalized()
	if err != nil {
		return string(id), nil
	}
	return string(normalized), nil
}
