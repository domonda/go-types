# money

Monetary amounts (float64, integer cents and exact fixed-point), ISO 4217 currency codes, currency+amount pairs, and conversion rates — with locale-aware parsing/formatting and SQL/JSON integration.

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
| `a.CentAmount(mode)`                               | Exact whole cents with a `RoundingMode`.           |
| `a.WithinOneCent(b)`                               | True if `abs(a - b)` ≤ 0.01.                       |
| `a.RoundToInt()` / `RoundToCents()` / `RoundToDecimals(n)` | Rounding helpers.                                  |
| `a.FormatSep(...)`                                 | Wraps `float.Format` for locale output.            |
| `a.Valid()`                                        | Not NaN, not Inf.                                  |
| `a.Ptr()`                                          | Pointer to a copy of the value.                    |
| `a.ScanString(src, validate)`                      | Assign from string, validating only if asked.      |
| `a.GoString()`                                     | Exact Go source literal, e.g. `money.Amount(0.5)`. |

`NullableAmount` is `nullable.Type[Amount]`. Constructors `NullableAmountFrom(v)` and `NullableAmountFromPtr(*Amount)`.

`AmountFinder` matches localized amount patterns (`-?\d+,\d{2}`, `-?\d{1,3}(?:\.\d{3})*(?:,\d{2})`, etc.); `StringIsAmount(str, acceptInt)` is the boolean shortcut.

## DecimalAmount

```go
type DecimalAmount struct{ /* packed int64 */ }
type NullableDecimalAmount = nullable.Type[DecimalAmount]
```

Exact fixed-point money, an alternative to the `float64`-based `Amount`. The value is `Coefficient × 10^-Scale`, packed into a single `int64`: the low 5 bits hold the scale (0–18) and the high 59 bits the signed coefficient (~18 significant digits, ±2.88×10¹⁷). Finite values are exact — `0.10` and `99999999999999.99` keep every digit — and the per-value scale is preserved, so `1.50` (scale 2) and `1.5` (scale 1) are distinct representations of the same value.

The reserved scale `31` encodes the non-finite states **NaN**, **+Inf** and **-Inf**. Arithmetic never panics on data: overflow yields ±Inf, `x/0` yields ±Inf (or NaN for `0/0`), and NaN propagates like IEEE floating point. Only misuse — an out-of-range `scale`/`decimals` argument — panics.

Implements `fmt.Stringer`, `fmt.GoStringer`, `fmt.Formatter`, `driver.Valuer`, `sql.Scanner`, `json.Marshaler`/`Unmarshaler`, `encoding.TextMarshaler`/`Unmarshaler`, `encoding.BinaryMarshaler`/`Unmarshaler` and `JSONSchema`.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `MakeDecimalAmount(coeff, scale)`                  | From integer coefficient and scale (panics if out of range). |
| `NewDecimalAmount(coeff, scale)`                   | Same, as a pointer (`New*` returns a pointer, like `NewAmount`). |
| `ParseDecimalAmount(str, decimals...)`             | Exact locale-aware parse (no `float64` round-trip). Reads `NaN`/`Inf`/`Infinity`. |
| `DecimalAmountFrom(v)`                             | Generic conversion from integer types (exact, scale 0) and float types incl. `Amount`/`Rate` (shortest exact decimal of the float value). |
| `DecimalAmountNaN()` / `DecimalAmountInf(sign)`    | Non-finite constructors.                           |
| `a.Coefficient()` / `a.Scale()`                    | Raw parts (finite values only).                    |
| `a.Add(b)` / `a.Sub(b)` / `a.MulInt(n)` / `a.MulInt64(n)` | Exact arithmetic; integer overflow → ±Inf.  |
| `a.Mul(b, mode)` / `a.Div(b, mode)`                | 128-bit exact multiply/divide at full result precision; `mode` only applies when the exact result exceeds the representable precision. |
| `a.RoundToDecimals(n, mode)` / `RoundToCents(mode)` / `RoundToInt(mode)` | Rounding with an explicit `RoundingMode`. |
| `a.MultipliedByRate(r)` / `DividedByRate(r)` / `Percentage(p)` | Apply a `float64` `Rate`; result is the exact shortest decimal of the `float64` result. |
| `a.Cmp(b)` / `a.Equal(b)` / `a.Sign()` / `a.IsZero()` | Value comparison (total order `-Inf < finite < +Inf < NaN`). |
| `a.IsFinite()` / `a.Valid()` / `a.IsNaN()` / `a.IsInf(sign)` | Non-finite tests (mirror `Amount`).      |
| `a.Abs()` / `a.Neg()`                              | Sign helpers (propagate non-finite).               |
| `a.Float()` / `a.Amount()`                         | Back to `float64` / `Amount` (may lose precision). |
| `a.String()` / `a.FormatSep(thousands, decimal)`   | Exact rendering; `fmt` verbs via `fmt.Formatter` (`%v %s %q %f %d`, flags). |

