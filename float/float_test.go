package float

import (
	"fmt"
	"math"
	"testing"
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
