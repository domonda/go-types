package notnull

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

// IntArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of int64.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type IntArray []int64

// String implements the fmt.Stringer interface.
func (a IntArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("IntArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a IntArray) Value() (driver.Value, error) {
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

	return errors.Errorf("can't convert %T to sqlarray.IntArray", src)
}

func (a *IntArray) scanBytes(src []byte) (err error) {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("can't parse '%s' as sqlarray.IntArray", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(IntArray, len(elements))
	for i, elem := range elements {
		newArray[i], err = strconv.ParseInt(elem, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "Can't parse '%s' as sqlarray.IntArray", string(src))
		}
	}
	*a = newArray

	return nil
}

// MarshalJSON returns a as the JSON encoding of a.
// MarshalJSON implements encoding/json.Marshaler.
func (a IntArray) MarshalJSON() ([]byte, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal([]int64(a))
}
