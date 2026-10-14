package strfmt

import (
	"reflect"
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
			en:     "03/05/2024",
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
			en:     "03/05/2024 06:07:08 UTC",
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
			assert.Equal(t, tt.en, Format(tt.value, NewUSFormatConfig()), "English")
		})
	}
}

// TestFormatConfig_MoneyTypesFormatAlike checks that money.CentAmount and
// money.DecimalAmount are formatted like money.Amount with the locale's
// money format. Without their type formatters a CentAmount was written as
// its raw cent count ("123450" for 1,234.50, a hundred times too much),
// a zero CentAmount or DecimalAmount as the Nil placeholder, which looks
// like missing data in an export, and a DecimalAmount ignored the decimal
// comma and thousands separator of the locale.
func TestFormatConfig_MoneyTypesFormatAlike(t *testing.T) {
	tests := []struct {
		amount money.Amount
		german string
		en     string
	}{
		{amount: 1234.5, german: "1.234,50", en: "1,234.50"},
		{amount: 0, german: "0,00", en: "0.00"},
		{amount: -0.05, german: "-0,05", en: "-0.05"},
		{amount: 1234567.89, german: "1.234.567,89", en: "1,234,567.89"},
	}
	for _, tt := range tests {
		cents := tt.amount.CentAmount(money.RoundHalfAwayFromZero)
		decimal := tt.amount.DecimalAmount(2, money.RoundHalfAwayFromZero)
		for _, value := range []any{tt.amount, cents, &cents, decimal, &decimal} {
			assert.Equal(t, tt.german, Format(value, NewGermanFormatConfig()), "German %T(%v)", value, value)
			assert.Equal(t, tt.en, Format(value, NewUSFormatConfig()), "English %T(%v)", value, value)
		}
	}

	// A DecimalAmount is rounded or padded to the precision like an Amount
	assert.Equal(t, "12,35", Format(money.NewDecimalAmount(12345, 3), NewGermanFormatConfig()), "12.345 rounds half away from zero")
	assert.Equal(t, "5.00", Format(money.NewDecimalAmount(5, 0), NewUSFormatConfig()), "padded to two decimals")
	assert.Equal(t, "0.00", Format(money.DecimalAmount{}, NewUSFormatConfig()), "zero value is zero, not Nil")
	assert.Equal(t, "NaN", Format(money.DecimalAmountNaN(), NewGermanFormatConfig()))
}

// TestFormatConfig_CurrencyDecimalAmount checks that money.CurrencyDecimalAmount
// is formatted with the locale's money format like money.CurrencyAmount.
// Without its type formatter the String method was used, which ignores the
// locale, so a German export wrote "EUR 1234.56" for a CurrencyDecimalAmount
// next to "EUR 1.234,56" for the same amount as CurrencyAmount.
func TestFormatConfig_CurrencyDecimalAmount(t *testing.T) {
	tests := []struct {
		value  money.CurrencyDecimalAmount
		german string
		en     string
	}{
		{value: money.NewCurrencyDecimalAmount(money.EUR, money.NewDecimalAmount(123456, 2)), german: "EUR 1.234,56", en: "EUR 1,234.56"},
		{value: money.NewCurrencyDecimalAmount(money.EUR, money.NewDecimalAmount(12345, 3)), german: "EUR 12,35", en: "EUR 12.35"},
		{value: money.NewCurrencyDecimalAmount(money.USD, money.NewDecimalAmount(5, 0)), german: "USD 5,00", en: "USD 5.00"},
		{value: money.CurrencyDecimalAmount{Amount: money.NewDecimalAmount(-5, 1)}, german: "-0,50", en: "-0.50"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.german, Format(tt.value, NewGermanFormatConfig()), "German %v", tt.value)
		assert.Equal(t, tt.en, Format(tt.value, NewUSFormatConfig()), "English %v", tt.value)
		assert.Equal(t, tt.german, Format(&tt.value, NewGermanFormatConfig()), "German pointer %v", tt.value)

		// Formatted exactly like the CurrencyAmount of the same value
		currencyAmount := tt.value.CurrencyAmount()
		assert.Equal(t, Format(currencyAmount, NewGermanFormatConfig()), Format(tt.value, NewGermanFormatConfig()), "German like CurrencyAmount %v", tt.value)
	}
}

// TestFormatConfig_NullableMoneyTypes checks that the nullable money types
// format a non-null value like the money type they wrap and null as the
// configured Nil string, through the generic nullable.Type[T] path without
// formatters of their own. Formatted by kind, a non-null
// money.NullableAmount would be written as the raw struct "{1234.5 true}".
func TestFormatConfig_NullableMoneyTypes(t *testing.T) {
	decimal := money.NewDecimalAmount(123450, 2)
	for _, config := range []*FormatConfig{NewGermanFormatConfig(), NewUSFormatConfig()} {
		config.Nil = "N/A"
		want := Format(money.Amount(1234.5), config)

		nullableAmount := money.NullableAmountFrom(1234.5)
		nullableCents := money.NullableCentAmountFrom(123450)
		nullableDecimal := money.NullableDecimalAmountFromPtr(&decimal)
		for _, value := range []any{nullableAmount, &nullableAmount, nullableCents, &nullableCents, nullableDecimal, &nullableDecimal} {
			assert.Equal(t, want, Format(value, config), "Format(%T)", value)
		}

		for _, value := range []any{money.NullableAmount{}, money.NullableCentAmount{}, money.NullableDecimalAmount{}, (*money.NullableAmount)(nil)} {
			assert.Equal(t, "N/A", Format(value, config), "Format(null %T)", value)
		}
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

// TestFormatConfig_NilPlaceholderForDates checks that a null date.NullableDate
// and a zero date.Date are formatted as the configured Nil string, like a
// null nullable.Time and a zero time.Time, instead of as "". An export with
// Nil "-" had empty date cells next to "-" for every other missing value,
// and a Nil placeholder that the ScanConfig reads as nil string, like
// "NULL", has to scan back as null.
func TestFormatConfig_NilPlaceholderForDates(t *testing.T) {
	for _, config := range []*FormatConfig{NewFormatConfig(), NewGermanFormatConfig(), NewUSFormatConfig()} {
		config.Nil = "NULL"
		for _, value := range []any{date.Null, date.NullableDate("0001-01-01"), date.Date(""), date.Date("0001-01-01"), date.Date("0000-00-00")} {
			assert.Equal(t, "NULL", Format(value, config), "Format(%T(%q))", value, value)
		}
		assert.Equal(t, Format(date.Date("2024-03-05"), config), Format(date.NullableDate("2024-03-05"), config), "date.NullableDate like date.Date")
		assert.NotEqual(t, "NULL", Format(date.Date("2024-03-05"), config), "date.Date")

		nullDate := date.NullableDate("2024-03-05")
		assert.NoError(t, Scan(reflect.ValueOf(&nullDate).Elem(), Format(date.Null, config), nil))
		assert.True(t, nullDate.IsNull(), "the Nil placeholder scans back as null")
	}
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
