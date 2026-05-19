package vat

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var validVATIDs = map[string]string{
	// "ATU Nr. 1-022 3 006": "ATU10223006", // Real world encounter, but we'll probably never support this mess
	"atu10223006":      "ATU10223006", // Example
	"ATU10223006":      "ATU10223006", // Example
	"ATU67554568":      "ATU67554568", // Real but inactive by now
	"ATU68765099":      "ATU68765099", // Real
	"ATU46983509":      "ATU46983509", // Real
	"ATU65785527":      "ATU65785527", // Real
	"ESW0184081H":      "ESW0184081H", // Real: Amazon EU Sarl Sucursal EN España
	"DE111111125":      "DE111111125", // Example
	"LT347776113":      "LT347776113",
	"DE 167015661":     "DE167015661",
	"ATU 10223006":     "ATU10223006",
	"AT U 10223006":    "ATU10223006",
	"at U 10223006":    "ATU10223006",
	"ATU.10223006":     "ATU10223006",
	"GB123456782":      "GB123456782", // 9-digit VRN, mod-97 valid
	"GB123456782012":   "GB123456782012",
	"GB 123456782012":  "GB123456782012",
	"GBGD001":          "GBGD001",
	"GBHA599":          "GBHA599",
	"GB GD001":         "GBGD001",
	"GB HA599":         "GBHA599",
	"IE9S99999L":       "IE9S99999L",
	"IE 9999999LI":     "IE9999999LI",
	// Belgium — first digit 0 or 1, mod-97 control on last 2 digits.
	"BE0776091951": "BE0776091951",
	"BE1234567894": "BE1234567894",
	// Bulgaria — 9-digit EIK and 10-digit EGN both verified.
	"BG175074752":  "BG175074752",
	"BG7523169263": "BG7523169263",
	// Cyprus — 8 digits + control letter.
	"CY10001049T": "CY10001049T",
	// Czech Republic — 8-digit legal entity (mod-11).
	"CZ00177041": "CZ00177041",
	// Denmark — mod-11 weighted.
	"DK25313763": "DK25313763",
	// Estonia — mod-10 weighted.
	"EE100094916": "EE100094916",
	// Greece — mod-11 with powers-of-two weights.
	"EL094019245": "EL094019245",
	// Finland — mod-11 weighted.
	"FI09853608": "FI09853608",
	// France — 2 digit key + 9-digit SIREN, mod-97 derived.
	"FR40303265045": "FR40303265045",
	// Croatia — ISO 7064 MOD 11-10.
	"HR38192148118": "HR38192148118",
	// Hungary — mod-10 weighted.
	"HU12892312": "HU12892312",
	// Italy — Luhn over 11 digits.
	"IT00488410010": "IT00488410010",
	// Lithuania — 9-digit and 12-digit forms.
	"LT119511515":    "LT119511515",
	"LT290061371314": "LT290061371314",
	// Luxembourg — mod-89.
	"LU10000356": "LU10000356",
	// Latvia — legal entity (first digit > 3), mod-11 weighted.
	"LV40003521600": "LV40003521600",
	// Malta — mod-37.
	"MT15121333": "MT15121333",
	// Netherlands — mod-11 weighted, then "B" + 2-digit branch.
	"NL005033019B01": "NL005033019B01",
	// Poland — mod-11 weighted.
	"PL5260250274": "PL5260250274",
	// Portugal — mod-11 weighted.
	"PT502757191": "PT502757191",
	// Romania — variable length, left-padded mod-11.
	"RO18158683": "RO18158683",
	// Sweden — Luhn on the 10-digit org number; literal "01" branch.
	"SE556677889901": "SE556677889901",
	// Slovenia — mod-11 weighted.
	"SI80267432": "SI80267432",
	// Slovakia — full number divisible by 11.
	"SK2020317068": "SK2020317068",
	// Northern Ireland — same shape and mod-97 checksum as GB.
	"XI123456782":    "XI123456782",
	"XI123456782012": "XI123456782012",
	"XIGD001":        "XIGD001",
	"XIHA599":        "XIHA599",
	// Iceland VSK — 5 or 6 digits, no checksum.
	"IS12345":  "IS12345",
	"IS123456": "IS123456",
	// Liechtenstein — 5 digits, no checksum.
	"LI12345": "LI12345",
	// San Marino — 5 digits, no checksum.
	"SM12345": "SM12345",
	"DE 1367 25570":    "DE136725570",
	"NO916634773":      "NO916634773",
	"NO 916634773":     "NO916634773",
	"NO 916634773 MVA": "NO916634773MVA",
	"NO977074010MVA":   "NO977074010MVA",
	"NO 977074010":     "NO977074010",
	"NO 977074010 MVA": "NO977074010MVA",
	"CHE-123.456.788":  "CHE123456788",
	"CHE123456788":     "CHE123456788",
	"EU372008134":      "EU372008134", // MOSS scheme VAT
	// Spain — coverage of all four sub-formats.
	"ES12345678Z":  "ES12345678Z",  // DNI / NIF, mod-23 letter = Z
	"ESY1234567X":  "ESY1234567X",  // NIE (Y → 1)
	"ESZ0000000M":  "ESZ0000000M",  // NIE (Z → 2)
	"ESX1234567L":  "ESX1234567L",  // NIE (X → 0, historic form)
	"ESA82018474":  "ESA82018474",  // CIF requiring numeric check
	"ESP0000000J":  "ESP0000000J",  // CIF requiring letter check
}

