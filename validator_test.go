package types

import (
	"fmt"
	"reflect"
	"testing"
)

func TestReflectCompare(t *testing.T) {
	tests := []struct {
		a    reflect.Value
		b    reflect.Value
		want int
	}{
		{reflect.ValueOf(0), reflect.ValueOf(1), -1},
		{reflect.ValueOf(0), reflect.ValueOf(0), 0},
		{reflect.ValueOf(1), reflect.ValueOf(0), +1},
		{reflect.ValueOf(0.0), reflect.ValueOf(1.0), -1},
		{reflect.ValueOf(0.0), reflect.ValueOf(0.0), 0},
		{reflect.ValueOf(1.0), reflect.ValueOf(0.0), +1},
		{reflect.ValueOf("0"), reflect.ValueOf("1"), -1},
		{reflect.ValueOf("0"), reflect.ValueOf("0"), 0},
		{reflect.ValueOf("1"), reflect.ValueOf("0"), +1},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%#v_%#v", tt.a.Interface(), tt.b.Interface()), func(t *testing.T) {
			if got := ReflectCompare(tt.a, tt.b); got != tt.want {
				t.Errorf("ReflectCompare(%#v, %#v) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
