package types

import "testing"

// TestSliceContainsAll checks the subset semantics of SliceContainsAll:
// the order of elements doesn't matter and an empty inner list is
// vacuously contained in any slice, including a nil one.
func TestSliceContainsAll(t *testing.T) {
	tests := []struct {
		name  string
		outer []string
		inner []string
		want  bool
	}{
		{name: "all present", outer: []string{"a", "b", "c"}, inner: []string{"a", "c"}, want: true},
		{name: "all present in different order", outer: []string{"a", "b", "c"}, inner: []string{"c", "b", "a"}, want: true},
		{name: "one missing", outer: []string{"a", "b", "c"}, inner: []string{"a", "x"}, want: false},
		{name: "none present", outer: []string{"a", "b"}, inner: []string{"x"}, want: false},
		{name: "outer shorter than inner", outer: []string{"a"}, inner: []string{"a", "b"}, want: false},
		{name: "empty inner", outer: []string{"a"}, inner: nil, want: true},
		{name: "nil outer and empty inner", outer: nil, inner: nil, want: true},
		{name: "nil outer", outer: nil, inner: []string{"a"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SliceContainsAll(tt.outer, tt.inner...); got != tt.want {
				t.Errorf("SliceContainsAll(%v, %v) = %t, want %t", tt.outer, tt.inner, got, tt.want)
			}
		})
	}
}

// TestSliceContainsAny checks that a single match is enough and that
// an empty inner list matches nothing.
func TestSliceContainsAny(t *testing.T) {
	tests := []struct {
		name  string
		outer []int
		inner []int
		want  bool
	}{
		{name: "one present", outer: []int{1, 2, 3}, inner: []int{9, 3}, want: true},
		{name: "all present", outer: []int{1, 2, 3}, inner: []int{1, 2}, want: true},
		{name: "none present", outer: []int{1, 2, 3}, inner: []int{4, 5}, want: false},
		{name: "empty inner", outer: []int{1}, inner: nil, want: false},
		{name: "nil outer", outer: nil, inner: []int{1}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SliceContainsAny(tt.outer, tt.inner...); got != tt.want {
				t.Errorf("SliceContainsAny(%v, %v) = %t, want %t", tt.outer, tt.inner, got, tt.want)
			}
		})
	}
}
