// Package date provides comprehensive date handling and validation utilities
// for Go applications with support for multiple date formats and internationalization.
//
// The package includes:
// - Date type with ISO 8601 format (YYYY-MM-DD) support
// - Flexible date parsing with language hints
// - Date arithmetic and comparison operations
// - Period range calculations (year, quarter, month, week)
// - Database integration (Scanner/Valuer interfaces)
// - JSON marshalling/unmarshalling
// - Nullable date support
// - Time zone handling
package date

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types/language"
	"github.com/domonda/go-types/strutil"
)

// Note: Date does not implement types.NormalizableValidator[Date] because
// its Normalized() method accepts optional language.Code parameters.
// This is intentional as it provides more functionality than the interface requires.

// Normalize returns str as normalized Date or an error.
// The first given lang argument is used as language hint for parsing.
func Normalize(str string, lang ...language.Code) (Date, error) {
	return Date(str).Normalized(lang...)
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
// Date implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and treats empty string or zero dates ("0000-00-00" and "0001-01-01") as SQL NULL.
type Date string

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
// - period of a ISO 8601 week of a year: YYYY-Wnn
// - period of a month of a year: YYYY-MM
// - period of a quarter of a year: YYYY-Qn
// - period of a half year: YYYY-Hn
// - period of full year: YYYY
//
// The returned from Date is the first day of the month of the period,
// the returned until Date is the last day of the month of the period.
//
// Examples:
// - Period of June 2018: PeriodRange("2018-06") == Date("2018-06-01"), Date("2018-06-30"), nil
// - Period of Q3 2018: PeriodRange("2018-Q3") == Date("2018-07-01"), Date("2018-09-30"), nil
// - Period of second half of 2018: PeriodRange("2018-H2") == Date("2018-07-01"), Date("2018-12-31"), nil
// - Period of year 2018: PeriodRange("2018") == Date("2018-01-01"), Date("2018-12-31"), nil
// - Period of week 1 2019: PeriodRange("2019-W01") == Date("2018-12-31"), Date("2019-01-06"), nil
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
	// January 1st of the year
	t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// Go back to Monday of week
	t = t.AddDate(0, 0, int(time.Monday-t.Weekday()))

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

// ScanString tries to parse and assign the passed source string as value of the Date.
//
// If validate is true, the source string is checked for validity before it is assigned.
// If validate is false and the source string can still be assigned in some non-normalized way,
// it will be assigned without returning an error.
//
// This method is part of the types.StringScanner interface.
func (date *Date) ScanString(source string, validate bool) error {
	newDate, err := Date(source).Normalized()
	if err != nil {
		if validate {
			return err
		}
		newDate = Date(source)
	}
	*date = newDate
	return nil
}

// ScanStringWithLang parses and assigns the source string as a Date using the given language hint.
// Returns wasNormalized to indicate if the source was already in normalized form,
// monthMustBeFirst to indicate if the date format requires month-first ordering,
// and an error if parsing fails.
func (date *Date) ScanStringWithLang(source string, lang language.Code) (wasNormalized bool, monthMustBeFirst bool, err error) {
	newDate, monthMustBeFirst, err := normalizeAndCheckDate(source, lang)
	if err != nil {
		return false, false, err
	}
	*date = newDate
	return newDate == Date(source), monthMustBeFirst, nil
}

// String returns the normalized date if possible, otherwise returns the unchanged string.
// String implements the fmt.Stringer interface.
func (date Date) String() string {
	norm, err := date.Normalized()
	if err != nil {
		return string(date)
	}
	return string(norm)
}

// WithinIncl returns true if the date is within the inclusive range [from, until].
func (date Date) WithinIncl(from, until Date) bool {
	t := date.MidnightUTC()
	tFrom := from.MidnightUTC()
	tUntil := until.MidnightUTC()
	return (t.Equal(tFrom) || t.After(tFrom)) && (t.Equal(tUntil) || t.Before(tUntil))
}

// BetweenExcl returns true if the date is strictly between (exclusive) after and before.
func (date Date) BetweenExcl(after, before Date) bool {
	t := date.MidnightUTC()
	return t.After(after.MidnightUTC()) && t.Before(before.MidnightUTC())
}

// Nullable returns the date as a NullableDate.
// Returns Null if the date is zero, otherwise returns the date as NullableDate.
func (date Date) Nullable() NullableDate {
	if date.IsZero() {
		return Null
	}
	return NullableDate(date)
}

// IsZero returns true when the date is any of ["", "0000-00-00", "0001-01-01"].
// "0001-01-01" is treated as zero because it's the zero value of time.Time.
// "0000-00-00" may be the zero value of other date implementations.
func (date Date) IsZero() bool {
	return date == "" || date == "0001-01-01" || date == "0000-00-00"
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
	_, err := time.Parse(Layout, string(date))
	return err == nil
}

// Time returns a time.Time for the date at the specified hour, minute, and second in the given location.
// Returns a zero time.Time if the date is zero or invalid.
func (date Date) Time(hour, minute, second int, location *time.Location) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	year, month, day := date.YearMonthDay()
	return time.Date(year, month, day, hour, minute, second, 0, location)
}

