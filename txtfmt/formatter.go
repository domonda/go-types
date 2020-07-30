package txtfmt

import (
	"reflect"

	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/strfmt"
)

type Formatter interface {
	FormatValue(val reflect.Value, config *FormatConfig) string
}

type FormatterFunc func(val reflect.Value, config *FormatConfig) string

func (f FormatterFunc) FormatValue(val reflect.Value, config *FormatConfig) string {
	return f(val, config)
}

type FloatFormat struct {
	ThousandsSep byte
	DecimalSep   byte
	Precision    int
	PadPrecision bool
}

func (format *FloatFormat) FormatFloat(f float64) string {
	return strfmt.FormatFloat(f, format.ThousandsSep, format.DecimalSep, format.Precision, format.PadPrecision)
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

func EnglishFloatFormat(precision int) FloatFormat {
	return FloatFormat{
		ThousandsSep: 0,
		DecimalSep:   '.',
		Precision:    precision,
		PadPrecision: false,
	}
}

func GermanFloatFormat(precision int) FloatFormat {
	return FloatFormat{
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
