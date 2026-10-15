package strfmt

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"unsafe"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/nullable"
)

// Format the passed value following the format config.
// A nil config formats with the defaults of NewFormatConfig,
// like Scan scans with DefaultScanConfig for a nil ScanConfig.
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

// defaultFormatConfig is the config for a nil config, built once
// instead of on every call. FormatValue and the default formatters
// only read it.
var defaultFormatConfig = sync.OnceValue(NewFormatConfig)

// FormatValue formats the passed reflect.Value following the format config.
// A nil config formats with the defaults of NewFormatConfig.
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
	if config == nil {
		config = defaultFormatConfig()
	}
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

	case reflect.Float32:
		// Format with float32 precision because the shortest float64
		// representation of a float32 shows its binary noise,
		// like 0.10000000149011612 for float32(0.1)
		return float.Format(
			float32(derefVal.Float()),
			config.Float.ThousandsSep,
			config.Float.DecimalSep,
			config.Float.Precision,
			config.Float.PadPrecision,
		)

	case reflect.Float64:
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

	// A nullable.Type[T] has no String or MarshalText method, so a non-null
	// one is formatted like the T it wraps instead of as a struct
	if _, ok := nullSetableValueType(derefType); ok {
		if get := derefVal.MethodByName("Get"); get.IsValid() {
			return FormatValue(get.Call(nil)[0], config)
		}
	}

	switch x := derefVal.Interface().(type) {
	case []byte:
		return string(x)
	default:
		return fmt.Sprint(val.Interface())
	}
}

// nullSetableMethods are the methods of nullable.NullSetable[T]
// that don't depend on T.
var nullSetableMethods = reflect.TypeFor[interface {
	nullable.Nullable
	SetNull()
}]()

// nullSetableValueType returns the type T if the pointer type of t has the
// methods of nullable.NullSetable[T], like nullable.Type[T] and the
// money.NullableAmount alias of it. The generic interface can't be asserted
// without knowing T, so the methods depending on T are checked by reflection.
func nullSetableValueType(t reflect.Type) (valueType reflect.Type, ok bool) {
	ptrType := reflect.PointerTo(t)
	if !ptrType.Implements(nullSetableMethods) {
		return nil, false
	}
	// The method types of a reflect.Type have the receiver as first argument
	get, ok := ptrType.MethodByName("Get")
	if !ok || get.Type.NumIn() != 1 || get.Type.NumOut() != 1 {
		return nil, false
	}
	valueType = get.Type.Out(0)
	set, ok := ptrType.MethodByName("Set")
	if !ok || set.Type.NumIn() != 2 || set.Type.NumOut() != 0 || set.Type.In(1) != valueType {
		return nil, false
	}
	return valueType, true
}

// derefValueAndType dereferences v through any non-nil pointers and returns
// the resulting value together with its type.
func derefValueAndType(v reflect.Value) (reflect.Value, reflect.Type) {
	for v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	return v, v.Type()
}
