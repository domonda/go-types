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

// Scan str to dest using the given ScanConfig.
func Scan(dest reflect.Value, str string, config *ScanConfig) (err error) {
	defer wraperr.WithFuncParams(&err, dest, str, config)

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
		return scaner.ScanString(dest, str, config)
	}

	switch x := dest.Addr().Interface().(type) {
	case Scannable:
		_, err = x.ScanString(str)
		return err
	case encoding.TextUnmarshaler:
		return x.UnmarshalText([]byte(str))
	}

	switch dest.Kind() {
	case reflect.Bool:
		s := strings.TrimSpace(str)
		switch {
		case strutil.StringIn(s, config.TrueStrings):
			dest.SetBool(true)
		case strutil.StringIn(s, config.FalseStrings):
			dest.SetBool(false)
		default:
			return wraperr.Errorf("can't scan %q as bool", str)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
		if err != nil {
			return wraperr.Errorf("can't scan %q as int because %w", str, err)
		}
		dest.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(strings.TrimSpace(str), 10, 64)
		if err != nil {
			return wraperr.Errorf("can't scan %q as uint because %w", str, err)
		}
		dest.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := ParseFloat(str)
		if err != nil {
			return wraperr.Errorf("can't scan %q as float because %w", str, err)
		}
		dest.SetFloat(f)

	case reflect.String:
		dest.SetString(str)

	default:
		return wraperr.Errorf("can't scan to type %s", dest.Type())
	}

	// Validate scanned value if dest implements types.ValidatErr or types.Validator
	switch x := dest.Interface().(type) {
	case types.ValidatErr:
		if err := x.Validate(); err != nil {
			return wraperr.Errorf("error validating %s value scanned from %q because %w", dest.Type(), str, err)
		}
		return nil

	case types.Validator:
		if !x.Valid() {
			return wraperr.Errorf("error validating %s value scanned from %q", dest.Type(), str)
		}
		return nil
	}

	// Validate scanned value if dest pointer implements types.ValidatErr or types.Validator
	switch x := dest.Addr().Interface().(type) {
	case types.ValidatErr:
		if err := x.Validate(); err != nil {
			return wraperr.Errorf("error validating %s value scanned from %q because %w", dest.Type(), str, err)
		}
		return nil

	case types.Validator:
		if !x.Valid() {
			return wraperr.Errorf("error validating %s value scanned from %q", dest.Type(), str)
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
