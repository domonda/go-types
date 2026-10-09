package strfmt

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/domonda/go-types/money"
)

// TestMoneyFormat_FormatCurrencyAmount checks that CurrencyFirst controls
// the position of the currency code, since locales differ in where the
// currency is written ("EUR 1.234,50" vs "1.234,50 EUR").
func TestMoneyFormat_FormatCurrencyAmount(t *testing.T) {
	amount := money.CurrencyAmount{Currency: money.EUR, Amount: 1234.5}

	tests := []struct {
		name   string
		format MoneyFormat
		want   string
	}{
		{name: "German currency first", format: GermanMoneyFormat(true), want: "EUR 1.234,50"},
		{name: "German currency last", format: GermanMoneyFormat(false), want: "1.234,50 EUR"},
		{name: "English currency first", format: EnglishMoneyFormat(true), want: "EUR 1,234.50"},
		{name: "English currency last", format: EnglishMoneyFormat(false), want: "1,234.50 EUR"},
		{name: "custom precision without thousands separator", format: MoneyFormat{DecimalSep: '.', Precision: 3}, want: "1234.500 EUR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.format.FormatCurrencyAmount(amount))
		})
	}

	// Without a currency there is nothing to place before or after the number
	noCurrency := money.CurrencyAmount{Amount: 1234.5}
	f := GermanMoneyFormat(true)
	assert.Equal(t, "1.234,50", f.FormatCurrencyAmount(noCurrency))
}

// TestFormatterFunc_customTypeFormatter checks that a FormatterFunc
// registered in TypeFormatters takes precedence over the generic
// conversions of the type's kind, which is the purpose of the registry.
func TestFormatterFunc_customTypeFormatter(t *testing.T) {
	type percent float64

	config := NewGermanFormatConfig()
	config.TypeFormatters[reflect.TypeFor[percent]()] = FormatterFunc(func(val reflect.Value, config *FormatConfig) string {
		return Format(val.Float()*100, config) + " %"
	})

	assert.Equal(t, "12,5 %", Format(percent(0.125), config))
	p := percent(0.125)
	assert.Equal(t, "12,5 %", Format(&p, config), "pointer is dereferenced before the formatter lookup")
	assert.Equal(t, "0,125", Format(0.125, config), "other float types are not affected")
}
