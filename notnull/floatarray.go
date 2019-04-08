package notnull

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

// FloatArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of float64.
// A nil slice is mapped to an empty SLQ or JSON array.
type FloatArray []float64

// String implements the fmt.Stringer interface.
func (a FloatArray) String() string {
	value, _ := a.Value()
	return fmt.Sprintf("FloatArray%v", value)
}

// Value implements the database/sql/driver.Valuer interface
func (a FloatArray) Value() (driver.Value, error) {
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
func (a *FloatArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)

	case string:
		return a.scanBytes([]byte(src))

	case nil:
		*a = nil
		return nil
	}

	return errors.Errorf("can't convert %T to FloatArray", src)
}

func (a *FloatArray) scanBytes(src []byte) (err error) {
	if len(src) == 0 {
		*a = nil
	}

	if src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("can't parse '%s' as FloatArray", string(src))
	}

	elements := strings.Split(string(src[1:len(src)-1]), ",")
	newArray := make(FloatArray, len(elements))
	for i, elem := range elements {
		newArray[i], err = strconv.ParseFloat(elem, 64)
		if err != nil {
			return errors.Wrapf(err, "Can't parse '%s' as FloatArray", string(src))
		}
	}
	*a = newArray

	return nil
}

// MarshalJSON returns a as the JSON encoding of a.
// MarshalJSON implements encoding/json.Marshaler.
func (a FloatArray) MarshalJSON() ([]byte, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal([]float64(a))
}
