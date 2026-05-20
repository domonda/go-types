package queue

import (
	"sync"
	"testing"
	"time"
)

// testChanLen is the consumer-channel capacity used throughout the tests.
const testChanLen = 32

func TestQueueFIFO(t *testing.T) {
	q := New[any](ChanLen(testChanLen))
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
	// Force the queue to buffer beyond the channel capacity without anyone
	// reading, so the ring buffer must grow.
	q := New[int](ChanLen(testChanLen))
	defer q.Close()

	n := testChanLen*4 + 7 // well past both the channel capacity and the initial buffer
	items := make([]int, n)
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

func TestQueueUnbufferedChannel(t *testing.T) {
	// ChanLen(0) is valid: the pump hands items off synchronously.
	q := New[int](ChanLen(0))
	defer q.Close()

	q.Add(1, 2, 3)
	for _, want := range []int{1, 2, 3} {
		select {
		case got := <-q.Next():
			if got != want {
				t.Errorf("Next() = %v, want %v", got, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for %v", want)
		}
	}
}

func TestQueueCloseWithActiveConsumer(t *testing.T) {
	q := New[string](ChanLen(testChanLen))

	var seen []string
	var wg sync.WaitGroup
	wg.Go(func() {
		for v := range q.Next() {
			seen = append(seen, v)
		}
	})

	q.Add("a", "b", "c")
	// Give the active consumer time to receive everything before closing.
	time.Sleep(50 * time.Millisecond)
	q.Close()
	wg.Wait()

	if len(seen) != 3 || seen[0] != "a" || seen[2] != "c" {
		t.Errorf("active consumer saw %v, want [a b c]", seen)
	}
}

func TestQueueCloseDiscards(t *testing.T) {
	q := New[int](ChanLen(testChanLen))

	q.Add(1, 2, 3, 4, 5)
	time.Sleep(50 * time.Millisecond) // let the pump stage the items into the channel
	q.Close()                         // synchronous: returns once the channel is drained and closed

	// Close discards everything still queued; Next() yields nothing.
	var seen []int
	for v := range q.Next() {
		seen = append(seen, v)
	}
	if len(seen) != 0 {
		t.Errorf("after Close, received %v, want everything discarded", seen)
	}
}

func TestQueueLenCap(t *testing.T) {
	q := New[int](ChanLen(testChanLen))
	defer q.Close()

	if q.Len() != 0 {
		t.Errorf("empty Len() = %d, want 0", q.Len())
	}
	if q.Cap() < testChanLen {
		t.Errorf("Cap() = %d, want >= %d", q.Cap(), testChanLen)
	}
}

func TestQueueCustomBufferSize(t *testing.T) {
	q := New[int](ChanLen(testChanLen), InitialBufferSize(64))
	defer q.Close()

	if got, want := q.Cap(), testChanLen+64; got != want {
		t.Errorf("Cap() = %d, want %d (ChanLen %d + InitialBufferSize 64)", got, want, testChanLen)
	}
}

func TestQueueNewPanicsOnInvalidConfig(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
	}{
		{"negative ChanLen", ChanLen(-1)},
		{"zero InitialBufferSize", InitialBufferSize(0)},
		{"negative InitialBufferSize", InitialBufferSize(-1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("New(%v) did not panic", tc.opt)
				}
			}()
			New[int](tc.opt)
		})
	}
}
