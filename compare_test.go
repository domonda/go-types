package types

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCompareReflectValue(t *testing.T) {
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
		// Bool: false < true
		{reflect.ValueOf(false), reflect.ValueOf(true), -1},
		{reflect.ValueOf(true), reflect.ValueOf(true), 0},
		{reflect.ValueOf(true), reflect.ValueOf(false), +1},
		// Complex: by real, then imaginary
		{reflect.ValueOf(1 + 2i), reflect.ValueOf(1 + 3i), -1},
		{reflect.ValueOf(1 + 2i), reflect.ValueOf(1 + 2i), 0},
		{reflect.ValueOf(2 + 0i), reflect.ValueOf(1 + 9i), +1},
		// Array: lexicographic
		{reflect.ValueOf([2]int{1, 2}), reflect.ValueOf([2]int{1, 3}), -1},
		{reflect.ValueOf([2]int{1, 2}), reflect.ValueOf([2]int{1, 2}), 0},
		{reflect.ValueOf([2]int{2, 0}), reflect.ValueOf([2]int{1, 9}), +1},
		// Struct: lexicographic by field
		{reflect.ValueOf(struct{ X, Y int }{1, 2}), reflect.ValueOf(struct{ X, Y int }{1, 3}), -1},
		{reflect.ValueOf(struct{ X, Y int }{1, 2}), reflect.ValueOf(struct{ X, Y int }{1, 2}), 0},
		{reflect.ValueOf(struct{ X, Y int }{2, 0}), reflect.ValueOf(struct{ X, Y int }{1, 9}), +1},
		// Slice: lexicographic, shorter first on tie
		{reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{1, 3}), -1},
		{reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{1, 2}), 0},
		{reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{1, 2, 3}), -1},
		{reflect.ValueOf([]int{1, 2, 3}), reflect.ValueOf([]int{1, 2}), +1},
		// []byte uses bytes.Compare
		{reflect.ValueOf([]byte("abc")), reflect.ValueOf([]byte("abd")), -1},
		{reflect.ValueOf([]byte("abc")), reflect.ValueOf([]byte("abc")), 0},
		{reflect.ValueOf([]byte("abd")), reflect.ValueOf([]byte("abc")), +1},
		{reflect.ValueOf([]byte("ab")), reflect.ValueOf([]byte("abc")), -1},
		// Mismatched types compare as equal instead of panicking
		{reflect.ValueOf(1), reflect.ValueOf("1"), 0},
		{reflect.ValueOf(int32(1)), reflect.ValueOf(int64(1)), 0},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%#v_%#v", tt.a.Interface(), tt.b.Interface()), func(t *testing.T) {
			if got := CompareReflectValue(tt.a, tt.b); got != tt.want {
				t.Errorf("CompareReflectValue(%#v, %#v) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
