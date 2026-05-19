# queue

Unbounded, concurrent-safe FIFO queue that exposes its output as a Go channel. Producers call `Add` without blocking; consumers read from the channel returned by `Next`. Items are buffered in a growing ring buffer when the channel is full, so adding is never blocked by slow consumers.

```
import "github.com/domonda/go-types/queue"
```

## API

```go
type Queue interface {
    Add(items ...any)
    Next() <-chan any
    Len() int
    Cap() int
    Close()
}

func New() Queue
```

- `Add` enqueues one or more items. Never blocks: items that don't fit in the channel are pushed onto an internal ring buffer that grows as needed.
- `Next` returns the read-only channel consumers range over.
- `Len` is the current number of items waiting (channel + buffer).
- `Cap` is the current capacity (channel cap + buffer length).
- `Close` shuts the queue down and closes the consumer channel.

## Tuning

Package-level variables set the initial capacities. Adjust at program start, before calling `New`.

| Variable             | Default | Meaning                          |
|----------------------|---------|----------------------------------|
| `ChanLen`            | `32`    | Size of the consumer channel.    |
| `InitialBufferSize`  | `8`     | Initial ring buffer length.      |

## Example

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/queue"
)

func main() {
	q := queue.New()
	defer q.Close()

	go func() {
		for item := range q.Next() {
			fmt.Println("got:", item)
		}
	}()

	q.Add("a", "b", "c")
	q.Add(1, 2, 3)
}
```

## Notes

- The queue holds `any`; type-assert in the consumer.
- Closing the queue closes the consumer channel; ranging consumers exit cleanly.
- The ring buffer doubles in size when full; it does not shrink.
