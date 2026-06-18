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
	"BE": regexp.MustCompile(`^BE[01]\d{9}$`),
	"BG": regexp.MustCompile(`^BG\d{9,10}$`),
	"CH": regexp.MustCompile(`^CHE\d{9}$`),
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
	"GB": regexp.MustCompile(`^GB(?:\d{9}|\d{12}|GD\d{3}|HA\d{3})$`),
	"HR": regexp.MustCompile(`^HR\d{11}$`),
	"HU": regexp.MustCompile(`^HU\d{8}$`),
	"IE": regexp.MustCompile(`^IE(?:\d[0-9A-Z]\d{5}[A-Z]|\d{7}[A-W][A-I])$`),
	"IS": regexp.MustCompile(`^IS\d{5,6}$`), // Iceland VSK (EFTA, non-EU)
	"IT": regexp.MustCompile(`^IT\d{11}$`),
	"LI": regexp.MustCompile(`^LI\d{5}$`), // Liechtenstein (EFTA, non-EU)
	"LT": regexp.MustCompile(`^LT(?:\d{9}|\d{12})$`),
	"LU": regexp.MustCompile(`^LU\d{8}$`),
	"LV": regexp.MustCompile(`^LV\d{11}$`),
	"MT": regexp.MustCompile(`^MT\d{8}$`),
	// Netherlands — "NL" + 9 digits + "B" + 2-digit branch. Format only, no
	// checksum: the pre-2020 number embedded an "11-proof" (elfproef) derived
	// from the BSN/RSIN, but the btw-id issued to sole proprietors since
	// 2020-01-01 is randomly generated and does NOT satisfy it. The two forms
	// share the same shape, so a checksum gate would reject valid btw-ids;
	// rely on the format here and VIES for the authoritative check.
	"NL": regexp.MustCompile(`^NL\d{9}B\d{2}$`),
	"NO": regexp.MustCompile(`^NO[89]\d{8}(?:MVA)?$`), // https://vatstack.com/articles/norway-vat-number-validation
	"PL": regexp.MustCompile(`^PL\d{10}$`),
	"PT": regexp.MustCompile(`^PT\d{9}$`),
	"RO": regexp.MustCompile(`^RO\d{2,10}$`),
	"SE": regexp.MustCompile(`^SE\d{10}01$`),
	"SI": regexp.MustCompile(`^SI\d{8}$`),
	"SK": regexp.MustCompile(`^SK\d{10}$`),
	"SM": regexp.MustCompile(`^SM\d{5}$`), // San Marino (EU-facing via Italian intermediary)
	// Northern Ireland — post-Brexit VAT prefix for EU goods regime.
	// Same shape and same mod-97 checksum (see checkSumGB) as GB.
	NorthernIrelandVATCountryCode: regexp.MustCompile(`^XI(?:\d{9}|\d{12}|GD\d{3}|HA\d{3})$`),
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
	"BE": checkSumBE,
	"BG": checkSumBG,
	"CY": checkSumCY,
	"CZ": checkSumCZ,
	"DE": checkSumDE,
	"DK": checkSumDK,
	"EE": checkSumEE,
	"EL": checkSumEL,
	"ES": checkSumES,
	"FI": checkSumFI,
	"FR": checkSumFR,
	"GB": checkSumGB,
	"HR": checkSumHR,
	"HU": checkSumHU,
	"IT": checkSumIT,
	"LT": checkSumLT,
	"LU": checkSumLU,
	"LV": checkSumLV,
	"MT": checkSumMT,
	// NL has no checksum entry: the post-2020 btw-id is randomly generated and
	// does not satisfy the old 11-proof. See the idRegex comment for "NL".
	"NO":                          checkSumNO,
	"PL":                          checkSumPL,
	"PT":                          checkSumPT,
	"RO":                          checkSumRO,
	"SE":                          checkSumSE,
	"SI":                          checkSumSI,
	"SK":                          checkSumSK,
	NorthernIrelandVATCountryCode: checkSumGB,
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

