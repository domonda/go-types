// Package nullable provides interfaces and utilities for handling nullable values in Go.
// It defines common interfaces for types that can represent null/zero states and provides
// reflection-based utilities for null checking.
//
// The package includes:
// - Nullable interface for types that can represent null values
// - NullSetable interface for types that can be set to null and retrieved
// - Zeroable interface for types that can represent zero values
// - Reflection utilities for null/zero checking
package nullable

import (
	"reflect"

	"github.com/domonda/go-types"
)

// Nullable is an interface with an IsNull method for types that can represent null values.
type Nullable interface {
	// IsNull returns true if the implementing value is considered null.
	IsNull() bool
}

// NullSetable is a generic interface that extends Nullable for types that can be
// set to null and have their values retrieved with fallback options.
// This interface is typically implemented by nullable wrapper types that provide
// safe handling of optional values with null state management.
type NullSetable[T any] interface {
	Nullable

	// SetNull sets the value to null/empty state.
	SetNull()

	// Set assigns a non-null value of type T.
	Set(T)

	// Get returns the non-null value or panics if the value is null.
	// Use IsNull() to check before calling Get() to avoid panics.
	Get() T

	// GetOr returns the non-null value or the provided default value if null.
	// This is a safe alternative to Get() that never panics.
	GetOr(T) T
}

// Zeroable is an interface with an IsZero method for types that can represent zero values.
type Zeroable interface {
	// IsZero returns true if the implementing value is considered zero.
	IsZero() bool
}

// NullableValidator is a generic interface that extends Nullable and Validator for types that can be
// set to null and have their values retrieved with fallback options.
// This interface is typically implemented by nullable wrapper types that provide
// safe handling of optional values with null state management.
type NullableValidator interface {
	Nullable
	types.Validator

	ValidAndNotNull() bool
}

// ReflectIsNull returns if a reflect.Value contains either a nil value
// or implements the Nullable interface and returns true from IsNull
// or implements the Zeroable interface and returns true from IsZero.
// It's safe to call ReflectIsNull on any reflect.Value
// with true returned for the zero value of reflect.Value.
func ReflectIsNull(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer:
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
