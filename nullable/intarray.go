package nullable

import (
	"database/sql/driver"
	"fmt"

	"github.com/domonda/go-types/notnull"
)

// IntArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of int64.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type IntArray []int64

// IsNull returns true if a is nil.
// IsNull implements the Nullable interface.
func (a IntArray) IsNull() bool { return a == nil }

// String implements the fmt.Stringer interface.
func (a IntArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("IntArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a IntArray) Value() (driver.Value, error) {
	if a.IsNull() {
		return nil, nil
	}
	return notnull.IntArray(a).Value()
}

// Scan implements the sql.Scanner interface.
func (a *IntArray) Scan(src any) error {
	return (*notnull.IntArray)(a).Scan(src)
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
