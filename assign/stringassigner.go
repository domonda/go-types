package assign

import (
	"reflect"
)

type StringAssigner interface {
	AssignString(dest reflect.Value, str string, parser *StringParser) error
}

type StringAssignerFunc func(dest reflect.Value, str string, parser *StringParser) error

func (f StringAssignerFunc) AssignString(dest reflect.Value, str string, parser *StringParser) error {
	return f(dest, str, parser)
}
