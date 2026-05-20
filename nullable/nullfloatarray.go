package nullable

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// NullFloatArray is a slice of nullable float64 values, each element
// stored as a Type[float64]. It implements the sql.Scanner and
// driver.Valuer interfaces for a SQL array. JSON marshalling needs no
// methods on the slice itself: encoding/json handles the slice and each
// Type[float64] element marshals a value as a number and a null element
// as JSON null, e.g. [1.5,null,3].
//
// A nil slice maps to SQL NULL and JSON null,
// a non-nil zero length slice to the empty SQL array '{}' and JSON [].
//
// Value and Scan use the PostgreSQL array text format ({1.5,NULL,3}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type NullFloatArray []Type[float64]

// IsNull returns true if a is nil.
// IsNull implements the Nullable interface.
func (a NullFloatArray) IsNull() bool { return a == nil }

// Floats returns all NullFloatArray elements as []float64
// with null elements set to 0.
func (a NullFloatArray) Floats() []float64 {
	if len(a) == 0 {
		return nil
	}
	floats := make([]float64, len(a))
	for i, n := range a {
		floats[i] = n.GetOr(0)
	}
	return floats
}

// String implements the fmt.Stringer interface.
func (a NullFloatArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("NullFloatArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface.
// A nil slice returns SQL NULL. Otherwise it returns the slice
// as a PostgreSQL float array literal like {1.5,NULL,3}, with
// null elements written as NULL.
func (a NullFloatArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, n := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		if n.IsNull() {
			b.WriteString("NULL")
		} else {
			b.WriteString(strconv.FormatFloat(n.Get(), 'f', -1, 64))
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface.
// A nil source scans to a nil slice (SQL NULL), the empty array {}
// to a non-nil empty slice. Otherwise it parses a PostgreSQL
// float array literal like {1.5,NULL,3} from a string or []byte,
// NULL elements scanning as a null Type[float64].
func (a *NullFloatArray) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("can't convert %T to NullFloatArray", src)
}

func (a *NullFloatArray) scanBytes(src []byte) error {
	if len(src) == 0 {
		// Reached only for a non-nil empty string or []byte,
		// because Scan maps an untyped nil source to a nil slice.
		*a = NullFloatArray{}
		return nil
	}
	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as NullFloatArray", string(src))
	}
	if len(src) == 2 { // src == "{}"
		*a = NullFloatArray{}
		return nil
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullFloatArray, len(elements))
	for i, elem := range elements {
		if elem == "NULL" || elem == "null" {
			continue // leave newArray[i] as a null Type[float64]
		}
		val, err := strconv.ParseFloat(elem, 64)
		if err != nil {
			return fmt.Errorf("can't parse %q as NullFloatArray because of: %w", string(src), err)
		}
		newArray[i] = TypeFrom(val)
	}
	*a = newArray
	return nil
}
