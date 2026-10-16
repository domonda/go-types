package date

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-errs"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/language"
	"github.com/domonda/go-types/strutil"
)

// Normalize returns str as a Date in the canonical YYYY-MM-DD form,
// or an error if the input cannot be parsed.
//
// A numeric date with the year last is ambiguous when both of its first
// numbers are valid months, like "05/06/2024". The separators decide:
// dots are read day first like in Germany, "05.06.2024" is 2024-06-05,
// and so are other separators, like "05-06-2024", with any language
// hint. Slashes are written month first in the US and day first in
// Britain, so for slashes the first lang argument, if present, is a
// hint: language.EN reads month first, every other language day first,
// so "05/06/2024" with language.DE is 2024-06-05. Without a hint
// slashes are read month first like in the US, "05/06/2024" is
// 2024-05-06. An English ordinal suffix marks the day and wins over
// the separators and the hint: "1st 02 2003" is 2003-02-01 also with
// language.EN.
//
// Accepted input shapes (case-insensitive, lenient whitespace):
//
//   - ISO 8601: "2006-01-02", "2006/01/02", "2006.01.02", "2006 01 02"
//   - ISO 8601 with time suffix, the date part is taken:
//     "2006-01-02T15:04:05Z07:00", "30.11.2023 00:00:00"
//   - dd.mm.yyyy / dd.mm.yy: "25.12.1975", "25.12.75"
//   - mm/dd/yyyy / mm/dd/yy: "12/25/1975", "12/25/75"
//   - "1st of 02/2003", "4th of dec. 2020"
//   - Month-first English / German: "jan. 24 2012", "1. Dezember 2016"
//   - Day-first English / German: "7 oct 1970", "1st of dec. 2020"
//   - Month-first with comma: "Oct 7, 1970", "May 17, 2012",
//     "September 17th, 2012"
//   - Short year with leading apostrophe: "Oct 7, '70"
//   - "yyyy-Mon-dd": "2013-Feb-03"
//   - "dd-Mon-yyyy" / "dd-Mon-yy": "07-Feb-2004", "07-Feb-04"
//   - With weekday prefix (dropped): "Mon Jan 02 2006",
//     "Fri, 03 Jul 2015", "Wednesday, 28 Feb 2018"
//   - CJK separators: "2014年04月08日"
//   - French, Italian and Spanish month names and abbreviations:
//     "2 janv. 2019", "23/gen/2019", "12 enero 2020", "2 févr. 2019"
//
// 2-digit years map to 2000..2044 (< 45) or 1900..1999 (>= 45).
//
// Explicitly NOT supported (use a different parser, e.g. time.Parse):
//
//   - Year-only or year-month-only strings ("2014", "201412",
//     "20140601", "2014-04", "2014.05")
//   - Single-digit day-month-year combinations under 5 total digits
//     ("8/8/71") — too ambiguous to disambiguate reliably
//   - RubyDate / RFC1123 with embedded time and timezone offset
//     ("Mon Jan 02 15:04:05 -0700 2006")
//   - Locale-specific written forms beyond the abbreviations above
//     (e.g. Arabic numerals, full Russian months)
func Normalize(str string, lang ...language.Code) (Date, error) {
	return Date(str).NormalizedLang(getLangHint(lang))
}

// StringIsDate returns if a string can be parsed as Date.
// The first given lang argument is used as language hint for parsing.
func StringIsDate(str string, lang ...language.Code) bool {
	_, err := Normalize(str, lang...)
	return err == nil
}

const (
	// Layout used for the Date type, compatible with time.Time.Format()
	Layout = time.DateOnly

	// Regular expression for the Layout
	Regex = `\d{4}-\d{2}-\d{2}`

	Length = 10 // len("2006-01-02")

	// MinLength is the minimum length of a valid date
	MinLength = 6

	// Invalid holds an empty string Date.
	// See NullableDate for where an empty string is a valid value.
	Invalid Date = ""
)

// Date represents a calendar date in ISO 8601 format (YYYY-MM-DD).
// See NullableDate for a nullable version of this type.
//
// A Date is zero, see IsZero, if it is empty, whitespace only, or one of
// the zero dates "0000-00-00" and "0001-01-01" in any spelling, which have
// no practical use as a date. A zero Date is not valid.
//
// Scan and Value fully normalize, because a database date is not free
// text: a valid date is stored as YYYY-MM-DD, a zero date (see IsZero)
// is "", Value writes SQL NULL for a zero date and returns an error for
// a date that can't be normalized, and a time.Time from the driver is
// its date. JSON, XML, MarshalText and ScanString without validate only
// trim surrounding whitespace, including a zero date, and do not parse
// another spelling into YYYY-MM-DD. ScanString with validate stores the
// normalized date or returns the error of Normalized. fmt prints the date
// as stored. Date has no UnmarshalText, so encoding/json reads and writes
// the keys of a map[Date]V as given. See "Reading and writing values"
// in the README of the module.
type Date string

// Compile-time check that Date implements types.NormalizableValidator[Date]
var _ types.NormalizableValidator[Date] = Date("")

// Must returns str as normalized Date or panics if str is not a valid Date.
func Must(str string) Date {
	d, err := Date(str).Normalized()
	if err != nil {
		panic(err)
	}
	return d
}

// Of returns a normalized Date for the given year, month, and day.
// The month, day values may be outside their usual ranges and will be normalized
// during the conversion. For example, October 32 converts to November 1.
func Of(year int, month time.Month, day int) Date {
	return OfTime(time.Date(year, month, day, 0, 0, 0, 0, time.Local))
}

// OfTime returns the date part of the passed time.Time or an empty string if t.IsZero().
// To get the date in a certain time zone, pass the time.Time with a location set to the time zone.
func OfTime(t time.Time) Date {
	if t.IsZero() {
		return ""
	}
	return Date(t.Format(Layout))
}

