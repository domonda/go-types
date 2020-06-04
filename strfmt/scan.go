package strfmt

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"
	"time"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-wraperr"
)

// Scan source into dest using the given ScanConfig.
// If dest is an assignable nil pointer variable,
// then a new object of the pointed to type will be allocated and set.
func Scan(dest reflect.Value, source string, config *ScanConfig) (err error) {
	defer wraperr.WithFuncParams(&err, dest, source, config)

	if config == nil {
		return wraperr.Errorf("nil ScanConfig")
	}

	// First priority is to check if there is a custom scanner for the type
	if scaner, ok := config.TypeScanners[dest.Type()]; ok {
		return scaner.ScanString(dest, source, config)
	}

	if dest.Kind() == reflect.Ptr {
		if config.IsNil(source) {
			// If dest is a pointer type and source is a nil string
			// then set pointer to nil (the zero value of the pointer)
			dest.Set(reflect.Zero(dest.Type()))
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
		_, err = x.ScanString(source)
		return err

	case encoding.TextUnmarshaler:
		return x.UnmarshalText([]byte(source))
	}

	switch dest.Kind() {
	case reflect.String:
		dest.SetString(source)

	case reflect.Bool:
		s := strings.TrimSpace(source)
		switch {
		case config.IsTrue(s):
			dest.SetBool(true)
		case config.IsFalse(s):
			dest.SetBool(false)
		default:
			return wraperr.Errorf("can't scan %q as bool", source)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(strings.TrimSpace(source), 10, 64)
		if err != nil {
			return wraperr.Errorf("can't scan %q as int because %w", source, err)
		}
		dest.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(strings.TrimSpace(source), 10, 64)
		if err != nil {
			return wraperr.Errorf("can't scan %q as uint because %w", source, err)
		}
		dest.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := ParseFloat(source)
		if err != nil {
			return wraperr.Errorf("can't scan %q as float because %w", source, err)
		}
		dest.SetFloat(f)

	default:
		return wraperr.Errorf("can't scan %q as destination type %s", source, dest.Type())
	}

	// Validate scanned value if dest or dest pointer implements types.ValidatErr or types.Validator
	err, isValidator := types.TryValidate(dest.Interface())
	if !isValidator {
		err, isValidator = types.TryValidate(dest.Addr().Interface())
	}
	if err != nil {
		return wraperr.Errorf("error validating %s value scanned from %q because %w", dest.Type(), source, err)
	}

	return nil
}

func scanTimeString(dest reflect.Value, str string, config *ScanConfig) error {
	t, ok := config.ParseTime(strings.TrimSpace(str))
	if !ok {
		return wraperr.Errorf("can't scan %q as time.Time", str)
	}
	dest.Set(reflect.ValueOf(t))
	return nil
}

func scanDurationString(dest reflect.Value, str string, config *ScanConfig) error {
	d, err := time.ParseDuration(strings.TrimSpace(str))
	if err != nil {
		return wraperr.Errorf("can't scan %q as time.Duration because %w", str, err)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}
