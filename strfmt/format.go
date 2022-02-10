package strfmt

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/nullable"
	"github.com/ungerik/go-reflection"
)

// Format the passed value following the format config.
// If the value implements encoding.TextMarshaler and MarshalText
// does not return an error, then this string is returned instead
// of more generic type conversions.
func Format(value interface{}, config *FormatConfig) string {
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
func FormatValue(val reflect.Value, config *FormatConfig) string {
	if !val.IsValid() {
		return config.Nil
	}
	derefVal, derefType := reflection.DerefValueAndType(val)
	if f, ok := config.TypeFormatters[derefType]; ok && derefVal.IsValid() {
		return f.FormatValue(derefVal, config)
	}

	if nullable.ReflectIsNull(val) {
		return config.Nil
	}

	textMarshaller, _ := val.Interface().(encoding.TextMarshaler)
	if textMarshaller == nil && val.CanAddr() {
		textMarshaller, _ = val.Addr().Interface().(encoding.TextMarshaler)
	}
	if textMarshaller == nil {
		textMarshaller, _ = derefVal.Interface().(encoding.TextMarshaler)
	}
	if textMarshaller != nil {
		text, err := textMarshaller.MarshalText()
		if err != nil {
			return string(text)
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
