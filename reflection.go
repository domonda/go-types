package types

import "reflect"

// ReflectTypeOf returns the reflect.Type
// of the generic type parameter T
// wich can also be an interface type like error.
func ReflectTypeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
