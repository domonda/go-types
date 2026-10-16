package date

import (
	"database/sql/driver"
	"encoding/xml"
	"strings"
	"time"

	"github.com/invopop/jsonschema"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/language"
	"github.com/domonda/go-types/nullable"
)

// Null is an empty string and will be treated as SQL NULL.
// date.Null.IsZero() == true
const Null NullableDate = ""

// NullableDate is identical to Date, except that a zero date (see Date.IsZero)
// is null and considered valid by the Valid() and Validate() methods.
// The main difference between Date and NullableDate is:
// Date("").Valid() == false
// NullableDate("").Valid() == true
//
// NullableDate reads and writes like Date, except that null is SQL NULL
// and JSON null. See "Reading and writing values" in the README of the module.
type NullableDate string

// Compile-time check that NullableDate implements nullable.NullSetable[Date]
var _ nullable.NullSetable[Date] = (*NullableDate)(nil)

// Compile-time check that NullableDate implements types.NormalizableValidator[NullableDate]
var _ types.NormalizableValidator[NullableDate] = NullableDate("")

// NormalizeNullable returns str as normalized NullableDate or an error.
// The first given lang argument is used as language hint for parsing.
func NormalizeNullable(str string, lang ...language.Code) (NullableDate, error) {
	return NullableDate(str).NormalizedLang(getLangHint(lang))
}

// MustNullable returns str as normalized NullableDate or panics if str is not
// neither a valid Date nor Null ("").
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

// IsZero returns true if the date is zero like Date.IsZero,
// which is the same as IsNull.
func (n NullableDate) IsZero() bool {
	return Date(n).IsZero()
}

// IsNull returns true if the NullableDate is null or zero, see Date.IsZero,
// which includes a string of only the runes strutil.TrimSpace removes
// like for country.NullableCode, and a zero date in another format
// like "01.01.0001".
// IsNull implements the nullable.Nullable interface.
func (n NullableDate) IsNull() bool {
	return Date(n).IsZero()
}

// IsNotNull returns true if the NullableDate is not null.
func (n NullableDate) IsNotNull() bool {
	return !n.IsNull()
}

// Set sets a Date for this NullableDate.
func (n *NullableDate) Set(d Date) {
	*n = NullableDate(d)
}

// SetNull sets the NullableDate to null.
func (n *NullableDate) SetNull() {
	*n = Null
}

// Get returns the non-nullable Date value or panics if the NullableDate is null.
// Note: check with IsNull before using Get!
func (n NullableDate) Get() Date {
	if n.IsNull() {
		panic("date.Null")
	}
	return Date(n)
}

// GetOr returns the non-nullable Date value or the passed defaultDate if the NullableDate is null.
func (n NullableDate) GetOr(defaultDate Date) Date {
	if n.IsNull() {
		return defaultDate
	}
	return Date(n)
}

// Valid returns if the format of the date is correct, see Format.
// A null date is valid, see IsNull.
func (n NullableDate) Valid() bool {
	return n.Validate() == nil
}

// ValidAndNotNull returns if the date is valid and not Null or Zero.
func (n NullableDate) ValidAndNotNull() bool {
	return n.IsNotNull() && Date(n).Valid()
}

// ValidAndNormalized returns if the date is valid and already normalized.
func (n NullableDate) ValidAndNormalized() bool {
	norm, err := n.Normalized()
	return err == nil && norm == n
}

// Validate returns an error if the date is not in a valid, normalizable format.
// A null date is valid, see IsNull.
func (n NullableDate) Validate() error {
	if n.IsNull() {
		return nil
	}
	return Date(n).Validate()
}

// ScanString implements the strfmt.Scannable interface.
// The strings "NULL", "null" and "nil" set the NullableDate to null.
// Without validate it stores source with surrounding whitespace trimmed,
// including a zero date. With validate it stores the normalized date or
// returns the error of Normalized.
func (n *NullableDate) ScanString(source string, validate bool) error {
	if validate {
		return internal.ScanStringNormalized(n, source, true, true, NullableDate.Normalized, NullableDate.IsNull)
	}
	return internal.ScanStringNormalized(n, source, false, true, internal.Trim[NullableDate], nil)
}

