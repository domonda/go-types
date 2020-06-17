package bank

import (
	"database/sql/driver"
	"fmt"

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

// BIC is a SWIFT Business Identifier Code.
// BIC implements the database/sql.Scanner and database/sql/driver.Valuer interfaces
// and will treat an empty BIC string as SQL NULL value.
type BIC string

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (bic *BIC) ScanString(source string) (normalized bool, err error) {
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

// Nullable returns the BIC as NullableBIC
func (bic BIC) Nullable() NullableBIC {
	return NullableBIC(bic)
}

// Validate returns an error if this is not a valid SWIFT Business Identifier Code
func (bic BIC) Validate() error {
	length := len(bic)
	if !(length == BICMinLength || length == BICMaxLength) {
		return fmt.Errorf("invalid BIC %q length: %d", string(bic), length)
	}
	subMatches := bicExactRegex.FindStringSubmatch(string(bic))
	// fmt.Println(subMatches)
	if len(subMatches) != 5 {
		return fmt.Errorf("invalid BIC %q: no regex match", string(bic))
	}
	countryCode := country.Code(subMatches[2])
	_, isValidCountry := ibanCountryLengthMap[countryCode]
	if !isValidCountry {
		return fmt.Errorf("invalid BIC %q country code: %q", string(bic), countryCode)
	}
	if _, isFalse := falseBICs[bic]; isFalse {
		return fmt.Errorf("BIC %q is in list of invalid BICs", string(bic))
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
		*bic = BIC(BICNull)
	default:
		return fmt.Errorf("can't scan SQL value of type %T as BIC", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (bic BIC) Value() (driver.Value, error) {
	return string(bic), nil
}

var falseBICs = map[BIC]struct{}{
	"AUTOBANK":    {},
	"DIENSTGEBER": {},
	"GELISTET":    {},
	"FACILITY":    {},
}
