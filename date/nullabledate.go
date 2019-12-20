package date

import (
	"database/sql/driver"
	"time"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/language"
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

// MidnightTimePtrOrNil returns the address of a midnight (00:00) time.Time of date,
// or nil if date.IsZero() returns true.
func (n NullableDate) MidnightTimePtrOrNil() *time.Time {
	return Date(n).MidnightTimePtrOrNil()
}

// Scan implements the database/sql.Scanner interface.
func (n *NullableDate) Scan(value interface{}) (err error) {
	switch x := value.(type) {
	case string:
		d := Date(x)
		if !d.IsZero() {
			d, err = d.Normalized()
			if err != nil {
				return err
			}
		}
		*n = d.NullableDate()
		return nil

	case time.Time:
		*n = OfTime(x).NullableDate()
		return nil

	case nil:
		*n = Null
		return nil
	}

	return errors.Errorf("can't scan value '%#v' of type %T as data.NullableDate", value, value)
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

// IsZero returns true when the date is any of ["", "0000-00-00", "0001-01-01", "null", "NULL"]
// "0001-01-01" is treated as zero because it's the zero value of time.Time.
func (n NullableDate) IsZero() bool {
	return Date(n).IsZero()
}

func (n NullableDate) Date() Date {
	return Date(n)
}

// MidnightTime returns the midnight (00:00) time.Time of the date in UTC,
// or a zero time.Time value if the date is not valid.
func (n NullableDate) MidnightTime() time.Time {
	if n.IsZero() {
		return time.Time{}
	}
	t, err := time.Parse(Layout, string(n))
	if err != nil {
		return time.Time{}
	}
	return t
}

// MidnightTime returns the midnight (00:00) time.Time of the date
// in the given location,
// or a zero time.Time value if the date is not valid.
func (n NullableDate) MidnightTimeInLocation(loc *time.Location) time.Time {
	if n.IsZero() {
		return time.Time{}
	}
	t, err := time.ParseInLocation(Layout, string(n), loc)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ISOWeek returns the ISO 8601 year and week number in which the date occurs.
// Week ranges from 1 to 53. Jan 01 to Jan 03 of year n might belong to
// week 52 or 53 of year n-1, and Dec 29 to Dec 31 might belong to week 1
// of year n+1.
func (n NullableDate) ISOWeek() (year, week int) {
	// Date.ISOWeek can handle zero/null
	return Date(n).ISOWeek()
}

// Format returns n.MidnightTime().Format(layout),
// or an empty string if n is Null or layout is an empty string.
func (n NullableDate) Format(layout string) string {
	if n == Null || layout == "" {
		return ""
	}
	if layout == Layout {
		return string(n)
	}
	return n.MidnightTime().Format(layout)
}
