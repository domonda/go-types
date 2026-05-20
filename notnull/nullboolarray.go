package notnull

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// NullBoolArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of sql.NullBool.
// The nil default value of the slice is returned as an empty (non null) array
// for SQL and JSON.
// Use nullable.NullBoolArray if the nil value should be treated as SQL and JSON null.
//
// Value and Scan use the PostgreSQL array text format ({t,f,NULL}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type NullBoolArray []sql.NullBool

// Bools returns all NullBoolArray elements as []bool with NULL elements set to false.
func (a NullBoolArray) Bools() []bool {
	if len(a) == 0 {
		return nil
	}

	bools := make([]bool, len(a))
	for i, nb := range a {
		if nb.Valid && nb.Bool {
			bools[i] = true
		}
	}
	return bools
}

// String implements the fmt.Stringer interface.
func (a NullBoolArray) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for i := range a {
		if i > 0 {
			b.WriteString(", ")
		}
		if a[i].Valid {
			b.WriteString(strconv.FormatBool(a[i].Bool))
		} else {
			b.WriteString("NULL")
		}
	}
	b.WriteByte(']')
	return b.String()
}

// Value implements the database/sql/driver.Valuer interface.
// It returns the slice as a PostgreSQL boolean array literal
// like {t,f,NULL}, with invalid (NULL) elements written as NULL.
// A nil or empty slice returns the empty array {}.
func (a NullBoolArray) Value() (driver.Value, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		if a[i].Valid {
			if a[i].Bool {
				b.WriteByte('t')
			} else {
				b.WriteByte('f')
			}
		} else {
			b.WriteString("NULL")
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface.
// It parses a PostgreSQL boolean array literal like {t,f,NULL}
// from a string or []byte; elements other than t and f
// (including NULL) scan as an invalid sql.NullBool.
// A nil source (SQL NULL), an empty string, or {} all scan to
// a non-nil empty slice; a notnull array is never nil.
func (a *NullBoolArray) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = NullBoolArray{}
		return nil
	}

	return fmt.Errorf("can't convert %T to notnull.NullBoolArray", src)
}

func (a *NullBoolArray) scanBytes(src []byte) error {
	if len(src) == 0 {
		*a = NullBoolArray{}
		return nil
	}
	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as notnull.NullBoolArray", string(src))
	}
	if len(src) == 2 { // src == "{}"
		*a = NullBoolArray{}
		return nil
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullBoolArray, len(elements))
	for i, elem := range elements {
		switch elem {
		case "t":
			newArray[i] = sql.NullBool{Valid: true, Bool: true}
		case "f":
			newArray[i] = sql.NullBool{Valid: true, Bool: false}
		}
	}
	*a = newArray

	return nil
}

// MarshalJSON implements encoding/json.Marshaler.
// Valid elements are encoded as JSON true or false,
// invalid (NULL) elements as JSON null.
// A nil or empty array is encoded as the empty JSON array []
// instead of null; a notnull array is never null.
func (a NullBoolArray) MarshalJSON() ([]byte, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	elements := make([]*bool, len(a))
	for i, nb := range a {
		if nb.Valid {
			b := nb.Bool
			elements[i] = &b
		}
	}
	return json.Marshal(elements)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// JSON null and [] both unmarshal to a non-nil empty slice;
// a notnull array is never nil.
// JSON null elements unmarshal to an invalid sql.NullBool.
func (a *NullBoolArray) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = NullBoolArray{}
		return nil
	}
	var elements []*bool
	if err := json.Unmarshal(data, &elements); err != nil {
		return err
	}
	newArray := make(NullBoolArray, len(elements))
	for i, e := range elements {
		if e != nil {
			newArray[i] = sql.NullBool{Valid: true, Bool: *e}
		}
	}
	*a = newArray
	return nil
}
