# nullable

Nullable wrapper types for `database/sql` and `encoding/json`, and the interfaces that describe nullable values across this module. Every type has one null state that round-trips as SQL `NULL` and JSON `null`, without pointer fields. Use it for optional strings, times, numbers, JSON documents and PostgreSQL arrays.

```
import "github.com/domonda/go-types/nullable"
```

## Why "zero == null"?

Go has no nullable value types, and `*string` or `*time.Time` fields are awkward to build and compare. This package uses the zero value of a type as its null state where the zero value is never a meaningful value: `""` for the string types, `time.Time{}` for `Time`, a nil slice for the arrays. Methods like `IsNull`, `Get`, `GetOr`, `Set` and `SetNull` make the convention explicit.

Where the zero value is a legal value, like `0` or `false`, use `Type[T]`, a generic struct with a separate valid flag. The domain packages of this module follow the same split: `uu.NullableID` or `date.NullableDate` use the zero value as null, `money.NullableAmount` is a `Type[money.Amount]`.

## Interfaces and functions

| Identifier                       | Description                                        |
|----------------------------------|----------------------------------------------------|
| `Nullable`                       | `IsNull() bool`. Implemented by every type of this package except the `StringArray` and `BoolArray` aliases, and by the `NullableX` types of the domain packages. |
| `NullSetable[T]`                 | `Nullable` plus `SetNull()`, `Set(T)`, `Get() T` (panics if null) and `GetOr(T) T`. Implemented by `*Type[T]`, `*Time` (with `time.Time`) and pointers to the `NullableX` types of the domain packages, like `*uu.NullableID` (with `uu.ID`) and `*date.NullableDate` (with `date.Date`). Not implemented by the two string types of this package, which have no `GetOr`. |
| `Zeroable`                       | `IsZero() bool`, like `time.Time`.                 |
| `NullableValidator`              | `Nullable` plus `types.Validator` plus `ValidAndNotNull() bool`, for nullable types whose non-null values can be invalid. Implemented by types like `bank.NullableIBAN`, `date.NullableDate` and `vat.NullableID`. |
| `ReflectIsNull(v reflect.Value)` | True for the zero `reflect.Value`, for a nil pointer, interface, map, slice, channel, function or `unsafe.Pointer` anywhere along a chain of pointers and interfaces, and for values whose `IsNull()` returns true. If `Nullable` is not implemented, `IsZero()` of `Zeroable` decides. Never panics. |
| `IsNull(v any)`                  | `ReflectIsNull(reflect.ValueOf(v))`. Reports a typed nil pointer in an `any` as null. |

## Type[T]

`Type[T any]` wraps a value of any type with a valid flag. The zero `Type[T]{}` is null, `TypeFrom(0)` is a valid zero.

| Function or method        | Description                                        |
|---------------------------|----------------------------------------------------|
| `TypeFrom(value)`         | Valid value.                                       |
| `TypeFromPtr(ptr)`        | Null for a nil pointer, otherwise a valid copy of `*ptr`. |
| `IsNull()`, `IsNotNull()` | Whether the value is null.                         |
| `Get()`                   | The value, panics with `Get() called on NULL <T>` if null. |
| `GetOr(defaultValue)`     | The value, or `defaultValue` if null.              |
| `Ptr()`                   | Pointer to a copy of the value, or nil if null.    |
| `Set(value)`              | Sets a valid value.                                |
| `SetNull()`               | Sets null and resets the value to the zero value of `T`. |
| `Value()`                 | nil if null. Otherwise `T`'s own `Value()` if `T` implements `driver.Valuer`, else `driver.DefaultParameterConverter`, which accepts booleans, numbers, strings, `[]byte` and `time.Time` but rejects structs and `uint64` values of 2⁶³ and above. |
| `Scan(src)`               | nil sets null. Other values are converted with a copy of the `database/sql` conversion rules, or with `T`'s `Scan` if `*T` implements `sql.Scanner`. Copies `[]byte` input. |
| `MarshalJSON()`           | `null`, or the JSON of the value.                  |
| `UnmarshalJSON(data)`     | `null` or empty input sets null, anything else is unmarshalled into a valid value. |
| `JSONSchema()`            | The schema of `T` as one of the options next to `null`. |

