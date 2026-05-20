# nullable

SQL-friendly nullable wrapper types and generic nullability interfaces. Each type has a single zero-like value (empty string, zero time, the zero-valued generic) that round-trips as SQL NULL and JSON `null`.

```
import "github.com/domonda/go-types/nullable"
```

## Why "zero == null"?

Go has no native nullable-with-default sentinel for `string`/`time.Time`, and `*string`/`*time.Time` is awkward in struct fields. This package leans on a convention: the type's natural zero value (`""`, `time.Time{}`) maps to SQL/JSON null on both directions. Methods like `IsNull`, `Get`, `GetOr`, `Set`, `SetNull` make the convention explicit.

For "empty is meaningful, null is separate" semantics, use `Type[T]` (a generic struct with a `valid` flag).

## Interfaces

```go
type Nullable interface { IsNull() bool }

type NullSetable[T any] interface {
    Nullable
    SetNull()
    Set(T)
    Get() T              // panics if null
    GetOr(T) T
}

type Zeroable interface { IsZero() bool }

type NullableValidator interface {
    Nullable
    types.Validator
    ValidAndNotNull() bool
}
```

Helper: `ReflectIsNull(v reflect.Value) bool` walks pointers/maps/slices/etc. and checks `Nullable` or `Zeroable` if implemented. Safe on the zero `reflect.Value`.

## Type[T] — generic nullable wrapper

```go
type Type[T any] struct { /* unexported */ }
```

The zero `Type[T]{}` represents null; a `valid` flag distinguishes "absent" from "present with the zero value of `T`". Useful for `int`, `bool`, structs — anywhere `0` or `false` is a legal value.

```go
nullable.TypeFrom(42)              // valid, value 42
nullable.TypeFromPtr[int](nil)     // null
nullable.TypeFromPtr(&someInt)     // valid

t.IsNull(); t.IsNotNull()
t.Get(); t.GetOr(defaultVal); t.Ptr()
t.Set(value); t.SetNull()
```

Implements `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`, `JSONSchema`. JSON `null` ↔ null; otherwise dispatch to `T`'s JSON behavior. SQL uses an embedded `convertAssign` (copied from `database/sql/convert.go`) so most scalar conversions work without `T` needing to implement `Scanner`.

## String wrappers

| Type             | Null sentinel         | Notes                                              |
|------------------|-----------------------|----------------------------------------------------|
| `NonEmptyString` | `""`                  | Empty string ↔ null. Can't represent an explicit empty string. |
| `TrimmedString`  | `""` (after trimming) | All marshal/unmarshal trims whitespace; whitespace-only ↔ null. |

Both ship constructors (`...From(str)`, `...FromPtr(*string)`, `...FromError(err)`, `...f("%s", x)`), null introspection (`IsNull`, `IsNotNull`, `Get`, `GetOr`, `Ptr`, `StringOr`), and SQL/JSON/text-marshaling interfaces.

`TrimmedString` additionally has the full `strings`-style API as methods (`ToUpper`, `ToLower`, `Contains`, `HasPrefix`, `HasSuffix`, `TrimPrefix`, `TrimSuffix`, `ReplaceAll`, `Split`, `ToValidUTF8`), each re-trimming the result. Also has XML marshal/unmarshal.

`JoinNonEmptyStrings(sep, strs...)` and `TrimmedStringJoin(sep, strs...)` skip null/empty entries.

## Time

```go
type Time struct { time.Time }
```

The zero `time.Time{}` ↔ null. `IsNull` uses `time.Time.IsZero()`. Constructors: `TimeNow`, `TimeFrom`, `TimeFromPtr`, `TimeParse(layout, value)`, `TimeParseInLocation`. The parsers treat `""`, `"null"`, `"NULL"` as null without error.

Wrapped methods (`Add`, `AddDate`, `UTC`, `Format`, `AppendFormat`, `Equal`) propagate null. `String()` returns `"NULL"` when null; use `StringOr(s)` to customize.

Text marshal: RFC 3339, with `"NULL"` for the null state. JSON: standard `time.Time` encoding with `null` for the null state. `PrettyPrint` integrates with `github.com/domonda/go-pretty`.

## Array types

PostgreSQL array types that round-trip as SQL arrays and JSON arrays. The purpose of these types is to keep SQL `NULL` distinct from an empty array: only a `nil` slice is `NULL`. Scanning `NULL` yields a `nil` slice, while scanning the empty array `'{}'` yields a non-nil empty slice — so the NULL-vs-empty distinction survives a round-trip. (For never-null arrays where a `nil` slice should serialize as `'{}'`/`[]`, use the sibling [`notnull`](../notnull/) package.)

| Go value / SQL / JSON                            | maps to                                          |
|--------------------------------------------------|--------------------------------------------------|
| `nil` slice                                      | SQL `NULL`, JSON `null`                          |
| empty non-nil slice                              | SQL `'{}'`, JSON `[]`                            |
| SQL `NULL`, JSON `null`                          | `nil` slice                                      |
| SQL empty array `'{}'`, JSON `[]`                | empty non-nil slice                              |

The `driver.Valuer` output and `sql.Scanner` input both use the [PostgreSQL array text format](https://www.postgresql.org/docs/current/arrays.html) (`{a,b,c}`). This works with PostgreSQL and array-compatible databases such as CockroachDB and YugabyteDB; databases without a native array type — MySQL, MariaDB, SQLite, SQL Server, Oracle — are not supported.

| Type             | Element                             |
|------------------|-------------------------------------|
| `StringArray`    | `string` (aliases `pq.StringArray`) |
| `IntArray`       | `int64`                             |
| `FloatArray`     | `float64`                           |
| `BoolArray`      | `bool`                              |
| `NullIntArray`   | nullable elements                   |
| `NullFloatArray` | nullable elements                   |
| `NullBoolArray`  | nullable elements                   |

For non-nullable arrays where `nil` should serialize as `'{}'`/`[]`, see the sibling `notnull` package.

Lower-level helpers: `SplitArray(s)` parses an SQL/JSON array literal into its top-level elements; `SQLArrayLiteral(s)` joins back. `null`/`NULL` decodes to a nil slice.

## Related

- `notnull` — sibling package with the same array types but never-null semantics.
- Most domain packages in `go-types` expose a parallel `NullableX` type using these conventions.
