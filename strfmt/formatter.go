package strfmt

import "reflect"

type Formatter interface {
	FormatString(val reflect.Value) (string, error)
}