Implements: `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`, `Nullable`, and with a pointer receiver `NullSetable[T]`. `Type[T]` has no `String` method.

## NonEmptyString

`type NonEmptyString string` treats the empty string as null. It does not trim, so `" "` is a non-null value.

| Function or method                  | Description                                      |
|-------------------------------------|--------------------------------------------------|
| `NonEmptyStringNull`                | The null value `""`.                             |
| `NonEmptyStringf(format, a...)`     | `fmt.Sprintf` result as `NonEmptyString`.        |
| `NonEmptyStringFromPtr(ptr)`        | Null for a nil pointer, otherwise `*ptr`.        |
| `NonEmptyStringFromError(err)`      | Null for a nil error, otherwise `err.Error()`.   |
| `NonEmptyStringTrimSpace(str)`      | `str` trimmed with `strutil.TrimSpace`.          |
| `JoinNonEmptyStrings(sep, strs...)` | Joins the non-null strings with `sep`.           |
| `IsNull()`, `IsNotNull()`           | Whether the string is `""`.                      |
| `Get()`                             | The string, panics if null.                      |
| `StringOr(nullString)`              | The string, or `nullString` if null.             |
| `Ptr()`                             | Pointer to a copy of the string, or nil if null. |
| `Set(s)`, `SetNull()`               | Assign a string or `""`.                         |
| `TrimSpace()`                       | The string trimmed with `strutil.TrimSpace`.     |

Implements: `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextUnmarshaler` (no `MarshalText`), `Nullable`.

## TrimmedString

`type TrimmedString string` trims every value with `strutil.TrimSpace`, which removes Unicode whitespace, non-printable runes like the BOM and zero-width characters, and invalid UTF-8 bytes. A string that is empty after trimming is null.

| Function or method                                 | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `TrimmedStringNull`                                | The null value `""`.                               |
| `TrimmedStringFrom(str)`                           | Trimmed `str`.                                     |
| `TrimmedStringf(format, a...)`                     | Trimmed `fmt.Sprintf` result.                      |
| `TrimmedStringFromPtr(ptr)`                        | Null for a nil pointer, otherwise trimmed `*ptr`.  |
| `TrimmedStringFromError(err)`                      | Null for a nil error, otherwise trimmed `err.Error()`. |
| `TrimmedStringJoin(sep, strs...)`                  | Trims every string of any `~string` type and joins the non-empty ones with `sep`. |
| `IsNull()`, `IsNotNull()`                          | Whether the trimmed string is empty.               |
| `String()`                                         | The trimmed string, `""` if null.                  |
| `StringOr(nullString)`                             | The trimmed string, or `nullString` if null.       |
| `Get()`                                            | The trimmed string, panics if null.                |
| `Ptr()`                                            | Pointer to a copy of the underlying string, or nil if null. |
| `Set(str)`, `SetNull()`                            | Assign a trimmed string or `""`.                   |
| `ToUpper()`, `ToLower()`                           | Case mapping of the trimmed string.                |
| `ToValidUTF8(replacement)`                         | Replaces invalid UTF-8 and trims the result.       |
| `Contains(substr)`, `ContainsAny(chars)`, `ContainsRune(r)` | `strings` functions on the trimmed string.         |
| `HasPrefix(prefix)`, `HasSuffix(suffix)`           | `strings` functions on the trimmed string.         |
| `TrimPrefix(prefix)`, `TrimSuffix(suffix)`, `ReplaceAll(old, new)` | `strings` functions, the result trimmed again.     |
| `Split(sep)`                                       | `strings.Split` of the trimmed string with every part trimmed. |
| `JSONSchema()`                                     | One of a string or `null`.                         |

