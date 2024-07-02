package date

import (
	"fmt"
	"reflect"
	"time"

	reflection "github.com/ungerik/go-reflection"

	"github.com/domonda/go-types/language"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
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

func (f *Format) AssignString(dest reflect.Value, source string /*, loc *time.Location*/) error {
	source = strutil.TrimSpace(source)

	tPtr := new(time.Time)
	if source != "" {
		if f.Layout == "" {
			d, err := Normalize(source)
			if err != nil {
				return err
			}
			t := d.MidnightInLocation(time.Local)
			tPtr = &t
		} else {
			t, err := time.Parse(f.Layout, source)
			if err != nil {
				return err
			}
			tPtr = &t
		}
		if tPtr.IsZero() {
			return fmt.Errorf("can't assign zero time")
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
			*ptr = OfTime(*tPtr).Nullable()
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

	case *nullable.Time:
		*ptr = nullable.TimeFromPtr(tPtr)
		return nil
	}

	return fmt.Errorf("AssignString destination type not supported: %s", dest.Type())
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

	return "", fmt.Errorf("could not format as date string: %s", val.Type())
}
