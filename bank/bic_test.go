package bank

import "testing"

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

var invalidBICs = []BIC{
	"BELADEBEXX",
	"bKAUATWW",
	"GIBAATWWX01",
	"GENODEFOJEV",
	"AUTOBANK",    // valid syntax, but not a BIC
	"DIENSTGEBER", // valid syntax, but not a BIC
	"GELISTET",    // valid syntax, but not a BIC
	"FACILITY",    // valid syntax, but not a BIC
}

func Test_BICValid(t *testing.T) {
	for _, bic := range validBICs {
		if !bic.Valid() {
			t.Errorf("Valid BIC not recognized: %s", bic)
		}
	}
	for _, bic := range invalidBICs {
		if bic.Valid() {
			t.Errorf("Invalid BIC not recognized: %s", bic)
		}
	}
}

var bicFinderData = map[string][][]int{
	"BKAUATWw":                  nil,
	"GIBAATWWX01 ":              nil, // TODO, detects too short instead of matching until end
	" XBKAUATWW ":               nil,
	"BKAUATWW":                  [][]int{[]int{0, 8}},
	" BKAUATWW ":                [][]int{[]int{1, 9}},
	"BKAUATWW. BIC:GIBAATWWXXX": [][]int{[]int{0, 8}, []int{14, 14 + 11}},
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
				t.Fatalf("Found BIC '%s' at wrong position in '%s'", bic, str)
			}
			if !bic.Valid() {
				t.Fatalf("Invalid BIC: %s", bic)
			}
		}
	}
}
