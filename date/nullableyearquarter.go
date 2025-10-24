package date

import (
	"strings"
	"time"

	"github.com/domonda/go-types/nullable"
)

const YearQuarterNull NullableYearQuarter = ""

// Compile-time check that NullableYearQuarter implements nullable.NullSetable[YearQuarter]
var _ nullable.NullSetable[YearQuarter] = (*NullableYearQuarter)(nil)

type NullableYearQuarter string

// Validate returns nil if the year-quarter is in valid YYYY-Q# format.
func (yq NullableYearQuarter) Validate() error {
	if yq.IsNull() {
		return nil
	}
	return YearQuarter(yq).Validate()
}

// Valid returns true if the year-quarter is in valid YYYY-Q# format.
func (yq NullableYearQuarter) Valid() bool {
	return yq.Validate() == nil
}

// ValidAndNotNull returns if the year-quarter is valid and not Null or Zero.
func (yq NullableYearQuarter) ValidAndNotNull() bool {
	return YearQuarter(yq).Valid()
}

// IsZero returns true when the year-quarter is any of ["", "0000-Q0", "0001-Q1"].
// "0001-Q1" is treated as zero because "0001-01-01" is the zero value of time.Time.
// "0000-Q0" may be the zero value of other date implementations.
func (yq NullableYearQuarter) IsZero() bool {
	return YearQuarter(yq).IsZero()
}

// IsNull returns true if the NullableYearQuarter is null or zero.
// IsNull implements the nullable.Nullable interface.
func (yq NullableYearQuarter) IsNull() bool {
	return yq == YearQuarterNull
}

// IsNotNull returns true if the NullableYearQuarter is not null.
func (yq NullableYearQuarter) IsNotNull() bool {
	return yq != YearQuarterNull
}

// Get returns the non-nullable YearQuarter value or panics if the NullableYearQuarter is null.
// Note: check with IsNull before using Get!
func (yq NullableYearQuarter) Get() YearQuarter {
	if yq.IsNull() {
		panic("NullableYearQuarter.Get() called on null value")
	}
	return YearQuarter(yq)
}

// GetOr returns the non-nullable YearQuarter value or the passed defaultYearQuarter if the NullableYearQuarter is null.
func (yq NullableYearQuarter) GetOr(defaultYearQuarter YearQuarter) YearQuarter {
	if yq.IsNull() {
		return defaultYearQuarter
	}
	return YearQuarter(yq)
}

// Set sets a YearQuarter for this NullableYearQuarter.
func (yq *NullableYearQuarter) Set(yearQuarter YearQuarter) {
	*yq = NullableYearQuarter(yearQuarter)
}

// SetNull sets the NullableYearQuarter to null.
func (yq *NullableYearQuarter) SetNull() {
	*yq = YearQuarterNull
}

// String returns the year-quarter as a string in YYYY-Q# format.
// String implements the fmt.Stringer interface.
func (yq NullableYearQuarter) String() string {
	return string(yq)
}

// StringOr returns the NullableYearQuarter as string
// or the passed nullString if the NullableYearQuarter is null.
func (yq NullableYearQuarter) StringOr(nullString string) string {
	if yq.IsNull() {
		return nullString
	}
	return yq.String()
}

// Compare compares the year-quarter with another NullableYearQuarter.
// Returns -1 if yq is before other, +1 if after, 0 if equal.
func (yq NullableYearQuarter) Compare(other NullableYearQuarter) int {
	return strings.Compare(string(yq), string(other))
}

// Year returns the year component of the year-quarter.
// Returns 0 if the year-quarter is null or not valid.
func (yq NullableYearQuarter) Year() int {
	if yq.IsNull() {
		return 0
	}
	return YearQuarter(yq).Year()
}

// Quarter returns the quarter component of the year-quarter (1-4).
// Returns 0 if the year-quarter is null or not valid.
func (yq NullableYearQuarter) Quarter() int {
	if yq.IsNull() {
		return 0
	}
	return YearQuarter(yq).Quarter()
}

// FirstMonth returns the first month of the quarter.
// Returns 0 if the year-quarter is null or not valid.
func (yq NullableYearQuarter) FirstMonth() time.Month {
	if yq.IsNull() {
		return 0
	}
	return YearQuarter(yq).FirstMonth()
}

// Date returns a Date for the given month (1-3 within quarter) and day.
// Returns an empty Date if the NullableYearQuarter is null.
func (yq NullableYearQuarter) Date(monthInQuarter, day int) Date {
	if yq.IsNull() {
		return ""
	}
	return YearQuarter(yq).Date(monthInQuarter, day)
}

// DateRange returns the first and last date of the quarter.
// Returns empty dates if the NullableYearQuarter is null.
func (yq NullableYearQuarter) DateRange() (fromDate, untilDate Date) {
	if yq.IsNull() {
		return "", ""
	}
	return YearQuarter(yq).DateRange()
}

// AddYears returns a new NullableYearQuarter with the specified number of years added.
// Returns null if the NullableYearQuarter is null.
func (yq NullableYearQuarter) AddYears(years int) NullableYearQuarter {
	if yq.IsNull() {
		return YearQuarterNull
	}
	return NullableYearQuarter(YearQuarter(yq).AddYears(years))
}

// AddQuarters returns a new NullableYearQuarter with the specified number of quarters added.
// Returns null if the NullableYearQuarter is null.
// Handles quarter overflow by adjusting the year accordingly.
func (yq NullableYearQuarter) AddQuarters(quarters int) NullableYearQuarter {
	if yq.IsNull() {
		return YearQuarterNull
	}
	return NullableYearQuarter(YearQuarter(yq).AddQuarters(quarters))
}

// ContainsTime returns true if the given time falls within this year-quarter.
// Returns false if the NullableYearQuarter is null.
func (yq NullableYearQuarter) ContainsTime(t time.Time) bool {
	if yq.IsNull() {
		return false
	}
	return YearQuarter(yq).ContainsTime(t)
}

// ContainsDate returns true if the given date falls within this year-quarter.
// Returns false if the NullableYearQuarter is null.
func (yq NullableYearQuarter) ContainsDate(date Date) bool {
	if yq.IsNull() {
		return false
	}
	return YearQuarter(yq).ContainsDate(date)
}

// ContainsYearMonth returns true if the given year-month falls within this year-quarter.
// Returns false if the NullableYearQuarter is null.
func (yq NullableYearQuarter) ContainsYearMonth(ym YearMonth) bool {
	if yq.IsNull() {
		return false
	}
	return YearQuarter(yq).ContainsYearMonth(ym)
}
