package nullable

import (
	"reflect"
	"testing"
	"time"
)

func TestReflectIsNull(t *testing.T) {
	tests := []struct {
		name string
		v    reflect.Value
		want bool
	}{
		{"zero reflect.Value", reflect.Value{}, true},
		{"nil interface", reflect.ValueOf(nil), true},
		{"nil int ptr", reflect.ValueOf((*int)(nil)), true},
		{"nil Nullable", reflect.ValueOf(Nullable(nil)), true},
		{"nil Zeroable", reflect.ValueOf(Zeroable(nil)), true},
		{"nil IntArray", reflect.ValueOf(IntArray(nil)), true},
		{"nil func", reflect.ValueOf((func())(nil)), true},
		{"zero time.Time", reflect.ValueOf(time.Time{}), true},
		{"zero time.Time ptr", reflect.ValueOf(new(time.Time)), true},
		{"nil time.Time ptr", reflect.ValueOf((*time.Time)(nil)), true},
		{"TimeNull", reflect.ValueOf(TimeNull), true},

		{"0", reflect.ValueOf(0), false},
		{"0 ptr", reflect.ValueOf(new(int)), false},
		{"empty IntArray", reflect.ValueOf(IntArray{}), false},
		{"empty string", reflect.ValueOf(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReflectIsNull(tt.v); got != tt.want {
				t.Errorf("ReflectIsNull() = %v, want %v", got, tt.want)
			}
		})
	}
}
