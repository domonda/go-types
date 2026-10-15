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

// waitFor polls cond until it returns true or fails the test after a second.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

// closeWithTimeout fails the test instead of hanging forever if Close deadlocks.
func closeWithTimeout(t *testing.T, q Queue[int]) {
	t.Helper()
	closed := make(chan struct{})
	go func() {
		q.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close did not return")
	}
}

// TestQueueLenCountsStagedItems protects that Len reports every item not yet
// received by a consumer, including those already staged in the consumer
// channel, not only those left in the internal ring-buffer.
func TestQueueLenCountsStagedItems(t *testing.T) {
	q := New[int](ChanLen(testChanLen))
	defer q.Close()

	q.Add(1, 2, 3, 4, 5)
	waitFor(t, "items staged in channel", func() bool { return len(q.Next()) == 5 })
	if got := q.Len(); got != 5 {
		t.Errorf("Len() = %d, want 5 with all items staged in the channel", got)
	}

	<-q.Next()
	<-q.Next()
	if got := q.Len(); got != 3 {
		t.Errorf("Len() = %d after receiving 2 of 5 items, want 3", got)
	}
}

// TestQueueAddNoItems protects that calling Add without items is a harmless
// no-op, as callers may forward an empty variadic slice.
func TestQueueAddNoItems(t *testing.T) {
	q := New[int](ChanLen(testChanLen))
	defer q.Close()

	q.Add()
	q.Add([]int{}...)
	if got := q.Len(); got != 0 {
		t.Errorf("Len() = %d after adding no items, want 0", got)
	}
	select {
	case v := <-q.Next():
		t.Errorf("received %v, want nothing after adding no items", v)
	case <-time.After(20 * time.Millisecond):
	}
}

// TestQueueAddAfterClose protects that Add after Close neither panics
// (no send on closed channel, no push into the dropped ring-buffer)
// nor delivers anything.
func TestQueueAddAfterClose(t *testing.T) {
	q := New[int](ChanLen(testChanLen))
	closeWithTimeout(t, q)

	q.Add(1, 2, 3)
	if got := q.Len(); got != 0 {
		t.Errorf("Len() = %d after Add on closed queue, want 0", got)
	}
	for v := range q.Next() {
		t.Errorf("received %v from closed queue, want nothing", v)
	}
}

// TestQueueCloseIdempotent protects that Close can be called more than once,
// e.g. by a deferred Close after an explicit one, without panicking on a
// double channel close or blocking.
func TestQueueCloseIdempotent(t *testing.T) {
	q := New[int](ChanLen(testChanLen))
	q.Add(1, 2, 3)
	closeWithTimeout(t, q)
	closeWithTimeout(t, q)

	if _, ok := <-q.Next(); ok {
		t.Error("channel must be closed after Close")
	}
}

// TestQueueCloseWhilePumpBlocked protects that Close does not deadlock when
// the pump goroutine is blocked sending an item to a full consumer channel
// that nobody reads from. Close must still return and the item is discarded.
func TestQueueCloseWhilePumpBlocked(t *testing.T) {
	// With an unbuffered channel and no consumer, the pump blocks on the
	// send as soon as it has taken the item out of the ring-buffer.
	q := New[int](ChanLen(0))
	q.Add(1)
	waitFor(t, "pump to take the item from the ring-buffer", func() bool { return q.Len() == 0 })

	closeWithTimeout(t, q)

	for v := range q.Next() {
		t.Errorf("received %v after Close, want the blocked item discarded", v)
	}
}

// TestQueueConcurrentAdd protects the goroutine-safety contract: items added
// concurrently by several producers are all delivered exactly once, and the
// items of each single producer keep their FIFO order.
func TestQueueConcurrentAdd(t *testing.T) {
	const (
		producers   = 8
		perProducer = 200
	)
	type item struct{ producer, seq int }

	q := New[item](ChanLen(4), InitialBufferSize(2)) // small sizes force buffer growth under contention
	defer q.Close()

	var wg sync.WaitGroup
	for p := range producers {
		wg.Go(func() {
			for s := range perProducer {
				q.Add(item{producer: p, seq: s})
			}
		})
	}

	nextSeq := make([]int, producers)
	for range producers * perProducer {
		select {
		case it := <-q.Next():
			if it.seq != nextSeq[it.producer] {
				t.Fatalf("producer %d: got seq %d, want %d (FIFO order violated)", it.producer, it.seq, nextSeq[it.producer])
			}
			nextSeq[it.producer]++
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out, received per producer: %v", nextSeq)
		}
	}
	wg.Wait()

	if got := q.Len(); got != 0 {
		t.Errorf("Len() = %d after receiving all items, want 0", got)
	}
}
