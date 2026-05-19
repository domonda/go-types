package queue

import (
	"sync"
	"testing"
	"time"
)

func TestQueueFIFO(t *testing.T) {
	q := New()
	defer q.Close()

	q.Add(1, 2, 3, "four", 5.0)

	want := []any{1, 2, 3, "four", 5.0}
	for _, w := range want {
		select {
		case got := <-q.Next():
			if got != w {
				t.Errorf("Next() = %v, want %v", got, w)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for %v", w)
		}
	}
}

func TestQueueBufferGrowth(t *testing.T) {
	// Force the queue to buffer beyond the channel capacity (default ChanLen=32)
	// without anyone reading, so the ring buffer must grow.
	q := New()
	defer q.Close()

	n := ChanLen*4 + 7 // well past both ChanLen and InitialBufferSize
	items := make([]any, n)
	for i := range items {
		items[i] = i
	}
	q.Add(items...)

	// All items must come out in order.
	for i := range n {
		select {
		case got := <-q.Next():
			if got != i {
				t.Errorf("position %d: got %v, want %d", i, got, i)
				return
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out at position %d (got %d of %d)", i, i, n)
		}
	}
}

func TestQueueCloseDrainsChannel(t *testing.T) {
	q := New()

	var seen []any
	var wg sync.WaitGroup
	wg.Go(func() {
		for v := range q.Next() {
			seen = append(seen, v)
		}
	})

	q.Add("a", "b", "c")
	// Give the pump a moment to deliver before closing.
	time.Sleep(50 * time.Millisecond)
	q.Close()
	wg.Wait()

	if len(seen) != 3 || seen[0] != "a" || seen[2] != "c" {
		t.Errorf("after Close, seen = %v, want [a b c]", seen)
	}
}

func TestQueueLenCap(t *testing.T) {
	q := New()
	defer q.Close()

	if q.Len() != 0 {
		t.Errorf("empty Len() = %d, want 0", q.Len())
	}
	if q.Cap() < ChanLen {
		t.Errorf("Cap() = %d, want >= %d", q.Cap(), ChanLen)
	}
}
