package strfmt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
)

// TestFormatConfig_TypeFormatters checks the type formatters that
// NewFormatConfig registers. They exist because the generic conversions
// would format these types wrongly for the configured locale: dates and
// times through their own String/MarshalText methods instead of the
// configured layouts, and money with the float precision instead of
// two fixed decimals and the locale's separators.
func TestFormatConfig_TypeFormatters(t *testing.T) {
	someTime := time.Date(2024, 3, 5, 6, 7, 8, 0, time.UTC)

	tests := []struct {
		name   string
		value  any
		german string
		en     string
	}{
		{
			name:   "date.NullableDate uses the configured date layout",
			value:  date.NullableDate("2024-03-05"),
			german: "05.03.2024",
			en:     "05/03/2024",
		},
		{
			name:   "date.NullableDate null",
			value:  date.Null,
			german: "",
			en:     "",
		},
		{
			name:   "nullable.Time uses the configured time layout",
			value:  nullable.TimeFrom(someTime),
			german: "05.03.2024 06:07:08 UTC",
			en:     "05/03/2024 06:07:08 UTC",
		},
		{
			name:   "time.Duration uses Go duration syntax independent of locale",
			value:  90 * time.Minute,
			german: "1h30m0s",
			en:     "1h30m0s",
		},
		{
			name:   "money.CurrencyAmount with currency first and locale separators",
			value:  money.CurrencyAmount{Currency: money.EUR, Amount: 1234.5},
			german: "EUR 1.234,50",
			en:     "EUR 1,234.50",
		},
		{
			name:   "money.CurrencyAmount without currency formats only the amount",
			value:  money.CurrencyAmount{Amount: -0.5},
			german: "-0,50",
			en:     "-0.50",
		},
		{
			name:   "pointer to money.CurrencyAmount is dereferenced",
			value:  &money.CurrencyAmount{Currency: money.USD, Amount: 1},
			german: "USD 1,00",
			en:     "USD 1.00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.german, Format(tt.value, NewGermanFormatConfig()), "German")
			assert.Equal(t, tt.en, Format(tt.value, NewEnglishFormatConfig()), "English")
		})
	}
}

// TestFormatConfig_NilPlaceholder checks that the configured Nil string
// is used for null time values instead of a formatted zero time like
// "01.01.0001 00:00:00 UTC", which would look like real data in a report.
func TestFormatConfig_NilPlaceholder(t *testing.T) {
	config := NewGermanFormatConfig()
	config.Nil = "N/A"

	assert.Equal(t, "N/A", Format(nullable.TimeNull, config), "nullable.Time null")
	assert.Equal(t, "N/A", Format(time.Time{}, config), "zero time.Time")
	assert.Equal(t, "N/A", Format((*int)(nil), config), "nil pointer")
}

// TestNewFormatConfig_defaults locks the locale conventions of the
// constructors, which callers rely on when exporting CSV or spreadsheets.
func TestNewFormatConfig_defaults(t *testing.T) {
	someTime := time.Date(2024, 3, 5, 6, 7, 8, 0, time.UTC)

	config := NewFormatConfig()
	assert.Equal(t, "2024-03-05", Format(date.Date("2024-03-05"), config), "ISO date layout")
	assert.Equal(t, "2024-03-05T06:07:08Z", Format(someTime, config), "RFC3339 time layout")
	assert.Equal(t, "true", Format(true, config))
	assert.Equal(t, "false", Format(false, config))
	assert.Equal(t, "1234.5", Format(1234.5, config), "float without thousands separator")
	assert.Equal(t, "EUR 1,234.50", Format(money.CurrencyAmount{Currency: money.EUR, Amount: 1234.5}, config))

	german := NewGermanFormatConfig()
	assert.Equal(t, "1234,5", Format(1234.5, german), "German decimal comma")
	assert.Equal(t, "1234,5", Format(float32(1234.5), german), "German decimal comma for float32")
}
