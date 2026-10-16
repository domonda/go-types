package date

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/invopop/jsonschema"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
)

// YearMonthNull is the null value for NullableYearMonth (empty string).
// Value writes it as SQL NULL and MarshalJSON as JSON null.
const YearMonthNull NullableYearMonth = ""

// Compile-time check that NullableYearMonth implements nullable.NullSetable[YearMonth]
var _ nullable.NullSetable[YearMonth] = (*NullableYearMonth)(nil)

// Compile-time check that NullableYearMonth implements types.NormalizableValidator[NullableYearMonth]
var _ types.NormalizableValidator[NullableYearMonth] = NullableYearMonth("")

// NullableYearMonth is a year-month string in YYYY-MM format that
// can also represent a NULL value as an empty string.
type NullableYearMonth string

// Validate returns nil if the year-month is null or can be normalized,
// see YearMonth.Validate.
func (ym NullableYearMonth) Validate() error {
	if ym.IsNull() {
		return nil
	}
	return YearMonth(ym).Validate()
}

// Valid returns true if the year-month is null or can be normalized,
// see YearMonth.Validate.
func (ym NullableYearMonth) Valid() bool {
	return ym.Validate() == nil
}

// ValidAndNormalized returns true if the year-month is valid and already normalized,
// or if it is the empty YearMonthNull. A whitespace-only or zero value is null and
// valid, but not normalized.
func (ym NullableYearMonth) ValidAndNormalized() bool {
	norm, err := ym.Normalized()
	return err == nil && ym == norm
}

// Normalized returns YearMonthNull for a null year-month, see IsNull,
// else the year-month normalized like YearMonth.Normalized,
// or unchanged together with an error if it can't be normalized.
func (ym NullableYearMonth) Normalized() (NullableYearMonth, error) {
	if ym.IsNull() {
		return YearMonthNull, nil
	}
	norm, err := YearMonth(ym).Normalized()
	return NullableYearMonth(norm), err
}

// ValidAndNotNull returns if the year-month is valid and not null, see IsNull.
func (ym NullableYearMonth) ValidAndNotNull() bool {
	return ym.IsNotNull() && YearMonth(ym).Valid()
}

// IsZero returns true if the year-month is zero like YearMonth.IsZero,
// which is the same as IsNull.
func (ym NullableYearMonth) IsZero() bool {
	return YearMonth(ym).IsZero()
}

// IsNull returns true if the NullableYearMonth is null or zero, see YearMonth.IsZero,
// which includes a string of only whitespace and the zero values
// like "0000-00", like NullableDate.IsNull includes the zero dates.
// IsNull implements the nullable.Nullable interface.
func (ym NullableYearMonth) IsNull() bool {
	return YearMonth(ym).IsZero()
}

// IsNotNull returns true if the NullableYearMonth is not null, see IsNull.
func (ym NullableYearMonth) IsNotNull() bool {
	return !ym.IsNull()
}

// Get returns the non-nullable YearMonth value or panics if the NullableYearMonth is null.
// Note: check with IsNull before using Get!
func (ym NullableYearMonth) Get() YearMonth {
	if ym.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", ym))
	}
	return YearMonth(ym)
}

// GetOr returns the non-nullable YearMonth value or the passed defaultYearMonth if the NullableYearMonth is null.
func (ym NullableYearMonth) GetOr(defaultYearMonth YearMonth) YearMonth {
	if ym.IsNull() {
		return defaultYearMonth
	}
	return YearMonth(ym)
}

// Set sets a YearMonth for this NullableYearMonth.
func (ym *NullableYearMonth) Set(yearMonth YearMonth) {
	*ym = NullableYearMonth(yearMonth)
}

// SetNull sets the NullableYearMonth to null.
func (ym *NullableYearMonth) SetNull() {
	*ym = YearMonthNull
}

// String returns the year-month as a string in YYYY-MM format.
// String implements the fmt.Stringer interface.
func (ym NullableYearMonth) String() string {
	return string(ym)
}

// StringOr returns the NullableYearMonth as string
// or the passed nullString if the NullableYearMonth is null.
func (ym NullableYearMonth) StringOr(nullString string) string {
	if ym.IsNull() {
		return nullString
	}
	return ym.String()
}

// ScanString implements the strfmt.Scannable interface.
// The strings "NULL", "null" and "nil" set the NullableYearMonth to null.
// Without validate it stores source like Scan, with validate it stores
// the normalized year-month or returns the error of Normalized.
func (ym *NullableYearMonth) ScanString(source string, validate bool) error {
	if validate {
		return internal.ScanStringNormalized(ym, source, true, true, NullableYearMonth.Normalized, NullableYearMonth.IsNull)
	}
	return internal.ScanStringNormalized(ym, source, false, true, internal.Trim[NullableYearMonth], nil)
}

// Scan implements the database/sql.Scanner interface.
// Surrounding whitespace is trimmed. A time.Time is stored as its
// year-month, and the zero time.Time as YearMonthNull. SQL NULL and a
// blank string are YearMonthNull. A zero year-month such as "0001-01"
// is kept. An invalid year-month is kept after the trim, without an error.
func (ym *NullableYearMonth) Scan(value any) error {
	if t, ok := value.(time.Time); ok {
		*ym = NullableYearMonth(YearMonthOfTime(t))
		return nil
	}
	return internal.ScanNormalized(ym, value, internal.Trim[NullableYearMonth], nil)
}

