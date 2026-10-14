package types

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
)

// Validator can be implemented by types that can validate their data.
// The Valid method should return true if the data is valid, false otherwise.
//
// Because Valid can't tell why the data is invalid, Validate, TryValidate,
// DeepValidate and ValidatorAsValidatErr return an error naming the type
// and the value formatted with %#v, like `invalid main.Status value: "unknown"`.
// So the whole value is part of the error message. Implement ValidatErr
// instead for other errors, like for a value that must not be logged,
// such as a password, or that can be large, because the error
// of ValidatErr.Validate is returned unchanged.
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
// Elements that [Validate] does not validate are skipped:
// nil, a nil func, and a nil pointer to a type implementing
// Validator with a value receiver.
func (v Validators) Valid() bool {
	for _, validator := range v {
		if !validOrSkipped(validator) {
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
// Elements that [Validate] does not validate are skipped:
// nil, a nil func, and a nil pointer to a type implementing
// ValidatErr with a value receiver.
//
// Because an element is called also after an earlier one failed,
// no element may rely on an earlier one, like a check of a field
// that relies on an earlier check that the pointer to the struct
// is not nil. Make such dependent checks one ValidatErr instead.
func (v ValidatErrs) Validate() (err error) {
	for _, validatErr := range v {
		err = errors.Join(err, Validate(validatErr))
	}
	return err
}

// CombinedValidatErr creates a ValidatErr whose Validate method
// returns the errors of all the passed validatErrs Validate methods
// joined with errors.Join, or nil if none returned an error.
// Every validatErr is called, also after an earlier one failed,
// see [ValidatErrs.Validate].
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
// A Validator that [Validate] does not validate returns nil:
// nil, a nil func, and a nil pointer to a type implementing
// Validator with a value receiver.
func (v ValidatorAsValidatErr) Validate() error {
	if validOrSkipped(v.Validator) {
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
// If v is a nil func like ValidatorFunc(nil), whose method would call
// the nil func, then nil is returned, like [DeepValidate] skips nil funcs.
// If v embeds a nil pointer that its method is promoted from, like
// struct{ *T } with a nil *T and a value receiver on T, then nil is
// returned too, because the method can't be called without the value.
func Validate(v any) error {
	err, _ := TryValidate(v)
	return err
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
// If v is a nil func like ValidatorFunc(nil), whose method would call
// the nil func, then nil and true are returned, like [DeepValidate]
// skips nil funcs. The same is returned if v embeds a nil pointer
// that its method is promoted from, see [Validate].
// The boolean return value indicates whether the value was validatable.
func TryValidate(v any) (err error, isValidatable bool) {
	switch x := v.(type) {
	case ValidatErr:
		if isNilFunc(v) || isNilPointerWithValueMethod(v, validatErrType) {
			return nil, true
		}
		if callSkippingNilEmbedded(v, func() { err = x.Validate() }) {
			return nil, true
		}
		return err, true
	case Validator:
		if !validOrSkipped(x) {
			return fmt.Errorf("invalid %T value: %#v", v, v), true
		}
		return nil, true
	default:
		return nil, false
	}
}

// validOrSkipped returns the result of the Valid method of v,
// or true if [Validate] does not validate v: if v is nil, a nil func,
// a nil pointer to a type implementing Validator with a value receiver,
// or a value whose Valid method is promoted from a nil embedded pointer.
func validOrSkipped(v Validator) bool {
	if v == nil || isNilFunc(v) || isNilPointerWithValueMethod(v, validatorType) {
		return true
	}
	valid := true
	callSkippingNilEmbedded(v, func() { valid = v.Valid() })
	return valid
}

// isNilPointerWithValueMethod reports if v is a nil pointer that has the
// method of the interface type iface only because its element type
// implements iface with a value receiver. Calling that method would
// dereference the nil pointer and panic.
func isNilPointerWithValueMethod(v any, iface reflect.Type) bool {
	val := reflect.ValueOf(v)
	return val.Kind() == reflect.Pointer && val.IsNil() && val.Type().Elem().Implements(iface)
}

// callSkippingNilEmbedded calls call, which must call the Valid or Validate
// method of v directly, and reports if it panicked because the method
// is promoted from a nil pointer that v embeds, like Validate of
// struct{ *T } with a nil *T and a value receiver on T. The method wrapper
// that the compiler generates for the promotion dereferences the nil
// pointer before T's method is called. Such a value is not validated,
// like a nil pointer, because the method can't be called without the
// embedded value. A struct can declare the method itself instead, which
// reflection can't tell apart, so the call is made and only a nil pointer
// dereference in the generated wrapper called by call is recovered.
// Every other panic is passed on, like a nil pointer dereference
// in the method itself, which is a bug of the method.
func callSkippingNilEmbedded(v any, call func()) (skipped bool) {
	if !hasNilEmbeddedPointer(reflect.ValueOf(v), 0) {
		call()
		return false
	}
	defer func() {
		if r := recover(); r != nil {
			if rErr, ok := r.(runtime.Error); ok && strings.Contains(rErr.Error(), "nil pointer dereference") && panickedInWrapperCalledBy(call) {
				skipped = true
				return
			}
			panic(r)
		}
	}()
	call()
	return false
}

// panickedInWrapperCalledBy reports if the panic being recovered by the
// deferred function that calls it is a memory fault in a method wrapper
// generated by the compiler, called directly by call. The runtime reports
// the file of generated functions as "<autogenerated>", and a method
// promoted through several embedded fields can have a chain of wrappers.
func panickedInWrapperCalledBy(call func()) bool {
	callName := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	pcs := make([]uintptr, 64)
	frames := runtime.CallersFrames(pcs[:runtime.Callers(1, pcs)])
	// The frame of the faulting function follows runtime.sigpanic
	for {
		frame, more := frames.Next()
		if frame.Function == "runtime.sigpanic" {
			break
		}
		if !more {
			return false
		}
	}
	wrappers := 0
	for {
		frame, more := frames.Next()
		if frame.File != "<autogenerated>" {
			return wrappers > 0 && frame.Function == callName
		}
		wrappers++
		if !more {
			return false
		}
	}
}

// hasNilEmbeddedPointer reports if the struct v, or the struct v points to,
// embeds a nil pointer, also within its embedded fields.
// The depth is limited because embedded pointers can form cycles.
func hasNilEmbeddedPointer(v reflect.Value, depth int) bool {
	if depth > 8 {
		return false
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	for i := range v.NumField() {
		if !v.Type().Field(i).Anonymous {
			continue
		}
		field := v.Field(i)
		if field.Kind() == reflect.Pointer && field.IsNil() {
			return true
		}
		if hasNilEmbeddedPointer(field, depth+1) {
			return true
		}
	}
	return false
}

// isNilFunc reports if v is a nil func, like ValidatorFunc(nil).
// A Valid or Validate method of a func type would call the nil func and panic.
func isNilFunc(v any) bool {
	val := reflect.ValueOf(v)
	return val.Kind() == reflect.Func && val.IsNil()
}

// DeepValidate recursively validates v and everything reachable from it,
// returning every validation error encountered.
//
// The traversal:
//   - calls [Validate] on v itself, then on every reachable element,
//     through a pointer to the element if only the pointer type
//     implements ValidatErr or Validator, like [LenString]
//     (an element that is not addressable is copied for that,
//     but not v itself: like [Validate], v passed by value is validated
//     with the methods of its own type, so a pointer receiver Validate
//     calling DeepValidate(*receiver) does not call itself again;
//     such a nested call has its own cycle detection though,
//     and its errors are reported again by the descent of an outer
//     DeepValidate, so prefer to check only cross-field rules in Validate),
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
	// The value passed to DeepValidate is not copied: a pointer receiver
	// Validate calling DeepValidate(*receiver) would otherwise be called
	// again for every copy without end
	isPassedValue := len(path) == 0
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
			err, isValidatable := TryValidate(validationTarget(v, !isPassedValue))
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
		isPassedValue = false
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
		// Iterate the entries instead of looking up the values by key,
		// because the value of a NaN key can't be looked up
		type entry struct{ key, value reflect.Value }
		entries := make([]entry, 0, v.Len())
		for iter := v.MapRange(); iter.Next(); {
			entries = append(entries, entry{iter.Key(), iter.Value()})
		}
		slices.SortFunc(entries, func(a, b entry) int { return CompareReflectValue(a.key, b.key) })
		for _, e := range entries {
			name := fmt.Sprintf("map value [%#v]", e.key.Interface())
			deepValidate(e.value, visiting, onError, append(path, name)...)
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
// case it is a pointer to v, or to a copy of v if v is not addressable
// and copyUnaddressable is true.
// Like Validate, a ValidatErr is preferred over a Validator.
func validationTarget(v reflect.Value, copyUnaddressable bool) any {
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
	if !copyUnaddressable {
		return val
	}
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)
	return ptr.Interface()
}
