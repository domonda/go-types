package bank

import (
	"database/sql/driver"
	"regexp"
	"unicode/utf8"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/strutil"
	"github.com/guregu/null"
)

var (
	bicFindRegex  = regexp.MustCompile(`[A-Z]{4}([A-Z]{2})[A-Z2-9][A-NP-Z0-9](?:XXX|[A-WY-Z0-9][A-Z0-9]{2})?`)
	bicExactRegex = regexp.MustCompile(`^([A-Z]{4})([A-Z]{2})([A-Z2-9][A-NP-Z0-9])(XXX|[A-WY-Z0-9][A-Z0-9]{2})?$`)
)

const (
	BICMinLength = 8
	BICMaxLength = 11
)

var BICFinder bicFinder

type bicFinder struct{}

func (bicFinder) FindAllIndex(str []byte, n int) [][]int {
	// fmt.Println(string(str))
	indices := bicFindRegex.FindAllSubmatchIndex(str, n)
	if len(indices) == 0 {
		return nil
	}
	result := make([][]int, 0, len(indices))
	for _, matchIndices := range indices {
		if len(matchIndices) != 2*2 {
			panic(errors.Errorf("Expected 4 match indices but len(matchIndices) = %d", len(matchIndices)))
		}
		// for _, i := range matchIndices {
		// 	if i < 0 || i > len(str) {
		// 		fmt.Println("bicFinder invalid index", i, len(str))
		// 		continue
		// 	}
		// }
		bic := str[matchIndices[0]:matchIndices[1]]
		countryCode := country.Code(str[matchIndices[2]:matchIndices[3]])
		_, isValidCountry := ibanCountryLengthMap[countryCode]
		_, isFalse := falseBICs[BIC(bic)]
		if isValidCountry && !isFalse && bicExactRegex.Match(bic) {
			// BIC must also be surrounded by line bounds,
			// or a separator rune
			if matchIndices[0] > 0 {
				r, _ := utf8.DecodeLastRune(str[:matchIndices[0]])
				if !strutil.IsWordSeparator(r) {
					continue
				}
			}
			if matchIndices[1] < len(str) {
				r, _ := utf8.DecodeRune(str[matchIndices[1]:])
				if !strutil.IsWordSeparator(r) {
					continue
				}
			}

			result = append(result, matchIndices[:2])
		}
	}
	return result
}

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
// BIC implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty string BIC as SQL NULL value.
type BIC string

// AssignString implements strfmt.StringAssignable
func (bic *BIC) AssignString(str string) error {
	err := BIC(str).Validate()
	if err != nil {
		return err
	}
	*bic = BIC(str)
	return nil
}

// Valid returns if this is a valid SWIFT Business Identifier Code
func (bic BIC) Valid() bool {
	return bic.Validate() == nil
}

// Validate returns an error if this is not a valid SWIFT Business Identifier Code
func (bic BIC) Validate() error {
	length := len(bic)
	if !(length == BICMinLength || length == BICMaxLength) {
		return errors.Errorf("Invalid BIC '%s' length: %d", bic, length)
	}
	subMatches := bicExactRegex.FindStringSubmatch(string(bic))
	// fmt.Println(subMatches)
	if len(subMatches) != 5 {
		return errors.Errorf("Invalid BIC '%s': no regex match", bic)
	}
	countryCode := country.Code(subMatches[2])
	_, isValidCountry := ibanCountryLengthMap[countryCode]
	if !isValidCountry {
		return errors.Errorf("Invalid BIC '%s' country code: '%s'", bic, countryCode)
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
	// fmt.Println("BIC.Scan", value)
	var ns null.String
	err := ns.Scan(value)
	if err != nil {
		return err
	}
	if ns.Valid {
		*bic = BIC(ns.String)
	} else {
		*bic = ""
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (bic BIC) Value() (driver.Value, error) {
	if bic == "" {
		return nil, nil
	}
	return string(bic), nil
}

var falseBICs = map[BIC]struct{}{
	"AUTOBANK":    struct{}{},
	"DIENSTGEBER": struct{}{},
	"GELISTET":    struct{}{},
}
