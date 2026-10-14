package date

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/jsonschema"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/strutil"
)

// Compile-time check that YearQuarter implements types.NormalizableValidator[YearQuarter]
var _ types.NormalizableValidator[YearQuarter] = YearQuarter("")

// YearQuarter represents a calendar year and quarter in format YYYY-Q# (e.g., "2023-Q1").
// See NullableYearQuarter for a nullable version of this type.
//
// YearQuarter reads and writes like Date.
// Normalized trims the surrounding whitespace and upper-cases a "q" like
// PeriodRange accepts it.
// Scan, UnmarshalJSON, UnmarshalXML and ScanString store a valid
// year-quarter normalized, a zero one (see IsZero) as "" and an invalid
// one unchanged. Value writes the normalized year-quarter, SQL NULL for a
// zero one, and returns an error for an invalid one. MarshalJSON and
// MarshalText write the normalized year-quarter, "" for a zero one, or an
// invalid one unchanged, and fmt prints the year-quarter as stored. See "Reading and writing values"
// in the README of the module.
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
// 1 to 9999 like Date, which has no year 0, and a quarter from 1 to 4,
// or if it is the zero year-quarter "0001-Q1", see IsZero.
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
	if err != nil || year < 1 {
		return yq, fmt.Errorf("invalid year: %q", yearStr)
	}
	quarterStr := string(norm)[6:7]
	quarter, err := strconv.ParseUint(quarterStr, 10, 8)
	if err != nil || quarter < 1 || quarter > 4 {
		return yq, fmt.Errorf("invalid quarter: %q", quarterStr)
	}
	if norm == "0001-Q1" {
		// The year-quarter of the zero time.Time has no practical use, like the zero Date
		return yq, fmt.Errorf("zero year quarter: %q", yq)
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

// ScanString implements the strfmt.Scannable interface.
// Without validate it stores source like Scan, with validate it stores
// the normalized year-quarter or returns the error of Normalized,
// also for a zero source.
func (yq *YearQuarter) ScanString(source string, validate bool) error {
	return internal.ScanStringNormalized(yq, source, validate, false, YearQuarter.Normalized, YearQuarter.IsZero)
}

// Scan implements the database/sql.Scanner interface.
// A valid year-quarter from a string or []byte is stored normalized, a time.Time
// as its year-quarter in the location of the time like Date.Scan, SQL NULL and a zero year-quarter (see IsZero) as "", and a
// year-quarter that can't be normalized unchanged without an error, so that one
// invalid value doesn't fail a whole query; Validate reports it.
func (yq *YearQuarter) Scan(value any) error {
	if t, ok := value.(time.Time); ok {
		value = string(YearQuarterOfTime(t))
	}
	return internal.ScanNormalized(yq, value, YearQuarter.Normalized, YearQuarter.IsZero)
}

// Value implements the database/sql/driver.Valuer interface.
// It returns the normalized year-quarter, SQL NULL for a zero year-quarter (see IsZero),
// or the error of Normalized for a year-quarter that can't be normalized.
func (yq YearQuarter) Value() (driver.Value, error) {
	return internal.ValueNormalized(yq, YearQuarter.Normalized, YearQuarter.IsZero)
}

// MarshalJSON implements encoding/json.Marshaler by writing the
// normalized year-quarter, "" for a zero year-quarter (see IsZero),
// or the year-quarter unchanged if it can't be normalized.
func (yq YearQuarter) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(yq, false, YearQuarter.Normalized, YearQuarter.IsZero)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null leaves the
// YearQuarter unchanged like encoding/json does for a string type,
// and any other JSON value is an error.
func (yq *YearQuarter) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(yq, data, false, YearQuarter.Normalized, YearQuarter.IsZero)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// so that encoding/xml and strfmt write the normalized year-quarter.
// encoding/json writes map keys of string types as they are
// and doesn't call MarshalText for them.
func (yq YearQuarter) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(yq, YearQuarter.Normalized, YearQuarter.IsZero)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (yq *YearQuarter) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(yq, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (yq *YearQuarter) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(yq, d, start, YearQuarter.Normalized, YearQuarter.IsZero)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (yq *YearQuarter) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(yq, attr, YearQuarter.Normalized, YearQuarter.IsZero)
}

// JSONSchema returns the JSON schema definition for the YearQuarter type.
// Implements the jsonschema.JSONSchemaProvider interface.
func (YearQuarter) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:   "Year Quarter",
		Type:    "string",
		Pattern: `^[0-9]{4}-Q[1-4]$`,
	}
}

// IsZero returns true when the year-quarter is any of ["", "0000-Q0", "0001-Q1"],
// also with surrounding whitespace or a lower case "q".
// "0001-Q1" is treated as zero because "0001-01-01" is the zero value of time.Time.
// "0000-Q0" may be the zero value of other date implementations.
// A zero year-quarter has no practical use like a zero Date: reading stores it
// as "", Value writes it as SQL NULL, and it is not a valid YearQuarter.
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
	// Last day of quarter is last day of third month, the 0th day of the
	// month after it, which Of computes also for the quarter 9999-Q4,
	// while adding a month to 9999-12-01 makes no valid Date
	untilDate = Of(yq.Year(), firstMonth+3, 0)
	return fromDate, untilDate
}

// Nullable returns the year-quarter as a NullableYearQuarter.
func (yq YearQuarter) Nullable() NullableYearQuarter {
	return NullableYearQuarter(yq)
}

// AddYears returns a new year-quarter with the specified number of years added.
// Returns the year-quarter unchanged if it is not valid.
// A result outside the years 1 to 9999, like "10000-Q4", is not valid,
// so that Value returns an error for it.
func (yq YearQuarter) AddYears(years int) YearQuarter {
	if !yq.Valid() {
		return yq
	}
	return YearQuarterFrom(yq.Year()+years, yq.Quarter())
}

// AddQuarters returns a new year-quarter with the specified number of quarters added.
// Handles quarter overflow by adjusting the year accordingly.
// Returns the year-quarter unchanged if it is not valid.
// A result outside the years 1 to 9999, like "10000-Q1" for
// "9999-Q4" plus one quarter, is not valid like for AddYears.
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
// YearQuarter, a zero one (see IsZero) as "" like Date.Compare does,
// and an invalid one as it is.
// Returns -1 if yq is before other, +1 if after, 0 if equal.
func (yq YearQuarter) Compare(other YearQuarter) int {
	yq = internal.NormalizedOrRaw(yq, YearQuarter.Normalized, YearQuarter.IsZero)
	other = internal.NormalizedOrRaw(other, YearQuarter.Normalized, YearQuarter.IsZero)
	return strings.Compare(string(yq), string(other))
}
