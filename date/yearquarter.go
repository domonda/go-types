package date

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// YearQuarter represents a calendar year and quarter in format YYYY-Q# (e.g., "2023-Q1").
type YearQuarter string

// YearQuarterFrom creates a YearQuarter from the given year and quarter values.
// Quarter must be 1-4. Returns the year-quarter in normalized YYYY-Q# format.
func YearQuarterFrom(year int, quarter int) YearQuarter {
	return YearQuarter(fmt.Sprintf("%04d-Q%d", year, quarter))
}

// YearQuarterOfTime returns the year-quarter part of the passed time.Time.
// Returns an empty string if t.IsZero().
func YearQuarterOfTime(t time.Time) YearQuarter {
	if t.IsZero() {
		return ""
	}
	year, month, _ := t.Date()
	quarter := (int(month)-1)/3 + 1
	return YearQuarterFrom(year, quarter)
}

// YearQuarterOfToday returns the year-quarter of today in the local timezone.
func YearQuarterOfToday() YearQuarter {
	return YearQuarterOfTime(time.Now())
}

// Validate returns an error if the year-quarter is not in valid YYYY-Q# format.
// Checks that the year is within reasonable range (≤3000) and quarter is 1-4.
func (yq YearQuarter) Validate() error {
	if len(yq) != 7 || yq[4] != '-' || yq[5] != 'Q' {
		return fmt.Errorf("invalid year quarter: %q", yq)
	}
	yearStr := string(yq)[:4]
	year, err := strconv.ParseUint(yearStr, 10, 16)
	if err != nil || year > 3000 {
		return fmt.Errorf("invalid year: %q", yearStr)
	}
	quarterStr := string(yq)[6:7]
	quarter, err := strconv.ParseUint(quarterStr, 10, 8)
	if err != nil || quarter < 1 || quarter > 4 {
		return fmt.Errorf("invalid quarter: %q", quarterStr)
	}
	return nil
}

// Valid returns true if the year-quarter is in valid YYYY-Q# format.
func (yq YearQuarter) Valid() bool {
	return yq.Validate() == nil
}

// IsZero returns true when the year-quarter is any of ["", "0000-Q0", "0001-Q1"].
// "0001-Q1" is treated as zero because "0001-01-01" is the zero value of time.Time.
// "0000-Q0" may be the zero value of other date implementations.
func (yq YearQuarter) IsZero() bool {
	return yq == "" || yq == "0000-Q0" || yq == "0001-Q1"
}

// String returns the year-quarter as a string in YYYY-Q# format.
// String implements the fmt.Stringer interface.
func (yq YearQuarter) String() string {
	return string(yq)
}

// Year returns the year component of the year-quarter.
// Returns 0 if the year-quarter is not valid.
func (yq YearQuarter) Year() int {
	if len(yq) < 4 {
		return 0
	}
	year, _ := strconv.Atoi(string(yq)[:4])
	return year
}

// Quarter returns the quarter component of the year-quarter (1-4).
// Returns 0 if the year-quarter is not valid.
func (yq YearQuarter) Quarter() int {
	if len(yq) < 7 {
		return 0
	}
	quarter, _ := strconv.Atoi(string(yq)[6:7])
	return quarter
}

// FirstMonth returns the first month of the quarter (1, 4, 7, or 10).
// Returns 0 if the year-quarter is not valid.
func (yq YearQuarter) FirstMonth() time.Month {
	q := yq.Quarter()
	if q == 0 {
		return 0
	}
	return time.Month((q-1)*3 + 1)
}

// Date returns a Date for the given month (1-3 within quarter) and day.
// The month parameter is 1-3 representing the month within the quarter.
// The values are formatted as-is without validation.
func (yq YearQuarter) Date(monthInQuarter, day int) Date {
	firstMonth := yq.FirstMonth()
	if firstMonth == 0 {
		return ""
	}
	month := int(firstMonth) + monthInQuarter - 1
	return Of(yq.Year(), time.Month(month), day)
}

// DateRange returns the first and last date of the quarter.
// Returns the date range [fromDate, untilDate] representing the entire quarter.
func (yq YearQuarter) DateRange() (fromDate, untilDate Date) {
	firstMonth := yq.FirstMonth()
	if firstMonth == 0 {
		return "", ""
	}
	fromDate = Of(yq.Year(), firstMonth, 1)
	// Last day of quarter is last day of third month
	untilDate = Of(yq.Year(), firstMonth+2, 1).AddMonths(1).AddDays(-1)
	return fromDate, untilDate
}

// Nullable returns the year-quarter as a NullableYearQuarter.
func (yq YearQuarter) Nullable() NullableYearQuarter {
	return NullableYearQuarter(yq)
}

// AddYears returns a new year-quarter with the specified number of years added.
func (yq YearQuarter) AddYears(years int) YearQuarter {
	return YearQuarterFrom(yq.Year()+years, yq.Quarter())
}

// AddQuarters returns a new year-quarter with the specified number of quarters added.
// Handles quarter overflow by adjusting the year accordingly.
func (yq YearQuarter) AddQuarters(quarters int) YearQuarter {
	year := yq.Year()
	quarter := yq.Quarter()

	totalQuarters := quarter + quarters

	// Handle positive overflow
	for totalQuarters > 4 {
		year++
		totalQuarters -= 4
	}

	// Handle negative overflow
	for totalQuarters < 1 {
		year--
		totalQuarters += 4
	}

	return YearQuarterFrom(year, totalQuarters)
}

// ContainsTime returns true if the given time falls within this year-quarter.
func (yq YearQuarter) ContainsTime(t time.Time) bool {
	from, until := yq.DateRange()
	if from == "" || until == "" {
		return false
	}
	return !t.Before(from.Midnight()) && !t.After(until.Midnight().AddDate(0, 0, 1).Add(-time.Nanosecond))
}

// ContainsDate returns true if the given date falls within this year-quarter.
func (yq YearQuarter) ContainsDate(date Date) bool {
	return yq.ContainsTime(date.Midnight())
}

// ContainsYearMonth returns true if the given year-month falls within this year-quarter.
func (yq YearQuarter) ContainsYearMonth(ym YearMonth) bool {
	if !ym.Valid() {
		return false
	}
	return yq.Year() == ym.Year() && ym.Month() >= yq.FirstMonth() && ym.Month() <= yq.FirstMonth()+2
}

// Compare compares the year-quarter with another YearQuarter.
// Returns -1 if yq is before other, +1 if after, 0 if equal.
func (yq YearQuarter) Compare(other YearQuarter) int {
	return strings.Compare(string(yq), string(other))
}
