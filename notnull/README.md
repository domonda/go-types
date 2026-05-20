# notnull

Sibling of [`nullable`](../nullable/) for "never-null" SQL/JSON values. These types model columns and values that are **never** NULL: a `nil` or empty slice serializes as `'{}'` (SQL) and `[]` (JSON) — never as NULL — and scanning a NULL or empty array always yields a non-nil empty slice. Useful when a column is `NOT NULL` and you'd rather not babysit `nil` checks on the Go side.

```
import "github.com/domonda/go-types/notnull"
```

## Why a nil slice maps to an empty array

`nil` is the zero value of a Go slice — what a slice variable or struct field holds before anything is assigned to it. Because this package exists to model values that are **NOT NULL**, that default `nil` must not become a SQL or JSON `null`. Instead the `nil` default is treated as an *empty array*:

- `Value` and `MarshalJSON` encode a `nil` (or empty) slice as the empty SQL array `'{}'` and the empty JSON array `[]`, never as `NULL` / `null`.
- `Scan` and `UnmarshalJSON` never produce a `nil` slice — SQL `NULL`, the empty array `'{}'`, JSON `null` and `[]` all decode to a non-nil empty slice.

So a notnull array is **never nil** and can always be used without a nil check. If you need to tell "no value" (NULL) apart from "empty array", use the sibling [`nullable`](../nullable/) package instead.

| Go value / SQL / JSON                            | maps to                                          |
|--------------------------------------------------|--------------------------------------------------|
| `nil` slice                                      | SQL `'{}'`, JSON `[]`                            |
| empty non-nil slice                              | SQL `'{}'`, JSON `[]`                            |
| SQL `NULL`, JSON `null`                          | empty non-nil slice                              |
| SQL empty array `'{}'`, JSON `[]`                | empty non-nil slice                              |

## Array types

| Type            | Element                  |
|-----------------|--------------------------|
| `StringArray`   | `string`                 |
| `IntArray`      | `int64`                  |
| `FloatArray`    | `float64`                |
| `NullBoolArray` | nullable `bool` elements |

Each implements `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, and `json.Unmarshaler`. Element-nullable arrays let individual entries be SQL NULL even though the array itself is never NULL — e.g. `NullBoolArray` encodes its invalid elements as JSON `null` (`[true,false,null]`) while still rendering an empty or nil array as `[]`.

The `driver.Valuer` output and `sql.Scanner` input both use the [PostgreSQL array text format](https://www.postgresql.org/docs/current/arrays.html) (`{a,b,c}`). This works with PostgreSQL and array-compatible databases such as CockroachDB and YugabyteDB; databases without a native array type — MySQL, MariaDB, SQLite, SQL Server, Oracle — are not supported.

Convenience: `StringArray.Contains(value)`.

For a slice that *should* be NULL when nil, use `nullable.StringArray` etc.

## TrimmedString

```go
type TrimmedString string
```

A `string` newtype where every marshal/unmarshal path trims leading/trailing whitespace. Empty strings are allowed (this is the not-null cousin of `nullable.TrimmedString`).

```go
notnull.TrimmedStringf("  %s  ", name) // → trimmed
notnull.TrimmedStringFrom("  hi  ")    // "hi"
notnull.TrimmedStringJoin(", ", a, b)  // each trimmed, then joined
```

API includes the `strings`-style methods: `ToUpper`, `ToLower`, `Contains`, `ContainsAny`, `ContainsRune`, `HasPrefix`, `HasSuffix`, `TrimPrefix`, `TrimSuffix`, `ReplaceAll`, `Split`, `ToValidUTF8`, `IsEmpty`. Also implements text, JSON, and XML marshaling.

## Helpers

```go
notnull.SplitArray("{a,b,c}")     // []string{"a","b","c"}
notnull.SplitArray("{}")          // []string{} (non-nil)
notnull.SplitArray("NULL")        // []string{} (non-nil)
notnull.SQLArrayLiteral(nil)      // "{}"
notnull.SQLArrayLiteral([]string{"a","b"}) // "{a,b}"
```

`SplitArray` parses an SQL or JSON array literal into top-level elements (quoted strings are returned still quoted). Unlike the nullable version, `nil` collapses to an empty non-nil slice.

`SQLArrayLiteral` always emits `'{}'` for empty input.

## Related

- `nullable` — same shapes with null semantics (nil/empty → NULL).
