package date

import (
	"reflect"
	"strings"
	"time"

	"github.com/guregu/null"
	reflection "github.com/ungerik/go-reflection"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/language"
)

type Format struct {
	Layout     string `json:"layout"`
	NilString  string `json:"nilString"`
	ZeroString string `json:"zeroString"`
}

func (f *Format) Format(date Date) string {
	return date.Format(f.Layout)
}

// Parse implements the strfmt.Parser interface
func (f *Format) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	date, err := Normalize(str, langHints...)
	if err != nil {
		return "", err
	}
	return f.Format(date), nil
}

func (f *Format) AssignString(dest reflect.Value, source string) error {
	source = strings.TrimSpace(source)

	tPtr := new(time.Time)
	if source != "" {
		if f.Layout == "" {
			d, err := Normalize(source)
			if err != nil {
				return err
			}
			t := d.MidnightTime()
			tPtr = &t
		} else {
			t, err := time.Parse(f.Layout, source)
			if err != nil {
				return err
			}
			tPtr = &t
		}
		if tPtr.IsZero() {
			return errors.Errorf("can't assign zero time")
		}
		// if !f.TimeZone.IsLocal() {
		// 	*tPtr = tPtr.In(f.TimeZone.Get())
		// }
	}

	switch ptr := dest.Addr().Interface().(type) {
	case *Date:
		if tPtr == nil {
			*ptr = ""
		} else {
			*ptr = OfTime(*tPtr)
		}
		return nil

	case *NullableDate:
		if tPtr == nil {
			*ptr = Null
		} else {
			*ptr = OfTime(*tPtr).NullableDate()
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

	return errors.Errorf("AssignString destination type not supported: %s", dest.Type())
}

func (f *Format) FormatString(val reflect.Value) (string, error) {
	v := reflection.DerefValue(val)
	if reflection.IsNil(v) {
		return f.NilString, nil
	}

	type dateOrTime interface {
		// Format as implemented by time.Time and Date
		Format(layout string) string

		// IsZero as implemented by time.Time and Date
		IsZero() bool
	}

	switch x := val.Interface().(type) {
	case dateOrTime:
		if x.IsZero() {
			return f.ZeroString, nil
		}
		return x.Format(f.Layout), nil
	}

	return "", errors.Errorf("could not format as date string: %s", val.Type())
}
