package nullable

import (
	"database/sql/driver"
	"strconv"
	"strings"

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
	if a.IsNull() {
		return "NULL"
	}
	var b strings.Builder
	b.WriteByte('[')
	for i := range a {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.FormatFloat(a[i], 'f', -1, 64))
	}
	b.WriteByte(']')
	return b.String()
}

func (a FloatArray) StringOr(nilStr string) string {
	if a.IsNull() {
		return nilStr
	}
	return a.String()
}

// Value implements the database/sql/driver.Valuer interface
func (a FloatArray) Value() (driver.Value, error) {
	if a.IsNull() {
		return nil, nil
	}
	return notnull.FloatArray(a).Value()
}

// Scan implements the sql.Scanner interface.
func (a *FloatArray) Scan(src any) error {
	return (*notnull.FloatArray)(a).Scan(src)
}

// Len is the number of elements in the collection.
// One of the methods to implement sort.Interface.
func (a FloatArray) Len() int { return len(a) }

// Less reports whether the element with
// index i should sort before the element with index j.
// One of the methods to implement sort.Interface.
func (a FloatArray) Less(i, j int) bool { return a[i] < a[j] }

// Swap swaps the elements with indexes i and j.
// One of the methods to implement sort.Interface.
func (a FloatArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
