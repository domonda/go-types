package nullable

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// NullIntArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of sql.NullInt64.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type NullIntArray []sql.NullInt64

// IsNull returns true if a is nil.
// IsNull implements the Nullable interface.
func (a NullIntArray) IsNull() bool { return a == nil }

// Ints returns all NullIntArray elements as []int64 with NULL elements set to 0.
func (a NullIntArray) Ints() []int64 {
	if len(a) == 0 {
		return nil
	}

	ints := make([]int64, len(a))
	for i, n := range a {
		if n.Valid {
			ints[i] = n.Int64
		}
	}
	return ints
}

// String implements the fmt.Stringer interface.
func (a NullIntArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("NullIntArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a NullIntArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	var b strings.Builder
	b.WriteByte('{')
	for i, n := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		if n.Valid {
			b.WriteString(strconv.FormatInt(n.Int64, 10))
		} else {
			b.WriteString("NULL")
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface
func (a *NullIntArray) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("can't convert %T to NullIntArray", src)
}

func (a *NullIntArray) scanBytes(src []byte) error {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as NullIntArray", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullIntArray, len(elements))
	for i, elem := range elements {
		if elem != "NULL" && elem != "null" {
			val, err := strconv.ParseInt(elem, 10, 64)
			if err != nil {
				return fmt.Errorf("can't parse %q as NullIntArray because of: %w", string(src), err)
			}
			newArray[i] = sql.NullInt64{Valid: true, Int64: val}
		}
	}
	*a = newArray

	return nil
}
