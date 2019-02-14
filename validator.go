package types

import "errors"

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

// ValidatorAsValidatErr wraps a Validator as a ValidatErr,
// returning ErrInvalidValue when Validator.Valid() returns false.
type ValidatorAsValidatErr struct {
	Validator
}

func (v ValidatorAsValidatErr) Validate() error {
	if v.Valid() {
		return nil
	}
	return ErrInvalidValue
}

// ErrInvalidValue means that a value is not valid,
// returned by Validate() and ValidatorAsValidatErr.Validate().
var ErrInvalidValue = errors.New("invalid value")

// Validate returns an error if v implements ValidatErr or Validator
// and the methods ValidatErr.Validate() or Validator.Valid()
// indicate an invalid value.
// The error from ValidatErr.Validate() is returned directly,
// and ErrInvalidValue is returned if Validator.Valid() is false.
// If v does not implement ValidatErr or Validator then nil will be returned.
func Validate(v interface{}) error {
	switch x := v.(type) {
	case ValidatErr:
		return x.Validate()
	case Validator:
		if !x.Valid() {
			return ErrInvalidValue
		}
	}
	return nil
}
