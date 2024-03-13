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
//
// If validate is true, the source string is checked
// for validity before it is assigned to the type.
//
// If validate is false and the source string
// can still be assigned in some non-normalized way
// it will be assigned without returning an error.
func (iban *IBAN) ScanString(source string, validate bool) error {
	newIBAN, err := IBAN(source).Normalized()
	if err != nil {
		if validate {
			return err
		}
		newIBAN = IBAN(source)
	}
	*iban = newIBAN
	return nil
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
	countryLength, found := countryIBANLength[country.Code(iban[:2])]
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

func (iban *IBAN) BankAndAccountNumbers() (bankNo, accountNo string, err error) {
	country := iban.CountryCode()
	if country == "" {
		return "", "", fmt.Errorf("invalid IBAN: %q", string(*iban))
	}
	getNumbers, found := getBankAndAccountNumbers[country]
	if !found {
		return "", "", fmt.Errorf("can't extract bank and account numbers from IBAN: %q", string(*iban))
	}
	return getNumbers(string(*iban))
}

var getBankAndAccountNumbers = map[country.Code]func(string) (string, string, error){
	country.AT: func(iban string) (string, string, error) {
		if len(iban) < countryIBANLength[country.AT] {
			return "", "", errors.New("IBAN too short")
		}
		return iban[4:9], iban[9:], nil
	},
	country.CH: func(iban string) (string, string, error) {
		if len(iban) < countryIBANLength[country.CH] {
			return "", "", errors.New("IBAN too short")
		}
		return iban[4:9], iban[9:], nil
	},
	country.DE: func(iban string) (string, string, error) {
		if len(iban) < countryIBANLength[country.DE] {
			return "", "", errors.New("IBAN too short")
		}
		return iban[4:12], iban[12:], nil
	},
}

var countryIBANLength = map[country.Code]int{
	country.AL: 28,
	country.AD: 24,
	country.AT: 20,
	country.AZ: 28,
	country.BH: 22,
	country.BY: 28,
	country.BE: 16,
	country.BA: 20,
	country.BR: 29,
	country.BG: 22,
	country.CR: 22,
	country.HR: 21,
	country.CY: 28,
	country.CZ: 24,
	country.DK: 18,
	country.DO: 28,
	country.SV: 28,
	country.EE: 20,
	country.FO: 18,
	country.FI: 18,
	country.FR: 27,
	country.GE: 22,
	country.DE: 22,
	country.GI: 23,
	country.GR: 27,
	country.GL: 18,
	country.GT: 28,
	country.HU: 28,
	country.IS: 26,
	country.IQ: 23,
	country.IE: 22,
	country.IL: 23,
	country.IT: 27,
	country.JO: 30,
	country.KZ: 20,
	country.XK: 20,
	country.KW: 30,
	country.LV: 21,
	country.LB: 28,
	country.LI: 21,
	country.LT: 20,
	country.LU: 20,
	country.MK: 19,
	country.MT: 31,
	country.MR: 27,
	country.MU: 30,
	country.MD: 24,
	country.MC: 27,
	country.ME: 22,
	country.NL: 18,
	country.NO: 15,
	country.PK: 24,
	country.PS: 29,
	country.PL: 28,
	country.PT: 25,
	country.QA: 29,
	country.RO: 24,
	country.LC: 32,
	country.SM: 27,
	country.ST: 25,
	country.SA: 24,
	country.RS: 22,
	country.SC: 31,
	country.SK: 24,
	country.SI: 19,
	country.ES: 24,
	country.SE: 24,
	country.CH: 21,
	country.TL: 23,
	country.TN: 24,
	country.TR: 26,
	country.UA: 29,
	country.AE: 23,
	country.GB: 22,
	country.VG: 24,

	country.GG: 22, // valid BIC but, can use GB or FR in IBAN
	country.JE: 22, // valid BIC but, can use GB or FR in IBAN
}
