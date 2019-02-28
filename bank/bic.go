package bank

import (
	"database/sql/driver"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/country"
)

// ValidateBIC returns str as valid BIC or an error.
func ValidateBIC(str string) (BIC, error) {
	err := BIC(str).Validate()
	if err != nil {
		return "", err
	}
	return BIC(str), nil
}

func StringIsBIC(str string) bool {
	return BIC(str).Valid()
}

// BICNull is an empty string and will be treatet as SQL NULL.
const BICNull = ""

// BIC is a SWIFT Business Identifier Code.
// BIC implements the database/sql.Scanner and database/sql/driver.Valuer interfaces
// and will treat an empty BIC string as SQL NULL value.
type BIC string

// NullableBIC is a BIC value which can hold an emtpy string ("") as the null value.
type NullableBIC = BIC

// AssignString tries to parse and assign the passed
// source string as value of the implementing object.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for normalized and nil for err.
// AssignString implements strfmt.StringAssignable
func (bic *BIC) AssignString(source string) (normalized bool, err error) {
	err = BIC(source).Validate()
	if err != nil {
		return false, err
	}
	*bic = BIC(source)
	return true, nil
}

// Valid returns if this is a valid SWIFT Business Identifier Code
func (bic BIC) Valid() bool {
	return bic.Validate() == nil
}

// Validate returns an error if this is not a valid SWIFT Business Identifier Code
func (bic BIC) Validate() error {
	length := len(bic)
	if !(length == BICMinLength || length == BICMaxLength) {
		return errors.Errorf("invalid BIC '%s' length: %d", bic, length)
	}
	subMatches := bicExactRegex.FindStringSubmatch(string(bic))
	// fmt.Println(subMatches)
	if len(subMatches) != 5 {
		return errors.Errorf("invalid BIC '%s': no regex match", bic)
	}
	countryCode := country.Code(subMatches[2])
	_, isValidCountry := ibanCountryLengthMap[countryCode]
	if !isValidCountry {
		return errors.Errorf("invalid BIC '%s' country code: '%s'", bic, countryCode)
	}
	if _, isFalse := falseBICs[bic]; isFalse {
		return errors.Errorf("BIC '%s' is in list of invalid BICs", bic)
	}
	return nil
}

func (bic BIC) Parse() (bankCode string, countryCode country.Code, branchCode string, isValid bool) {
	length := len(bic)
	if !(length == BICMinLength || length == BICMaxLength) {
		return "", "", "", false
	}
	subMatches := bicExactRegex.FindStringSubmatch(string(bic))
	// fmt.Println(subMatches)
	if len(subMatches) != 5 {
		return "", "", "", false
	}
	countryCode = country.Code(subMatches[2])
	_, isValidCountry := ibanCountryLengthMap[countryCode]
	if !isValidCountry {
		return "", "", "", false
	}
	_, isFalse := falseBICs[bic]
	if isFalse {
		return "", "", "", false
	}
	bankCode = subMatches[1]
	branchCode = subMatches[4]
	return bankCode, countryCode, branchCode, true
}

func (bic BIC) TrimBranchCode() BIC {
	if len(bic) <= 8 {
		return bic
	}
	return bic[:8]
}

func (bic BIC) IsTestBIC() bool {
	return bic.Valid() && bic[7] == '0'
}

func (bic BIC) IsPassiveSWIFT() bool {
	return bic.Valid() && bic[7] == '1'
}

func (bic BIC) ReceiverPaisFees() bool {
	return bic.Valid() && bic[7] == '2'
}

// Scan implements the database/sql.Scanner interface.
func (bic *BIC) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*bic = BIC(x)
	case []byte:
		*bic = BIC(x)
	case nil:
		*bic = BICNull
	default:
		return errors.Errorf("can't scan SQL value of type %T as BIC", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (bic BIC) Value() (driver.Value, error) {
	if bic == BICNull {
		return nil, nil
	}
	return string(bic), nil
}

var falseBICs = map[BIC]struct{}{
	"AUTOBANK":    {},
	"DIENSTGEBER": {},
	"GELISTET":    {},
}
