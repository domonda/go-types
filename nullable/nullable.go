package nullable

import "reflect"

// Nullable is an interface with a IsNull method
type Nullable interface {
	// IsNull returns true if the implementing value is considered null.
	IsNull() bool
}

// Zeroable is an interface with a IsZero method
type Zeroable interface {
	// IsZero returns true if the implementing value is considered zero.
	IsZero() bool
}

// ReflectIsNull returns if a reflect.Value contains either a nil value
// or implements the Nullable interface and returns true from IsNull
// or implements the Zeroable interface and returns true from IsZero.
// It's safe to call ReflectIsNull on any reflect.Value
// with true returned for the zero value of reflect.Value.
func ReflectIsNull(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		v = v.Elem()

	case reflect.Map, reflect.Slice, reflect.Interface, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		if v.IsNil() {
			return true
		}

	case reflect.Invalid:
		return true
	}

	nullable, _ := v.Interface().(Nullable)
	if nullable == nil && v.CanAddr() {
		nullable, _ = v.Addr().Interface().(Nullable)
	}
	if nullable != nil {
		return nullable.IsNull()
	}

	zeroable, _ := v.Interface().(Zeroable)
	if zeroable == nil && v.CanAddr() {
		zeroable, _ = v.Addr().Interface().(Zeroable)
	}
	if zeroable != nil {
		return zeroable.IsZero()
	}

	return false
}
