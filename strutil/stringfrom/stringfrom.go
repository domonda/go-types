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
// All functions handle nil pointers gracefully by returning customizable default strings,
// making them safe for use in scenarios where pointers might be nil.
package stringfrom

import (
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

// Interface converts any interface{} value to a string using reflection.
// Returns empty string for nil or invalid values.
func Interface(i any) string {
	v := reflect.ValueOf(i)

	if v.IsNil() || !v.IsValid() {
		return ""
	}

	return v.Elem().String()
}