// ScanStringWithLang parses and assigns the source string as a NullableDate using the given language hint.
// Returns wasNormalized to indicate if the source was already in normalized form,
// monthMustBeFirst to indicate if the date format requires month-first ordering
// like Date.ScanStringWithLang, and an error if parsing fails.
// Null values (empty, whitespace only or zero) are treated as Null.
func (n *NullableDate) ScanStringWithLang(source string, lang language.Code) (wasNormalized bool, monthMustBeFirst bool, err error) {
	if NullableDate(source).IsNull() {
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

// StringOr returns the NullableDate as stored,
// or the passed nullString if the NullableDate is null.
func (n NullableDate) StringOr(nullString string) string {
	if n.IsNull() {
		return nullString
	}
	return string(n)
}

// Normalized returns the date in normalized form,
// or an error if the format can't be detected.
// If n.IsNull() then Null, nil will be returned.
// Returns the NullableDate unchanged in case of an error.
// Normalized implements types.Normalizable[NullableDate],
// use NormalizedLang to pass a language hint.
func (n NullableDate) Normalized() (NullableDate, error) {
	return n.NormalizedLang("")
}

// NormalizedLang returns the date in normalized form like Normalized,
// with lang as hint for the order of day and month like Date.NormalizedLang.
func (n NullableDate) NormalizedLang(lang language.Code) (NullableDate, error) {
	if n.IsNull() {
		return Null, nil
	}
	d, err := Date(n).NormalizedLang(lang)
	return NullableDate(d), err
}

// NormalizedOrNull returns the normalized date as NullableDate if possible, otherwise returns Null.
// The first given lang argument is used as language hint for parsing.
func (n NullableDate) NormalizedOrNull(lang ...language.Code) NullableDate {
	norm, err := n.NormalizedLang(getLangHint(lang))
	if err != nil {
		return Null
	}
	return norm
}

// NormalizedOrUnchanged returns the normalized date if possible, otherwise returns the date unchanged.
// The first given lang argument is used as language hint for parsing.
func (n NullableDate) NormalizedOrUnchanged(lang ...language.Code) NullableDate {
	normalized, err := n.NormalizedLang(getLangHint(lang))
	if err != nil {
		return n
	}
	return normalized
}

// MidnightUTC returns the midnight (00:00) nullable.Time of the date in UTC,
// or a null nullable.Time value if the date is null or not valid.
// A valid date that is not in the normalized form is normalized first.
func (n NullableDate) MidnightUTC() nullable.Time {
	return nullable.TimeFrom(Date(n).MidnightUTC())
}

// Midnight returns the midnight (00:00) nullable.Time of the date
// in the local time zone,
// or a null nullable.Time value if the date is not valid.
func (n NullableDate) Midnight() nullable.Time {
	return n.MidnightInLocation(time.Local)
}

// MidnightInLocation returns the midnight (00:00) nullable.Time of the date in the given location.
// Returns a null nullable.Time if the date is not valid or null.
// A valid date that is not in the normalized form is normalized first.
func (n NullableDate) MidnightInLocation(loc *time.Location) nullable.Time {
	return nullable.TimeFrom(Date(n).MidnightInLocation(loc))
}

// NormalizedEqual returns true if two dates are equal after normalization.
// Null and the zero dates are equal to each other.
// An invalid date has no normalized value and is not equal
// to any date, not even to Null or an identical invalid string.
func (n NullableDate) NormalizedEqual(other NullableDate) bool {
	a, err := n.Normalized()
	if err != nil {
		return false
	}
	b, err := other.Normalized()
	return err == nil && a == b
}

// Compare compares the date with another NullableDate.
// Returns -1 if n is before other, +1 if after, 0 if equal.
// A null date is always before a non-null date.
// An invalid date compares like Null, see Date.Compare.
func (n NullableDate) Compare(other NullableDate) int {
	return strings.Compare(string(n.NormalizedOrNull()), string(other.NormalizedOrNull()))
}

// After returns true if the date is after the other date.
// A null date is always before any non-null date,
// so any non-null date is after a null date
// and a null date is never after any date.
// Returns false if either date is not null and not valid, see Date.After.
func (n NullableDate) After(other NullableDate) bool {
	if n.IsNull() {
		return false
	}
	if other.IsNull() {
		return Date(n).Valid()
	}
	return Date(n).After(Date(other))
}

// EqualOrAfter returns true if the date is equal to or after the other date.
func (n NullableDate) EqualOrAfter(other NullableDate) bool {
	return n.NormalizedEqual(other) || n.After(other)
}

// Before returns true if the date is before the other date.
// A null date is always before any non-null date.
// Returns false if either date is not null and not valid, see Date.After.
func (n NullableDate) Before(other NullableDate) bool {
	if other.IsNull() {
		return false
	}
	if n.IsNull() {
		return Date(other).Valid()
	}
	return Date(n).Before(Date(other))
}

// EqualOrBefore returns true if the date is equal to or before the other date.
func (n NullableDate) EqualOrBefore(other NullableDate) bool {
	return n.NormalizedEqual(other) || n.Before(other)
}

// AfterTime returns true if midnight of the date (in the location of the passed time) is after the time.
// Returns false if the date is null or not valid, see Date.After.
func (n NullableDate) AfterTime(other time.Time) bool {
	if n.IsNull() {
		return false
	}
	return Date(n).AfterTime(other)
}

// BeforeTime returns true if midnight of the date (in the location of the passed time) is before the time.
// Returns true if the date is null and the time is not zero,
// and false if the date is not valid, see Date.After.
func (n NullableDate) BeforeTime(other time.Time) bool {
	if n.IsNull() {
		return !other.IsZero()
	}
	return Date(n).BeforeTime(other)
}

// AddDate returns a new date with the specified years, months, and days added.
// Returns Null if the NullableDate is null.
func (n NullableDate) AddDate(years int, months int, days int) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).AddDate(years, months, days).Nullable()
}

