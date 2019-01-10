package bank

import (
	"database/sql/driver"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/strutil"
	"github.com/guregu/null"
)

var ibanRegex = regexp.MustCompile(`^([A-Z]{2})(\d{2})([A-Z\d]{8,30})$`)

const (
	IBANMinLength = 15
	IBANMaxLength = 32
)

// NormalizeIBAN returns str as normalized IBAN or an error.
func NormalizeIBAN(str string) (IBAN, error) {
	return IBAN(str).Normalized()
}

// StringIsIBAN returns if a string can be parsed as IBAN.
func StringIsIBAN(str string) bool {
	_, err := NormalizeIBAN(str)
	return err == nil
}

var IBANFinder ibanFinder

type ibanFinder struct{}

func (ibanFinder) FindAllIndex(str []byte, n int) (result [][]int) {
	strLen := len(str)
	max := strLen - IBANMinLength
	for i := 0; i <= max; i++ {
		countryCode := country.Code(str[i : i+2])
		countryLength, found := ibanCountryLengthMap[countryCode]
		if found {
			end := i + countryLength
			if end <= strLen {
				if IBAN(str[i:end]).Valid() {
					result = append(result, []int{i, end})
					i = end - 1
					continue
				}
			}
		}
	}
	return result
}

// IBAN is a International Bank Account Number.
// IBAN implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty string IBAN as SQL NULL value.
type IBAN string

// AssignString tries to parse and assign the passed
// source string as value of the implementing object.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for normalized and nil for err.
// AssignString implements strfmt.StringAssignable
func (iban *IBAN) AssignString(source string) (normalized bool, err error) {
	newIBAN, err := IBAN(source).Normalized()
	if err != nil {
		return false, err
	}
	*iban = newIBAN
	return newIBAN == IBAN(source), nil
}

// Valid returns if this is a valid SWIFT Business Identifier Code
func (iban IBAN) Valid() bool {
	_, err := iban.Normalized()
	return err == nil
}

func (iban IBAN) ValidAndNormalized() bool {
	norm, err := iban.Normalized()
	return err == nil && iban == norm
}

// CountryCode returns the country code of the IBAN
func (iban IBAN) CountryCode() country.Code {
	if !iban.Valid() {
		return ""
	}
	return country.Code(iban[:2])
}

// Normalized returns the iban in normalized form,
// or an error if the format can't be detected.
func (iban IBAN) Normalized() (IBAN, error) {
	switch {
	case iban == "":
		return "", errors.New("empty IBAN")
	case len(iban) < IBANMinLength:
		return "", errors.New("IBAN too short")
	}
	countryLength, found := ibanCountryLengthMap[country.Code(iban[:2])]
	if !found {
		return "", errors.New("invalid IBAN country code")
	}
	normalized := IBAN(strutil.RemoveRunesString(string(iban), unicode.IsSpace))
	if len(normalized) != countryLength {
		// fmt.Println(normalized, len(normalized), countryLength)
		return "", errors.New("wrong IBAN length")
	}
	if !ibanRegex.MatchString(string(normalized)) {
		return "", errors.New("invalid IBAN characters")
	}
	if !normalized.isCheckSumValid() {
		return "", errors.New("invalid IBAN check sum")
	}
	return normalized, nil
}

func (iban IBAN) NormalizedOrUnchanged() IBAN {
	normalized, err := iban.Normalized()
	if err != nil {
		return iban
	}
	return normalized
}

func (iban IBAN) NormalizedOrEmpty() IBAN {
	normalized, err := iban.Normalized()
	if err != nil {
		return ""
	}
	return normalized
}

// NormalizedWithSpaces returns the iban in normalized form with spaces every 4 characters,
// or an error if the format can't be detected.
func (iban IBAN) NormalizedWithSpaces() (IBAN, error) {
	norm, err := iban.Normalized()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	normLen := len(norm)
	for i := 0; i < normLen; i += 4 {
		if i > 0 {
			b.WriteByte(' ')
		}
		end := i + 4
		if end > normLen {
			end = normLen
		}
		b.WriteString(string(norm)[i:end])
	}
	return IBAN(b.String()), nil
}

func writeIBANRuneToCheckSumBuf(r rune, b *strings.Builder) {
	if r >= 'A' && r <= 'Z' {
		i := int(r - 'A' + 10)
		b.WriteString(strconv.Itoa(i))
	} else {
		b.WriteRune(r)
	}
}

