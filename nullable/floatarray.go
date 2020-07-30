package nullable

import (
	"database/sql/driver"
	"fmt"

	"github.com/domonda/go-types/notnull"
)

// FloatArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of float64.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type FloatArray []float64

// IsNull returns true if a is nil.
// IsNull implements the Nullable interface.
func (a FloatArray) IsNull() bool { return a == nil }

// String implements the fmt.Stringer interface
func (a FloatArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("FloatArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a FloatArray) Value() (driver.Value, error) {
	if a.IsNull() {
		return nil, nil
	}
	return notnull.FloatArray(a).Value()
}

// Scan implements the sql.Scanner interface.
func (a *FloatArray) Scan(src interface{}) error {
	return (*notnull.FloatArray)(a).Scan(src)
}
