package strfmt

import "reflect"

type StringFormatter interface {
	// json.Marshaler
	// json.Unmarshaler

	ReflectAssignString(val reflect.Value, str string) error
	FormatString(val interface{}) (string, error)
}
