package vat

import (
	"regexp"
	"strings"

	"github.com/domonda/go-types/country"
)

const (
	// IDMinLength is the minium length of a VAT ID
	IDMinLength = 4

	// IDMaxLength is the maximum length of a VAT ID
	IDMaxLength = 14 + 2 // allow 2 spaces
)

// idRegex are matched with the result of ID.NormalizedUnchecked()
// that uppercases everything and removes spaces and punctuation
// https://de.wikipedia.org/wiki/Umsatzsteuer-Identifikationsnummer
// http://www.pruefziffernberechnung.de/U/USt-IdNr.shtml
// Online debug tool: https://regex101.com/
var idRegex = map[country.Code]*regexp.Regexp{
	"AT": regexp.MustCompile(`^ATU\d{8}$`),
	"BE": regexp.MustCompile(`^BE\d{10}$`),
	"BG": regexp.MustCompile(`^BG\d{9,10}$`),
	"CH": regexp.MustCompile(`^CHE-?(?:\d{9}|(?:\d{3}\.\d{3}\.\d{3}))$`),
	"CY": regexp.MustCompile(`^CY\d{8}[A-Z]$`),
	"CZ": regexp.MustCompile(`^CZ\d{8,10}$`),
	"DE": regexp.MustCompile(`^DE[1-9]\d{8}$`),
	"DK": regexp.MustCompile(`^DK\d{8}$`),
	"EE": regexp.MustCompile(`^EE\d{9}$`),
	"EL": regexp.MustCompile(`^EL\d{9}$`), // greece GR
	// Spain — four distinct sub-formats. checkSumES below disambiguates
	// and validates the appropriate control digit/letter:
	//   [ABEH]\d{8}            CIF requiring numeric check (juridical)
	//   [PQSNWR]\d{7}[A-J]     CIF requiring letter check  (juridical)
	//   [CDFGJUV]\d{7}[0-9A-J] CIF accepting either form   (juridical)
	//   \d{8}[A-Z]             DNI / NIF (citizens)
	//   [YZ]\d{7}[A-Z]         NIE (foreigners, Y/Z form)
	//   [KLMX]\d{7}[A-Z]       NIE (X) and historical K, L, M
	// See https://en.wikipedia.org/wiki/VAT_identification_number and
	// https://en.wikipedia.org/wiki/NIE_number.
	"ES": regexp.MustCompile(`^ES(?:[ABEH]\d{8}|[PQSNWR]\d{7}[A-J]|[CDFGJUV]\d{7}[0-9A-J]|\d{8}[A-Z]|[YZ]\d{7}[A-Z]|[KLMX]\d{7}[A-Z])$`),
	"FI": regexp.MustCompile(`^FI\d{8}$`),
	"FR": regexp.MustCompile(`^FR[0-9A-Z][0-9A-Z]\d{9}$`),
	"GB": regexp.MustCompile(`^GB(?:\d{9})|(?:\d{12})|(?:GD\d{3})|(?:HA\d{3})$`),
	"HR": regexp.MustCompile(`^HR\d{11}$`),
	"HU": regexp.MustCompile(`^HU\d{8,9}$`),
	"IE": regexp.MustCompile(`^IE(?:\d[0-9A-Z]\d{5}[A-Z])|(?:\d{7}[A-W][A-I])$`),
	"IT": regexp.MustCompile(`^IT\d{11}$`),
	"LT": regexp.MustCompile(`^LT(?:\d{9}|\d{12})$`),
	"LU": regexp.MustCompile(`^LU\d{8}$`),
	"LV": regexp.MustCompile(`^LV\d{11}$`),
	"MT": regexp.MustCompile(`^MT\d{8}$`),
	"NL": regexp.MustCompile(`^NL\d{9}B\d{2}$`),
	"NO": regexp.MustCompile(`^NO[89]\d{8}(?:MVA)?$`), // https://vatstack.com/articles/norway-vat-number-validation
	"PL": regexp.MustCompile(`^PL\d{10}$`),
	"PT": regexp.MustCompile(`^PT\d{9}$`),
	"RO": regexp.MustCompile(`^RO\d{2,10}$`),
	"SE": regexp.MustCompile(`^SE\d{12}$`),
	"SI": regexp.MustCompile(`^SI\d{8}$`),
	"SK": regexp.MustCompile(`^SK\d{10}$`),
	// > For the non-Union scheme, the taxable person can choose any Member State to be
	// > the Member State of identification. That Member State will allocate an individual
	// > VAT identification number to the taxable person (using the format EUxxxyyyyyz).
	// Taken straight from: https://ec.europa.eu/taxation_customs/sites/taxation/files/resources/documents/taxation/vat/how_vat_works/telecom/one-stop-shop-guidelines_en.pdf
	MOSSSchemaVATCountryCode: regexp.MustCompile(`^EU\d{9}$`),
}

// checkSumFuncs assume that a idRegex matched before calling
// the matched check-sum function for further format checks.
//
// List of check-sum algorithms: https://www.bmf.gv.at/dam/jcr:9f9f8d5f-5496-4886-aa4f-81a4e39ba83e/BMF_UID_Konstruktionsregeln.pdf
var checkSumFuncs = map[country.Code]func(raw, normalized ID) bool{
	"AT": checkSumAT,
	"DE": checkSumDE,
	"ES": checkSumES,
	"NO": checkSumNO,
}

