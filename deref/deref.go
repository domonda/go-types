package deref

import "time"

func Bool(ptr *bool, defaultVal bool) bool {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func String(ptr *string, defaultVal string) string {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func Int(ptr *int, defaultVal int) int {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func Uint(ptr *uint, defaultVal uint) uint {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func Uint64(ptr *uint64, defaultVal uint64) uint64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func Int32(ptr *int32, defaultVal int32) int32 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func Int64(ptr *int64, defaultVal int64) int64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func Float32(ptr *float32, defaultVal float32) float32 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func Float64(ptr *float64, defaultVal float64) float64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func Time(ptr *time.Time, defaultVal time.Time) time.Time {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}
