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

// Cached type descriptors used by ReflectIsNull to avoid boxing
// via v.Interface() when the value's type does not implement
// Nullable or Zeroable.
var (
	nullableType = reflect.TypeFor[Nullable]()
	zeroableType = reflect.TypeFor[Zeroable]()
)

// ReflectIsNull returns true if v represents a "null" value, defined as:
//   - the zero reflect.Value (Kind Invalid),
//   - a nil pointer, interface, map, slice, channel, function, or unsafe.Pointer
//     anywhere along the chain of pointer/interface unwrapping,
//   - a value whose type (or addressable pointer-to-type) implements Nullable
//     and whose IsNull method returns true,
//   - a value whose type (or addressable pointer-to-type) implements Zeroable
//     and whose IsZero method returns true (only when Nullable is not implemented).
//
// Pointers and interfaces are unwrapped iteratively, so a non-nil **T pointing
// at a nil *T, or a non-nil interface holding a typed-nil pointer, are both
// reported as null without invoking IsNull on the nil receiver.
//
// ReflectIsNull never panics. Values that cannot be passed to reflect.Value.Interface
// (for example, unexported struct fields) are reported as not-null instead.
func ReflectIsNull(v reflect.Value) bool {
	for {
		switch v.Kind() {
		case reflect.Invalid:
			return true
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return true
			}
			v = v.Elem()
			continue
		}
		break
	}

	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		if v.IsNil() {
			return true
		}
	}

	if !v.CanInterface() {
		return false
	}

	t := v.Type()
	if t.Implements(nullableType) {
		return v.Interface().(Nullable).IsNull()
	}
	if v.CanAddr() && reflect.PointerTo(t).Implements(nullableType) {
		addr := v.Addr()
		if addr.CanInterface() {
			return addr.Interface().(Nullable).IsNull()
		}
	}

	if t.Implements(zeroableType) {
		return v.Interface().(Zeroable).IsZero()
	}
	if v.CanAddr() && reflect.PointerTo(t).Implements(zeroableType) {
		addr := v.Addr()
		if addr.CanInterface() {
			return addr.Interface().(Zeroable).IsZero()
		}
	}

	return false
}

// IsNull is the any-typed convenience wrapper around [ReflectIsNull].
// See ReflectIsNull for the full unwrapping and dispatch rules.
//
// IsNull never panics.
func IsNull(v any) bool {
	return ReflectIsNull(reflect.ValueOf(v))
}
