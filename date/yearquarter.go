package date

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/strutil"
)

// Compile-time check that YearQuarter implements types.NormalizableValidator[YearQuarter]
var _ types.NormalizableValidator[YearQuarter] = YearQuarter("")

// YearQuarter represents a calendar year and quarter in format YYYY-Q# (e.g., "2023-Q1").
// Normalized trims the surrounding whitespace and upper-cases a "q" like
// PeriodRange accepts it, UnmarshalJSON stores a valid year-quarter
// normalized and an invalid one unchanged.
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

// Normalized returns the year-quarter with the surrounding whitespace
// trimmed and an upper case "Q", or the YearQuarter unchanged together
// with an error if that is not in valid YYYY-Q# format with a year from
// 1 to 3000, because Date has no year 0, and a quarter from 1 to 4.
func (yq YearQuarter) Normalized() (YearQuarter, error) {
	norm := strutil.TrimSpace(yq)
	if len(norm) == 7 && norm[5] == 'q' {
		norm = norm[:5] + "Q" + norm[6:]
	}
	if len(norm) != 7 || norm[4] != '-' || norm[5] != 'Q' {
		return yq, fmt.Errorf("invalid year quarter: %q", yq)
	}
	yearStr := string(norm)[:4]
	year, err := strconv.ParseUint(yearStr, 10, 16)
	if err != nil || year < 1 || year > 3000 {
		return yq, fmt.Errorf("invalid year: %q", yearStr)
	}
	quarterStr := string(norm)[6:7]
	quarter, err := strconv.ParseUint(quarterStr, 10, 8)
	if err != nil || quarter < 1 || quarter > 4 {
		return yq, fmt.Errorf("invalid quarter: %q", quarterStr)
	}
	return norm, nil
}

// Validate returns the error of Normalized,
// so "2023-q1" and a year-quarter with surrounding whitespace are valid.
func (yq YearQuarter) Validate() error {
	_, err := yq.Normalized()
	return err
}

// Valid returns true if the year-quarter can be normalized, see Validate.
func (yq YearQuarter) Valid() bool {
	return yq.Validate() == nil
}

// ValidAndNormalized returns true if the year-quarter is valid and already normalized.
func (yq YearQuarter) ValidAndNormalized() bool {
	norm, err := yq.Normalized()
	return err == nil && yq == norm
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored normalized if it is a valid year-quarter,
// blank as "" and unchanged if it is invalid, without an error, so that
// one invalid value doesn't fail a whole document; Validate reports it.
// JSON null leaves the YearQuarter unchanged like encoding/json does for a
// string type, and any other JSON value is an error.
func (yq *YearQuarter) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(yq, data, false, YearQuarter.Normalized, nil)
}

// IsZero returns true when the year-quarter is any of ["", "0000-Q0", "0001-Q1"],
// also with surrounding whitespace or a lower case "q".
// "0001-Q1" is treated as zero because "0001-01-01" is the zero value of time.Time.
// "0000-Q0" may be the zero value of other date implementations.
func (yq YearQuarter) IsZero() bool {
	switch strings.ToUpper(string(strutil.TrimSpace(yq))) {
	case "", "0000-Q0", "0001-Q1":
		return true
	}
	return false
}

// String returns the year-quarter as a string in YYYY-Q# format.
// String implements the fmt.Stringer interface.
func (yq YearQuarter) String() string {
	return string(yq)
}

// Year returns the year component of the year-quarter.
// Returns 0 if the year-quarter is not valid.
func (yq YearQuarter) Year() int {
	norm, err := yq.Normalized()
	if err != nil {
		return 0
	}
	year, _ := strconv.Atoi(string(norm)[:4])
	return year
}

// Quarter returns the quarter component of the year-quarter (1-4).
// Returns 0 if the year-quarter is not valid.
func (yq YearQuarter) Quarter() int {
	norm, err := yq.Normalized()
	if err != nil {
		return 0
	}
	quarter, _ := strconv.Atoi(string(norm)[6:7])
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
// Returns the year-quarter unchanged if it is not valid.
func (yq YearQuarter) AddYears(years int) YearQuarter {
	if !yq.Valid() {
		return yq
	}
	return YearQuarterFrom(yq.Year()+years, yq.Quarter())
}

// AddQuarters returns a new year-quarter with the specified number of quarters added.
// Handles quarter overflow by adjusting the year accordingly.
// Returns the year-quarter unchanged if it is not valid.
func (yq YearQuarter) AddQuarters(quarters int) YearQuarter {
	if !yq.Valid() {
		return yq
	}
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

// Compare compares the normalized year-quarter with another normalized
// YearQuarter, an invalid one as it is.
// Returns -1 if yq is before other, +1 if after, 0 if equal.
func (yq YearQuarter) Compare(other YearQuarter) int {
	yq, _ = yq.Normalized()
	other, _ = other.Normalized()
	return strings.Compare(string(yq), string(other))
}
