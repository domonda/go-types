package bank

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var bicFinderData = map[string][][]int{
	"":                          nil,
	"no bic in this text":       nil,
	"BKAUATWw":                  nil,      // lowercase 'w' is not a valid BIC character
	"GIBAATWWX01":               nil,      // "X01" is not a valid branch code, so this is not a BIC
	"GIBAATWWX01 ":              nil,      // a trailing separator does not rescue an invalid code
	"GIBAATWW X01":              {{0, 8}}, // only the standalone 8-char BIC is matched
	" XBKAUATWW ":               nil,      // BKAUATWW is not separator-bounded on the left
	"AMTSGERICHT":               nil,      // syntactically a BIC but listed in falseBICs
	"BKAUATWW":                  {{0, 8}},
	" BKAUATWW ":                {{1, 9}},
	"DEUTDEDBMAN":               {{0, 11}},
	"see BELADEBEXXX, ok":       {{4, 15}},
	"BKAUATWW. BIC:GIBAATWWXXX": {{0, 8}, {14, 14 + 11}},
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

func Test_BICFinder_LongText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []BIC
	}{
		{
			name: "no BIC in prose",
			text: "This paragraph mentions banks and money but contains no code at all.",
			want: nil,
		},
		{
			name: "multiple BICs across a longer text",
			text: "For the Austrian vendor send the SEPA transfer to BKAUATWW, " +
				"while the German subsidiary is paid via MARKDEFF instead. " +
				"Route the customer refund through BELADEBEXXX, and keep " +
				"DEUTDEDBMAN reserved for emergency settlements only.",
			want: []BIC{"BKAUATWW", "MARKDEFF", "BELADEBEXXX", "DEUTDEDBMAN"},
		},
		{
			name: "BIC at start and end of text",
			text: "GIBAATWWXXX is the opening account and the closing one is BCEELULL",
			want: []BIC{"GIBAATWWXXX", "BCEELULL"},
		},
		{
			name: "BIC-like words that are not BICs are ignored",
			text: "The AMTSGERICHT and the AUTOBANK both look like a BIC but only " +
				"HELADEFF is a real one.",
			want: []BIC{"HELADEFF"},
		},
		{
			name: "BICs separated only by punctuation",
			text: "BKAUATWW;MARKDEFF/BELADEBEXXX",
			want: []BIC{"BKAUATWW", "MARKDEFF", "BELADEBEXXX"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spans := BICFinder.FindAllIndex([]byte(tc.text), -1)

			var got []BIC
			prevEnd := 0
			for i, span := range spans {
				require.Lenf(t, span, 2, "span %d must be a [start, end] pair", i)
				require.GreaterOrEqualf(t, span[0], prevEnd, "span %d must not overlap the previous match", i)
				require.Lessf(t, span[0], span[1], "span %d must be non-empty", i)
				require.LessOrEqualf(t, span[1], len(tc.text), "span %d must stay within the text", i)
				prevEnd = span[1]

				bic := BIC(tc.text[span[0]:span[1]])
				require.Truef(t, bic.Valid(), "found BIC %q must be valid", bic)
				got = append(got, bic)
			}
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_BICFinder_N(t *testing.T) {
	// AMTSGERICHT is a regex candidate but filtered out (falseBICs);
	// the two valid BICs come after it in the text.
	text := []byte("AMTSGERICHT, then BKAUATWW, then MARKDEFF.")

	require.Len(t, BICFinder.FindAllIndex(text, -1), 2, "n < 0 returns all valid BICs")
	require.Empty(t, BICFinder.FindAllIndex(text, 0), "n == 0 returns no matches")
	require.Len(t, BICFinder.FindAllIndex(text, 99), 2, "n larger than available returns all matches")

	// n must limit valid BICs, not raw regex candidates: even though
	// AMTSGERICHT is the first regex match, n == 1 must still yield BKAUATWW.
	got := BICFinder.FindAllIndex(text, 1)
	require.Len(t, got, 1, "n == 1 returns at most 1 valid BIC")
	require.Equal(t, BIC("BKAUATWW"), BIC(text[got[0][0]:got[0][1]]), "n == 1 skips filtered-out candidates")
}
