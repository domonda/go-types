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

	// The currency is written normalized like money.Currency.String
	assert.Equal(t, "EUR 1.234,50", f.FormatCurrencyAmount(money.CurrencyAmount{Currency: "eur", Amount: 1234.5}))
}

// TestMoneyFormat_FormatCurrencyDecimalAmount checks that a
// money.CurrencyDecimalAmount is formatted like a money.CurrencyAmount:
// CurrencyFirst places the currency code, and the amount gets the
// separators and the precision of the MoneyFormat like FormatDecimalAmount.
func TestMoneyFormat_FormatCurrencyDecimalAmount(t *testing.T) {
	amount := money.NewCurrencyDecimalAmount(money.EUR, money.NewDecimalAmount(12345, 1))

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
		{name: "precision -1 keeps the scale", format: MoneyFormat{DecimalSep: ',', Precision: -1}, want: "1234,5 EUR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.format.FormatCurrencyDecimalAmount(amount))
		})
	}

	// Without a currency there is nothing to place before or after the number
	f := GermanMoneyFormat(true)
	assert.Equal(t, "1.234,50", f.FormatCurrencyDecimalAmount(money.CurrencyDecimalAmount{Amount: amount.Amount}))

	// The currency is written normalized like by FormatCurrencyAmount
	assert.Equal(t, "EUR 1.234,50", f.FormatCurrencyDecimalAmount(money.CurrencyDecimalAmount{Currency: "€", Amount: amount.Amount}))
}

// TestMoneyFormat_CurrencyMinorUnits protects that the money formats keep
// every decimal of the minor unit of a currency: a Kuwaiti dinar has 1000
// fils, and the two decimals of the presets rounded KWD 1.234 to "KWD 1.23",
// losing a fils that money.CurrencyAmount.String keeps. The precision is
// only raised, so JPY keeps the two decimals of the presets like String.
// A precision of 0 or 1 rounds below the cents on purpose and is kept for
// every currency, and a precision of -1 keeps writing the shortest form.
// CurrencyDecimalAmount is formatted the same way as CurrencyAmount.
func TestMoneyFormat_CurrencyMinorUnits(t *testing.T) {
	precision0 := MoneyFormat{CurrencyFirst: true, ThousandsSep: ',', DecimalSep: '.', Precision: 0}
	tests := []struct {
		name   string
		format MoneyFormat
		amount money.CurrencyAmount
		want   string
	}{
		{name: "English KWD", format: EnglishMoneyFormat(true), amount: money.CurrencyAmount{Currency: money.KWD, Amount: 1.234}, want: "KWD 1.234"},
		{name: "German KWD", format: GermanMoneyFormat(true), amount: money.CurrencyAmount{Currency: money.KWD, Amount: 1.234}, want: "KWD 1,234"},
		{name: "German KWD rounded to fils", format: GermanMoneyFormat(false), amount: money.CurrencyAmount{Currency: money.KWD, Amount: 1234.5678}, want: "1.234,568 KWD"},
		{name: "currency normalized for its minor units", format: EnglishMoneyFormat(true), amount: money.CurrencyAmount{Currency: "bhd", Amount: 0.5}, want: "BHD 0.500"},
		{name: "JPY keeps the two decimals of the presets", format: EnglishMoneyFormat(true), amount: money.CurrencyAmountJPY(1234.5), want: "JPY 1,234.50"},
		{name: "precision 0 writes whole yen", format: precision0, amount: money.CurrencyAmountJPY(1234.5), want: "JPY 1,235"},
		// A precision below 2 rounds below the cents on purpose,
		// also for a currency with more decimals and without currency
		{name: "precision 0 writes whole euros", format: precision0, amount: money.CurrencyAmountEUR(1234.5), want: "EUR 1,235"},
		{name: "precision 0 writes whole dinars", format: precision0, amount: money.CurrencyAmount{Currency: money.KWD, Amount: 1234.5678}, want: "KWD 1,235"},
		{name: "precision 1 is kept", format: MoneyFormat{DecimalSep: '.', Precision: 1}, amount: money.CurrencyAmount{Currency: money.KWD, Amount: 1.234}, want: "1.2 KWD"},
		{name: "precision 0 without currency", format: precision0, amount: money.CurrencyAmount{Amount: 1234.5}, want: "1,235"},
		// A blank currency has no minor unit to keep
		{name: "precision 3 without currency", format: MoneyFormat{DecimalSep: '.', Precision: 3}, amount: money.CurrencyAmount{Amount: 1.5}, want: "1.500"},
		{name: "precision above the minor units", format: MoneyFormat{DecimalSep: '.', Precision: 4}, amount: money.CurrencyAmount{Currency: money.KWD, Amount: 1.23456}, want: "1.2346 KWD"},
		{name: "precision -1 writes the shortest form", format: MoneyFormat{DecimalSep: ',', Precision: -1}, amount: money.CurrencyAmount{Currency: money.KWD, Amount: 1.5}, want: "1,5 KWD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.format.FormatCurrencyAmount(tt.amount), "FormatCurrencyAmount")
			decimal := money.NewCurrencyDecimalAmount(tt.amount.Currency, money.DecimalAmountFrom(tt.amount.Amount))
			assert.Equal(t, tt.want, tt.format.FormatCurrencyDecimalAmount(decimal), "FormatCurrencyDecimalAmount")
		})
	}
}

