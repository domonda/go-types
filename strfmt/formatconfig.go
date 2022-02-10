package strfmt

import (
	"reflect"
	"time"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
)

type FormatConfig struct {
	Float          float.FormatDef
	MoneyAmount    MoneyFormat
	Percent        float.FormatDef
	Time           string
	Date           string
	Nil            string // Also used for nullable.Nullable and Zeroable
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
			reflect.TypeOf(date.Date("")):          FormatterFunc(formatDateString),
			reflect.TypeOf(date.NullableDate("")):  FormatterFunc(formatNullableDateString),
			reflect.TypeOf(time.Time{}):            FormatterFunc(formatTimeString),
			reflect.TypeOf(nullable.Time{}):        FormatterFunc(formatNullableTimeString),
			reflect.TypeOf(time.Duration(0)):       FormatterFunc(formatDurationString),
			reflect.TypeOf(money.Amount(0)):        FormatterFunc(formatMoneyAmountString),
			reflect.TypeOf(money.CurrencyAmount{}): FormatterFunc(formatMoneyCurrencyAmountString),
		},
	}
}

func NewEnglishFormatConfig() *FormatConfig {
	config := NewFormatConfig()
	config.Date = "02/01/2006"
	config.Time = "02/01/2006 15:04:05 MST"
	config.True = "yes"
	config.False = "no"
	return config
}

func NewGermanFormatConfig() *FormatConfig {
	config := NewFormatConfig()
	config.Float = GermanFloatFormat(-1)
	config.MoneyAmount = GermanMoneyFormat(true)
	config.Percent = GermanFloatFormat(-1)
	config.Date = "02.01.2006"
	config.Time = "02.01.2006 15:04:05 MST"
	config.True = "ja"
	config.False = "nein"
	return config
}

func formatDateString(val reflect.Value, config *FormatConfig) string {
	return val.Interface().(date.Date).Format(config.Date)
}

func formatNullableDateString(val reflect.Value, config *FormatConfig) string {
	return val.Interface().(date.NullableDate).Format(config.Date)
}

func formatTimeString(val reflect.Value, config *FormatConfig) string {
	t := val.Interface().(time.Time)
	if t.IsZero() {
		return config.Nil
	}
	return t.Format(config.Time)
}

func formatNullableTimeString(val reflect.Value, config *FormatConfig) string {
	t := val.Interface().(nullable.Time)
	if t.IsNull() {
		return config.Nil
	}
	return val.Interface().(nullable.Time).Format(config.Time)
}

func formatDurationString(val reflect.Value, config *FormatConfig) string {
	return val.Interface().(time.Duration).String()
}

func formatMoneyAmountString(val reflect.Value, config *FormatConfig) string {
	return config.MoneyAmount.FormatAmount(val.Interface().(money.Amount))
}

func formatMoneyCurrencyAmountString(val reflect.Value, config *FormatConfig) string {
	return config.MoneyAmount.FormatCurrencyAmount(val.Interface().(money.CurrencyAmount))
}
