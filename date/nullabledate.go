package date

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/domonda/go-types/language"
	"github.com/domonda/go-types/nullable"
)

// Null is an empty string and will be treatet as SQL NULL.
// date.Null.IsZero() == true
var Null NullableDate

// NullableDate is identical to Date, except that IsZero() is considered valid
// by the Valid() and Validate() methods.
// NullableDate implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty/zero Date string as SQL NULL value.
// The main difference between Date and NullableDate is:
// Date("").Valid() == false
// NullableDate("").Valid() == true
type NullableDate string

// NormalizeNullable returns str as normalized NullableDate or an error.
// The first given lang argument is used as language hint.
func NormalizeNullable(str string, lang ...language.Code) (NullableDate, error) {
	return NullableDate(str).Normalized(lang...)
}

// MustNullable returns str as normalized NullableDate or panics if str is not neither a valid Date nor Null ("").
func MustNullable(str string) NullableDate {
	d, err := NullableDate(str).Normalized()
	if err != nil {
		panic(err)
	}
	return d
}

// Date returns the NullableDate as Date without checking if it's null.
// See also Get which panics on null.
func (n NullableDate) Date() Date {
	return Date(n)
}

// DateOr returns the NullableDate as Date if it is not null,
// else useIfNull will be returned.
func (n NullableDate) DateOr(useIfNull Date) Date {
	if n.IsNull() {
		return useIfNull
	}
	return Date(n)
}

// IsZero returns true when the date is any of ["", "0000-00-00", "0001-01-01", "null", "NULL"]
// "0001-01-01" is treated as zero because it's the zero value of time.Time.
func (n NullableDate) IsZero() bool {
	return Date(n).IsZero()
}

// IsNull returns true if the NullableDate is null or zero.
// IsNull implements the nullable.Nullable interface.
func (n NullableDate) IsNull() bool {
	return n.IsZero()
}

// IsNotNull returns true if the NullableDate is not null.
func (n NullableDate) IsNotNull() bool {
	return !n.IsNull()
}

// Set sets an Date for this NullableDate
func (n *NullableDate) Set(d Date) {
	*n = NullableDate(d)
}

// SetNull sets the NullableDate to null
func (n *NullableDate) SetNull() {
	*n = Null
}

// Get returns the non nullable Date value
// or panics if the NullableDate is null.
// Note: check with IsNull before using Get!
func (n NullableDate) Get() Date {
	if n.IsNull() {
		panic("date.Null")
	}
	return Date(n)
}

// GetOr returns the non nullable Date value
// or the passed defaultDate if the NullableDate is null.
func (n NullableDate) GetOr(defaultDate Date) Date {
	if n.IsNull() {
		return defaultDate
	}
	return Date(n)
}

// Valid returns if the format of the date is correct, see Format
// n.IsZero() is valid
func (n NullableDate) Valid() bool {
	return n.Validate() == nil
}

// ValidAndNotNull returns if the date is valid and not Null or Zero.
func (n NullableDate) ValidAndNotNull() bool {
	return Date(n).Valid()
}

func (n NullableDate) ValidAndNormalized() bool {
	norm, err := n.Normalized()
	return err == nil && norm == n
}

// Validate returns an error if the date is not in a valid, normalizeable format.
// n.IsZero() is valid
func (n NullableDate) Validate() error {
	if n.IsZero() {
		return nil
	}
	return Date(n).Validate()
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// If validate is true, the source string is checked
// for validity before it is assigned to the type.
//
// If validate is false and the source string
// can still be assigned in some non-normalized way
// it will be assigned without returning an error.
func (n *NullableDate) ScanString(source string, validate bool) error {
	switch source {
	case "", "NULL", "null", "nil":
		n.SetNull()
		return nil
	}
	newDate, err := NullableDate(source).Normalized()
	if err != nil {
		if validate {
			return err
		}
		newDate = NullableDate(source)
	}
	*n = newDate
	return nil
}

func (n *NullableDate) ScanStringWithLang(source string, lang language.Code) (wasNormalized bool, monthMustBeFirst bool, err error) {
	if NullableDate(source).IsZero() {
		*n = Null
		return false, false, nil
	}
	newDate, monthMustBeFirst, err := normalizeAndCheckDate(source, lang)
	if err != nil {
		return false, false, err
	}
	*n = newDate.Nullable()
	return newDate == Date(source), monthMustBeFirst, nil
}

// String returns the normalized date if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (n NullableDate) String() string {
	norm, err := n.Normalized()
	if err != nil {
		return string(n)
	}
	return string(norm)
}

