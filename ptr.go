package types

// FromPtr returns the dereferenced value of the pointer if it is not nil,
// otherwise it returns the zero value of the type T.
// This safely dereferences a pointer without panicking on nil.
func FromPtr[T any](ptr *T) T {
	if ptr == nil {
		return *new(T)
	}
	return *ptr
}

// FromPtrOr returns the dereferenced value of the pointer if it is not nil,
// otherwise it returns the passed defaultValue.
// This provides a safe way to dereference a pointer with a fallback value.
func FromPtrOr[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
