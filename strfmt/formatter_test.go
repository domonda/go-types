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

// TestMoneyFormat_FormatCentAndDecimalAmount checks that the exact money
// types honor the separators and the precision of a MoneyFormat like
// FormatAmount does for money.Amount, so a custom format applies to all
// money types alike. A negative precision keeps the exact digits.
func TestMoneyFormat_FormatCentAndDecimalAmount(t *testing.T) {
	tests := []struct {
		name    string
		format  MoneyFormat
		cents   money.CentAmount
		decimal money.DecimalAmount
		want    string
	}{
		{name: "German", format: GermanMoneyFormat(true), cents: 123456, decimal: money.NewDecimalAmount(123456, 2), want: "1.234,56"},
		{name: "English", format: EnglishMoneyFormat(true), cents: -123456, decimal: money.NewDecimalAmount(-123456, 2), want: "-1,234.56"},
		{name: "precision 3 pads", format: MoneyFormat{DecimalSep: '.', Precision: 3}, cents: 150, decimal: money.NewDecimalAmount(15, 1), want: "1.500"},
		{name: "precision 0 rounds half away from zero", format: MoneyFormat{DecimalSep: ',', Precision: 0}, cents: -250, decimal: money.NewDecimalAmount(-25, 1), want: "-3"},
		{name: "precision -1 keeps the cents", format: MoneyFormat{DecimalSep: '.', Precision: -1}, cents: 150, decimal: money.NewDecimalAmount(150, 2), want: "1.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.format.FormatCentAmount(tt.cents), "FormatCentAmount")
			assert.Equal(t, tt.want, tt.format.FormatDecimalAmount(tt.decimal), "FormatDecimalAmount")
		})
	}

	// Beyond the DecimalAmount range a CentAmount still formats exactly
	f := EnglishMoneyFormat(true)
	assert.Equal(t, "92,233,720,368,547,758.07", f.FormatCentAmount(money.MaxCentAmount))
	// A finite DecimalAmount never formats as Inf, even if its value
	// can't be represented with more decimals
	big := money.NewDecimalAmount(200_000_000_000_000_000, 0)
	assert.Equal(t, "200,000,000,000,000,000.00", f.FormatDecimalAmount(big))
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
