package nullable

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// NullIntArray is a slice of nullable int64 values, each element stored
// as a Type[int64]. It implements the sql.Scanner and driver.Valuer
// interfaces for a SQL array. JSON marshalling needs no methods on the
// slice itself: encoding/json handles the slice and each Type[int64]
// element marshals a value as a number and a null element as JSON null,
// e.g. [1,null,3].
//
// A nil slice maps to SQL NULL and JSON null,
// a non-nil zero length slice to the empty SQL array '{}' and JSON [].
//
// Value and Scan use the PostgreSQL array text format ({1,NULL,3}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type NullIntArray []Type[int64]

// IsNull returns true if a is nil.
// IsNull implements the Nullable interface.
func (a NullIntArray) IsNull() bool { return a == nil }

// Ints returns all NullIntArray elements as []int64
// with null elements set to 0.
func (a NullIntArray) Ints() []int64 {
	if len(a) == 0 {
		return nil
	}
	ints := make([]int64, len(a))
	for i, n := range a {
		ints[i] = n.GetOr(0)
	}
	return ints
}

// String implements the fmt.Stringer interface.
func (a NullIntArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("NullIntArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface.
// A nil slice returns SQL NULL. Otherwise it returns the slice
// as a PostgreSQL integer array literal like {1,NULL,3}, with
// null elements written as NULL.
func (a NullIntArray) Value() (driver.Value, error) {
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
			b.WriteString(strconv.FormatInt(n.Get(), 10))
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface.
// A nil source scans to a nil slice (SQL NULL), the empty array {}
// to a non-nil empty slice. Otherwise it parses a PostgreSQL
// integer array literal like {1,NULL,3} from a string or []byte,
// NULL elements scanning as a null Type[int64].
func (a *NullIntArray) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("can't convert %T to NullIntArray", src)
}

func (a *NullIntArray) scanBytes(src []byte) error {
	if len(src) == 0 {
		// Reached only for a non-nil empty string or []byte,
		// because Scan maps an untyped nil source to a nil slice.
		*a = NullIntArray{}
		return nil
	}
	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as NullIntArray", string(src))
	}
	if len(src) == 2 { // src == "{}"
		*a = NullIntArray{}
		return nil
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullIntArray, len(elements))
	for i, elem := range elements {
		if elem == "NULL" || elem == "null" {
			continue // leave newArray[i] as a null Type[int64]
		}
		val, err := strconv.ParseInt(elem, 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse %q as NullIntArray because of: %w", string(src), err)
		}
		newArray[i] = TypeFrom(val)
	}
	*a = newArray
	return nil
}
