package strfmt

import (
	"reflect"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/money"
)

type Formatter interface {
	FormatValue(val reflect.Value, config *FormatConfig) string
}

type FormatterFunc func(val reflect.Value, config *FormatConfig) string

func (f FormatterFunc) FormatValue(val reflect.Value, config *FormatConfig) string {
	return f(val, config)
}

type MoneyFormat struct {
	CurrencyFirst bool
	ThousandsSep  byte
	DecimalSep    byte
	Precision     int
}

func (format *MoneyFormat) FormatAmount(amount money.Amount) string {
	return amount.Format(format.ThousandsSep, format.DecimalSep, format.Precision)
}

func (format *MoneyFormat) FormatCurrencyAmount(currencyAmount money.CurrencyAmount) string {
	return currencyAmount.Format(format.CurrencyFirst, format.ThousandsSep, format.DecimalSep, format.Precision)
}

func EnglishFloatFormat(precision int) float.FormatDef {
	return float.FormatDef{
		ThousandsSep: 0,
		DecimalSep:   '.',
		Precision:    precision,
		PadPrecision: false,
	}
}

func GermanFloatFormat(precision int) float.FormatDef {
	return float.FormatDef{
		ThousandsSep: 0,
		DecimalSep:   ',',
		Precision:    precision,
		PadPrecision: false,
	}
}

func EnglishMoneyFormat(currencyFirst bool) MoneyFormat {
	return MoneyFormat{
		CurrencyFirst: currencyFirst,
		ThousandsSep:  ',',
		DecimalSep:    '.',
		Precision:     2,
	}
}

func GermanMoneyFormat(currencyFirst bool) MoneyFormat {
	return MoneyFormat{
		CurrencyFirst: currencyFirst,
		ThousandsSep:  '.',
		DecimalSep:    ',',
		Precision:     2,
	}
}