Implements: `fmt.Stringer`, `driver.Valuer`, `sql.Scanner`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `json.Marshaler`, `json.Unmarshaler`, `xml.Marshaler`, `xml.Unmarshaler`, `xml.MarshalerAttr` (pointer), `Nullable`.

## Time

`type Time struct{ time.Time }` treats the zero `time.Time` as null. `IsNull` uses `time.Time.IsZero`, so the zero instant in any location is null.

| Function or method                        | Description                                        |
|-------------------------------------------|----------------------------------------------------|
| `TimeNull`                                | The null value `Time{}`. Compare with `IsNull()`, not `==`. |
| `TimeNow()`                               | The current time.                                  |
| `TimeFrom(t)`                             | Wraps `t`. A zero `t` is null.                     |
| `TimeFromPtr(ptr)`                        | Null for a nil pointer, otherwise `*ptr`.          |
| `TimeParse(layout, value)`                | `time.Parse`, but `""`, `"null"` and `"NULL"` return null without an error. |
| `TimeParseInLocation(layout, value, loc)` | `time.ParseInLocation` with the same null strings. |
| `IsNull()`, `IsNotNull()`                 | Whether the time is zero.                          |
| `Get()`                                   | The `time.Time`, panics if null.                   |
| `GetOr(defaultTime)`                      | The `time.Time`, or `defaultTime` if null.         |
| `Ptr()`                                   | Pointer to a copy of the `time.Time`, or nil if null. |
| `Set(t)`, `SetNull()`                     | Assign a time (a zero `t` is null) or null.        |
| `UTC()`, `Add(d)`, `AddDate(y, m, d)`     | Like the `time.Time` methods, but return a `Time` and keep null as null. |
| `Equal(o)`                                | Same instant, or both null.                        |
| `Format(layout)`                          | `time.Time.Format`, or `""` if null.               |
| `AppendFormat(b, layout)`                 | `time.Time.AppendFormat`, or `b` unchanged if null. |
| `String()`                                | `time.Time.String()`, or `NULL` if null.           |
| `StringOr(nullStr)`                       | `time.Time.String()`, or `nullStr` if null.        |
| `PrettyPrint(w)`                          | `NULL` or the time, for `github.com/domonda/go-pretty`. |
| `JSONSchema()`                            | One of a `date-time` string or `null`.             |

All other `time.Time` methods are promoted from the embedded field. They don't know about null and return plain `time.Time` values: `nullable.Time{}.In(loc)` is a zero `time.Time`.

