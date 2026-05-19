# notnull

Sibling of [`nullable`](../nullable/) for "never-null" SQL/JSON values. A `nil` or empty slice serializes as `'{}'` (SQL) and `[]` (JSON) — not as NULL. Useful when a column is `NOT NULL` and you'd rather not babysit `nil` checks on the Go side.

```
import "github.com/domonda/go-types/notnull"
```

## Array types

| Type                                             | Element                                          |
|--------------------------------------------------|--------------------------------------------------|
| `StringArray`                                    | `string`                                         |
| `IntArray`                                       | `int64`                                          |
| `FloatArray`                                     | `float64`                                        |
| `NullBoolArray`                                  | nullable `bool` elements                         |

Each implements `sql.Scanner`, `driver.Valuer`, and `json.Marshaler`. Element-nullable arrays let individual entries be SQL NULL even though the array itself is never NULL.

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