// TimeLocal returns a time.Time for the date at the specified hour, minute, and second in local time.
// Returns a zero time.Time if the date is zero or invalid.
func (date Date) TimeLocal(hour, minute, second int) time.Time {
	return date.Time(hour, minute, second, time.Local)
}

// TimeUTC returns a time.Time for the date at the specified hour, minute, and second in UTC.
// Returns a zero time.Time if the date is zero or invalid.
func (date Date) TimeUTC(hour, minute, second int) time.Time {
	return date.Time(hour, minute, second, time.UTC)
}

// MidnightUTC returns the midnight (00:00) time.Time of the date in UTC,
// or a zero time.Time value if the date is not valid.
func (date Date) MidnightUTC() time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	// time.Parse uses UTC
	t, err := time.Parse(Layout, string(date))
	if err != nil {
		return time.Time{}
	}
	return t
}

// Midnight returns the midnight (00:00) time.Time of the date in the local time zone.
// Returns a zero time.Time if the date is not valid.
func (date Date) Midnight() time.Time {
	return date.MidnightInLocation(time.Local)
}

// MidnightInLocation returns the midnight (00:00) time.Time of the date in the given location.
// Returns a zero time.Time if the date is not valid.
func (date Date) MidnightInLocation(loc *time.Location) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	t, err := time.ParseInLocation(Layout, string(date), loc)
	if err != nil {
		return time.Time{}
	}
	return t
}

// Format formats the date using the given layout string (see time.Time.Format).
// Returns an empty string if date or layout are empty.
// If layout equals Layout constant, returns the date as-is for efficiency.
func (date Date) Format(layout string) string {
	if date == "" || layout == "" {
		return ""
	}
	if layout == Layout {
		return string(date)
	}
	return date.MidnightUTC().Format(layout)
}

// NormalizedOrUnchanged returns the normalized date if possible, otherwise returns the date unchanged.
// The first given lang argument is used as language hint for parsing.
func (date Date) NormalizedOrUnchanged(lang ...language.Code) Date {
	normalized, err := date.Normalized(lang...)
	if err != nil {
		return date
	}
	return normalized
}

// func (date Date) NormalizedOrInvalid(lang ...language.Code) Date {
// 	normalized, err := date.Normalized(lang...)
// 	if err != nil {
// 		return Invalid
// 	}
// 	return normalized
// }

// NormalizedOrNull returns the normalized date as NullableDate if possible, otherwise returns Null.
// The first given lang argument is used as language hint for parsing.
func (date Date) NormalizedOrNull(lang ...language.Code) NullableDate {
	normalized, err := date.Normalized(lang...)
	if err != nil {
		return Null
	}
	return NullableDate(normalized)
}

// NormalizedEqual returns true if two dates are equal after normalization.
func (date Date) NormalizedEqual(other Date) bool {
	a, _ := date.Normalized()
	b, _ := other.Normalized()
	return a == b
}

// Compare compares the date with another Date.
// Returns -1 if date is before other, +1 if after, 0 if equal.
func (date Date) Compare(other Date) int {
	a, _ := date.Normalized()
	b, _ := other.Normalized()
	return strings.Compare(string(a), string(b))
}

// After returns true if the date is after the other date.
func (date Date) After(other Date) bool {
	return date.MidnightUTC().After(other.MidnightUTC())
}

// EqualOrAfter returns true if the date is equal to or after the other date.
func (date Date) EqualOrAfter(other Date) bool {
	return date.NormalizedEqual(other) || date.After(other)
}

// Before returns true if the date is before the other date.
func (date Date) Before(other Date) bool {
	return date.MidnightUTC().Before(other.MidnightUTC())
}

// EqualOrBefore returns true if the date is equal to or before the other date.
func (date Date) EqualOrBefore(other Date) bool {
	return date.NormalizedEqual(other) || date.Before(other)
}

// AfterTime returns true if midnight of the date (in the location of the passed time) is after the time.
func (date Date) AfterTime(other time.Time) bool {
	return date.MidnightInLocation(other.Location()).After(other)
}