Implements: `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `xml.MarshalerAttr` (pointer), `fmt.Stringer`, `Nullable`, `Zeroable`, and with a pointer receiver `NullSetable[time.Time]`.

## JSON

`type JSON []byte` holds raw JSON text, with nil as the null value. A non-nil empty `JSON` is null too: `IsNull` reports it, `Value` writes SQL `NULL` and `MarshalJSON` writes `null`.

| Function or method    | Description                                        |
|-----------------------|----------------------------------------------------|
| `MarshalJSON(source)` | `json.Marshal(source)` as `JSON`. A nil `source` returns `JSON("null")`, which is not nil. |
| `MarshalFrom(source)` | Marshals `source` into the receiver and sets nil if the result is `null`. |
| `UnmarshalTo(dest)`   | `json.Unmarshal` into `dest`. A nil `JSON` returns an error. |
| `IsNull()`            | Whether the value is nil or empty.                 |
| `IsEmpty()`           | True for nil, `""`, `{}` and `[]`.                 |
| `Valid()`             | True for nil and empty, otherwise `json.Valid`.    |
| `String()`            | The text, or `null` for nil and empty.             |
| `GoString()`          | ``nullable.JSON(`...`)``.                          |
| `PrettyPrint(w)`      | The text in backticks, for `github.com/domonda/go-pretty`. |
| `Clone()`             | Copy. nil stays nil.                               |

Implements: `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`, `fmt.Stringer`, `fmt.GoStringer`, `Nullable`.

## Arrays

PostgreSQL array types where only a nil slice is SQL `NULL` and JSON `null`, so `NULL` and the empty array `'{}'` stay distinct through a round trip. Use [`notnull`](../notnull/) for `NOT NULL` columns where a nil slice should be written as `'{}'`.

| Type             | Element         | Methods besides `Value` and `Scan`                 |
|------------------|-----------------|----------------------------------------------------|
| `StringArray`    | `string`        | none, alias of the module's internal copy of `lib/pq`'s `StringArray` |
| `BoolArray`      | `bool`          | none, alias of the internal copy of `lib/pq`'s `BoolArray` |
| `IntArray`       | `int64`         | `IsNull`, `Contains`, `String`, `Len`, `Less`, `Swap` |
| `FloatArray`     | `float64`       | `IsNull`, `Contains`, `String`, `StringOr`, `Len`, `Less`, `Swap` |
| `NullIntArray`   | `Type[int64]`   | `IsNull`, `Ints` (null as `0`), `String`           |
| `NullFloatArray` | `Type[float64]` | `IsNull`, `Floats` (null as `0`), `String`         |
| `NullBoolArray`  | `Type[bool]`    | `IsNull`, `Bools` (null as `false`), `String`      |

None of the array types has JSON methods. `encoding/json` encodes them as plain slices, so nil is `null`, an empty slice `[]`, and a null element of the `NullX` arrays `null`. `Ints`, `Floats` and `Bools` return nil for an empty array.

| Function                  | Description                                        |
|---------------------------|----------------------------------------------------|
| `SplitArray(array)`       | Splits an SQL or JSON array literal into its top level elements and leaves quoted elements quoted. `{}`, `[]`, `null` and `NULL` return nil. An unquoted SQL `NULL` element is returned as `NULL`. A trailing comma yields an empty last element. |
| `SplitArrayValues(array)` | Like `SplitArray`, but returns the value of quoted elements with quotes and escapes removed. Every SQL element is unescaped, in a JSON array only quoted strings are. A quoted `"NULL"` and an unquoted `NULL` both return `NULL`. |
| `SQLArrayLiteral(strs)`   | Joins `strs` as a quoted PostgreSQL array literal like `{"a","b"}`. nil returns `NULL`, an empty slice `{}`. |

`Value` and `Scan` use the [PostgreSQL array text format](https://www.postgresql.org/docs/current/arrays.html), which PostgreSQL, CockroachDB and YugabyteDB understand. Databases without array types, like MySQL, MariaDB, SQLite, SQL Server and Oracle, are not supported.

## Serialization

| Type                  | Null value       | SQL `Value()` of null and of a value         | JSON of null and of a value | Text of null     |
|-----------------------|------------------|----------------------------------------------|-----------------------------|------------------|
| `Type[T]`             | `Type[T]{}`      | `NULL`, converted `T`                        | `null`, JSON of `T`         | no text methods  |
| `NonEmptyString`      | `""`             | `NULL`, the string                           | `null`, string              | no `MarshalText` |
| `TrimmedString`       | trims to `""`    | `NULL`, trimmed string                       | `null`, trimmed string      | empty            |
| `Time`                | zero `time.Time` | `NULL`, `time.Time`                          | `null`, RFC 3339 string     | `NULL`           |
| `JSON`                | nil or empty     | `NULL`, the bytes                            | `null`, the raw JSON        | no text methods  |
| `IntArray` and others | nil slice        | `NULL`; `'{}'` for an empty slice, `'{1,2}'` | `null`; `[]`, `[1,2]`       | no text methods  |

Reading:

| Input                 | `Type[T]`                                          | `NonEmptyString`  | `TrimmedString`   | `Time`                  | `JSON`                |
|-----------------------|----------------------------------------------------|-------------------|-------------------|-------------------------|-----------------------|
| SQL `NULL`            | null                                               | null              | null              | null                    | nil                   |
| SQL `''`              | valid `""` for `Type[string]`, error for `Type[int]` | error             | null              | error                   | nil                   |
| SQL text value        | converted                                          | the string        | trimmed           | error, only `time.Time` | copied, `null` is nil |
| JSON `null`           | null                                               | null              | null              | null                    | nil                   |
| JSON `""`             | valid `""` for `Type[string]`, error for `Type[int]` | null              | null              | error                   | `JSON("\"\"")`        |
| Text empty            | no text methods                                    | null              | null              | null                    | no text methods       |
| Text `null`, any case | no text methods                                    | the string `null` | the string `null` | null                    | no text methods       |

| Input          | `StringArray` | `BoolArray` | `IntArray`, `FloatArray` | `NullIntArray` and others |
|----------------|---------------|-------------|--------------------------|---------------------------|
| SQL `NULL`     | nil           | nil         | nil                      | nil                       |
| SQL `''`       | error         | error       | empty slice              | empty slice               |
| SQL `'{}'`     | empty slice   | empty slice | empty slice              | empty slice               |
| `NULL` element | error         | error       | error                    | null element              |
| JSON `null`    | nil           | nil         | nil                      | nil                       |
| JSON `[]`      | empty slice   | empty slice | empty slice              | empty slice               |

## Gotchas

- `Type[string]` keeps JSON `""` as a valid empty string, it is not null. Only `null` is.
- `omitempty` never omits a `Type[T]` or `Time`, because they are structs. Use `omitzero`: `Type[T]{}` and the zero `Time` are then left out.
- `Type[T]` has no `String` method, so `fmt.Println(nullable.TypeFrom(42))` prints `{42 true}`.
- `Type[T].Value()` fails for a struct `T` that does not implement `driver.Valuer` (`unsupported type main.S, a struct`) and for `uint64` values of 2⁶³ and above.
- A failed `Type[T].Scan` keeps the previous state: a null value stays null, a valid value keeps its old value.
- `NonEmptyString` does not trim, so `" "` is written as `' '`. `NonEmptyString.Scan("")` returns an error, while `TrimmedString.Scan("")` sets null.
- `TrimmedString` created by a plain conversion like `TrimmedString(" x ")` is trimmed by every method except `Ptr`, which points to a copy of the untrimmed underlying string. Use `TrimmedStringFrom` or `Set`.
- `Time.Scan` accepts only `time.Time` and nil. A driver that returns timestamps as text gives `can't scan string as nullable.Time`.
- `Time` writes null as `NULL` in text form, `TrimmedString` writes it as empty text.
- `Time.UnmarshalJSON` returns an error for `""`. `Time.UnmarshalText` treats `null` in any letter case as null, `TimeParse` only `null` and `NULL`.
- The promoted `time.Time` methods ignore null: `nullable.Time{}.Before(t)` compares the zero instant.
- `MarshalJSON(nil)` returns `JSON("null")`, which is not nil, so `IsNull()` is false. `MarshalFrom(nil)` sets nil.
- `IntArray` and `FloatArray` reject `NULL` elements: `{1,NULL,3}` fails, use `NullIntArray` or `NullFloatArray`.
- `StringArray.Scan("")` and `BoolArray.Scan("")` return an error, the other arrays return an empty slice.
- `StringArray.Scan` follows `lib/pq` and reads the unquoted `{a\,b}` as two elements, PostgreSQL reads it as `a,b`. `SplitArrayValues` follows PostgreSQL.
- `String()` formats differ: `IntArray{1, 2}` prints `IntArray{1,2}` and nil prints `IntArray<nil>`, `FloatArray{1.5, 2}` prints `[1.5, 2]` and nil prints `NULL`.
- `NullBoolArray.Scan` accepts `t`, `f`, `true`, `false` and `NULL` in any letter case and returns an error for any other element.

