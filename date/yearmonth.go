package date

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// YearMonthLayout is the layout string for formatting YearMonth values (YYYY-MM).
// Compatible with time.Time.Format().
const YearMonthLayout = "2006-01"

// YearMonth represents a calendar year and month in ISO 8601 format (YYYY-MM).
type YearMonth string

// YearMonthFrom creates a YearMonth from the given year and month values.
// Returns the year-month in normalized YYYY-MM format.
func YearMonthFrom(year int, month time.Month) YearMonth {
	return YearMonth(fmt.Sprintf("%04d-%02d", year, month))
}

// YearMonthOfTime returns the year-month part of the passed time.Time.
// Returns an empty string if t.IsZero().
func YearMonthOfTime(t time.Time) YearMonth {
	if t.IsZero() {
		return ""
	}
	return YearMonth(t.Format(YearMonthLayout))
}

// YearMonthOfToday returns the year-month of today in the local timezone.
func YearMonthOfToday() YearMonth {
	return YearMonthOfTime(time.Now())
}

// Validate returns an error if the year-month is not in valid YYYY-MM format.
// Checks that the year is within reasonable range (≤3000) and month is 1-12.
func (ym YearMonth) Validate() error {
	if len(ym) != 7 || ym[4] != '-' {
		return fmt.Errorf("invalid year month: %q", ym)
	}
	yearStr := string(ym)[:4]
	year, err := strconv.ParseUint(yearStr, 10, 16)
	if err != nil || year > 3000 {
		return fmt.Errorf("invalid year: %q", yearStr)
	}
	monthStr := string(ym)[5:7]
	month, err := strconv.ParseUint(monthStr, 10, 8)
	if err != nil || month < 1 || month > 12 {
		return fmt.Errorf("invalid month: %q", monthStr)
	}
	return nil
}

// Valid returns true if the year-month is in valid YYYY-MM format.
func (ym YearMonth) Valid() bool {
	return ym.Validate() == nil
}

// IsZero returns true when the year-month is any of ["", "0000-00", "0001-01"].
// "0001-01" is treated as zero because "0001-01-01" is the zero value of time.Time.
// "0000-00" may be the zero value of other date implementations.
func (ym YearMonth) IsZero() bool {
	return ym == "" || ym == "0000-00" || ym == "0001-01"
}

// String returns the year-month as a string in YYYY-MM format.
// String implements the fmt.Stringer interface.
func (ym YearMonth) String() string {
	return string(ym)
}

// Year returns the year component of the year-month.
// Returns 0 if the year-month is not valid.
func (ym YearMonth) Year() int {
	year, _ := strconv.Atoi(string(ym)[:4])
	return year
}

// Month returns the month component of the year-month.
// Returns 0 if the year-month is not valid.
func (ym YearMonth) Month() time.Month {
	month, _ := strconv.Atoi(string(ym)[5:7])
	return time.Month(month)
}

// Date returns a Date for the given day within this year-month.
// The day value is formatted as-is without validation.
func (ym YearMonth) Date(day int) Date {
	return Date(fmt.Sprintf("%s-%02d", ym, day))
}

// Nullable returns the year-month as a NullableYearMonth.
func (ym YearMonth) Nullable() NullableYearMonth {
	return NullableYearMonth(ym)
}

// DateRange returns the first and last date of the month.
// Returns the date range [fromDate, untilDate] representing the entire month
// where untilDate is the last day of the month.
func (ym YearMonth) DateRange() (fromDate, untilDate Date) {
	fromDate = ym.Date(1)
	untilDate = Date(fromDate.Midnight().
		AddDate(0, 1, 0).
		AddDate(0, 0, -1).
		Format(YearMonthLayout),
	)
	return fromDate, untilDate
}

// Format formats the year-month using the given layout string (see time.Time.Format).
// Returns an empty string if ym or layout are empty.
// If layout equals YearMonthLayout constant, returns the year-month as-is for efficiency.
// For other layouts, converts to Date using the 1st day of the month and delegates to Date.Format.
func (ym YearMonth) Format(layout string) string {
	if ym == "" || layout == "" {
		return ""
	}
	if layout == YearMonthLayout {
		return string(ym)
	}
	return ym.Date(1).Format(layout)
}

// AddYears returns a new year-month with the specified number of years added.
func (ym YearMonth) AddYears(years int) YearMonth {
	return YearMonthFrom(ym.Year()+years, ym.Month())
}

// AddMonths returns a new year-month with the specified number of months added.
// Handles month overflow by adjusting the year accordingly.
func (ym YearMonth) AddMonths(months int) YearMonth {
	return YearMonthOfTime(ym.Date(1).AddMonths(months).Midnight())
}

// ContainsTime returns true if the given time falls within this year-month.
func (ym YearMonth) ContainsTime(t time.Time) bool {
	from := ym.Date(1).Midnight()
	switch t.Compare(from) {
	case 0: // t is equal to from
		return true
	case -1: // t is before from
		return false
	}
	return t.Before(from.AddDate(0, 1, 0))
}

// ContainsDate returns true if the given date falls within this year-month.
func (ym YearMonth) ContainsDate(date Date) bool {
	return ym.ContainsTime(date.Midnight())
}

// Compare compares the year-month with another YearMonth.
// Returns -1 if ym is before other, +1 if after, 0 if equal.
func (ym YearMonth) Compare(other YearMonth) int {
	return strings.Compare(string(ym), string(other))
}
