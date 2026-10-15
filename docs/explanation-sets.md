# Sets

This page explains why the module has six ways to hold a set, how they relate,
which one to use in new code, and how their SQL and JSON encodings differ. The
API of each type is in the package READMEs linked at the end.

## The problem

Go has no built-in set type. The idiom is a map with empty struct values,
`map[T]struct{}`: the keys are the members and the values take no memory. The
idiom leaves several questions open:

- Where do union, intersection and difference live?
- Is a nil map an empty set, and what happens when you insert into it?
- How does a set print, and how is it stored in SQL and JSON?

This module answered these questions more than once over the years, so it now
has six set implementations. They share one implementation of the set
operations, but not one encoding.

## The six set implementations

| Implementation      | What it is                                         | Elements         | SQL                 | JSON                |
|---------------------|----------------------------------------------------|------------------|---------------------|---------------------|
| `mapset`            | functions over `~map[K]struct{}` and `~map[K]bool` | any `comparable` | none                | Go map default      |
| `set`               | functions over `map[T]struct{}`                    | `cmp.Ordered`    | none                | Go map default      |
| `types.Set[T]`      | `map[T]struct{}` with methods                      | `cmp.Ordered`    | none                | array, three states |
| `uu.IDSet`          | `map[uu.ID]struct{}` with methods                  | `uu.ID`          | array, three states | array, three states |
| `strutil.StringSet` | `map[string]struct{}` with methods                 | `string`         | none                | Go map default      |
| `email.AddressSet`  | `map[email.Address]struct{}` with methods          | `email.Address`  | array, three states | array, three states |

