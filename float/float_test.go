package float

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidAndHasSign(t *testing.T) {
	tests := []struct {
		f    float64
		sign int
		want bool
	}{
		{f: 0, sign: 0, want: true},
		{f: 0, sign: +1, want: true},
		{f: 0, sign: -1, want: false},

		{f: -1, sign: 0, want: true},
		{f: -1, sign: +1, want: false},
		{f: -1, sign: -1, want: true},

		{f: +1, sign: 0, want: true},
		{f: +1, sign: +1, want: true},
		{f: +1, sign: -1, want: false},

		{f: math.NaN(), sign: 0, want: false},
		{f: math.NaN(), sign: +1, want: false},
		{f: math.NaN(), sign: -1, want: false},
		{f: math.Inf(0), sign: 0, want: false},
		{f: math.Inf(+1), sign: +1, want: false},
		{f: math.Inf(-1), sign: -1, want: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%#v_%#v", tt.f, tt.sign), func(t *testing.T) {
			if got := ValidAndHasSign(tt.f, tt.sign); got != tt.want {
				t.Errorf("ValidAndHasSign(%#v, %#v) = %#v, want %#v", tt.f, tt.sign, got, tt.want)
			}
		})
	}
}

// TestRoundToDecimals protects the rounding mode: values exactly halfway
// round away from zero (not banker's rounding), so amounts like 2.5
// become 3 and -2.5 becomes -3, symmetric around zero.
// Negative decimals round to tens, hundreds, etc.
func TestRoundToDecimals(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		decimals int
		want     float64
	}{
		{name: "zero", f: 0, decimals: 2, want: 0},
		{name: "round down", f: 1.2345678, decimals: 2, want: 1.23},
		{name: "round up", f: 1.2345678, decimals: 3, want: 1.235},
		{name: "negative round toward zero", f: -1.2345678, decimals: 2, want: -1.23},
		{name: "already rounded", f: 1.5, decimals: 2, want: 1.5},
		{name: "half away from zero positive", f: 2.5, decimals: 0, want: 3},
		{name: "half away from zero negative", f: -2.5, decimals: 0, want: -3},
		{name: "exactly representable half", f: 0.125, decimals: 2, want: 0.13},
		{name: "negative decimals round to hundreds", f: 1234.5, decimals: -2, want: 1200},
		{name: "negative decimals half up", f: 1250, decimals: -2, want: 1300},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, RoundToDecimals(tt.f, tt.decimals))
		})
	}

	t.Run("float32 keeps its type", func(t *testing.T) {
		got := RoundToDecimals(float32(1.2345678), 2)
		assert.Equal(t, float32(1.23), got)
	})

	t.Run("NaN and Inf pass through", func(t *testing.T) {
		assert.True(t, math.IsNaN(RoundToDecimals(math.NaN(), 2)))
		assert.True(t, math.IsInf(RoundToDecimals(math.Inf(1), 2), 1))
		assert.True(t, math.IsInf(RoundToDecimals(math.Inf(-1), 2), -1))
	})
}

// TestRoundToDecimals_shortestDecimalTies protects the rounding contract
// shared with Format and all money rounding: the value that is rounded is
// the shortest decimal representation of the float, the decimal a human
// typed or a JSON number carried, not its binary approximation.
// 1.005 is stored as 1.00499999999999989342, so rounding the binary value
// yields 1 while every money method and Format yield 1.01 for it.
// Rounding the same number differently depending on the function used
// makes sums and their formatted strings disagree by a cent.
func TestRoundToDecimals_shortestDecimalTies(t *testing.T) {
	tests := []struct {
		f        float64
		decimals int
		want     float64
	}{
		{f: 1.005, decimals: 2, want: 1.01},
		{f: -1.005, decimals: 2, want: -1.01},
		{f: 0.145, decimals: 2, want: 0.15},
		{f: 2.675, decimals: 2, want: 2.68},
		{f: -2.675, decimals: 2, want: -2.68},
		{f: 8.345, decimals: 2, want: 8.35},
		{f: 1234.565, decimals: 2, want: 1234.57},
		{f: 1.0005, decimals: 3, want: 1.001},
		{f: 1.45, decimals: 1, want: 1.5},
		{f: 0.5, decimals: 0, want: 1},
		{f: 9.995, decimals: 2, want: 10},
		{f: 0.144999, decimals: 2, want: 0.14}, // not a tie
		{f: 0.1 + 0.2, decimals: 2, want: 0.3},
		{f: 15, decimals: -1, want: 20},
		{f: -25, decimals: -1, want: -30},
		{f: 4.9, decimals: -1, want: 0},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, RoundToDecimals(tt.f, tt.decimals), "RoundToDecimals(%v, %d)", tt.f, tt.decimals)
	}

	// float32(1.005) is 1.00499999523162841797 in binary,
	// its shortest float32 decimal is still 1.005
	assert.Equal(t, float32(1.01), RoundToDecimals(float32(1.005), 2))
	assert.Equal(t, float32(2.68), RoundToDecimals(float32(2.675), 2))

	// Negative zero keeps its sign like math.Round does
	assert.True(t, math.Signbit(RoundToDecimals(-0.001, 2)), "-0.001 rounds to -0")
}

