package stringfrom

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
	// Only the unambiguous cases are tested here,
	// see the bug report about non-pointer and untyped nil values.
	s := "Grüße"
	assert.Equal(t, "Grüße", Interface(&s))
	assert.Equal(t, "", Interface((*string)(nil)), "typed nil pointer")
}
