package nullable

import (
	"database/sql/driver"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/domonda/go-types/notnull"
)

// IntArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of int64.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
//
// Value and Scan use the PostgreSQL array text format ({1,2,3}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type IntArray []int64

// IsNull returns true if a is nil.
// IsNull implements the Nullable interface.
func (a IntArray) IsNull() bool { return a == nil }

// String implements the fmt.Stringer interface.
func (a IntArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("IntArray%v", value)
}

// Contains reports if the passed value is present in a.
func (a IntArray) Contains(value int64) bool {
	return slices.Contains(a, value)
}

// Value implements the database/sql/driver.Valuer interface.
// A nil slice returns SQL NULL. Otherwise it returns the slice
// as a PostgreSQL integer array literal like {1,2,3}, an empty
// slice returning the empty array {}.
func (a IntArray) Value() (driver.Value, error) {
	if a.IsNull() {
		return nil, nil
	}
	return notnull.IntArray(a).Value()
}

// Scan implements the sql.Scanner interface.
// A nil source scans to a nil slice (SQL NULL). Any other source
// scans to a non-nil slice: {} to an empty slice, otherwise the
// parsed PostgreSQL integer array literal like {1,2,3}.
func (a *IntArray) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("can't convert %T to nullable.IntArray", src)
}

func (a *IntArray) scanBytes(src []byte) (err error) {
	if len(src) == 0 {
		// Reached only for a non-nil empty string or []byte,
		// because Scan maps an untyped nil source to a nil slice.
		*a = IntArray{}
		return nil
	}
	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as nullable.IntArray", string(src))
	}
	if len(src) == 2 { // src == "{}"
		*a = IntArray{}
		return nil
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(IntArray, len(elements))
	for i, elem := range elements {
		newArray[i], err = strconv.ParseInt(elem, 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse %q as nullable.IntArray because of: %w", string(src), err)
		}
	}
	*a = newArray
	return nil
}

// Len is the number of elements in the collection.
// One of the methods to implement sort.Interface.
func (a IntArray) Len() int { return len(a) }

// Less reports whether the element with
// index i should sort before the element with index j.
// One of the methods to implement sort.Interface.
func (a IntArray) Less(i, j int) bool { return a[i] < a[j] }

// Swap swaps the elements with indexes i and j.
// One of the methods to implement sort.Interface.
func (a IntArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
