package strfmt

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/strutil"
)

// nullSetter is the SetNull method of the nullable.NullSetable interface
// implemented by types that can represent a null value.
type nullSetter interface {
	SetNull()
}

// Scan source into dest using the given ScanConfig.
//
// A Scanner registered for the destination type in config.TypeScanners is
// asked before all built-in scanning logic, including the nil source string
// handling below, so that it owns its type completely.
//
// A source string that config.IsNil reports as nil, ignoring surrounding
// whitespace, means "no value" like an empty cell of a CSV file or a
// spreadsheet. A kind that can hold nil is set to nil, a nullable type
// is set to null with its SetNull method, and a string destination
// without a scanning method of its own is assigned the source string.
//
// Any other destination type can't represent "no value", so what a nil
// source string means for it depends on config.StrictEmptyStringParsing:
// with strict parsing it is an error, unless the type has a scanning
// method of its own that gives it a meaning, and without strict parsing
// the zero value of the destination type is assigned.
//
// A destination set to nil, null or its zero value for a nil source
// string is not passed to config.ValidateFunc because it holds the
// absence of a value instead of a scanned one.
//
// A failed scan never modifies dest: a destination that a scanning
// method or the validation left partially assigned is restored to the
// value it had before, so a caller scanning a row of cells never ends
// up with a half written field.
//
// A nil pointer destination is allocated, an already allocated one is
// scanned in place and keeps the fields that the scan doesn't assign.
// Multiple levels of indirection are allocated and scanned as well.
//
// An error is returned instead of a panic if dest is not valid
// or not settable, like the value of an unexported struct field.
// A pointer that is not settable itself, like reflect.ValueOf(&x),
// is scanned by scanning the non-nil value it points to.
func Scan(dest reflect.Value, source string, config *ScanConfig) error {
	// Errors are wrapped below instead of with a deferred call
	// so that a successful scan doesn't have to box the destination
	// value for an error message that is never built
	err := scanValue(dest, source, config)
	if err == nil {
		return nil
	}
	// dest.Interface() panics for a destination that is invalid or can't
	// be interfaced, like the value of an unexported struct field, so
	// the wrapping uses nil for such a destination instead of leaving
	// the errors of those cases unwrapped.
	var destVal any
	if dest.IsValid() && dest.CanInterface() {
		destVal = dest.Interface()
	}
	errs.WrapWith3FuncParams(&err, destVal, source, config)
	return err
}

// scanValue implements Scan without wrapping the returned error so that
// the recursion for pointer destinations wraps it only once, in Scan.
func scanValue(dest reflect.Value, source string, config *ScanConfig) error {
	if !dest.IsValid() {
		return fmt.Errorf("can't scan %q into an invalid destination value", source)
	}
	if config == nil {
		return fmt.Errorf("can't scan %q using nil ScanConfig", source)
	}
	if !dest.CanSet() {
		// dest.Addr() and the Set methods panic instead of returning an
		// error for a non-settable destination like the value of an
		// unexported struct field. Only a non-nil pointer can still be
		// scanned, by scanning the settable value it points to, and only
		// when the pointer itself is not what a scanner registered for
		// its type would set.
		_, hasTypeScanner := config.TypeScanners[dest.Type()]
		if dest.Kind() == reflect.Pointer && !hasTypeScanner && dest.Elem().CanSet() {
			return scanValue(dest.Elem(), source, config)
		}
		return fmt.Errorf("can't scan %q into a non-settable destination of type %s", source, dest.Type())
	}

	// The scanning methods below write into dest directly, and ValidateFunc
	// runs after the kind conversions have assigned, so a failed scan can
	// leave a partially assigned destination behind. Restoring the previous
	// value keeps a failed scan from modifying dest at all, so that a caller
	// scanning a row of cells never ends up with a half written field.
	// A pointer restores its own pointed to value in the recursion below,
	// which is why the copy here only has to be a shallow one.
	previous := reflect.New(dest.Type()).Elem()
	previous.Set(dest)
	err := scanSettableValue(dest, source, config)
	if err != nil {
		dest.Set(previous)
	}
	return err
}

