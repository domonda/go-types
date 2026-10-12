# queue

Unbounded, goroutine-safe FIFO queue that delivers its items through a Go channel. Producers call `Add`, which never blocks; consumers receive from the channel returned by `Next`. Items wait in a ring buffer that grows as needed until the channel can take them, so a slow consumer never blocks a producer.

```
import "github.com/domonda/go-types/queue"
```

## API

| Identifier                 | Description                                        |
|----------------------------|----------------------------------------------------|
| `Queue[T]`                 | Interface with the methods below, returned by `New`. |
| `New[T](options...)`       | Starts a queue of items of type `T` and its pump goroutine. Panics for a `ChanLen` below 0 or an `InitialBufferSize` below 1. |
| `Option`                   | Interface of the options for `New`. Its method is unexported, so only `ChanLen` and `InitialBufferSize` implement it. |
| `ChanLen`                  | Option: capacity of the consumer channel, default `DefaultChanLen`. `ChanLen(0)` is a valid unbuffered hand-off. |
| `InitialBufferSize`        | Option: initial length of the ring buffer, default `DefaultInitialBufferSize`. |
| `DefaultChanLen`           | `32`.                                              |
| `DefaultInitialBufferSize` | `32`.                                              |

| Method          | Description                                        |
|-----------------|----------------------------------------------------|
| `Add(items...)` | Appends the items in order. Never blocks. A no-op without items and after `Close`. |
| `Next()`        | The receive-only channel that delivers the items in FIFO order. It is closed by `Close`. |
| `Len()`         | Number of items in the ring buffer plus the items staged in the channel. |
| `Cap()`         | Channel capacity plus the current ring buffer length. |
| `Close()`       | Shuts the queue down, see below.                   |

A background pump goroutine moves items one at a time from the ring buffer into the channel. It is the only sender and closer of the channel, which keeps delivery strictly FIFO. When the channel is full, the pump blocks in the send and does not spin.

## Close

`Close` discards every item that has not been received yet, closes the channel and waits until the pump goroutine has exited. After `Close` returns:

- The items still in the ring buffer and in the channel are gone. A consumer that receives concurrently with `Close` may still get some of the items staged in the channel, up to `ChanLen`, before they are discarded.
- A `range q.Next()` loop ends.
- `Add` is a no-op, the items are dropped silently.
- Calling `Close` again returns without effect, so a deferred `Close` after an explicit one is safe.
- `Cap()` returns the channel capacity only, because the ring buffer was dropped.

There is no drain operation. To process every item before shutting down, count the items the consumer has handled, for example with a `sync.WaitGroup`, and call `Close` after the last one, as in the first example below.

## Tuning

- `ChanLen` is a throughput setting: a larger channel lets more items sit ready for the consumer and wakes the pump less often. It also bounds how many items a racing consumer can still receive during `Close`.
- `InitialBufferSize` sets the starting length of the ring buffer. The buffer doubles when it is full and never shrinks.

```go
q := queue.New[string](queue.ChanLen(64), queue.InitialBufferSize(16))
```

## Gotchas

- `defer q.Close()` right after `Add` in `main` loses items: `Close` discards everything the consumer has not received yet, and `main` can return before the consumer prints anything.
- Waiting for `Len() == 0` before `Close` is not enough. The pump holds one item outside the buffer and the channel while it waits to send it, so `Len()` can be one lower than the number of undelivered items, and `0` while an item is still on its way. Count handled items instead.
- `Len()` and `Cap()` are snapshots that can change as soon as they return.
- With several producers calling `Add` concurrently, the items of each producer keep their order, but the items of different producers interleave.
- The ring buffer never shrinks, so a burst of items keeps its memory allocated for the lifetime of the queue. Delivered items are released for garbage collection.
- Each queue runs one goroutine until `Close` is called. A queue that is never closed leaks it.

## Examples

Process every item, then shut down cleanly:

```go
package main

import (
	"fmt"
	"sync"

	"github.com/domonda/go-types/queue"
)

func main() {
	q := queue.New[string]()

	items := []string{"a", "b", "c", "d", "e"}
	var processed sync.WaitGroup
	processed.Add(len(items))

	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		for item := range q.Next() {
			fmt.Println("got:", item)
			processed.Done()
		}
	}()

	q.Add(items[:2]...)
	q.Add(items[2:]...)

	processed.Wait() // every item has been handled, nothing is left to discard
	q.Close()        // closes the channel, so the consumer's range loop ends
	<-consumerDone
	fmt.Println("closed, len:", q.Len())

	// Output:
	// got: a
	// got: b
	// got: c
	// got: d
	// got: e
	// closed, len: 0
}
```

`Close` discards queued items:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/queue"
)

func main() {
	q := queue.New[int](queue.ChanLen(4), queue.InitialBufferSize(8))
	q.Add(1, 2, 3) // nobody reads these

	q.Close() // discards the queued items
	q.Close() // Close is idempotent
	q.Add(4)  // Add after Close is a no-op

	received := 0
	for range q.Next() {
		received++
	}
	fmt.Println("received:", received, "len:", q.Len(), "cap:", q.Cap())

	// Output:
	// received: 0 len: 0 cap: 4
}
```

## Related

- [docs index](../docs/README.md): explanations and how-to guides for the other packages of this module.
