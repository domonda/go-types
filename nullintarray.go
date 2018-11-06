package types

import (
	"database/sql/driver"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

type NullIntArray []*int64

func (a NullIntArray) Value() (driver.Value, error) {
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

// Scan implements the sql.Scanner interface.
func (a *NullIntArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return errors.Errorf("Can't convert %T to NullIntArray", src)
}

func (a *NullIntArray) scanBytes(src []byte) error {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("Can't parse '%s' as NullIntArray", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(NullIntArray, len(elements))
	for i, elem := range elements {
		if elem != "NULL" {
			val, err := strconv.ParseInt(elem, 10, 64)
			if err != nil {
				return errors.Wrapf(err, "Can't parse '%s' as NullIntArray", string(src))
			}
			newArray[i] = &val
		}
	}
	*a = newArray

	return nil
}
