package strfmt

import "reflect"

// split in two?
type StringFormatter interface {
	Assigntring(dest reflect.Value, str string) error
	FormatString(val reflect.Value) (string, error)
}
