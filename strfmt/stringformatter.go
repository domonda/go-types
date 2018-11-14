package strfmt

import "reflect"

type StringFormatter interface {
	// json.Marshaler
	// json.Unmarshaler

	Assigntring(dest reflect.Value, str string) error
	FormatString(val reflect.Value) (string, error)
}
