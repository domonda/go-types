package types

import "github.com/domonda/go-wraperr/sentinel"

// Validator can be implemented by types that can validate their data.
type Validator interface {
	// Valid returns if the data of the implementation is valid.
	Valid() bool
}

// ValidatorFunc implements the Validator interface with a function.
type ValidatorFunc func() bool

// Valid returns if the data of the implementation is valid.
func (f ValidatorFunc) Valid() bool {
	return f()
}

// StaticValidator implements the Validator interface a bool validity value.
type StaticValidator bool

// Valid returns if the data of the implementation is valid.
func (valid StaticValidator) Valid() bool {
	return bool(valid)
}

type Validators []Validator

// Valid returns if the data of the implementation is valid.
func (v Validators) Valid() bool {
	for _, validator := range v {
		if !validator.Valid() {
			return false
		}
	}
	return true
}

// CombinedValidator creates a Validator whose Valid method
// returns false if any of the passed validators Valid methods
// returned false, else it returns true.
func CombinedValidator(validators ...Validator) Validator {
	return Validators(validators)
}

// ValidatErr can be implemented by types that can validate their data.
type ValidatErr interface {
	// Validate returns an error if the data of the implementation is not valid.
	Validate() error
}

// ValidatErrFunc implements the ValidatErr interface with a function.
type ValidatErrFunc func() error

// Validate returns an error if the data of the implementation is not valid.
func (f ValidatErrFunc) Validate() error {
	return f()
}

// StaticValidatErr implements the ValidatErr interface for a validation error value.
type StaticValidatErr struct {
	Err error
}

// Validate returns an error if the data of the implementation is not valid.
func (v StaticValidatErr) Validate() error {
	return v.Err
}

type ValidatErrs []ValidatErr

// Validate returns an error if the data of the implementation is not valid.
func (v ValidatErrs) Validate() error {
	for _, validatErr := range v {
		if err := validatErr.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// CombinedValidatErr creates a ValidatErr whose Validate method
// returns the first error from the passed validatErrs Validate methods
// or nil if none returned an error.
func CombinedValidatErr(validatErrs ...ValidatErr) ValidatErr {
	return ValidatErrs(validatErrs)
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
const ErrInvalidValue = sentinel.Error("invalid value")

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