// checkSumGB validates the United Kingdom mod-97 VAT checksum.
// Shared with the Northern Ireland "XI" prefix, which uses the same algorithm.
//
// After normalization the ID is GB/XI followed by either:
//
//   - 9 digits  — standard VRN, checksum applies
//   - 12 digits — branch trader VAT; the first 9 are a VRN and carry the checksum,
//     the trailing 3 are the branch code
//   - GD\d{3}   — government department
//   - HA\d{3}   — health authority
//
// The GD and HA forms have no checksum.
//
// The 9-digit VRN uses MOD-97 with two accepted residues — "old" format requires
// the weighted sum plus the 2-digit check to be divisible by 97; "new" format
// (introduced 2010) adds a constant 55 before the mod. References:
//
//   - https://en.wikipedia.org/wiki/VAT_identification_number#United_Kingdom
//   - https://library.croneri.co.uk/cch_uk/btr/85-260
func checkSumGB(raw, normalized ID) bool {
	_ = raw
	body := normalized[2:]
	// GD###/HA### have no checksum; the regex already validated the shape.
	if body[0] < '0' || body[0] > '9' {
		return true
	}
	weights := [7]int{8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := range 7 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	check := int(body[7]-'0')*10 + int(body[8]-'0')
	return (sum+check)%97 == 0 || (sum+check+55)%97 == 0
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

// checkSumBE validates a Belgian VAT ID.
//
// After normalization the ID is BE + 10 digits where the first digit is
// 0 or 1 (enforced by the regex). The last 2 digits are the control:
// expected = 97 - (first8 mod 97).
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumBE(_, normalized ID) bool {
	body := normalized[2:]
	n := 0
	for i := range 8 {
		n = n*10 + int(body[i]-'0')
	}
	check := int(body[8]-'0')*10 + int(body[9]-'0')
	return 97-n%97 == check
}

// checkSumBG validates a Bulgarian VAT ID.
//
// 9-digit form is the EIK (Unified Identification Code) for legal
// entities: weighted mod-11 with a fallback weight set when the first
// remainder is 10. 10-digit form is the EGN (personal civil number):
// fixed weights [2,4,8,5,10,9,7,3,6] mod 11, with 10 mapped to 0.
//
// References:
//   - https://en.wikipedia.org/wiki/Unique_citizenship_number
//   - https://nra.bg/wps/portal/nra/elektronni-uslugi
func checkSumBG(_, normalized ID) bool {
	body := normalized[2:]
	if len(body) == 9 {
		sum := 0
		for i := range 8 {
			sum += int(body[i]-'0') * (i + 1)
		}
		r := sum % 11
		if r == 10 {
			sum = 0
			for i := range 8 {
				sum += int(body[i]-'0') * (i + 3)
			}
			r = sum % 11
			if r == 10 {
				r = 0
			}
		}
		return int(body[8]-'0') == r
	}
	weights := [9]int{2, 4, 8, 5, 10, 9, 7, 3, 6}
	sum := 0
	for i := range 9 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	r := sum % 11
	if r == 10 {
		r = 0
	}
	return int(body[9]-'0') == r
}

// checkSumCY validates a Cypriot VAT ID. 8 digits + 1 letter, where the
// letter is computed by mapping odd-position digits (1-indexed) through
// a fixed table, summing with even-position digits taken at face value,
// taking the result mod 26, and indexing the alphabet (A=0..Z=25).
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumCY(_, normalized ID) bool {
	body := normalized[2:]
	oddMap := [10]int{1, 0, 5, 7, 9, 13, 15, 17, 19, 21}
	sum := 0
	for i := range 8 {
		d := int(body[i] - '0')
		if i%2 == 0 {
			sum += oddMap[d]
		} else {
			sum += d
		}
	}
	return body[8] == byte('A'+sum%26)
}

// checkSumCZ validates an 8-digit Czech legal-entity VAT ID.
// 9- and 10-digit personal identifiers (rodné číslo) follow date-based
// algorithms that are out of scope here — those forms pass shape-only.
//
// Reference: https://www.mfcr.cz/cs/legislativa/legislativni-dokumenty
func checkSumCZ(_, normalized ID) bool {
	body := normalized[2:]
	if len(body) != 8 {
		return true
	}
	weights := [7]int{8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := range 7 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	r := sum % 11
	var check int
	switch r {
	case 0:
		check = 1
	case 1:
		check = 0
	default:
		check = 11 - r
	}
	return int(body[7]-'0') == check
}

// checkSumDK validates a Danish VAT ID — 8 digits, weighted sum
// [2,7,6,5,4,3,2,1] divisible by 11.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumDK(_, normalized ID) bool {
	body := normalized[2:]
	weights := [8]int{2, 7, 6, 5, 4, 3, 2, 1}
	sum := 0
	for i := range 8 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	return sum%11 == 0
}

// checkSumEE validates an Estonian VAT ID — 9 digits, weighted sum
// [3,7,1,3,7,1,3,7,1] divisible by 10.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumEE(_, normalized ID) bool {
	body := normalized[2:]
	weights := [9]int{3, 7, 1, 3, 7, 1, 3, 7, 1}
	sum := 0
	for i := range 9 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	return sum%10 == 0
}

// checkSumEL validates a Greek VAT ID — 9 digits, weighted sum
// [256,128,64,32,16,8,4,2] mod 11 yields the 9th digit (10 → 0).
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumEL(_, normalized ID) bool {
	body := normalized[2:]
	weights := [8]int{256, 128, 64, 32, 16, 8, 4, 2}
	sum := 0
	for i := range 8 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	r := sum % 11
	if r == 10 {
		r = 0
	}
	return int(body[8]-'0') == r
}

// checkSumFI validates a Finnish VAT ID — 8 digits, weighted sum
// [7,9,10,5,8,4,2] mod 11; check digit = (11 - r) for r > 1, 0 for r == 0,
// and r == 1 makes the underlying business ID invalid by construction.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumFI(_, normalized ID) bool {
	body := normalized[2:]
	weights := [7]int{7, 9, 10, 5, 8, 4, 2}
	sum := 0
	for i := range 7 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	r := sum % 11
	if r == 1 {
		return false
	}
	check := 0
	if r != 0 {
		check = 11 - r
	}
	return int(body[7]-'0') == check
}

