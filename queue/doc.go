// Package queue provides an unbounded, goroutine-safe FIFO queue that
// delivers its items through a Go channel.
//
// [New] returns a [Queue] and starts a pump goroutine. Producers call
// Queue.Add, which never blocks because the items wait in a ring buffer
// that grows as needed. Consumers receive the items in FIFO order from the
// channel returned by Queue.Next. The options [ChanLen] and
// [InitialBufferSize] set the channel capacity and the initial buffer
// length.
//
// Queue.Close discards every item that has not been received yet, closes
// the channel and waits for the pump goroutine to exit. Close is
// idempotent, and Add after Close is a no-op. There is no drain operation:
// to process every item, count the handled items, for example with a
// sync.WaitGroup, and call Close after the last one.
//
// See the [README] for the full method table, the Close semantics and a
// clean shutdown example.
//
// [README]: https://github.com/domonda/go-types/blob/master/queue/README.md
package queue
