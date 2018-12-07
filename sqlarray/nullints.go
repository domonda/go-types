package sqlarray

import (
	"database/sql"
	"database/sql/driver"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

// NullInts implements the sql.Scanner and driver.Valuer interfaces
// for a slice of sql.NullInt64.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type NullInts []sql.NullInt64

// Ints returns all NullInts elements as []int64 with NULL elements set to 0.
func (a NullInts) Ints() []int64 {
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

// Value implements the database/sql/driver.Valuer interface
func (a NullInts) Value() (driver.Value, error) {
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
			newArray[i] = sql.NullInt64{Valid: true, Int64: val}
		}
	}
	*a = newArray

	return nil
}
