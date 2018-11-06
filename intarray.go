package types

import (
	"database/sql/driver"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

type IntArray []int64

func (a IntArray) Value() (driver.Value, error) {
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
func (a *IntArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return errors.Errorf("Can't convert %T to IntArray", src)
}

func (a *IntArray) scanBytes(src []byte) (err error) {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("Can't parse '%s' as IntArray", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(IntArray, len(elements))
	for i, elem := range elements {
		newArray[i], err = strconv.ParseInt(elem, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "Can't parse '%s' as IntArray", string(src))
		}
	}
	*a = newArray

	return nil
}
