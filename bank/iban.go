package bank

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/strutil"
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

// IBAN is a International Bank Account Number.
// IBAN implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty IBAN string as SQL NULL value.
type IBAN string

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (iban *IBAN) ScanString(source string) (normalized bool, err error) {
	newIBAN, err := IBAN(source).Normalized()
	if err != nil {
		return false, err
	}
	*iban = newIBAN
	return newIBAN == IBAN(source), nil
}

// Valid returns if this is a valid International Bank Account Number
func (iban IBAN) Valid() bool {
	return iban.Validate() == nil
}

// Validate returns an error if this is not a valid International Bank Account Number
func (iban IBAN) Validate() error {
	_, err := iban.Normalized()
	return err
}

func (iban IBAN) ValidAndNormalized() bool {
	norm, err := iban.Normalized()
	return err == nil && iban == norm
}

// CountryCode returns the country code of the IBAN.
// May be invalid if the IBAN is invalid.
func (iban IBAN) CountryCode() country.Code {
	norm, err := iban.Normalized()
	if err != nil {
		return country.Invalid
	}
	return country.Code(norm[:2])
}

// Normalized returns the iban in normalized form,
// or an error if the format can't be detected.
func (iban IBAN) Normalized() (IBAN, error) {
	switch {
	case iban.Nullable().IsNull():
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

func (iban IBAN) NormalizedOrNull() NullableIBAN {
	normalized, err := iban.Normalized()
	if err != nil {
		return IBANNull
	}
	return NullableIBAN(normalized)
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

// String returns the normalized IBAN string if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (iban IBAN) String() string {
	norm, err := iban.Normalized()
	if err != nil {
		return string(iban)
	}
	return string(norm)
}

// Nullable returns the IBAN as NullableIBAN
func (iban IBAN) Nullable() NullableIBAN {
	return NullableIBAN(iban)
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
func (iban *IBAN) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*iban = IBAN(x)
	case []byte:
		*iban = IBAN(x)
	case nil:
		*iban = IBAN(IBANNull)
	default:
		return fmt.Errorf("can't scan SQL value of type %T as IBAN", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (iban IBAN) Value() (driver.Value, error) {
	return string(iban), nil
}

func (iban *IBAN) BankAndAccountNumbers() (bankNr, accountNr string, err error) {
	country := iban.CountryCode()
	if country == "" {
		return "", "", fmt.Errorf("invalid IBAN: %q", string(*iban))
	}
	getNumbers, found := bankAndAccountNumbers[country]
	if !found {
		return "", "", fmt.Errorf("can't extract bank and account numbers from IBAN: %q", string(*iban))
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