var invalidVATIDs = []ID{
	"atu12345678",
	"AT/U.12345678",
	" ATU12345678 ",
	"No. 62-1764389",
	"No.821764389",
	// Spain — must reject:
	"EST99600678",  // T is not a valid CIF entity prefix
	"ES12345678A",  // wrong DNI check letter (should be Z)
	"ESX1234567A",  // wrong NIE check letter (should be L)
	"ESA82018470",  // wrong CIF numeric check (should be 4)
	"ESP00000000",  // P requires a letter check, not a digit
	"ESA0000000J",  // A requires a digit check, not a letter
	// Norway — must reject (mod-11 checksum fails):
	"NO916634770", // valid format, last digit broken (3 → 0)
	// GB — regex precedence bug: each alternative must be anchored to ^GB...$
	"GBXGD123",    // 'X' between GB and GD must not slip through
	"GBHA999Y",    // trailing 'Y' after HA\d{3} must not slip through
	"GB12345678",  // 8 digits is not a valid GB length
	"GB123456789", // 9-digit shape OK but mod-97 checksum fails
	// Northern Ireland — shares GB regex + checksum.
	"XI123456789", // checksum fails
	"XIXGD123",    // precedence bug guard
	"XIHA999Y",    // precedence bug guard
	// Iceland — 4 or 7 digits are out of range.
	"IS1234",
	"IS1234567",
	// Liechtenstein — must be exactly 5 digits.
	"LI1234",
	"LI123456",
	// San Marino — must be exactly 5 digits.
	"SM1234",
	"SM123456",
	// IE — regex precedence bug: second alternative must be anchored to ^IE
	"IEX9999999LI", // unexpected 'X' before \d{7}[A-W][A-I] must not slip through
	// Belgium — first digit must be 0 or 1.
	"BE2345678901", // starts with 2
	"BE9999999999", // starts with 9
	"BE123456789",  // only 9 digits
	"BE0776091950", // valid format, mod-97 control broken
	// Bulgaria — 9-digit EIK and 10-digit EGN checksums.
	"BG175074750",  // EIK: last digit broken
	"BG7523169260", // EGN: last digit broken
	// Cyprus — control letter must match.
	"CY10001049X", // wrong control letter (should be T)
	// Czech Republic — 8-digit mod-11.
	"CZ00177040",
	// Denmark — mod-11.
	"DK25313764",
	// Estonia — mod-10.
	"EE100094917",
	// Greece — mod-11.
	"EL094019246",
	// Finland — mod-11.
	"FI09853609",
	// France — mod-97 SIREN key.
	"FR41303265045",
	// Croatia — ISO 7064 MOD 11-10.
	"HR38192148119",
	// Hungary — exactly 8 digits, not 9.
	"HU123456789",
	"HU12892313", // valid format, mod-10 control broken
	// Italy — Luhn.
	"IT00488410011",
	// Lithuania — first-pass and fallback both fail.
	"LT119511516",
	"LT290061371315",
	// Luxembourg — mod-89.
	"LU10000357",
	// Latvia — legal entity mod-11.
	"LV40003521601",
	// Malta — mod-37.
	"MT15121334",
	// Netherlands — mod-11 (after stripping B-suffix).
	"NL005033018B01",
	// Poland — mod-11.
	"PL5260250275",
	// Portugal — mod-11.
	"PT502757192",
	// Romania — mod-11 over left-padded body.
	"RO18158684",
	// Sweden — last two digits must be "01"; Luhn over the org number.
	"SE123456789002",  // ends in 02
	"SE123456789010",  // ends in 10
	"SE12345678901",   // only 11 digits
	"SE556677889001",  // valid format, Luhn fails (org digit moved)
	// Slovenia — mod-11.
	"SI80267433",
	// Slovakia — full number divisible by 11.
	"SK2020317069",
	// Switzerland — no alternative dotted format after normalization.
	"CHE12345678",   // 8 digits
	"CHE1234567890", // 10 digits
}

