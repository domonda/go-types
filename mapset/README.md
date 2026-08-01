# mapset

Set operations on sets represented as `map[K]struct{}` or `map[K]bool` (where the boolean value is ignored).

```
import "github.com/domonda/go-types/mapset"
```

This package mirrors, function for function, the `container/mapset` package proposed for Go 1.28 ([go.dev/issue/77052](https://go.dev/issue/77052), CL 724420). Once that package ships, this one can be replaced by changing the import path.

It is the shared implementation behind the set methods of `types.Set[T]`, `uu.IDSet`, `strutil.StringSet` and `email.AddressSet`, which together implement the abstract set interface of [go.dev/issue/80590](https://go.dev/issue/80590).

All functions take a `~map[K]V` constraint, so they accept defined map types directly — no conversion needed — and the ones returning a set return that same defined type.

## API

| Function                            | Purpose                                            |
|-------------------------------------|----------------------------------------------------|
| `Of(elems...)` / `OfBool(elems...)` | New set from the passed elements.                  |
| `Collect(seq)` / `CollectBool(seq)` | New set from an `iter.Seq`.                        |
| `String(x)`                         | `{a b c}`, sorted by formatted element.            |
| `Equal(x, y)`                       | Same elements.                                     |
| `Contains(x, k)`                    | `k` is an element of `x`.                          |
| `ContainsAll(x, seq)`               | Every element of `seq` is in `x`; true if empty.   |
| `All(x)`                            | `iter.Seq` over the elements in undefined order.   |
| `Intersects(x, y)`                  | At least one element in common.                    |
| `Union(x, y)`                       | New set with the elements of both.                 |
| `Intersection(x, y)`                | New set with the elements in both.                 |
| `Difference(x, y)`                  | New set with the elements of `x` not in `y`.       |
| `SymmetricDifference(x, y)`         | New set with the elements in exactly one.          |
| `Insert(x, elem)`                   | Add one element, reports whether `x` changed.      |
| `InsertAll(x, seq)`                 | Add all of `seq`, reports whether `x` changed.     |
| `Delete(x, k)`                      | Remove one element, reports whether `x` changed.   |
| `DeleteAll(x, seq)`                 | Remove all of `seq`, reports whether `x` changed.  |
| `DeleteFunc(x, f)`                  | Remove where `f` is true, reports whether changed. |
| `UnionWith(x, y)`                   | In-place `Union` into `x`.                         |
| `IntersectionWith(x, y)`            | In-place `Intersection` into `x`.                  |
| `DifferenceWith(x, y)`              | In-place `Difference` into `x`.                    |
| `SymmetricDifferenceWith(x, y)`     | In-place `SymmetricDifference` into `x`.           |

`Difference` is the asymmetric difference. Use `SymmetricDifference` for the elements in exactly one of the two sets.

## Nil sets

A nil set is a valid empty set for every operation that does not have to store an element: `Contains`, `ContainsAll`, `All`, `Equal`, `Intersects`, `String`, `Delete`, `DeleteAll`, `DeleteFunc`, `IntersectionWith` and `DifferenceWith` all accept one. `Union`, `Intersection`, `Difference` and `SymmetricDifference` accept nil operands and always return a newly allocated set.

`Insert`, `InsertAll`, `UnionWith` and `SymmetricDifferenceWith` panic on a nil set if they actually have to store an element, exactly like an assignment to a nil Go map.

## Example

```go
package main

import (
	"fmt"
	"slices"

	"github.com/domonda/go-types/mapset"
)

func main() {
	x := mapset.Of("go", "types")
	y := mapset.Of("types", "json")

	fmt.Println(mapset.String(mapset.Union(x, y)))                // {go json types}
	fmt.Println(mapset.String(mapset.Intersection(x, y)))         // {types}
	fmt.Println(mapset.String(mapset.Difference(x, y)))           // {go}
	fmt.Println(mapset.String(mapset.SymmetricDifference(x, y)))  // {go json}

	fmt.Println(mapset.Insert(x, "set"))                          // true
	fmt.Println(mapset.Insert(x, "set"))                          // false
	fmt.Println(mapset.ContainsAll(x, slices.Values([]string{"go", "set"}))) // true
}
```