// TestMoneyFormat_CurrencyAmountRoundTrip protects that Scan reads back what
// the presets write for a currency amount, for currencies with 0, 2 and 3
// decimals: the German "KWD 1,234" is 1.234 and not rejected as a thousands
// grouping, and no decimal of KWD was rounded away by the format.
func TestMoneyFormat_CurrencyAmountRoundTrip(t *testing.T) {
	amounts := []money.CurrencyAmount{
		{Currency: money.KWD, Amount: 1.234},
		{Currency: money.KWD, Amount: 1234.567},
		{Currency: money.KWD, Amount: -0.001},
		{Currency: money.JPY, Amount: 1234},
		{Currency: money.JPY, Amount: 1234.5},
		{Currency: money.EUR, Amount: 1234.56},
		{Currency: money.EUR, Amount: 0.01},
	}
	germanShortest := NewGermanFormatConfig()
	germanShortest.MoneyAmount.Precision = -1
	configs := []struct {
		name    string
		config  *FormatConfig
		amounts []money.CurrencyAmount
	}{
		{name: "default", config: NewFormatConfig(), amounts: amounts},
		{name: "English", config: NewUSFormatConfig(), amounts: amounts},
		{name: "German", config: NewGermanFormatConfig(), amounts: amounts},
		// The shortest form of 1234.5 has one decimal, which no currency accepts
		{name: "German precision -1", config: germanShortest, amounts: []money.CurrencyAmount{{Currency: money.KWD, Amount: 1.234}, {Currency: money.EUR, Amount: 1234.56}}},
	}
	for _, c := range configs {
		for _, ca := range c.amounts {
			str := Format(ca, c.config)
			var scanned money.CurrencyAmount
			if assert.NoError(t, Scan(reflect.ValueOf(&scanned).Elem(), str, NewScanConfig()), "%s: Scan(%q)", c.name, str) {
				assert.Equal(t, ca, scanned, "%s: Scan(%q)", c.name, str)
			}

			decimal := money.NewCurrencyDecimalAmount(ca.Currency, money.DecimalAmountFrom(ca.Amount))
			str = Format(decimal, c.config)
			var scannedDecimal money.CurrencyDecimalAmount
			if assert.NoError(t, Scan(reflect.ValueOf(&scannedDecimal).Elem(), str, NewScanConfig()), "%s: Scan(%q)", c.name, str) {
				assert.Equal(t, decimal.Currency, scannedDecimal.Currency, "%s: Scan(%q)", c.name, str)
				assert.True(t, decimal.Amount.Equal(scannedDecimal.Amount), "%s: Scan(%q) is %s, want %s", c.name, str, scannedDecimal.Amount, decimal.Amount)
			}
		}
	}
}

// TestMoneyFormat_FormatCentAndDecimalAmount checks that the exact money
// types honor the separators and the precision of a MoneyFormat like
// FormatAmount does for money.Amount, so a custom format applies to all
// money types alike. A negative precision keeps the exact digits.
// A MoneyFormat without DecimalSep pads with a point like FormatAmount,
// the zero rune must not end up in the text where it is invisible
// but breaks CSV files and string comparisons.
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
		{name: "zero DecimalSep pads with a point", format: MoneyFormat{Precision: 2}, cents: 500, decimal: money.NewDecimalAmount(5, 0), want: "5.00"},
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