// BeforeTime returns true if midnight of the date (in the location of the passed time) is before the time.
func (date Date) BeforeTime(other time.Time) bool {
	return date.MidnightInLocation(other.Location()).Before(other)
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
func (date Date) IsToday() bool {
	return date == OfToday()
}

// IsTodayInUTC returns true if the date is today in UTC time.
func (date Date) IsTodayInUTC() bool {
	return date == OfNowInUTC()
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
// Accepts string, time.Time, or nil values.
// Empty strings and zero dates ("0000-00-00", "0001-01-01") are treated as empty Date.
func (date *Date) Scan(value any) (err error) {
	switch x := value.(type) {
	case string:
		d := Date(x)
		if !d.IsZero() {
			d, err = d.Normalized()
			if err != nil {
				return err
			}
		}
		*date = d
		return nil

	case time.Time:
		*date = OfTime(x)
		return nil

	case nil:
		*date = ""
		return nil
	}

	return fmt.Errorf("can't scan value '%#v' of type %T as date.Date", value, value)
}

// Value implements the database/sql/driver.Valuer interface.
// Returns nil for zero dates, otherwise returns the normalized date string.
func (date Date) Value() (driver.Value, error) {
	if date.IsZero() {
		return nil, nil
	}
	normalized, err := date.Normalized()
	if err != nil {
		return nil, err
	}
	return string(normalized), nil
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
	return unicode.IsSpace(r) || r == '.' || r == '/' || r == '-'
}

func isDateTrimRune(r rune) bool {
	return unicode.IsSpace(r) || (!unicode.IsLetter(r) && !unicode.IsDigit(r))
}

// Normalized returns the date in normalized form,
// or an error if the format can't be detected.
// The first given lang argument is used as language hint.
func (date Date) Normalized(lang ...language.Code) (Date, error) {
	normalized, _, err := normalizeAndCheckDate(string(date), getLangHint(lang))
	return normalized, err
}

func normalizeAndCheckDate(str string, langHint language.Code) (Date, bool, error) {
	normalized, monthMustBeFirst, err := normalizeDate(str, langHint)
	if err != nil {
		return "", monthMustBeFirst, err
	}
	_, err = time.Parse(Layout, normalized)
	if err != nil {
		return "", monthMustBeFirst, err
	}
	return Date(normalized), monthMustBeFirst, nil
}

func normalizeDate(str string, langHint language.Code) (string, bool, error) {
	trimmed := strings.TrimSuffix(str, "00:00:00") // Trim zero time part
	trimmed = strings.TrimFunc(trimmed, isDateTrimRune)
	if len(trimmed) < MinLength {
		return "", false, fmt.Errorf("too short for a date: %q", str)
	}

	if len(trimmed) > 10 && trimmed[10] == 'T' {
		// Use date part of this date-time format: "2006-01-02T15:04:05"
		trimmed = trimmed[:10]
	}

	trimmed = strings.ToLower(trimmed)

	langHint, _ = langHint.Normalized()

	parts := strings.FieldsFunc(trimmed, isDateSeparatorRune)
	if len(parts) == 4 {
		i := strutil.IndexInStrings("of", parts)
		if i > 0 && i <= 2 {
			// remove the word "of" within date
			parts = append(parts[:i], parts[i+1:]...)
		}
	}
	if len(parts) != 3 {
		return "", false, fmt.Errorf("date must have 3 parts: %q", str)
	}
	dayHint := -1
	totalLen := 0
	for i := range parts {
		l := len(parts[i])
		totalLen += l
		if l == 1 {
			parts[i] = "0" + parts[i]
		} else if parts[i] == "1st" {
			parts[i] = "01"
			dayHint = i
		} else if parts[i] == "2nd" {
			parts[i] = "02"
			dayHint = i
		} else if parts[i] == "3rd" {
			parts[i] = "03"
			dayHint = i
		} else if strings.HasSuffix(parts[i], "th") {
			parts[i] = strings.TrimSuffix(parts[i], "th")
			if len(parts[i]) == 1 {
				parts[i] = "0" + parts[i]
			}
			dayHint = i
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

	// fmt.Println(len0, len1, len2)
	// fmt.Println(val0, val1, val2)

	expandVal2ToFullYear := func() {
		if len2 != 2 {
			panic("len2")
		}
		if val2 < 45 {
			parts[2] = "20" + parts[2]
			val2 = 2000 + val2
		} else {
			parts[2] = "19" + parts[2]
			val2 = 1900 + val2
		}
		len2 = 4
	}

	switch {
	case month0 != 0:
		if len2 == 2 {
			expandVal2ToFullYear()
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
			// DD m YY
		}
		expandVal2ToFullYear()
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
		expandVal2ToFullYear()
		fallthrough

	case len0 == 2 && len1 == 2 && len2 == 4:
		monthMustBeFirst := validMonth(val0) && !validMonth(val1)
		if (!validMonth(val1) && validMonth(val0)) || dayHint == 1 || langHint == "en" {
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

func validYear(year int) bool {
	return year > 0
}

func validMonth(month int) bool {
	return month >= 1 && month <= 12
}

func validDay(day int) bool {
	return day >= 1 && day <= 31
}

func getLangHint(lang []language.Code) language.Code {
	if len(lang) == 0 || len(lang[0]) < 2 {
		return ""
	}
	return lang[0][:2]
}
