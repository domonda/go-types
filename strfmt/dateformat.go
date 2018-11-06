package strfmt

import (
	"reflect"

	"github.com/domonda/errors"
	reflection "github.com/ungerik/go-reflection"
)

type DateFormat struct {
	Layout     string `json:"layout"`
	NilString  string `json:"nilString"`
	ZeroString string `json:"zeroString"`
}

func (f *DateFormat) ReflectAssignString(val reflect.Value, str string) error {
	return nil
}

func (f *DateFormat) FormatString(val interface{}) (string, error) {
	v := reflection.DerefValue(val)
	if reflection.IsNil(v) {
		return f.NilString, nil
	}
	val = v.Interface()

	switch x := val.(type) {
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
