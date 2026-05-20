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

| Method                 | Description                                        |
|------------------------|----------------------------------------------------|
| `Valid()`              | True if the code is already in the strict ISO 639-1 table. |
| `ValidAndNormalized()` | True if valid AND already lowercase.               |
| `Normalized()`         | Trim, lowercase, strip BCP-47 region/script subtags, map ISO 639-2/3 to 639-1. |
| `LanguageName()`       | English name (e.g. `"German"`).                    |
| `String()`             | Normalized form if valid, raw otherwise.           |

`Normalized()` accepts a broader range of inputs than `Valid()`:

- ISO 639-1 two-letter codes, case-insensitive (`"en"`, `"EN"`, `" en "`).
- ISO 639-2/T and 639-3 three-letter codes (`"eng"`, `"deu"`, `"fra"`).
- ISO 639-2/B bibliographic variants (`"ger"` → `"de"`, `"fre"` → `"fr"`, `"chi"` → `"zh"`).
- BCP-47 tags by extracting the primary language subtag (`"en-US"` → `"en"`, `"zh-Hant-CN"` → `"zh"`, `"sr-Latn"` → `"sr"`).
- POSIX locale separators (`"en_US"` → `"en"`).

Languages without an ISO 639-1 assignment (most of ISO 639-3's ~7000 entries) are rejected with a wrapped error.

## Constants

The package defines a constant for every ISO 639-1 code — `AA, AB, AF, …, EN, DE, FR, ZH, …`. See `constants.go` for the full list.

`iso6393.go` holds the ISO 639-3 → 639-1 mapping that `Normalized()` consults (including 639-2/B bibliographic variants). The older `iso6393macro.go` and `iso6393names.go` files contain related ISO 639-3 macrolanguage and English-name data with their maps still commented out — see TODOS for the decision on whether to revive or delete them.

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

## Related

- `country` — sibling package for ISO 3166-1 country codes.
