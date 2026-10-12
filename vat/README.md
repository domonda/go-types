# vat

European VAT identification numbers. `ID` normalizes VAT IDs as they appear on invoices and in forms (`"atu 1022 3006"`, `"DE 167-015-661"`) to the canonical form, checks the country format and, for most countries, the check digit, finds VAT IDs in text, and reads and writes SQL. The checks run offline: they do not ask VIES whether a number is registered. Use `ID` for required VAT IDs and `NullableID` for optional ones.

    import "github.com/domonda/go-types/vat"

## ID

```go
type ID string

const (
    IDMinLength = 4      // minimum length after normalization
    IDMaxLength = 14 + 2 // maximum length after normalization
)

const MOSSSchemaVATCountryCode = "EU"      // VAT Mini One Stop Shop (MOSS) prefix
const NorthernIrelandVATCountryCode = "XI" // Northern Ireland prefix

const ErrInvalidID errs.Sentinel = "invalid VAT ID"
```

### Normalization

`Normalized()` runs these steps and stops at the first failure:

1. Upper-cases the input and removes all whitespace and Unicode punctuation: `" at u.1022-3006"` becomes `"ATU10223006"`.
2. Checks the length of the result against `IDMinLength` and `IDMaxLength` ("is too short", "is too long").
3. Checks that the first two characters are a valid `country.Code`, `EU` or `XI` ("has an invalid country code").
4. Checks that the package has a format for that prefix ("has an unsupported country code", for example `US` or `GR`).
5. Matches the country format ("has an invalid format").
6. Verifies the check digit if the country has a checksum ("has an invalid check-sum").

Every error wraps `ErrInvalidID`, so `errors.Is(err, vat.ErrInvalidID)` detects invalid input, and the message quotes the input and names the failed step. On failure `Normalized` returns the input unchanged.

### Supported prefixes

| Prefix | Format after the prefix                       | Check digit                               |
|--------|-----------------------------------------------|-------------------------------------------|
| `AT`   | `U` + 8 digits                                | yes                                       |
| `BE`   | `0` or `1` + 9 digits                         | yes                                       |
| `BG`   | 9 or 10 digits                                | yes                                       |
| `CH`   | `E` + 9 digits                                | no                                        |
| `CY`   | 8 digits + letter                             | yes                                       |
| `CZ`   | 8 to 10 digits                                | 8 digits only                             |
| `DE`   | 9 digits, not starting with 0                 | yes                                       |
| `DK`   | 8 digits                                      | yes                                       |
| `EE`   | 9 digits                                      | yes                                       |
| `EL`   | 9 digits (Greece)                             | yes                                       |
| `ES`   | NIF, NIE and CIF forms                        | yes                                       |
| `FI`   | 8 digits                                      | yes                                       |
| `FR`   | 2 key characters + 9 digits                   | if the key is numeric                     |
| `GB`   | 9 or 12 digits, or `GD`/`HA` + 3 digits       | not for `GD`/`HA`                         |
| `HR`   | 11 digits                                     | yes                                       |
| `HU`   | 8 digits                                      | yes                                       |
| `IE`   | 8 or 9 characters in the two Irish forms      | no                                        |
| `IS`   | 5 or 6 digits                                 | no                                        |
| `IT`   | 11 digits                                     | yes                                       |
| `LI`   | 5 digits                                      | no                                        |
| `LT`   | 9 or 12 digits                                | yes                                       |
| `LU`   | 8 digits                                      | yes                                       |
| `LV`   | 11 digits                                     | legal entities (first digit above 3) only |
| `MT`   | 8 digits                                      | yes                                       |
| `NL`   | 9 digits + `B` + 2 digits                     | no, post-2020 numbers are random          |
| `NO`   | 9 digits starting with 8 or 9, optional `MVA` | yes                                       |
| `PL`   | 10 digits                                     | yes                                       |
| `PT`   | 9 digits                                      | yes                                       |
| `RO`   | 2 to 10 digits                                | yes                                       |
| `SE`   | 10 digits + `01`                              | yes                                       |
| `SI`   | 8 digits                                      | yes                                       |
| `SK`   | 10 digits                                     | yes                                       |
| `SM`   | 5 digits                                      | no                                        |
| `XI`   | like `GB` (Northern Ireland)                  | like `GB`                                 |
| `EU`   | 9 digits (MOSS)                               | no                                        |

The formats and checksum functions live in `formats.go` (`idRegex` and `checkSumFuncs`). To support another country, add a regular expression for the normalized ID to `idRegex` and, if the country has a check digit, a function to `checkSumFuncs`.

### Functions and methods

