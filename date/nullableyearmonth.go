package date

import (
	"strings"
	"time"

	"github.com/domonda/go-types/nullable"
)

const YearMonthNull = ""

// Compile-time check that NullableYearMonth implements nullable.NullSetable[Date]
var _ nullable.NullSetable[YearMonth] = (*NullableYearMonth)(nil)

type NullableYearMonth string

// Validate returns nil if the year-month is in valid YYYY-MM format.
func (ym NullableYearMonth) Validate() error {
	if ym.IsNull() {
		return nil
	}
	return YearMonth(ym).Validate()
}

// Valid returns true if the year-month is in valid YYYY-MM format.
func (ym NullableYearMonth) Valid() bool {
	return ym.Validate() == nil
}

// ValidAndNotNull returns if the year-month is valid and not Null or Zero.
func (ym NullableYearMonth) ValidAndNotNull() bool {
	return YearMonth(ym).Valid()
}

// IsZero returns true when the year-month is any of ["", "0000-00", "0001-01"].
// "0001-01" is treated as zero because "0001-01-01" is the zero value of time.Time.
// "0000-00" may be the zero value of other date implementations.
func (ym NullableYearMonth) IsZero() bool {
	return YearMonth(ym).IsZero()
}

// IsNull returns true if the NullableYearMonth is null or zero.
// IsNull implements the nullable.Nullable interface.
func (ym NullableYearMonth) IsNull() bool {
	return ym == YearMonthNull
}

// IsNotNull returns true if the NullableYearMonth is not null.
func (ym NullableYearMonth) IsNotNull() bool {
	return ym != YearMonthNull
}

// Get returns the non-nullable YearMonth value or panics if the NullableYearMonth is null.
// Note: check with IsNull before using Get!
func (ym NullableYearMonth) Get() YearMonth {
	if ym.IsNull() {
		panic("NullableYearMonth.Get() called on null value")
	}
	return YearMonth(ym)
}

// GetOr returns the non-nullable YearMonth value or the passed defaultYearMonth if the NullableYearMonth is null.
func (ym NullableYearMonth) GetOr(defaultYearMonth YearMonth) YearMonth {
	if ym.IsNull() {
		return defaultYearMonth
	}
	return YearMonth(ym)
}

// Set sets a YearMonth for this NullableYearMonth.
func (ym *NullableYearMonth) Set(yearMonth YearMonth) {
	*ym = NullableYearMonth(yearMonth)
}

// SetNull sets the NullableYearMonth to null.
func (ym *NullableYearMonth) SetNull() {
	*ym = YearMonthNull
}

// String returns the year-month as a string in YYYY-MM format.
// String implements the fmt.Stringer interface.
func (ym NullableYearMonth) String() string {
	return string(ym)
}

// StringOr returns the NullableYearMonth as string
// or the passed nullString if the NullableYearMonth is null.
func (ym NullableYearMonth) StringOr(nullString string) string {
	if ym.IsNull() {
		return nullString
	}
	return ym.String()
}

// Compare compares the year-month with another NullableYearMonth.
// Returns -1 if ym is before other, +1 if after, 0 if equal.
func (ym NullableYearMonth) Compare(other NullableYearMonth) int {
	return strings.Compare(string(ym), string(other))
}

// Year returns the year component of the year-month.
// Returns 0 if the year-month is null or not valid.
func (ym NullableYearMonth) Year() int {
	if ym.IsNull() {
		return 0
	}
	return YearMonth(ym).Year()
}

// Month returns the month component of the year-month.
// Returns 0 if the year-month is null or not valid.
func (ym NullableYearMonth) Month() time.Month {
	if ym.IsNull() {
		return 0
	}
	return YearMonth(ym).Month()
}

// Date returns a NullableDate for the given day within this year-month.
// Returns Null if the NullableYearMonth is null.
// The day value is formatted as-is without validation.
func (ym NullableYearMonth) Date(day int) NullableDate {
	if ym.IsNull() {
		return Null
	}
	return NullableDate(YearMonth(ym).Date(day))
}

// DateRange returns the first and last NullableDate of the month.
// Returns Null dates if the NullableYearMonth is null.
// Returns the date range [fromDate, untilDate] representing the entire month.
func (ym NullableYearMonth) DateRange() (fromDate, untilDate NullableDate) {
	if ym.IsNull() {
		return Null, Null
	}
	from, until := YearMonth(ym).DateRange()
	return NullableDate(from), NullableDate(until)
}

// Format formats the year-month using the given layout string (see time.Time.Format).
// Returns an empty string if ym or layout are empty, or if ym is null.
// If layout equals YearMonthLayout constant, returns the year-month as-is for efficiency.
// For other layouts, converts to Date using the 1st day of the month and delegates to Date.Format.
func (ym NullableYearMonth) Format(layout string) string {
	if ym.IsNull() {
		return ""
	}
	return YearMonth(ym).Format(layout)
}

// AddYears returns a new NullableYearMonth with the specified number of years added.
// Returns null if the NullableYearMonth is null.
func (ym NullableYearMonth) AddYears(years int) NullableYearMonth {
	if ym.IsNull() {
		return YearMonthNull
	}
	return NullableYearMonth(YearMonth(ym).AddYears(years))
}

// AddMonths returns a new NullableYearMonth with the specified number of months added.
// Returns null if the NullableYearMonth is null.
// Handles month overflow by adjusting the year accordingly.
func (ym NullableYearMonth) AddMonths(months int) NullableYearMonth {
	if ym.IsNull() {
		return YearMonthNull
	}
	return NullableYearMonth(YearMonth(ym).AddMonths(months))
}

// ContainsTime returns true if the given time falls within this year-month.
// Returns false if the NullableYearMonth is null.
func (ym NullableYearMonth) ContainsTime(t time.Time) bool {
	if ym.IsNull() {
		return false
	}
	return YearMonth(ym).ContainsTime(t)
}

// ContainsDate returns true if the given date falls within this year-month.
// Returns false if the NullableYearMonth is null.
func (ym NullableYearMonth) ContainsDate(date Date) bool {
	if ym.IsNull() {
		return false
	}
	return YearMonth(ym).ContainsDate(date)
}
