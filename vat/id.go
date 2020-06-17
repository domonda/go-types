package vat

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"unicode"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/strutil"
)

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
func (id ID) NormalizedUnchecked() ID {
	return ID(strutil.RemoveRunesString(string(id), unicode.IsSpace, unicode.IsPunct))
}

// Normalized returns the id in normalized form,
// or an error if the VAT ID is not valid.
func (id ID) Normalized() (ID, error) {
	normalized := id.NormalizedUnchecked()

	// Check length
	if len(normalized) < IDMinLength {
		return "", fmt.Errorf("VAT ID %q is too short", string(id))
	}
	if len(normalized) > IDMaxLength {
		return "", fmt.Errorf("VAT ID %q is too long", string(id))
	}

	// Check country code
	countryCode := country.Code(normalized[:2])
	if !countryCode.Valid() {
		return "", fmt.Errorf("VAT ID %q has an invalid country code: %q", string(id), string(countryCode))
	}

	// Check format with country specific regex
	regex, found := vatidRegex[countryCode]
	if !found {
		return "", fmt.Errorf("VAT ID %q has an unsupported country code: %q", string(id), string(countryCode))
	}
	if !regex.MatchString(string(normalized)) {
		return "", fmt.Errorf("VAT ID %q has an invalid format", string(id))
	}

	// Test check-sum if a function is available for the country
	check, found := vatidCheckSum[countryCode]
	if found && !check(normalized) {
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

// NullableID returns the id as NullableID
func (id ID) NullableID() NullableID {
	return NullableID(id)
}

// CountryCode returns the country.Code of the VAT ID,
// or country.Invalid if the id is not valid.
func (id ID) CountryCode() country.Code {
	norm, err := id.Normalized()
	if err != nil {
		return country.Invalid
	}
	code := country.Code(norm[:2])
	if code == "EL" {
		return "GR"
	}
	return code
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
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (id *ID) ScanString(source string) (normalized bool, err error) {
	newID, err := ID(source).Normalized()
	if err != nil {
		return false, err
	}
	*id = newID
	return string(newID) == source, nil
}

// Scan implements the database/sql.Scanner interface.
func (id *ID) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*id = ID(x).NormalizedUnchecked()
	case []byte:
		*id = ID(x).NormalizedUnchecked()
	case nil:
		return errors.New("can't scan SQL NULL as vat.ID")
	default:
		return fmt.Errorf("can't scan SQL value of type %T as vat.ID", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (id ID) Value() (driver.Value, error) {
	normalized := id.NormalizedUnchecked()
	if normalized == "" {
		return nil, nil
	}
	return string(normalized), nil
}

func vatidCheckSumAT(id ID) bool {
	nonSpaceCount := 0
	sum := 0
	for _, r := range id {
		if unicode.IsSpace(r) {
			continue
		}
		nonSpaceCount++
		if nonSpaceCount > 3 {
			intVal := int(r - '0')
			if nonSpaceCount == 11 {
				sum := (10 - (sum+4)%10) % 10
				return intVal == sum
			}
			if nonSpaceCount&1 == 0 {
				// C2, C4, C6, C8
				sum += intVal
			} else {
				// C3, C5, C7
				sum += intVal/5 + intVal*2%10
			}
		}
	}
	return false
}

func vatidCheckSumDE(id ID) bool {
	nonSpaceCount := 0
	P := 10
	for _, r := range id {
		if unicode.IsSpace(r) {
			continue
		}
		nonSpaceCount++
		if nonSpaceCount > 2 {
			intVal := int(r - '0')
			if nonSpaceCount == 11 {
				// fmt.Println("final C:", string(r), "P:", P)
				return intVal == (11-P)%10
			}
			M := (intVal + P) % 10
			if M == 0 {
				M = 10
			}
			P = (2 * M) % 11
			// fmt.Println("C:", string(r), "P:", P)
		}
	}
	return false
}
