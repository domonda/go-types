package notnull

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// IntArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of int64.
// The nil default value of the slice is returned as an empty (non null) array
// for SQL and JSON.
// Use nullable.IntArray if the nil value should be treated as SQL and JSON null.
type IntArray []int64

// String implements the fmt.Stringer interface.
func (a IntArray) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for i := range a {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.FormatInt(a[i], 10))
	}
	b.WriteByte(']')
	return b.String()
}

// Value implements the database/sql/driver.Valuer interface
func (a IntArray) Value() (driver.Value, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatInt(a[i], 10))
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface.
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

	return fmt.Errorf("can't convert %T to notnull.IntArray", src)
}

func (a *IntArray) scanBytes(src []byte) (err error) {
	if len(src) == 0 {
		*a = nil
		return nil
	}
	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as notnull.IntArray", string(src))
	}
	if len(src) == 2 { // src == "{}"
		*a = nil
		return nil
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(IntArray, len(elements))
	for i, elem := range elements {
		newArray[i], err = strconv.ParseInt(elem, 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse %q as notnull.IntArray because of: %w", string(src), err)
		}
	}
	*a = newArray
	return nil
}

// MarshalJSON returns a as the JSON encoding of a.
// MarshalJSON implements encoding/json.Marshaler.
func (a IntArray) MarshalJSON() ([]byte, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal([]int64(a))
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
