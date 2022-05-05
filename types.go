package types

import "reflect"

// ReflectTypeOf returns the reflect.Type
// of the generic type parameter T
// wich can also be an interface type like error.
func ReflectTypeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// Zero value of the generic type T
func Zero[T any]() T {
	var zero T
	return zero
}
