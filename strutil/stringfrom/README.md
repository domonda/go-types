# stringfrom

Convert pointers and other values to strings for display, with a chosen text for nil. Use it to render optional fields in templates, CSV files or log messages without writing a nil check for every field.

```
import "github.com/domonda/go-types/strutil/stringfrom"
```

## Pointer functions

Every `*Ptr` function takes a pointer and optional `strForNil ...string` arguments. For a nil pointer it returns the `strForNil` arguments joined without a separator, so `Ptr(nil)` is `""`, `Ptr(nil, "n/a")` is `"n/a"` and `Ptr(nil, "<", "nil", ">")` is `"<nil>"`.

| Function                                           | Non-nil result                                     | Returns `strForNil` for                            |
|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|
| `Ptr(ptr *string, strForNil ...string) string`     | `*ptr`                                             | nil                                                |
| `TimePtr(ptr *time.Time, strForNil ...string) string` | `ptr.String()`, like `"2024-03-01 12:30:00 +0000 UTC"` | nil and a pointer to the zero time                 |
| `UUIDPtr(ptr *uu.ID, strForNil ...string) string`  | `ptr.String()`, the canonical UUID form            | nil only, a pointer to `uu.IDNil` gives `"00000000-0000-0000-0000-000000000000"` |
| `IntPtr(ptr *int, strForNil ...string) string`     | decimal, like `"-3"`                               | nil                                                |
| `UintPtr(ptr *uint, strForNil ...string) string`   | decimal, like `"7"`                                | nil                                                |
| `Int64Ptr(ptr *int64, strForNil ...string) string` | decimal                                            | nil                                                |
| `Float64Ptr(ptr *float64, strForNil ...string) string` | `strconv.FormatFloat(*ptr, 'f', -1, 64)`: no exponent, like `"0.1"` or `"1000000000000000000000"` for `1e21`; `"NaN"`, `"+Inf"`, `"-Inf"` | nil                                                |
| `BoolPtr(ptr *bool, strForNil ...string) string`   | `"true"` or `"false"`                              | nil                                                |

## Other functions

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `Bool(boolVal bool, trueString, falseString string) string` | `trueString` if `boolVal` is true, else `falseString`. |
| `Interface(i any) string`                          | Converts `i`, or the value `i` points to: a value of string kind as is, any other value with `fmt.Sprint`. A non-pointer value gives the same string as a pointer to it. Returns `""` for nil, a nil pointer, and a nil map, slice, func, channel or interface value. |

None of the functions panics for nil input.

## Gotchas

- `TimePtr` uses `time.Time.String`, which is not RFC 3339 and includes the monotonic clock reading for times from `time.Now()`, like `"2026-10-12 14:15:32.261641 +0200 CEST m=+0.000682459"`. Call `Round(0)` on the time or format it yourself if you need a stable layout.
- `UUIDPtr` does not treat `uu.IDNil` as nil, unlike `TimePtr` with the zero time.
- `Interface` dereferences one pointer level before formatting, so methods with a pointer receiver are lost: a type whose `String` method has a pointer receiver prints as its fields, like `"{1}"`, and `Interface(errors.New("boom"))` gives `"{boom}"`. Pass `err.Error()` or call `fmt.Sprint` yourself for such values.
- `Interface` returns values of string kind without calling their `String` method: for `type color string` with a `String` method, `Interface(color("red"))` is `"red"`.
- `Interface` dereferences only one level: a pointer to a non-nil pointer prints as an address.

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/domonda/go-types/strutil/stringfrom"
	"github.com/domonda/go-types/uu"
)

func main() {
	var name *string
	fmt.Println(stringfrom.Ptr(name, "<unknown>"))

	paid := time.Date(2024, 3, 1, 12, 30, 0, 0, time.UTC)
	var zero time.Time
	fmt.Println(stringfrom.TimePtr(&paid, "unpaid"))
	fmt.Println(stringfrom.TimePtr(&zero, "unpaid"))

	fmt.Println(stringfrom.UUIDPtr(nil, "no ID"))
	fmt.Println(stringfrom.UUIDPtr(&uu.IDNil, "no ID"))

	amount := 1234.5
	fmt.Println(stringfrom.Float64Ptr(&amount))
	fmt.Println(stringfrom.Bool(true, "yes", "no"))

	n := 42
	fmt.Printf("%q %q %q\n", stringfrom.Interface(&n), stringfrom.Interface(n), stringfrom.Interface(nil))
	// Output:
	// <unknown>
	// 2024-03-01 12:30:00 +0000 UTC
	// unpaid
	// no ID
	// 00000000-0000-0000-0000-000000000000
	// 1234.5
	// yes
	// "42" "42" ""
}
```

## Related

- [`strutil`](../README.md): string helpers of the parent package, including `Ptr`, `DerefPtr` and `StringToPtrEmptyToNil` for `*string`.
- [Root package reference](../../docs/reference-types.md): `types.Ptr`, `types.FromPtr` and `types.FromPtrOr` for pointers of any type.
- [`uu`](../../uu/README.md): the `uu.ID` type formatted by `UUIDPtr`.
- [`nullable`](../../nullable/README.md): optional values as types instead of pointers.
- [How to store optional values](../../docs/howto-store-optional-values.md)
