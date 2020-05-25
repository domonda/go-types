package strfmt

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"
	"time"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/strutil"
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

	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		dest = dest.Elem()
	}

	if scaner, ok := config.TypeScanners[dest.Type()]; ok {
		return scaner.ScanString(dest, source, config)
	}

	switch x := dest.Addr().Interface().(type) {
	case Scannable:
		_, err = x.ScanString(source)
		return err

	case encoding.TextUnmarshaler:
		return x.UnmarshalText([]byte(source))
	}

	switch dest.Kind() {
	case reflect.Bool:
		s := strings.TrimSpace(source)
		switch {
		case strutil.StringIn(s, config.TrueStrings):
			dest.SetBool(true)
		case strutil.StringIn(s, config.FalseStrings):
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

	case reflect.String:
		dest.SetString(source)

	default:
		return wraperr.Errorf("can't scan %q as destination type %s", source, dest.Type())
	}

	// Validate scanned value if dest implements types.ValidatErr or types.Validator
	switch x := dest.Interface().(type) {
	case types.ValidatErr:
		if err := x.Validate(); err != nil {
			return wraperr.Errorf("error validating %s value scanned from %q because %w", dest.Type(), source, err)
		}
		return nil

	case types.Validator:
		if !x.Valid() {
			return wraperr.Errorf("error validating %s value scanned from %q", dest.Type(), source)
		}
		return nil
	}

	// Validate scanned value if dest pointer implements types.ValidatErr or types.Validator
	switch x := dest.Addr().Interface().(type) {
	case types.ValidatErr:
		if err := x.Validate(); err != nil {
			return wraperr.Errorf("error validating %s value scanned from %q because %w", dest.Type(), source, err)
		}
		return nil

	case types.Validator:
		if !x.Valid() {
			return wraperr.Errorf("error validating %s value scanned from %q", dest.Type(), source)
		}
		return nil
	}

	return nil
}

func scanTimeString(dest reflect.Value, str string, config *ScanConfig) error {
	s := strings.TrimSpace(str)
	for _, config := range config.TimeFormats {
		t, err := time.Parse(config, s)
		if err == nil {
			dest.Set(reflect.ValueOf(t))
			return nil
		}
	}
	return wraperr.Errorf("can't scan %q as time.Time", str)
}

func scanDurationString(dest reflect.Value, str string, config *ScanConfig) error {
	d, err := time.ParseDuration(strings.TrimSpace(str))
	if err != nil {
		return wraperr.Errorf("can't scan %q as time.Duration because %w", str, err)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}
