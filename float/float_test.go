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
// round away from zero (math.Round semantics, not banker's rounding), so
// amounts like 2.5 become 3 and -2.5 becomes -3, symmetric around zero.
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
