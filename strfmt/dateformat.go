package strfmt

import (
	"reflect"
	"strings"
	"time"

	"github.com/guregu/null"
	reflection "github.com/ungerik/go-reflection"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/date"
)

type DateFormat struct {
	Layout     string `json:"layout"`
	NilString  string `json:"nilString"`
	ZeroString string `json:"zeroString"`
}

func (f *DateFormat) AssignString(dest reflect.Value, str string) error {
	str = strings.TrimSpace(str)

	tPtr := new(time.Time)
	if str != "" {
		if f.Layout == "" {
			d, err := date.Normalize(str)
			if err != nil {
				return err
			}
			t := d.MidnightTime()
			tPtr = &t
		} else {
			t, err := time.Parse(f.Layout, str)
			if err != nil {
				return err
			}
			tPtr = &t
		}
		if tPtr.IsZero() {
			return errors.Errorf("Can't assign zero time")
		}
		// if !f.TimeZone.IsLocal() {
		// 	*tPtr = tPtr.In(f.TimeZone.Get())
		// }
	}

	switch ptr := dest.Addr().Interface().(type) {
	case *date.Date:
		if tPtr == nil {
			*ptr = ""
		} else {
			*ptr = date.OfTime(*tPtr)
		}
		return nil

	case *time.Time:
		if tPtr == nil {
			*ptr = time.Time{}
		} else {
			*ptr = *tPtr
		}
		return nil

	case **time.Time:
		*ptr = tPtr
		return nil

	case *null.Time:
		if tPtr == nil {
			*ptr = null.Time{}
		} else {
			*ptr = null.TimeFrom(*tPtr)
		}
		return nil
	}
	return errors.Errorf("AssignString destination type not supported: %T", dest.Interface())
}

func (f *DateFormat) FormatString(val reflect.Value) (string, error) {
	v := reflection.DerefValue(val)
	if reflection.IsNil(v) {
		return f.NilString, nil
	}

	switch x := val.Interface().(type) {
	case dateOrTime:
		if x.IsZero() {
			return f.ZeroString, nil
		}
		return x.Format(f.Layout), nil
	}

	return "", errors.Errorf("Could not format as date string: %T", val)
}

type dateOrTime interface {
	// Format as implemented by time.Time and date.Date
	Format(layout string) string

	// IsZero as implemented by time.Time and date.Date
	IsZero() bool
}
