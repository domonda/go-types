package set

import (
	"testing"
)

func TestNewAndContains(t *testing.T) {
	s := New("a", "b", "c")
	if len(s) != 3 {
		t.Errorf("len(New(\"a\",\"b\",\"c\")) = %d, want 3", len(s))
	}
	if !Contains(s, "a") {
		t.Error("Contains(s, \"a\") = false, want true")
	}
	if Contains(s, "z") {
		t.Error("Contains(s, \"z\") = true, want false")
	}

	// Duplicate values collapse.
	s2 := New(1, 1, 2, 2, 3)
	if len(s2) != 3 {
		t.Errorf("len(New(1,1,2,2,3)) = %d, want 3", len(s2))
	}
}

func TestAdd(t *testing.T) {
	// nil set -> new set
	var s map[int]struct{}
	s = Add(s, 1, 2, 3)
	if len(s) != 3 {
		t.Errorf("Add(nil, 1,2,3) len = %d, want 3", len(s))
	}

	// add to existing
	s = Add(s, 3, 4)
	if !Contains(s, 4) || len(s) != 4 {
		t.Errorf("Add(s, 3, 4) len = %d (want 4) contains 4 = %v", len(s), Contains(s, 4))
	}
}

func TestContainsAllAny(t *testing.T) {
	s := New("a", "b", "c")

	if !ContainsAll(s, "a", "b") {
		t.Error("ContainsAll(s, a, b) = false, want true")
	}
	if ContainsAll(s, "a", "z") {
		t.Error("ContainsAll(s, a, z) = true, want false")
	}

	if !ContainsAny(s, "z", "a") {
		t.Error("ContainsAny(s, z, a) = false, want true")
	}
	if ContainsAny(s, "x", "y", "z") {
		t.Error("ContainsAny(s, x, y, z) = true, want false")
	}
}

func TestContainsOther(t *testing.T) {
	s := New(1, 2, 3, 4)
	subset := New(2, 3)
	disjoint := New(99)
	partial := New(3, 99)

	if !ContainsAllOther(s, subset) {
		t.Error("ContainsAllOther(s, subset) = false, want true")
	}
	if ContainsAllOther(s, partial) {
		t.Error("ContainsAllOther(s, partial) = true, want false")
	}

	if !ContainsAnyOther(s, partial) {
		t.Error("ContainsAnyOther(s, partial) = false, want true")
	}
	if ContainsAnyOther(s, disjoint) {
		t.Error("ContainsAnyOther(s, disjoint) = true, want false")
	}
}
