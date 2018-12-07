package sqlarray

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

// Ints implements the sql.Scanner and driver.Valuer interfaces
// for a slice of int64.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type Ints []int64

// String implements the fmt.Stringer interface.
func (a Ints) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("Ints%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a Ints) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

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
func (a *Ints) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return errors.Errorf("Can't convert %T to sqlarray.Ints", src)
}

func (a *Ints) scanBytes(src []byte) (err error) {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("Can't parse '%s' as sqlarray.Ints", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(Ints, len(elements))
	for i, elem := range elements {
		newArray[i], err = strconv.ParseInt(elem, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "Can't parse '%s' as sqlarray.Ints", string(src))
		}
	}
	*a = newArray

	return nil
}