`RoundingMode`: `RoundHalfAwayFromZero` (zero value / default), `RoundHalfToEven`, `RoundHalfUp`, `RoundHalfDown`, `RoundDown`, `RoundUp`, `RoundFloor`, `RoundCeil`.

### Calculate exactly, round once at the end

Arithmetic results carry the scale that holds the full precision of the result: `Add`/`Sub` use the larger operand scale, `Mul` the sum of the operand scales, and `Div` extends the scale as far as needed (up to the 18-place maximum for non-terminating quotients). So a chain of calculations loses no data along the way — round to the final precision (typically 2 decimal places for cents, or 4 in accounting) exactly once, at the end:

```go
price := money.MakeDecimalAmount(1999, 2) // 19.99 per unit
qty := money.DecimalAmountFrom(7)
vatFactor := money.MakeDecimalAmount(119, 2) // 1.19 → 19% VAT

net := price.Mul(qty, money.RoundHalfAwayFromZero)   // 139.93   (scale 2, exact)
gross := net.Mul(vatFactor, money.RoundHalfAwayFromZero) // 166.5167 (scale 4, exact)
perMonth := gross.Div(money.DecimalAmountFrom(12), money.RoundHalfAwayFromZero)
// 13.8763916666666667 — non-terminating, so it is rounded at the
// maximum representable precision (this is the only step that rounds)

cents := perMonth.RoundToCents(money.RoundHalfToEven)        // 13.88
account := perMonth.RoundToDecimals(4, money.RoundHalfToEven) // 13.8764
```

Had every step been rounded to cents instead, the monthly amount would come out as `166.52 / 12 → 13.88` here, but chains of such intermediate roundings drift by whole cents; keeping full precision until the final rounding avoids that. The float-based `MultipliedByRate`/`DividedByRate`/`Percentage` follow the same pattern: they return the exact shortest decimal of the `float64` result (e.g. `0.10 × Rate(3)` → `0.30000000000000004`) for you to round once at the end.

`NullableDecimalAmount` is `nullable.Type[DecimalAmount]`. Wrap a value with `a.Nullable()` or a pointer with `NullableDecimalAmountFromPtr(*DecimalAmount)` (nil → null).

## CentAmount

```go
type CentAmount int64
type NullableCentAmount = nullable.Type[CentAmount]
```