// TestRoundToDecimals_overflow protects that a huge number of decimals or a
// huge value never turns a finite float into NaN or Inf. Scaling by 10^decimals
// overflowed: 10^400 is Inf and f*Inf/Inf is NaN, and 1e300*10^10 is Inf.
// A value that has no more decimals than requested is returned unchanged,
// and rounding to a power of ten beyond the value yields zero.
func TestRoundToDecimals_overflow(t *testing.T) {
	assert.Equal(t, 1.5, RoundToDecimals(1.5, 400))
	assert.Equal(t, 1.5, RoundToDecimals(1.5, 309))
	assert.Equal(t, 1e300, RoundToDecimals(1e300, 10))
	assert.Equal(t, -1e300, RoundToDecimals(-1e300, 10))
	assert.Equal(t, math.MaxFloat64, RoundToDecimals(math.MaxFloat64, 2))
	assert.Equal(t, math.SmallestNonzeroFloat64, RoundToDecimals(math.SmallestNonzeroFloat64, 400))
	assert.Equal(t, 1.2345e-300, RoundToDecimals(1.2345e-300, 310))
	assert.Equal(t, 0.0, RoundToDecimals(1.5, -400))
	assert.Equal(t, 0.0, RoundToDecimals(1e300, -400))
	assert.Equal(t, float32(math.MaxFloat32), RoundToDecimals(float32(math.MaxFloat32), 2))
	assert.Equal(t, float32(1.5), RoundToDecimals(float32(1.5), 400))

	// Only a rounded value that is itself beyond the float range is ±Inf,
	// like any other float64 operation overflowing:
	// 1.7976931348623157e308 rounded to a multiple of 10^307 is 1.8e308
	assert.True(t, math.IsInf(RoundToDecimals(math.MaxFloat64, -307), 1))
	assert.True(t, math.IsInf(RoundToDecimals(-math.MaxFloat64, -307), -1))
}

// TestDerefOr protects the nil-safety contract: a nil pointer yields the
// default, while a non-nil pointer to zero must yield zero (not the default),
// because zero is a legitimate value distinct from "not set".
func TestDerefOr(t *testing.T) {
	assert.Equal(t, 7.5, DerefOr(nil, 7.5), "nil pointer returns default")

	zero := 0.0
	assert.Equal(t, 0.0, DerefOr(&zero, 7.5), "pointer to zero must not be replaced by default")

	val := -1.25
	assert.Equal(t, -1.25, DerefOr(&val, 7.5))

	f32 := float32(3.5)
	assert.Equal(t, float32(3.5), DerefOr(&f32, 1))
	assert.Equal(t, float32(1), DerefOr[float32](nil, 1))
}

func TestValid(t *testing.T) {
	assert.True(t, Valid(0.0))
	assert.True(t, Valid(-1.5))
	assert.True(t, Valid(math.MaxFloat64))
	assert.True(t, Valid(float32(1)))
	assert.False(t, Valid(math.NaN()))
	assert.False(t, Valid(math.Inf(1)))
	assert.False(t, Valid(math.Inf(-1)))
	assert.False(t, Valid(float32(math.Inf(1))))
}
