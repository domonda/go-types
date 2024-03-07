package types

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

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

// DeepValidate validates all fields of a struct, all elements of a slice or array,
// and all values of a map by recursively calling Validate or Valid methods.
func DeepValidate(v any) error {
	return deepValidate(reflect.ValueOf(v))
}

func deepValidate(v reflect.Value, path ...string) error {
	err := Validate(v.Interface())
	if err != nil && len(path) > 0 {
		err = fmt.Errorf("%s: %w", strings.Join(path, " -> "), err)
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return err
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			name := fmt.Sprintf("struct field %s", v.Type().Field(i).Name)
			err = errors.Join(err, deepValidate(v.Field(i), append(path, name)...))
		}
	case reflect.Map:
		keys := v.MapKeys()
		slices.SortFunc(keys, ReflectCompare)
		for _, key := range keys {
			name := fmt.Sprintf("map value [%#v]", key.Interface())
			err = errors.Join(err, deepValidate(v.MapIndex(key), append(path, name)...))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			name := fmt.Sprintf("elememt [%d]", i)
			err = errors.Join(err, deepValidate(v.Index(i), append(path, name)...))
		}
	}
	return err
}

// ReflectCompare compares two reflect.Values of the same type.
// The function panics if the types of a and b
// are not idential or not orderable.
// Orderable types are variantes of integers, floats, and strings.
func ReflectCompare(a, b reflect.Value) int {
	if a.Type() != b.Type() {
		panic("values are not of the same type")
	}
	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmp.Compare(a.Int(), b.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return cmp.Compare(a.Uint(), b.Uint())
	case reflect.Float32, reflect.Float64:
		return cmp.Compare(a.Float(), b.Float())
	case reflect.String:
		return cmp.Compare(a.String(), b.String())
	default:
		panic("values are not of an orderable type")
	}
}
