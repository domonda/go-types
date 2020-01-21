package strfmt

type Scannable interface {
	// ScanString tries to parse and assign the passed
	// source string as value of the implementing type.
	// It returns an error if source could not be parsed.
	// If the source string could be parsed, but was not
	// in the expected normalized format, then false is
	// returned for sourceWasNormalized and nil for err.
	ScanString(source string) (sourceWasNormalized bool, err error)
}
