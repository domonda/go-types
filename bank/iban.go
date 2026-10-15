package bank

import (
	"database/sql/driver"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/strutil"
)

const (
	// IBANRegex is the regular expression pattern for validating an IBAN.
	IBANRegex = `^([A-Z]{2})(\d{2})([A-Z\d]{8,30})$`
	// IBANMinLength is the minimum length of a valid IBAN (15 characters for Norway).
	IBANMinLength = 15
	// IBANMaxLength is the length of the longest IBAN known to this
	// package (32 characters for Saint Lucia). The ISO 13616 standard
	// permits up to 34 characters.
	IBANMaxLength = 32
)

var ibanRegexp = regexp.MustCompile(IBANRegex)

// Compile-time check that IBAN implements types.NormalizableValidator[IBAN]
var _ types.NormalizableValidator[IBAN] = IBAN("")

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
// See NullableIBAN for a nullable version of this type.
//
// Scan, UnmarshalJSON, UnmarshalXML and ScanString store a valid IBAN
// normalized, a blank one as "" and an invalid one unchanged.
// Value writes the normalized IBAN, SQL NULL for a blank one, and returns
// an error for an invalid one. MarshalJSON and MarshalText write the
// normalized IBAN or an invalid one unchanged, and fmt prints the IBAN
// as stored. IBAN has no UnmarshalText, so encoding/json reads and writes
// the keys of a map[IBAN]V as given. See "Reading and writing values"
// in the README of the module.
type IBAN string

// ScanString implements the strfmt.Scannable interface.
// Without validate it stores source like Scan, with validate it stores
// the normalized IBAN or returns the error of Normalized,
// also for a blank source.
func (iban *IBAN) ScanString(source string, validate bool) error {
	return internal.ScanStringNormalized(iban, source, validate, false, IBAN.Normalized, nil)
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

// ValidAndNormalized returns true if the IBAN is valid and already in normalized form.
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
// All whitespace, including leading whitespace, is removed
// before the IBAN is checked.
// Returns the IBAN unchanged in case of an error.
func (iban IBAN) Normalized() (IBAN, error) {
	normalized := IBAN(strutil.RemoveRunesString(string(iban), strutil.IsSpace))
	switch {
	case normalized == "":
		return iban, errors.New("empty IBAN")
	case len(normalized) < IBANMinLength:
		return iban, errors.New("IBAN too short")
	}
	countryLength, found := countryIBANLength[country.Code(normalized[:2])]
	if !found {
		return iban, errors.New("invalid IBAN country code")
	}
	if len(normalized) != countryLength {
		return iban, errors.New("wrong IBAN length")
	}
	if !ibanRegexp.MatchString(string(normalized)) {
		return iban, errors.New("invalid IBAN characters")
	}
	if !normalized.isCheckSumValid() {
		return iban, errors.New("invalid IBAN check sum")
	}
	return normalized, nil
}

// NormalizedOrNull returns the IBAN as NullableIBAN in normalized form,
// or IBANNull if normalization fails.
func (iban IBAN) NormalizedOrNull() NullableIBAN {
	normalized, err := iban.Normalized()
	if err != nil {
		return IBANNull
	}
	return NullableIBAN(normalized)
}

// NormalizedWithSpaces returns the iban in normalized form with spaces every 4 characters,
// or an error if the format can't be detected.
// Returns the IBAN unchanged in case of an error.
func (iban IBAN) NormalizedWithSpaces() (IBAN, error) {
	norm, err := iban.Normalized()
	if err != nil {
		return iban, err
	}
	var b strings.Builder
	normLen := len(norm)
	for i := 0; i < normLen; i += 4 {
		if i > 0 {
			b.WriteByte(' ')
		}
		end := min(i+4, normLen)
		b.WriteString(string(norm)[i:end])
	}
	return IBAN(b.String()), nil
}

// Nullable returns the IBAN as NullableIBAN
func (iban IBAN) Nullable() NullableIBAN {
	return NullableIBAN(iban)
}

func (iban IBAN) isCheckSumValid() bool {
	const (
		ibanCheckSumModulo = 97
	)

	if len(iban) < IBANMinLength {
		return false
	}

	rearranged := []byte(iban[4:] + iban[:4])

	var remainder int
	for _, r := range rearranged {
		if !(isNum(r) || isUpperAZ(r)) {
			return false
		}

		if isNum(r) {
			remainder = (remainder*10 + int(r-'0')) % ibanCheckSumModulo
			continue
		}

		n := int(r - 'A' + 10)

		remainder = (remainder*10 + n/10) % ibanCheckSumModulo
		remainder = (remainder*10 + n%10) % ibanCheckSumModulo
	}

	return remainder == 1
}

// Scan implements the database/sql.Scanner interface.
// A valid IBAN is stored normalized, SQL NULL and a blank string as "",
// and a IBAN that can't be normalized unchanged without an error,
// so that one invalid value doesn't fail a whole query; Validate reports it.
func (iban *IBAN) Scan(value any) error {
	return internal.ScanNormalized(iban, value, IBAN.Normalized, nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing the
// normalized IBAN, "" for a blank IBAN, or the IBAN unchanged
// if it can't be normalized.
func (iban IBAN) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(iban, false, IBAN.Normalized, nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null leaves the IBAN
// unchanged like encoding/json does for a string type, and any other
// JSON value is an error.
func (iban *IBAN) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(iban, data, false, IBAN.Normalized, nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// so that encoding/xml and strfmt write the normalized IBAN.
// encoding/json writes map keys of string types as they are
// and doesn't call MarshalText for them.
func (iban IBAN) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(iban, IBAN.Normalized, nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (iban *IBAN) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(iban, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (iban *IBAN) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(iban, d, start, IBAN.Normalized, nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (iban *IBAN) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(iban, attr, IBAN.Normalized, nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the normalized IBAN, SQL NULL for a blank IBAN,
// or the error of Normalized for a IBAN that can't be normalized.
func (iban IBAN) Value() (driver.Value, error) {
	return internal.ValueNormalized(iban, IBAN.Normalized, nil)
}

// JSONSchema returns the JSON schema definition for the IBAN type.
func (IBAN) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:   "IBAN",
		Type:    "string",
		Pattern: IBANRegex,
	}
}

// BankAndAccountNumbers extracts the bank number and account number from the
// normalized IBAN (whitespace removed) for supported countries (AT, CH, DE).
// Returns an error if the IBAN is invalid or the country is not supported.
func (iban *IBAN) BankAndAccountNumbers() (bankNo, accountNo string, err error) {
	norm, err := iban.Normalized()
	if err != nil {
		return "", "", fmt.Errorf("invalid IBAN: %q", string(*iban))
	}
	getNumbers, found := getBankAndAccountNumbers[country.Code(norm[:2])]
	if !found {
		return "", "", fmt.Errorf("can't extract bank and account numbers from IBAN: %q", string(*iban))
	}
	return getNumbers(string(norm))
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
	country.EG: 29,

	country.IM: 22, // valid BIC but, can use GB or FR in IBAN
	country.GG: 22, // valid BIC but, can use GB or FR in IBAN
	country.JE: 22, // valid BIC but, can use GB or FR in IBAN
}
