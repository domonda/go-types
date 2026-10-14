package internal

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFastRound pins when the fast float64 rounding of the float and money
// packages may skip their exact fallback. A false ok only costs time, but a
// true ok next to a half-integer returns a rounding that differs from the
// exact decimal one: 1.005·100 is 100.49999999999999 in float64 and rounds
// to 100, while the decimal 1.005 that the caller rounds is the tie 100.5,
// which rounds to 101.
func TestFastRound(t *testing.T) {
	// A variable, because the constant expression 1.005 * 100 is the exact 100.5
	f := 1.005
	tests := []struct {
		name        string
		x           float64
		wantRounded float64
		wantOK      bool
	}{
		{name: "zero", x: 0, wantRounded: 0, wantOK: true},
		{name: "below half", x: 2.4, wantRounded: 2, wantOK: true},
		{name: "above half", x: 2.6, wantRounded: 3, wantOK: true},
		{name: "negative above half", x: -2.6, wantRounded: -3, wantOK: true},
		{name: "integer", x: 1234, wantRounded: 1234, wantOK: true},
		{name: "largest power of two that can be ok", x: 0x1p47, wantRounded: 0x1p47, wantOK: true},
		// The ties are left to the exact rounding of the caller,
		// because the exact value may lie on either side of them
		{name: "tie", x: 2.5, wantOK: false},
		{name: "negative tie", x: -2.5, wantOK: false},
		{name: "float below the decimal tie of 1.005·100", x: f * 100, wantOK: false},
		// From 2^48 on the bound |x|·2^-49 reaches 1/2,
		// so no distance to a half-integer exceeds it
		{name: "2^48 is too large", x: 0x1p48, wantOK: false},
		{name: "NaN", x: math.NaN(), wantOK: false},
		{name: "+Inf", x: math.Inf(1), wantOK: false},
		{name: "-Inf", x: math.Inf(-1), wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rounded, ok := FastRound(tt.x)
			assert.Equal(t, tt.wantOK, ok, "ok of FastRound(%v)", tt.x)
			if tt.wantOK {
				assert.Equal(t, tt.wantRounded, rounded, "FastRound(%v)", tt.x)
			}
		})
	}
}
