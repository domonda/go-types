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
			reflect.TypeOf((*date.Date)(nil)).Elem():            FormatterFunc(formatDateString),
			reflect.TypeOf((*date.NullableDate)(nil)).Elem():    FormatterFunc(formatNullableDateString),
			reflect.TypeOf((*time.Time)(nil)).Elem():            FormatterFunc(formatTimeString),
			reflect.TypeOf((*nullable.Time)(nil)).Elem():        FormatterFunc(formatNullableTimeString),
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
	config.Date = "02/01/2006"
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

func formatDateString(val reflect.Value, config *FormatConfig) string {
	return val.Interface().(date.Date).Format(config.Date)
}

func formatNullableDateString(val reflect.Value, config *FormatConfig) string {
	return val.Interface().(date.NullableDate).Format(config.Date)
}

func formatTimeString(val reflect.Value, config *FormatConfig) string {
	return val.Interface().(time.Time).Format(config.Time)
}

func formatNullableTimeString(val reflect.Value, config *FormatConfig) string {
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