// StringOr returns the NullableDate as string
// or the passed nullString if the NullableDate is null.
func (n NullableDate) StringOr(nullString string) string {
	if n.IsNull() {
		return nullString
	}
	return n.String()
}

// Normalized returns the date in normalized form,
// or an error if the format can't be detected.
// The first given lang argument is used as language hint.
// if n.IsZero() then Null, nil will be returned.
func (n NullableDate) Normalized(lang ...language.Code) (NullableDate, error) {
	if n.IsZero() {
		return Null, nil
	}
	d, err := Date(n).Normalized(lang...)
	return NullableDate(d), err
}

func (n NullableDate) NormalizedOrNull(lang ...language.Code) NullableDate {
	norm, err := n.Normalized(lang...)
	if err != nil {
		return Null
	}
	return norm
}

func (n NullableDate) NormalizedOrUnchanged(lang ...language.Code) NullableDate {
	normalized, err := n.Normalized(lang...)
	if err != nil {
		return n
	}
	return normalized
}

// MidnightUTC returns the midnight (00:00) nullable.Time of the date in UTC,
// or a null nullable.Time value if the date is not valid.
func (n NullableDate) MidnightUTC() nullable.Time {
	if n.IsZero() {
		return nullable.TimeNull
	}
	t, err := nullable.TimeParse(Layout, string(n))
	if err != nil {
		return nullable.TimeNull
	}
	return t
}

// Midnight returns the midnight (00:00) nullable.Time of the date
// in the local time zone,
// or a null nullable.Time value if the date is not valid.
func (n NullableDate) Midnight() nullable.Time {
	return n.MidnightInLocation(time.Local)
}

// MidnightTime returns the midnight (00:00) nullable.Time of the date
// in the given location,
// or a null nullable.Time value if the date is not valid.
func (n NullableDate) MidnightInLocation(loc *time.Location) nullable.Time {
	if n.IsZero() {
		return nullable.TimeNull
	}
	t, err := nullable.TimeParseInLocation(Layout, string(n), loc)
	if err != nil {
		return nullable.TimeNull
	}
	return t
}

// NormalizedEqual returns if two dates are equal in normalized form.
func (n NullableDate) NormalizedEqual(other NullableDate) bool {
	a, _ := n.Normalized()
	b, _ := other.Normalized()
	return a == b
}

// Compare compares n with the passed other NullableDate.
// If n is before the other, it returns -1;
// if n is after the other, it returns +1;
// if they're the same, it returns 0.
// A null date is always before a non-null date.
func (n NullableDate) Compare(other NullableDate) int {
	a, _ := n.Normalized()
	b, _ := other.Normalized()
	return strings.Compare(string(a), string(b))
}

// After returns if the date is after the passed other one.
// Returns false if any of the dates is null.
func (n NullableDate) After(other NullableDate) bool {
	if n.IsNull() || other.IsNull() {
		return false
	}
	return n.MidnightUTC().Get().After(other.MidnightUTC().Get())
}

// EqualOrAfter returns if the date is equal or after the passed other one.
func (n NullableDate) EqualOrAfter(other NullableDate) bool {
	return n.NormalizedEqual(other) || n.After(other)
}

// Before returns if the date is before the passed other one.
// A null date is always before any other date.
func (n NullableDate) Before(other NullableDate) bool {
	if other.IsNull() {
		return false
	}
	if n.IsNull() {
		return true
	}
	return Date(n).Before(Date(other))
}

// EqualOrBefore returns if the date is equal or before the passed other one.
func (n NullableDate) EqualOrBefore(other NullableDate) bool {
	return n.NormalizedEqual(other) || n.Before(other)
}

// AfterTime returns if midnight of the date in location of the passed
// time is after the time.
// Returns false if the date is null.
func (n NullableDate) AfterTime(other time.Time) bool {
	if n.IsNull() {
		return false
	}
	if other.IsZero() {
		return true
	}
	return n.MidnightInLocation(other.Location()).After(other)
}

// BeforeTime returns if midnight of the date in location of the passed
// time is before the time.
// Returns true if the date is null.
func (n NullableDate) BeforeTime(other time.Time) bool {
	if n.IsNull() {
		return !other.IsZero()
	}
	return n.MidnightInLocation(other.Location()).Before(other)
}

func (n NullableDate) AddDate(years int, months int, days int) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).AddDate(years, months, days).Nullable()
}

func (n NullableDate) AddYears(years int) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).AddYears(years).Nullable()
}

func (n NullableDate) AddMonths(months int) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).AddMonths(months).Nullable()
}

func (n NullableDate) AddDays(days int) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).AddDays(days).Nullable()
}

func (n NullableDate) Add(d time.Duration) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).Add(d).Nullable()
}