| Identifier                     | Description                                        |
|--------------------------------|----------------------------------------------------|
| `NormalizeVATID(str)`          | `ID(str).Normalized()`.                            |
| `StringIsVATID(str)`           | `ID(str).Valid()`.                                 |
| `BytesAreVATID(b)`             | `ID(b).Valid()`.                                   |
| `Normalized()`                 | The canonical ID or a wrapped `ErrInvalidID`, see above. |
| `NormalizedOrNull()`           | The canonical ID as `NullableID`, or `Null` if invalid. |
| `Valid()`                      | True if `Normalized()` succeeds, whatever the spelling. |
| `Validate()`                   | The error of `Normalized()`, or nil.               |
| `ValidAndNormalized()`         | True only if the ID is valid and already canonical. |
| `ValidateIsNormalized()`       | A wrapped `ErrInvalidID` if invalid, a different error (not `ErrInvalidID`) if valid but not canonical, nil if canonical. |
| `CountryCode()`                | `country.Code` of the prefix: `EU` gives `country.BE`, `XI` gives `country.GB`, `EL` stays `country.EL`. `country.Invalid` if invalid. |
| `Number()`                     | Everything after the two-letter prefix, `""` if invalid. For Austria it includes the `U`: `"U10223006"`. |
| `IsMOSS()`                     | True for a valid ID with the `EU` prefix.          |
| `String()`                     | Canonical ID, or the raw string if invalid.        |
| `Nullable()`                   | Plain conversion to `NullableID`, no normalization. |
| `ScanString(source, validate)` | Assigns the canonical `source`. If that fails, returns the error when `validate` is true, else assigns `source` as is. |

Implements: `fmt.Stringer`, `database/sql.Scanner`, `database/sql/driver.Valuer`, `strfmt.Scannable`, `types.NormalizableValidator[ID]` and `JSONSchema()` (string with `minLength` 4 and `maxLength` 16). `ID` has no JSON methods.

## NullableID

```go
type NullableID string

const Null NullableID = ""
```

`NullableID` is `ID` where `Null` (the empty string) is a valid null value. A whitespace only value is null too, like for `country.NullableCode`.

| Method                             | Description                                        |
|------------------------------------|----------------------------------------------------|
| `Normalized()`                     | `Null` for null, otherwise like `ID`.              |
| `NormalizedOrNull()`               | The canonical ID, or `Null` if invalid.            |
| `NormalizedNotNull()`              | `ID(n).Normalized()`: an error for `Null`.         |
| `Valid()`, `Validate()`            | `Null` is valid, otherwise like `ID`.              |
| `ValidAndNotNull()`                | Valid and not `Null`.                              |
| `ValidAndNormalized()`             | True for `Null` and for a canonical ID.            |
| `ValidateIsNormalized()`           | Like `ID`, but nil for `Null`.                     |
| `ValidateIsNormalizedAndNotNull()` | `ID(n).ValidateIsNormalized()`: an error for `Null`. |
| `IsNull()`, `IsNotNull()`          | Null is `""` or whitespace only.                   |
| `Get()`                            | The `ID`; panics if null.                          |
| `GetOr(defaultID)`                 | The `ID`, or `defaultID` if null.                  |
| `Set(id)`, `SetNull()`             | Assign an `ID` or `Null`.                          |
| `String()`                         | Canonical ID, `""` for `Null`, or the raw string if invalid. |
| `StringOr(nullString)`             | The raw string (not normalized), or `nullString` if null. |
| `CountryCode()`                    | Like `ID`, as `country.NullableCode`; `country.Null` if null or invalid. |
| `IsMOSS()`, `Number()`             | Like `ID`; false and `""` for `Null`.              |
| `ScanString(source, validate)`     | `""`, `"NULL"`, `"null"` and `"nil"` set `Null`; otherwise like `ID`. |

