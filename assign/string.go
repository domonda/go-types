package assign

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"
	"time"

	types "github.com/domonda/go-types"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/strfmt"
	"github.com/domonda/go-types/strutil"
	"github.com/domonda/go-wraperr"
)

// String assigns str to dest using the given StringParser.
func String(dest reflect.Value, str string, parser *StringParser) (err error) {
	defer wraperr.WithFuncParams(&err, dest, str, parser)

	if parser == nil {
		return wraperr.Errorf("nil StringParser")
	}

	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		dest = dest.Elem()
	}

	if assigner, ok := parser.TypeAssigners[dest.Type()]; ok {
		return assigner.AssignString(dest, str, parser)
	}

	switch x := dest.Addr().Interface().(type) {
	case StringAssignable:
		_, err = x.AssignString(str)
		return err
	case encoding.TextUnmarshaler:
		return x.UnmarshalText([]byte(str))
	}

	switch dest.Kind() {
	case reflect.Bool:
		s := strings.TrimSpace(str)
		switch {
		case strutil.StringIn(s, parser.TrueStrings):
			dest.SetBool(true)
		case strutil.StringIn(s, parser.FalseStrings):
			dest.SetBool(false)
		default:
			return wraperr.Errorf("can't assign %q as bool", str)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
		if err != nil {
			return wraperr.Errorf("can't assign %q as int because %w", str, err)
		}
		dest.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(strings.TrimSpace(str), 10, 64)
		if err != nil {
			return wraperr.Errorf("can't assign %q as uint because %w", str, err)
		}
		dest.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strfmt.ParseFloat(str)
		if err != nil {
			return wraperr.Errorf("can't assign %q as float because %w", str, err)
		}
		dest.SetFloat(f)

	case reflect.String:
		dest.SetString(str)

	default:
		return wraperr.Errorf("can't assign to type %s", dest.Type())
	}

	// Validate assigned value if dest implements types.ValidatErr or types.Validator
	switch x := dest.Interface().(type) {
	case types.ValidatErr:
		if err := x.Validate(); err != nil {
			return wraperr.Errorf("error validating %s value assigned from %q because %w", dest.Type(), str, err)
		}
		return nil

	case types.Validator:
		if !x.Valid() {
			return wraperr.Errorf("error validating %s value assigned from %q", dest.Type(), str)
		}
		return nil
	}

	// Validate assigned value if dest pointer implements types.ValidatErr or types.Validator
	switch x := dest.Addr().Interface().(type) {
	case types.ValidatErr:
		if err := x.Validate(); err != nil {
			return wraperr.Errorf("error validating %s value assigned from %q because %w", dest.Type(), str, err)
		}
		return nil

	case types.Validator:
		if !x.Valid() {
			return wraperr.Errorf("error validating %s value assigned from %q", dest.Type(), str)
		}
		return nil
	}

	return nil
}

func assignTimeString(dest reflect.Value, str string, parser *StringParser) error {
	s := strings.TrimSpace(str)
	for _, parser := range parser.TimeFormats {
		t, err := time.Parse(parser, s)
		if err == nil {
			dest.Set(reflect.ValueOf(t))
			return nil
		}
	}
	return wraperr.Errorf("can't assign %q as time.Time", str)
}

func assignDurationString(dest reflect.Value, str string, parser *StringParser) error {
	d, err := time.ParseDuration(strings.TrimSpace(str))
	if err != nil {
		return wraperr.Errorf("can't assign %q as time.Duration because %w", str, err)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}

func assignDateString(dest reflect.Value, str string, parser *StringParser) error {
	d, err := date.Normalize(str)
	if err != nil {
		return wraperr.Errorf("can't assign %q as date.Date because %w", str, err)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}

func assignNullableDateString(dest reflect.Value, str string, parser *StringParser) error {
	d, err := date.NormalizeNullable(str)
	if err != nil {
		return wraperr.Errorf("can't assign %q as date.NullableDate because %w", str, err)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}

func assignMoneyAmountString(dest reflect.Value, str string, parser *StringParser) error {
	a, err := money.ParseAmount(str, parser.AcceptedMoneyAmountDecimals...)
	if err != nil {
		return wraperr.Errorf("can't assign %q as money.Amount because %w", str, err)
	}
	dest.Set(reflect.ValueOf(a))
	return nil
}

func assignMoneyCurrencyAmountString(dest reflect.Value, str string, parser *StringParser) error {
	ca, err := money.ParseCurrencyAmount(str, parser.AcceptedMoneyAmountDecimals...)
	if err != nil {
		return wraperr.Errorf("can't assign %q as money.CurrencyAmount because %w", str, err)
	}
	dest.Set(reflect.ValueOf(ca))
	return nil
}