// checkSumFR validates a French VAT ID — 2 alphanumeric check chars + 9
// digit SIREN. When the 2 leading chars are digits the key is
// (12 + 3 * (SIREN mod 97)) mod 97. The letter-bearing variant uses a
// different algorithm and is not validated here (shape-only).
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumFR(_, normalized ID) bool {
	body := normalized[2:]
	if body[0] > '9' || body[1] > '9' {
		return true
	}
	siren := 0
	for i := 2; i < 11; i++ {
		siren = siren*10 + int(body[i]-'0')
	}
	key := (12 + 3*(siren%97)) % 97
	actual := int(body[0]-'0')*10 + int(body[1]-'0')
	return actual == key
}

// checkSumHR validates a Croatian OIB — 11 digits, ISO 7064 MOD 11-10.
//
// Reference: https://en.wikipedia.org/wiki/Personal_identification_number_(Croatia)
func checkSumHR(_, normalized ID) bool {
	body := normalized[2:]
	p := 10
	for i := range 10 {
		s := (int(body[i]-'0') + p) % 10
		if s == 0 {
			s = 10
		}
		p = (2 * s) % 11
	}
	return int(body[10]-'0') == (11-p)%10
}

// checkSumHU validates a Hungarian VAT ID — 8 digits, weighted sum
// [9,7,3,1,9,7,3] mod 10 yields the check via (10 - r) mod 10.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumHU(_, normalized ID) bool {
	body := normalized[2:]
	weights := [7]int{9, 7, 3, 1, 9, 7, 3}
	sum := 0
	for i := range 7 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	return int(body[7]-'0') == (10-sum%10)%10
}

// checkSumIT validates an Italian VAT ID — 11 digits, plain Luhn
// (positions 2,4,6,8,10 doubled with digit-sum if > 9).
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumIT(_, normalized ID) bool {
	body := normalized[2:]
	sum := 0
	for i := range 11 {
		d := int(body[i] - '0')
		if i%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}

// checkSumLT validates a Lithuanian VAT ID (9 or 12 digits).
// Weights start at 1 and cycle 1..9 across the body; if the first pass
// gives remainder 10, retry with weights shifted by 2.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumLT(_, normalized ID) bool {
	body := normalized[2:]
	n := len(body) - 1
	sum := 0
	for i := range n {
		sum += int(body[i]-'0') * (i%9 + 1)
	}
	r := sum % 11
	if r == 10 {
		sum = 0
		for i := range n {
			sum += int(body[i]-'0') * ((i+2)%9 + 1)
		}
		r = sum % 11
		if r == 10 {
			r = 0
		}
	}
	return int(body[n]-'0') == r
}

// checkSumLU validates a Luxembourg VAT ID — 8 digits, where the first
// 6 digits mod 89 equal the last 2.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumLU(_, normalized ID) bool {
	body := normalized[2:]
	n := 0
	for i := range 6 {
		n = n*10 + int(body[i]-'0')
	}
	check := int(body[6]-'0')*10 + int(body[7]-'0')
	return n%89 == check
}