// AddYears returns a new date with the specified number of years added.
// Returns Null if the NullableDate is null.
func (n NullableDate) AddYears(years int) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).AddYears(years).Nullable()
}

// AddMonths returns a new date with the specified number of months added.
// Returns Null if the NullableDate is null.
func (n NullableDate) AddMonths(months int) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).AddMonths(months).Nullable()
}

// AddDays returns a new date with the specified number of days added.
// Returns Null if the NullableDate is null.
func (n NullableDate) AddDays(days int) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).AddDays(days).Nullable()
}

// Add returns a new date with the specified duration added.
// Returns Null if the NullableDate is null.
func (n NullableDate) Add(d time.Duration) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).Add(d).Nullable()
}

// Sub returns the duration between the date and the other date (n - other).
// Returns 0 if either date is null or both dates are equal.
func (n NullableDate) Sub(other NullableDate) time.Duration {
	if n.IsNull() || other.IsNull() || other == n {
		return 0
	}
	return Date(n).Sub(Date(other))
}

// BeginningOfWeek returns the date of the first day of the week containing this date.
// The startDay parameter specifies which day is considered the start of the week.
// Returns Null if the NullableDate is null.
func (n NullableDate) BeginningOfWeek(startDay time.Weekday) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).BeginningOfWeek(startDay).Nullable()
}

// BeginningOfMonth returns the date of the first day of the month containing this date.
// Returns Null if the NullableDate is null.
func (n NullableDate) BeginningOfMonth() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).BeginningOfMonth().Nullable()
}

// BeginningOfQuarter returns the date of the first day of the quarter containing this date.
// Returns Null if the NullableDate is null.
func (n NullableDate) BeginningOfQuarter() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).BeginningOfQuarter().Nullable()
}

// BeginningOfYear returns the date of the first day of the year containing this date.
// Returns Null if the NullableDate is null.
func (n NullableDate) BeginningOfYear() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).BeginningOfYear().Nullable()
}

// EndOfWeek returns the date of the last day of the week containing this date.
// The weekday parameter specifies which day is considered the start of the week.
// Returns Null if the NullableDate is null.
func (n NullableDate) EndOfWeek(weekday time.Weekday) NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).EndOfWeek(weekday).Nullable()
}

// EndOfMonth returns the date of the last day of the month containing this date.
// Returns Null if the NullableDate is null.
func (n NullableDate) EndOfMonth() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).EndOfMonth().Nullable()
}

// EndOfQuarter returns the date of the last day of the quarter containing this date.
// Returns Null if the NullableDate is null.
func (n NullableDate) EndOfQuarter() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).EndOfQuarter().Nullable()
}

// EndOfYear returns the date of the last day of the year containing this date.
// Returns Null if the NullableDate is null.
func (n NullableDate) EndOfYear() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).EndOfYear().Nullable()
}

// LastMonday returns the date of the Monday on or before this date.
// Returns Null if the NullableDate is null.
func (n NullableDate) LastMonday() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).LastMonday().Nullable()
}

// NextSunday returns the date of the Sunday on or after this date.
// Returns Null if the NullableDate is null.
func (n NullableDate) NextSunday() NullableDate {
	if n.IsNull() {
		return Null
	}
	return Date(n).NextSunday().Nullable()
}

// YearMonthDay returns the year, month, and day components of the date.
// Returns zero values if the date is not valid or null.
func (n NullableDate) YearMonthDay() (year int, month time.Month, day int) {
	if n.IsNull() {
		return 0, 0, 0
	}
	return Date(n).YearMonthDay()
}

// Year returns the year component of the date.
// Returns 0 if the date is null.
func (n NullableDate) Year() int {
	year, _, _ := n.YearMonthDay()
	return year
}

