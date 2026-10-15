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

// YearQuarterNull is the null value for NullableYearQuarter (empty string).
// Value writes it as SQL NULL and MarshalJSON as JSON null.
const YearQuarterNull NullableYearQuarter = ""

// Compile-time check that NullableYearQuarter implements nullable.NullSetable[YearQuarter]
var _ nullable.NullSetable[YearQuarter] = (*NullableYearQuarter)(nil)

// Compile-time check that NullableYearQuarter implements types.NormalizableValidator[NullableYearQuarter]
var _ types.NormalizableValidator[NullableYearQuarter] = NullableYearQuarter("")

// NullableYearQuarter is a year-quarter string in YYYY-Q# format (e.g. "2024-Q3")
// that can also represent a NULL value as an empty string.
type NullableYearQuarter string

// Validate returns nil if the year-quarter is null or can be normalized,
// see YearQuarter.Validate.
func (yq NullableYearQuarter) Validate() error {
	if yq.IsNull() {
		return nil
	}
	return YearQuarter(yq).Validate()
}

// Valid returns true if the year-quarter is null or can be normalized,
// see YearQuarter.Validate.
func (yq NullableYearQuarter) Valid() bool {
	return yq.Validate() == nil
}

// ValidAndNormalized returns true if the year-quarter is valid and already normalized,
// or if it is the empty YearQuarterNull. A whitespace-only or zero value is null and
// valid, but not normalized.
func (yq NullableYearQuarter) ValidAndNormalized() bool {
	norm, err := yq.Normalized()
	return err == nil && yq == norm
}

// Normalized returns YearQuarterNull for a null year-quarter, see IsNull,
// else the year-quarter normalized like YearQuarter.Normalized,
// or unchanged together with an error if it can't be normalized.
func (yq NullableYearQuarter) Normalized() (NullableYearQuarter, error) {
	if yq.IsNull() {
		return YearQuarterNull, nil
	}
	norm, err := YearQuarter(yq).Normalized()
	return NullableYearQuarter(norm), err
}

// ValidAndNotNull returns if the year-quarter is valid and not null, see IsNull.
func (yq NullableYearQuarter) ValidAndNotNull() bool {
	return yq.IsNotNull() && YearQuarter(yq).Valid()
}

// IsZero returns true if the year-quarter is zero like YearQuarter.IsZero,
// which is the same as IsNull.
func (yq NullableYearQuarter) IsZero() bool {
	return YearQuarter(yq).IsZero()
}

// IsNull returns true if the NullableYearQuarter is null or zero, see YearQuarter.IsZero,
// which includes a string of only whitespace and the zero values
// like "0000-Q0", like NullableDate.IsNull includes the zero dates.
// IsNull implements the nullable.Nullable interface.
func (yq NullableYearQuarter) IsNull() bool {
	return YearQuarter(yq).IsZero()
}

// IsNotNull returns true if the NullableYearQuarter is not null, see IsNull.
func (yq NullableYearQuarter) IsNotNull() bool {
	return !yq.IsNull()
}

// Get returns the non-nullable YearQuarter value or panics if the NullableYearQuarter is null.
// Note: check with IsNull before using Get!
func (yq NullableYearQuarter) Get() YearQuarter {
	if yq.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", yq))
	}
	return YearQuarter(yq)
}

// GetOr returns the non-nullable YearQuarter value or the passed defaultYearQuarter if the NullableYearQuarter is null.
func (yq NullableYearQuarter) GetOr(defaultYearQuarter YearQuarter) YearQuarter {
	if yq.IsNull() {
		return defaultYearQuarter
	}
	return YearQuarter(yq)
}

// Set sets a YearQuarter for this NullableYearQuarter.
func (yq *NullableYearQuarter) Set(yearQuarter YearQuarter) {
	*yq = NullableYearQuarter(yearQuarter)
}

// SetNull sets the NullableYearQuarter to null.
func (yq *NullableYearQuarter) SetNull() {
	*yq = YearQuarterNull
}

// String returns the year-quarter as a string in YYYY-Q# format.
// String implements the fmt.Stringer interface.
func (yq NullableYearQuarter) String() string {
	return string(yq)
}

// StringOr returns the NullableYearQuarter as string
// or the passed nullString if the NullableYearQuarter is null.
func (yq NullableYearQuarter) StringOr(nullString string) string {
	if yq.IsNull() {
		return nullString
	}
	return yq.String()
}

// ScanString implements the strfmt.Scannable interface.
// The strings "NULL", "null" and "nil" set the NullableYearQuarter to null.
// Without validate it stores source like Scan, with validate it stores
// the normalized year-quarter or returns the error of Normalized.
func (yq *NullableYearQuarter) ScanString(source string, validate bool) error {
	return internal.ScanStringNormalized(yq, source, validate, true, NullableYearQuarter.Normalized, NullableYearQuarter.IsNull)
}

// Scan implements the database/sql.Scanner interface.
// A valid year-quarter from a string or []byte is stored normalized, a time.Time
// as its year-quarter, SQL NULL and a null year-quarter (see IsNull) as YearQuarterNull, and a
// year-quarter that can't be normalized unchanged without an error, so that one
// invalid value doesn't fail a whole query; Validate reports it.
func (yq *NullableYearQuarter) Scan(value any) error {
	if t, ok := value.(time.Time); ok {
		value = string(YearQuarterOfTime(t))
	}
	return internal.ScanNormalized(yq, value, NullableYearQuarter.Normalized, NullableYearQuarter.IsNull)
}

