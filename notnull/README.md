# notnull

Never-null counterparts of some [`nullable`](../nullable/) types for `NOT NULL` columns and JSON fields that must not be `null`. The zero value of every type is written as a value: a nil array as the empty SQL array `'{}'` and JSON `[]`, a nil `JSON` as the empty object `{}`, and `TrimmedString` as a possibly empty string. Use it when a column is `NOT NULL DEFAULT '{}'` and the Go code should not need nil checks.

```
import "github.com/domonda/go-types/notnull"
```

## Why a nil slice maps to an empty array

`nil` is the zero value of a Go slice, what a slice variable or struct field holds before anything is assigned to it. Because this package models values that are `NOT NULL`, that default must not become SQL `NULL` or JSON `null`. The arrays treat it as the empty array instead:

- `Value` and `MarshalJSON` write a nil or empty slice as `'{}'` and `[]`, never as `NULL` or `null`.
- `Scan` and `UnmarshalJSON` never produce a nil slice: SQL `NULL`, `'{}'`, JSON `null` and `[]` all read as a non-nil empty slice.

To keep `NULL` apart from the empty array, use the `nullable` arrays instead.

## Arrays

| Type            | Element        | Methods besides `Value`, `Scan`, `MarshalJSON` and `UnmarshalJSON` |
|-----------------|----------------|----------------------------------------------------|
| `StringArray`   | `string`       | `Contains`, `Len`, `Less`, `Swap`                  |
| `IntArray`      | `int64`        | `Contains`, `String`, `Len`, `Less`, `Swap`        |
| `FloatArray`    | `float64`      | `Contains`, `String`, `Len`, `Less`, `Swap`        |
| `NullBoolArray` | `sql.NullBool` | `Bools` (NULL elements as `false`), `String`       |

| Method                              | Description                                        |
|-------------------------------------|----------------------------------------------------|
| `Contains(value)`                   | Whether `value` is an element.                     |
| `Len()`, `Less(i, j)`, `Swap(i, j)` | `sort.Interface` in ascending element order.       |
| `String()`                          | `[1, 2]` for `IntArray` and `FloatArray`, `[true, NULL]` for `NullBoolArray`. A nil array prints `[]`. `StringArray` has no `String` method. |
| `Bools()`                           | `[]bool` with NULL elements as `false`. Returns nil for an empty array. |
| `Value()`                           | PostgreSQL array literal: `{"a","b"}` with quoted and escaped elements for `StringArray`, `{1,2}`, `{1.5,2}`, `{t,f,NULL}`. nil and empty write `{}`. |
| `Scan(src)`                         | Parses a PostgreSQL array literal from a `string` or `[]byte`. nil and `{}` read as an empty slice. |
| `MarshalJSON()`                     | JSON array, `[]` for nil and empty. NULL elements of `NullBoolArray` as `null`. A NaN or ±Inf element of `FloatArray` returns a `*json.UnsupportedValueError`, like `encoding/json` for a `float64`. |
| `UnmarshalJSON(data)`               | JSON array, `null` reads as an empty slice. `null` elements of `NullBoolArray` read as NULL. |

Every array implements `driver.Valuer`, `sql.Scanner`, `json.Marshaler` and `json.Unmarshaler`. `IntArray`, `FloatArray` and `NullBoolArray` also implement `fmt.Stringer`.

