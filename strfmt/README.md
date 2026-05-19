# strfmt

Reflection-driven string scanning and formatting with locale presets. One pair of entry points (`Scan` / `Format`) plus pluggable per-type scanners/formatters lets you bridge between user-facing strings (CSV rows, form fields, CLI args) and typed Go values without writing per-type glue.

```
import "github.com/domonda/go-types/strfmt"
```

## Interfaces

```go
// Implemented by package types that know how to parse themselves.
// Most domain types in go-types (date.Date, money.Amount, vat.ID, ...) satisfy this.
type Scannable interface {
    ScanString(source string, validate bool) error
}

type Scanner interface {
    ScanString(dest reflect.Value, str string, config *ScanConfig) error
}
type ScannerFunc func(reflect.Value, string, *ScanConfig) error

type Formatter interface {
    FormatValue(val reflect.Value, config *FormatConfig) string
}
type FormatterFunc func(reflect.Value, *FormatConfig) string
```

## Scanning

```go
err := strfmt.Scan(dest, source, strfmt.DefaultScanConfig)
```

Resolution order:

1. Custom `Scanner` registered for `dest.Type()` in `config.TypeScanners`.
2. `dest` is a nil pointer + source is a "nil string" → leave nil. Otherwise allocate and dereference.
3. `dest` implements `Scannable` → `ScanString(source, validate)`.
4. `dest` implements `encoding.TextUnmarshaler` → `UnmarshalText`.
5. Built-in scalars: string, bool (via `IsTrue`/`IsFalse`), int/uint (decimal), float (via `float.Parse`).
6. If `config.ValidateFunc` is set, run it on the final value.

### ScanConfig

```go
type ScanConfig struct {
    TrueStrings, FalseStrings, NilStrings []string
    TimeFormats                           []string
    AcceptedMoneyAmountDecimals           []int

    TypeScanners map[reflect.Type]Scanner
    ValidateFunc func(any) error // nil disables validation
}

strfmt.NewScanConfig() // defaults below
```

Defaults: `TrueStrings = {true, True, TRUE, yes, Yes, YES, 1}`, mirror for false, `NilStrings = {"", nil, <nil>, null, NULL}`, `TimeFormats` covers RFC 3339 nano/sec, `time.DateTime`, `time.DateOnly`, browser `datetime-local`. Money decimals `{0, 2, 4}`. `ValidateFunc = types.Validate`.

Built-in type scanners cover `time.Time` (any registered layout) and `time.Duration`. Add more via `SetTypeScanner`.

## Formatting

```go
str := strfmt.Format(value, strfmt.NewFormatConfig())
// or, with a reflect.Value already in hand:
str := strfmt.FormatValue(val, config)
```

Resolution order:

1. Invalid value → `config.Nil`.
2. Custom `Formatter` registered for the dereferenced type.
3. Null-or-zero (via `nullable.ReflectIsNull`) → `config.Nil`.
4. `encoding.TextMarshaler` → `MarshalText`.
5. Bool/string/int/uint/float built-ins (float goes through `float.Format`).
6. `fmt.Stringer` on value, addressed value, or dereferenced value.
7. `[]byte` → `string([]byte)`.
8. Final fallback: `fmt.Sprint`.

### FormatConfig

```go
type FormatConfig struct {
    Float          float.FormatDef
    MoneyAmount    MoneyFormat
    Percent        float.FormatDef
    Time           string // time layout
    Date           string // date layout
    Nil, True, False string
    TypeFormatters map[reflect.Type]Formatter
}
```

`NewFormatConfig()` registers formatters for `date.Date`, `date.NullableDate`, `time.Time`, `nullable.Time`, `time.Duration`, `money.Amount`, `money.CurrencyAmount`.

### Locale presets

```go
strfmt.NewEnglishFormatConfig() // "02/01/2006", yes/no, dot decimal, comma thousands
strfmt.NewGermanFormatConfig()  // "02.01.2006", ja/nein, comma decimal, dot thousands
```

Float/money helpers: `EnglishFloatFormat(precision)`, `GermanFloatFormat(precision)`, `EnglishMoneyFormat(currencyFirst)`, `GermanMoneyFormat(currencyFirst)`.

## MoneyFormat

```go
type MoneyFormat struct {
    CurrencyFirst bool
    ThousandsSep  rune
    DecimalSep    rune
    Precision     int
}

mf.FormatAmount(money.Amount(1234.5))
mf.FormatCurrencyAmount(money.CurrencyAmount{Currency: money.EUR, Amount: 1234.5})
```

## Parser interface

Domain packages (`date.Parser{}`, `money.AmountParser{...}`, `money.CurrencyParser{}`, `vat.IDParser{}`, `bank.IBANParser{}`) implement a parser shape used in older callsites:

```go
type Parser interface {
    Parse(str string, langHints ...language.Code) (normalized string, err error)
}
```

This precedes the reflection-driven `Scan`/`Format` API; both styles coexist.

## Related

- `float`, `money`, `date`, `nullable` — types that drop in via the default config.
- `types.Validate` — used as the default `ValidateFunc`.
