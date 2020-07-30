package nullable

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// NullFloatArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of sql.NullFloat64.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type NullFloatArray []sql.NullFloat64

// IsNull returns true if a is nil.
// IsNull implements the Nullable interface.
func (a NullFloatArray) IsNull() bool { return a == nil }

// Floats returns all NullFloatArray elements as []float64 with NULL elements set to 0.
func (a NullFloatArray) Floats() []float64 {
	if len(a) == 0 {
		return nil
	}

	floats := make([]float64, len(a))
	for i, n := range a {
		if n.Valid {
			floats[i] = n.Float64
		}
	}
	return floats
}

// String implements the fmt.Stringer interface.
func (a NullFloatArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("NullFloatArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a NullFloatArray) Value() (driver.Value, error) {
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
			b.WriteString(strconv.FormatFloat(n.Float64, 'f', -1, 64))
		} else {
			b.WriteString("NULL")
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface
func (a *NullFloatArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("can't convert %T to NullFloatArray", src)
}

func (a *NullFloatArray) scanBytes(src []byte) error {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as NullFloatArray", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullFloatArray, len(elements))
	for i, elem := range elements {
		if elem != "NULL" && elem != "null" {
			val, err := strconv.ParseFloat(elem, 64)
			if err != nil {
				return fmt.Errorf("can't parse %q as NullFloatArray because of: %w", string(src), err)
			}
			newArray[i] = sql.NullFloat64{Valid: true, Float64: val}
		}
	}
	*a = newArray

	return nil
}
