package nullable

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/domonda/go-types/notnull"
)

// NullBoolArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of sql.NullBool.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type NullBoolArray []sql.NullBool

// IsNull returns true if a is nil.
// IsNull implements the Nullable interface.
func (a NullBoolArray) IsNull() bool { return a == nil }

// Bools returns all NullBoolArray elements as []bool with NULL elements set to false.
func (a NullBoolArray) Bools() []bool {
	return notnull.NullBoolArray(a).Bools()
}

// String implements the fmt.Stringer interface.
func (a NullBoolArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("NullBoolArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a NullBoolArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return notnull.NullBoolArray(a).Value()
}

// Scan implements the sql.Scanner interface
func (a *NullBoolArray) Scan(src interface{}) error {
	return (*notnull.NullBoolArray)(a).Scan(src)
}
