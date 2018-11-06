package strutil

import (
	"math/rand"
	"time"

	"github.com/guregu/null"
)

// NullStringFromPtr converts a string pointer to null.String
// func NullStringFromPtr(stringPtr *string) (ns null.String) {
// 	if stringPtr == nil {
// 		return ns
// 	}
// 	ns.String = *stringPtr
// 	ns.Valid = true
// 	return ns
// }

// // NullStringValid converts a string to valid null.String
// func NullStringValid(str string) (ns null.String) {
// 	return null.String{
// 		String: str,
// 		Valid:  true,
// 	}
// }

// // NullStringEmptyInvalid converts a string to valid null.String
// func NullStringEmptyInvalid(str string) (ns null.String) {
// 	return null.String{
// 		String: str,
// 		Valid:  str != "",
// 	}
// }

// StringOrNil returns the string that stringPtr points,
// or nil if stringPtr points to an empty string or is nil.
// func StringOrNil(stringPtr *string) interface{} {
// 	if stringPtr == nil || len(*stringPtr) == 0 {
// 		return nil
// 	}
// 	return *stringPtr
// }

// EmptyStringToNil returns str or nil if it is empty.
func EmptyStringToNil(str string) interface{} {
	if str == "" {
		return nil
	}
	return str
}

// StringToNullString returns correct null.String
func StringToNullString(str string) null.String {
	if str == "" {
		return null.NewString("", false)
	}
	return null.StringFrom(str)
}

// PtrFromString returns a pointer to a string
func PtrFromString(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

// IndexInStrings returns the index of where str
// can be found in slice.
func IndexInStrings(str string, slice []string) int {
	for i := range slice {
		if str == slice[i] {
			return i
		}
	}
	return -1
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringOfSize(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func Truncate(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i])
	}
	return s
}

func TruncateWithEllipsis(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i-1]) + "…"
	}
	return s
}

// DerefPtr returns the string ptr points to
// or an empty string if ptr is nil.
func DerefPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// EqualPtrOrString returns if a and b are equal pointers
// or if the pointed to strings are equal
func EqualPtrOrString(a, b *string) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
