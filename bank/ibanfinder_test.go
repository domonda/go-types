package bank

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IBANFinder(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []IBAN
	}{
		{
			name: "empty string",
			text: "",
			want: nil,
		},
		{
			name: "no IBAN in prose",
			text: "There is no bank account number anywhere in this sentence.",
			want: nil,
		},
		{
			name: "shorter than IBANMinLength",
			text: "AT12345",
			want: nil,
		},
		{
			name: "lowercase is not detected",
			text: "at611904300234573201",
			want: nil,
		},
		{
			name: "invalid checksum is rejected",
			text: "AT611904300234573200",
			want: nil,
		},
		{
			name: "spaces inside the IBAN prevent detection",
			text: "AT61 1904 3002 3457 3201",
			want: nil,
		},
		{
			name: "single IBAN only",
			text: "AT611904300234573201",
			want: []IBAN{"AT611904300234573201"},
		},
		{
			name: "single IBAN surrounded by text",
			text: "Please pay to IBAN AT611904300234573201 by Friday.",
			want: []IBAN{"AT611904300234573201"},
		},
		{
			name: "multiple IBANs across a longer text",
			text: "Transfer the deposit from AT611904300234573201 to the supplier " +
				"account DE89370400440532013000, then forward the remainder to " +
				"the Norwegian account NO9386011117947. The Belgian reserve " +
				"account BE62510007547061 must stay untouched, and " +
				"NL39RABO0300065264 is only used for refunds.",
			want: []IBAN{
				"AT611904300234573201",
				"DE89370400440532013000",
				"NO9386011117947",
				"BE62510007547061",
				"NL39RABO0300065264",
			},
		},
		{
			name: "IBANs separated only by punctuation",
			text: "AT611904300234573201;DE89370400440532013000,NO9386011117947",
			want: []IBAN{
				"AT611904300234573201",
				"DE89370400440532013000",
				"NO9386011117947",
			},
		},
		{
			name: "IBAN with alphanumeric BBAN inside text",
			text: "International example FR1420041010050500013M02606 from France.",
			want: []IBAN{"FR1420041010050500013M02606"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spans := IBANFinder.FindAllIndex([]byte(tc.text), -1)

			var got []IBAN
			prevEnd := 0
			for i, span := range spans {
				require.Lenf(t, span, 2, "span %d must be a [start, end] pair", i)
				require.GreaterOrEqualf(t, span[0], prevEnd, "span %d must not overlap the previous match", i)
				require.Lessf(t, span[0], span[1], "span %d must be non-empty", i)
				require.LessOrEqualf(t, span[1], len(tc.text), "span %d must stay within the text", i)
				prevEnd = span[1]

				iban := IBAN(tc.text[span[0]:span[1]])
				require.Truef(t, iban.Valid(), "found IBAN %q must be valid", iban)
				got = append(got, iban)
			}
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_IBANFinder_N(t *testing.T) {
	// text contains exactly 3 IBANs
	text := []byte("First AT611904300234573201, second DE89370400440532013000, " +
		"third NO9386011117947.")

	require.Len(t, IBANFinder.FindAllIndex(text, -1), 3, "n < 0 returns all matches")
	require.Empty(t, IBANFinder.FindAllIndex(text, 0), "n == 0 returns no matches")
	require.Len(t, IBANFinder.FindAllIndex(text, 1), 1, "n == 1 returns at most 1 match")
	require.Len(t, IBANFinder.FindAllIndex(text, 2), 2, "n == 2 returns at most 2 matches")
	require.Len(t, IBANFinder.FindAllIndex(text, 99), 3, "n larger than available returns all matches")
}
