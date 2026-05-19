# deref

Safe pointer dereferencing for common Go types. Each function returns the pointed-to value when the pointer is non-nil and a caller-supplied default when it is nil, avoiding the panic that direct dereferencing would cause.

```
import "github.com/domonda/go-types/deref"
```

## API

All functions follow the same shape: `Func(ptr *T, defaultVal T) T`.

| Function                                         | Pointer type                                     |
|--------------------------------------------------|--------------------------------------------------|
| `Bool`                                           | `*bool`                                          |
| `String`                                         | `*string`                                        |
| `Int`                                            | `*int`                                           |
| `Int32`                                          | `*int32`                                         |
| `Int64`                                          | `*int64`                                         |
| `Uint`                                           | `*uint`                                          |
| `Uint64`                                         | `*uint64`                                        |
| `Float32`                                        | `*float32`                                       |
| `Float64`                                        | `*float64`                                       |
| `Time`                                           | `*time.Time`                                     |

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/domonda/go-types/deref"
)

type Config struct {
	Verbose *bool
	Name    *string
	Started *time.Time
}

func main() {
	var cfg Config

	verbose := deref.Bool(cfg.Verbose, false)
	name := deref.String(cfg.Name, "anonymous")
	started := deref.Time(cfg.Started, time.Now())

	fmt.Println(verbose, name, started)
}
```

## Related

- Root package `types` provides `Ptr`, `FromPtr`, and `FromPtrOr` for generic pointer conversions.