// Value implements the database/sql/driver.Valuer interface.
// It returns the normalized year-quarter, SQL NULL if the NullableYearQuarter is null,
// or the error of Normalized for a year-quarter that can't be normalized.
func (yq NullableYearQuarter) Value() (driver.Value, error) {
	return internal.ValueNormalized(yq, NullableYearQuarter.Normalized, NullableYearQuarter.IsNull)
}

// MarshalJSON implements encoding/json.Marshaler by writing the JSON null
// value if the NullableYearQuarter is null, else the normalized year-quarter,
// or the year-quarter unchanged if it can't be normalized.
func (yq NullableYearQuarter) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(yq, true, NullableYearQuarter.Normalized, NullableYearQuarter.IsNull)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null sets the
// NullableYearQuarter to null, so that decoding null into a field that holds
// a value resets it, and any other JSON value is an error.
func (yq *NullableYearQuarter) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(yq, data, true, NullableYearQuarter.Normalized, NullableYearQuarter.IsNull)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// but with "" for null, so that encoding/xml and strfmt write the
// normalized year-quarter. encoding/json writes map keys of string types
// as they are and doesn't call MarshalText for them.
func (yq NullableYearQuarter) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(yq, NullableYearQuarter.Normalized, NullableYearQuarter.IsNull)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (yq *NullableYearQuarter) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(yq, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (yq *NullableYearQuarter) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(yq, d, start, NullableYearQuarter.Normalized, NullableYearQuarter.IsNull)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (yq *NullableYearQuarter) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(yq, attr, NullableYearQuarter.Normalized, NullableYearQuarter.IsNull)
}

// JSONSchema returns the JSON schema definition for the NullableYearQuarter type.
// Implements the jsonschema.JSONSchemaProvider interface.
func (NullableYearQuarter) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable Year Quarter",
		OneOf: []*jsonschema.Schema{
			{
				Type:    "string",
				Pattern: `^[0-9]{4}-Q[1-4]$`,
			},
			{Type: "null"},
		},
		Default: YearQuarterNull,
	}
}

// Compare compares the normalized year-quarter with another normalized
// NullableYearQuarter, an invalid one as it is. Null is before every year-quarter.
// Returns -1 if yq is before other, +1 if after, 0 if equal.
func (yq NullableYearQuarter) Compare(other NullableYearQuarter) int {
	yq, _ = yq.Normalized()
	other, _ = other.Normalized()
	return strings.Compare(string(yq), string(other))
}

// Year returns the year component of the year-quarter.
// Returns 0 if the year-quarter is null or not valid.
func (yq NullableYearQuarter) Year() int {
	if yq.IsNull() {
		return 0
	}
	return YearQuarter(yq).Year()
}

// Quarter returns the quarter component of the year-quarter (1-4).
// Returns 0 if the year-quarter is null or not valid.
func (yq NullableYearQuarter) Quarter() int {
	if yq.IsNull() {
		return 0
	}
	return YearQuarter(yq).Quarter()
}

// FirstMonth returns the first month of the quarter.
// Returns 0 if the year-quarter is null or not valid.
func (yq NullableYearQuarter) FirstMonth() time.Month {
	if yq.IsNull() {
		return 0
	}
	return YearQuarter(yq).FirstMonth()
}

// Date returns a NullableDate for the given month (1-3 within quarter) and day.
// Returns Null if the NullableYearQuarter is null or not valid.
func (yq NullableYearQuarter) Date(monthInQuarter, day int) NullableDate {
	if yq.IsNull() {
		return Null
	}
	return NullableDate(YearQuarter(yq).Date(monthInQuarter, day))
}

// DateRange returns the first and last NullableDate of the quarter.
// Returns Null dates if the NullableYearQuarter is null or not valid.
func (yq NullableYearQuarter) DateRange() (fromDate, untilDate NullableDate) {
	if yq.IsNull() {
		return Null, Null
	}
	from, until := YearQuarter(yq).DateRange()
	return NullableDate(from), NullableDate(until)
}

// AddYears returns a new NullableYearQuarter with the specified number of years added.
// Returns null if the NullableYearQuarter is null
// and an invalid NullableYearQuarter unchanged.
func (yq NullableYearQuarter) AddYears(years int) NullableYearQuarter {
	if yq.IsNull() {
		return YearQuarterNull
	}
	return NullableYearQuarter(YearQuarter(yq).AddYears(years))
}

// AddQuarters returns a new NullableYearQuarter with the specified number of quarters added.
// Returns null if the NullableYearQuarter is null
// and an invalid NullableYearQuarter unchanged.
// Handles quarter overflow by adjusting the year accordingly.
func (yq NullableYearQuarter) AddQuarters(quarters int) NullableYearQuarter {
	if yq.IsNull() {
		return YearQuarterNull
	}
	return NullableYearQuarter(YearQuarter(yq).AddQuarters(quarters))
}

// ContainsTime returns true if the given time falls within this year-quarter.
// Returns false if the NullableYearQuarter is null.
func (yq NullableYearQuarter) ContainsTime(t time.Time) bool {
	if yq.IsNull() {
		return false
	}
	return YearQuarter(yq).ContainsTime(t)
}

// ContainsDate returns true if the given date falls within this year-quarter.
// Returns false if the NullableYearQuarter is null.
func (yq NullableYearQuarter) ContainsDate(date Date) bool {
	if yq.IsNull() {
		return false
	}
	return YearQuarter(yq).ContainsDate(date)
}

// ContainsYearMonth returns true if the given year-month falls within this year-quarter.
// Returns false if the NullableYearQuarter is null.
func (yq NullableYearQuarter) ContainsYearMonth(ym YearMonth) bool {
	if yq.IsNull() {
		return false
	}
	return YearQuarter(yq).ContainsYearMonth(ym)
}
