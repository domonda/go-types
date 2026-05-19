# vat

European VAT identification numbers with per-country format/checksum validation, normalization, finding VAT IDs in free text, and SQL/JSON integration.

```
import "github.com/domonda/go-types/vat"
```

## Types

| Type                                             | Purpose                                               |
|--------------------------------------------------|-------------------------------------------------------|
| `ID`                                             | Non-nullable VAT ID. Empty string is treated as NULL. |
| `NullableID`                                     | Same; empty string is a *valid* null value.           |

Both implement `fmt.Stringer`, `driver.Valuer`, `sql.Scanner`, `JSONSchema`, and the `ScanString(source, validate)` helper. `ID` implements `types.NormalizableValidator[ID]`. `NullableID` implements `nullable.NullSetable[ID]`.

## Constants & errors

```go
const (
    IDMinLength = 4
    IDMaxLength = 16 // 14 chars + 2 allowed spaces
)

const MOSSSchemaVATCountryCode = "EU" // EU MOSS scheme prefix

const ErrInvalidID errs.Sentinel = "invalid VAT ID"

var Null NullableID // ""
```

## Normalization & validation

`Normalized()` performs:

1. Uppercase + remove spaces/punctuation.
2. Verify length (`IDMinLength..IDMaxLength`).
3. Verify the leading two characters are a valid country code (or the MOSS prefix `EU`).
4. Match against the country-specific regex in `formats.go`.
5. Run a country-specific checksum if available.

```go
vat.ID(" at u 1234 5678").Normalized() // "ATU12345678", nil
vat.ID("AT123").Validate()             // wrapped ErrInvalidID — too short
```

Convenience: `NormalizeVATID(str)`, `StringIsVATID(str)`, `BytesAreVATID(b)`.

## Inspection

```go
id.CountryCode()   // country.Code — for MOSS "EU" this returns country.BE
id.Number()        // digits after the 2-letter prefix
id.IsMOSS()        // true if prefix is "EU"
id.Valid()
id.ValidAndNormalized()
id.Validate()
id.ValidateIsNormalized() // distinguishes "invalid" from "valid but not normalized"
id.Nullable()
```

`NullableID` adds `IsNull`, `IsNotNull`, `Get`, `GetOr`, `StringOr`, `Set`, `SetNull`, `NormalizedOrNull`, `NormalizedNotNull`, `ValidAndNotNull`, `ValidateIsNormalizedAndNotNull`.

## Finding VAT IDs in text

`IDFinder` implements `strutil.Finder` and scans byte slices for valid VAT IDs:

```go
text := []byte("Invoice from supplier ATU12345678 (VAT)")
for _, idx := range vat.IDFinder.FindAllIndex(text, -1) {
    fmt.Println(string(text[idx[0]:idx[1]])) // "ATU12345678"
}
```

The finder splits on whitespace and `:`, then probes 1–3-word windows. It tolerates internal punctuation since `Normalized()` strips it.

## strfmt integration

`IDParser` implements `strfmt.Parser`:

```go
parser := vat.IDParser{}
normalized, err := parser.Parse(" at u 1234 5678")
```

## Country coverage

Country regexes live in `formats.go`. Checksum functions live alongside the regex map (`checkSumFuncs`). Add a new country by editing both maps; see existing entries (AT/DE/FR…) as a template.

## Related

- `country` — used for the leading country-code prefix.
- `strfmt` — the `Parser` interface implemented by `IDParser`.
