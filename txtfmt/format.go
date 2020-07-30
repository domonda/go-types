package txtfmt

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strfmt"
	"github.com/ungerik/go-reflection"
)

// FormatValue formats the passed value following the format config.
func FormatValue(val reflect.Value, config *FormatConfig) string {
	derefVal, derefType := reflection.DerefValueAndType(val)
	if f, ok := config.TypeFormatters[derefType]; ok && derefVal.IsValid() {
		return f.FormatValue(derefVal, config)
	}

	if nullable.ReflectIsNull(val) {
		return config.Nil
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
		return strfmt.FormatFloat(
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
