package vat

import (
	"database/sql/driver"
	"unicode"

	"github.com/guregu/null"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/strutil"
)

// ID is a european VAT ID.
// ID implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty string ID as SQL NULL value.
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

// AssignString tries to parse and assign the passed
// source string as value of the implementing object.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expeced normalized format, then false is
// returned for normalized and nil for err.
// AssignString implements strfmt.StringAssignable
func (id *ID) AssignString(source string) (normalized bool, err error) {
	newID, err := ID(source).Normalized()
	if err != nil {
		return false, err
	}
	*id = newID
	return newID == ID(source), nil
}

// Scan implements the database/sql.Scanner interface.
func (id *ID) Scan(value interface{}) error {
	var ns null.String
	err := ns.Scan(value)
	if err != nil {
		return err
	}
	if ns.Valid {
		*id = ID(ns.String)
	} else {
		*id = ""
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (id ID) Value() (driver.Value, error) {
	if id == "" {
		return nil, nil
	}
	return string(id), nil
}

func (id ID) Valid() bool {
	_, err := id.Normalized()
	return err == nil
}

func (id ID) ValidAndNormalized() bool {
	norm, err := id.Normalized()
	return err == nil && id == norm
}

func (id ID) CountryCode() country.Code {
	if !id.Valid() {
		return ""
	}
	code := country.Code(id[:2])
	if code == "EL" {
		return "GR"
	}
	return code
}

func (id ID) Number() string {
	if !id.Valid() {
		return ""
	}
	return string(id[2:])
}

func (id ID) Normalized() (ID, error) {
	if len(id) < IDMinLength {
		return "", errors.New("VAT ID is too short")
	}
	countryCode := country.Code(id[:2])
	regex, found := vatidRegex[countryCode]
	if !found {
		return "", errors.New("invalid VAT ID country code: " + string(countryCode))
	}
	normalized := strutil.RemoveRunesString(string(id), unicode.IsSpace)
	if !regex.MatchString(normalized) {
		return "", errors.New("invalid VAT ID format")
	}
	check, found := vatidCheckSum[countryCode]
	if found && !check(normalized) {
		return "", errors.New("invalid VAT ID check-sum")
	}
	return ID(normalized), nil
}

func (id ID) NormalizedOrEmpty() ID {
	normalized, err := id.Normalized()
	if err != nil {
		return ""
	}
	return normalized
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