Whole cents (hundredths of a currency unit) as `int64` — the exact integer counterpart of the `float64`-based `Amount` for the common two decimal places case. Every value from `MinCentAmount` to `MaxCentAmount` is a valid amount; there are no NaN or infinite states. `MinCentAmount` is one above `math.MinInt64` so negating a valid amount can never overflow, the same trick the `DecimalAmount` coefficient range uses.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseCentAmount(str, mode, decimals...)`          | Exact locale-aware parse via `ParseDecimalAmount`, applying `mode` beyond 2 decimals. `NaN`/`Inf` are errors. |
| `NewCentAmount(v)` / `CentAmountFromPtr`           | Pointer round-trip helpers.                        |
| `c.Cents()`                                        | The raw cent count (`int64`).                      |
| `c.Amount()` / `a.CentAmount(mode)`                | Convert to/from the `float64` `Amount`.            |
| `c.DecimalAmount()` / `d.CentAmount(mode)`         | Convert to/from `DecimalAmount` (scale 2).         |
| `c.RoundToInt()`                                   | Round to whole currency units (multiple of 100).   |
| `c.WithinOneCent(b)`                               | True if `abs(c - b)` ≤ 1 cent.                     |
| `c.String()` / `c.FormatSep(thousands, decimal)`   | Exact rendering with always 2 decimal places.      |
| `c.Sign()` / `IsZero()` / `Signbit()` / `Abs()` / `Copysign(s)` | Sign helpers.                          |
| `c.MultipliedByRate(r)` / `DividedByRate(r)` / `Percentage(p)` | Apply a `float64` `Rate`, rounded back to cents. |
| `c.SplitEqually(n)` / `SplitProportionally(w)`     | Exact integer splits whose parts sum up to the initial amount. |

`ParseCentAmount` and `Amount.CentAmount(mode)` work on the decimal digits directly rather than through a `DecimalAmount`, so they cover the **full** `CentAmount` range — the `DecimalAmount` coefficient stops at `2^58-1`, which is 32× narrower. That is what makes `ParseCentAmount(c.String(), mode)` round-trip every valid `CentAmount`, and it also means `ParseCentAmount` accepts any number of decimal places (the extra ones only feed the rounding) rather than inheriting the 18-place `DecimalAmount` scale cap.

Conversions from a value that has no cent representation never produce a platform-dependent result. `DecimalAmount.CentAmount(mode)` is the only one that returns an error — for non-finite amounts and for amounts beyond the `DecimalAmount` coefficient range. `Amount.CentAmount(mode)` maps NaN to zero and clamps to `MinCentAmount`/`MaxCentAmount` instead, as do `MultipliedByRate`/`DividedByRate`/`Percentage`; `CentAmount.DecimalAmount()` saturates to ±Inf like every other `DecimalAmount` overflow.

`Amount.CentAmount(mode)` rounds the *shortest decimal representation* of the float rather than `float64(a) * 100`, so `Amount(0.145).CentAmount(RoundHalfAwayFromZero)` is 15 cents — what the literal reads — not the 14 cents that multiplying first would give.

`SplitEqually` and `SplitProportionally` are exact: the parts always sum up to the initial amount, and no part is ever more than one cent away from its fair share or carries a sign opposite to the split amount. `SplitEqually` hands the indivisible cents to the last parts; `SplitProportionally` hands them to the largest truncated fractions (the largest remainder method), uses only the weight magnitudes, and computes the shares in 128-bit integer arithmetic so `amount × weight` cannot overflow; if the weight magnitudes themselves sum past `uint64` it falls back to exact `big.Int` arithmetic rather than rescaling the weights, which would skew their ratios. This is a deliberate divergence from the `float64` `Amount` splits, which give the whole rounding difference to the last part.

JSON uses the plain cent count (`12345`), not the decimal `String` form (`"123.45"`) — the value *is* a cent count, and a fractional JSON number is rejected rather than silently rounded. Unlike `Amount` and `DecimalAmount`, a JSON *string* (`"123.45"`) is rejected too, so swapping a field's type to `CentAmount` is a breaking wire change. SQL works through the native `int64` handling of `database/sql`, like `Amount` does for `float64`.

`NullableCentAmount` is `nullable.Type[CentAmount]`. Constructors `NullableCentAmountFrom(v)` and `NullableCentAmountFromPtr(*CentAmount)` (nil → null).

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

## CurrencyDecimalAmount

```go
type CurrencyDecimalAmount struct {
    Currency Currency
    Amount   DecimalAmount
}
```

The exact counterpart of `CurrencyAmount`, pairing a `Currency` with a `DecimalAmount`. Same shape — constructors `NewCurrencyDecimalAmount`, `CurrencyDecimalAmountUSD/EUR/CHF/GBP/JPY`, and `ParseCurrencyDecimalAmount(str, decimals...)` — plus `FormatSep`, `String`, `GoString`, `ScanString`, `sql.Scanner`/`driver.Valuer`. Formatting uses the amount's own scale rather than forcing two decimals, and `ca.CurrencyAmount()` bridges back to the `float64` form.

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