// Month returns the month component of the date.
// Returns 0 if the date is null.
func (n NullableDate) Month() time.Month {
	_, month, _ := n.YearMonthDay()
	return month
}

// Day returns the day of the month component of the date.
// Returns 0 if the date is null.
func (n NullableDate) Day() int {
	_, _, day := n.YearMonthDay()
	return day
}

// Weekday returns the day of the week for this date.
// Returns 0 (Sunday) if the date is null.
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

// IsToday returns true if the date is today in local time.
// Returns false if the date is null.
func (n NullableDate) IsToday() bool {
	if n.IsNull() {
		return false
	}
	return Date(n).IsToday()
}

// IsTodayInUTC returns true if the date is today in UTC time.
// Returns false if the date is null.
func (n NullableDate) IsTodayInUTC() bool {
	if n.IsNull() {
		return false
	}
	return Date(n).IsTodayInUTC()
}

// AfterToday returns true if the date is after today in local time.
// Returns false if the date is null.
func (n NullableDate) AfterToday() bool {
	if n.IsNull() {
		return false
	}
	return Date(n).AfterToday()
}

// AfterTodayInUTC returns true if the date is after today in UTC time.
// Returns false if the date is null.
func (n NullableDate) AfterTodayInUTC() bool {
	if n.IsNull() {
		return false
	}
	return Date(n).AfterTodayInUTC()
}

// BeforeToday returns true if the date is before today in local time.
// A null date is always before today.
func (n NullableDate) BeforeToday() bool {
	if n.IsNull() {
		return true
	}
	return Date(n).BeforeToday()
}

// BeforeTodayInUTC returns true if the date is before today in UTC time.
// A null date is always before today.
func (n NullableDate) BeforeTodayInUTC() bool {
	if n.IsNull() {
		return true
	}
	return Date(n).BeforeTodayInUTC()
}

// Format returns the date formatted like Date.Format,
// or an empty string if n is null (see IsNull) or layout is an empty string.
// A date that is not valid is returned unchanged.
func (n NullableDate) Format(layout string) string {
	if n.IsNull() || layout == "" {
		return ""
	}
	return Date(n).Format(layout)
}

// Scan implements the database/sql.Scanner interface.
// A valid date from a string or []byte is stored normalized, SQL NULL
// and a zero date (see IsNull) as Null, and a date that can't be
// normalized unchanged without an error, so that one invalid value
// doesn't fail a whole query; Validate reports it. A time.Time is stored
// as OfTime, which is already YYYY-MM-DD or Null for the zero time,
// without parsing it again.
func (n *NullableDate) Scan(value any) error {
	if t, ok := value.(time.Time); ok {
		*n = NullableDate(OfTime(t))
		return nil
	}
	return internal.ScanNormalized(n, value, NullableDate.Normalized, NullableDate.IsNull)
}

// Value implements the database/sql/driver.Valuer interface.
// It returns the normalized date, SQL NULL if the NullableDate is null,
// or the error of Normalized for a date that can't be normalized.
func (n NullableDate) Value() (driver.Value, error) {
	return internal.ValueNormalized(n, NullableDate.Normalized, NullableDate.IsNull)
}

// MarshalJSON implements encoding/json.Marshaler by writing JSON null
// if the NullableDate is blank, else the date with surrounding whitespace
// trimmed. A zero date is written as that text, unlike Value.
func (n NullableDate) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(n, true, internal.Trim[NullableDate], nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// but with "" for null, so that encoding/xml and strfmt write the
// trimmed date. encoding/json writes map keys of string types as they
// are and doesn't call MarshalText for them.
func (n NullableDate) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(n, internal.Trim[NullableDate], nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (n *NullableDate) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(n, name)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored with surrounding whitespace trimmed, including
// a zero date, unlike Scan. JSON null sets the NullableDate to null, so
// that decoding null into a field that holds a value resets it, and any
// other JSON value is an error.
func (n *NullableDate) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(n, data, true, internal.Trim[NullableDate], nil)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like UnmarshalJSON stores it.
func (n *NullableDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(n, d, start, internal.Trim[NullableDate], nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like UnmarshalJSON stores it.
func (n *NullableDate) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(n, attr, internal.Trim[NullableDate], nil)
}

// JSONSchema returns the JSON schema definition for the NullableDate type.
// Implements the jsonschema.JSONSchemaProvider interface.
func (NullableDate) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable Date",
		OneOf: []*jsonschema.Schema{
			{
				Type:   "string",
				Format: "date",
			},
			{Type: "null"},
		},
		Default: Null,
	}
}
