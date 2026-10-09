package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAmountParser_Parse pins how AmountParser normalizes user input for
// strfmt: any separator convention becomes a plain point decimal number
// padded to the largest accepted decimal count (2 if none are given),
// and a decimal count that is not accepted is an error instead of being
// guessed, which keeps "1.234" from being read as one euro twenty three
// when only whole and cent amounts are expected.
func TestAmountParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		decimals []int
		str      string
		want     string
		wantErr  bool
	}{
		{name: "German", str: "1.234,56", want: "1234.56"},
		{name: "English", str: "1,234.56", want: "1234.56"},
		{name: "French", str: "1 234,56", want: "1234.56"},
		{name: "Swiss", str: "1'234.56", want: "1234.56"},
		{name: "integer padded to cents", str: "1234", want: "1234.00"},
		{name: "trailing minus", str: "12,34-", want: "-12.34"},
		{name: "any decimals rounded to cents", str: "1,2378", want: "1.24"},

		{name: "cents accepted", decimals: []int{0, 2}, str: "1.234,50", want: "1234.50"},
		{name: "integer accepted", decimals: []int{0, 2}, str: "1234", want: "1234.00"},
		{name: "ambiguous group separator rejected", decimals: []int{0, 2}, str: "1.234", wantErr: true},
		{name: "one decimal rejected", decimals: []int{0, 2}, str: "12,5", wantErr: true},
		{name: "padded to max accepted decimals", decimals: []int{3}, str: "1,500", want: "1.500"},
		{name: "integer only", decimals: []int{0}, str: "1.234.567", want: "1234567"},

		{name: "empty", str: "", wantErr: true},
		{name: "with currency", str: "EUR 12.34", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAmountParser(tt.decimals...).Parse(tt.str)
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
