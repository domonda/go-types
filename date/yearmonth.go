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

// YearMonthLayout is the layout string for formatting YearMonth values (YYYY-MM).
// Compatible with time.Time.Format().
const YearMonthLayout = "2006-01"

// Compile-time check that YearMonth implements types.NormalizableValidator[YearMonth]
var _ types.NormalizableValidator[YearMonth] = YearMonth("")

// YearMonth represents a calendar year and month in ISO 8601 format (YYYY-MM).
// Normalized trims the surrounding whitespace, UnmarshalJSON stores a valid
// year-month normalized and an invalid one unchanged.
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

// Normalized returns the year-month with the surrounding whitespace trimmed,
// or the YearMonth unchanged together with an error if that is not in
// valid YYYY-MM format with a year from 1 to 3000, because Date has no
// year 0, and a month from 1 to 12.
func (ym YearMonth) Normalized() (YearMonth, error) {
	norm := strutil.TrimSpace(ym)
	if len(norm) != 7 || norm[4] != '-' {
		return ym, fmt.Errorf("invalid year month: %q", ym)
	}
	yearStr := string(norm)[:4]
	year, err := strconv.ParseUint(yearStr, 10, 16)
	if err != nil || year < 1 || year > 3000 {
		return ym, fmt.Errorf("invalid year: %q", yearStr)
	}
	monthStr := string(norm)[5:7]
	month, err := strconv.ParseUint(monthStr, 10, 8)
	if err != nil || month < 1 || month > 12 {
		return ym, fmt.Errorf("invalid month: %q", monthStr)
	}
	return norm, nil
}

// Validate returns the error of Normalized,
// so a year-month with surrounding whitespace is valid.
func (ym YearMonth) Validate() error {
	_, err := ym.Normalized()
	return err
}

// Valid returns true if the year-month can be normalized, see Validate.
func (ym YearMonth) Valid() bool {
	return ym.Validate() == nil
}

// ValidAndNormalized returns true if the year-month is valid and already normalized.
func (ym YearMonth) ValidAndNormalized() bool {
	norm, err := ym.Normalized()
	return err == nil && ym == norm
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored normalized if it is a valid year-month,
// blank as "" and unchanged if it is invalid, without an error, so that
// one invalid value doesn't fail a whole document; Validate reports it.
// JSON null leaves the YearMonth unchanged like encoding/json does for a
// string type, and any other JSON value is an error.
func (ym *YearMonth) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(ym, data, false, YearMonth.Normalized, nil)
}

// IsZero returns true when the year-month is any of ["", "0000-00", "0001-01"],
// also with surrounding whitespace.
// "0001-01" is treated as zero because "0001-01-01" is the zero value of time.Time.
// "0000-00" may be the zero value of other date implementations.
func (ym YearMonth) IsZero() bool {
	switch strutil.TrimSpace(ym) {
	case "", "0000-00", "0001-01":
		return true
	}
	return false
}

// String returns the year-month as a string in YYYY-MM format.
// String implements the fmt.Stringer interface.
func (ym YearMonth) String() string {
	return string(ym)
}

// Year returns the year component of the year-month.
// Returns 0 if the year-month is not valid.
func (ym YearMonth) Year() int {
	norm, err := ym.Normalized()
	if err != nil {
		return 0
	}
	year, _ := strconv.Atoi(string(norm)[:4])
	return year
}

// Month returns the month component of the year-month.
// Returns 0 if the year-month is not valid.
func (ym YearMonth) Month() time.Month {
	norm, err := ym.Normalized()
	if err != nil {
		return 0
	}
	month, _ := strconv.Atoi(string(norm)[5:7])
	return time.Month(month)
}

// Date returns a Date for the given day within this year-month.
// The year-month is normalized if it is valid, the day value
// is formatted as-is without validation.
func (ym YearMonth) Date(day int) Date {
	norm, _ := ym.Normalized() // unchanged if invalid
	return Date(fmt.Sprintf("%s-%02d", norm, day))
}

// Nullable returns the year-month as a NullableYearMonth.
func (ym YearMonth) Nullable() NullableYearMonth {
	return NullableYearMonth(ym)
}

// DateRange returns the first and last date of the month.
// Returns the date range [fromDate, untilDate] representing the entire month
// where untilDate is the last day of the month.
// Returns empty dates if the year-month is not valid.
func (ym YearMonth) DateRange() (fromDate, untilDate Date) {
	if !ym.Valid() {
		return "", ""
	}
	fromDate = ym.Date(1)
	untilDate = Date(fromDate.Midnight().
		AddDate(0, 1, 0).
		AddDate(0, 0, -1).
		Format(Layout),
	)
	return fromDate, untilDate
}

// Format formats the year-month using the given layout string (see time.Time.Format).
// Returns an empty string if ym or layout are empty.
// If layout equals YearMonthLayout constant, returns the normalized year-month,
// or an invalid one unchanged. For other layouts, converts to Date using
// the 1st day of the month and delegates to Date.Format.
func (ym YearMonth) Format(layout string) string {
	if ym == "" || layout == "" {
		return ""
	}
	if layout == YearMonthLayout {
		norm, _ := ym.Normalized() // unchanged if invalid
		return string(norm)
	}
	return ym.Date(1).Format(layout)
}

// AddYears returns a new year-month with the specified number of years added.
// Returns the year-month unchanged if it is not valid.
func (ym YearMonth) AddYears(years int) YearMonth {
	if !ym.Valid() {
		return ym
	}
	return YearMonthFrom(ym.Year()+years, ym.Month())
}

// AddMonths returns a new year-month with the specified number of months added.
// Handles month overflow by adjusting the year accordingly.
// Returns the year-month unchanged if it is not valid.
func (ym YearMonth) AddMonths(months int) YearMonth {
	if !ym.Valid() {
		return ym
	}
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

// Compare compares the normalized year-month with another normalized
// YearMonth, an invalid one as it is.
// Returns -1 if ym is before other, +1 if after, 0 if equal.
func (ym YearMonth) Compare(other YearMonth) int {
	ym, _ = ym.Normalized()
	other, _ = other.Normalized()
	return strings.Compare(string(ym), string(other))
}
