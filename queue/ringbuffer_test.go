package queue

import "testing"

// TestRingBufferGrowWhileWrapped protects the growth copy logic: when the
// buffer is full and its content wraps around the end of the backing slice,
// growing must preserve FIFO order. A wrong copy offset would reorder or
// lose queued items.
func TestRingBufferGrowWhileWrapped(t *testing.T) {
	b := ringBuffer[int]{items: make([]int, 4)}
	for i := range 4 {
		b.push(i)
	}
	// Remove two from the front and push two more so the content
	// wraps around: items = [4 5 2 3], first = 2.
	if got := b.shift(); got != 0 {
		t.Fatalf("shift() = %d, want 0", got)
	}
	if got := b.shift(); got != 1 {
		t.Fatalf("shift() = %d, want 1", got)
	}
	b.push(4)
	b.push(5)
	if b.first == 0 {
		t.Fatal("test setup: buffer content is expected to wrap around")
	}

	// Buffer is full and wrapped, this push has to grow it
	b.push(6)
	if len(b.items) != 8 {
		t.Fatalf("len(items) = %d after growth, want 8", len(b.items))
	}

	for want := 2; want <= 6; want++ {
		if b.isEmpty() {
			t.Fatalf("buffer empty, want %d", want)
		}
		if got := b.shift(); got != want {
			t.Fatalf("shift() = %d, want %d", got, want)
		}
	}
	if !b.isEmpty() {
		t.Errorf("buffer not empty after shifting all items, count = %d", b.count)
	}
}

// TestRingBufferShiftReleasesItem protects against memory retention: a queue
// is long-lived and its buffer only grows, so a slot that still references a
// delivered item keeps that item reachable until the slot happens to be
// overwritten, which may never happen after a burst.
func TestRingBufferShiftReleasesItem(t *testing.T) {
	b := ringBuffer[*int]{items: make([]*int, 4)}
	for i := range 3 {
		b.push(&i)
	}
	for range 3 {
		if b.shift() == nil {
			t.Fatal("shift() = nil, want pushed item")
		}
	}
	for i, item := range b.items {
		if item != nil {
			t.Errorf("items[%d] still references the shifted item %d", i, *item)
		}
	}
}

// TestRingBufferShiftEmptyPanics protects that a pump bug shifting an empty
// buffer fails loudly instead of silently delivering a zero value.
func TestRingBufferShiftEmptyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("shift() on empty buffer did not panic")
		}
	}()
	b := ringBuffer[int]{items: make([]int, 2)}
	b.shift()
}