func checkSumAT(raw, normalized ID) bool {
	nonSpaceCount := 0
	sum := 0
	for _, r := range normalized {
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

func checkSumDE(raw, normalized ID) bool {
	nonSpaceCount := 0
	P := 10
	for _, r := range normalized {
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

// dniCheckLetter returns the Spanish DNI/NIE control letter for a
// non-negative integer using the standard mod-23 table.
func dniCheckLetter(n int) byte {
	const table = "TRWAGMYFPDXBNJZSQVHLCKE"
	return table[n%23]
}

// cifLuhnControl computes the CIF mod-10 control value from the seven
// middle digits using the [2,1,2,1,2,1,2] Luhn-style weights.
func cifLuhnControl(digits string) int {
	sum := 0
	for i := range 7 {
		d := int(digits[i] - '0')
		if i%2 == 0 { // positions 1,3,5,7 (1-indexed) → 0,2,4,6 here
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return (10 - sum%10) % 10
}

// checkSumES validates a Spanish VAT ID. The package-level regex picks
// one of six sub-formats (CIF numeric / CIF letter / CIF either /
// DNI / NIE Y-Z / NIE X-K-L-M) and this function applies the matching
// control-character algorithm.
//
// References:
//   - https://en.wikipedia.org/wiki/VAT_identification_number#Examples
//   - https://en.wikipedia.org/wiki/NIE_number
//   - https://www.boe.es/buscar/act.php?id=BOE-A-2007-21421 (RD 1065/2007)
func checkSumES(raw, normalized ID) bool {
	_ = raw
	if len(normalized) != 11 { // "ES" + 9 chars
		return false
	}
	body := string(normalized[2:])
	first := body[0]
	last := body[8]

	switch {
	case first >= '0' && first <= '9':
		// DNI / NIF: 8 digits + letter.
		n := 0
		for i := range 8 {
			n = n*10 + int(body[i]-'0')
		}
		return dniCheckLetter(n) == last

	case first == 'Y' || first == 'Z' || first == 'X':
		// NIE: replace prefix letter with a digit (X=0, Y=1, Z=2),
		// form an 8-digit number, apply DNI algorithm.
		var prefix byte
		switch first {
		case 'X':
			prefix = '0'
		case 'Y':
			prefix = '1'
		case 'Z':
			prefix = '2'
		}
		n := int(prefix - '0')
		for i := 1; i <= 7; i++ {
			n = n*10 + int(body[i]-'0')
		}
		return dniCheckLetter(n) == last

	case first == 'K' || first == 'L' || first == 'M':
		// Historical/special NIF starting with K, L, or M: the seven
		// digits (no prefix substitution) are run through the mod-23
		// table.
		n := 0
		for i := 1; i <= 7; i++ {
			n = n*10 + int(body[i]-'0')
		}
		return dniCheckLetter(n) == last

	default:
		// CIF: 1 entity-type letter + 7 digits + 1 control char.
		control := cifLuhnControl(body[1:8])
		expectedDigit := byte('0' + control)
		expectedLetter := "JABCDEFGHI"[control]

		onlyLetter := strings.ContainsRune("PQSNWR", rune(first))
		onlyDigit := strings.ContainsRune("ABEH", rune(first))

		switch {
		case last >= '0' && last <= '9':
			return !onlyLetter && last == expectedDigit
		case last >= 'A' && last <= 'J':
			return !onlyDigit && last == expectedLetter
		default:
			return false
		}
	}
}

// checkSumNO validates a Norwegian VAT ID.
//
// After normalization the ID has the form NO + 9-digit organisation
// number + optional "MVA". The 9-digit organisation number uses a
// MOD-11 checksum with weights 3, 2, 7, 6, 5, 4, 3, 2 applied to the
// first 8 digits; the 9th digit is the control digit.
//
// References:
//   - https://en.wikipedia.org/wiki/National_identification_numbers_in_Norway
//   - https://www.brreg.no/om-oss/oppgavene-vare/alle-registrene-vare/om-enhetsregisteret/organisasjonsnummeret/
//   - https://vatstack.com/articles/norway-vat-number-validation
func checkSumNO(raw, normalized ID) bool {
	// "No." is far more commonly an English number abbreviation than a
	// Norwegian VAT prefix, so reject before any checksum work.
	if strings.HasPrefix(string(raw), "No.") || strings.HasPrefix(string(raw), "no.") {
		return false
	}
	if len(normalized) < 11 {
		return false
	}
	// Defensive — the regex already enforces [89] in this position.
	if d := normalized[2]; d != '8' && d != '9' {
		return false
	}
	digits := normalized[2:11]
	weights := [8]int{3, 2, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := range 8 {
		sum += int(digits[i]-'0') * weights[i]
	}
	rem := sum % 11
	var check int
	switch rem {
	case 0:
		check = 0
	case 1:
		// remainder 1 would require a control digit of 10, which is
		// not representable in a single character, so the underlying
		// organisation number is invalid by construction.
		return false
	default:
		check = 11 - rem
	}
	return int(digits[8]-'0') == check
}
