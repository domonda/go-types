package strutil

import "strings"

// Ptr returns the passed string as pointer value.
func Ptr(str string) *string { return &str }

// EmptyStringToNil returns str or nil if it is empty.
func EmptyStringToNil(str string) any {
	if str == "" {
		return nil
	}
	return str
}

// StringToPtrEmptyToNil returns the address of a string
// or nil if the string is empty.
func StringToPtrEmptyToNil(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

// IndexInStrings returns the index of where str
// can be found in slice, or -1 if it was not found.
func IndexInStrings(str string, slice []string) int {
	for i := range slice {
		if str == slice[i] {
			return i
		}
	}
	return -1
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

// Join concatenates the elements of its first argument to create a single string. The separator
// string sep is placed between elements in the resulting string.
func Join[T ~string](elems []T, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return string(elems[0])
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(string(elems[0]))
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(string(s))
	}
	return b.String()
}