// Value implements the database/sql/driver.Valuer interface.
// It returns the trimmed year-month, or SQL NULL for a blank one.
// A zero year-month such as "0001-01" is written as that text.
// An invalid year-month is written after the trim, not rejected.
func (ym NullableYearMonth) Value() (driver.Value, error) {
	return internal.ValueNormalized(ym, internal.Trim[NullableYearMonth], nil)
}

// MarshalJSON implements encoding/json.Marshaler by writing JSON null
// if the NullableYearMonth is blank, else the trimmed year-month.
// A zero year-month is written as that text.
func (ym NullableYearMonth) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(ym, true, internal.Trim[NullableYearMonth], nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null sets the
// NullableYearMonth to null, so that decoding null into a field that holds
// a value resets it, and any other JSON value is an error.
func (ym *NullableYearMonth) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(ym, data, true, internal.Trim[NullableYearMonth], nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// but with "" for null, so that encoding/xml and strfmt write the
// normalized year-month. encoding/json writes map keys of string types
// as they are and doesn't call MarshalText for them.
func (ym NullableYearMonth) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(ym, internal.Trim[NullableYearMonth], nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (ym *NullableYearMonth) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(ym, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (ym *NullableYearMonth) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(ym, d, start, internal.Trim[NullableYearMonth], nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (ym *NullableYearMonth) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(ym, attr, internal.Trim[NullableYearMonth], nil)
}

// JSONSchema returns the JSON schema definition for the NullableYearMonth type.
// Implements the jsonschema.JSONSchemaProvider interface.
func (NullableYearMonth) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable Year Month",
		OneOf: []*jsonschema.Schema{
			{
				Type:    "string",
				Pattern: `^[0-9]{4}-(0[1-9]|1[0-2])$`,
			},
			{Type: "null"},
		},
		Default: YearMonthNull,
	}
}

// Compare compares the normalized year-month with another normalized
// NullableYearMonth, an invalid one as it is. Null is before every year-month.
// Returns -1 if ym is before other, +1 if after, 0 if equal.
func (ym NullableYearMonth) Compare(other NullableYearMonth) int {
	ym, _ = ym.Normalized()
	other, _ = other.Normalized()
	return strings.Compare(string(ym), string(other))
}

// Year returns the year component of the year-month.
// Returns 0 if the year-month is null or not valid.
func (ym NullableYearMonth) Year() int {
	if ym.IsNull() {
		return 0
	}
	return YearMonth(ym).Year()
}

// Month returns the month component of the year-month.
// Returns 0 if the year-month is null or not valid.
func (ym NullableYearMonth) Month() time.Month {
	if ym.IsNull() {
		return 0
	}
	return YearMonth(ym).Month()
}

// Date returns a NullableDate for the given day within this year-month.
// Returns Null if the NullableYearMonth is null.
// The day value is formatted as-is without validation.
func (ym NullableYearMonth) Date(day int) NullableDate {
	if ym.IsNull() {
		return Null
	}
	return NullableDate(YearMonth(ym).Date(day))
}

// DateRange returns the first and last NullableDate of the month.
// Returns Null dates if the NullableYearMonth is null or not valid.
// Returns the date range [fromDate, untilDate] representing the entire month.
func (ym NullableYearMonth) DateRange() (fromDate, untilDate NullableDate) {
	if ym.IsNull() {
		return Null, Null
	}
	from, until := YearMonth(ym).DateRange()
	return NullableDate(from), NullableDate(until)
}

// Format formats the year-month using the given layout string (see time.Time.Format).
// Returns an empty string if ym or layout are empty, or if ym is null.
// If layout equals YearMonthLayout constant, returns the normalized year-month,
// or an invalid one unchanged.
// For other layouts, converts to Date using the 1st day of the month and delegates to Date.Format.
func (ym NullableYearMonth) Format(layout string) string {
	if ym.IsNull() {
		return ""
	}
	return YearMonth(ym).Format(layout)
}

// AddYears returns a new NullableYearMonth with the specified number of years added.
// Returns null if the NullableYearMonth is null
// and an invalid NullableYearMonth unchanged.
func (ym NullableYearMonth) AddYears(years int) NullableYearMonth {
	if ym.IsNull() {
		return YearMonthNull
	}
	return NullableYearMonth(YearMonth(ym).AddYears(years))
}

// AddMonths returns a new NullableYearMonth with the specified number of months added.
// Returns null if the NullableYearMonth is null
// and an invalid NullableYearMonth unchanged.
// Handles month overflow by adjusting the year accordingly.
func (ym NullableYearMonth) AddMonths(months int) NullableYearMonth {
	if ym.IsNull() {
		return YearMonthNull
	}
	return NullableYearMonth(YearMonth(ym).AddMonths(months))
}

// ContainsTime returns true if the given time falls within this year-month.
// Returns false if the NullableYearMonth is null.
func (ym NullableYearMonth) ContainsTime(t time.Time) bool {
	if ym.IsNull() {
		return false
	}
	return YearMonth(ym).ContainsTime(t)
}

// ContainsDate returns true if the given date falls within this year-month.
// Returns false if the NullableYearMonth is null.
func (ym NullableYearMonth) ContainsDate(date Date) bool {
	if ym.IsNull() {
		return false
	}
	return YearMonth(ym).ContainsDate(date)
}
