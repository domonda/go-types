package stringfrom

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/domonda/go-types/uu"
)

func Ptr(ptr *string, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return *ptr
}

func TimePtr(ptr *time.Time, strForNil ...string) string {
	if ptr == nil || ptr.IsZero() {
		return strings.Join(strForNil, "")
	}
	return ptr.String()
}

func UUIDPtr(ptr *uu.ID, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return ptr.String()
}

func IntPtr(ptr *int, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return strconv.Itoa(*ptr)
}

func UintPtr(ptr *uint, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return strconv.FormatUint(uint64(*ptr), 10)
}

func Int64Ptr(ptr *int64, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return strconv.FormatInt(*ptr, 10)
}

func Float64Ptr(ptr *float64, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	return strconv.FormatFloat(*ptr, 'f', -1, 64)
}

func BoolPtr(ptr *bool, strForNil ...string) string {
	if ptr == nil {
		return strings.Join(strForNil, "")
	}
	if *ptr {
		return "true"
	}
	return "false"
}

func Bool(boolVal bool, trueString, falseString string) string {
	if boolVal {
		return trueString
	}
	return falseString
}

func Interface(i any) string {
	v := reflect.ValueOf(i)

	if v.IsNil() || !v.IsValid() {
		return ""
	}

	return v.Elem().String()
}