// scanSettableValue scans source into the valid and settable dest without
// restoring it if the scan fails, which scanValue does for it.
func scanSettableValue(dest reflect.Value, source string, config *ScanConfig) error {
	// A scanner registered for the destination type owns that type
	// completely: it is asked before all built-in scanning logic,
	// including the nil source string handling below, so that it can
	// give a nil source string its own meaning. The scanners that this
	// package registers by default apply the same nil source string
	// handling themselves, see scanNilStringToNonOptional.
	if scanner, ok := config.TypeScanners[dest.Type()]; ok {
		return scanner.ScanString(dest, source, config)
	}

	trimmed := strutil.TrimSpace(source)

	// A nil source string means "no value" like an empty cell of a CSV
	// file or a spreadsheet. Handled here so that every type gets the
	// same nil handling instead of every scanning method repeating it.
	// The assignments below are not passed to config.ValidateFunc
	// because the absence of a value is not a value to validate, and
	// calling a Valid method with a value receiver through a nil
	// pointer would panic.
	if config.IsNil(trimmed) {
		// Kinds that can hold nil are set to nil
		if nilableKind(dest.Kind()) {
			dest.SetZero()
			return nil
		}
		// Nullable types are set to null
		destPtr := dest.Addr().Interface()
		if setter, ok := destPtr.(nullSetter); ok {
			setter.SetNull()
			return nil
		}
		switch {
		case dest.Kind() == reflect.String && !hasScanMethod(destPtr):
			// A string destination without a scanning method of its own
			// can hold the source string itself like any other string,
			// which the scanning below assigns

		case !config.StrictEmptyStringParsing:
			// Without strict parsing the destination type doesn't have
			// to be able to represent "no value", it's assigned its
			// zero value as the closest thing to an empty one
			dest.SetZero()
			return nil

		case hasScanMethod(destPtr):
			// With strict parsing a type with a scanning method of its
			// own decides what a nil source string means for it,
			// which the scanning below asks it for

		default:
			return fmt.Errorf("can't scan nil string %q into non-optional destination type %s with strict empty string parsing", source, dest.Type())
		}
	}

	if dest.Kind() == reflect.Pointer {
		// Only a nil pointer is allocated, an already allocated one is
		// scanned in place so that it keeps its identity and the fields
		// that the scanning below doesn't assign. A scanning method that
		// assigns only some fields of a struct would lose the others by
		// scanning into a newly allocated value instead.
		// The recursion also scans through any further levels of
		// indirection, and restores the pointed to value if it fails,
		// while scanValue restores this pointer itself.
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		return scanValue(dest.Elem(), source, config)
	}

	switch x := dest.Addr().Interface().(type) {
	case Scannable:
		return x.ScanString(source, config.ValidateFunc != nil)

	case encoding.TextUnmarshaler:
		return x.UnmarshalText([]byte(source))
	}

	// The numbers below are parsed with the bit size of the destination
	// type because the SetInt, SetUint and SetFloat methods would
	// silently truncate a value that doesn't fit into it.
	switch dest.Kind() {
	case reflect.String:
		dest.SetString(source)

	case reflect.Bool:
		switch {
		case config.IsTrue(trimmed):
			dest.SetBool(true)
		case config.IsFalse(trimmed):
			dest.SetBool(false)
		default:
			return fmt.Errorf("can't scan %q as %s", source, dest.Type())
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(trimmed, 10, dest.Type().Bits())
		if err != nil {
			return fmt.Errorf("can't scan %q as %s because %w", source, dest.Type(), err)
		}
		dest.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(trimmed, 10, dest.Type().Bits())
		if err != nil {
			return fmt.Errorf("can't scan %q as %s because %w", source, dest.Type(), err)
		}
		dest.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := float.Parse(source)
		if err != nil {
			return fmt.Errorf("can't scan %q as %s because %w", source, dest.Type(), err)
		}
		if dest.OverflowFloat(f) {
			return fmt.Errorf("can't scan %q as %s because the value is out of range", source, dest.Type())
		}
		dest.SetFloat(f)

	default:
		return fmt.Errorf("can't scan %q as destination type %s", source, dest.Type())
	}

	if config.ValidateFunc != nil {
		err := config.ValidateFunc(dest.Interface())
		if err != nil {
			return fmt.Errorf("error validating %s value scanned from %q because %w", dest.Type(), source, err)
		}
	}

	return nil
}

// hasScanMethod reports whether destPtr, the pointer to a scan
// destination, has a scanning method that lets the type itself
// decide what a source string means for it.
func hasScanMethod(destPtr any) bool {
	switch destPtr.(type) {
	case Scannable, encoding.TextUnmarshaler:
		return true
	}
	return false
}

// nilableKind reports whether a destination of the passed kind
// can hold nil as representation of "no value".
// Those are the same kinds that nullable.ReflectIsNull
// reports as null when they are nil.
func nilableKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Pointer,
		reflect.Interface,
		reflect.Map,
		reflect.Slice,
		reflect.Chan,
		reflect.Func,
		reflect.UnsafePointer:
		return true
	}
	return false
}
