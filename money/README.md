# money

Monetary amounts, ISO 4217 currency codes, currency+amount pairs, and conversion rates — with locale-aware parsing/formatting and SQL/JSON integration.

```
import "github.com/domonda/go-types/money"
```

## Amount

```go
type Amount float64
type NullableAmount = nullable.Type[Amount]
```

`Amount` is a `float64` with money-specific behavior. Helpers:

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseAmount(str, decimals...)`                    | Locale-aware parse via `float.ParseDetails`. Optional decimal allowlist. |
| `NewAmount(v)` / `AmountFromPtr`                   | Pointer round-trip helpers.                        |
| `a.Cents()`                                        | Rounded to integer cents (`int64`).                |
| `a.WithinOneCent(b)`                               | True if `abs(a - b)` ≤ 0.01.                       |
| `a.RoundToInt()` / `RoundToCents()` / `RoundToDecimals(n)` | Rounding helpers.                                  |
| `a.Format(...)`                                    | Wraps `float.Format` for locale output.            |
| `a.Valid()`                                        | Not NaN, not Inf.                                  |
| `a.Ptr()`                                          | Pointer to a copy of the value.                    |
| `a.ScanString(src, validate)`                      | Assign from string, validating only if asked.      |

`NullableAmount` is `nullable.Type[Amount]`. Constructors `NullableAmountFrom(v)` and `NullableAmountFromPtr(*Amount)`.

`AmountFinder` matches localized amount patterns (`-?\d+,\d{2}`, `-?\d{1,3}(?:\.\d{3})*(?:,\d{2})`, etc.); `StringIsAmount(str, acceptInt)` is the boolean shortcut.

## Currency

```go
type Currency string         // empty → SQL NULL; Currency("").Valid() == false
type NullableCurrency string // empty → SQL NULL; NullableCurrency("").Valid() == true
```

ISO 4217 alphabetic codes (`USD`, `EUR`, `CHF`, …). Each currency has a package-level constant — see `constants.go`. Both types implement `fmt.Stringer`, `driver.Valuer`, `sql.Scanner`, `JSONSchema`, and the `ScanString(src, validate)` helper.

| Method                                  | Description                                        |
|-----------------------------------------|----------------------------------------------------|
| `c.Normalized()`                        | Trim, uppercase, resolve symbols (`€`, `$`) or English aliases (`"Euro"`, `"Swiss Franc"`). |
| `c.Valid()` / `c.Validate()`            | Pass/error variants.                               |
| `c.ValidAndNormalized()`                | Already in canonical form.                         |
| `c.NullableCurrency()` / `n.Currency()` | Convert between the two flavors.                   |
| `c.GetOrDefault(def)`                   | Pointer-safe with fallback.                        |

Helpers: `NormalizeCurrency(str)`, `StringIsCurrency(str)`. `CurrencyParser{}` implements `strfmt.Parser`.

## CurrencyAmount

```go
type CurrencyAmount struct {
    Currency Currency
    Amount   Amount
}
```

Constructors: `NewCurrencyAmount`, `CurrencyAmountUSD`, `CurrencyAmountEUR`, `CurrencyAmountCHF`, `CurrencyAmountGBP`, `CurrencyAmountJPY`.

`ParseCurrencyAmount(str, decimals...)` accepts `"EUR 1.234,56"`, `"1,234.56 USD"`, etc. — it finds the first/last separator between letters and digits, normalizes each part, and validates decimal counts if an allowlist is given.

## Rate

```go
type Rate float64
type NullableRate = nullable.Type[Rate]
```

For conversion rates and percentages. `ParseRate(str, decimals...)` accepts a trailing `%` and divides by 100.

```go
r, _ := money.ParseRate("19%")  // → Rate(0.19)
r, _ := money.ParseRate("1,2345") // → Rate(1.2345)
```

## strfmt integration

`AmountParser`, `CurrencyParser`, and (via `bank.IBANParser`) the broader ecosystem implement `strfmt.Parser`. Construct an `AmountParser` with an optional decimal allowlist:

```go
p := money.NewAmountParser(0, 2) // accept integers or 2-decimal amounts
normalized, err := p.Parse("1.234,56")
```

## Constants

Every ISO 4217 alphabetic code (`AED`, `AFN`, …, `USD`, `EUR`, `XPF`, `ZWL`, …) is defined as an untyped string constant — see `constants.go`. Also: `CurrencyNull = ""`.

## Related

- `float` — locale-aware parsing/formatting that backs `Amount` and `Rate`.
- `nullable` — `NullableAmount` / `NullableRate` reuse `nullable.Type[T]`.
- `bank.CAMT53` — embeds `Currency` and `Amount` in statement entries.
