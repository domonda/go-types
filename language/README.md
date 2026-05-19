# language

ISO 639-1 two-letter language codes with validation, normalization, and SQL/JSON integration.

```
import "github.com/domonda/go-types/language"
```

## Type

`Code` is a `string`-based type backed by an internal `codeNames` map. Empty string equals `Null` and is treated as SQL NULL. The package exports constants for every ISO 639-1 code: `language.EN`, `language.DE`, `language.FR`, …

```go
const Null Code = ""
```

`Code` implements `fmt.Stringer`, `driver.Valuer`, `sql.Scanner`, `JSONSchema`, and `types.NormalizableValidator[Code]`.

## API

| Method                                           | Description                                      |
|--------------------------------------------------|--------------------------------------------------|
| `Valid()`                                        | True if the code is in the ISO 639-1 table.      |
| `ValidAndNormalized()`                           | True if valid AND already lowercase.             |
| `Normalized()`                                   | Lowercase + lookup; returns error if unknown.    |
| `LanguageName()`                                 | English name (e.g. `"German"`).                  |
| `String()`                                       | Normalized form if valid, raw otherwise.         |

## Constants

The package defines a constant for every ISO 639-1 code — `AA, AB, AF, …, EN, DE, FR, ZH, …`. See `constants.go` for the full list.

Helpers in `iso6393macro.go` and `iso6393names.go` cover ISO 639-3 macrolanguage and English-name lookups when you need to bridge two-letter and three-letter standards.

## Example

```go
code, err := language.Code("EN").Normalized()
if err != nil {
    log.Fatal(err)
}
fmt.Println(code)                  // "en"
fmt.Println(code.LanguageName())   // "English"

// Using a constant directly:
fmt.Println(language.DE.LanguageName()) // "German"
```

## SQL

```go
type Document struct {
    Lang language.Code // "" → NULL
}
```

`Scan` accepts `string`, `[]byte`, and `nil`. Values are stored raw; call `Normalized()` after reading if you need canonical form.

## TODOs

The current `Normalized()` does not yet handle:

- Three-letter ISO 639-2/3 codes.
- BCP-47 tags (`en-US`, `sr-Latn`).

See the inline `TODO` comments in `code.go`.

## Related

- `country` — sibling package for ISO 3166-1 country codes.
