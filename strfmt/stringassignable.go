package strfmt

type StringAssignable interface {
	// AssignString tries to parse and assign the passed string
	// as value of the implementing object, or return an error
	// if the string could not be parsed as valid value.
	AssignString(string) error
}
