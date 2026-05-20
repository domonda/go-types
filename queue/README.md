# queue

Unbounded, concurrent-safe FIFO queue that exposes its output as a Go channel. Producers call `Add` without blocking; consumers read from the channel returned by `Next`. Items are held in a growing ring buffer until the channel can take them, so adding is never blocked by slow consumers.

```
import "github.com/domonda/go-types/queue"
```

## API

```go
type Queue[T any] interface {
    Add(items ...T)
    Next() <-chan T
    Len() int
    Cap() int
    Close()
}

func New[T any](options ...Option) Queue[T]

// Options
type ChanLen int           // consumer-channel capacity
type InitialBufferSize int // initial ring-buffer capacity
```

- `New[T]` creates a queue of items of type `T`. With no options it uses `DefaultChanLen` (`32`) and `DefaultInitialBufferSize` (`32`).
- `Add` enqueues one or more items onto an internal ring buffer that grows as needed, so it never blocks. A background pump goroutine delivers items to the consumer channel.
- `Next` returns the read-only channel consumers range over.
- `Len` is the current number of items waiting (channel + buffer).
- `Cap` is the current capacity (channel cap + buffer length).
- `Close` shuts the queue down, discards every item still queued, and closes the consumer channel. It blocks until the pump goroutine has exited.

## Tuning

`New` takes optional configuration values:

- `ChanLen` — capacity of the consumer channel. A pure throughput knob: a larger channel lets more items sit ready to consume and runs the pump goroutine less often. Must be `>= 0` (`0` is a valid unbuffered hand-off); `New` panics otherwise.
- `InitialBufferSize` — initial length of the ring buffer that absorbs items the channel cannot hold yet. The buffer grows automatically as needed. Must be `>= 1`; `New` panics otherwise.

```go
q := queue.New(queue.ChanLen(64), queue.InitialBufferSize(16))
```

## Example

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/queue"
)

func main() {
	q := queue.New[string]() // DefaultChanLen, DefaultInitialBufferSize
	defer q.Close()

	go func() {
		for item := range q.Next() {
			fmt.Println("got:", item)
		}
	}()

	q.Add("a", "b", "c")
	q.Add("d", "e", "f")
}
```

## Notes

- `Queue[T]` is typed; no type assertions needed in the consumer.
- Closing the queue closes the consumer channel; ranging consumers exit cleanly.
- `Close` discards every item still queued — the ring buffer and whatever was staged in the channel — then closes the channel. It blocks until the pump goroutine has exited, so the queue is fully shut down once `Close` returns.
- A consumer reading concurrently with `Close` may still win some of the channel-staged items (up to `ChanLen`) before they're discarded; the discard guarantee covers the queue's own backlog, not an actively racing consumer.
- The ring buffer doubles in size when full; it does not shrink.
- The pump goroutine does a blocking send into the channel, so a full channel parks it in the scheduler rather than busy-spinning.
