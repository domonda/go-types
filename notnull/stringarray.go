package notnull

import (
	"database/sql/driver"
	"encoding/json"
	"slices"

	"github.com/domonda/go-types/internal/pq"
)

// StringArray implements the sql.Scanner, driver.Valuer, and encoding/json.Marshaler interfaces
// for a slice of strings.
// The nil default value of the slice is returned as an empty (non null) array
// for SQL and JSON.
// Use nullable.StringArray if the nil value should be treated as SQL and JSON null.
//
// Value and Scan use the PostgreSQL array text format ({"a","b"}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type StringArray []string

// Contains reports if the passed value is present in a.
func (a StringArray) Contains(value string) bool {
	return slices.Contains(a, value)
}

// Scan implements the sql.Scanner interface.
// It parses a PostgreSQL text array literal like {"a","b"}
// (quoted or unquoted elements) from a string or []byte.
// A nil source (SQL NULL) or {} scan to a non-nil empty slice;
// a notnull array is never nil.
func (a *StringArray) Scan(src any) error {
	if src == nil {
		*a = StringArray{}
		return nil
	}
	return (*pq.StringArray)(a).Scan(src)
}

// Value implements the driver.Valuer interface.
// It returns the slice as a PostgreSQL text array literal like
// {"a","b"} with each element quoted and escaped. A nil or empty
// slice returns the empty array {}.
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	return pq.StringArray(a).Value()
}

// MarshalJSON implements encoding/json.Marshaler.
// A nil or empty array is encoded as the empty JSON array []
// instead of null; a notnull array is never null.
func (a StringArray) MarshalJSON() ([]byte, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(a))
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// JSON null and [] both unmarshal to a non-nil empty slice;
// a notnull array is never nil.
func (a *StringArray) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = StringArray{}
		return nil
	}
	var s []string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a = StringArray(s)
	return nil
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
