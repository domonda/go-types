package notnull

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
)

// NullBoolArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of sql.NullBool.
// The nil default value of the slice is returned as an empty (non null) array
// for SQL and JSON.
// Use nullable.NullBoolArray if the nil value should be treated as SQL and JSON null.
type NullBoolArray []sql.NullBool

// Bools returns all NullBoolArray elements as []bool with NULL elements set to false.
func (a NullBoolArray) Bools() []bool {
	if len(a) == 0 {
		return nil
	}

	bools := make([]bool, len(a))
	for i, nb := range a {
		if nb.Valid && nb.Bool {
			bools[i] = true
		}
	}
	return bools
}

// String implements the fmt.Stringer interface.
func (a NullBoolArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("NullBoolArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a NullBoolArray) Value() (driver.Value, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		if a[i].Valid {
			if a[i].Bool {
				b.WriteByte('t')
			} else {
				b.WriteByte('f')
			}
		} else {
			b.WriteString("NULL")
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface
func (a *NullBoolArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("can't convert %T to NullBoolArray", src)
}

func (a *NullBoolArray) scanBytes(src []byte) error {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as NullBoolArray", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullBoolArray, len(elements))
	for i, elem := range elements {
		switch elem {
		case "t":
			newArray[i] = sql.NullBool{Valid: true, Bool: true}
		case "f":
			newArray[i] = sql.NullBool{Valid: true, Bool: false}
		}
	}
	*a = newArray

	return nil
}
