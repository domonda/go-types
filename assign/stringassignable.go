package assign

type StringAssignable interface {
	// AssignString tries to parse and assign the passed
	// source string as value of the implementing object.
	// It returns an error if source could not be parsed.
	// If the source string could be parsed, but was not
	// in the expected normalized format, then false is
	// returned for normalized and nil for err.
	AssignString(source string) (normalized bool, err error)
}
