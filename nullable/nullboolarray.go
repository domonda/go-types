package nullable

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// NullBoolArray is a slice of nullable bool values, each element stored
// as a Type[bool]. It implements the sql.Scanner and driver.Valuer
// interfaces for a SQL array. JSON marshalling needs no methods on the
// slice itself: encoding/json handles the slice and each Type[bool]
// element marshals a value as true/false and a null element as JSON
// null, e.g. [true,null,false].
//
// A nil slice maps to SQL NULL and JSON null,
// a non-nil zero length slice to the empty SQL array '{}' and JSON [].
//
// Value and Scan use the PostgreSQL array text format ({t,NULL,f}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type NullBoolArray []Type[bool]

// IsNull returns true if a is nil.
// IsNull implements the Nullable interface.
func (a NullBoolArray) IsNull() bool { return a == nil }

// Bools returns all NullBoolArray elements as []bool
// with null elements set to false.
func (a NullBoolArray) Bools() []bool {
	if len(a) == 0 {
		return nil
	}
	bools := make([]bool, len(a))
	for i, n := range a {
		bools[i] = n.GetOr(false)
	}
	return bools
}

// String implements the fmt.Stringer interface.
func (a NullBoolArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("NullBoolArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface.
// A nil slice returns SQL NULL. Otherwise it returns the slice
// as a PostgreSQL boolean array literal like {t,NULL,f}, with
// null elements written as NULL.
func (a NullBoolArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, n := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		switch {
		case n.IsNull():
			b.WriteString("NULL")
		case n.Get():
			b.WriteByte('t')
		default:
			b.WriteByte('f')
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface.
// A nil source scans to a nil slice (SQL NULL), the empty array {}
// to a non-nil empty slice. Otherwise it parses a PostgreSQL
// boolean array literal like {t,NULL,f} from a string or []byte;
// elements other than t and f scan as a null Type[bool].
func (a *NullBoolArray) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("can't convert %T to NullBoolArray", src)
}

func (a *NullBoolArray) scanBytes(src []byte) error {
	if len(src) == 0 {
		// Reached only for a non-nil empty string or []byte,
		// because Scan maps an untyped nil source to a nil slice.
		*a = NullBoolArray{}
		return nil
	}
	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as NullBoolArray", string(src))
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
			newArray[i] = TypeFrom(true)
		case "f":
			newArray[i] = TypeFrom(false)
			// Any other element (NULL, etc.) stays a null Type[bool].
		}
	}
	*a = newArray
	return nil
}