// OfTimePtr returns the date part of the passed time.Time or Null (an empty string)
// if t is nil or t.IsZero().
func OfTimePtr(t *time.Time) NullableDate {
	if t == nil || t.IsZero() {
		return Null
	}
	return NullableDate(OfTime(*t))
}

// OfToday returns the date of today in the local timezone.
func OfToday() Date {
	return OfTime(time.Now())
}

// OfNowInUTC returns the date of the current time in the UTC timezone.
func OfNowInUTC() Date {
	return OfTime(time.Now().UTC())
}

// OfTodayIn returns the date of today in the timezone of the passed location.
func OfTodayIn(loc *time.Location) Date {
	return OfTime(time.Now().In(loc))
}

// OfYesterday returns the date of yesterday in the local timezone.
func OfYesterday() Date {
	return OfTime(time.Now().Add(-24 * time.Hour))
}

// OfTomorrow returns the date of tomorrow in the local timezone.
func OfTomorrow() Date {
	return OfTime(time.Now().Add(24 * time.Hour))
}

// Parse returns the date part from time.Parse(layout, value).
func Parse(layout, value string) (Date, error) {
	t, err := time.Parse(layout, value)
	if t.IsZero() || err != nil {
		return "", err
	}
	return OfTime(t), nil
}

// PeriodRange returns the dates [from, until] for a period defined in one of the following formats:
//   - period of a ISO 8601 week of a year: YYYY-Wnn
//   - period of a month of a year: YYYY-MM
//   - period of a quarter of a year: YYYY-Qn
//   - period of a half year: YYYY-Hn
//   - period of full year: YYYY
//
// The returned from Date is the first day of the period,
// the returned until Date is the last day of the period.
// For a week these are the Monday and the Sunday of the ISO 8601 week,
// which can fall into the previous or next year.
//
// Examples:
//   - Period of June 2018: PeriodRange("2018-06") == Date("2018-06-01"), Date("2018-06-30"), nil
//   - Period of Q3 2018: PeriodRange("2018-Q3") == Date("2018-07-01"), Date("2018-09-30"), nil
//   - Period of second half of 2018: PeriodRange("2018-H2") == Date("2018-07-01"), Date("2018-12-31"), nil
//   - Period of year 2018: PeriodRange("2018") == Date("2018-01-01"), Date("2018-12-31"), nil
//   - Period of week 1 2019: PeriodRange("2019-W01") == Date("2018-12-31"), Date("2019-01-06"), nil
func PeriodRange(period string) (from, until Date, err error) {
	if len(period) != 4 && len(period) != 7 && len(period) != 8 {
		return "", "", fmt.Errorf("invalid period format length: %q", period)
	}

	if len(period) == 4 {
		year, err := strconv.Atoi(period)
		if err != nil || year <= 0 {
			return "", "", fmt.Errorf("invalid period format: %q", period)
		}
		from = Date(period + "-01-01")
		until = Date(period + "-12-31")
		return from, until, nil
	}

	if period[4] != '-' {
		return "", "", fmt.Errorf("invalid period format, expected '-' after year: %q", period)
	}

	year, err := strconv.Atoi(period[:4])
	if err != nil {
		return "", "", fmt.Errorf("invalid period format, can't parse year: %q", period)
	}

	switch period[5] {
	case 'W', 'w':
		week, err := strconv.Atoi(period[6:])
		if err != nil || week < 1 || week > 53 {
			return "", "", fmt.Errorf("invalid period format, can't parse week: %q", period)
		}
		// December 28th is always in the last ISO week of its year,
		// a year with 52 weeks has no week 53
		if _, lastWeek := time.Date(year, 12, 28, 0, 0, 0, 0, time.UTC).ISOWeek(); week > lastWeek {
			return "", "", fmt.Errorf("invalid period, year %d has no ISO week %d: %q", year, week, period)
		}
		from, until = YearWeekRange(year, week)
		return from, until, nil

	case 'Q', 'q':
		quarter, err := strconv.Atoi(period[6:])
		if err != nil || quarter < 1 || quarter > 4 {
			return "", "", fmt.Errorf("invalid period format, can't parse quarter: %q", period)
		}
		from = Of(year, time.Month(quarter-1)*3+1, 1)
		until = Of(year, time.Month(quarter)*3+1, 0) // 0th day is the last day of the previous month
		return from, until, nil

	case 'H', 'h':
		half, err := strconv.Atoi(period[6:])
		if err != nil || half < 1 || half > 2 {
			return "", "", fmt.Errorf("invalid period format, can't parse half-year: %q", period)
		}
		from = Of(year, time.Month(half-1)*6+1, 1)
		until = Of(year, time.Month(half)*6+1, 0) // 0th day is the last day of the previous month
		return from, until, nil
	}

	month, err := strconv.Atoi(period[5:])
	if err != nil || month < 1 || month > 12 {
		return "", "", fmt.Errorf("invalid period format, can't parse month: %q", period)
	}

	from = Of(year, time.Month(month), 1)
	until = Of(year, time.Month(month)+1, 0) // 0th day is the last day of the previous month
	return from, until, nil
}

// YearRange returns the date range from
// first of January to 31st of December of a year.
func YearRange(year int) (from, until Date) {
	yyyy := fmt.Sprintf("%04d", year)
	return Date(yyyy + "-01-01"), Date(yyyy + "-12-31")
}

// YearWeekMonday returns the date of Monday of an ISO 8601 week.
func YearWeekMonday(year, week int) (monday Date) {
	// January 4th is always in week 1 of the ISO 8601 year
	t := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)

	// Go back to Monday of week 1
	daysSinceMonday := (int(t.Weekday()) + 6) % 7
	t = t.AddDate(0, 0, -daysSinceMonday)

	// Add week days
	t = t.AddDate(0, 0, (week-1)*7)

	return OfTime(t)
}

