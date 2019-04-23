package notnull

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/lib/pq"
)

// StringArray implements the sql.Scanner, driver.Valuer, and encoding/json.Marshaler interfaces
// for a slice of strings.
// The nil default value of the slice is returned as an empty (non null) array
// for SQL and JSON.
// Use nullable.StringArray if the nil value should be treated as SQL and JSON null.
type StringArray []string

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(src interface{}) error {
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
