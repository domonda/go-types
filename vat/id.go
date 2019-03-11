package vat

import (
	"database/sql/driver"
	"unicode"

	"github.com/domonda/errors"
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
	l := len(str)
	if l < IDMinLength || l > IDMaxLength {
		return false
	}
	countryCode := country.Code(str[:2])
	regex, found := vatidRegex[countryCode]
	// return found && regex.Match(str)
	if !found || !regex.Match(str) {
		return false
	}
	check, found := vatidCheckSum[countryCode]
	return !found || check(string(str))
}

func isVATIDSplitRune(r rune) bool {
	return unicode.IsSpace(r) || r == ':'
}

func isVATIDTrimRune(r rune) bool {
	return unicode.IsPunct(r)
}

// Normalized returns the id in normalized form,
// or an error if the VAT ID is not valid.
func (id ID) Normalized() (ID, error) {
	if len(id) < IDMinLength {
		return "", errors.Errorf("VAT ID '%s' is too short", id)
	}
	countryCode := country.Code(id[:2])
	regex, found := vatidRegex[countryCode]
	if !found {
		return "", errors.Errorf("VAT ID '%s' has an unsupported country code: '%s'", id, countryCode)
	}
	normalized := strutil.RemoveRunesString(string(id), unicode.IsSpace, unicode.IsPunct)
	if !regex.MatchString(normalized) {
		return "", errors.Errorf("VAT ID '%s' has an invalid format", id)
	}
	check, found := vatidCheckSum[countryCode]
	if found && !check(normalized) {
		return "", errors.Errorf("VAT ID '%s' has an invalid check-sum", id)
	}
	return ID(normalized), nil
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
		return errors.Errorf("VAT ID is valid but not normalized: '%s'", id)
	}
	return nil
}

// NullableID returns the id as NullableID
func (id ID) NullableID() NullableID {
	return NullableID(id)
}

// CountryCode returns the country.Code of the VAT ID,
// or country.Null if the id is not valid.
func (id ID) CountryCode() country.Code {
	norm, err := id.Normalized()
	if err != nil {
		return country.Null
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

// AssignString tries to parse and assign the passed
// source string as value of the implementing object.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for normalized and nil for err.
// AssignString implements strfmt.StringAssignable
func (id *ID) AssignString(source string) (normalized bool, err error) {
	newID, err := ID(source).Normalized()
	if err != nil {
		return false, err
	}
	*id = newID
	return string(newID) == source, nil
}

// Scan implements the database/sql.Scanner interface.
func (id *ID) Scan(value interface{}) error {
	var newID ID
	switch x := value.(type) {
	case string:
		newID = ID(x)
	case []byte:
		newID = ID(x)
	case nil:
		return errors.New("can't scan SQL NULL as vat.ID")
	default:
		return errors.Errorf("can't scan SQL value of type %T as vat.ID", value)
	}
	newID, err := newID.Normalized()
	if err != nil {
		return err
	}
	*id = newID
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (id ID) Value() (driver.Value, error) {
	norm, err := id.Normalized()
	return string(norm), err
}

func vatidCheckSumAT(id string) bool {
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

func vatidCheckSumDE(id string) bool {
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
