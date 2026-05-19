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
	// Belgium — first digit must be 0 or 1.
	"BE0123456789": "BE0123456789",
	"BE1234567890": "BE1234567890",
	// Hungary — exactly 8 digits.
	"HU12345678": "HU12345678",
	// Sweden — 10-digit org number followed by literal "01".
	"SE123456789001": "SE123456789001",
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
	// Hungary — exactly 8 digits, not 9.
	"HU123456789",
	// Sweden — last two digits must be "01".
	"SE123456789002", // ends in 02
	"SE123456789010", // ends in 10
	"SE12345678901",  // only 11 digits
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
