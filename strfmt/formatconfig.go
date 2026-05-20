package strfmt

import (
	"reflect"
	"time"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
)

// FormatConfig holds the locale-specific formatting settings used
// by Format and FormatValue when converting typed values to strings.
// It covers numeric separators, date/time layouts, boolean labels,
// a nil/null placeholder, and an optional per-type formatter registry.
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

// NewFormatConfig returns a FormatConfig with English-style defaults:
// dot decimal separator, RFC3339 time layout, ISO date layout,
// empty nil string, and "true"/"false" boolean labels.
// It pre-registers type-specific formatters for date.Date,
// date.NullableDate, time.Time, nullable.Time, time.Duration,
// money.Amount, and money.CurrencyAmount.
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
			reflect.TypeFor[date.Date]():            FormatterFunc(formatDateString),
			reflect.TypeFor[date.NullableDate]():    FormatterFunc(formatNullableDateString),
			reflect.TypeFor[time.Time]():            FormatterFunc(formatTimeString),
			reflect.TypeFor[nullable.Time]():        FormatterFunc(formatNullableTimeString),
			reflect.TypeFor[time.Duration]():        FormatterFunc(formatDurationString),
			reflect.TypeFor[money.Amount]():         FormatterFunc(formatMoneyAmountString),
			reflect.TypeFor[money.CurrencyAmount](): FormatterFunc(formatMoneyCurrencyAmountString),
		},
	}
}

// NewEnglishFormatConfig returns a FormatConfig based on NewFormatConfig
// adjusted for English locale conventions: DD/MM/YYYY date layout,
// a long-form time layout, and "yes"/"no" boolean labels.
func NewEnglishFormatConfig() *FormatConfig {
	config := NewFormatConfig()
	config.Date = "02/01/2006"
	config.Time = "02/01/2006 15:04:05 MST"
	config.True = "yes"
	config.False = "no"
	return config
}

// NewGermanFormatConfig returns a FormatConfig based on NewFormatConfig
// adjusted for German locale conventions: comma decimal separator,
// dot thousands separator, DD.MM.YYYY date layout, a long-form time
// layout, and "ja"/"nein" boolean labels.
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