func (iban IBAN) isCheckSumValid() bool {
	// fmt.Println("IsCheckSumValid", iban)
	if len(iban) < IBANMinLength {
		return false
	}
	var b strings.Builder
	for _, r := range iban[4:] {
		writeIBANRuneToCheckSumBuf(r, &b)
	}
	for _, r := range iban[:4] {
		writeIBANRuneToCheckSumBuf(r, &b)
	}
	str := b.String()
	sum64, err := strconv.ParseUint(str, 10, 64)
	if err == nil {
		// If the checksum string fits into a uint64,
		// use it as fasted way to calculate
		valid := sum64%97 == 1
		// fmt.Println("Valid IBAN:", iban)
		return valid
	}
	// Checksum string is to big to be parsed as uin64,
	// so parse it as big.Int
	sumBig, ok := big.NewInt(0).SetString(str, 10)
	if !ok {
		return false
	}
	valid := sumBig.Mod(sumBig, big.NewInt(97)).Int64() == 1
	// fmt.Println("Valid IBAN:", iban)
	return valid
}

// Scan implements the database/sql.Scanner interface.
func (iban *IBAN) Scan(value interface{}) error {
	var ns null.String
	err := ns.Scan(value)
	if err != nil {
		return err
	}
	if ns.Valid {
		*iban = IBAN(ns.String)
	} else {
		*iban = ""
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (iban IBAN) Value() (driver.Value, error) {
	if iban == "" {
		return nil, nil
	}
	return string(iban), nil
}

func (iban *IBAN) BankAndAccountNumbers() (bankNr, accountNr string, err error) {
	country := iban.CountryCode()
	if country == "" {
		return "", "", errors.Errorf("Invalid IBAN: '%s'", *iban)
	}
	getNumbers, found := bankAndAccountNumbers[country]
	if !found {
		return "", "", errors.Errorf("Can't extract bank and account numbers from IBAN: '%s'", *iban)
	}
	bankNr, accountNr = getNumbers(string(*iban))
	return bankNr, accountNr, nil
}

var bankAndAccountNumbers = map[country.Code]func(string) (string, string){
	"AT": func(iban string) (string, string) {
		return iban[4:9], iban[9:]
	},
	"CH": func(iban string) (string, string) {
		return iban[4:9], iban[9:]
	},
	"DE": func(iban string) (string, string) {
		return iban[4:12], iban[12:]
	},
}

var ibanCountryLengthMap = map[country.Code]int{
	"AL": 28,
	"AD": 24,
	"AT": 20,
	"AZ": 28,
	"BH": 22,
	"BY": 28,
	"BE": 16,
	"BA": 20,
	"BR": 29,
	"BG": 22,
	"CR": 22,
	"HR": 21,
	"CY": 28,
	"CZ": 24,
	"DK": 18,
	"DO": 28,
	"SV": 28,
	"EE": 20,
	"FO": 18,
	"FI": 18,
	"FR": 27,
	"GE": 22,
	"DE": 22,
	"GI": 23,
	"GR": 27,
	"GL": 18,
	"GT": 28,
	"HU": 28,
	"IS": 26,
	"IQ": 23,
	"IE": 22,
	"IL": 23,
	"IT": 27,
	"JO": 30,
	"KZ": 20,
	"XK": 20,
	"KW": 30,
	"LV": 21,
	"LB": 28,
	"LI": 21,
	"LT": 20,
	"LU": 20,
	"MK": 19,
	"MT": 31,
	"MR": 27,
	"MU": 30,
	"MD": 24,
	"MC": 27,
	"ME": 22,
	"NL": 18,
	"NO": 15,
	"PK": 24,
	"PS": 29,
	"PL": 28,
	"PT": 25,
	"QA": 29,
	"RO": 24,
	"LC": 32,
	"SM": 27,
	"ST": 25,
	"SA": 24,
	"RS": 22,
	"SC": 31,
	"SK": 24,
	"SI": 19,
	"ES": 24,
	"SE": 24,
	"CH": 21,
	"TL": 23,
	"TN": 24,
	"TR": 26,
	"UA": 29,
	"AE": 23,
	"GB": 22,
	"VG": 24,

	"GG": 22, // valid BIC but, can use GB or FR in IBAN
	"JE": 22, // valid BIC but, can use GB or FR in IBAN
}