func (n NullableDate) Sub(other NullableDate) time.Duration {
	if n.IsNull() || other.IsNull() || other == n {
		return 0
	}
	return Date(n).Sub(Date(other))
}

func (n NullableDate) BeginningOfWeek() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).BeginningOfWeek().Nullable()
}

func (n NullableDate) BeginningOfMonth() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).BeginningOfMonth().Nullable()
}

func (n NullableDate) BeginningOfQuarter() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).BeginningOfQuarter().Nullable()
}

func (n NullableDate) BeginningOfYear() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).BeginningOfYear().Nullable()
}

func (n NullableDate) EndOfWeek() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).EndOfWeek().Nullable()
}

func (n NullableDate) EndOfMonth() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).EndOfMonth().Nullable()
}

func (n NullableDate) EndOfQuarter() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).EndOfQuarter().Nullable()
}

func (n NullableDate) EndOfYear() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).EndOfYear().Nullable()
}

func (n NullableDate) LastMonday() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).LastMonday().Nullable()
}

func (n NullableDate) NextSunday() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).NextSunday().Nullable()
}

// YearMonthDay returns the year, month, day components of the Date.
// Zero values will be returned when the date is not valid or null.
func (n NullableDate) YearMonthDay() (year int, month time.Month, day int) {
	if n.IsNull() {
		return 0, 0, 0
	}
	return Date(n).YearMonthDay()
}

// Year of the date
// or zero if the date is null.
func (n NullableDate) Year() int {
	year, _, _ := n.YearMonthDay()
	return year
}

// Month of the date
// or zero if the date is null.
func (n NullableDate) Month() time.Month {
	_, month, _ := n.YearMonthDay()
	return month
}

// Day within the month of the date
// or zero if the date is null.
func (n NullableDate) Day() int {
	_, _, day := n.YearMonthDay()
	return day
}

// Weekday returns the date's day of the week
// or zero if the date is null.
func (n NullableDate) Weekday() time.Weekday {
	if n.IsNull() {
		return 0
	}
	return Date(n).Weekday()
}

// ISOWeek returns the ISO 8601 year and week number in which the date occurs.
// Week ranges from 1 to 53. Jan 01 to Jan 03 of year n might belong to
// week 52 or 53 of year n-1, and Dec 29 to Dec 31 might belong to week 1
// of year n+1.
// Returns zeros if the date is not valid or null.
func (n NullableDate) ISOWeek() (year, week int) {
	if n.IsNull() {
		return 0, 0
	}
	return Date(n).ISOWeek()
}

func (n NullableDate) IsToday() bool {
	if n.IsNull() {
		return false
	}
	return Date(n).IsToday()
}

func (n NullableDate) IsTodayInUTC() bool {
	if n.IsNull() {
		return false
	}
	return Date(n).IsTodayInUTC()
}

func (n NullableDate) AfterToday() bool {
	if n.IsNull() {
		return false
	}
	return Date(n).AfterToday()
}

func (n NullableDate) AfterTodayInUTC() bool {
	if n.IsNull() {
		return false
	}
	return Date(n).AfterTodayInUTC()
}

// Null is always before today.
func (n NullableDate) BeforeToday() bool {
	if n.IsNull() {
		return true
	}
	return Date(n).BeforeToday()
}

// Null is always before today.
func (n NullableDate) BeforeTodayInUTC() bool {
	if n.IsNull() {
		return true
	}
	return Date(n).BeforeTodayInUTC()
}

// Format returns n.MidnightUTC().Format(layout),
// or an empty string if n is Null or layout is an empty string.
func (n NullableDate) Format(layout string) string {
	if n == Null || layout == "" {
		return ""
	}
	if layout == Layout {
		return string(n)
	}
	return n.MidnightUTC().Format(layout)
}

// Scan implements the database/sql.Scanner interface.
func (n *NullableDate) Scan(value any) (err error) {
	switch x := value.(type) {
	case string:
		d := Date(x)
		if !d.IsZero() {
			d, err = d.Normalized()
			if err != nil {
				return err
			}
		}
		*n = d.Nullable()
		return nil

	case time.Time:
		*n = OfTime(x).Nullable()
		return nil

	case nil:
		*n = Null
		return nil
	}

	return fmt.Errorf("can't scan value '%#v' of type %T as data.NullableDate", value, value)
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n NullableDate) Value() (driver.Value, error) {
	if n.IsZero() {
		return nil, nil
	}
	normalized, err := Date(n).Normalized()
	if err != nil {
		return nil, err
	}
	return string(normalized), nil
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value for an empty (null) string.
func (n NullableDate) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(string(n))
}
