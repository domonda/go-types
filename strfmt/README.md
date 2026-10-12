# strfmt

Reflection-driven conversion between strings and typed Go values. `Scan` parses a string into a `reflect.Value`, `Format` and `FormatValue` turn a value into a string, and the configs decide the locale details: true/false/nil strings, time layouts, number and money separators, and per-type scanners and formatters. Use it to read CSV or spreadsheet cells into struct fields and to write typed values into localized exports, without writing glue code per type.

```
import "github.com/domonda/go-types/strfmt"
```

## Interfaces

```go
type Scannable interface {
    ScanString(source string, validate bool) error
}

type Scanner interface {
    ScanString(dest reflect.Value, str string, config *ScanConfig) error
}
type ScannerFunc func(dest reflect.Value, str string, config *ScanConfig) error

type Formatter interface {
    FormatValue(val reflect.Value, config *FormatConfig) string
}
type FormatterFunc func(val reflect.Value, config *FormatConfig) string

type Parser interface {
    Parse(str string, langHints ...language.Code) (normalized string, err error)
}
```

| Interface   | Description                                        |
|-------------|----------------------------------------------------|
| `Scannable` | Implemented (pointer receiver) by types that parse themselves, like `date.Date`, `money.Amount`, `money.Currency`, `bank.IBAN`, `vat.ID`. `validate` asks for an error instead of assigning an invalid value as it is. `Scan` passes `config.ValidateFunc != nil` as `validate`. |
| `Scanner`   | Per-type scanning logic registered with `ScanConfig.SetTypeScanner`. `ScannerFunc.ScanString` calls the function. |
| `Formatter` | Per-type formatting logic registered in `FormatConfig.TypeFormatters`. `FormatterFunc.FormatValue` calls the function. |
| `Parser`    | The older string normalization interface, not used by `Scan` or `Format`. Implemented by `bank.IBANParser`, `bank.BICParser`, `date.Parser`, `date.Formatter`, `*date.Format`, `*float.FormatDef`, `*money.AmountParser`, `money.CurrencyParser`, `*money.CurrencyAmountParser`, and `vat.IDParser`. |

## Scan

```go
func Scan(dest reflect.Value, source string, config *ScanConfig) error
```

A nil `config` scans with `DefaultScanConfig`. `Scan` resolves the destination in this order:

1. An invalid `dest` is an error. A `dest` that is not settable, like `reflect.ValueOf(x)` or an unexported struct field, is an error, except a non-nil pointer without a registered `Scanner` for its pointer type, whose pointed to value is scanned instead.
2. A `Scanner` registered for `dest.Type()` in `config.TypeScanners` owns the type completely, including the nil string handling of step 3.
3. A nil string, meaning `config.IsNil` of the source with surrounding whitespace trimmed:
   - a pointer, interface, map, slice, chan, func, or unsafe pointer is set to nil,
   - a type with a `SetNull()` method (pointer receiver counts) is set to null,
   - a `string` kind without `ScanString` or `UnmarshalText` method gets the source string itself in step 7, so `"NULL"` scans as `"NULL"`,
   - every other type gets its zero value, unless `config.StrictEmptyStringParsing` is true: then a `Scannable` or `encoding.TextUnmarshaler` continues with step 5 or 6, and any other type is an error.

   Nil, null, or zero values set here are not passed to `ValidateFunc`.
4. A pointer is allocated if it is nil, then the pointed to value is scanned in place, recursively for every level of indirection. An already allocated pointer keeps its identity and every field the scan doesn't assign.
5. A `Scannable` gets `ScanString(source, config.ValidateFunc != nil)` with the untrimmed source.
6. An `encoding.TextUnmarshaler` gets `UnmarshalText([]byte(source))`.
7. Kind conversions:
   - `string`: the untrimmed source,
   - `bool`: the trimmed source must be in `TrueStrings` or `FalseStrings`, compared case-sensitive,
   - signed and unsigned integers: `strconv.ParseInt`/`ParseUint` of the trimmed source in base 10 with the bit size of the destination, so `"300"` into a `uint8` is an error,
   - `float32`, `float64`: `float.Parse`, with an error if the value overflows the destination,
   - a `nullable.Type[T]`, or another type whose pointer has the methods of `nullable.NullSetable[T]`, is set to the `T` scanned from the source with these steps, including the validation of `T`: `"1,5"` into a `money.NullableAmount` is 1.5,
   - every other kind is an error.
8. `config.ValidateFunc`, if not nil, validates a value scanned in step 7. Values from steps 2, 5, and 6 are not passed to it.