`Value` and `Scan` use the [PostgreSQL array text format](https://www.postgresql.org/docs/current/arrays.html), which PostgreSQL, CockroachDB and YugabyteDB understand. Databases without array types, like MySQL, MariaDB, SQLite, SQL Server and Oracle, are not supported.

## Array helpers

| Function                  | Description                                        |
|---------------------------|----------------------------------------------------|
| `SplitArray(array)`       | Splits an SQL or JSON array literal into its top level elements and leaves quoted elements quoted. `{}`, `[]`, `null` and `NULL` return a non-nil empty slice. An unquoted SQL `NULL` element is returned as `NULL`. A trailing comma yields an empty last element. |
| `SplitArrayValues(array)` | Like `SplitArray`, but returns the value of quoted elements with quotes and escapes removed. Every SQL element is unescaped, in a JSON array only quoted strings are. |
| `SQLArrayLiteral(strs)`   | Joins `strs` as a quoted PostgreSQL array literal like `{"a","b"}`. nil and an empty slice return `{}`. |

## TrimmedString

`type TrimmedString string` trims every value with `strutil.TrimSpace`, which removes Unicode whitespace, non-printable runes like the BOM and zero-width characters, and invalid UTF-8 bytes. Unlike `nullable.TrimmedString` it has no null state: the empty string is a value and is written as `''` and `""`.

| Function or method                                 | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `TrimmedStringFrom(str)`                           | Trimmed `str`.                                     |
| `TrimmedStringf(format, a...)`                     | Trimmed `fmt.Sprintf` result.                      |
| `TrimmedStringJoin(sep, strs...)`                  | Trims every string of any `~string` type and joins all of them with `sep`, empty ones included. |
| `String()`                                         | The trimmed string.                                |
| `IsEmpty()`                                        | Whether the trimmed string is empty.               |
| `Set(str)`                                         | Assigns the trimmed `str`.                         |
| `ToUpper()`, `ToLower()`                           | Case mapping of the trimmed string.                |
| `ToValidUTF8(replacement)`                         | Replaces invalid UTF-8 and trims the result.       |
| `Contains(substr)`, `ContainsAny(chars)`, `ContainsRune(r)` | `strings` functions on the trimmed string.         |
| `HasPrefix(prefix)`, `HasSuffix(suffix)`           | `strings` functions on the trimmed string.         |
| `TrimPrefix(prefix)`, `TrimSuffix(suffix)`, `ReplaceAll(old, new)` | `strings` functions, the result trimmed again.     |
| `Split(sep)`                                       | `strings.Split` of the trimmed string with every part trimmed. |

Implements: `fmt.Stringer`, `driver.Valuer`, `sql.Scanner`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `json.Marshaler`, `json.Unmarshaler`, `xml.Marshaler`, `xml.Unmarshaler`, `xml.MarshalerAttr` (pointer). It has no `JSONSchema` method.

## JSON

`type JSON []byte` holds raw JSON text. A nil or empty `JSON` stands for the empty object `{}`.

| Function or method    | Description                                        |
|-----------------------|----------------------------------------------------|
| `MarshalJSON(source)` | `json.Marshal(source)` as `JSON`.                  |
| `MarshalFrom(source)` | Marshals `source` into the receiver. A nil `source` stores `null`. |
| `UnmarshalTo(dest)`   | `json.Unmarshal` into `dest`. A nil or empty `JSON` unmarshals `{}`. |
| `IsEmpty()`           | True for nil, `""`, `{}` and `[]`.                 |
| `Valid()`             | `json.Valid`, but true for nil and empty.          |
| `String()`            | The text, or `{}` for nil and empty.               |
| `GoString()`          | ``notnull.JSON(`...`)``.                           |
| `PrettyPrint(w)`      | The text in backticks, `{}` for nil and empty, for `github.com/domonda/go-pretty`. |
| `Clone()`             | Copy. nil stays nil.                               |

Implements: `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`, `fmt.Stringer`, `fmt.GoStringer`.

## Serialization

| Go value                                       | SQL `Value()`      | JSON          | Text            |
|------------------------------------------------|--------------------|---------------|-----------------|
| `StringArray(nil)`, `StringArray{}`            | `'{}'`             | `[]`          | no text methods |
| `StringArray{"a", "b c"}`                      | `'{"a","b c"}'`    | `["a","b c"]` | no text methods |
| `IntArray(nil)`, `IntArray{}`                  | `'{}'`             | `[]`          | no text methods |
| `IntArray{1, 2}`                               | `'{1,2}'`          | `[1,2]`       | no text methods |
| `NullBoolArray{{Valid: true, Bool: true}, {}}` | `'{t,NULL}'`       | `[true,null]` | no text methods |
| `TrimmedString("")`                            | `''`               | `""`          | empty           |
| `TrimmedString(" x ")`                         | `'x'`              | `"x"`         | `x`             |
| `JSON(nil)`, `JSON("")`                        | `'{}'` as `[]byte` | `{}`          | no text methods |

| Input          | Arrays                                  | `TrimmedString` | `JSON`              |
|----------------|-----------------------------------------|-----------------|---------------------|
| SQL `NULL`     | empty slice                             | error           | `{}`                |
| SQL `''`       | empty slice, error for `StringArray`    | `""`            | `{}`                |
| SQL `'{}'`     | empty slice                             | the string `{}` | `{}`                |
| `NULL` element | error, NULL element for `NullBoolArray` | not an array    | not an array        |
| JSON `null`    | empty slice                             | `""`            | the text `null`     |
| JSON `[]`      | empty slice                             | error           | the text `[]`       |

## Gotchas

- Only the arrays turn SQL `NULL` into an empty value. `TrimmedString.Scan(nil)` returns `can't scan <nil> as notnull.TrimmedString`, and `JSON.Scan(nil)` stores `{}`.
- JSON `null` is not mapped to `{}` by `JSON`: `UnmarshalJSON` stores the text `null`, which `Value` then writes as the JSON value `null`.
- `JSON(nil)` is written as `{}`, an object, not as `[]`, and `JSON(nil).Valid()` is true because it stands for the valid `{}`.
- `StringArray.Scan("")` returns an error, the other arrays return an empty slice.
- `NULL` elements fail in `StringArray`, `IntArray` and `FloatArray`: `IntArray.Scan("{1,NULL}")` returns an error. Only `NullBoolArray` has nullable elements, and it accepts `t`, `f`, `true`, `false` and `NULL` in any letter case.
- `notnull.NullBoolArray` holds `sql.NullBool` elements, `nullable.NullBoolArray` holds `nullable.Type[bool]`. There is no `notnull.BoolArray`, `NullIntArray` or `NullFloatArray`.
- `notnull.TrimmedStringJoin` keeps empty elements, `nullable.TrimmedStringJoin` skips them: `TrimmedStringJoin(",", "a", " ", "b")` returns `a,,b` here and `a,b` there.
- Switching a field between `notnull` and `nullable` changes what a nil slice writes: `'{}'` here, `NULL` there.

## Examples

A struct with never-null fields:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/notnull"
)

