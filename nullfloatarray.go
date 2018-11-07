package types

import (
	"database/sql/driver"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

// NullFloatArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of float64 pointers.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
// A nil float64 pointer element is mapped to SQL NULL.
type NullFloatArray []*float64

// Value implements the database/sql/driver.Valuer interface
func (a NullFloatArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	var b strings.Builder
	b.WriteByte('{')
	for i, floatPtr := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		if floatPtr == nil {
			b.WriteString("NULL")
		} else {
			b.WriteString(strconv.FormatFloat(*floatPtr, 'f', -1, 64))
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

	return errors.Errorf("Can't convert %T to NullFloatArray", src)
}

func (a *NullFloatArray) scanBytes(src []byte) error {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("Can't parse '%s' as NullFloatArray", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullFloatArray, len(elements))
	for i, elem := range elements {
		if elem != "NULL" {
			val, err := strconv.ParseFloat(elem, 64)
			if err != nil {
				return errors.Wrapf(err, "Can't parse '%s' as NullFloatArray", string(src))
			}
			newArray[i] = &val
		}
	}
	*a = newArray

	return nil
}
