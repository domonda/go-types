package stringfrom

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/uu"
)

// All *Ptr functions share the same nil contract: a nil pointer never
// panics and is converted to the concatenation of the strForNil
// arguments, which is the empty string if none are passed.
// That way templates and log messages can render optional values
// without nil checks and choose their own placeholder.

func TestPtr(t *testing.T) {
	s := "Grüße"
	empty := ""
	assert.Equal(t, "Grüße", Ptr(&s))
	assert.Equal(t, "Grüße", Ptr(&s, "default"), "default is only used for nil")
	// A pointer to an empty string is not nil, so the default is not used
	assert.Equal(t, "", Ptr(&empty, "default"))

	assert.Equal(t, "", Ptr(nil))
	assert.Equal(t, "n/a", Ptr(nil, "n/a"))
	assert.Equal(t, "<nil>", Ptr(nil, "<", "nil", ">"), "strForNil is concatenated")
}

func TestTimePtr(t *testing.T) {
	tm := time.Date(2024, 2, 29, 13, 14, 15, 0, time.UTC)
	assert.Equal(t, tm.String(), TimePtr(&tm))
	assert.Equal(t, tm.String(), TimePtr(&tm, "never"))

	assert.Equal(t, "", TimePtr(nil))
	assert.Equal(t, "never", TimePtr(nil, "never"))
	// Unlike the other functions, the zero value is treated like nil
	// because a zero time.Time means "no time" and would otherwise be
	// rendered as the meaningless "0001-01-01 00:00:00 +0000 UTC"
	var zero time.Time
	assert.Equal(t, "never", TimePtr(&zero, "never"))
	assert.Equal(t, "", TimePtr(&zero))
}

func TestUUIDPtr(t *testing.T) {
	id := uu.IDFrom("ad5d4c20-6be6-4f22-a39f-8b1c4bd7f7a0")
	assert.Equal(t, "ad5d4c20-6be6-4f22-a39f-8b1c4bd7f7a0", UUIDPtr(&id))

	// Only a nil pointer gets the default, uu.IDNil is a valid value
	nilID := uu.IDNil
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", UUIDPtr(&nilID, "none"))

	assert.Equal(t, "", UUIDPtr(nil))
	assert.Equal(t, "none", UUIDPtr(nil, "none"))
}

func TestIntPtr(t *testing.T) {
	assert.Equal(t, "0", IntPtr(new(0), "-"), "zero is a value, not nil")
	assert.Equal(t, "42", IntPtr(new(42)))
	assert.Equal(t, "-42", IntPtr(new(-42)))
	assert.Equal(t, "", IntPtr(nil))
	assert.Equal(t, "-", IntPtr(nil, "-"))
}

func TestUintPtr(t *testing.T) {
	assert.Equal(t, "0", UintPtr(new(uint(0)), "-"), "zero is a value, not nil")
	assert.Equal(t, "4294967295", UintPtr(new(uint(math.MaxUint32))))
	assert.Equal(t, "", UintPtr(nil))
	assert.Equal(t, "-", UintPtr(nil, "-"))
}

func TestInt64Ptr(t *testing.T) {
	assert.Equal(t, "0", Int64Ptr(new(int64(0)), "-"), "zero is a value, not nil")
	assert.Equal(t, "9223372036854775807", Int64Ptr(new(int64(math.MaxInt64))))
	assert.Equal(t, "-9223372036854775808", Int64Ptr(new(int64(math.MinInt64))))
	assert.Equal(t, "", Int64Ptr(nil))
	assert.Equal(t, "-", Int64Ptr(nil, "-"))
}

func TestFloat64Ptr(t *testing.T) {
	// Floats are formatted with the minimal number of digits that
	// still parse back to the same value and never in exponent notation,
	// so large and small numbers stay readable as plain decimals.
	tests := []struct {
		f    float64
		want string
	}{
		{f: 0, want: "0"},
		{f: 1, want: "1"},
		{f: -1.5, want: "-1.5"},
		{f: 0.1, want: "0.1"},
		{f: 1234567.891, want: "1234567.891"},
		{f: 1e21, want: "1000000000000000000000"},
		{f: 1e-7, want: "0.0000001"},
		{f: math.Inf(1), want: "+Inf"},
		{f: math.NaN(), want: "NaN"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, Float64Ptr(&tt.f, "-"))
		})
	}
	assert.Equal(t, "", Float64Ptr(nil))
	assert.Equal(t, "-", Float64Ptr(nil, "-"))
}

func TestBoolPtr(t *testing.T) {
	assert.Equal(t, "true", BoolPtr(new(true), "-"))
	assert.Equal(t, "false", BoolPtr(new(false), "-"), "false is a value, not nil")
	assert.Equal(t, "", BoolPtr(nil))
	assert.Equal(t, "unknown", BoolPtr(nil, "unknown"))
}

func TestBool(t *testing.T) {
	assert.Equal(t, "Ja", Bool(true, "Ja", "Nein"))
	assert.Equal(t, "Nein", Bool(false, "Ja", "Nein"))
	assert.Equal(t, "", Bool(false, "yes", ""))
}

func TestInterface(t *testing.T) {
	s := "Grüße"
	assert.Equal(t, "Grüße", Interface(&s))
	assert.Equal(t, "", Interface((*string)(nil)), "typed nil pointer")
}

