package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StringIsAmount(t *testing.T) {
	for str := range amountTable2Decimals {
		if !StringIsAmount(str, false) {
			t.Errorf("String not detected as amount: '%s'", str)
		}
	}
	for _, str := range invalidAmounts {
		if StringIsAmount(str, false) {
			t.Errorf("Invalid string detected as amount: '%s'", str)
		}
	}
}

// Test_StringIsAmount_acceptInt pins that bare integers like years,
// invoice or article numbers are only detected as amounts when the
// caller explicitly accepts them.
func Test_StringIsAmount_acceptInt(t *testing.T) {
	for _, str := range []string{"1234", "-12", "0"} {
		assert.False(t, StringIsAmount(str, false), "StringIsAmount(%q, false)", str)
		assert.True(t, StringIsAmount(str, true), "StringIsAmount(%q, true)", str)
	}
	// Not an integer and not an amount with two decimals
	for _, str := range []string{"12.5", "1,5", "12.345", "1 234"} {
		assert.False(t, StringIsAmount(str, true), "StringIsAmount(%q, true)", str)
	}
}

// Test_StringIsAmount_nonFiniteTokensMustBeWholeString pins that the
// NaN and Inf tokens are only accepted as the whole string. AmountFinder
// tests every word of a text with the same regex, so a token matched
// anywhere inside a word turned words like "Information" into amounts
// that then fail to parse or parse as infinity.
func Test_StringIsAmount_nonFiniteTokensMustBeWholeString(t *testing.T) {
	for _, str := range []string{"NaN", "Inf", "+Inf", "-Inf"} {
		assert.True(t, StringIsAmount(str, false), "StringIsAmount(%q, false)", str)
	}
	for _, str := range []string{"Information", "Infinity", "Infos", "NaNa", "ANaN", "xInf", "-Info", "Inf.", "1,00Inf"} {
		assert.False(t, StringIsAmount(str, false), "StringIsAmount(%q, false)", str)
		assert.False(t, StringIsAmount(str, true), "StringIsAmount(%q, true)", str)
	}
	assert.Nil(t, AmountFinder.FindAllIndex([]byte("Information about the Infrastructure"), -1))
}

// TestAmountFinder_FindAllIndex_limit pins the n argument of the
// types.Finder contract: n < 0 returns all matches, n == 0 none and
// n > 0 at most n, in text order. Callers pass n > 0 to stop after the
// first amounts, so returning more than asked for breaks them.
func TestAmountFinder_FindAllIndex_limit(t *testing.T) {
	text := []byte("1,00 2,00 3,00")
	all := [][]int{{0, 4}, {5, 9}, {10, 14}}

	assert.Equal(t, all, AmountFinder.FindAllIndex(text, -1), "n < 0")
	assert.Nil(t, AmountFinder.FindAllIndex(text, 0), "n == 0")
	assert.Equal(t, all[:1], AmountFinder.FindAllIndex(text, 1), "n == 1")
	assert.Equal(t, all[:2], AmountFinder.FindAllIndex(text, 2), "n == 2")
	assert.Equal(t, all, AmountFinder.FindAllIndex(text, 3), "n == number of matches")
	assert.Equal(t, all, AmountFinder.FindAllIndex(text, 10), "n > number of matches")
}

// TestAmountFinder_FindAllIndex pins that amounts are found in invoice
// text with both decimal conventions, separated by whitespace or colons
// and with trailing punctuation trimmed, while integers like years
// are not reported as amounts.
func TestAmountFinder_FindAllIndex(t *testing.T) {
	tests := []struct {
		text string
		want []string
	}{
		{
			text: "Netto: 100,00 USt: 20,00 Brutto: 120,00.",
			want: []string{"100,00", "20,00", "120,00"},
		},
		{
			text: "Summe 1.234,56; Datum 2024\ntotal:99.99, Rabatt -5.00",
			want: []string{"1.234,56", "99.99", "-5.00"},
		},
		{
			text: "Amount due 1,234,567.89 by 2024-05-01",
			want: []string{"1,234,567.89"},
		},
		{
			text: "no amounts in here 2024",
			want: nil,
		},
		{
			text: "",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			var found []string
			for _, pos := range AmountFinder.FindAllIndex([]byte(tt.text), -1) {
				found = append(found, tt.text[pos[0]:pos[1]])
			}
			assert.Equal(t, tt.want, found)
		})
	}
}
