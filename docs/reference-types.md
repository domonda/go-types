# types (root package)

The root package of the module holds the building blocks the subpackages share: the generic `Set[T]`, the `Validator` and `ValidatErr` interfaces with `Validate`, `TryValidate` and `DeepValidate`, the `Normalizable` interfaces of the domain types, pointer and iterator helpers, the length-limited `LenString`, the `Finder` interface, the per-key `KeyMutex`, and two reflection helpers. Import it when you validate values or structs of `go-types` values, or need one of these helpers.

```
import "github.com/domonda/go-types"
```

The package name is `types`.

## Set

`Set[T cmp.Ordered]` is the defined map type `map[T]struct{}`. Its methods delegate to [`mapset`](../mapset/README.md) and mirror the abstract set interface of the Go collections proposal ([go.dev/issue/80590](https://go.dev/issue/80590)).

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NewSet[T cmp.Ordered](vals ...T) Set[T]`          | New set with `vals`, duplicates removed. Never nil. |
| `SetToSortedSlice[S ~map[T]struct{}, T cmp.Ordered](set S) []T` | Values in ascending order, nil for an empty or nil set. |
| `SetToRandomizedSlice[S ~map[T]struct{}, T cmp.Ordered](set S) []T` | Values in map iteration order, which is unspecified but not a shuffle. nil for an empty or nil set. |
| `ReduceSet[S ~map[T]struct{}, T cmp.Ordered, R any](set S, reduceFunc func(last R, val T) R) (result R)` | Fold the values, starting with the zero `R`, in undefined order. |
| `ReduceSlice[S ~[]T, T cmp.Ordered, R any](slice S, reduceFunc func(last R, val T) R) (result R)` | Fold the slice elements in order, starting with the zero `R`. |
| `SliceContainsAll[T comparable](outer []T, inner ...T) bool` | Every value of `inner` is in `outer`. Duplicates in `inner` are not counted, so `SliceContainsAll([]int{1}, 1, 1)` is true. True for an empty `inner`. |
| `SliceContainsAny[T comparable](outer []T, inner ...T) bool` | Any value of `inner` is in `outer`. False for an empty `inner`. |

| Method                                     | Description                                        |
|--------------------------------------------|----------------------------------------------------|
| `Len() int`                                | Number of values, 0 for nil.                       |
| `IsEmpty() bool`                           | True for an empty or nil set.                      |
| `IsNull() bool`                            | True only for a nil set. Implements `nullable.Nullable`. |
| `Contains(val T) bool`                     | `val` is in the set.                               |
| `ContainsAny(vals ...T) bool`              | Any of `vals` is in the set. False without arguments. |
| `ContainsAll(seq iter.Seq[T]) bool`        | Every value of `seq` is in the set. True for an empty `seq`. |
| `ContainsSet(other Set[T]) bool`           | Every value of `other` is in the set.              |
| `GetOne() T`                               | Some value of the set, or the zero `T` for an empty set. |
| `All() iter.Seq[T]`                        | Iterate in undefined order.                        |
| `Sorted() []T`                             | `SetToSortedSlice(set)`.                           |
| `String() string`                          | Sorted values formatted with `%#v`: `[1, 2, 3]`, `["a", "b"]`, `<nil>` for nil. |
| `Equal(other Set[T]) bool`                 | Same values. A nil set equals an empty one.        |
| `Intersects(other Set[T]) bool`            | At least one value in common.                      |
| `Clone() Set[T]`                           | New set with the same values, nil for nil.         |
| `Map(mapFunc func(T) (T, bool)) Set[T]`    | New set of the mapped values for which `mapFunc` returns true. Never nil. |
| `Insert(val T) bool`                       | Add `val`, report whether the set changed.         |
| `InsertAll(seq iter.Seq[T]) bool`          | Add the values of `seq`, report whether the set changed. |
| `Delete(val T) bool`                       | Remove `val`, report whether the set changed.      |
| `DeleteAll(seq iter.Seq[T]) bool`          | Remove the values of `seq`, report whether the set changed. |
| `DeleteFunc(del func(T) bool) bool`        | Remove the values where `del` is true, report whether the set changed. |
| `Clear()`                                  | Remove all values.                                 |
| `Union(other Set[T]) Set[T]`               | New set with the values of both.                   |
| `Intersection(other Set[T]) Set[T]`        | New set with the values in both.                   |
| `Difference(other Set[T]) Set[T]`          | New set with the values of `set` that are not in `other`. |
| `SymmetricDifference(other Set[T]) Set[T]` | New set with the values in exactly one of the two. |
| `UnionWith(other Set[T])`                  | In-place `Union`.                                  |
| `IntersectionWith(other Set[T])`           | In-place `Intersection`.                           |
| `DifferenceWith(other Set[T])`             | In-place `Difference`.                             |
| `SymmetricDifferenceWith(other Set[T])`    | In-place `SymmetricDifference`.                    |

Implements: `json.Marshaler` (value receiver), `json.Unmarshaler` (pointer receiver), `fmt.Stringer`, `nullable.Nullable`. There is no `sql.Scanner`, `driver.Valuer`, text marshaling or `JSONSchema` method.

A nil `Set` is a valid empty set for every method except `Insert`, `InsertAll`, `UnionWith` and `SymmetricDifferenceWith`, which panic like an assignment to a nil map when they have to add a value. `Union`, `Intersection`, `Difference`, `SymmetricDifference` and `Map` always return a new non-nil set.

`Difference` is the asymmetric difference of the collections proposal. Before the 2026-09-02 release it returned the symmetric difference, which is now `SymmetricDifference`. The change does not break compilation, so check old call sites.

## Validation

Two interfaces describe validatable types:

| Interface    | Method             | Use it when                                        |
|--------------|--------------------|----------------------------------------------------|
| `Validator`  | `Valid() bool`     | A yes/no answer is enough.                         |
| `ValidatErr` | `Validate() error` | The caller should learn why a value is invalid. Any non-nil error means invalid. |

Composition helpers:

| Identifier                                         | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ValidatorFunc`                                    | `func() bool` that implements `Validator`.         |
| `StaticValidator`                                  | `bool` that implements `Validator`.                |
| `Validators`                                       | `[]Validator`. `Valid()` is false if any element is false. nil elements, nil funcs and nil pointers that `Validate` does not validate are skipped. |
| `CombinedValidator(validators ...Validator) Validator` | Returns `Validators(validators)`.                  |
| `ValidatErrFunc`                                   | `func() error` that implements `ValidatErr`.       |
| `StaticValidatErr`                                 | `struct{ Err error }` whose `Validate()` returns `Err`. |
| `ValidatErrs`                                      | `[]ValidatErr`. `Validate()` calls `Validate` for every element and returns all errors joined with `errors.Join`, or nil. |
| `CombinedValidatErr(validatErrs ...ValidatErr) ValidatErr` | Returns `ValidatErrs(validatErrs)`.                |
| `ValidatorAsValidatErr`                            | `struct{ Validator }` whose `Validate()` returns an error naming the type and value of the wrapped `Validator` if it is invalid. |

Entry points:

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `Validate(v any) error`                            | Validate one value, see the rules below. nil if `v` implements neither interface. |
| `TryValidate(v any) (err error, isValidatable bool)` | Like `Validate`, plus whether `v` implements one of the interfaces. |
| `DeepValidate(v any) []error`                      | Validate `v` and everything reachable from it. Returns nil if nothing is invalid. |

`Validate` and `TryValidate` follow these rules:

1. If `v` implements `ValidatErr`, only `Validate()` is called and its error is returned unchanged, even if `v` also implements `Validator`.
2. Otherwise, if `v` implements `Validator` and `Valid()` is false, the error is `fmt.Errorf("invalid %T value: %#v", v, v)`, like `invalid main.Status value: "bad"`. The whole value is part of the message, so implement `ValidatErr` for a value that must not be logged, like a password, or that can be large.
3. A nil pointer whose element type implements the method with a value receiver is not validated: `Validate` returns nil and `TryValidate` returns `nil, true`. A method with a pointer receiver is called with the nil pointer, so it can report nil itself. A struct whose method is promoted from an embedded nil pointer or interface, like `struct{ *T }` with a nil `*T` and a value receiver on `T`, or `struct{ types.ValidatErr }` with a nil `ValidatErr`, is not validated either. A panic in a method that the struct declares itself is passed on.

`DeepValidate` walks `v` with reflection:

- It calls `Validate` on `v` and then on every reachable value. If only the pointer type implements an interface, like `*LenString`, it validates through a pointer to the value, or to a copy if the value is not addressable. `v` itself is not copied: passed by value it is validated with the methods of its own type like by `Validate`, so a pointer receiver `Validate` calling `DeepValidate(*receiver)` doesn't call itself again. The nested call has its own cycle detection, and an outer `DeepValidate` reports the field errors twice.
- It follows pointers and interfaces to any depth. nil pointers, interfaces and funcs stop the descent.
- It descends into exported struct fields, the exported fields of unexported embedded structs, slice and array elements, and map values. Unexported fields and map keys are not validated. Map keys are sorted with `CompareReflectValue`, so the error order is deterministic.
- A value that is validatable itself is still descended into, so a struct can report its own error and errors of its fields.
- A pointer, map or slice that is already on the current path is not entered again, so cyclic data is validated once.
- Every error except one of `v` itself is prefixed with its path and wraps the original error, so `errors.Is` and `errors.As` work:

| Path element            | Example error                                      |
|-------------------------|----------------------------------------------------|
| `struct field <Name>`   | `struct field Status: invalid main.Status value: "bad"` |
| `element [<index>]`     | `struct field Tags -> element [1]: invalid main.Status value: "nope"` |
| `map value [<key %#v>]` | `map value ["a"]: invalid main.Status value: "y"`  |

Use `errors.Join(types.DeepValidate(v)...)` to get one error, which is nil if the slice is empty.

## Normalization

| Interface                      | Methods                                            |
|--------------------------------|----------------------------------------------------|
| `Normalizable[T any]`          | `Normalized() (T, error)`: the value in its canonical form, or an error. Should be idempotent. |
| `NormalizableValidator[T any]` | `Validator`, `Normalizable[T]`, and `ValidAndNormalized() bool`, true if the value is valid and already canonical. |

These types assert `NormalizableValidator` at compile time: `bank.IBAN`, `bank.NullableIBAN`, `bank.BIC`, `bank.NullableBIC`, `country.Code`, `country.NullableCode`, `email.Address`, `email.NullableAddress`, `email.AddressList`, `email.AddressSet`, `language.Code`, `money.Currency`, `money.NullableCurrency`, `vat.ID`, `vat.NullableID`, `date.Date`, `date.NullableDate`, `date.YearMonth`, `date.NullableYearMonth`, `date.YearQuarter` and `date.NullableYearQuarter`. The dates take a language hint with `NormalizedLang` instead of `Normalized`.

## Pointers

| Function                                     | Description                        |
|----------------------------------------------|------------------------------------|
| `FromPtr[T any](ptr *T) T`                   | `*ptr`, or the zero `T` for nil.   |
| `FromPtrOr[T any](ptr *T, defaultValue T) T` | `*ptr`, or `defaultValue` for nil. |

## Iterators

| Function                                           | Description                                  |
|----------------------------------------------------|----------------------------------------------|
| `Yield[V any](value V) iter.Seq[V]`                | Sequence of the single `value`.              |
| `Yield2[K, V any](key K, value V) iter.Seq2[K, V]` | Sequence of the single pair `key, value`.    |
| `YieldErr[K any](err error) iter.Seq2[K, error]`   | Sequence of the single pair zero `K`, `err`. |
| `Seq2NilError[K any](seq iter.Seq[K]) iter.Seq2[K, error]` | Pairs every value of `seq` with a nil error. |

## LenString

`LenString` holds a string with a minimum and maximum length, counted in bytes.

| Identifier                                         | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NewLenString(str string, min, max int) *LenString` | New pointer without validation.                    |
| `MustLenString(str string, min, max int) LenString` | New value, panics if `Validate` fails.             |
| `Validate() error`                                 | Error for a nil pointer, a negative `min` or `max`, `min > max`, or a length outside `[min, max]`. |
| `SetString(str string) error`                      | Set `str` if its length is within the limits, else return an error and keep the old string. |
| `ScanString(source string, validate bool) error`   | Like `SetString` if `validate` is true. With `false` it sets `source` without any check. |
| `String() string`                                  | The string, `"<nil>"` for a nil pointer.           |
| `MinLen() int` / `MaxLen() int`                    | The limits.                                        |
| `MarshalText() ([]byte, error)` / `UnmarshalText(text []byte) error` | Text form. `UnmarshalText` validates like `SetString`. |
| `MarshalJSON() ([]byte, error)` / `UnmarshalJSON(text []byte) error` | JSON string. `UnmarshalJSON` validates like `SetString`. |

Implements, all with pointer receivers: `ValidatErr`, `fmt.Stringer`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `json.Marshaler`, `json.Unmarshaler`, `strfmt.Scannable`. There is no SQL support.

## Finder

`Finder` is implemented by types that locate values in text:

```go
type Finder interface {
	FindAllIndex(str []byte, n int) [][]int
}
```

`FindAllIndex` returns `[start, end)` byte offsets, so `str[start:end]` is a match. `n < 0` returns all matches, `n == 0` none, `n > 0` at most `n`. Implementations: `bank.IBANFinder`, `bank.BICFinder`, `*date.Finder` (from `date.NewFinder`), `vat.IDFinder` and `money.AmountFinder`.

## KeyMutex

`KeyMutex[T comparable]` holds one mutex per key. A key's mutex is created by the first `Lock` and deleted when no goroutine holds or waits for it. The zero value is usable.

| Identifier                                 | Description                                        |
|--------------------------------------------|----------------------------------------------------|
| `NewKeyMutex[T comparable]() *KeyMutex[T]` | New mutex. Optional, the zero value works too.     |
| `Lock(key T)`                              | Block until `key` is free, then lock it.           |
| `Unlock(key T)`                            | Unlock `key`. Panics with a message like `KeyMutex[string].Unlock("k") called for non locked key` if `key` is not locked. |
| `IsLocked(key T) bool`                     | `key` is locked or a `Lock` call for it is in progress. Only a snapshot. |

`strutil.StrMutex` and `uu.IDMutex` implement the same algorithm for `string` and `uu.ID` keys. `KeyMutex` is the generic version. `uu.IDMutex` has `IsLocked` too and adds lock and unlock callbacks, `strutil.StrMutex` has no `IsLocked`.

## Reflection helpers

| Function                                      | Description                                        |
|-----------------------------------------------|----------------------------------------------------|
| `CanMarshalJSON(t reflect.Type) bool`         | Static check whether `encoding/json` can encode a value of type `t`. True for interfaces, types implementing `json.Marshaler` or `encoding.TextMarshaler` (also through the pointer type), bool, integer, float and string kinds, arrays, slices, structs, maps with string, integer or text-marshaling keys, and pointers to such types. False for nil, complex, chan, func and `unsafe.Pointer`. |
| `CompareReflectValue(a, b reflect.Value) int` | Three-way comparison for `slices.SortFunc`: numbers by value, strings lexicographically, `false < true`, complex by real then imaginary part, pointers and channels by address, arrays, slices and structs element by element, interfaces nil first, then by dynamic type name, then by value. The invalid zero `reflect.Value` sorts first. 0 for different types and kinds like map or func. |

## Serialization

`Set[T]` has a JSON codec that keeps nil, empty and populated sets apart:

| Go value           | `json.Marshal` | `String()`   |
|--------------------|----------------|--------------|
| `Set[int](nil)`    | `null`         | `<nil>`      |
| `Set[int]{}`       | `[]`           | `[]`         |
| `NewSet(3, 1, 2)`  | `[1,2,3]`      | `[1, 2, 3]`  |
| `NewSet("b", "a")` | `["a","b"]`    | `["a", "b"]` |

`MarshalJSON` writes elements of a normalizing string type, one that implements `types.Normalizable`, as stored, not with their own `MarshalJSON`, so elements whose `MarshalJSON` normalizes, like `money.Currency`, are not merged: `NewSet[money.Currency]("eur", "EUR")` marshals as `["EUR","eur"]` and reads back with both elements. Other elements, also of other string types, are written with their own `MarshalJSON`.

`UnmarshalJSON` turns `null` into a nil set, `[]` into an empty non-nil set, and an array into a new set without duplicates, elements of a normalizing string type as given and other elements with their own `UnmarshalJSON`, so its validation still applies. It always assigns a new map, so a map shared with another variable is never changed. A JSON object is an error.

`*LenString` marshals as its string:

| Go value                                      | `json.Marshal` | `MarshalText` |
|-----------------------------------------------|----------------|---------------|
| `(*LenString)(nil)`                           | `null`         | nil           |
| `NewLenString("abc", 1, 3)`                   | `"abc"`        | `abc`         |
| `LenString` field of a struct passed by value | `{}`           | not called    |

`UnmarshalJSON` accepts a JSON string within the limits and leaves the value unchanged for `null` or an error. `UnmarshalText` accepts any text within the limits.

## Gotchas

- `var s types.Set[int]; s.Insert(1)` panics, and JSON `null` unmarshals to such a nil set. Create sets with `NewSet` or `make`, or check `IsNull` after unmarshalling.
- `T` of `Set[T]` must be `cmp.Ordered`, so `Set` can't hold `uu.ID` or structs. Use `uu.IDSet` or a `map[K]struct{}` with the `mapset` functions.
- `Set` is not safe for concurrent use, like any map.
- `ReduceSet` visits values in undefined order. Only an order independent function like a sum gives a stable result.
- `Set.String()` gives `[1, 2]`, while `mapset.String` gives `{1 2}`.
- A `ValidatErr` method wins over a `Validator` method on the same type, so `Validate` can accept a value whose `Valid()` is false.
- The error for a failed `Validator` contains the value formatted with `%#v`: a pointer shows up as an address, and a secret like a password ends up in the message, for example `invalid main.secret value: "pw"`. Implement `ValidatErr` with your own message for such types.
- A nil func like `types.ValidatorFunc(nil)` counts as valid: `Validate` returns nil for it, and `Validators`, `ValidatErrs` and `DeepValidate` skip it, also if its method could handle nil. So do nil elements of `Validators` and `ValidatErrs`.
- `DeepValidate` calls `Validate` methods with pointer receivers through a pointer to addressable data, so a `Validate` that changes its receiver changes the caller's data when you pass a pointer. Map values and the fields and elements of values passed by value are validated through a copy.
- `LenString` counts bytes: `NewLenString("äö", 0, 3).Validate()` fails because `"äö"` has 4 bytes.
- All `LenString` methods have pointer receivers. A `LenString` field of a struct marshalled by value encodes as `{}`, and `types.MustLenString("abc", 1, 3).String()` does not compile. Use `*LenString` fields or marshal a pointer to the struct.
- The zero `LenString{}` has `min` and `max` 0 and accepts only `""`.
- `KeyMutex` is not reentrant, and `Unlock` of a key that is not locked panics with the key in the message.
- Sorting values of different types with `CompareReflectValue` does not panic, but their order is unspecified, because the result is not a strict weak ordering.
- `CanMarshalJSON` can return true for types that fail at runtime: struct fields, slice elements and map values are not inspected, so `struct{ C chan int }` passes, and pointer-receiver marshalers are not called for values passed by value.

## Examples

Sets and their JSON form:

```go
package main

import (
	"encoding/json"
	"fmt"

	types "github.com/domonda/go-types"
)

func main() {
	a := types.NewSet(3, 1, 2)
	b := types.NewSet(2, 4)
	fmt.Println(a.Insert(5), a.Insert(5))
	fmt.Println(a.Difference(b), a.SymmetricDifference(b), a.Intersects(b))

	for _, s := range []types.Set[int]{nil, {}, a} {
		j, _ := json.Marshal(s)
		fmt.Println(string(j))
	}

	var decoded types.Set[string]
	_ = json.Unmarshal([]byte(`["b","a","b"]`), &decoded)
	fmt.Println(decoded, decoded.Len())
	_ = json.Unmarshal([]byte(`null`), &decoded)
	fmt.Println(decoded.IsNull())
	// Output:
	// true false
	// [1, 3, 5] [1, 3, 4, 5] true
	// null
	// []
	// [1,2,3,5]
	// ["a", "b"] 2
	// true
}
```

Validating a struct with `DeepValidate`:

```go
package main

import (
	"errors"
	"fmt"

	types "github.com/domonda/go-types"
)

type Status string

func (s Status) Valid() bool { return s == "active" || s == "blocked" }

type User struct {
	Name   *types.LenString
	Status Status
	Tags   []Status
}

func main() {
	fmt.Println(types.Validate(Status("unknown")))
	err, ok := types.TryValidate(42)
	fmt.Println(err, ok)

	user := User{
		Name:   types.NewLenString("Alexander", 1, 5),
		Status: "unknown",
		Tags:   []Status{"active", "nope"},
	}
	errs := types.DeepValidate(user)
	for _, err := range errs {
		fmt.Println(err)
	}
	fmt.Println(errors.Join(errs...) != nil)
	// Output:
	// invalid main.Status value: "unknown"
	// <nil> false
	// struct field Name: length 9 of LenString "Alexander" is longer than maximum of 5
	// struct field Status: invalid main.Status value: "unknown"
	// struct field Tags -> element [1]: invalid main.Status value: "nope"
	// true
}
```

Pointers, iterators, `LenString` and the zero `KeyMutex`:

```go
package main

import (
	"encoding/json"
	"fmt"

	types "github.com/domonda/go-types"
)

type Profile struct {
	Nick types.LenString
}

func main() {
	fmt.Println(*new(42), types.FromPtr[int](nil), types.FromPtrOr(nil, "default"))
	for v, err := range types.Seq2NilError(types.Yield("only")) {
		fmt.Println(v, err)
	}

	nick := types.NewLenString("", 2, 8)
	fmt.Println(json.Unmarshal([]byte(`"gopher"`), nick), nick)
	fmt.Println(json.Unmarshal([]byte(`"a"`), nick), nick)

	p := Profile{Nick: types.MustLenString("gopher", 2, 8)}
	byValue, _ := json.Marshal(p)
	byPointer, _ := json.Marshal(&p)
	fmt.Println(string(byValue), string(byPointer))

	var locks types.KeyMutex[string] // usable without NewKeyMutex
	locks.Lock("invoice-1")
	fmt.Println(locks.IsLocked("invoice-1"))
	locks.Unlock("invoice-1")
	fmt.Println(locks.IsLocked("invoice-1"))
	// Output:
	// 42 0 default
	// only <nil>
	// <nil> gopher
	// length 1 of LenString "a" is shorter than minimum of 2 gopher
	// {"Nick":{}} {"Nick":"gopher"}
	// true
	// false
}
```

## Related

- [Validation and normalization explained](../docs/explanation-validation-and-normalization.md)
- [How to validate structs](../docs/howto-validate-structs.md)
- [Sets: nil, empty and the collections proposal](../docs/explanation-sets.md)
- [`mapset`](../mapset/README.md): the set functions behind `Set[T]`, `uu.IDSet`, `strutil.StringSet` and `email.AddressSet`.
- [`set`](../set/README.md): older helper functions for `map[T]struct{}`.
- [`uu`](../uu/README.md): `uu.IDSet` for UUIDs and `uu.IDMutex`.
- [`strutil`](../strutil/README.md): `StringSet`, `StrMutex` and `SplitAndTrimIndex`, which several finders use.
- [`nullable`](../nullable/README.md): the `nullable.Nullable` interface that `Set[T]` implements.
- [`strfmt`](../strfmt/README.md): the `strfmt.Scannable` interface that `*LenString` implements.
- Packages with `Finder` implementations: [`bank`](../bank/README.md), [`date`](../date/README.md), [`vat`](../vat/README.md), [`money`](../money/README.md).
