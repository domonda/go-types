package nullable

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/domonda/go-pretty"
)

// TimeNull is a null Time value.
//
// Note: use Time.IsNull or IsNotNull to check for null
// instead of comparing a Time with TimeNull
// because Time.IsNull uses time.Time.IsZero internally
// which can return true for times that are not
// the empty time.Time{} default value.
var TimeNull Time

// Time represents a time.Time where the zero time instant
// (which is the empty default value of the type)
// is interpreted as SQL NULL and JSON null.
// It implements the sql.Scanner and driver.Valuer interfaces
// and also json.Marshaler and json.Unmarshaler.
// It assumes that zero time instant is never used
// in any real life application so it's free
// to be used as magical value for representing NULL.
// IsNull uses time.Time.IsZero internally.
type Time struct {
	time.Time
}

// TimeNow returns the current time
func TimeNow() Time {
	return Time{Time: time.Now()}
}

// TimeParse parses a time value with the provided layout
// using time.Parse(layout, value)
// except for when value is on of "", "null", "NULL",
// then a null/zero time and no error are returned.
func TimeParse(layout, value string) (Time, error) {
	if value == "" || value == "null" || value == "NULL" {
		return Time{}, nil
	}
	t, err := time.Parse(layout, value)
	if err != nil {
		return Time{}, err
	}
	return Time{Time: t}, nil
}

// TimeParseInLocation parses a time value with the provided layout
// and location using time.ParseInLocation(layout, value, loc)
// except for when value is on of "", "null", "NULL",
// then a null/zero time and no error are returned.
func TimeParseInLocation(layout, value string, loc *time.Location) (Time, error) {
	if value == "" || value == "null" || value == "NULL" {
		return Time{}, nil
	}
	t, err := time.ParseInLocation(layout, value, loc)
	if err != nil {
		return Time{}, err
	}
	return Time{Time: t}, nil
}

// TimeFrom returns a nullable.Time from a time.Time
func TimeFrom(t time.Time) Time {
	return Time{Time: t}
}

// TimeFromPtr returns a nullable.Time from a time.Time pointer
// with nil interpreted as null.
func TimeFromPtr(ptr *time.Time) Time {
	if ptr == nil {
		return Time{}
	}
	return Time{Time: *ptr}
}

// Ptr returns a pointer to Time or nil if n is null
func (n Time) Ptr() *time.Time {
	if n.IsNull() {
		return nil
	}
	return &n.Time
}

// UTC returns the time in UTC or null if n is null
func (n Time) UTC() Time {
	if n.IsNull() {
		return n
	}
	return Time{Time: n.Time.UTC()}
}

// Add returns n+duration or null if n is null
func (n Time) Add(duration time.Duration) Time {
	if n.IsNull() {
		return n
	}
	return Time{Time: n.Time.Add(duration)}
}

// AddDate returns the nullable time corresponding to adding the
// given number of years, months, and days to t.
// For example, AddDate(-1, 2, 3) applied to January 1, 2011
// returns March 4, 2010.
//
// AddDate normalizes its result in the same way that Date does,
// so, for example, adding one month to October 31 yields
// December 1, the normalized form for November 31.
//
// Returns null if n is null.
func (n Time) AddDate(years int, months int, days int) Time {
	if n.IsNull() {
		return n
	}
	return Time{Time: n.Time.AddDate(years, months, days)}
}

// Equal reports whether n and o represent the same time instant
// or both are null.
// Two times can be equal even if they are in different locations.
// For example, 6:00 +0200 and 4:00 UTC are Equal.
// See the documentation on the Time type for the pitfalls of using == with
// Time values; most code should use Equal instead.
func (n Time) Equal(o Time) bool {
	return (n.IsNull() && o.IsNull()) || n.Time.Equal(o.Time)
}

// IsNull returns true if the Time is null.
// Uses time.Time.IsZero internally.
// IsNull implements the Nullable interface.
func (n Time) IsNull() bool {
	return n.Time.IsZero()
}

// IsNotNull returns true if the Time is not null.
// Uses time.Time.IsZero internally.
func (n Time) IsNotNull() bool {
	return !n.IsNull()
}

// String returns Time.String() or "NULL" if n is null.
func (n Time) String() string {
	return n.StringOr("NULL")
}

// StringOr returns Time.String() or the passed nullStr if n is null.
func (n Time) StringOr(nullStr string) string {
	if n.IsNull() {
		return nullStr
	}
	return n.Time.String()
}

// Format the time using time.Time.Format
// or return and empty string if n is null.
func (n Time) Format(layout string) string {
	if n.IsNull() {
		return ""
	}
	return n.Time.Format(layout)
}

// AppendFormat the time to b using time.Time.AppendFormat
// or b if n is null.
func (n Time) AppendFormat(b []byte, layout string) []byte {
	if n.IsNull() {
		return b
	}
	return n.Time.AppendFormat(b, layout)
}

// Get returns the non nullable time.Time value
// or panics if the Time is null.
// Note: check with IsNull before using Get!
func (n Time) Get() time.Time {
	if n.IsNull() {
		panic("NULL nullable.Time")
	}
	return n.Time
}

// Set a time.Time.
// Note that if t.IsZero() then n will be set to null.
func (n *Time) Set(t time.Time) {
	n.Time = t
}

// SetNull sets the time to its null value
func (n *Time) SetNull() {
	n.Time = time.Time{}
}

// Scan implements the database/sql.Scanner interface.
func (n *Time) Scan(value any) error {
	switch t := value.(type) {
	case nil:
		*n = Time{}
		return nil

	case time.Time:
		n.Time = t
		return nil

	default:
		return fmt.Errorf("can't scan %T as nullable.Time", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n Time) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.Time, nil
}

// UnarshalJSON implements encoding/json.Unmarshaler.
// Interprets []byte(nil), []byte(""), []byte("null") as null.
func (n *Time) UnmarshalJSON(sourceJSON []byte) error {
	if len(sourceJSON) == 0 || bytes.Equal(sourceJSON, []byte("null")) /*|| bytes.Equal(sourceJSON, []byte(`"NULL"`))*/ {
		*n = Time{}
		return nil
	}
	return json.Unmarshal(sourceJSON, &n.Time)
}

// MarshalJSON implements encoding/json.Marshaler
func (n Time) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte("null"), nil
	}
	return json.Marshal(n.Time)
}

// MarshalText implements the encoding.TextMarshaler interface.
// The time is formatted in RFC 3339 format, with sub-second precision added if present.
// "NULL" is returned as text if the time is null.
func (n Time) MarshalText() ([]byte, error) {
	if n.IsNull() {
		return []byte("NULL"), nil
	}
	return n.Time.MarshalText()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The time is expected to be in RFC 3339 format.
// Empty text, "null", or "NULL" will set the time to null.
func (n *Time) UnmarshalText(text []byte) error {
	if len(text) == 0 || bytes.EqualFold(text, []byte("NULL")) {
		n.SetNull()
		return nil
	}
	return n.Time.UnmarshalText(text)
}

// PrettyPrint implements the pretty.Printable interface
func (n Time) PrettyPrint(w io.Writer) {
	if n.IsNull() {
		w.Write([]byte("null")) //#nosec G104 -- go-pretty does not check write errors
	} else {
		pretty.Fprint(w, n.Time)
	}
}
