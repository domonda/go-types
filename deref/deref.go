// Package deref provides safe pointer dereferencing utilities for Go applications.
// These functions safely dereference pointers and return default values when the pointer is nil,
// preventing panic errors that would occur with direct pointer dereferencing.
//
// The package includes:
// - Safe dereferencing for all basic Go types
// - Default value fallback for nil pointers
// - Type-specific functions for bool, string, numeric types, and time.Time
package deref

import "time"

// Bool safely dereferences a *bool pointer and returns the value or defaultVal if ptr is nil.
func Bool(ptr *bool, defaultVal bool) bool {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// String safely dereferences a *string pointer and returns the value or defaultVal if ptr is nil.
func String(ptr *string, defaultVal string) string {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Int safely dereferences a *int pointer and returns the value or defaultVal if ptr is nil.
func Int(ptr *int, defaultVal int) int {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Uint safely dereferences a *uint pointer and returns the value or defaultVal if ptr is nil.
func Uint(ptr *uint, defaultVal uint) uint {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Uint64 safely dereferences a *uint64 pointer and returns the value or defaultVal if ptr is nil.
func Uint64(ptr *uint64, defaultVal uint64) uint64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Int32 safely dereferences a *int32 pointer and returns the value or defaultVal if ptr is nil.
func Int32(ptr *int32, defaultVal int32) int32 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Int64 safely dereferences a *int64 pointer and returns the value or defaultVal if ptr is nil.
func Int64(ptr *int64, defaultVal int64) int64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Float32 safely dereferences a *float32 pointer and returns the value or defaultVal if ptr is nil.
func Float32(ptr *float32, defaultVal float32) float32 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Float64 safely dereferences a *float64 pointer and returns the value or defaultVal if ptr is nil.
func Float64(ptr *float64, defaultVal float64) float64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Time safely dereferences a *time.Time pointer and returns the value or defaultVal if ptr is nil.
func Time(ptr *time.Time, defaultVal time.Time) time.Time {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}
