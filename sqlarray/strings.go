package sqlarray

import (
	"database/sql/driver"

	"github.com/lib/pq"
)

// Strings implements the sql.Scanner and driver.Valuer interfaces
// for a slice of string.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type Strings = pq.StringArray

// StringsNotNull always returns an non NULL SQL array, even if the slice is nil.
type StringsNotNull []string

// Scan implements the sql.Scanner interface.
func (a *StringsNotNull) Scan(src interface{}) error {
	return (*pq.StringArray)(a).Scan(src)
}

// Value implements the driver.Valuer interface.
func (a StringsNotNull) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	return pq.StringArray(a).Value()
}
