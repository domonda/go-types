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