## Examples

`Type[T]`, `TrimmedString` and `Time` in a JSON document:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/nullable"
)

type Product struct {
	Discount nullable.Type[int]     `json:"discount"`
	Comment  nullable.TrimmedString `json:"comment"`
	Deleted  nullable.Time          `json:"deleted,omitzero"`
}

func main() {
	var p Product
	j, _ := json.Marshal(p)
	fmt.Println(string(j))

	// Zero is a value for Type[int], not null
	p.Discount.Set(0)
	p.Comment = nullable.TrimmedStringFrom("  needs review  ")
	j, _ = json.Marshal(p)
	fmt.Println(string(j))

	var in Product
	err := json.Unmarshal([]byte(`{"discount":null,"comment":"   "}`), &in)
	fmt.Println(err, in.Discount.IsNull(), in.Comment.IsNull(), in.Deleted.IsNull())

	v, err := p.Discount.Value()
	fmt.Printf("%#v %v\n", v, err)
	v, err = in.Discount.Value()
	fmt.Printf("%#v %v\n", v, err)
	fmt.Println(in.Discount.GetOr(-1), p.Comment.StringOr("none"), in.Comment.StringOr("none"))

	// Output:
	// {"discount":null,"comment":null}
	// {"discount":0,"comment":"needs review"}
	// <nil> true true true
	// 0 <nil>
	// <nil> <nil>
	// -1 needs review none
}
```

SQL and JSON of null values, and NULL array elements:

```go
package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/notnull"
	"github.com/domonda/go-types/nullable"
)

