package types

import "testing"

// TestFromPtr checks that FromPtr safely dereferences nil to the zero value.
func TestFromPtr(t *testing.T) {
	if got := FromPtr[string](nil); got != "" {
		t.Errorf("FromPtr(nil) = %q, want zero value", got)
	}
	type pair struct{ A, B int }
	if got := FromPtr[pair](nil); got != (pair{}) {
		t.Errorf("FromPtr(nil) = %#v, want zero value", got)
	}
	if got := FromPtr(new("x")); got != "x" {
		t.Errorf("FromPtr(&\"x\") = %q, want %q", got, "x")
	}
}

// TestFromPtrOr checks that the default is only used for a nil pointer,
// a non-nil pointer to the zero value must still return the zero value.
func TestFromPtrOr(t *testing.T) {
	if got := FromPtrOr(nil, "default"); got != "default" {
		t.Errorf("FromPtrOr(nil, default) = %q, want %q", got, "default")
	}
	if got := FromPtrOr(new("x"), "default"); got != "x" {
		t.Errorf("FromPtrOr(&\"x\", default) = %q, want %q", got, "x")
	}
	if got := FromPtrOr(new(""), "default"); got != "" {
		t.Errorf("FromPtrOr(&\"\", default) = %q, want the pointed to empty string", got)
	}
	if got := FromPtrOr(new(0), 42); got != 0 {
		t.Errorf("FromPtrOr(&0, 42) = %d, want the pointed to 0", got)
	}
}
