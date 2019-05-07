package assign

import (
	"reflect"
)

type StringAssigner interface {
	AssignString(str string, config *StringConfig, dest reflect.Value) error
}

type StringAssignerFunc func(str string, config *StringConfig, dest reflect.Value) error

func (f StringAssignerFunc) AssignString(str string, config *StringConfig, dest reflect.Value) error {
	return f(str, config, dest)
}
