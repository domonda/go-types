package bank

import (
	"github.com/domonda/go-types/country"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var validBICs = []BIC{
	"BKAUATWW",
	"GIBAATWWXXX",
	"BELADEBEXXX",
	"RBOSGGSX",
	"CHASGB2LXXX",
	"RZTIAT22263",
	"BCEELULL",
	"MARKDEFF",
	"MARKDEFFXXX",
	"GENODEF1JEV",
	"UBSWCHZH80A",
	"CEDELULLXXX",
	"HELADEF1RRS",
	"HELADEFF",
	"GENODEF1S04",
	"GENODE61MA2",
	"DEUTDEDBMAN",
	"SOLADES1HDB",
}

var validBICsWithSpaces = []BIC{
	"BKAU AT WW",
	"GIBAATWW XXX",
	" BELADEBEXXX ",
	"RBOSGGSX   ",
}

var invalidBICs = []BIC{
	"BELADEBEXX",
	"bKAUATWW",
	"GIBAATWWX01",
	"GENODEFOJEV",
	"AMTSGERICHT", // valid syntax, but not a BIC
	"AUTOBANK",    // valid syntax, but not a BIC
	"DEUTSCHLAND", // valid syntax, but not a BIC
	"DIENSTGEBER", // valid syntax, but not a BIC
	"DOCUMENT",    // valid syntax, but not a BIC
	"DOKUMENT",    // valid syntax, but not a BIC
	"FACILITY",    // valid syntax, but not a BIC
	"GELISTET",    // valid syntax, but not a BIC
	"GESAMTNETTO", // valid syntax, but not a BIC
}

func Test_BICValid(t *testing.T) {
	for _, bic := range validBICs {
		if !bic.Valid() {
			t.Errorf("Valid BIC not recognized: %s", string(bic))
		}
	}
	for _, bic := range invalidBICs {
		if bic.Valid() {
			t.Errorf("Invalid BIC not recognized: %s", string(bic))
		}
	}
}

var bicFinderData = map[string][][]int{
	"BKAUATWw":                  nil,
	"GIBAATWWX01 ":              nil, // TODO, detects too short instead of matching until end
	" XBKAUATWW ":               nil,
	"BKAUATWW":                  {[]int{0, 8}},
	" BKAUATWW ":                {[]int{1, 9}},
	"BKAUATWW. BIC:GIBAATWWXXX": {[]int{0, 8}, []int{14, 14 + 11}},
}

func Test_bicFinder(t *testing.T) {
	for str, allIndices := range bicFinderData {
		allResult := BICFinder.FindAllIndex([]byte(str), -1)
		if len(allResult) != len(allIndices) {
			// bic := str[allResult[0][0]:allResult[0][1]]
			// fmt.Println(bic)
			t.Fatalf("Found %d BICs in '%s', but expected %d", len(allResult), str, len(allIndices))
		}
		for i := range allIndices {
			indices := allIndices[i]
			result := allResult[i]
			if len(result) != 2 {
				t.Fatalf("Did not find BIC in '%s'", str)
			}
			bic := BIC(str[result[0]:result[1]])
			if result[0] != indices[0] || result[1] != indices[1] {
				t.Fatalf("Found BIC '%s' at wrong position in '%s'", string(bic), str)
			}
			if !bic.Valid() {
				t.Fatalf("Invalid BIC: %s", string(bic))
			}
		}
	}
}

func TestBIC_Normalized(t *testing.T) {
	for _, bic := range append(validBICs, validBICsWithSpaces...) {
		normalized, err := bic.Normalized()
		if err != nil {
			t.Errorf("Error normalizing BIC: %s", err)
		}
		expected := BIC(strings.ReplaceAll(string(bic), " ", ""))
		if len(expected) == 8 {
			expected += "XXX"
		}
		require.Equalf(t, expected, normalized, "Normalized BIC %q", bic)
	}
	for _, bic := range invalidBICs {
		norm, err := bic.Normalized()
		require.Error(t, err, "Normalized invalid BIC error expected")
		require.Equal(t, bic, norm, "Normalized invalid BIC returned unchanged")
	}
}

func TestBIC_Parse(t *testing.T) {
	tests := []struct {
		name            string
		input           BIC
		wantIsValid     bool
		wantBankCode    string
		wantCountryCode country.Code
		wantBranchCode  string
	}{
		{
			name:        "Invalid length",
			input:       "ABCDE",
			wantIsValid: false,
		},
		{
			name:        "Regexp no match",
			input:       "12345678",
			wantIsValid: false,
		},
		{
			name:        "Country code invalid",
			input:       "DEUTZZFF",
			wantIsValid: false,
		},
		{
			name:        "In falseBICs",
			input:       "FAKEBIC12",
			wantIsValid: false,
		},
		{
			name:            "Valid BIC 11 chars",
			input:           "DEUTDEFF500",
			wantIsValid:     true,
			wantBankCode:    "DEUT",
			wantCountryCode: "DE",
			wantBranchCode:  "500",
		},
		{
			name:            "Valid BIC 8 chars",
			input:           "DEUTDEFF",
			wantIsValid:     true,
			wantBankCode:    "DEUT",
			wantCountryCode: "DE",
			wantBranchCode:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bankCode, countryCode, branchCode, isValid := tc.input.Parse()
			require.Equal(t, tc.wantIsValid, isValid)

			assert.Equal(t, tc.wantBankCode, bankCode)
			assert.Equal(t, tc.wantCountryCode, countryCode)
			assert.Equal(t, tc.wantBranchCode, branchCode)
		})
	}
}
