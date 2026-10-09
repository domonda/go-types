package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCurrencyAmountParser_Parse pins how CurrencyAmountParser normalizes
// user input for strfmt: currency symbols and names become the ISO code
// placed first, any separator convention becomes a plain point decimal
// number padded to the largest accepted decimal count (2 if none are
// given), and unknown currencies or unaccepted decimal counts are errors.
func TestCurrencyAmountParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		decimals []int
		str      string
		want     string
		wantErr  bool
	}{
		{name: "German with symbol after", str: "1.234,56 €", want: "EUR 1234.56"},
		{name: "English with symbol before", str: "$1,234.56", want: "USD 1234.56"},
		{name: "Swiss", str: "CHF 1'234.5", want: "CHF 1234.50"},
		{name: "code after without space", str: "12,34EUR", want: "EUR 12.34"},
		{name: "negative", str: "EUR -12,34", want: "EUR -12.34"},
		{name: "no currency", str: "12", want: "12.00"},

		{name: "integer only", decimals: []int{0}, str: "EUR 12", want: "EUR 12"},
		{name: "cents not accepted", decimals: []int{0}, str: "EUR 12.50", wantErr: true},
		{name: "padded to max accepted decimals", decimals: []int{0, 2}, str: "€ 12", want: "EUR 12.00"},

		{name: "empty", str: "", wantErr: true},
		{name: "unknown currency", str: "XYZ 12.00", wantErr: true},
		{name: "currency only", str: "EUR", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCurrencyAmountParser(tt.decimals...).Parse(tt.str)
			if tt.wantErr {
				assert.Error(t, err, "Parse(%q)", tt.str)
				assert.Empty(t, got, "no normalized value on error")
				return
			}
			assert.NoError(t, err, "Parse(%q)", tt.str)
			assert.Equal(t, tt.want, got, "Parse(%q)", tt.str)
		})
	}
}
