package sqlarray

import (
	"database/sql/driver"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

// NullInts implements the sql.Scanner and driver.Valuer interfaces
// for a slice of int64 pointers.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
// A nil int64 pointer element is mapped to SQL NULL.
// Note that allocating many individual memory chunks
// for every slice element may lead to poor performance.
type NullInts []*int64

// Ints returns a int64 slice where all non NULL
// elements of a are set, and all NULL elements are 0.
func (a NullInts) Ints() []int64 {
	if len(a) == 0 {
		return nil
	}

	ints := make([]int64, len(a))
	for i, ptr := range a {
		if ptr != nil {
			ints[i] = *a[i]
		}
	}
	return ints
}

// Value implements the database/sql/driver.Valuer interface
func (a NullInts) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	var b strings.Builder
	b.WriteByte('{')
	for i, intPtr := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		if intPtr == nil {
			b.WriteString("NULL")
		} else {
			b.WriteString(strconv.FormatInt(*intPtr, 10))
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface
func (a *NullInts) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return errors.Errorf("Can't convert %T to sqlarray.NullInts", src)
}

func (a *NullInts) scanBytes(src []byte) error {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("Can't parse '%s' as sqlarray.NullInts", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullInts, len(elements))
	for i, elem := range elements {
		if elem != "NULL" && elem != "null" {
			val, err := strconv.ParseInt(elem, 10, 64)
			if err != nil {
				return errors.Wrapf(err, "Can't parse '%s' as sqlarray.NullInts", string(src))
			}
			newArray[i] = &val
		}
	}
	*a = newArray

	return nil
}
