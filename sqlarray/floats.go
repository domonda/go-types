package sqlarray

import (
	"database/sql/driver"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

// Floats implements the sql.Scanner and driver.Valuer interfaces
// for a slice of float64.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type Floats []float64

// Value implements the database/sql/driver.Valuer interface
func (a Floats) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	var b strings.Builder
	b.WriteByte('{')
	for i := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(a[i], 'f', -1, 64))
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements the sql.Scanner interface.
func (a *Floats) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return errors.Errorf("Can't convert %T to Floats", src)
}

func (a *Floats) scanBytes(src []byte) (err error) {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("Can't parse '%s' as Floats", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(Floats, len(elements))
	for i, elem := range elements {
		newArray[i], err = strconv.ParseFloat(elem, 64)
		if err != nil {
			return errors.Wrapf(err, "Can't parse '%s' as Floats", string(src))
		}
	}
	*a = newArray

	return nil
}
