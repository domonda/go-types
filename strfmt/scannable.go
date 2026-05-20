package strfmt

// Scannable is implemented by types that can parse and assign a string
// value to themselves, optionally enforcing validity constraints.
type Scannable interface {
	// ScanString tries to parse and assign the passed
	// source string as value of the implementing type.
	//
	// If validate is true, the source string is checked
	// for validity before it is assigned to the type.
	//
	// If validate is false and the source string
	// can still be assigned in some non-normalized way
	// it will be assigned without returning an error.
	ScanString(source string, validate bool) error
}
