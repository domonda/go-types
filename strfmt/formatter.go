package strfmt

import (
	"reflect"
	"strings"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/strutil"
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
// For the currency amount types the Precision is raised to the decimals
// of the minor unit of a currency with more than 2, see FormatCurrencyAmount.
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
// For a currency with more than 2 decimals in its minor unit, like KWD,
// see money.Currency.MinorUnits, a precision of zero or more is raised to
// them, so that "KWD 1,234.567" keeps the fils of the Kuwaiti dinar like
// money.CurrencyAmount.String does, and is never written as "KWD 1,234",
// which money.CurrencyAmount.ScanString reads back as 1.234. For other
// currencies and without currency the precision is kept, so a precision
// of 0 writes whole euros. A precision of -1 writes the shortest form.
func (format *MoneyFormat) FormatCurrencyAmount(currencyAmount money.CurrencyAmount) string {
	precision := format.currencyPrecision(currencyAmount.Currency)
	return currencyAmount.FormatSep(format.CurrencyFirst, format.ThousandsSep, format.DecimalSep, precision)
}

// FormatCurrencyDecimalAmount formats a currency-tagged decimal amount like
// FormatCurrencyAmount formats a currency-tagged amount, placing the normalized
// currency code before or after the number according to format.CurrencyFirst,
// with the number formatted by FormatDecimalAmount with the precision raised
// to the decimals of the minor unit of the currency like FormatCurrencyAmount.
func (format *MoneyFormat) FormatCurrencyDecimalAmount(currencyAmount money.CurrencyDecimalAmount) string {
	amountStr := format.formatDecimalAmount(currencyAmount.Amount, format.currencyPrecision(currencyAmount.Currency))
	currency, _ := currencyAmount.Currency.MarshalText()
	return internal.JoinCurrencyAmount(string(currency), amountStr, format.CurrencyFirst)
}

// currencyPrecision returns format.Precision raised to the decimals of the
// minor unit of a currency with more than 2 of them, because fewer decimals
// would scan back as an amount 10 or 1000 times smaller. A precision of -1
// writes all decimals already, and the precision of other currencies and
// of a blank currency, which has no minor unit, is kept.
func (format *MoneyFormat) currencyPrecision(currency money.Currency) int {
	if format.Precision < 0 || strutil.TrimSpace(currency) == "" {
		return format.Precision
	}
	if units := currency.MinorUnits(); units > 2 {
		return max(format.Precision, units)
	}
	return format.Precision
}

// FormatCentAmount formats a cent amount like FormatAmount formats an amount,
// see FormatDecimalAmount. A cent amount beyond the DecimalAmount range,
// which no real money amount reaches, is formatted with its two decimals
// ignoring the precision.
func (format *MoneyFormat) FormatCentAmount(amount money.CentAmount) string {
	if decimal := amount.DecimalAmount(); decimal.IsFinite() {
		return format.FormatDecimalAmount(decimal)
	}
	return amount.FormatSep(format.ThousandsSep, format.DecimalSep)
}

// FormatDecimalAmount formats a decimal amount like FormatAmount formats an
// amount: with the thousands and decimal separators of format, rounded half
// away from zero to format.Precision decimal places and padded with zeros to
// them. A negative precision formats all decimal places of the amount's scale.
func (format *MoneyFormat) FormatDecimalAmount(amount money.DecimalAmount) string {
	return format.formatDecimalAmount(amount, format.Precision)
}

// formatDecimalAmount is FormatDecimalAmount with the passed precision
// instead of format.Precision.
func (format *MoneyFormat) formatDecimalAmount(amount money.DecimalAmount, precision int) string {
	if precision < 0 || !amount.IsFinite() {
		return amount.FormatSep(format.ThousandsSep, format.DecimalSep)
	}
	if precision < amount.Scale() {
		amount = amount.RoundToDecimals(precision, money.RoundHalfAwayFromZero)
	}
	str := amount.FormatSep(format.ThousandsSep, format.DecimalSep)
	// Pad the text instead of increasing the scale of the amount,
	// which would overflow to Inf for amounts close to the range limit
	padding := precision - amount.Scale()
	if padding <= 0 {
		return str
	}
	if amount.Scale() == 0 {
		decimalSep := format.DecimalSep
		if decimalSep == 0 {
			decimalSep = '.' // like DecimalAmount.FormatSep
		}
		str += string(decimalSep)
	}
	return str + strings.Repeat("0", padding)
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