// checkSumLV validates a Latvian legal-entity VAT ID — 11 digits with
// first digit > 3, weighted mod-11 with check = (3 - r) mod 11.
// Natural-person IDs (first digit ≤ 3) follow a different birthdate-based
// algorithm and pass shape-only here.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumLV(_, normalized ID) bool {
	body := normalized[2:]
	if body[0] <= '3' {
		return true
	}
	weights := [10]int{9, 1, 4, 8, 3, 10, 2, 5, 7, 6}
	sum := 0
	for i := range 10 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	r := sum % 11
	if r == 4 {
		return false
	}
	check := (3 - r + 11) % 11
	if check == 10 {
		return false
	}
	return int(body[10]-'0') == check
}

// checkSumMT validates a Maltese VAT ID — 8 digits, weighted sum
// [3,4,6,7,8,9] of the first 6 digits; check = 37 - (sum mod 37) must
// equal the 2-digit tail.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumMT(_, normalized ID) bool {
	body := normalized[2:]
	weights := [6]int{3, 4, 6, 7, 8, 9}
	sum := 0
	for i := range 6 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	check := 37 - sum%37
	return int(body[6]-'0')*10+int(body[7]-'0') == check
}

// checkSumPL validates a Polish NIP — 10 digits, weighted sum
// [6,5,7,2,3,4,5,6,7] mod 11 yields the check (mod 10 invalid).
//
// Reference: https://en.wikipedia.org/wiki/PESEL#Structure
func checkSumPL(_, normalized ID) bool {
	body := normalized[2:]
	weights := [9]int{6, 5, 7, 2, 3, 4, 5, 6, 7}
	sum := 0
	for i := range 9 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	r := sum % 11
	if r == 10 {
		return false
	}
	return int(body[9]-'0') == r
}

// checkSumPT validates a Portuguese VAT ID — 9 digits, weighted sum
// [9,8,7,6,5,4,3,2] mod 11; check = 11 - r, normalized to 0 for r ≤ 1.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumPT(_, normalized ID) bool {
	body := normalized[2:]
	weights := [8]int{9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := range 8 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	check := 11 - sum%11
	if check >= 10 {
		check = 0
	}
	return int(body[8]-'0') == check
}

// checkSumRO validates a Romanian VAT ID — 2..10 digits, left-padded
// with zeros to 10. Weighted sum [7,5,3,2,1,7,5,3,2] over the first 9
// digits, then check = (sum * 10) mod 11, mapping 10 → 0.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumRO(_, normalized ID) bool {
	body := normalized[2:]
	var padded [10]byte
	for i := range padded {
		padded[i] = '0'
	}
	copy(padded[10-len(body):], body)
	weights := [9]int{7, 5, 3, 2, 1, 7, 5, 3, 2}
	sum := 0
	for i := range 9 {
		sum += int(padded[i]-'0') * weights[i]
	}
	check := (sum * 10) % 11
	if check == 10 {
		check = 0
	}
	return int(padded[9]-'0') == check
}

// checkSumSE validates a Swedish VAT ID. After normalization the ID is
// SE + 10-digit organisation number + literal "01" (regex-enforced).
// The 10-digit org number uses standard Luhn (positions 1,3,5,7,9 doubled).
//
// Reference: https://www.skatteverket.se/foretagochorganisationer/registrering.4.html
func checkSumSE(_, normalized ID) bool {
	body := normalized[2:]
	sum := 0
	for i := range 10 {
		d := int(body[i] - '0')
		if i%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}

// checkSumSI validates a Slovenian VAT ID — 8 digits, weighted sum
// [8,7,6,5,4,3,2] mod 11; check = 11 - r normalized to 0 for r ≤ 1.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumSI(_, normalized ID) bool {
	body := normalized[2:]
	weights := [7]int{8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := range 7 {
		sum += int(body[i]-'0') * weights[i] //#nosec G602 -- index bounded by the country idRegex matched before dispatch
	}
	check := 11 - sum%11
	if check >= 10 {
		check = 0
	}
	return int(body[7]-'0') == check
}

// checkSumSK validates a Slovak VAT ID — 10 digits divisible by 11.
//
// Reference: https://en.wikipedia.org/wiki/VAT_identification_number
func checkSumSK(_, normalized ID) bool {
	body := normalized[2:]
	var n int64
	for i := range 10 {
		n = n*10 + int64(body[i]-'0')
	}
	return n%11 == 0
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
		sum += int(digits[i]-'0') * weights[i] //#nosec G602 -- digits is normalized[2:11], a fixed length-9 slice
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