func Test_NormalizeVATID(t *testing.T) {
	for testID, refID := range validVATIDs {
		result, err := NormalizeVATID(testID)
		if err != nil {
			t.Errorf("NormalizeVATID(%s): %s", string(testID), err)
		} else if string(result) != refID {
			t.Errorf("NormalizeVATID(%s): %s != %s", string(testID), string(result), refID)
		}
	}
}

func Test_VATIDValid(t *testing.T) {
	for _, invalidID := range invalidVATIDs {
		assert.Falsef(t, invalidID.Valid(), "vat.ID should be invalid: %s, Regex uses NormalizedUnchecked: %s", string(invalidID), string(invalidID /*.NormalizedUnchecked()*/))
	}
}

var vatidTestIndices = map[string][][]int{
	"":                         nil,
	"ATU10223006":              {{0, 11}},
	"  ATU10223006":            {{2, 13}},
	"UID: ATU10223006":         {{5, 16}},
	"UID AT U 10223006":        {{4, 17}},
	"UID:AT U 10223006":        {{4, 17}},
	"ATU10223006 ":             {{0, 11}},
	"ATU10223006 ATU 10223006": {{0, 11}, {12, 24}},
	" AT U 10223006 ATU10223006 ATU 10223006 ": {{1, 14}, {15, 26}, {27, 39}},
	"USt-IdNr. DE 136725570":                   {{10, 22}},
}

func Test_VATIDFinder_FindAllIndex(t *testing.T) {
	for str, refIndices := range vatidTestIndices {
		indices := IDFinder.FindAllIndex([]byte(str), -1)
		if len(indices) != len(refIndices) {
			var words []string
			for i := range indices {
				words = append(words, "'"+str[indices[i][0]:indices[i][1]]+"'")
			}
			t.Errorf("VATIDFinder.FindAllIndex('%s') expected %d words but got %d: %s", str, len(refIndices), len(indices), strings.Join(words, " "))
		} else {
			for i := range indices {
				if indices[i][0] != refIndices[i][0] || indices[i][1] != refIndices[i][1] {
					// t.Error(i, indices[i], refIndices[i], len(str))
					t.Errorf("VATIDFinder.FindAllIndex('%s') word %d expected %v '%s' but got %v '%s'", str, i, refIndices[i], str[refIndices[i][0]:refIndices[i][1]], indices[i], str[indices[i][0]:indices[i][1]])
				}
			}
		}
	}
}