A failed scan never modifies `dest`. A scanning method or the validation can assign before it fails, so `Scan` restores the previous value, also the pointed to value of an allocated pointer and a pointer it allocated in step 4, which is set back to nil. Errors are wrapped with the destination value, the source, and the config.

## ScanConfig

```go
type ScanConfig struct {
    TrueStrings                 []string `json:"trueStrings"`
    FalseStrings                []string `json:"falseStrings"`
    NilStrings                  []string `json:"nilStrings"`
    TimeFormats                 []string `json:"timeFormats"`
    AcceptedMoneyAmountDecimals []int    `json:"acceptedMoneyAmountDecimals,omitempty"`
    StrictEmptyStringParsing    bool     `json:"strictEmptyStringParsing"`
    TypeScanners                map[reflect.Type]Scanner `json:"-"`
    ValidateFunc                func(any) error          `json:"-"`
}

var DefaultScanConfig = NewScanConfig()
```

| Field                         | `NewScanConfig()` default                          |
|-------------------------------|----------------------------------------------------|
| `TrueStrings`                 | `true`, `True`, `TRUE`, `yes`, `Yes`, `YES`, `ja`, `Ja`, `JA`, `1` |
| `FalseStrings`                | `false`, `False`, `FALSE`, `no`, `No`, `NO`, `nein`, `Nein`, `NEIN`, `0` |
| `NilStrings`                  | `""`, `nil`, `<nil>`, `null`, `NULL`               |
| `TimeFormats`                 | In this order: `time.RFC3339Nano`, `time.RFC3339`, `"2006-01-02 15:04:05.999999999 -0700 MST"` (`time.Time.String`), `time.DateTime`, `"2006-01-02 15:04"`, `"2006-01-02T15:04"` (browser `datetime-local`), `"02/01/2006 15:04:05 MST"` (English preset), `"02.01.2006 15:04:05 MST"` (German preset), `time.DateOnly` |
| `AcceptedMoneyAmountDecimals` | `0, 2, 4`. Not used by `Scan`, it's there to pass on to `money.ParseAmount` or `money.NewAmountParser`. |
| `StrictEmptyStringParsing`    | `false`: a nil string scans as the zero value of a type that can't be nil or null. |
| `TypeScanners`                | Scanners for `time.Time`, `nullable.Time`, and `time.Duration`, see below. |
| `ValidateFunc`                | `types.Validate`. Set it to nil to disable validation. |

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NewScanConfig() *ScanConfig`                      | A new config with the defaults above.              |
| `DefaultScanConfig`                                | The shared config that a nil config means. Changing it changes every caller that passes nil. |
| `(*ScanConfig).IsTrue(str) bool`                   | `str` is in `TrueStrings`, exact match.            |
| `(*ScanConfig).IsFalse(str) bool`                  | `str` is in `FalseStrings`, exact match.           |
| `(*ScanConfig).IsNil(str) bool`                    | `str` is in `NilStrings`, exact match.             |
| `(*ScanConfig).ParseTime(str) (time.Time, bool)`   | The result of the first of `TimeFormats` that `time.Parse` accepts, or the zero time and false. |
| `(*ScanConfig).SetTypeScanner(t reflect.Type, s Scanner)` | Registers or replaces the `Scanner` for `t`, allocating `TypeScanners` if it is nil. |

The scanners registered by `NewScanConfig`:

| Type            | Scanning                                           |
|-----------------|----------------------------------------------------|
| `time.Time`     | `ParseTime` of the trimmed source. A nil string is the zero time, or an error with `StrictEmptyStringParsing`. |
| `nullable.Time` | Like `time.Time`, but a nil string is always null, also with `StrictEmptyStringParsing`. |
| `time.Duration` | `time.ParseDuration` like `"1h30m"`, or an integer as nanoseconds, so `"1000"` is 1µs. A nil string like for `time.Time`. |

## Format and FormatValue

```go
func Format(value any, config *FormatConfig) string
func FormatValue(val reflect.Value, config *FormatConfig) string
```

`Format` calls `FormatValue` with `reflect.ValueOf(value)`, or with `value` itself if it is a `reflect.Value`. A nil `config` formats with a new `NewFormatConfig()`. `FormatValue` resolves the value in this order:

1. An invalid value, like the one of a nil `any`, is `config.Nil`.
2. The addressable value of an unexported struct field is read again through its address, so the steps below can call its methods. A non-addressable one, like a field of a struct not passed as pointer, skips steps 3 and 5 and ends with `fmt.Sprint` after step 6.
3. The `Formatter` registered in `config.TypeFormatters` for the type after dereferencing all non-nil pointers.
4. A null value is `config.Nil`: nil pointers, interfaces, maps, slices, chans, and funcs, and values whose `IsNull()` or `IsZero()` method returns true, see `nullable.ReflectIsNull`. `0`, `""`, and `false` are not null.
5. `encoding.TextMarshaler` of the value, its address, or the dereferenced value. If `MarshalText` returns an error, the steps below are used.
6. The kind of the dereferenced value:
   - `bool`: `config.True` or `config.False`,
   - `string`: the string unchanged,
   - `float32`, `float64`: `float.Format` with `config.Float`, a `float32` with 32-bit precision,
   - signed and unsigned integers: base 10 without grouping.
7. `fmt.Stringer` of the value, its address, or the dereferenced value.
8. A `nullable.Type[T]`, or another type whose pointer has the methods of `nullable.NullSetable[T]`, as the value of its `Get` method formatted with these steps: `nullable.TypeFrom(1234)` is `1234`, `money.NullableRate` like `money.Rate`.
9. A `[]byte` as string, everything else with `fmt.Sprint`.

## FormatConfig

```go
type FormatConfig struct {
    Float          float.FormatDef
    MoneyAmount    MoneyFormat
    Time           string // time.Time layout
    Date           string // date.Date layout
    Nil            string // placeholder for null values
    True           string
    False          string
    TypeFormatters map[reflect.Type]Formatter
}
```

`FormatConfig` has no JSON tags. The presets:

| Field         | `NewFormatConfig()`        | `NewEnglishFormatConfig()`  | `NewGermanFormatConfig()`   |
|---------------|----------------------------|-----------------------------|-----------------------------|
| `Float`       | `EnglishFloatFormat(-1)`   | `EnglishFloatFormat(-1)`    | `GermanFloatFormat(-1)`     |
| `MoneyAmount` | `EnglishMoneyFormat(true)` | `EnglishMoneyFormat(true)`  | `GermanMoneyFormat(true)`   |
| `Time`        | `time.RFC3339`             | `"02/01/2006 15:04:05 MST"` | `"02.01.2006 15:04:05 MST"` |
| `Date`        | `date.Layout` (2006-01-02) | `"02/01/2006"` (DD/MM/YYYY) | `"02.01.2006"`              |
| `Nil`         | `""`                       | `""`                        | `""`                        |
| `True`        | `"true"`                   | `"yes"`                     | `"ja"`                      |
| `False`       | `"false"`                  | `"no"`                      | `"nein"`                    |

All three register the same `TypeFormatters`:

| Type                                               | Formatted as                                       |
|----------------------------------------------------|----------------------------------------------------|
| `date.Date`, `date.NullableDate`                   | `Format(config.Date)`, `config.Nil` for the zero date or null, like `0001-01-01`. |
| `time.Time`, `nullable.Time`                       | `Format(config.Time)`, `config.Nil` for the zero time or null. |
| `time.Duration`                                    | `String()`, like `1h30m0s`.                        |
| `money.Amount`                                     | `config.MoneyAmount.FormatAmount`, also for zero.  |
| `money.CentAmount`                                 | `config.MoneyAmount.FormatCentAmount`, also for zero. |
| `money.DecimalAmount`                              | `config.MoneyAmount.FormatDecimalAmount`, also for zero. |
| `money.CurrencyAmount`                             | `config.MoneyAmount.FormatCurrencyAmount`.         |
| `money.CurrencyDecimalAmount`                      | `config.MoneyAmount.FormatCurrencyDecimalAmount`.  |
| `money.NullableAmount`, `money.NullableCentAmount`, `money.NullableDecimalAmount` | Like the wrapped type, `config.Nil` for null.      |

## MoneyFormat

```go
type MoneyFormat struct {
    CurrencyFirst bool
    ThousandsSep  rune
    DecimalSep    rune
    Precision     int
}
```

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `EnglishMoneyFormat(currencyFirst) MoneyFormat`    | `{currencyFirst, ',', '.', 2}`: `1,234.50`.        |
| `GermanMoneyFormat(currencyFirst) MoneyFormat`     | `{currencyFirst, '.', ',', 2}`: `1.234,50`.        |
| `(*MoneyFormat).FormatAmount(money.Amount) string` | `amount.FormatSep(ThousandsSep, DecimalSep, Precision)`: rounded half away from zero and padded to `Precision`, `-1` for the shortest exact form. |
| `(*MoneyFormat).FormatCentAmount(money.CentAmount) string` | Like `FormatDecimalAmount` of the amount with scale 2. |
| `(*MoneyFormat).FormatDecimalAmount(money.DecimalAmount) string` | Rounded half away from zero to `Precision` if it has more decimals, padded with zeros if it has fewer. A negative `Precision` writes the decimals of the amount's scale. NaN and ±Inf as `NaN`, `Inf`, `-Inf`. |
| `(*MoneyFormat).FormatCurrencyAmount(money.CurrencyAmount) string` | The amount like `FormatAmount`, with the currency code and a space before it if `CurrencyFirst`, else after it. No code for an empty currency. |
| `(*MoneyFormat).FormatCurrencyDecimalAmount(money.CurrencyDecimalAmount) string` | Like `FormatCurrencyAmount` with the amount from `FormatDecimalAmount`. |
| `EnglishFloatFormat(precision) float.FormatDef`    | `{0, '.', precision, false}`: no grouping, trailing zeros trimmed. |
| `GermanFloatFormat(precision) float.FormatDef`     | `{0, ',', precision, false}`.                      |

## Format and Scan round trip

What the presets write, and whether `Scan` with the default `ScanConfig` reads it back:

| Value                                           | `NewFormatConfig`      | English                   | German                    | Scans back               |
|-------------------------------------------------|------------------------|---------------------------|---------------------------|--------------------------|
| `true`                                          | `true`                 | `yes`                     | `ja`                      | yes                      |
| `1234.5` (`float64`)                            | `1234.5`               | `1234.5`                  | `1234,5`                  | yes                      |
| `money.Amount(1234.5)`                          | `1,234.50`             | `1,234.50`                | `1.234,50`                | yes                      |
| `money.CentAmount(123450)`                      | `1,234.50`             | `1,234.50`                | `1.234,50`                | yes                      |
| `money.CurrencyAmountEUR(1234.5)`               | `EUR 1,234.50`         | `EUR 1,234.50`            | `EUR 1.234,50`            | yes                      |
| `money.CurrencyAmountJPY(1234.5)`               | `JPY 1,234.50`         | `JPY 1,234.50`            | `JPY 1.234,50`            | yes                      |
| `date.Date("2024-01-31")`                       | `2024-01-31`           | `31/01/2024`              | `31.01.2024`              | yes                      |
| `time.Date(2024, 1, 31, 10, 0, 0, 0, time.UTC)` | `2024-01-31T10:00:00Z` | `31/01/2024 10:00:00 UTC` | `31.01.2024 10:00:00 UTC` | yes, see the zone gotcha |
| `90 * time.Minute`                              | `1h30m0s`              | `1h30m0s`                 | `1h30m0s`                 | yes                      |
| zero `time.Time`, null nullable types           | `config.Nil`           | `config.Nil`              | `config.Nil`              | yes                      |

## Gotchas

- **`"1,000"` scans as 1.** Floats and the money types parse with `float.Parse`, where a single `.` or `,` is the decimal separator. Integers are plain base 10, so `"1.000"` and `"1,000"` into an `int` are errors.
- **Empty cells scan as zero.** `""`, `NULL`, and the other nil strings scan as `0` into numbers and `money.Amount`, which can't be told apart from a real zero. Use a pointer field like `*money.Amount`, a nullable type like `money.NullableAmount` or `date.NullableDate`, or set `StrictEmptyStringParsing`.
- **A nil string into a `string` field is kept.** `"NULL"` and `"nil"` scan as the strings `"NULL"` and `"nil"`. Use `nullable.TrimmedString` or a `*string` to get null.
- **Strings are not trimmed.** A `string` destination and `Scannable.ScanString` get the source with its whitespace, bools and integers are compared and parsed trimmed.
- **Bool strings are an exact list.** `y`, `on`, and `X` are errors with the defaults.
- **The English preset writes dates as DD/MM/YYYY**, the British order: `31/01/2024`, not the US `01/31/2024`.
- **Zone abbreviations lose the offset.** The preset time layouts end with `MST`. `time.Parse` uses the offset of an abbreviation only if the local time zone knows it, so `31.01.2024 10:00:00 CET` scanned on a server in UTC is 10:00 in a zone named CET with offset 0. Times without a zone scan as UTC. Use `NewFormatConfig` with RFC 3339 for lossless times.
- **Money presets use 2 decimals for every currency.** `CurrencyAmountJPY(1234.5)` formats as `JPY 1,234.50`, while `CurrencyAmount.String` writes `JPY 1235`.
- **Registered formatters run before the null check.** A zero `money.Amount` formats as `0.00`, not as `config.Nil`.
- **Kinds win over `String` methods.** A type with a `string` or number kind and no registered formatter is formatted by its kind: `money.Currency("eur")` formats as `eur` although its `String` returns `EUR`, and `money.Rate(0.19)` as `0.19`.
- **Formatters are looked up by the dereferenced type.** A formatter registered for `*T` only runs for a nil `*T`. Register it for `T`.
- **Don't change `DefaultScanConfig` in place.** It is shared by every caller that passes nil. Create your own with `NewScanConfig`.
- **Zero value configs.** `FormatConfig{}` has `Float.DecimalSep` 0, so formatting a float panics in `float.Format`, and it writes `""` for both bools. `ScanConfig{}` has no nil, true, or false strings. Start from the constructors.
- **`any`, slice, and map destinations without a scanning method can't be scanned** from a non-nil string, only set to nil by a nil string.
- **A nil `FormatConfig` allocates.** `Format(v, nil)` builds a new default config on every call. Create the config once when formatting many values.

## Examples

Scan a spreadsheet row into a struct and write it with the presets:

```go
package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/strfmt"
)

