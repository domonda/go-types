package types

import (
	"fmt"
	htemplate "html/template"
	"reflect"
	"testing"
	ttemplate "text/template"
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

// TestCompareReflectValue_Interface checks that interface values are totally
// ordered even when their dynamic types differ. DeepValidate sorts the keys
// of a map[any]T with CompareReflectValue, and returning 0 for differently
// typed keys made the sort order, and so the order of the returned errors,
// depend on the random map iteration order.
func TestCompareReflectValue_Interface(t *testing.T) {
	// iface returns v as a reflect.Value of kind Interface,
	// like the keys of a map[any]T
	iface := func(v any) reflect.Value { return reflect.ValueOf(&v).Elem() }

	tests := []struct {
		name string
		a, b reflect.Value
		want int
	}{
		{"nil before non-nil", iface(nil), iface(1), -1},
		{"nil equals nil", iface(nil), iface(nil), 0},
		{"non-nil after nil", iface(1), iface(nil), +1},
		{"same dynamic type by value", iface(1), iface(2), -1},
		{"same dynamic type equal", iface("a"), iface("a"), 0},
		{"different dynamic types by type name", iface(2), iface("1"), -1}, // "int" < "string"
		{"different dynamic types by type name reversed", iface("1"), iface(2), +1},
		{"different dynamic types are never equal", iface(int32(1)), iface(int64(1)), -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareReflectValue(tt.a, tt.b); got != tt.want {
				t.Errorf("CompareReflectValue(%#v, %#v) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}

	t.Run("different types with the same name", func(t *testing.T) {
		a := iface((*ttemplate.Template)(nil))
		b := iface((*htemplate.Template)(nil))
		if a.Elem().Type().String() != b.Elem().Type().String() {
			t.Fatalf("test needs types with the same name, got %s and %s", a.Elem().Type(), b.Elem().Type())
		}
		ab, ba := CompareReflectValue(a, b), CompareReflectValue(b, a)
		if ab == 0 || ab != -ba {
			t.Errorf("CompareReflectValue = %d and reversed %d, want a non-zero antisymmetric order", ab, ba)
		}
	})
}

// TestCompareReflectValue_Invalid checks that the zero reflect.Value,
// which has no type, is ordered before every valid value instead of
// panicking, like a nil interface value is ordered before non-nil ones.
// The function is documented as safe to use as compare function
// for sorting arbitrary reflect.Values, and reflect.ValueOf(nil)
// or the Elem of a nil interface return the zero reflect.Value.
func TestCompareReflectValue_Invalid(t *testing.T) {
	var invalid reflect.Value
	for _, valid := range []reflect.Value{reflect.ValueOf(0), reflect.ValueOf(""), reflect.ValueOf((*int)(nil))} {
		if got := CompareReflectValue(invalid, valid); got != -1 {
			t.Errorf("CompareReflectValue(invalid, %#v) = %d, want -1", valid, got)
		}
		if got := CompareReflectValue(valid, invalid); got != +1 {
			t.Errorf("CompareReflectValue(%#v, invalid) = %d, want +1", valid, got)
		}
	}
	if got := CompareReflectValue(invalid, reflect.ValueOf(nil)); got != 0 {
		t.Errorf("CompareReflectValue(invalid, invalid) = %d, want 0", got)
	}
}

// TestCompareReflectValue_Chan checks that channels are ordered by address
// as documented instead of all comparing as equal.
func TestCompareReflectValue_Chan(t *testing.T) {
	a, b := reflect.ValueOf(make(chan int)), reflect.ValueOf(make(chan int))
	if got := CompareReflectValue(a, a); got != 0 {
		t.Errorf("CompareReflectValue(a, a) = %d, want 0", got)
	}
	ab, ba := CompareReflectValue(a, b), CompareReflectValue(b, a)
	if ab == 0 || ab != -ba {
		t.Errorf("CompareReflectValue = %d and reversed %d, want a non-zero antisymmetric order", ab, ba)
	}
}
