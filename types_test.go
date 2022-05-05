package types

import (
	"reflect"
	"testing"
)

func TestReflectTypeOf(t *testing.T) {
	{
		want := reflect.TypeOf((*error)(nil)).Elem()
		got := ReflectTypeOf[error]()
		if got != want {
			t.Errorf("ReflectTypeOf() = %v, want %v", got, want)
		}
	}
	{
		// Works also with non interface types
		want := reflect.TypeOf((*int)(nil)).Elem()
		got := ReflectTypeOf[int]()
		if got != want {
			t.Errorf("ReflectTypeOf() = %v, want %v", got, want)
		}
	}
	{
		// Works also with non interface types
		want := reflect.TypeOf((**string)(nil)).Elem()
		got := ReflectTypeOf[*string]()
		if got != want {
			t.Errorf("ReflectTypeOf() = %v, want %v", got, want)
		}
	}
}
