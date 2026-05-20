# set

Generic set helpers built on the idiomatic Go `map[T]struct{}` representation. There is no `Set` type — the package operates directly on map values, which keeps zero values, JSON marshaling, and range iteration native.

```
import "github.com/domonda/go-types/set"
```

For an opaque `Set[T]` struct with methods (union, intersection, difference, Sorted, JSON), see the root package `types.Set[T]` in `set.go`.

## API

All functions are generic over `T cmp.Ordered`.

| Function                       | Purpose                                            |
|--------------------------------|----------------------------------------------------|
| `New(values...)`               | Create a new set populated with `values`.          |
| `Add(set, values...)`          | Add values; creates the set if `nil`. Returns the (possibly new) set. |
| `Contains(set, v)`             | Whether `v` is in the set.                         |
| `ContainsAll(set, values...)`  | All values are present.                            |
| `ContainsAny(set, values...)`  | At least one value is present.                     |
| `ContainsAllOther(set, other)` | All keys of `other` are present.                   |
| `ContainsAnyOther(set, other)` | At least one key of `other` is present.            |

## Example

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/set"
)

func main() {
	tags := set.New("go", "types", "validation")
	tags = set.Add(tags, "json")

	fmt.Println(set.Contains(tags, "go"))           // true
	fmt.Println(set.ContainsAll(tags, "go", "json")) // true
	fmt.Println(set.ContainsAny(tags, "rust"))      // false
}
```

## Related

- `types.Set[T]` — struct-based set with rich method API.
- `strutil.StringSet` — string-specialized set with case folding helpers.
- `uu.IDSet` — UUID-specialized set.
