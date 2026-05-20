package notnull

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// FloatArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of float64.
// The nil default value of the slice is returned as an empty (non null) array
// for SQL and JSON.
// Use nullable.FloatArray if the nil value should be treated as SQL and JSON null.
//
// Value and Scan use the PostgreSQL array text format ({1.5,2,3}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type FloatArray []float64

// String implements the fmt.Stringer interface.
func (a FloatArray) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for i := range a {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.FormatFloat(a[i], 'f', -1, 64))
	}
	b.WriteByte(']')
	return b.String()
}

// Contains reports if the passed value is present in a.
func (a FloatArray) Contains(value float64) bool {
	return slices.Contains(a, value)
}

// Value implements the database/sql/driver.Valuer interface.
// It returns the slice as a PostgreSQL float array literal
// like {1.5,2,3}. A nil or empty slice returns the empty array {}.
func (a FloatArray) Value() (driver.Value, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(a[i], 'f', -1, 64))
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface.
// It parses a PostgreSQL float array literal like {1.5,2,3}
// from a string or []byte. A nil source (SQL NULL), an empty
// string, or {} all scan to a non-nil empty slice;
// a notnull array is never nil.
func (a *FloatArray) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = FloatArray{}
		return nil
	}

	return fmt.Errorf("can't convert %T to notnull.FloatArray", src)
}

func (a *FloatArray) scanBytes(src []byte) (err error) {
	if len(src) == 0 {
		*a = FloatArray{}
		return nil
	}
	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as notnull.FloatArray", string(src))
	}
	if len(src) == 2 { // src == "{}"
		*a = FloatArray{}
		return nil
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(FloatArray, len(elements))
	for i, elem := range elements {
		newArray[i], err = strconv.ParseFloat(elem, 64)
		if err != nil {
			return fmt.Errorf("can't parse %q as notnull.FloatArray because of: %w", string(src), err)
		}
	}
	*a = newArray
	return nil
}

// MarshalJSON implements encoding/json.Marshaler.
// A nil or empty array is encoded as the empty JSON array []
// instead of null; a notnull array is never null.
func (a FloatArray) MarshalJSON() ([]byte, error) {
	b := make([]byte, 1, 32)
	b[0] = '['
	for i := range a {
		if i > 0 {
			b = append(b, ',')
		}
		b = strconv.AppendFloat(b, a[i], 'f', -1, 64)
	}
	b = append(b, ']')
	return b, nil
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// JSON null and [] both unmarshal to a non-nil empty slice;
// a notnull array is never nil.
func (a *FloatArray) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = FloatArray{}
		return nil
	}
	var s []float64
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a = FloatArray(s)
	return nil
}

// Len is the number of elements in the collection.
// One of the methods to implement sort.Interface.
func (a FloatArray) Len() int { return len(a) }

// Less reports whether the element with
// index i should sort before the element with index j.
// One of the methods to implement sort.Interface.
func (a FloatArray) Less(i, j int) bool { return a[i] < a[j] }

// Swap swaps the elements with indexes i and j.
// One of the methods to implement sort.Interface.
func (a FloatArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