// Interface renders arbitrary, often optional values like the other
// functions of this package, so it must not panic for nil or for a
// value that is not a pointer: it used to panic for an untyped nil,
// for any non-pointer value except nil maps, slices, funcs and channels,
// and for any non-nil map, slice, func or channel.
// Nil returns "" like documented and a value returns the same string
// as a pointer to it, so callers don't have to care which one they have.
func TestInterface_NilAndNonPointer(t *testing.T) {
	type named string
	var (
		nilMap   map[string]int
		nilSlice []int
		nilPtr   *string
	)
	for _, tt := range []struct {
		name string
		val  any
		want string
	}{
		{name: "untyped nil", val: nil, want: ""},
		{name: "typed nil pointer", val: (*int)(nil), want: ""},
		{name: "pointer to nil pointer", val: &nilPtr, want: ""},
		{name: "nil map", val: nilMap, want: ""},
		{name: "pointer to nil map", val: &nilMap, want: ""},
		{name: "nil slice", val: nilSlice, want: ""},
		{name: "pointer to nil slice", val: &nilSlice, want: ""},
		{name: "string", val: "Grüße", want: "Grüße"},
		{name: "named string type", val: named("x"), want: "x"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				assert.Equal(t, tt.want, Interface(tt.val))
			})
		})
	}

	// A non-pointer value returns the same string as a pointer to it
	same := func(t *testing.T, name string, val, ptr any) {
		t.Helper()
		require.NotPanics(t, func() {
			assert.Equal(t, Interface(ptr), Interface(val), name)
		})
	}
	str, n, m, sl, st := "x", 42, map[string]int{"a": 1}, []int{1}, struct{ A int }{1}
	same(t, "string", str, &str)
	same(t, "int", n, &n)
	same(t, "map", m, &m)
	same(t, "slice", sl, &sl)
	same(t, "struct", st, &st)
	same(t, "nil map", nilMap, &nilMap)
	same(t, "nil slice", nilSlice, &nilSlice)
}

// Interface renders values for humans like the other functions of this
// package, so a non-string value must be formatted like fmt.Sprint does,
// not as reflect's "<int Value>" placeholder that hides the value.
func TestInterface_NonStringValues(t *testing.T) {
	n := 42
	pn := &n
	for _, tt := range []struct {
		name string
		val  any
		want string
	}{
		{name: "int", val: 42, want: "42"},
		{name: "pointer to int", val: &n, want: "42"},
		{name: "float", val: 1.5, want: "1.5"},
		{name: "bool", val: true, want: "true"},
		{name: "struct", val: struct{ A int }{1}, want: "{1}"},
		{name: "map", val: map[string]int{"a": 1}, want: "map[a:1]"},
		{name: "slice", val: []int{1, 2}, want: "[1 2]"},
		// fmt.Stringer is used like by fmt.Sprint
		{name: "Stringer", val: time.Second, want: "1s"},
		{name: "pointer to Stringer", val: &uu.IDNil, want: uu.IDNil.String()},
		// Only one pointer level is dereferenced, the inner pointer
		// is formatted by fmt.Sprint like any other non-nil pointer
		{name: "pointer to pointer", val: &pn, want: fmt.Sprint(pn)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Interface(tt.val))
		})
	}
}

// ptrStringer implements fmt.Stringer with a pointer receiver.
type ptrStringer struct{ n int }

func (p *ptrStringer) String() string { return fmt.Sprintf("ptrStringer %d", p.n) }

// ptrFormatter implements fmt.Formatter with a pointer receiver.
type ptrFormatter struct{ n int }

func (p *ptrFormatter) Format(f fmt.State, verb rune) { fmt.Fprintf(f, "ptrFormatter %d", p.n) }

// color is of string kind and implements fmt.Stringer with a value receiver.
type color string

func (c color) String() string { return "color " + string(c) }

// errCode is of string kind and implements error with a value receiver.
type errCode string

func (e errCode) Error() string { return "error code " + string(e) }

// A type's own String or Error method is how it wants to be shown to
// humans, so Interface must use it like fmt.Sprint does, also if the
// method has a pointer receiver, like the one of the error returned by
// errors.New, or if the type is of string kind. Formatting the fields
// of the dereferenced value instead, like "{boom}", hides the message.
// A value still returns the same string as a pointer to it,
// and the nil handling stays: a nil pointer returns "" and its method
// is not called.
func TestInterface_FormattingMethods(t *testing.T) {
	c := color("red")
	for _, tt := range []struct {
		name string
		val  any
		want string
	}{
		{name: "errors.New", val: errors.New("boom"), want: "boom"},
		{name: "wrapped error", val: fmt.Errorf("outer: %w", errors.New("inner")), want: "outer: inner"},
		{name: "pointer receiver Stringer", val: &ptrStringer{n: 1}, want: "ptrStringer 1"},
		{name: "pointer receiver Stringer by value", val: ptrStringer{n: 2}, want: "ptrStringer 2"},
		{name: "pointer receiver Formatter", val: &ptrFormatter{n: 3}, want: "ptrFormatter 3"},
		{name: "string kind value receiver Stringer", val: color("red"), want: "color red"},
		{name: "pointer to string kind value receiver Stringer", val: &c, want: "color red"},
		{name: "string kind value receiver error", val: errCode("E1"), want: "error code E1"},
		{name: "nil pointer receiver Stringer", val: (*ptrStringer)(nil), want: ""},
		{name: "nil pointer to string kind Stringer", val: (*color)(nil), want: ""},
		// Values without formatting methods are unchanged
		{name: "int", val: 42, want: "42"},
		{name: "string", val: "plain", want: "plain"},
		{name: "struct", val: struct{ A int }{1}, want: "{1}"},
		{name: "pointer to struct", val: &struct{ A int }{1}, want: "{1}"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				assert.Equal(t, tt.want, Interface(tt.val))
			})
		})
	}
}
