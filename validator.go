package types

// Validator can be implemented by types that can validate their data.
type Validator interface {
	// Valid returns if the data of the implementation is valid.
	Valid() bool
}

// ValidatErr can be implemented by types that can validate their data.
type ValidatErr interface {
	// Validate returns an error if the data of the implementation is not valid.
	Validate() error
}