type Invoice struct {
	Number string
	Amount money.Amount
	Paid   bool
	Due    date.NullableDate
	Items  int
}

func main() {
	cells := []string{"INV-1", "1.234,50", "ja", "31.01.2024", ""}

	var inv Invoice
	row := reflect.ValueOf(&inv).Elem()
	for i, cell := range cells {
		err := strfmt.Scan(row.Field(i), cell, nil)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("%+v\n", inv)

	for _, config := range []*strfmt.FormatConfig{strfmt.NewEnglishFormatConfig(), strfmt.NewGermanFormatConfig()} {
		var out []string
		for i := range row.NumField() {
			out = append(out, strfmt.FormatValue(row.Field(i), config))
		}
		fmt.Println(strings.Join(out, ";"))
	}

	// Output:
	// {Number:INV-1 Amount:1234.50 Paid:true Due:2024-01-31 Items:0}
	// INV-1;1,234.50;yes;31/01/2024;0
	// INV-1;1.234,50;ja;31.01.2024;0
}
```

Reject empty cells and register a scanner and a formatter for your own type:

```go
package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/strfmt"
)

// Percent holds 19.5% as 0.195
type Percent float64

func main() {
	scanConfig := strfmt.NewScanConfig()
	scanConfig.StrictEmptyStringParsing = true
	scanConfig.SetTypeScanner(reflect.TypeFor[Percent](), strfmt.ScannerFunc(
		func(dest reflect.Value, str string, config *strfmt.ScanConfig) error {
			f, err := float.Parse(strings.TrimSuffix(strings.TrimSpace(str), "%"))
			if err != nil {
				return err
			}
			dest.SetFloat(f / 100)
			return nil
		},
	))

	count := 7
	err := strfmt.Scan(reflect.ValueOf(&count).Elem(), "", scanConfig)
	fmt.Println(count, err != nil) // unchanged after the error

	var optional *int
	err = strfmt.Scan(reflect.ValueOf(&optional).Elem(), "", scanConfig)
	fmt.Println(optional, err)

	var vat Percent
	err = strfmt.Scan(reflect.ValueOf(&vat).Elem(), "19,5 %", scanConfig)
	fmt.Println(float64(vat), err)

	formatConfig := strfmt.NewGermanFormatConfig()
	formatConfig.TypeFormatters[reflect.TypeFor[Percent]()] = strfmt.FormatterFunc(
		func(val reflect.Value, config *strfmt.FormatConfig) string {
			return float.Format(val.Float()*100, 0, config.Float.DecimalSep, 1, false) + " %"
		},
	)
	fmt.Println(strfmt.Format(vat, formatConfig))

	// Output:
	// 7 true
	// <nil> <nil>
	// 0.195 <nil>
	// 19,5 %
}
```

## Related

- [`float`](../float/README.md): `float.Parse` scans floats, `float.FormatDef` formats them.
- [`money`](../money/README.md): the money types and their `ScanString` methods.
- [`date`](../date/README.md): `date.Date` and `date.NullableDate` parse many date formats in their `ScanString`.
- [`nullable`](../nullable/README.md): `nullable.ReflectIsNull` decides what formats as `config.Nil`, `nullable.Time` and `nullable.Type`.
- [Scan and format table rows](../docs/howto-scan-and-format-table-rows.md): reading and writing CSV rows step by step.
- [Lenient parsing](../docs/explanation-lenient-parsing.md): why `"1,000"` is 1.
- [Store optional values](../docs/howto-store-optional-values.md): optional fields for empty cells.
