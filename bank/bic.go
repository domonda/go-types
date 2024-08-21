package bank

import (
	"database/sql/driver"
	"fmt"
	"strings"

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

// BIC is a SWIFT Business Identifier Code (also know as SWIFT-Code).
// BIC implements the database/sql.Scanner and database/sql/driver.Valuer interfaces
// and will treat an empty BIC string as SQL NULL value.
type BIC string

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// If validate is true, the source string is checked
// for validity before it is assigned to the type.
//
// If validate is false and the source string
// can still be assigned in some non-normalized way
// it will be assigned without returning an error.
func (bic *BIC) ScanString(source string, validate bool) error {
	if validate && BIC(source).Validate() != nil {
		return BIC(source).Validate()
	}
	*bic = BIC(source)
	return nil
}

// Valid returns if this is a valid SWIFT Business Identifier Code
func (bic BIC) Valid() bool {
	return bic.Validate() == nil
}

// Validate returns an error if this is not a valid SWIFT Business Identifier Code
func (bic BIC) Validate() error {
	length := len(bic)
	if length != BICMinLength && length != BICMaxLength {
		return fmt.Errorf("invalid BIC %q length: %d", string(bic), length)
	}
	subMatches := bicExactRegex.FindStringSubmatch(string(bic))
	// fmt.Println(subMatches)
	if len(subMatches) != 5 {
		return fmt.Errorf("invalid BIC %q: no regex match", string(bic))
	}
	countryCode := country.Code(subMatches[2])
	_, isValidCountry := countryIBANLength[countryCode]
	if !isValidCountry {
		return fmt.Errorf("invalid BIC %q country code: %q", string(bic), countryCode)
	}
	if _, isFalse := falseBICs[bic]; isFalse {
		return fmt.Errorf("BIC %q is in list of invalid BICs", string(bic))
	}
	return nil
}

// Normalized returns the BIC normalized to a length of 11 characters
// by removing spaces and appending "XXX" in case of a valid length of 8 charaters.
// Returns the BIC unchanged in case of an error.
func (bic BIC) Normalized() (BIC, error) {
	norm := BIC(strings.ReplaceAll(string(bic), " ", ""))
	if err := norm.Validate(); err != nil {
		return bic, err
	}
	if len(norm) == 8 {
		norm += "XXX"
	}
	return norm, nil
}

// NormalizedShort returns the BIC normalized to a length of 8 or 11 characters
// by removing spaces trimming the "XXX" suffix in case of a valid length of 8 charaters.
// Returns the BIC unchanged in case of an error.
func (bic BIC) NormalizedShort() (BIC, error) {
	norm := BIC(strings.ReplaceAll(string(bic), " ", ""))
	if err := norm.Validate(); err != nil {
		return bic, err
	}
	if len(norm) == 11 && strings.HasSuffix(string(norm), "XXX") {
		norm = norm[:8]
	}
	return norm, nil
}

// Nullable returns the BIC as NullableBIC
func (bic BIC) Nullable() NullableBIC {
	return NullableBIC(bic)
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
	_, isValidCountry := countryIBANLength[countryCode]
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

// CountryCode of the BIC.
// May be invalid if the BIC is invalid.
func (bic BIC) CountryCode() country.Code {
	_, cc, _, _ := bic.Parse()
	return cc
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
func (bic *BIC) Scan(value any) error {
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
	"AMTSGERICHT": {},
	"AUTOBANK":    {},
	"DEUTSCHLAND": {},
	"DIENSTGEBER": {},
	"DOCUMENT":    {},
	"DOKUMENT":    {},
	"FACILITY":    {},
	"GELISTET":    {},
	"GESAMTNETTO": {},
}
