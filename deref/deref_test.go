package deref

import (
	"testing"
	"time"
)

func TestDeref(t *testing.T) {
	b := true
	if got := Bool(&b, false); got != true {
		t.Errorf("Bool(&true, false) = %v, want true", got)
	}
	if got := Bool(nil, true); got != true {
		t.Errorf("Bool(nil, true) = %v, want true", got)
	}

	s := "hi"
	if got := String(&s, "default"); got != "hi" {
		t.Errorf("String(&\"hi\", \"default\") = %q, want %q", got, "hi")
	}
	if got := String(nil, "default"); got != "default" {
		t.Errorf("String(nil, \"default\") = %q, want %q", got, "default")
	}

	i := 42
	if got := Int(&i, -1); got != 42 {
		t.Errorf("Int(&42, -1) = %d, want 42", got)
	}
	if got := Int(nil, -1); got != -1 {
		t.Errorf("Int(nil, -1) = %d, want -1", got)
	}

	var u32 int32 = 7
	if got := Int32(&u32, 0); got != 7 {
		t.Errorf("Int32(&7, 0) = %d, want 7", got)
	}
	if got := Int32(nil, -3); got != -3 {
		t.Errorf("Int32(nil, -3) = %d, want -3", got)
	}

	var i64 int64 = 1234567890123
	if got := Int64(&i64, 0); got != i64 {
		t.Errorf("Int64(&v, 0) = %d, want %d", got, i64)
	}
	if got := Int64(nil, 99); got != 99 {
		t.Errorf("Int64(nil, 99) = %d, want 99", got)
	}

	var ui uint = 5
	if got := Uint(&ui, 0); got != 5 {
		t.Errorf("Uint(&5, 0) = %d, want 5", got)
	}
	if got := Uint(nil, 9); got != 9 {
		t.Errorf("Uint(nil, 9) = %d, want 9", got)
	}

	var u64 uint64 = 9876543210
	if got := Uint64(&u64, 0); got != u64 {
		t.Errorf("Uint64(&v, 0) = %d, want %d", got, u64)
	}
	if got := Uint64(nil, 7); got != 7 {
		t.Errorf("Uint64(nil, 7) = %d, want 7", got)
	}

	var f32 float32 = 3.14
	if got := Float32(&f32, 0); got != f32 {
		t.Errorf("Float32(&3.14, 0) = %v, want %v", got, f32)
	}
	if got := Float32(nil, 2.72); got != 2.72 {
		t.Errorf("Float32(nil, 2.72) = %v, want 2.72", got)
	}

	f64 := 2.71828
	if got := Float64(&f64, 0); got != f64 {
		t.Errorf("Float64(&2.71828, 0) = %v, want %v", got, f64)
	}
	if got := Float64(nil, 1.41); got != 1.41 {
		t.Errorf("Float64(nil, 1.41) = %v, want 1.41", got)
	}

	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	zero := time.Time{}
	if got := Time(&now, zero); !got.Equal(now) {
		t.Errorf("Time(&now, zero) = %v, want %v", got, now)
	}
	if got := Time(nil, now); !got.Equal(now) {
		t.Errorf("Time(nil, now) = %v, want %v", got, now)
	}
}
