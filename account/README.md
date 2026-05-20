# account

Account-number types with validation, parsing, and SQL/JSON/XML integration. An account number is alphanumeric and may contain `_ - / : . ; ,`; it must start with `[0-9A-Za-z]`.

```
import "github.com/domonda/go-types/account"
```

## Types

| Type             | Purpose                                            |
|------------------|----------------------------------------------------|
| `Number`         | Non-nullable account number (a `string` underneath). |
| `NullableNumber` | Same shape; empty string is treated as SQL NULL / JSON null. |

Both implement `fmt.Stringer`, `driver.Valuer`, `sql.Scanner`, `json.Marshaler`/`Unmarshaler`, `xml.Unmarshaler`, `encoding.TextUnmarshaler`, and `JSONSchema`.

## Constants & errors

```go
const NumberRegex = `^[0-9A-Za-z][0-9A-Za-z_\-\/:.;,]*$`
const NumberNull NullableNumber = ""

var (
    ErrInvalidNumber      errs.Sentinel = "invalid account number"
    ErrAlphanumericNumber errs.Sentinel = "account number is alphanumeric"
)
```

## Constructors

```go
account.NumberFrom(str)            // validate & trim whitespace
account.NumberFromUint(u)          // format uint64 as Number
account.NullableNumberFrom(str)    // empty → NumberNull
account.NullableNumberFromUint(u)  // 0 → NumberNull
```

## Inspection

| Method                    | Description                                        |
|---------------------------|----------------------------------------------------|
| `Valid()`                 | True if matches `NumberRegex` (or null, for `NullableNumber`). |
| `Validate()`              | Returns wrapped `ErrInvalidNumber` if invalid.     |
| `IsNumeric()`             | True if contains only digits (`0-9`).              |
| `ValidateNumeric()`       | Returns wrapped `ErrAlphanumericNumber` if not purely numeric. |
| `HasPrefix` / `HasSuffix` | String-prefix/suffix tests.                        |
| `Cut(sep)`                | Like `strings.Cut`, returning `Number` halves.     |
| `TrimLeadingZeros()`      | Drops leading `'0'`s.                              |

## Numeric conversion

`Uint() (uint64, error)`, `UintPtr() (*uint64, error)`, `Int() (int64, error)` — return `ErrAlphanumericNumber` if the value isn't purely digits. This is deliberate: it prevents accidental parsing of hex-shaped strings via `strconv`.

`NullableNumber.Uint()` returns `(0, nil)` for NULL; `UintPtr()` returns `(nil, nil)`.

## Nullable helpers

```go
nn.IsNull(); nn.IsNotNull()
nn.Get()                // panics if null — check IsNull first
nn.GetOr(defaultNumber) // safe fallback
nn.Set(num); nn.SetNull()
n.Nullable()            // Number → NullableNumber
```

## Example

```go
n, err := account.NumberFrom("AT-12345/67")
if err != nil {
    log.Fatal(err)
}
fmt.Println(n.IsNumeric()) // false

if u, err := account.Number("123456").Uint(); err == nil {
    fmt.Println(u) // 123456
}
```

## SQL

```go
type Row struct {
    Primary account.Number         // empty string → NULL
    Backup  account.NullableNumber // empty string → NULL
}
```

`Scan` accepts `string`, `[]byte`, `int64` (non-negative), `float64` (non-negative), and `nil`.
