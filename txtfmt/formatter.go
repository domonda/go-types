package txtfmt

import (
	"reflect"
	"time"

	"github.com/domonda/go-types/date"
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

type FormatConfig struct {
	Float          FloatFormat
	MoneyAmount    MoneyFormat
	Percent        FloatFormat
	Time           string
	Date           string
	Nil            string
	True           string
	False          string
	TypeFormatters map[reflect.Type]Formatter
}

func NewFormatConfig() *FormatConfig {
	return &FormatConfig{
		Float:       EnglishFloatFormat(-1),
		MoneyAmount: EnglishMoneyFormat(true),
		Percent:     EnglishFloatFormat(-1),
		Time:        time.RFC3339,
		Date:        date.Layout,
		Nil:         "",
		True:        "true",
		False:       "false",
		TypeFormatters: map[reflect.Type]Formatter{
			reflect.TypeOf((*date.Date)(nil)).Elem():            FormatterFunc(formatDateString),
			reflect.TypeOf((*date.NullableDate)(nil)).Elem():    FormatterFunc(formatNullableDateString),
			reflect.TypeOf((*time.Time)(nil)).Elem():            FormatterFunc(formatTimeString),
			reflect.TypeOf((*time.Duration)(nil)).Elem():        FormatterFunc(formatDurationString),
			reflect.TypeOf((*money.Amount)(nil)).Elem():         FormatterFunc(formatMoneyAmountString),
			reflect.TypeOf((*money.CurrencyAmount)(nil)).Elem(): FormatterFunc(formatMoneyCurrencyAmountString),
		},
	}
}

func NewEnglishFormatConfig() *FormatConfig {
	config := NewFormatConfig()
	config.True = "YES"
	config.False = "NO"
	return config
}

func NewGermanFormatConfig() *FormatConfig {
	config := NewFormatConfig()
	config.Float = GermanFloatFormat(-1)
	config.MoneyAmount = GermanMoneyFormat(true)
	config.Percent = GermanFloatFormat(-1)
	config.Date = "02.01.2006"
	config.True = "JA"
	config.False = "NEIN"
	return config
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
