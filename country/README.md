# country

ISO 3166-1 alpha-2 country codes with normalization, EU membership lookup, English names, and full SQL/JSON support.

```
import "github.com/domonda/go-types/country"
```

## Types

| Type           | Purpose                                           |
|----------------|---------------------------------------------------|
| `Code`         | Non-nullable country code (string underneath).    |
| `NullableCode` | Same shape; empty string is SQL NULL / JSON null. |

Both implement `fmt.Stringer`, `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `JSONSchema`, and the `ScanString(source, validate)` helper. `Code` implements `types.NormalizableValidator[Code]`.

## Constants

```go
const Invalid Code = ""        // sentinel for invalid Code
const Null NullableCode = ""   // SQL NULL
```

## Constructors & conversion

```go
country.Code("at")            // raw — call Normalized() to get "AT"
country.Code("DE").Nullable() // → NullableCode("DE")
```

There's no `From(str)` constructor — assign the literal and call `Normalized()` or `Validate()`.

## Inspection

| Method                 | Description                                        |
|------------------------|----------------------------------------------------|
| `Valid()`              | True if normalization succeeds.                    |
| `ValidAndNormalized()` | True if valid AND already in canonical form.       |
| `Validate()`           | Returns error if normalization fails.              |
| `Normalized()`         | Trim, uppercase, match ISO map or `AltCodes`; returns error. |
| `IsEU()`               | True if the country is currently in the European Union. |
| `EnglishName()`        | English country name (empty for invalid codes).    |

`NullableCode` adds `IsNull`, `IsNotNull`, `Get`, `GetOr`, `Set`, `SetNull`, `StringOr`, `NormalizedOrNull`, and `NormalizedWithAltCodes`.

## Normalization

`Normalized()` accepts:
1. The canonical 2-letter code in any case (`"at"`, `" AT "`, `"At"` → `"AT"`).
2. Any key from the `AltCodes` map (ITU codes, German names, common synonyms). See `data.go` for the table.

If neither matches, the original value is returned unchanged with a non-nil error.

## Example

```go
code, err := country.Code("at").Normalized()
if err != nil {
    log.Fatal(err)
}
fmt.Println(code)            // "AT"
fmt.Println(code.IsEU())     // true
fmt.Println(code.EnglishName()) // "Austria"
```

## SQL

```go
type Customer struct {
    Country country.Code         // "" → NULL on write
    Billing country.NullableCode // "" → NULL on write
}
```

`Scan` accepts `string`, `[]byte`, and `nil`; values are stored raw (call `Normalized()` after reading if needed).

## Related

- `country.AltCodes` (in `data.go`) — alternative-spelling lookup map.
- `language` — sibling package for ISO 639-1 language codes.
