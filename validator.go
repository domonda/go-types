package types

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// Validator can be implemented by types that can validate their data.
// The Valid method should return true if the data is valid, false otherwise.
type Validator interface {
	// Valid returns if the data of the implementation is valid.
	Valid() bool
}

// ValidatorFunc implements the Validator interface with a function.
// This allows you to use a function as a validator.
type ValidatorFunc func() bool

// Valid returns if the data of the implementation is valid.
func (f ValidatorFunc) Valid() bool {
	return f()
}

// StaticValidator implements the Validator interface with a bool validity value.
// This is useful when you have a pre-computed validation result.
type StaticValidator bool

// Valid returns if the data of the implementation is valid.
func (valid StaticValidator) Valid() bool {
	return bool(valid)
}

// Validators is a slice of Validator implementations.
// It implements the Validator interface itself, returning true only if all validators are valid.
type Validators []Validator

// Valid returns if the data of the implementation is valid.
// Returns false if any validator in the slice returns false.
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
// This is equivalent to creating a Validators slice.
func CombinedValidator(validators ...Validator) Validator {
	return Validators(validators)
}

// ValidatErr can be implemented by types that can validate their data and return detailed error information.
// This is more informative than the Validator interface as it provides specific error details.
type ValidatErr interface {
	// Validate returns an error if the data of the implementation is not valid.
	Validate() error
}

// ValidatErrFunc implements the ValidatErr interface with a function.
// This allows you to use a function as a validator that returns detailed errors.
type ValidatErrFunc func() error

// Validate returns an error if the data of the implementation is not valid.
func (f ValidatErrFunc) Validate() error {
	return f()
}

// StaticValidatErr implements the ValidatErr interface for a validation error value.
// This is useful when you have a pre-computed validation error.
type StaticValidatErr struct {
	Err error
}

// Validate returns an error if the data of the implementation is not valid.
func (v StaticValidatErr) Validate() error {
	return v.Err
}

// ValidatErrs is a slice of ValidatErr implementations.
// It implements the ValidatErr interface itself, returning the first error encountered.
type ValidatErrs []ValidatErr

// Validate returns an error if the data of the implementation is not valid.
// Returns the first error from any validator in the slice, or nil if all are valid.
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
// This is equivalent to creating a ValidatErrs slice.
func CombinedValidatErr(validatErrs ...ValidatErr) ValidatErr {
	return ValidatErrs(validatErrs)
}

// ValidatorAsValidatErr wraps a Validator as a ValidatErr,
// returning ErrInvalidValue when Validator.Valid() returns false.
// This allows you to use a simple boolean validator in contexts that require detailed error information.
type ValidatorAsValidatErr struct {
	Validator
}

// Validate returns an error if the wrapped validator is not valid.
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
func Validate(v any) error {
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

// TryValidate returns an error and true if v implements ValidatErr or Validator
// and the methods ValidatErr.Validate() or Validator.Valid()
// indicate an invalid value.
// The error from ValidatErr.Validate() is returned directly,
// and ErrInvalidValue is returned if Validator.Valid() is false.
// If v does not implement ValidatErr or Validator then nil and false
// will be returned.
// The boolean return value indicates whether the value was validatable.
func TryValidate(v any) (err error, isValidatable bool) {
	switch x := v.(type) {
	case ValidatErr:
		return x.Validate(), true
	case Validator:
		if x.Valid() {
			return nil, true
		} else {
			return ErrInvalidValue, true
		}
	default:
		return nil, false
	}
}

// DeepValidate recursively validates v and everything reachable from it,
// returning every validation error encountered.
//
// The traversal:
//   - calls [Validate] on v itself, then on every reachable element,
//   - follows pointers and interfaces to any depth (nil values stop the descent),
//   - recurses into exported struct fields, slice and array elements, and map values
//     (map keys are not validated, but are sorted with [CompareReflectValue]
//     so error order is deterministic),
//   - skips unexported struct fields, which cannot be inspected via reflection.
//
// Errors are prefixed with a human-readable path to the offending value,
// e.g. "struct field Users -> element [1] -> struct field Name: invalid value".
//
// Cycles via self-referential pointers or maps are not detected and will
// recurse until the stack overflows; pass acyclic data.
//
// Use errors.Join(DeepValidate(v)...) to join the errors into a single error.
func DeepValidate(v any) []error {
	var errs []error
	deepValidate(reflect.ValueOf(v), func(err error) {
		errs = append(errs, err)
	})
	return errs
}

// deepValidate is the internal implementation of DeepValidate.
// It recursively validates nested structures and provides path information for errors.
func deepValidate(v reflect.Value, onError func(error), path ...string) {
	if !v.IsValid() {
		return
	}

	// Try Validate at every level of the pointer/interface chain. The method
	// set of **T does not include T's value-receiver Validators, so calling
	// Validate only at the outer level would miss validators reachable only
	// after one or more dereferences. Stop at the first level that produces
	// an error to avoid duplicate reports for the same logical value.
	for {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return
			}
		}
		if v.CanInterface() && !isTypedNilPointerInInterface(v) {
			if err := Validate(v.Interface()); err != nil {
				if len(path) > 0 {
					err = fmt.Errorf("%s: %w", strings.Join(path, " -> "), err)
				}
				onError(err)
				break
			}
		}
		if k := v.Kind(); k != reflect.Pointer && k != reflect.Interface {
			break
		}
		v = v.Elem()
	}

	// Fully unwrap so the descent below sees the concrete container kind.
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := range v.NumField() {
			if !t.Field(i).IsExported() {
				continue
			}
			name := fmt.Sprintf("struct field %s", t.Field(i).Name)
			deepValidate(v.Field(i), onError, append(path, name)...)
		}
	case reflect.Map:
		keys := v.MapKeys()
		slices.SortFunc(keys, CompareReflectValue)
		for _, key := range keys {
			name := fmt.Sprintf("map value [%#v]", key.Interface())
			deepValidate(v.MapIndex(key), onError, append(path, name)...)
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			name := fmt.Sprintf("element [%d]", i)
			deepValidate(v.Index(i), onError, append(path, name)...)
		}
	}
}

// isTypedNilPointerInInterface reports whether v is an interface value
// wrapping a nil pointer. Calling Validate on such a value would route
// through the wrapped type's Valid/Validate method with a nil receiver
// and panic, so the caller should skip it.
func isTypedNilPointerInInterface(v reflect.Value) bool {
	if v.Kind() != reflect.Interface {
		return false
	}
	e := v.Elem()
	return e.Kind() == reflect.Pointer && e.IsNil()
}
