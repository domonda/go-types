# set

Generic helper functions for sets represented as plain `map[T]struct{}` values. The package defines no set type: it works on map values directly, so zero values, `range` iteration and `len` stay the native map ones.

```
import "github.com/domonda/go-types/set"
```

> **Prefer [`mapset`](../mapset/) for new code.** It covers the full set algebra, works with `map[K]bool` sets and any comparable element type, and mirrors the `container/mapset` package proposed for Go 1.28, so it can be swapped for the standard library one by changing the import path. This package predates it and is kept unchanged for its existing callers. Two helpers here have no direct `mapset` counterpart: `Add(set, values...)` allocates a nil set and returns it, where `mapset.Insert` panics on a nil set and takes a single element, and `ContainsAny` and `ContainsAll` take variadic values, where `mapset` offers `Intersects` and `ContainsAll(x, iter.Seq[K])`.

For a set type with methods, see the root package type `types.Set[T]`. It is a defined map type `map[T]struct{}` with the set algebra of `mapset` as methods and JSON marshalling as a sorted array.

## Functions

All functions take and return the unnamed type `map[T]struct{}` with `T cmp.Ordered`. Values of a defined map type with that underlying type, like `types.Set[T]` or `strutil.StringSet`, can be passed directly.

| Function                       | Description                                        |
|--------------------------------|----------------------------------------------------|
| `New(values...)`               | New set with `values`. Never nil, `New[int]()` is an empty set. |
| `Add(set, values...)`          | Adds `values` and returns the set. A nil set is allocated first, so always use the result. |
| `Contains(set, v)`             | Whether `v` is in the set.                         |
| `ContainsAll(set, values...)`  | Whether every value is in the set. True for no values. |
| `ContainsAny(set, values...)`  | Whether at least one value is in the set. False for no values. |
| `ContainsAllOther(set, other)` | Whether every element of `other` is in the set. True for an empty `other`. |
| `ContainsAnyOther(set, other)` | Whether at least one element of `other` is in the set. False for an empty `other`. |

Every function that takes a set accepts a nil set as the empty set. Only `Add` writes to it.

Migrating to `mapset`:

| `set`                          | `mapset`                                           |
|--------------------------------|----------------------------------------------------|
| `New(values...)`               | `Of(values...)`                                    |
| `Add(set, values...)`          | `InsertAll(set, slices.Values(values))` on a non-nil set |
| `Contains(set, v)`             | `Contains(set, v)`                                 |
| `ContainsAll(set, values...)`  | `ContainsAll(set, slices.Values(values))`          |
| `ContainsAny(set, values...)`  | `Intersects(set, Of(values...))`                   |
| `ContainsAllOther(set, other)` | `ContainsAll(set, All(other))`                     |
| `ContainsAnyOther(set, other)` | `Intersects(set, other)`                           |

## Serialization

The package has no serialization code. A `map[T]struct{}` encodes with `encoding/json` as an object with empty object values, like `{"a":{},"b":{}}`, a nil map as `null`. Use `types.Set[T]` for JSON arrays. There is no SQL support.

## Gotchas

- `T` must be `cmp.Ordered` although the functions only need `comparable`, so `uu.IDSet` (`[16]byte` keys) and struct keys don't work. Use `mapset`.
- `Add` returns the set. Calling `set.Add(s, v)` on a nil `s` without assigning the result loses the new map.
- The functions take `map[T]struct{}`, not `map[T]bool`. Use `mapset` for `map[T]bool` sets.
- Like any Go map, a set is not safe for concurrent writes.

## Examples

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/set"
)

func main() {
	tags := set.New("go", "types", "validation")
	tags = set.Add(tags, "json")

	fmt.Println(set.Contains(tags, "go"))
	fmt.Println(set.ContainsAll(tags, "go", "json"))
	fmt.Println(set.ContainsAny(tags, "rust"))
	fmt.Println(set.ContainsAll(tags), set.ContainsAny(tags))

	// Add allocates the map for a nil set
	var ids map[int]struct{}
	ids = set.Add(ids, 3, 1, 3)
	fmt.Println(len(ids), set.ContainsAllOther(ids, set.New(1)), set.ContainsAnyOther(ids, set.New(2)))

	// The native JSON of a map[T]struct{} is an object, not an array
	j, _ := json.Marshal(set.New("a", "b"))
	fmt.Println(string(j))

	// Output:
	// true
	// true
	// false
	// true false
	// 2 true false
	// {"a":{},"b":{}}
}
```

## Related

- [mapset](../mapset/): the full set algebra on `map[K]struct{}` and `map[K]bool` sets.
- `types.Set[T]` in the [root package](../README.md): set type with methods and JSON support.
- `strutil.StringSet` in [strutil](../strutil/), `uu.IDSet` in [uu](../uu/) and `email.AddressSet` in [email](../email/): set types for strings, UUIDs and email addresses.
- [Sets in go-types](../docs/explanation-sets.md): which set type to use.
