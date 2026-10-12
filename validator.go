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
// A nil element panics unless an earlier element returned false.
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
//
// Any non-nil error means the data is invalid,
// so the error should describe what is invalid.
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
// It implements the ValidatErr interface itself,
// returning the errors of all its elements joined.
type ValidatErrs []ValidatErr

// Validate calls Validate on every element of v,
// also after one of them returned an error,
// and returns all their errors joined with errors.Join,
// or nil if no element returned an error.
// errors.Is and errors.As find each of the joined errors.
// A nil element panics.
func (v ValidatErrs) Validate() (err error) {
	for _, validatErr := range v {
		err = errors.Join(err, validatErr.Validate())
	}
	return err
}

// CombinedValidatErr creates a ValidatErr whose Validate method
// returns the errors of all the passed validatErrs Validate methods
// joined with errors.Join, or nil if none returned an error.
// This is equivalent to creating a ValidatErrs slice.
func CombinedValidatErr(validatErrs ...ValidatErr) ValidatErr {
	return ValidatErrs(validatErrs)
}

// ValidatorAsValidatErr wraps a Validator as a ValidatErr,
// returning an error naming the type and value of the Validator
// when Validator.Valid() returns false.
// This allows you to use a simple boolean validator in contexts that require detailed error information.
type ValidatorAsValidatErr struct {
	Validator
}

// Validate returns an error if the wrapped validator is not valid.
func (v ValidatorAsValidatErr) Validate() error {
	if v.Valid() {
		return nil
	}
	return fmt.Errorf("invalid %T value: %#v", v.Validator, v.Validator)
}

// Validate returns an error if v implements ValidatErr or Validator
// and the methods ValidatErr.Validate() or Validator.Valid()
// indicate an invalid value.
// If v implements both interfaces, only ValidatErr.Validate() is called.
// The error from ValidatErr.Validate() is returned unchanged.
// If Validator.Valid() is false, which can't tell why v is invalid,
// an error naming the type and value of v is returned,
// like `invalid main.Status value: "unknown"`.
// Any non-nil error means v is invalid.
// If v does not implement ValidatErr or Validator then nil will be returned.
// If v is a nil pointer to a type implementing the method with a value
// receiver, which can't be called without a value, then nil is returned,
// like [DeepValidate] skips nil pointers. A method with a pointer receiver
// is called with the nil pointer, so it can handle nil itself.
func Validate(v any) error {
	switch x := v.(type) {
	case ValidatErr:
		if isNilPointerWithValueMethod(v, validatErrType) {
			return nil
		}
		return x.Validate()
	case Validator:
		if isNilPointerWithValueMethod(v, validatorType) {
			return nil
		}
		if !x.Valid() {
			return fmt.Errorf("invalid %T value: %#v", v, v)
		}
	}
	return nil
}

// TryValidate returns an error and true if v implements ValidatErr or Validator
// and the methods ValidatErr.Validate() or Validator.Valid()
// indicate an invalid value.
// If v implements both interfaces, only ValidatErr.Validate() is called.
// The error from ValidatErr.Validate() is returned unchanged.
// If Validator.Valid() is false, which can't tell why v is invalid,
// an error naming the type and value of v is returned,
// like `invalid main.Status value: "unknown"`.
// Any non-nil error means v is invalid.
// If v does not implement ValidatErr or Validator then nil and false
// will be returned.
// If v is a nil pointer to a type implementing the method with a value
// receiver, which can't be called without a value, then nil and true are
// returned, like [DeepValidate] skips nil pointers. A method with a pointer
// receiver is called with the nil pointer, so it can handle nil itself.
// The boolean return value indicates whether the value was validatable.
func TryValidate(v any) (err error, isValidatable bool) {
	switch x := v.(type) {
	case ValidatErr:
		if isNilPointerWithValueMethod(v, validatErrType) {
			return nil, true
		}
		return x.Validate(), true
	case Validator:
		if isNilPointerWithValueMethod(v, validatorType) {
			return nil, true
		}
		if !x.Valid() {
			return fmt.Errorf("invalid %T value: %#v", v, v), true
		}
		return nil, true
	default:
		return nil, false
	}
}

// isNilPointerWithValueMethod reports if v is a nil pointer that has the
// method of the interface type iface only because its element type
// implements iface with a value receiver. Calling that method would
// dereference the nil pointer and panic.
func isNilPointerWithValueMethod(v any, iface reflect.Type) bool {
	val := reflect.ValueOf(v)
	return val.Kind() == reflect.Pointer && val.IsNil() && val.Type().Elem().Implements(iface)
}

