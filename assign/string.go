package assign

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/domonda/errors"
	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/strfmt"
	"github.com/domonda/go-types/strutil"
)

// String assigns str to dest using config.
func String(str string, config *StringConfig, dest reflect.Value) error {
	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		dest = dest.Elem()
	}

	if assigner, ok := config.TypeAssigners[dest.Type()]; ok {
		return assigner.AssignString(str, config, dest)
	}

	if assignable, ok := dest.Addr().Interface().(strfmt.StringAssignable); ok {
		_, err := assignable.AssignString(str)
		return err
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
			return errors.Errorf("can't assign %q as bool", str)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
		if err != nil {
			return errors.Wrapf(err, "can't assign %q as int", str)
		}
		dest.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(strings.TrimSpace(str), 10, 64)
		if err != nil {
			return errors.Wrapf(err, "can't assign %q as uint", str)
		}
		dest.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strfmt.ParseFloat(str)
		if err != nil {
			return errors.Wrapf(err, "can't assign %q as float", str)
		}
		dest.SetFloat(f)

	case reflect.String:
		dest.SetString(str)

	default:
		return errors.Errorf("can't assign to type %s", dest.Type())
	}

	// Validate assigned value if dest implements types.ValidatErr or types.Validator
	switch x := dest.Interface().(type) {
	case types.ValidatErr:
		return errors.Wrapf(x.Validate(), "error validating %s value assigned from %q", dest.Type(), str)

	case types.Validator:
		if x.Valid() {
			return nil
		} else {
			return errors.Errorf("error validating %s value assigned from %q", dest.Type(), str)
		}
	}

	// Validate assigned value if dest pointer implements types.ValidatErr or types.Validator
	switch x := dest.Addr().Interface().(type) {
	case types.ValidatErr:
		return errors.Wrapf(x.Validate(), "error validating %s value assigned from %q", dest.Type(), str)

	case types.Validator:
		if x.Valid() {
			return nil
		} else {
			return errors.Errorf("error validating %s value assigned from %q", dest.Type(), str)
		}
	}

	return nil
}

func assignTimeString(str string, config *StringConfig, dest reflect.Value) error {
	s := strings.TrimSpace(str)
	for _, format := range config.TimeFormats {
		t, err := time.Parse(format, s)
		if err == nil {
			dest.Set(reflect.ValueOf(t))
			return nil
		}
	}
	return errors.Errorf("can't assign %q as time.Time", str)
}

func assignDurationString(str string, config *StringConfig, dest reflect.Value) error {
	d, err := time.ParseDuration(strings.TrimSpace(str))
	if err != nil {
		return errors.Wrapf(err, "can't assign %q as time.Duration", str)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}

func assignDateString(str string, config *StringConfig, dest reflect.Value) error {
	d, err := date.Normalize(str)
	if err != nil {
		return errors.Wrapf(err, "can't assign %q as date.Date", str)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}

func assignNullableDateString(str string, config *StringConfig, dest reflect.Value) error {
	d, err := date.NormalizeNullable(str)
	if err != nil {
		return errors.Wrapf(err, "can't assign %q as date.NullableDate", str)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}

func assignMoneyAmountString(str string, config *StringConfig, dest reflect.Value) error {
	a, err := money.ParseAmount(str, config.MoneyAmountDecimals...)
	if err != nil {
		return errors.Wrapf(err, "can't assign %q as money.Amount", str)
	}
	dest.Set(reflect.ValueOf(a))
	return nil
}

func assignMoneyCurrencyAmountString(str string, config *StringConfig, dest reflect.Value) error {
	ca, err := money.ParseCurrencyAmount(str, config.MoneyAmountDecimals...)
	if err != nil {
		return errors.Wrapf(err, "can't assign %q as money.CurrencyAmount", str)
	}
	dest.Set(reflect.ValueOf(ca))
	return nil
}
