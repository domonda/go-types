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
	// TODO improve ES like https://gist.github.com/svschannak/e79892f4fbc56df15bdb5496d0e67b85
	// `^(ES)([A-Z]\d{8})$/`                       //** Spain (National juridical entities)
	// `^(ES)([A-HN-SW]\d{7}[A-J])$/`              //** Spain (Other juridical entities)
	// `^(ES)([0-9YZ]\d{7}[A-Z])$/`                //** Spain (Personal entities type 1)
	// `^(ES)([KLMX]\d{7}[A-Z])$/`                 //** Spain (Personal entities type 2)
	"ES": regexp.MustCompile(`^ES[0-9A-Z]\d{7}[0-9A-Z]$`),
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

func checkSumNO(raw, normalized ID) bool {
	// https://vatstack.com/articles/norway-vat-number-validation
	if len(normalized) < 11 {
		return false
	}
	if d := normalized[2]; d != '8' && d != '9' {
		return false
	}
	// "No." is more often used as prefix for numbers than for Norwegian VAT IDs
	if strings.HasPrefix(string(raw), "No.") || strings.HasPrefix(string(raw), "no.") {
		return false
	}
	// TODO full check-sum implementation
	return true
}