// DeepValidate recursively validates v and everything reachable from it,
// returning every validation error encountered.
//
// The traversal:
//   - calls [Validate] on v itself, then on every reachable element,
//     through a pointer to the element if only the pointer type
//     implements ValidatErr or Validator, like [LenString]
//     (an element that is not addressable is copied for that),
//   - follows pointers and interfaces to any depth
//     (nil pointers, interfaces, and funcs stop the descent),
//   - recurses into exported struct fields, including the exported fields
//     of unexported embedded structs, slice and array elements, and map values
//     (map keys are not validated, but are sorted with [CompareReflectValue]
//     so error order is deterministic, apart from among multiple NaN keys),
//   - skips unexported struct fields, which cannot be inspected via reflection,
//   - does not descend again into a pointer, map, or slice that it is
//     already descending into higher up on the path, so cyclic data
//     is validated once instead of recursing until the stack overflows.
//
// Errors are prefixed with a human-readable path to the offending value,
// e.g. `struct field Users -> element [1] -> struct field Status: invalid main.Status value: "unknown"`.
// An error of v itself is returned without a prefix.
// The prefixed errors wrap the errors returned by [Validate],
// so errors.Is and errors.As see through the path.
//
// Use errors.Join(DeepValidate(v)...) to join the errors into a single error.
func DeepValidate(v any) []error {
	var errs []error
	visiting := make(map[visitKey]struct{})
	deepValidate(reflect.ValueOf(v), visiting, func(err error) {
		errs = append(errs, err)
	})
	return errs
}

// visitKey identifies a pointer, map, or slice that deepValidate
// is descending into. The type is part of the key because a pointer
// to a struct and a pointer to its first field share an address,
// and the length because slices of different length can share
// the same underlying array.
type visitKey struct {
	ptr uintptr
	typ reflect.Type
	len int
}

// deepValidate is the internal implementation of DeepValidate.
// It recursively validates nested structures and provides path information for errors.
// visiting holds the pointers, maps, and slices on the current path to detect cycles.
func deepValidate(v reflect.Value, visiting map[visitKey]struct{}, onError func(error), path ...string) {
	if !v.IsValid() {
		return
	}

	// Leave the pointers, maps, and slices entered by this call
	// so that data shared by multiple paths is not taken for a cycle.
	var entered []visitKey
	defer func() {
		for _, key := range entered {
			delete(visiting, key)
		}
	}()
	// enter returns false if v is already being descended into,
	// which means that descending again would follow a cycle.
	enter := func(v reflect.Value) bool {
		key := visitKey{ptr: v.Pointer(), typ: v.Type()}
		if v.Kind() == reflect.Slice {
			key.len = v.Len()
		}
		if _, ok := visiting[key]; ok {
			return false
		}
		visiting[key] = struct{}{}
		entered = append(entered, key)
		return true
	}

	// Try Validate at every level of the pointer chain. The method
	// set of **T does not include T's value-receiver Validators, so calling
	// Validate only at the outer level would miss validators reachable only
	// after one or more dereferences. Stop at the first level that implements
	// ValidatErr or Validator, because the levels below call the same
	// methods again, which would duplicate errors and side effects.
	// An interface is not validated itself, but its dynamic value
	// after unwrapping it, which has the same methods.
	validated := false
	for {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Func:
			// A Valid or Validate method of a nil func type
			// would call the nil func and panic
			if v.IsNil() {
				return
			}
		}
		switch v.Kind() {
		case reflect.Pointer, reflect.Map, reflect.Slice:
			if !enter(v) {
				// Already validated and descended into higher up on the path
				return
			}
		}
		if !validated && v.Kind() != reflect.Interface && v.CanInterface() {
			err, isValidatable := TryValidate(validationTarget(v))
			if err != nil {
				if len(path) > 0 {
					err = fmt.Errorf("%s: %w", strings.Join(path, " -> "), err)
				}
				onError(err)
			}
			validated = isValidatable
		}
		// Fully unwrap so the descent below sees the concrete container kind
		if k := v.Kind(); k != reflect.Pointer && k != reflect.Interface {
			break
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := range v.NumField() {
			field := t.Field(i)
			// The exported fields of an unexported embedded struct
			// are promoted to v and can be inspected
			embeddedStruct := field.Anonymous &&
				(field.Type.Kind() == reflect.Struct ||
					field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct)
			if !field.IsExported() && !embeddedStruct {
				continue
			}
			name := fmt.Sprintf("struct field %s", field.Name)
			deepValidate(v.Field(i), visiting, onError, append(path, name)...)
		}
	case reflect.Map:
		keys := v.MapKeys()
		slices.SortFunc(keys, CompareReflectValue)
		for _, key := range keys {
			name := fmt.Sprintf("map value [%#v]", key.Interface())
			deepValidate(v.MapIndex(key), visiting, onError, append(path, name)...)
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			name := fmt.Sprintf("element [%d]", i)
			deepValidate(v.Index(i), visiting, onError, append(path, name)...)
		}
	}
}

var (
	validatErrType = reflect.TypeFor[ValidatErr]()
	validatorType  = reflect.TypeFor[Validator]()
)

// validationTarget returns the value to pass to Validate for v,
// which must be interfaceable and not an interface.
// That is v itself, unless only the pointer type of v implements
// ValidatErr, or only the pointer type implements Validator, in which
// case it is a pointer to v, or to a copy of v if v is not addressable.
// Like Validate, a ValidatErr is preferred over a Validator.
func validationTarget(v reflect.Value) any {
	val := v.Interface()
	if _, ok := val.(ValidatErr); ok {
		return val
	}
	_, isValidator := val.(Validator)
	// The pointer type of a pointer has no methods
	ptrType := reflect.PointerTo(v.Type())
	if !ptrType.Implements(validatErrType) && (isValidator || !ptrType.Implements(validatorType)) {
		return val
	}
	if v.CanAddr() {
		return v.Addr().Interface()
	}
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)
	return ptr.Interface()
}
