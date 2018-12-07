package sqlarray

import (
	"database/sql"
	"database/sql/driver"
	"strings"

	"github.com/domonda/errors"
)

// NullBools implements the sql.Scanner and driver.Valuer interfaces
// for a slice of sql.NullBool.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type NullBools []sql.NullBool

// Bools returns all NullBools elements as []float64 with NULL elements set to false.
func (a NullBools) Bools() []bool {
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

// Value implements the database/sql/driver.Valuer interface
func (a NullBools) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

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
func (a *NullBools) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return errors.Errorf("Can't convert %T to sqlarray.NullBools", src)
}

func (a *NullBools) scanBytes(src []byte) error {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("Can't parse '%s' as sqlarray.NullBools", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullBools, len(elements))
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
