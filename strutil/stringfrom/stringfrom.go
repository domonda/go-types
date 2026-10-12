// Package stringfrom provides utilities for converting various Go types to strings
// with safe handling of nil pointers and customizable default values.
//
// The package includes:
//   - [Ptr], [TimePtr], [UUIDPtr], [IntPtr], [UintPtr], [Int64Ptr],
//     [Float64Ptr] and [BoolPtr], which dereference a pointer and format
//     the value. For a nil pointer they return their strForNil arguments
//     joined without a separator, or "" if none are passed.
//     [TimePtr] treats a pointer to the zero time like nil.
//   - [Bool], which returns one of two passed strings.
//   - [Interface], which converts any value, or the value a pointer points to:
//     a value with a Format, Error or String method, also one with a pointer
//     receiver, with that method, a value of string kind as is,
//     other values with fmt.Sprint.
//     It returns "" for nil, a nil pointer,
//     or a nil map, slice, func, channel or interface value.
//
// None of the functions panics for nil input.
package stringfrom

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/domonda/go-types/uu"
)

// Ptr safely dereferences a *string pointer and returns the value or a custom string for nil.
// If ptr is nil, returns the concatenation of strForNil arguments.
func Ptr(ptr *string, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return *ptr
}

// TimePtr safely dereferences a *time.Time pointer and returns its string representation
// or a custom string for nil/zero values.
func TimePtr(ptr *time.Time, strForNil ...string) string {
	if ptr == nil || ptr.IsZero() {
		return strings.Join(strForNil, "")
	}
	return ptr.String()
}

// UUIDPtr safely dereferences a *uu.ID pointer and returns its string representation
// or a custom string for nil.
func UUIDPtr(ptr *uu.ID, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return ptr.String()
}

// IntPtr safely dereferences a *int pointer and returns its string representation
// or a custom string for nil.
func IntPtr(ptr *int, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return strconv.Itoa(*ptr)
}

// UintPtr safely dereferences a *uint pointer and returns its string representation
// or a custom string for nil.
func UintPtr(ptr *uint, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return strconv.FormatUint(uint64(*ptr), 10)
}

// Int64Ptr safely dereferences a *int64 pointer and returns its string representation
// or a custom string for nil.
func Int64Ptr(ptr *int64, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return strconv.FormatInt(*ptr, 10)
}

// Float64Ptr safely dereferences a *float64 pointer and returns its string representation
// or a custom string for nil.
func Float64Ptr(ptr *float64, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return strconv.FormatFloat(*ptr, 'f', -1, 64)
}

// BoolPtr safely dereferences a *bool pointer and returns "true"/"false" string
// or a custom string for nil.
func BoolPtr(ptr *bool, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	if *ptr {
		return "true"
	}
	return "false"
}

// Bool converts a boolean value to a string using custom true/false strings.
func Bool(boolVal bool, trueString, falseString string) string {
	if boolVal {
		return trueString
	}
	return falseString
}

// Interface converts i, or the value i points to, to a string,
// so a non-pointer value returns the same string as a pointer to it.
// If the value or a pointer to it has a Format, Error or String method,
// the value is formatted with fmt.Sprint through a pointer to it,
// so methods with a pointer receiver are used too.
// Otherwise a value of string kind is returned as is,
// and other values are formatted with fmt.Sprint.
// Returns an empty string for nil, a nil pointer,
// or a nil map, slice, func, channel or interface value.
func Interface(i any) string {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Pointer {
		v = v.Elem() // zero Value for a nil pointer
	}
	switch v.Kind() {
	case reflect.Invalid:
		return ""
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		if v.IsNil() {
			return ""
		}
	}
	// The method set of a pointer includes the methods with a value receiver
	if ptrType := reflect.PointerTo(v.Type()); ptrType.Implements(formatterType) || ptrType.Implements(errorType) || ptrType.Implements(stringerType) {
		if !v.CanAddr() {
			// v is not a pointer target, format a pointer to a copy
			ptr := reflect.New(v.Type())
			ptr.Elem().Set(v)
			v = ptr.Elem()
		}
		return fmt.Sprint(v.Addr().Interface())
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	return fmt.Sprint(v.Interface())
}

var (
	formatterType = reflect.TypeFor[fmt.Formatter]()
	errorType     = reflect.TypeFor[error]()
	stringerType  = reflect.TypeFor[fmt.Stringer]()
)