type Document struct {
	Tags    notnull.StringArray   `json:"tags"`
	Pages   notnull.IntArray      `json:"pages"`
	Title   notnull.TrimmedString `json:"title"`
	Payload notnull.JSON          `json:"payload"`
}

func main() {
	// The zero value is written as empty arrays and an empty object, never null
	var doc Document
	j, _ := json.Marshal(doc)
	fmt.Println(string(j))

	tags, _ := doc.Tags.Value()
	payload, _ := doc.Payload.Value()
	fmt.Printf("%#v %s\n", tags, payload)

	// SQL NULL and JSON null read as empty, non-nil arrays
	var in Document
	err := json.Unmarshal([]byte(`{"tags":null,"pages":[1,2],"title":"  Report  "}`), &in)
	fmt.Println(err, in.Tags != nil, len(in.Tags), in.Pages, in.Title)
	fmt.Println(in.Pages.Scan(nil), in.Pages != nil, len(in.Pages))

	fmt.Printf("%q\n", notnull.TrimmedStringJoin(",", "a", " ", "b"))

	// Output:
	// {"tags":[],"pages":[],"title":"","payload":{}}
	// "{}" {}
	// <nil> true 0 [1, 2] Report
	// <nil> true 0
	// "a,,b"
}
```

Array literals:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/notnull"
)

func main() {
	elems, err := notnull.SplitArray(`{a,"b,c",NULL}`)
	fmt.Printf("%q %v\n", elems, err)
	values, err := notnull.SplitArrayValues(`{a,"b,c",NULL}`)
	fmt.Printf("%q %v\n", values, err)
	empty, err := notnull.SplitArray("NULL")
	fmt.Printf("%q %t %v\n", empty, empty != nil, err)

	fmt.Println(notnull.SQLArrayLiteral(nil))
	fmt.Println(notnull.SQLArrayLiteral([]string{"a", `b"c`}))

	var tags notnull.StringArray
	err = tags.Scan(notnull.SQLArrayLiteral([]string{"a", `b"c`}))
	fmt.Printf("%q %v\n", tags, err)

	// Output:
	// ["a" "\"b,c\"" "NULL"] <nil>
	// ["a" "b,c" "NULL"] <nil>
	// [] true <nil>
	// {}
	// {"a","b\"c"}
	// ["a" "b\"c"] <nil>
}
```

## Related

- [nullable](../nullable/): the counterparts where nil is SQL `NULL` and JSON `null`.
- [Null conventions](../docs/explanation-null-conventions.md): when to choose `notnull` or `nullable`.
- [How to store optional values](../docs/howto-store-optional-values.md)
