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

// Scan source into dest using the given ScanConfig.
// If dest is an assignable nil pointer variable,
// then a new object of the pointed to type will be allocated and set.
func Scan(dest reflect.Value, source string, config *ScanConfig) (err error) {
	defer errs.WrapWithFuncParams(&err, dest.Interface(), source, config)

	if config == nil {
		return fmt.Errorf("can't scan %q using nil ScanConfig", source)
	}

	// First priority is to check if there is a custom scanner for the type
	if scaner, ok := config.TypeScanners[dest.Type()]; ok {
		return scaner.ScanString(dest, source, config)
	}

	if dest.Kind() == reflect.Pointer {
		if config.IsNil(source) {
			// If dest is a pointer type and source is a nil string
			// then set pointer to nil (the zero value of the pointer)
			dest.SetZero()
			return nil
		}
		if dest.IsNil() {
			// If source is not a nil string and dest is nil
			// then allocate and set pointer
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		// Use pointed to type in further code because dest.Addr()
		// will be used where only a pointer makes sense
		dest = dest.Elem()
	}

	switch x := dest.Addr().Interface().(type) {
	case Scannable:
		return x.ScanString(source, config.ValidateFunc != nil)

	case encoding.TextUnmarshaler:
		return x.UnmarshalText([]byte(source))
	}

	switch dest.Kind() {
	case reflect.String:
		dest.SetString(source)

	case reflect.Bool:
		s := strutil.TrimSpace(source)
		switch {
		case config.IsTrue(s):
			dest.SetBool(true)
		case config.IsFalse(s):
			dest.SetBool(false)
		default:
			return fmt.Errorf("can't scan %q as bool", source)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(strutil.TrimSpace(source), 10, 64)
		if err != nil {
			return fmt.Errorf("can't scan %q as int because %w", source, err)
		}
		dest.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(strutil.TrimSpace(source), 10, 64)
		if err != nil {
			return fmt.Errorf("can't scan %q as uint because %w", source, err)
		}
		dest.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := float.Parse(source)
		if err != nil {
			return fmt.Errorf("can't scan %q as float because %w", source, err)
		}
		dest.SetFloat(f)

	default:
		return fmt.Errorf("can't scan %q as destination type %s", source, dest.Type())
	}

	if config.ValidateFunc != nil {
		err = config.ValidateFunc(dest.Interface())
		if err != nil {
			return fmt.Errorf("error validating %s value scanned from %q because %w", dest.Type(), source, err)
		}
	}

	return nil
}
