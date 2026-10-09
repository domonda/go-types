package types

import "testing"

// TestPtr checks that Ptr points to a copy of the value,
// so later changes of the source variable don't leak through the pointer
// and every call returns a distinct pointer.
func TestPtr(t *testing.T) {
	v := 1
	p := Ptr(v)
	v = 2
	if *p != 1 {
		t.Errorf("*Ptr(v) = %d after changing v, want the copied 1", *p)
	}
	if Ptr(v) == Ptr(v) {
		t.Error("Ptr returned the same pointer for two calls, want distinct copies")
	}
}

// TestFromPtr checks that FromPtr safely dereferences nil to the zero value.
func TestFromPtr(t *testing.T) {
	if got := FromPtr[string](nil); got != "" {
		t.Errorf("FromPtr(nil) = %q, want zero value", got)
	}
	type pair struct{ A, B int }
	if got := FromPtr[pair](nil); got != (pair{}) {
		t.Errorf("FromPtr(nil) = %#v, want zero value", got)
	}
	if got := FromPtr(Ptr("x")); got != "x" {
		t.Errorf("FromPtr(&\"x\") = %q, want %q", got, "x")
	}
}

// TestFromPtrOr checks that the default is only used for a nil pointer,
// a non-nil pointer to the zero value must still return the zero value.
func TestFromPtrOr(t *testing.T) {
	if got := FromPtrOr(nil, "default"); got != "default" {
		t.Errorf("FromPtrOr(nil, default) = %q, want %q", got, "default")
	}
	if got := FromPtrOr(Ptr("x"), "default"); got != "x" {
		t.Errorf("FromPtrOr(&\"x\", default) = %q, want %q", got, "x")
	}
	if got := FromPtrOr(Ptr(""), "default"); got != "" {
		t.Errorf("FromPtrOr(&\"\", default) = %q, want the pointed to empty string", got)
	}
	if got := FromPtrOr(Ptr(0), 42); got != 0 {
		t.Errorf("FromPtrOr(&0, 42) = %d, want the pointed to 0", got)
	}
}
