package deref

import "time"

func Bool(ptr *bool, nilVal bool) bool {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

func String(ptr *string, nilVal string) string {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

func Int(ptr *int, nilVal int) int {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

func Uint(ptr *uint, nilVal uint) uint {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

func Uint64(ptr *uint64, nilVal uint64) uint64 {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

func Int32(ptr *int32, nilVal int32) int32 {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

func Int64(ptr *int64, nilVal int64) int64 {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

func Float32(ptr *float32, nilVal float32) float32 {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

func Float64(ptr *float64, nilVal float64) float64 {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

func Time(ptr *time.Time, nilVal time.Time) time.Time {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}
