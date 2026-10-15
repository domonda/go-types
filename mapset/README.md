# mapset

Set operations on sets represented as `map[K]struct{}` or `map[K]bool`, where the boolean value is ignored. The functions work on any map type with one of these underlying types, so you get the set algebra for your own defined map types without wrapping them.

```
import "github.com/domonda/go-types/mapset"
```

This package mirrors, function for function, the `container/mapset` package proposed for Go 1.28 ([go.dev/issue/77052](https://go.dev/issue/77052), CL 724420). Once that package ships, this one can be replaced by changing the import path. `String` is the one function whose output format may differ from the standard library version.

It is the shared implementation behind the set methods of `types.Set[T]`, `uu.IDSet`, `strutil.StringSet` and `email.AddressSet`, which together implement the abstract set interface of the Go collections proposal ([go.dev/issue/80590](https://go.dev/issue/80590)).

## Functions

`K` is any comparable type and `V` is `struct{}` or `bool`. Functions that take a set use a `~map[K]V` constraint, so they accept defined map types directly, and the ones returning a set return the type of their first operand.

| Function                        | Description                                        |
|---------------------------------|----------------------------------------------------|
| `Of(elems...)`                  | New `map[K]struct{}` with the elements.            |
| `OfBool(elems...)`              | New `map[K]bool` with the elements set to `true`.  |
| `Collect(seq)`                  | New `map[K]struct{}` with the elements of an `iter.Seq[K]`. |
| `CollectBool(seq)`              | New `map[K]bool` with the elements of an `iter.Seq[K]`. |
| `String(x)`                     | `{a b c}` with the elements formatted by `fmt` and sorted by their formatted text. |
| `Equal(x, y)`                   | Whether both sets have the same elements. nil and empty are equal. |
| `Contains(x, k)`                | Whether `k` is an element of `x`.                  |
| `ContainsAll(x, seq)`           | Whether every element of `seq` is in `x`. True for an empty `seq`. |
| `All(x)`                        | `iter.Seq[K]` over the elements in undefined order. |
| `Intersects(x, y)`              | Whether `x` and `y` have at least one element in common. |
| `Union(x, y)`                   | New set with the elements of both.                 |
| `Intersection(x, y)`            | New set with the elements in both.                 |
| `Difference(x, y)`              | New set with the elements of `x` that are not in `y`. |
| `SymmetricDifference(x, y)`     | New set with the elements in exactly one of `x` and `y`. |
| `Insert(x, elem)`               | Adds one element, reports whether `x` changed.     |
| `InsertAll(x, seq)`             | Adds the elements of `seq`, reports whether `x` changed. |
| `Delete(x, k)`                  | Removes one element, reports whether `x` changed.  |
| `DeleteAll(x, seq)`             | Removes the elements of `seq`, reports whether `x` changed. |
| `DeleteFunc(x, f)`              | Removes every element for which `f` returns true, reports whether `x` changed. |
| `UnionWith(x, y)`               | Adds the elements of `y` to `x`.                   |
| `IntersectionWith(x, y)`        | Removes the elements of `x` that are not in `y`.   |
| `DifferenceWith(x, y)`          | Removes the elements of `y` from `x`.              |
| `SymmetricDifferenceWith(x, y)` | Replaces `x` with the elements in exactly one of `x` and `y`. |

`Difference` is the asymmetric difference. Use `SymmetricDifference` for the elements in exactly one of the two sets.

## Operand types

The functions with two set operands differ in which map types they accept. This split is the one of upstream CL 724420.

| Functions                                          | Operands                                           |
|----------------------------------------------------|----------------------------------------------------|
| `Union`, `Intersection`, `Difference`, `SymmetricDifference`, `UnionWith` | Two different map types, also mixing `map[K]struct{}` and `map[K]bool`. |
| `Intersects`                                       | Two different map types with the same value type `V`. `Intersects(map[K]struct{}, map[K]bool)` does not compile. |
| `Equal`, `IntersectionWith`, `DifferenceWith`, `SymmetricDifferenceWith` | The same map type `M` for both operands.           |

`Of`, `OfBool`, `Collect` and `CollectBool` take no set and return the unnamed types `map[K]struct{}` and `map[K]bool`, not a defined type.

## Nil sets

A nil set is a valid empty set for every operation that does not have to store an element: `Contains`, `ContainsAll`, `All`, `Equal`, `Intersects`, `String`, `Delete`, `DeleteAll`, `DeleteFunc`, `IntersectionWith` and `DifferenceWith` all accept one. `Union`, `Intersection`, `Difference` and `SymmetricDifference` accept nil operands and always return a newly allocated set, never nil.

`Insert`, `InsertAll`, `UnionWith` and `SymmetricDifferenceWith` panic on a nil set if they actually have to store an element, exactly like an assignment to a nil Go map.

## Serialization

The package has no serialization code. A `map[K]struct{}` encodes with `encoding/json` as an object with empty object values, like `{"a":{}}`, and a `map[K]bool` as `{"a":true}`. The set types of this module that need JSON arrays, like `types.Set[T]` and `uu.IDSet`, implement it themselves.

## Gotchas

- In a `map[K]bool` set, membership means the key is present. A key with the value `false` is a member for `Contains`, `Equal` and all other functions.
- Functions that store an element always store `true` in a `map[K]bool` set, so `Difference(map[string]bool{"a": false}, ...)` returns `map[a:true]`.
- `String` sorts by the formatted text, so numbers sort as strings: `String(Of(10, 2, 1))` is `{1 10 2}`.
- The set types of this module don't use `String`, each has its own `String` method with its own format.
- Like any Go map, a set is not safe for concurrent writes.

## Examples

Set algebra:

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

	fmt.Println(mapset.String(mapset.Union(x, y)))
	fmt.Println(mapset.String(mapset.Intersection(x, y)))
	fmt.Println(mapset.String(mapset.Difference(x, y)))
	fmt.Println(mapset.String(mapset.SymmetricDifference(x, y)))

	fmt.Println(mapset.Insert(x, "set"))
	fmt.Println(mapset.Insert(x, "set"))
	fmt.Println(mapset.ContainsAll(x, slices.Values([]string{"go", "set"})))

	// Output:
	// {go json types}
	// {types}
	// {go}
	// {go json}
	// true
	// false
	// true
}
```

Defined map types and `map[K]bool` sets:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/mapset"
)

type Tags map[string]struct{}

func main() {
	tags := Tags{"go": {}}
	legacy := map[string]bool{"go": true, "old": false}

	// The result has the type of the first operand
	u := mapset.Union(tags, legacy)
	fmt.Printf("%T %s\n", u, mapset.String(u))
	d := mapset.Difference(legacy, tags)
	fmt.Printf("%T %v\n", d, d)

	// A key with the value false is still a member
	fmt.Println(mapset.Contains(legacy, "old"))

	// Intersects needs the same value type on both sides,
	// mapset.Intersects(tags, legacy) does not compile
	fmt.Println(mapset.Intersects(tags, mapset.Of("go")))

	fmt.Println(mapset.Equal(Tags(nil), Tags{}))
	fmt.Println(mapset.String(mapset.Of(10, 2, 1)))

	var empty Tags
	fmt.Println(mapset.Delete(empty, "go"), mapset.Union(empty, empty) != nil)

	// Output:
	// main.Tags {go old}
	// map[string]bool map[old:true]
	// true
	// true
	// true
	// {1 10 2}
	// false true
}
```

## Related

- [set](../set/): older helpers for `map[T]struct{}` sets, kept for existing callers.
- `types.Set[T]` in the [root package](../README.md), [uu](../uu/) `IDSet`, [strutil](../strutil/) `StringSet` and [email](../email/) `AddressSet`: set types whose methods delegate here.
- [Sets in go-types](../docs/explanation-sets.md): which set type to use and why the API follows the Go proposal.