// YearWeekRange returns the dates of Monday and Sunday of an ISO 8601 week.
func YearWeekRange(year, week int) (monday, sunday Date) {
	monday = YearWeekMonday(year, week)
	sunday = monday.AddDays(6)
	return monday, sunday
}

// FromUntilFromYearAndMonths returns the first and last date of a month range
// within a given year. The year argument must be a 4-digit year string (e.g. "2024").
// The months argument is either a single month number ("3") or a dash-separated
// range ("3-9"); an empty months string defaults to the full year ("1-12").
// Returns empty dates without an error when year is empty.
func FromUntilFromYearAndMonths(year, months string) (fromDate, untilDate Date, err error) {
	if year == "" {
		return "", "", nil
	}
	if months == "" {
		months = "1-12"
	}
	parts := strings.Split(months, "-")
	fromMonth := parts[0]
	untilMonth := parts[0]
	if len(parts) == 2 {
		untilMonth = parts[1]
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		return "", "", err
	}
	fromMonthInt, err := strconv.Atoi(fromMonth)
	if err != nil {
		return "", "", err
	}
	untilMonthInt, err := strconv.Atoi(untilMonth)
	if err != nil {
		return "", "", err
	}
	fromDate = Of(yearInt, time.Month(fromMonthInt), 1)
	untilDate = Of(yearInt, time.Month(untilMonthInt+1), 0) // 0th day is the last day of the previous month

	return fromDate, untilDate, nil
}

// ScanString implements the strfmt.Scannable interface.
// Without validate it trims surrounding whitespace, including a zero
// date. With validate it stores the normalized date or returns the
// error of Normalized, also for a blank or zero source.
func (date *Date) ScanString(source string, validate bool) error {
	if validate {
		return internal.ScanStringNormalized(date, source, true, false, Date.Normalized, Date.IsZero)
	}
	return internal.ScanStringNormalized(date, source, false, false, internal.Trim[Date], nil)
}

// ScanStringWithLang parses and assigns the source string as a Date using the given language hint.
// Returns wasNormalized to indicate if the source was already in normalized form,
// monthMustBeFirst to indicate if the date format requires month-first ordering,
// and an error if parsing fails, also for a blank source. A zero date in
// any spelling, see IsZero, is assigned as "" like by Scan.
// monthMustBeFirst is only true if month first is the only valid reading,
// like for "12/25/2023", not for an ambiguous date like "05/06/2024"
// that is read month first because of its slashes or the hint.
func (date *Date) ScanStringWithLang(source string, lang language.Code) (wasNormalized bool, monthMustBeFirst bool, err error) {
	if strutil.TrimSpace(source) != "" && Date(source).IsZero() {
		*date = ""
		return false, false, nil
	}
	newDate, monthMustBeFirst, err := normalizeAndCheckDate(source, lang)
	if err != nil {
		return false, false, err
	}
	*date = newDate
	return newDate == Date(source), monthMustBeFirst, nil
}

// WithinIncl returns true if the date is within the inclusive range [from, until].
// Returns false if any of the dates is not valid, see After.
func (date Date) WithinIncl(from, until Date) bool {
	t, valid := date.midnightInLocation(time.UTC)
	tFrom, fromValid := from.midnightInLocation(time.UTC)
	tUntil, untilValid := until.midnightInLocation(time.UTC)
	if !valid || !fromValid || !untilValid {
		return false
	}
	return (t.Equal(tFrom) || t.After(tFrom)) && (t.Equal(tUntil) || t.Before(tUntil))
}

// BetweenExcl returns true if the date is strictly between (exclusive) after and before.
// Returns false if any of the dates is not valid, see After.
func (date Date) BetweenExcl(after, before Date) bool {
	t, valid := date.midnightInLocation(time.UTC)
	tAfter, afterValid := after.midnightInLocation(time.UTC)
	tBefore, beforeValid := before.midnightInLocation(time.UTC)
	if !valid || !afterValid || !beforeValid {
		return false
	}
	return t.After(tAfter) && t.Before(tBefore)
}

// Nullable returns the date as a NullableDate.
// Returns Null if the date is zero, otherwise returns the date as NullableDate.
func (date Date) Nullable() NullableDate {
	if date.IsZero() {
		return Null
	}
	return NullableDate(date)
}

// zeroDate is the normalized form of the zero value of time.Time
const zeroDate Date = "0001-01-01"

// IsZero returns true if the date is empty or whitespace only, or one of the
// zero dates "0000-00-00" and "0001-01-01" in any spelling, like "01.01.0001",
// "0001-01-01T00:00:00Z", "00.00.0000" or "0000-00-00 00:00:00".
// "0001-01-01" is the zero value of time.Time, "0000-00-00" may be the zero
// value of other date implementations, like MySQL.
// A zero date has no practical use as a date and is not a valid Date.
// SQL Scan stores it as "" and Value writes SQL NULL. JSON and XML keep
// the trimmed text. IsNull is still true for that text, so Get panics.
func (date Date) IsZero() bool {
	switch date {
	case "", zeroDate, "0000-00-00":
		return true
	}
	if strutil.TrimSpace(date) == "" {
		return true
	}
	// Every spelling of a zero date has the year 0000 or 0001 in its date
	// part, so other dates are told apart without parsing them, which the
	// reading methods would do again to normalize the date. The time part
	// is left out, because its fractional seconds often contain "0000".
	if part := datePart(string(date)); !strings.Contains(part, "0000") && !strings.Contains(part, "0001") {
		return false
	}
	if isZerosDate(string(date)) {
		return true
	}
	norm, _, err := normalizeAndCheckDate(string(date), "")
	return err == nil && norm == zeroDate
}