"Go map default" is what `encoding/json` does with a map that has no
`MarshalJSON` method: an object with empty object values, `{"a":{}}`. "Three
states" means nil, empty and populated sets encode to three different values,
see [Null conventions](explanation-null-conventions.md#three-states-for-sets-and-slices).
`email.AddressSet` also reads the object shape, which its `MarshalJSON` wrote
before, and an address list in a JSON string.

Every set reads and writes its elements as given, also those of a string type
whose own methods normalize, like `types.Set[money.Currency]`: normalizing
could merge two elements, like `"eur"` and `"EUR"`, and the set would read
back with fewer elements than it was written with. Normalize explicitly, for
example with `email.AddressSet.Normalized` or `NormalizedAddressList`.

All six work on plain maps, so `len` and `range` work on each of them, and a
set converts to another set type with the same element type, like
`strutil.StringSet(types.NewSet("a"))`.

## How they relate

```
   container/mapset             proposed for Go 1.28 (go.dev/issue/77052),
   (standard library)           not released
          ^
          |  mirrored function for function
          |
       mapset  <---------- set methods delegate here ----------+
                                                               |
      types.Set[T]    uu.IDSet    strutil.StringSet    email.AddressSet
           |              |               |                   |
           +--------------+-------+-------+-------------------+
                                  |
            internal/collections.Set   compile-time interface check
            internal/collections/settest.Run   behaviour check

      set   older helper functions, independent of mapset
```

`mapset` arrived in the 2026-09-02 release as the one shared implementation
(see [CHANGELOG.md](../CHANGELOG.md)). It mirrors the `container/mapset`
package proposed for Go 1.28 (go.dev/issue/77052, CL 724420) function for
function, so that it can be replaced by an import path change once that
package ships. In the same release the four set types got the method set of
the abstract set interface from the Go collections proposal
(go.dev/issue/80590): `All`, `Insert`, `InsertAll`, `Delete` (now returning
`bool`), `DeleteAll`, `DeleteFunc`, `ContainsAll`, `Union`, `Intersection`,
`Difference`, `SymmetricDifference`, `Intersects` and the in-place `*With`
variants. These methods delegate to the `mapset` function of the same name.

The specialized types exist because their elements need more than set
algebra:

- `uu.IDSet` reads and writes PostgreSQL `uuid[]` arrays and JSON arrays of
  UUID strings.
- `email.AddressSet` normalizes and validates addresses on request
  (`Normalized`, `NormalizedAddressList`, `Validate`, `AddNormalized`,
  `AddAddressPart`), reads and writes PostgreSQL text arrays, and writes an
  address list as text and XML.
- `strutil.StringSet` adds string helpers like `StringContainsAnyOfSet`.

`set` and `types.Set[T]` require `cmp.Ordered` elements, so neither can hold a
`uu.ID` (an array) or a struct. `mapset` only needs `comparable`.

## Which one to use in new code

| You need                                           | Use                            |
|----------------------------------------------------|--------------------------------|
| a set in memory                                    | `map[K]struct{}` with `mapset` |
| a set of UUIDs in a `uuid[]` column or a JSON array | `uu.IDSet`                     |
| a set of email addresses, normalized or in a `text[]` column | `email.AddressSet`             |
| a JSON array codec for ordered values              | `types.Set[T]`                 |

`mapset` is the default because it works on any map type with comparable keys,
including your own defined map types, and because it is the piece that will
become standard library. The `set` package is superseded by `mapset` for new
code but not deprecated. `strutil.StringSet` has no SQL or JSON methods and no
`IsNull`; a plain `map[string]struct{}` with `mapset` does the same job.

```go
package main

import (
	"fmt"
	"slices"

	"github.com/domonda/go-types/mapset"
	"github.com/domonda/go-types/uu"
)

func main() {
	a := uu.IDMust("0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d")
	b := uu.IDMust("7b0a3c1e-2f4d-4e5f-8a9b-0c1d2e3f4a5b")

	seen := mapset.Of(a)
	fmt.Println(mapset.Insert(seen, b), mapset.Insert(seen, b))
	fmt.Println(mapset.ContainsAll(seen, slices.Values([]uu.ID{a, b})))

	// A legacy map[K]bool set works as the other operand,
	// its bool values are ignored.
	legacy := map[uu.ID]bool{a: true}
	fmt.Println(mapset.Difference(seen, legacy))

	// Output:
	// true false
	// true
	// map[7b0a3c1e-2f4d-4e5f-8a9b-0c1d2e3f4a5b:{}]
}
```

Mixed operations return the representation of the left operand. In a
`map[K]bool` set a key counts as a member whatever its value, so
`map[K]bool{k: false}` contains `k`.

## Nil sets

All six follow the rules of Go maps:

- A nil set is a valid empty set for every read and every removal:
  `Contains`, `Len`, `All`, `Equal`, `Delete`, `DeleteAll` and the like.
- `Union`, `Intersection`, `Difference` and `SymmetricDifference` accept nil
  operands and always return a newly allocated set.
- `Insert`, `InsertAll`, `UnionWith` and `SymmetricDifferenceWith` panic on a
  nil set when they have to store an element, like an assignment to a nil map.

SQL NULL and JSON `null` decode to a nil set, so a set read from storage may
need `make` before the first insert. Two helpers allocate for you:
`set.Add(nilSet, values...)` returns a new set, and the pointer-receiver
`Add`, `AddSet`, `AddNormalized` and `AddAddressPart` of `email.AddressSet`
allocate the map of a nil variable or struct field.

`strutil.StringSet` is the one set type without `IsNull`; it does not tell a
nil set from an empty one.

## The conformance specification

The set types are checked against the collections proposal in two ways.

`internal/collections` holds the proposal's abstract `Collection[E, C]` and
`Set[E, S]` interfaces, copied verbatim. Their self-referential constraint, as
in `Collection[E any, C Collection[E, C]]`, needs the Go 1.26 type checker,
which is why the module requires Go 1.26. Each set type asserts the interface
in its test file, for example `var _ collections.Set[ID, IDSet] = IDSet(nil)`
in `uu/idset_test.go`. Because the assertions live in `_test.go` files,
`go vet ./...` and `go test ./...` catch a missing method, `go build` does not.

`internal/collections/settest` holds the behaviour of the abstract set as a
reusable test. `settest.Run(t, makeSet, nilSet, a, b, c)` gets a constructor,
the nil value of the set type and three distinct elements, and checks every
method of the abstract set interface, including the nil set rules above. It never inserts into the nil set.
It runs for `types.Set[int]`, `types.Set[string]`, `uu.IDSet`,
`strutil.StringSet` and `email.AddressSet`.

The point of running the same specification at every type, instead of only
testing `mapset`, is the planned swap: if the standard library package behaves
differently from `mapset`, the failure shows at each type the difference
affects.

## The planned swap to the standard library

[TODOS.md](../TODOS.md) lists two items for after v1.0:

1. Replace `mapset` with the standard library `container/mapset` once Go 1.28
   ships. Because the API is mirrored, this should be an import path change.
2. Reconsider `types.Set` and the `set` package at the same time. The proposed
   `container/set.Set` covers both, and neither has many callers: 3 call sites
   for `types.Set` and 5 for `set` across all known consumers.

Both standard library packages are proposals, so their final API can still
change. `mapset.String` is already a known difference: the proposal does not
pin down the output format, so the doc comment of `String` marks it as the one
function whose output may differ from the standard library version.

## SQL and JSON encodings compared

JSON of a set holding the elements `a` and `b`, an empty set and a nil set:

| Implementation                                     | nil    | empty | populated                       |
|----------------------------------------------------|--------|-------|---------------------------------|
| `map[K]struct{}` with `mapset` or `set`, `strutil.StringSet` | `null` | `{}`  | `{"a":{},"b":{}}`               |
| `types.Set[T]`                                     | `null` | `[]`  | `["a","b"]`, sorted             |
| `uu.IDSet`                                         | `null` | `[]`  | `["<id>","<id>"]`, sorted       |
| `email.AddressSet`                                 | `null` | `[]`  | `["a@example.com",...]`, sorted |

SQL, only for the two types with a `Value` method:

| Implementation     | nil    | empty | populated                                   |
|--------------------|--------|-------|---------------------------------------------|
| `uu.IDSet`         | `NULL` | `{}`  | `{"<id>","<id>"}`, sorted by bytes          |
| `email.AddressSet` | `NULL` | `{}`  | `{"a@example.com","b@example.com"}`, sorted |

`String()` differs as well:

| Implementation      | nil     | empty   | populated                      |
|---------------------|---------|---------|--------------------------------|
| `mapset.String`     | `{}`    | `{}`    | `{a b}`                        |
| `types.Set[string]` | `<nil>` | `[]`    | `["a", "b"]`                   |
| `uu.IDSet`          | `set[]` | `set[]` | `set[<id>,<id>]`               |
| `strutil.StringSet` | `[]`    | `[]`    | `["a", "b"]`                   |
| `email.AddressSet`  | `<nil>` | empty   | `a@example.com, b@example.com` |

Other differences:

- `uu.IDSet.Scan("")` gives an empty non-nil set, `email.AddressSet.Scan("")`
  and a whitespace-only string the nil set. `email.AddressSet.Scan` also reads
  a JSON array or object from a `json` or `jsonb` column, and a plain string
  that is not an array literal as address list, split with
  `email.SplitAddressList`.
- `email.AddressSet.Value` and `MarshalText` return an error for an invalid
  address, `MarshalJSON` writes every set and round-trips it exactly.
- `types.Set.UnmarshalJSON` and `uu.IDSet.UnmarshalJSON` always assign a new
  map. Refilling the existing map would rewrite every other variable that
  shares it, like a default permission set copied into a struct field; the test
  `TestSet_UnmarshalJSON_DoesNotMutateSharedMap` in
  [set_test.go](../set_test.go) describes the privilege escalation this
  prevents.
- Use `omitzero`, not `omitempty`, on optional set fields. `omitempty` drops an
  empty set, which then decodes as nil.

```go
package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/email"
	"github.com/domonda/go-types/mapset"
	"github.com/domonda/go-types/strutil"
	"github.com/domonda/go-types/uu"
)

func show(name string, set any) {
	j, _ := json.Marshal(set)
	sqlValue := "none"
	if v, ok := set.(driver.Valuer); ok {
		val, _ := v.Value()
		sqlValue = fmt.Sprint(val)
	}
	fmt.Printf("%s: JSON %s, SQL %s\n", name, j, sqlValue)
}

func main() {
	id := uu.IDMust("0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d")
	show("mapset", mapset.Of("a"))
	show("types.Set", types.NewSet("a"))
	show("strutil.StringSet", strutil.NewStringSet("a"))
	show("email.AddressSet", email.MakeAddressSet("a@example.com"))
	show("uu.IDSet", uu.MakeIDSet(id))
	show("empty uu.IDSet", uu.IDSet{})
	show("nil uu.IDSet", uu.IDSet(nil))

	// Output:
	// mapset: JSON {"a":{}}, SQL none
	// types.Set: JSON ["a"], SQL none
	// strutil.StringSet: JSON {"a":{}}, SQL none
	// email.AddressSet: JSON ["a@example.com"], SQL {"a@example.com"}
	// uu.IDSet: JSON ["0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d"], SQL {"0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d"}
	// empty uu.IDSet: JSON [], SQL {}
	// nil uu.IDSet: JSON null, SQL <nil>
}
```

## Related

- [Null conventions](explanation-null-conventions.md): the three-state codec
  shared with `uu.IDSlice` and the `nullable` arrays.
- Reference: [mapset](../mapset/README.md), [set](../set/README.md),
  [root package `types.Set`](../README.md), [uu](../uu/README.md),
  [strutil](../strutil/README.md), [email](../email/README.md).