func main() {
	show := func(name string, v driver.Valuer) {
		val, err := v.Value()
		j, _ := json.Marshal(v)
		fmt.Printf("%-24s SQL %#v JSON %s err %v\n", name, val, j, err)
	}
	show("NonEmptyString(\"\")", nullable.NonEmptyString(""))
	show("NonEmptyString(\" \")", nullable.NonEmptyString(" "))
	show("TrimmedString(\" \")", nullable.TrimmedString(" "))
	show("IntArray(nil)", nullable.IntArray(nil))
	show("IntArray{}", nullable.IntArray{})
	show("notnull.IntArray(nil)", notnull.IntArray(nil))
	show("NullIntArray{1,null}", nullable.NullIntArray{nullable.TypeFrom[int64](1), {}})

	var ne nullable.NonEmptyString
	fmt.Println(ne.Scan(""))
	var ts nullable.TrimmedString
	fmt.Println(ts.Scan(""), ts.IsNull())

	var ints nullable.IntArray
	fmt.Println(ints.Scan("{1,NULL,3}"))
	var nullInts nullable.NullIntArray
	fmt.Println(nullInts.Scan("{1,NULL,3}"), nullInts, nullInts.Ints())

	// Output:
	// NonEmptyString("")       SQL <nil> JSON null err <nil>
	// NonEmptyString(" ")      SQL " " JSON " " err <nil>
	// TrimmedString(" ")       SQL <nil> JSON null err <nil>
	// IntArray(nil)            SQL <nil> JSON null err <nil>
	// IntArray{}               SQL "{}" JSON [] err <nil>
	// notnull.IntArray(nil)    SQL "{}" JSON [] err <nil>
	// NullIntArray{1,null}     SQL "{1,NULL}" JSON [1,null] err <nil>
	// can't scan empty string as nullable.NonEmptyString
	// <nil> true
	// can't parse "{1,NULL,3}" as nullable.IntArray because of: strconv.ParseInt: parsing "NULL": invalid syntax
	// <nil> NullIntArray{1,NULL,3} [1 0 3]
}
```

## Related

- [notnull](../notnull/): the never-null counterparts, where a nil slice is written as `'{}'` and `[]`.
- [uu](../uu/): `uu.NullableID` uses the zero UUID as null and implements `NullSetable[uu.ID]`.
- [Null conventions](../docs/explanation-null-conventions.md): when the zero value is null and when to use `Type[T]`.
- [How to store optional values](../docs/howto-store-optional-values.md)
- [How to validate structs](../docs/howto-validate-structs.md): `NullableValidator` and `types.Validator`.