// isZerosDate reports if the date part of str, see datePart, is the zero
// date "0000-00-00" in any order of its parts and with any separators,
// like "00.00.0000", which normalizeDate rejects because it has no valid
// month and day.
func isZerosDate(str string) bool {
	parts := strings.FieldsFunc(datePart(str), isDateSeparatorRune)
	if len(parts) != 3 {
		return false
	}
	years := 0
	for _, part := range parts {
		if strings.Trim(part, "0") != "" {
			return false
		}
		switch len(part) {
		case 4:
			years++
		case 1, 2:
		default:
			return false
		}
	}
	return years == 1
}

// Validate returns an error if the date is not in a valid, normalizable format.
func (date Date) Validate() error {
	_, err := date.Normalized()
	return err
}

// Valid returns true if the date is in a valid, normalizable format.
func (date Date) Valid() bool {
	return date.Validate() == nil
}

// ValidAndNormalized returns true if the date is both valid and already in normalized form (YYYY-MM-DD).
func (date Date) ValidAndNormalized() bool {
	// The shape check saves the failed time.Parse of every other format,
	// which NormalizedLang calls first and which allocates its error
	if len(date) != len(Layout) || date[4] != '-' || date[7] != '-' || date == zeroDate {
		return false
	}
	t, err := time.Parse(Layout, string(date))
	// time.Parse accepts the year 0, which Valid rejects
	return err == nil && validYear(t.Year())
}

// Time returns a time.Time for the date at the specified hour, minute, and second in the given location.
// Returns a zero time.Time if the date is zero.
// An invalid date is not detected and results in a meaningless time
// before year 1 (from time.Date(0, 0, 0, ...)), use Valid to check the date first.
func (date Date) Time(hour, minute, second int, location *time.Location) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	year, month, day := date.YearMonthDay()
	return time.Date(year, month, day, hour, minute, second, 0, location)
}

// TimeLocal returns a time.Time for the date at the specified hour, minute, and second in local time.
// Returns a zero time.Time if the date is zero, see Time for invalid dates.
func (date Date) TimeLocal(hour, minute, second int) time.Time {
	return date.Time(hour, minute, second, time.Local)
}

// TimeUTC returns a time.Time for the date at the specified hour, minute, and second in UTC.
// Returns a zero time.Time if the date is zero, see Time for invalid dates.
func (date Date) TimeUTC(hour, minute, second int) time.Time {
	return date.Time(hour, minute, second, time.UTC)
}

// MidnightUTC returns the midnight (00:00) time.Time of the date in UTC,
// or a zero time.Time value if the date is zero or not valid.
// A valid date that is not in the normalized form is normalized first.
func (date Date) MidnightUTC() time.Time {
	return date.MidnightInLocation(time.UTC)
}

// Midnight returns the midnight (00:00) time.Time of the date in the local time zone.
// Returns a zero time.Time if the date is not valid.
func (date Date) Midnight() time.Time {
	return date.MidnightInLocation(time.Local)
}

// MidnightInLocation returns the midnight (00:00) time.Time of the date in the given location.
// Returns a zero time.Time if the date is zero or not valid.
// A valid date that is not in the normalized form is normalized first.
func (date Date) MidnightInLocation(loc *time.Location) time.Time {
	t, _ := date.midnightInLocation(loc)
	return t
}

