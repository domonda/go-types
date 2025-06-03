package types

// Ptr returns a new pointer to a copy the passed value.
func Ptr[T any](value T) *T {
	return &value
}

// FromPtr returns the dereferenced value of the pointer if it is not nil,
// otherwise it returns the zero value of the type T.
func FromPtr[T any](ptr *T) T {
	if ptr == nil {
		return *new(T)
	}
	return *ptr
}

// FromPtrOr returns the dereferenced value of the pointer if it is not nil,
// otherwise it returns the passed defaultValue.
func FromPtrOr[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
