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

// String implements the fmt.Stringer interface.
func (a IntArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("IntArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a IntArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return notnull.IntArray(a).Value()
}

// Scan implements the sql.Scanner interface.
func (a *IntArray) Scan(src interface{}) error {
	return (*notnull.IntArray)(a).Scan(src)
}