Implements: `fmt.Stringer`, `database/sql.Scanner`, `database/sql/driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `strfmt.Scannable`, `nullable.NullSetable[ID]` and `JSONSchema()` (the `ID` string or null).

## Finding VAT IDs in text

```go
var IDFinder idFinder
```

`IDFinder.FindAllIndex(str []byte, n int) [][]int` returns the `[start, end]` byte offsets of valid VAT IDs in `str`, at most `n` (`n < 0` returns all, `n == 0` none). `IDFinder` implements `types.Finder`. It splits the text on whitespace and `:`, trims punctuation from each word, and tests runs of one to three words, so `"UID: ATU 10223006,"` yields `"ATU 10223006"`. The match is the raw text: pass it to `NormalizeVATID` to get the canonical ID.

## strfmt

`IDParser{}` implements `strfmt.Parser`: `Parse(str, langHints...)` returns the canonical ID as string or the error of `NormalizeVATID`. The language hints are ignored.

## Serialization

| Value                         | SQL `Value()`     | JSON (`json.Marshal`) | `String()`      |
|-------------------------------|-------------------|-----------------------|-----------------|
| `ID("")`                      | `""` (not `NULL`) | `""`                  | `""`            |
| `ID("atu 1022 3006")` (valid) | `"ATU10223006"`   | `"atu 1022 3006"`     | `"ATU10223006"` |
| `ID("ATU10223006")`           | `"ATU10223006"`   | `"ATU10223006"`       | `"ATU10223006"` |
| `ID("bad")` (invalid)         | `"bad"`, no error | `"bad"`               | `"bad"`         |
| `NullableID` `Null`, `"  "`   | `NULL`            | `null`                | `""`            |
| `NullableID("atu 1022 3006")` | `"ATU10223006"`   | `"atu 1022 3006"`     | `"ATU10223006"` |
| `NullableID("bad")` (invalid) | `"bad"`, no error | `"bad"`               | `"bad"`         |

| Input                          | `ID`                                    | `NullableID`         |
|--------------------------------|-----------------------------------------|----------------------|
| `Scan(nil)`                    | error wrapping `ErrInvalidID`           | `Null`               |
| `Scan(string)`, `Scan([]byte)` | stored as is, not normalized or checked | same                 |
| `Scan` of other types          | error wrapping `ErrInvalidID`           | error                |
| JSON string                    | stored as is, not normalized or checked | same                 |
| JSON `null`                    | field left unchanged                    | `Null`               |
| other JSON values              | error                                   | error                |

## Gotchas

- **`Valid` ignores the spelling.** `ID("atu 1022 3006").Valid()` is true. Use `ValidAndNormalized` or `ValidateIsNormalized` to require the canonical form, or store `Normalized()`.
- **Greece uses `EL`.** `"GR094019245"` fails with "unsupported country code", `"EL094019245"` is valid. `CountryCode()` returns `country.EL`, which `country.Code.IsEU` counts as an EU member.
- **Only listed prefixes.** A valid country without a format, like `US`, is rejected.
- **Format only for some countries.** CH, IE, IS, LI, NL, SM and `EU` have no checksum; CZ (9 and 10 digits), FR (letter keys), GB/XI (`GD`, `HA`) and LV (natural persons) are checked partly. A well-formed made-up number passes there.
- **The length limit applies after normalization.** `"AT - U - 1022 - 3006"` (20 characters) is valid. The JSON schema puts `maxLength` 16 on the raw string, so a schema validator rejects input that `Normalized` accepts.
- **`"No."` is not Norway.** Input starting with `"No."` or `"no."` is rejected, because it is far more often the English abbreviation for number.
- **`ID` is not nullable.** `ID("").Value()` writes an empty string, and `Scan(nil)` returns an error. Use `NullableID` for nullable columns.
- **Output is only partly normalized.** `Value` writes valid IDs canonical, but JSON encodes the raw string for both types. Normalize before encoding if consumers expect the canonical form.
- **`StringOr` is not normalized.** `NullableID("atu 10223006").StringOr("-")` is `"atu 10223006"`.
- **The finder reads at most three words.** `"DE 167 015 661"` spans four words and is not found; `"DE 167015661"` is.

## Examples

Validate and normalize input:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/domonda/go-types/vat"
)

func main() {
	for _, s := range []string{"ATU10223006", " at u 1022 3006", "DE 167-015-661", "ATU12345678", "AT123", "GR094019245"} {
		id, err := vat.NormalizeVATID(s)
		fmt.Printf("%q %v %v\n", string(id), errors.Is(err, vat.ErrInvalidID), err)
	}

	id := vat.ID("atu 1022 3006")
	fmt.Println(id.Valid(), id.ValidAndNormalized(), id.CountryCode(), id.Number())
	fmt.Println(vat.ID("EU372008134").CountryCode(), vat.ID("EU372008134").IsMOSS())

	// Output:
	// "ATU10223006" false <nil>
	// "ATU10223006" false <nil>
	// "DE167015661" false <nil>
	// "ATU12345678" true invalid VAT ID: "ATU12345678" has an invalid check-sum
	// "AT123" true invalid VAT ID: "AT123" has an invalid format
	// "GR094019245" true invalid VAT ID: "GR094019245" has an unsupported country code: "GR"
	// true false AT U10223006
	// BE true
}
```

Find VAT IDs in text and store them:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/vat"
)

type Invoice struct {
	SupplierVATID vat.ID         `json:"supplierVatId"`
	BuyerVATID    vat.NullableID `json:"buyerVatId"`
}

func main() {
	text := []byte("Lieferant: Muster GmbH, UID: ATU 10223006, USt-IdNr. DE 136725570")
	for _, idx := range vat.IDFinder.FindAllIndex(text, -1) {
		raw := string(text[idx[0]:idx[1]])
		id, _ := vat.NormalizeVATID(raw)
		fmt.Printf("%q -> %s\n", raw, id)
	}

	inv := Invoice{SupplierVATID: vat.ID("atu 1022 3006").NormalizedOrNull().Get()}
	b, _ := json.Marshal(inv)
	fmt.Println(string(b))

	v, _ := vat.ID("atu 1022 3006").Value()
	fmt.Println(v)
	v, _ = vat.Null.Value()
	fmt.Println(v)

	// Output:
	// "ATU 10223006" -> ATU10223006
	// "DE 136725570" -> DE136725570
	// {"supplierVatId":"ATU10223006","buyerVatId":null}
	// ATU10223006
	// <nil>
}
```

## Related

- [`country`](../country/README.md): the `country.Code` returned by `CountryCode`.
- [`strfmt`](../strfmt/README.md): uses `IDParser`.
- [Extract identifiers from text](../docs/howto-extract-identifiers-from-text.md): using `IDFinder`.
- [Validation and normalization](../docs/explanation-validation-and-normalization.md): `Valid` vs `ValidateIsNormalized`.
- [Null conventions](../docs/explanation-null-conventions.md): `ID` vs `NullableID`.
