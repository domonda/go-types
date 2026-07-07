package strfmt

import (
	"reflect"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/money"
)

// Formatter converts a reflect.Value to a string using the provided FormatConfig.
// It is the core extension point for registering custom per-type formatting
// logic in FormatConfig.TypeFormatters.
type Formatter interface {
	FormatValue(val reflect.Value, config *FormatConfig) string
}

// FormatterFunc is a function adapter that implements the Formatter interface,
// allowing an ordinary function to be used wherever a Formatter is expected.
type FormatterFunc func(val reflect.Value, config *FormatConfig) string

// FormatValue calls f(val, config), implementing the Formatter interface.
func (f FormatterFunc) FormatValue(val reflect.Value, config *FormatConfig) string {
	return f(val, config)
}

// MoneyFormat specifies how monetary amounts are serialized to strings,
// including digit grouping, decimal notation, precision, and whether
// the currency symbol appears before or after the numeric value.
type MoneyFormat struct {
	CurrencyFirst bool
	ThousandsSep  rune
	DecimalSep    rune
	Precision     int
}

// FormatAmount formats a monetary amount as a string using the
// thousands separator, decimal separator, and precision defined in format.
func (format *MoneyFormat) FormatAmount(amount money.Amount) string {
	return amount.FormatSep(format.ThousandsSep, format.DecimalSep, format.Precision)
}

// FormatCurrencyAmount formats a currency-tagged monetary amount as a string,
// placing the currency code before or after the number according to
// format.CurrencyFirst, and applying the configured separators and precision.
func (format *MoneyFormat) FormatCurrencyAmount(currencyAmount money.CurrencyAmount) string {
	return currencyAmount.FormatSep(format.CurrencyFirst, format.ThousandsSep, format.DecimalSep, format.Precision)
}

// EnglishFloatFormat returns a float.FormatDef using a dot as the decimal
// separator, no thousands separator, and the given precision.
// Pass -1 as precision to use the minimum number of digits necessary.
func EnglishFloatFormat(precision int) float.FormatDef {
	return float.FormatDef{
		ThousandsSep: 0,
		DecimalSep:   '.',
		Precision:    precision,
		PadPrecision: false,
	}
}

// GermanFloatFormat returns a float.FormatDef using a comma as the decimal
// separator, no thousands separator, and the given precision.
// Pass -1 as precision to use the minimum number of digits necessary.
func GermanFloatFormat(precision int) float.FormatDef {
	return float.FormatDef{
		ThousandsSep: 0,
		DecimalSep:   ',',
		Precision:    precision,
		PadPrecision: false,
	}
}

// EnglishMoneyFormat returns a MoneyFormat with English conventions:
// comma thousands separator, dot decimal separator, and 2 decimal places.
// currencyFirst controls whether the currency code precedes the amount.
func EnglishMoneyFormat(currencyFirst bool) MoneyFormat {
	return MoneyFormat{
		CurrencyFirst: currencyFirst,
		ThousandsSep:  ',',
		DecimalSep:    '.',
		Precision:     2,
	}
}

// GermanMoneyFormat returns a MoneyFormat with German conventions:
// dot thousands separator, comma decimal separator, and 2 decimal places.
// currencyFirst controls whether the currency code precedes the amount.
func GermanMoneyFormat(currencyFirst bool) MoneyFormat {
	return MoneyFormat{
		CurrencyFirst: currencyFirst,
		ThousandsSep:  '.',
		DecimalSep:    ',',
		Precision:     2,
	}
}
