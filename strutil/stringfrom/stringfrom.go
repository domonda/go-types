// Package stringfrom provides utilities for converting various Go types to strings
// with safe handling of nil pointers and customizable default values.
//
// The package includes:
// - Safe pointer dereferencing with string conversion
// - Type-specific conversion functions for common Go types
// - Customizable default values for nil pointers
// - Boolean to string conversion with custom true/false strings
// - Interface reflection-based string conversion
//
// All functions handle nil pointers gracefully, the *Ptr functions by returning
// customizable default strings and Interface by returning an empty string,
// making them safe for use in scenarios where pointers might be nil.
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
// A value of string kind is returned as is,
// other values are formatted with fmt.Sprint.
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
	case reflect.String:
		return v.String()
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		if v.IsNil() {
			return ""
		}
	}
	return fmt.Sprint(v.Interface())
}
