package notnull

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/domonda/go-types/internal/pq"
)

// StringArray implements the sql.Scanner, driver.Valuer, and encoding/json.Marshaler interfaces
// for a slice of strings.
// The nil default value of the slice is returned as an empty (non null) array
// for SQL and JSON.
// Use nullable.StringArray if the nil value should be treated as SQL and JSON null.
type StringArray []string

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(src any) error {
	return (*pq.StringArray)(a).Scan(src)
}

// Value implements the driver.Valuer interface.
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	return pq.StringArray(a).Value()
}

// MarshalJSON returns a as the JSON encoding of a.
// MarshalJSON implements encoding/json.Marshaler.
func (a StringArray) MarshalJSON() ([]byte, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(a))
}

// Len is the number of elements in the collection.
// One of the methods to implement sort.Interface.
func (a StringArray) Len() int { return len(a) }

// Less reports whether the element with
// index i should sort before the element with index j.
// One of the methods to implement sort.Interface.
func (a StringArray) Less(i, j int) bool { return a[i] < a[j] }

// Swap swaps the elements with indexes i and j.
// One of the methods to implement sort.Interface.
func (a StringArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
