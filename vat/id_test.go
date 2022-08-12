package vat

import (
	"strings"
	"testing"
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
	"GB123456789012":   "GB123456789012",
	"GB 123456789012":  "GB123456789012",
	"GBGD001":          "GBGD001",
	"GBHA599":          "GBHA599",
	"GB GD001":         "GBGD001",
	"GB HA599":         "GBHA599",
	"IE9S99999L":       "IE9S99999L",
	"IE 9999999LI":     "IE9999999LI",
	"DE 1367 25570":    "DE136725570",
	"NO916634773":      "NO916634773",
	"NO 916634773":     "NO916634773",
	"NO 916634773 MVA": "NO916634773MVA",
	"CHE-123.456.788":  "CHE123456788",
	"CHE123456788":     "CHE123456788",
	"EU372008134":      "EU372008134", // MOSS scheme VAT
	"EST 99600678":     "EST99600678", // Not a real ID, should also be invlaid, see also https://gist.github.com/svschannak/e79892f4fbc56df15bdb5496d0e67b85
}

var invalidVATIDs = []ID{
	"atu12345678",
	"AT/U.12345678",
	" ATU12345678 ",
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
		if invalidID.Valid() {
			t.Errorf("vat.ID should be invalid: %s", string(invalidID))
		}
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