// midnightInLocation returns MidnightInLocation(loc) and Valid()
// with a single parse for a date in the normalized form.
func (date Date) midnightInLocation(loc *time.Location) (t time.Time, valid bool) {
	if date.IsZero() {
		// A zero date is no date, so it is not valid
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(Layout, string(date), loc)
	if err != nil {
		// Not in the normalized YYYY-MM-DD form, but might still be valid
		norm, err := date.Normalized()
		if err != nil {
			return time.Time{}, false
		}
		return norm.midnightInLocation(loc)
	}
	// time.Parse accepts the year 0 that Normalized rejects
	if !validYear(t.Year()) {
		return time.Time{}, false
	}
	return t, true
}

// Format formats the date using the given layout string (see time.Time.Format).
// Returns an empty string if the date is zero (see IsZero) or layout is empty.
// If layout equals Layout constant, returns the normalized date
// without a time conversion.
// A date that is not valid is returned unchanged for every layout
// instead of formatting the zero time.Time like "01.01.0001".
func (date Date) Format(layout string) string {
	if date.IsZero() || layout == "" {
		return ""
	}
	if layout == Layout {
		return string(date.NormalizedOrUnchanged())
	}
	t, ok := date.midnightInLocation(time.UTC)
	if !ok {
		return string(date)
	}
	return t.Format(layout)
}

// NormalizedOrUnchanged returns the normalized date if possible, otherwise returns the date unchanged.
// The first given lang argument is used as language hint for parsing.
func (date Date) NormalizedOrUnchanged(lang ...language.Code) Date {
	normalized, err := date.NormalizedLang(getLangHint(lang))
	if err != nil {
		return date
	}
	return normalized
}

// NormalizedOrNull returns the normalized date as NullableDate if possible, otherwise returns Null.
// The first given lang argument is used as language hint for parsing.
func (date Date) NormalizedOrNull(lang ...language.Code) NullableDate {
	normalized, err := date.NormalizedLang(getLangHint(lang))
	if err != nil {
		return Null
	}
	return NullableDate(normalized)
}

// NormalizedEqual returns true if two dates are equal after normalization.
// An invalid date (including "") has no normalized value and is not equal
// to any date, not even to an identical invalid string.
func (date Date) NormalizedEqual(other Date) bool {
	a, err := date.Normalized()
	if err != nil {
		return false
	}
	b, err := other.Normalized()
	return err == nil && a == b
}

// Compare compares the date with another Date.
// Returns -1 if date is before other, +1 if after, 0 if equal.
// An invalid date compares like the empty string: equal to any other
// invalid date and before every valid date, so Compare is a consistent
// sort order for slices.SortFunc also for invalid dates.
func (date Date) Compare(other Date) int {
	return strings.Compare(string(date.NormalizedOrNull()), string(other.NormalizedOrNull()))
}

// After returns true if the date is after the other date.
// An invalid date (including "") has no position in time,
// so After returns false if either date is not valid,
// like NormalizedEqual returns false for an invalid date.
func (date Date) After(other Date) bool {
	t, valid := date.midnightInLocation(time.UTC)
	tOther, otherValid := other.midnightInLocation(time.UTC)
	return valid && otherValid && t.After(tOther)
}

// EqualOrAfter returns true if the date is equal to or after the other date.
func (date Date) EqualOrAfter(other Date) bool {
	return date.NormalizedEqual(other) || date.After(other)
}

// Before returns true if the date is before the other date.
// Returns false if either date is not valid, see After.
func (date Date) Before(other Date) bool {
	t, valid := date.midnightInLocation(time.UTC)
	tOther, otherValid := other.midnightInLocation(time.UTC)
	return valid && otherValid && t.Before(tOther)
}

// EqualOrBefore returns true if the date is equal to or before the other date.
func (date Date) EqualOrBefore(other Date) bool {
	return date.NormalizedEqual(other) || date.Before(other)
}

// AfterTime returns true if midnight of the date (in the location of the passed time) is after the time.
// Returns false if the date is not valid, see After.
func (date Date) AfterTime(other time.Time) bool {
	t, valid := date.midnightInLocation(other.Location())
	return valid && t.After(other)
}

// BeforeTime returns true if midnight of the date (in the location of the passed time) is before the time.
// Returns false if the date is not valid, see After.
func (date Date) BeforeTime(other time.Time) bool {
	t, valid := date.midnightInLocation(other.Location())
	return valid && t.Before(other)
}

// AddDate returns a new date with the specified years, months, and days added.
// The month and day values may be outside their usual ranges and will be normalized.
func (date Date) AddDate(years int, months int, days int) Date {
	return OfTime(date.MidnightUTC().AddDate(years, months, days))
}

// AddYears returns a new date with the specified number of years added.
func (date Date) AddYears(years int) Date {
	return OfTime(date.Midnight().AddDate(years, 0, 0))
}

// AddMonths returns a new date with the specified number of months added.
func (date Date) AddMonths(months int) Date {
	return OfTime(date.Midnight().AddDate(0, months, 0))
}

// AddDays returns a new date with the specified number of days added.
func (date Date) AddDays(days int) Date {
	return OfTime(date.Midnight().AddDate(0, 0, days))
}

// Add returns a new date with the specified duration added.
func (date Date) Add(d time.Duration) Date {
	return OfTime(date.MidnightUTC().Add(d))
}

// Sub returns the duration between the date and the other date (date - other).
func (date Date) Sub(other Date) time.Duration {
	return date.MidnightUTC().Sub(other.MidnightUTC())
}

// BeginningOfWeek returns the date of the first day of the week containing this date.
// The startDay parameter specifies which day is considered the start of the week.
// For example, time.Monday for ISO 8601 weeks, or time.Sunday for US-style weeks.
func (date Date) BeginningOfWeek(startDay time.Weekday) Date {
	t := date.MidnightUTC()
	// Calculate days to go back to reach the week start day
	daysBack := int(t.Weekday() - startDay)
	if daysBack < 0 {
		daysBack += 7
	}
	return OfTime(t.AddDate(0, 0, -daysBack))
}

// BeginningOfMonth returns the date of the first day of the month containing this date.
func (date Date) BeginningOfMonth() Date {
	y, m, _ := date.YearMonthDay()
	return Of(y, m, 1)
}

// BeginningOfQuarter returns the date of the first day of the quarter containing this date.
func (date Date) BeginningOfQuarter() Date {
	y, m, _ := date.YearMonthDay()
	// Calculate the first month of the quarter
	offset := (int(m) - 1) % 3
	return Of(y, m, 1).AddMonths(-offset)
}

// BeginningOfYear returns the date of the first day of the year containing this date.
func (date Date) BeginningOfYear() Date {
	y, _, _ := date.YearMonthDay()
	return Of(y, time.January, 1)
}

// EndOfWeek returns the date of the last day of the week containing this date.
// The weekday parameter specifies which day is considered the start of the week.
// The end of the week is calculated as 6 days after the beginning.
// For example, if weekday is time.Monday, the end will be Sunday.
func (date Date) EndOfWeek(weekday time.Weekday) Date {
	return date.BeginningOfWeek(weekday).AddDays(6)
}

// EndOfMonth returns the date of the last day of the month containing this date.
func (date Date) EndOfMonth() Date {
	return date.BeginningOfMonth().AddMonths(1).AddDays(-1)
}

// EndOfQuarter returns the date of the last day of the quarter containing this date.
func (date Date) EndOfQuarter() Date {
	return date.BeginningOfQuarter().AddMonths(3).AddDays(-1)
}

// EndOfYear returns the date of the last day of the year containing this date.
func (date Date) EndOfYear() Date {
	y, _, _ := date.YearMonthDay()
	return Of(y, time.December, 31)
}

// LastMonday returns the date of the Monday on or before this date.
func (date Date) LastMonday() Date {
	t := date.MidnightUTC()
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return OfTime(t.AddDate(0, 0, -(weekday - 1)))
}

// NextSunday returns the date of the Sunday on or after this date.
func (date Date) NextSunday() Date {
	t := date.MidnightUTC()
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return OfTime(t.AddDate(0, 0, 7-weekday))
}

// YearMonthDay returns the year, month, and day components of the date.
// Returns zero values if the date is not valid.
func (date Date) YearMonthDay() (year int, month time.Month, day int) {
	norm, err := date.Normalized()
	if err != nil {
		return 0, 0, 0
	}
	year, _ = strconv.Atoi(string(norm)[:4])
	monthInt, _ := strconv.Atoi(string(norm)[5:7])
	day, _ = strconv.Atoi(string(norm)[8:])
	return year, time.Month(monthInt), day
}

// Year returns the year component of the date.
// Returns 0 if the date is not valid.
func (date Date) Year() int {
	norm, err := date.Normalized()
	if err != nil {
		return 0
	}
	year, _ := strconv.Atoi(string(norm)[:4])
	return year
}

// Month returns the month component of the date.
// Returns 0 if the date is not valid.
func (date Date) Month() time.Month {
	norm, err := date.Normalized()
	if err != nil {
		return 0
	}
	monthInt, _ := strconv.Atoi(string(norm)[5:7])
	return time.Month(monthInt)
}

// Day returns the day of the month component of the date.
// Returns 0 if the date is not valid.
func (date Date) Day() int {
	norm, err := date.Normalized()
	if err != nil {
		return 0
	}
	day, _ := strconv.Atoi(string(norm)[8:])
	return day
}

// Weekday returns the day of the week for this date.
// Returns 0 (Sunday) if the date is not valid.
func (date Date) Weekday() time.Weekday {
	t := date.Midnight()
	if t.IsZero() {
		return 0
	}
	return t.Weekday()
}

// ISOWeek returns the ISO 8601 year and week number in which the date occurs.
// Week ranges from 1 to 53. Jan 01 to Jan 03 of year n might belong to
// week 52 or 53 of year n-1, and Dec 29 to Dec 31 might belong to week 1
// of year n+1.
// Returns zeros if the date is not valid.
func (date Date) ISOWeek() (year, week int) {
	t := date.MidnightUTC()
	if t.IsZero() {
		return 0, 0
	}
	return t.ISOWeek()
}

// IsToday returns true if the date is today in local time.
// A valid date that is not in the normalized form is normalized first.
func (date Date) IsToday() bool {
	// OfToday is normalized, so only the receiver needs normalization
	norm, err := date.Normalized()
	return err == nil && norm == OfToday()
}

// IsTodayInUTC returns true if the date is today in UTC time.
// A valid date that is not in the normalized form is normalized first.
func (date Date) IsTodayInUTC() bool {
	// OfNowInUTC is normalized, so only the receiver needs normalization
	norm, err := date.Normalized()
	return err == nil && norm == OfNowInUTC()
}

// AfterToday returns true if the date is after today in local time.
func (date Date) AfterToday() bool {
	return date.After(OfToday())
}

// AfterTodayInUTC returns true if the date is after today in UTC time.
func (date Date) AfterTodayInUTC() bool {
	return date.After(OfNowInUTC())
}

// BeforeToday returns true if the date is before today in local time.
func (date Date) BeforeToday() bool {
	return date.Before(OfToday())
}

// BeforeTodayInUTC returns true if the date is before today in UTC time.
func (date Date) BeforeTodayInUTC() bool {
	return date.Before(OfNowInUTC())
}

// Scan implements the database/sql.Scanner interface.
// A valid date from a string or []byte is stored normalized, SQL NULL
// and a zero date (see IsZero) as "", and a date that can't be normalized
// unchanged without an error, so that one invalid value doesn't fail a
// whole query; Validate reports it. A time.Time is stored as OfTime,
// which is already YYYY-MM-DD or "" for the zero time, without parsing
// it again.
func (date *Date) Scan(value any) error {
	if t, ok := value.(time.Time); ok {
		*date = OfTime(t)
		return nil
	}
	return internal.ScanNormalized(date, value, Date.Normalized, Date.IsZero)
}

// Value implements the database/sql/driver.Valuer interface.
// It returns the normalized date, SQL NULL for a zero date (see IsZero),
// or the error of Normalized for a date that can't be normalized.
func (date Date) Value() (driver.Value, error) {
	return internal.ValueNormalized(date, Date.Normalized, Date.IsZero)
}

// MarshalJSON implements encoding/json.Marshaler by writing the date
// with surrounding whitespace trimmed. A zero date is written as that
// text, unlike Value. A blank date is "".
func (date Date) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(date, false, internal.Trim[Date], nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored with surrounding whitespace trimmed, including
// a zero date, unlike Scan. JSON null leaves the Date unchanged like
// encoding/json does for a string type, and any other JSON value is an error.
func (date *Date) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(date, data, false, internal.Trim[Date], nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// so that encoding/xml and strfmt write the trimmed date.
// encoding/json writes map keys of string types as they are
// and doesn't call MarshalText for them.
func (date Date) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(date, internal.Trim[Date], nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (date *Date) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(date, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like UnmarshalJSON stores it.
func (date *Date) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(date, d, start, internal.Trim[Date], nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like UnmarshalJSON stores it.
func (date *Date) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(date, attr, internal.Trim[Date], nil)
}

// JSONSchema returns the JSON schema definition for the Date type.
// Implements the jsonschema.JSONSchemaProvider interface.
func (Date) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:  "Date",
		Type:   "string",
		Format: "date",
	}
}

func isDateSeparatorRune(r rune) bool {
	// "'" splits a leading apostrophe off short years like '70.
	return unicode.IsSpace(r) || r == '.' || r == '/' || r == '-' ||
		r == ',' || r == '\'' || r == '年' || r == '月' || r == '日'
}

func isDateTrimRune(r rune) bool {
	return unicode.IsSpace(r) || (!unicode.IsLetter(r) && !unicode.IsDigit(r))
}

// Normalized returns the date in normalized form,
// or an error if the format can't be detected or the date is zero,
// see IsZero. Returns the Date unchanged in case of an error.
// Normalized implements types.Normalizable[Date],
// use NormalizedLang to pass a language hint.
func (date Date) Normalized() (Date, error) {
	return date.NormalizedLang("")
}

// NormalizedLang returns the date in normalized form like Normalized,
// with lang as hint for the order of day and month in an ambiguous
// numeric date with slashes: language.EN reads "05/06/2024" month first
// and every other language day first. Other separators, like in
// "05.06.2024", are read day first with any hint.
// An empty lang or one that can't be normalized gives no hint,
// then slashes are read month first, see Normalize.
// Returns the Date unchanged in case of an error.
func (date Date) NormalizedLang(lang language.Code) (Date, error) {
	if date.ValidAndNormalized() {
		// Every reading and writing method normalizes, mostly a date
		// that already is, which this checks without normalizeDate
		return date, nil
	}
	normalized, _, err := normalizeAndCheckDate(string(date), langHint(lang))
	if err != nil {
		return date, err
	}
	if normalized == zeroDate {
		return date, errs.Errorf("zero date: %q", string(date))
	}
	return normalized, nil
}

func normalizeAndCheckDate(str string, langHint language.Code) (Date, bool, error) {
	normalized, monthMustBeFirst, err := normalizeDate(str, langHint)
	if err != nil {
		return "", monthMustBeFirst, err
	}
	t, err := time.Parse(Layout, normalized)
	if err != nil {
		return "", monthMustBeFirst, err
	}
	// time.Parse accepts the year 0, which Date rejects
	if !validYear(t.Year()) {
		return "", monthMustBeFirst, errs.Errorf("invalid year 0 in date: %q", str)
	}
	return Date(normalized), monthMustBeFirst, nil
}

// datePart returns str without the surrounding characters that are
// no letters or digits, and without its time part: the time of an
// ISO 8601 date-time like "2006-01-02T15:04:05Z", or "00:00:00".
func datePart(str string) string {
	trimmed := strings.TrimFunc(str, isDateTrimRune)
	if len(trimmed) > 11 && trimmed[10] == 'T' && trimmed[11] >= '0' && trimmed[11] <= '9' {
		// Use date part of this date-time format: "2006-01-02T15:04:05",
		// but not of a word like "2023-12-25TOTAL". Cut before trimming
		// "00:00:00", which would leave "2023-12-25T" of "2023-12-25T00:00:00".
		return trimmed[:10]
	}
	trimmed = strings.TrimSuffix(trimmed, "00:00:00") // Trim zero time part
	return strings.TrimFunc(trimmed, isDateTrimRune)
}

func normalizeDate(str string, lang language.Code) (string, bool, error) {
	trimmed := datePart(str)
	if len(trimmed) < MinLength {
		return "", false, fmt.Errorf("too short for a date: %q", str)
	}

	trimmed = strings.ToLower(trimmed)

	parts := strings.FieldsFunc(trimmed, isDateSeparatorRune)
	if len(parts) == 4 {
		// Drop the English filler "of" ("1st of Dec 2020").
		if i := strutil.IndexInStrings("of", parts); i > 0 && i <= 2 {
			parts = append(parts[:i], parts[i+1:]...)
		} else if _, isWeekday := weekdayNameSet[parts[0]]; isWeekday {
			// Drop a leading weekday name ("Mon Jan 02 2006",
			// "Fri, 03 Jul 2015", "Wednesday, 28 Feb 2018").
			parts = parts[1:]
		}
	}
	if len(parts) != 3 {
		return "", false, fmt.Errorf("date must have 3 parts: %q", str)
	}
	dayHint := -1
	totalLen := 0
	// A single digit is padded to two, but can't be a year,
	// so that text like "set 6 4" is not the date 2004-09-06
	var singleDigit [3]bool
	for i := range parts {
		l := len(parts[i])
		totalLen += l
		if l == 1 {
			parts[i] = "0" + parts[i]
			singleDigit[i] = true
			continue
		}
		// Strip English ordinal suffixes (1st, 2nd, 3rd, 4th, 21st,
		// 22nd, 23rd, 31st, …). The suffix is always at the end and
		// the preceding characters must be digits, so that a month
		// name like "august" keeps its "st".
		for _, suf := range [4]string{"st", "nd", "rd", "th"} {
			if before, ok := strings.CutSuffix(parts[i], suf); ok && before != "" && strings.Trim(before, "0123456789") == "" {
				parts[i] = before
				if len(parts[i]) == 1 {
					parts[i] = "0" + parts[i]
				}
				dayHint = i
				break
			}
		}
	}
	if totalLen < 5 {
		return "", false, fmt.Errorf("date is too short: %q", str)
	}

	len0 := len(parts[0])
	len1 := len(parts[1])
	len2 := len(parts[2])
	val0, _ := strconv.Atoi(parts[0])
	val1, _ := strconv.Atoi(parts[1])
	val2, _ := strconv.Atoi(parts[2])
	month0, _ := monthNameMap[parts[0]]
	month1, _ := monthNameMap[parts[1]]
	month2, _ := monthNameMap[parts[2]]

	expandVal2ToFullYear := func() bool {
		if len2 != 2 {
			panic("len2")
		}
		if singleDigit[2] {
			return false
		}
		if val2 < 45 {
			parts[2] = "20" + parts[2]
			val2 = 2000 + val2
		} else {
			parts[2] = "19" + parts[2]
			val2 = 1900 + val2
		}
		len2 = 4
		return true
	}

	switch {
	case month0 != 0:
		if len2 == 2 && !expandVal2ToFullYear() {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		if !validDay(val1) || !validYear(val2) {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		// m DD YYYY
		return fmt.Sprintf("%s-%02d-%s", parts[2], month0, parts[1]), false, nil

	case len0 == 2 && month1 != 0 && len2 == 2:
		if (!validDay(val0) && validDay(val2)) || dayHint == 2 {
			// YY m DD
			parts[0], parts[2] = parts[2], parts[0]
			val0, val2 = val2, val0
			singleDigit[0], singleDigit[2] = singleDigit[2], singleDigit[0]
			// DD m YY
		}
		if !expandVal2ToFullYear() {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		fallthrough

	case len0 == 2 && month1 != 0 && len2 == 4:
		if !validDay(val0) || !validYear(val2) {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		// DD m YYYY
		return fmt.Sprintf("%s-%02d-%s", parts[2], month1, parts[0]), false, nil

	case len0 == 4 && month1 != 0 && len2 == 2:
		if !validYear(val0) || !validDay(val2) {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		// YYYY m DD
		return fmt.Sprintf("%s-%02d-%s", parts[0], month1, parts[2]), false, nil

	case month2 != 0:
		if singleDigit[0] {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		if len0 == 2 {
			// YY DD m
			if val0 < 45 {
				parts[0] = "20" + parts[0]
				val0 = 2000 + val0
			} else {
				parts[0] = "19" + parts[0]
				val0 = 1900 + val0
			}
			len0 = 4
		}
		// YYYY DD m
		return fmt.Sprintf("%s-%02d-%s", parts[0], month2, parts[1]), false, nil

	case len0 == 4 && len1 == 2 && len2 == 2:
		if !validYear(val0) || !validMonth(val1) || !validDay(val2) {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		return strings.Join(parts, "-"), false, nil

	case len0 == 2 && len1 == 2 && len2 == 2:
		if !expandVal2ToFullYear() {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		fallthrough

	case len0 == 2 && len1 == 2 && len2 == 4:
		monthMustBeFirst := validMonth(val0) && !validMonth(val1)
		// The ordinal day hint, the language hint and the separators
		// only decide the order if both numbers could be the month
		ambiguous := validMonth(val0) && validMonth(val1)
		monthFirst := monthMustBeFirst || (ambiguous && ambiguousMonthFirst(trimmed, dayHint, lang))
		// A number with an ordinal suffix can't be the month
		if (monthFirst && dayHint == 0) || (!monthFirst && dayHint == 1) {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		if monthFirst {
			// MM DD YYYY
			parts[0], parts[1] = parts[1], parts[0]
			val0, val1 = val1, val0
			// DD MM YYYY
		}
		if !validDay(val0) || !validMonth(val1) || !validYear(val2) {
			return "", false, fmt.Errorf("invalid date: %q", str)
		}
		// DD MM YYYY
		parts[0], parts[2] = parts[2], parts[0]
		// YYYY MM DD
		return strings.Join(parts, "-"), monthMustBeFirst, nil
	}

	return "", false, fmt.Errorf("invalid date: %q", str)
}

// ambiguousMonthFirst reports if a numeric date with the year last,
// whose first two numbers can both be the month, is read month first.
// str is the date part of the date, see datePart, and dayHint the index
// of the number with an ordinal suffix, or -1.
//
// An ordinal suffix marks the day explicitly, so it decides first.
// Then the separators decide: "03.04.2024" with dots is German style and
// read day first, and so are other separators like "03-04-2024", which
// have no month first convention, also with an English hint.
// Only "03/04/2024" with slashes is written in both orders, US month
// first and British day first, so for slashes a language hint decides,
// because it knows the convention of the whole document: English reads
// month first, every other language day first. Without a hint slashes
// are read month first, because JSON, XML and SQL can't pass a hint and
// read a US date day first as 3 April otherwise.
func ambiguousMonthFirst(str string, dayHint int, lang language.Code) bool {
	if dayHint != -1 {
		return dayHint == 1
	}
	if !slashSeparated(str) {
		return false
	}
	if hint := langHint(lang); hint != "" {
		return hint == language.EN
	}
	return true
}

// slashSeparated reports if both separators between the last three
// parts of the date str have a '/', like in "03/04/2024", "3 / 4 / 24",
// "03/04/'24" and "Mon, 03/04/2024", but not "03/04.2024".
func slashSeparated(str string) bool {
	for range 2 {
		// Cut the last part, then the separators before it
		withSeparators := strings.TrimRightFunc(str, isNoDateSeparatorRune)
		str = strings.TrimRightFunc(withSeparators, isDateSeparatorRune)
		if !strings.Contains(withSeparators[len(str):], "/") {
			return false
		}
	}
	return true
}

func isNoDateSeparatorRune(r rune) bool {
	return !isDateSeparatorRune(r)
}

func validYear(year int) bool {
	return year > 0
}

func validMonth(month int) bool {
	return month >= 1 && month <= 12
}

func validDay(day int) bool {
	return day >= 1 && day <= 31
}

// getLangHint returns the first hint normalized by langHint,
// or "" if there is none.
func getLangHint(lang []language.Code) language.Code {
	if len(lang) == 0 {
		return ""
	}
	return langHint(lang[0])
}

// langHint returns lang normalized, or "" if it can't be normalized.
// It is normalized instead of cut to two bytes,
// because language.Code.Validate accepts hints like " en" and "EN-us".
func langHint(lang language.Code) language.Code {
	if lang == "" {
		// No hint, which every Normalized call passes, without building
		// the error of an invalid language.Code
		return ""
	}
	hint, err := lang.Normalized()
	if err != nil {
		return ""
	}
	return hint
}
