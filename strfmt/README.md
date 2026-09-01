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

1. `dest` is not settable → error, except for a non-nil pointer without a registered `Scanner`, whose pointed
   to value is scanned instead.
2. Source is a "nil string" (compared with surrounding whitespace trimmed): a kind that can be nil is set to nil,
   a type with a `SetNull()` method is set to null, a `string` without a scanning method of its own gets the source
   string (and a registered `Scanner` for a string kind still gets it, step 3). For every other type
   `config.StrictEmptyStringParsing` decides: `false` assigns the zero value, `true` passes it on to a
   `Scannable` or `encoding.TextUnmarshaler` destination (steps 5 and 6, or step 3 when a `Scanner` is
   registered for such a type as well, like `time.Time`) and errors for all the rest, including a type whose
   only scanning method is a registered `Scanner` of a non-string kind, like `time.Duration`.
   A destination set to nil, null or its zero value here is not passed to `ValidateFunc` — the absence of a
   value is not a value. A `string` assigned the source string is validated like any other scanned value.
3. Custom `Scanner` registered for `dest.Type()` in `config.TypeScanners`.
4. `dest` is a pointer → allocate it if it is nil, then scan the pointed to value in place, recursively for
   every level of indirection. An already allocated pointer keeps its identity and every field the scan
   doesn't assign.
5. `dest` implements `Scannable` → `ScanString(source, validate)`.
6. `dest` implements `encoding.TextUnmarshaler` → `UnmarshalText`.
7. Built-in scalars: string, bool (via `IsTrue`/`IsFalse`), int/uint (decimal, parsed with the bit size of the
   destination type), float (via `float.Parse`, checked for overflow).
8. If `config.ValidateFunc` is set, run it on a value scanned by step 7. Steps 3, 5 and 6 return before it —
   a `Scannable` validates itself and is only told whether `ValidateFunc` is set.

A failed scan never modifies `dest`. A scanning method can assign before it returns an error, and step 8 runs
after step 7 has assigned, so `Scan` restores the previous value on failure — including the pointed to value of
an already allocated pointer, and including a pointer it allocated in step 4, which is set back to nil.

### ScanConfig

```go
type ScanConfig struct {
    TrueStrings, FalseStrings, NilStrings []string
    TimeFormats                           []string
    AcceptedMoneyAmountDecimals           []int
    StrictEmptyStringParsing              bool

    TypeScanners map[reflect.Type]Scanner
    ValidateFunc func(any) error // nil disables validation
}

strfmt.NewScanConfig() // defaults below
```

Defaults: `TrueStrings = {true, True, TRUE, yes, Yes, YES, 1}`, mirror for false, `NilStrings = {"", nil, <nil>, null, NULL}`, `TimeFormats` covers RFC 3339 nano/sec, `time.DateTime`, `time.DateOnly`, browser `datetime-local`. Money decimals `{0, 2, 4}`. `StrictEmptyStringParsing = false` (a nil string scans as the zero value; set it to `true` to reject one for a non-optional type). `ValidateFunc = types.Validate`.

Built-in type scanners cover `time.Time` and `nullable.Time` (any registered layout) and `time.Duration` (a unit suffixed duration or plain nanoseconds). Add more via `SetTypeScanner`.

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
