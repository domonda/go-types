# float

Float utilities: locale-aware parsing & formatting (commas, dots, apostrophes, spaces), decimal rounding, NaN/Inf checks, and a JSON-tolerant float type.

```
import "github.com/domonda/go-types/float"
```

## Generic helpers

All work on both `float32` and `float64` via `T ~float32 | ~float64`:

| Function                                         | Description                                      |
|--------------------------------------------------|--------------------------------------------------|
| `RoundToDecimals(f, n)`                          | Round `f` to `n` decimal places.                 |
| `DerefOr(ptr, defaultVal)`                       | Safe pointer dereference with default.           |
| `Valid(f)`                                       | True if `f` is not NaN and not Inf.              |
| `ValidAndHasSign(f, sign)`                       | Valid AND sign matches (`sign==0` matches any).  |

## Parsing

```go
f, err := float.Parse("1.234,56") // → 1234.56 (auto-detects separators)

f, thousands, decimal, decimals, err := float.ParseDetails("1 234.56")
// f=1234.56, thousands=' ', decimal='.', decimals=2
```

`Parse` and `ParseDetails` accept any combination of:

- Decimal separators: `.`, `,`
- Thousands separators: `.`, `,`, `'`, space
- Signs: leading or trailing `+` / `-`
- Scientific notation: `1.5e10`
- Specials: `NaN`, `Inf`, `+Inf`, `-Inf`

Rules: separators alternate (a value can have at most one kind as thousands and one as decimal); thousands groups must be exactly 3 digits apart; mixing space-thousands with another thousands separator is an error.

## Formatting

```go
float.Format(1234.56, '.', ',', 2, false)
// "1.234,56"   (German style: dot thousands, comma decimal, 2 decimals)

float.Format(1234.5, ',', '.', 4, true)
// "1,234.5000" (US style with padding)
```

```go
func Format[T ~float32 | ~float64](
    f T,
    thousandsSep rune, // 0, ',', '.', '\'', or ' '
    decimalSep   rune, // ',' or '.'
    precision    int,  // -1 = minimum needed; otherwise digit count
    padPrecision bool, // pad fraction with trailing zeros
) string
```

`Format` panics on invalid separators (unrecognized rune, `thousandsSep == decimalSep`, or `precision < -1`).

## FormatDef

A serializable format spec, useful when storing user/locale preferences:

```go
type FormatDef struct {
    ThousandsSep rune `json:"thousandsSep,string,omitempty"`
    DecimalSep   rune `json:"decimalSep,string"`
    Precision    int  `json:"precision"`
    PadPrecision bool `json:"padPrecision"`
}

ff := float.NewFormatDef() // { '.', -1, false } — dot decimal, no precision limit
ff.Format(3.14)            // "3.14"
ff.Parse("3.14")           // normalize a string through Parse + Format
```

## Tolerant

`Tolerant` is a `float64` newtype that accepts numbers, strings, and `null` from JSON. Empty string and `null` decode as `0`. Strings are parsed with `float.Parse`, so locale separators are supported.

```go
type Payment struct {
    Amount float.Tolerant `json:"amount"`
}

// Accepts: {"amount": 12.5}, {"amount": "12,5"}, {"amount": ""}, {"amount": null}
```

Convenience methods: `Valid`, `ValidAndGreaterZero`, `ValidAndSmallerZero`, `ValidAndHasSign`, `IsNaN`, `IsInf`, `AsFloatPtr`.

## Related

- `money.Amount` — uses these utilities for currency-aware parsing/formatting.
- `strfmt` — broader format detection (dates, amounts, integers).
