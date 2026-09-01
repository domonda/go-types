package strfmt

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/nullable"
)

// Format the passed value following the format config.
// If the value implements encoding.TextMarshaler and MarshalText
// does not return an error, then this string is returned instead
// of more generic type conversions.
func Format(value any, config *FormatConfig) string {
	val, ok := value.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(value)
	}
	return FormatValue(val, config)
}

// FormatValue formats the passed reflect.Value following the format config.
// If the value implements encoding.TextMarshaler and MarshalText
// does not return an error, then this string is returned instead
// of more generic type conversions.
//
// The value of an unexported struct field is formatted like an exported
// one by reading it again through its address, instead of panicking in
// reflect.Value.Interface. Only a non-addressable one, like the field of
// a struct that was not passed as a pointer, is limited to the
// conversions based on its kind and fmt.Sprint.
func FormatValue(val reflect.Value, config *FormatConfig) string {
	if !val.IsValid() {
		return config.Nil
	}

	// A formatter, TextMarshaler or Stringer needs the value as any,
	// which reflect.Value.Interface panics for when the value was
	// obtained from an unexported struct field. Reading it again
	// through its address yields an interfaceable value of the same
	// type sharing the same memory, so those methods can be used.
	if !val.CanInterface() && val.CanAddr() {
		// The address is converted in the same expression as documented
		// for reflect.Value.UnsafeAddr and only read, never written
		val = reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem() //#nosec G103 -- reading an addressable value of its own type
	}
	derefVal, derefType := derefValueAndType(val)

	// Still false for a non-addressable value of an unexported field
	canInterface := val.CanInterface()

	if canInterface {
		if f, ok := config.TypeFormatters[derefType]; ok {
			return f.FormatValue(derefVal, config)
		}
	}

	// Asked for a value that can't be interfaced as well
	// because nullable.ReflectIsNull never panics for one
	if nullable.ReflectIsNull(val) {
		return config.Nil
	}

	if canInterface {
		textMarshaller, _ := val.Interface().(encoding.TextMarshaler)
		if textMarshaller == nil && val.CanAddr() {
			textMarshaller, _ = val.Addr().Interface().(encoding.TextMarshaler)
		}
		if textMarshaller == nil {
			textMarshaller, _ = derefVal.Interface().(encoding.TextMarshaler)
		}
		if textMarshaller != nil {
			text, err := textMarshaller.MarshalText()
			if err == nil {
				return string(text)
			}
			// Continue with the generic conversions below
			// when MarshalText did not yield a valid text
		}
	}

	switch derefType.Kind() {
	case reflect.Bool:
		if derefVal.Bool() {
			return config.True
		}
		return config.False

	case reflect.String:
		return derefVal.String()

	case reflect.Float32, reflect.Float64:
		return float.Format(
			derefVal.Float(),
			config.Float.ThousandsSep,
			config.Float.DecimalSep,
			config.Float.Precision,
			config.Float.PadPrecision,
		)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(derefVal.Int(), 10)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(derefVal.Uint(), 10)
	}

	if !canInterface {
		// No Stringer can be called for a non-addressable value of an
		// unexported struct field, so fall back to fmt, which formats
		// a reflect.Value as the value it holds, also one that can't
		// be passed to Interface()
		return fmt.Sprint(derefVal)
	}

	if s, ok := val.Interface().(fmt.Stringer); ok {
		return s.String()
	}
	if val.CanAddr() {
		if s, ok := val.Addr().Interface().(fmt.Stringer); ok {
			return s.String()
		}
	}
	if s, ok := derefVal.Interface().(fmt.Stringer); ok {
		return s.String()
	}

	switch x := derefVal.Interface().(type) {
	case []byte:
		return string(x)
	default:
		return fmt.Sprint(val.Interface())
	}
}

// derefValueAndType dereferences v through any non-nil pointers and returns
// the resulting value together with its type.
func derefValueAndType(v reflect.Value) (reflect.Value, reflect.Type) {
	for v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	return v, v.Type()
}
