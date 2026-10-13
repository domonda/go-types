# money

Money amounts, ISO 4217 currencies, currency and amount pairs, and rates, with locale-aware parsing (`1.234,56`, `1,234.56`, `1'234.56`), rounding with explicit modes, and SQL, JSON, and `strfmt` support. Use it for invoice, bank, and accounting data that arrives as text in unknown locales and has to end up in typed fields and database columns.

```
import "github.com/domonda/go-types/money"
```

## Choosing an amount type

`Amount` is a `float64`: tolerant to parse, with NaN and ±Inf, but inexact in binary. `DecimalAmount` is an exact fixed-point decimal in one `int64` that keeps its scale, with NaN and ±Inf states and explicit rounding modes. `CentAmount` is an exact `int64` count of cents without non-finite states. Prefer `CentAmount` for two-decimal ledger values in integer columns, `DecimalAmount` for more decimals, multi-step calculations, and `numeric` columns, and `Amount` for interoperability with existing code. See [Money types and rounding](../docs/explanation-money-types-and-rounding.md) for the reasoning.

All three round half away from zero by default, and conversions from `Amount` round the shortest decimal representation of the float, not its binary value:

| Call                                               | Result   |
|----------------------------------------------------|----------|
| `Amount(0.145).RoundToCents()`                     | 0.15     |
| `Amount(0.145).Cents()`                            | 15       |
| `Amount(0.145).CentAmount(RoundHalfAwayFromZero)`  | 15       |
| `Amount(0.145).DecimalAmount(2, RoundHalfAwayFromZero)` | 0.15     |
| `Amount(0.145).String()`                           | `"0.15"` |
| `Amount(0.145).FormatSep(0, '.', 2)`               | `"0.15"` |
| `DecimalAmountFrom(Amount(0.145))`                 | 0.145    |

## Amount

```go
type Amount float64
type NullableAmount = nullable.Type[Amount]
```

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseAmount(str, acceptedDecimals ...int) (Amount, error)` | Parses with `float.ParseDetails`. With `acceptedDecimals` the number of decimals written in `str` must be one of them and NaN and ±Inf are errors, without them NaN and ±Inf are returned without error. |
| `AmountFromPtr(ptr *Amount, defaultVal Amount) Amount` | `*ptr`, or `defaultVal` for nil.                   |
| `NullableAmountFrom(value float64) NullableAmount` | A non-null `NullableAmount`. Takes a `float64`, not an `Amount`. |
| `NullableAmountFromPtr(ptr *Amount) NullableAmount` | Null for a nil pointer.                            |
| `Ptr() *Amount`                                    | Pointer to a copy.                                 |
| `(*Amount).AmountOr(defaultVal) Amount`, `FloatOr(defaultVal float64) float64`, `StringOr(defaultVal string) string` | The pointed to value, its `float64`, or its `String()`, or the default for a nil pointer. |
| `Cents() int64`                                    | Rounded half away from zero to whole cents. Undefined for NaN and ±Inf. |
| `CentAmount(rounding RoundingMode) CentAmount`     | Rounded to whole cents with `rounding`. NaN is 0, ±Inf and values beyond the range clamp to `MinCentAmount`/`MaxCentAmount`. |
| `DecimalAmount(scale int, rounding RoundingMode) DecimalAmount` | Rounded to `scale` decimals with `rounding`. NaN and ±Inf map to the non-finite states, out of range values to ±Inf. Panics for a scale outside 0 to 18. |
| `BigFloat() *big.Float`                            | A new `big.Float` of the `float64`.                |
| `RoundToInt() Amount`                              | `math.Round`.                                      |
| `RoundToCents() Amount`, `RoundToDecimals(decimals int) Amount` | Half away from zero on the shortest decimal, see `float.RoundToDecimals`. Negative decimals round to tens, hundreds. NaN and ±Inf unchanged. |
| `String() string`                                  | Rounded to 2 decimals with a dot, no grouping: `"1234.50"`. |
| `FormatSep(thousandsSep, decimalSep rune, precision int) string` | `float.Format` with padding to `precision`: `FormatSep('.', ',', 2)` is `"1.234,50"`, `-1` writes the shortest exact form. Panics for invalid separators. |
| `GoString() string`                                | Exact Go literal like `money.Amount(0.1000000000000000055511151231257827021181583404541015625)`. |
| `WithinOneCent(b Amount) bool`                     | `Cents()` of both differ by at most 1, so the unrounded difference can be just under 2 cents: `Amount(10.004).WithinOneCent(10.0149)` is true. |
| `(*Amount).Equal(b *Amount) bool`                  | Both nil, the same pointer, or equal values.       |
| `Signbit() bool`, `Copysign(sign Amount) Amount`   | Like `math.Signbit`, `math.Copysign`.              |
| `Abs() Amount`, `AbsFloat() float64`               | Absolute value.                                    |
| `(*Amount).Invert()`, `Inverted() Amount`          | Negate in place, or return the negation.           |
| `WithPosSign(positive bool) Amount`, `WithNegSign(negative bool) Amount` | Force the sign.                                    |
| `MultipliedByRate(rate Rate) Amount`, `DividedByRate(rate Rate) Amount` | `a * rate`, `a / rate`, unrounded.                 |
| `Percentage(percent float64) Amount`               | `a * percent / 100`, unrounded.                    |
| `SplitEqually(count int) []Amount`                 | `count` parts rounded to cents that sum up to the amount rounded to cents. The indivisible cents go to the last parts: `Amount(1).SplitEqually(3)` is `[0.33 0.33 0.34]`. Nil for `count < 1`, the unrounded amount for 1. |
| `SplitProportionally(weights []Amount) []Amount`   | Parts rounded to cents by the largest remainder method, with the weight magnitudes, summing up to the rounded amount. All zero weights put everything in the last part. NaN, ±Inf, or amounts beyond the `CentAmount` range are split without rounding. |
| `Valid() bool`                                     | Neither NaN nor ±Inf.                              |
| `ValidAndPositive()`, `ValidAndGreaterZero()`, `ValidAndSmallerZero()` | Valid and `>= 0`, `> 0`, `< 0`.                    |
| `ValidAndHasSign(sign int) bool`                   | Valid and the sign of `sign`, any sign for 0. Zero counts as positive. |
| `IsNaN() bool`, `IsInf() bool`                     | NaN, or +Inf or -Inf. `IsInf` takes no sign, unlike `DecimalAmount.IsInf`. |
| `(*Amount).ScanString(source string, validate bool) error` | `float.Parse`, with `validate` NaN and ±Inf are errors. |
| `(*Amount).UnmarshalJSON(j []byte) error`          | See [Serialization](#serialization).               |

Implements: `fmt.Stringer`, `fmt.GoStringer`, `json.Unmarshaler` (pointer), `strfmt.Scannable` (pointer). No `MarshalJSON`, `sql.Scanner`, `driver.Valuer`, or text methods: JSON and SQL use the `float64` kind.

## DecimalAmount

```go
type DecimalAmount struct{ /* packed int64 */ }
type NullableDecimalAmount = nullable.Type[DecimalAmount]

const MaxDecimalAmountScale = 18
```

The value is `Coefficient × 10^-Scale`, packed into one `int64`: the low 5 bits hold the scale 0 to `MaxDecimalAmountScale` and the high 59 bits the signed coefficient within ±(2^58-1), about ±2.88×10^17. Finite values are exact, and the scale is kept, so `1.50` (scale 2) and `1.5` (scale 1) are different representations of the same value. The reserved scale 31 encodes NaN, +Inf, and -Inf. Arithmetic never panics on data: overflow yields ±Inf, `x/0` yields ±Inf, `0/0` NaN, and NaN propagates. Out of range arguments panic: a scale or `decimals` outside 0 to 18, and a coefficient outside ±(2^58-1) for `NewDecimalAmount`. The zero value is 0 with scale 0.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NewDecimalAmount(coefficient int64, scale int) DecimalAmount` | `coefficient × 10^-scale`: `NewDecimalAmount(1999, 2)` is 19.99. |
| `ParseDecimalAmount(str, acceptedDecimals ...int) (DecimalAmount, error)` | Exact parse from the digits with the separator rules of `float.ParseDetails`. Reads `NaN`, `Inf`, `+Inf`, `-Inf`, `Infinity`, `-Infinity`, and exponents (`1.5e2` is 150). `acceptedDecimals` is matched against the scale of the result, the non-finite tokens are errors with `acceptedDecimals`. Error for more than 18 decimals or a value beyond the coefficient range. |
| `DecimalAmountFrom[T DecimalAmountConvertible](value T) DecimalAmount` | Integers exactly with scale 0, floats (`float32`, `float64`, `Amount`, `Rate`) as their shortest decimal, rounded half away from zero at scale 18 if it needs more. NaN and ±Inf map to the non-finite states, out of range values to ±Inf. |
| `DecimalAmountConvertible`                         | Type constraint: the integer and float types, `Amount`, and `Rate`. |
| `DecimalAmountNaN() DecimalAmount`, `DecimalAmountInf(sign int) DecimalAmount` | NaN, +Inf for `sign >= 0`, else -Inf.              |
| `NullableDecimalAmountFromPtr(ptr *DecimalAmount) NullableDecimalAmount` | Null for a nil pointer.                            |
| `Nullable() NullableDecimalAmount`, `Ptr() *DecimalAmount` | A non-null wrapper, a pointer to a copy.           |
| `Coefficient() int64`, `Scale() int`               | The unscaled integer and the number of decimals, meaningful for finite values. |
| `Add(b)`, `Sub(b)`                                 | Exact with the larger scale. A sum that doesn't fit exactly yields ±Inf. |
| `MulInt(n int)`, `MulInt64(n int64)`               | Exact with the scale of `a`, fewer decimals only by dropping trailing zeros. Overflow yields ±Inf. |
| `Mul(b, rounding)`                                 | Exact product with scale `a.Scale()+b.Scale()` in 128-bit arithmetic. `rounding` only applies if the product needs more digits than fit. |
| `Div(b, rounding)`                                 | A quotient that terminates within 18 decimals is exact, a non-terminating one is rounded once with `rounding` at the largest scale that fits: `1/3` is `0.33333333333333333`. |
| `Neg()`, `Abs()`                                   | Sign changes keeping the scale. `Abs` of ±Inf is +Inf. |
| `MultipliedByRate(rate Rate)`, `DividedByRate(rate Rate)`, `Percentage(percent float64)` | Computed exactly with `Mul` or `Div` and the shortest decimal of the `float64` rate, without trailing zeros: `NewDecimalAmount(300, 2).MultipliedByRate(0.145)` is `0.435`, not the `0.43499999999999994` of a `float64` product. A rate with more than 18 decimals or out of the coefficient range is computed in `float64`. A result with more than 18 decimals is rounded half away from zero, so rounding it again to cents can differ by one cent from rounding the exact result. Round afterwards. |
| `RoundToDecimals(decimals int, rounding RoundingMode)` | Result has scale `decimals`, padding with zeros if it grows. Non-finite values unchanged, padding overflow yields ±Inf: `NewDecimalAmount(1, 0).RoundToDecimals(18, mode)` is +Inf. Panics for `decimals` outside 0 to 18. |
| `RoundToCents(rounding)`, `RoundToInt(rounding)`   | `RoundToDecimals(2, rounding)`, `RoundToDecimals(0, rounding)`. |
| `Cmp(b) int`, `Equal(b) bool`                      | Compare by value ignoring the scale. Total order `-Inf < finite < +Inf < NaN`, and NaN equals NaN. |
| `Sign() int`, `Signbit() bool`, `IsZero() bool`    | -1, 0, +1 (0 for NaN), negative including -Inf, exactly finite zero. |
| `IsFinite() bool`, `Valid() bool`                  | Neither NaN nor ±Inf.                              |
| `IsNaN() bool`, `IsInf(sign int) bool`             | `IsInf(+1)` tests +Inf, `IsInf(-1)` -Inf, `IsInf(0)` either. |
| `Float() float64`, `Amount() Amount`               | Conversion that may lose precision, non-finite states map to NaN and ±Inf. |
| `CentAmount(rounding RoundingMode) (CentAmount, error)` | Rounded to cents. Error for non-finite values and values beyond the `CentAmount` range. |
| `String() string`                                  | Exact digits with the scale and a dot: `"1234.50"`, `"NaN"`, `"Inf"`, `"-Inf"`. |
| `FormatSep(thousandsSep, decimalSep rune) string`  | Exact digits with separators, 0 for `decimalSep` means `'.'`. |
| `GoString() string`                                | `money.NewDecimalAmount(123450, 2)`, `money.DecimalAmountNaN()`. |
| `Format(f fmt.State, verb rune)`                   | `fmt` verbs: `%v` and `%s` like `String`, `%#v` like `GoString`, `%q` quoted, `%f` with the precision as decimals (default: the scale), `%d` rounded to an integer. Flags `+` and space for the sign, except `+` for `%v`, which means field names like in `fmt`: `%+v` of a struct prints `{T:1234.50}`. `#` groups with `,`, width with `-` and `0`. Display rounding is half away from zero. |
| `MarshalJSON`, `UnmarshalJSON`, `MarshalText`, `UnmarshalText`, `Scan`, `Value` | See [Serialization](#serialization).               |
| `MarshalBinary() ([]byte, error)`, `(*DecimalAmount).UnmarshalBinary(data []byte) error` | The packed `int64` as 8 big-endian bytes.          |
| `(*DecimalAmount).ScanString(source string, validate bool) error` | `ParseDecimalAmount`, with `validate` non-finite values are errors. |
| `DecimalAmount.JSONSchema() *jsonschema.Schema`    | A JSON number schema.                              |

Implements: `fmt.Stringer`, `fmt.GoStringer`, `fmt.Formatter`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `encoding.BinaryMarshaler`, `encoding.BinaryUnmarshaler`, `sql.Scanner`, `driver.Valuer`, `strfmt.Scannable`. The unmarshalers, `Scan`, and `ScanString` have pointer receivers.

Calculate at full precision and round once at the end. `Add` and `Sub` keep the larger scale, `Mul` adds the scales, and `Div` extends the scale as needed:

```go
price := money.NewDecimalAmount(1999, 2)                           // 19.99
net := price.MulInt(7)                                             // 139.93
gross := net.Mul(money.NewDecimalAmount(119, 2), money.RoundHalfAwayFromZero) // 166.5167
perMonth := gross.Div(money.DecimalAmountFrom(12), money.RoundHalfAwayFromZero) // 13.8763916666666667
cents := perMonth.RoundToCents(money.RoundHalfToEven)              // 13.88
```

### RoundingMode

```go
type RoundingMode uint8
```

| Constant                | Rounds                                             |
|-------------------------|----------------------------------------------------|
| `RoundHalfAwayFromZero` | To nearest, ties away from zero: 0.5 to 1, -0.5 to -1. The zero value. |
| `RoundHalfToEven`       | To nearest, ties to even (banker's rounding): 0.5 to 0, 1.5 to 2. |
| `RoundHalfUp`           | To nearest, ties toward +Inf.                      |
| `RoundHalfDown`         | To nearest, ties toward -Inf.                      |
| `RoundDown`             | Toward zero (truncate).                            |
| `RoundUp`               | Away from zero.                                    |
| `RoundFloor`            | Toward -Inf.                                       |
| `RoundCeil`             | Toward +Inf.                                       |

`String()` returns the constant name, or `Unknown rounding mode: 8` for other values. Used by `DecimalAmount` rounding, `Mul`, `Div`, `CentAmount` conversions, and `ParseCentAmount`.

### Constructor convention

A `New*` function does real work and returns a value: `NewDecimalAmount` checks its arguments, `NewCurrencyAmount` composes a pair. There are no `New*` functions that only allocate a pointer, Go 1.26 `new(money.Amount(1.23))` does that inline. `NewAmountParser` and `NewCurrencyAmountParser` return pointers because only the pointer types implement `strfmt.Parser`.

## CentAmount

```go
type CentAmount int64
type NullableCentAmount = nullable.Type[CentAmount]

const (
    MinCentAmount CentAmount = math.MinInt64 + 1
    MaxCentAmount CentAmount = math.MaxInt64
)
```

Whole cents, hundredths of a currency unit, as `int64`. Every value from `MinCentAmount` to `MaxCentAmount` is valid, there are no non-finite states. `MinCentAmount` is one above `math.MinInt64`, so negating a valid amount never overflows.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseCentAmount(str, rounding, acceptedDecimals ...int) (CentAmount, error)` | Exact parse from the digits with the separator rules of `float.ParseDetails`, rounding beyond 2 decimals with `rounding`. Accepts any number of decimals and exponents (`1.5e2` is 150.00, matched as 0 decimals against `acceptedDecimals`). NaN and Inf are errors. Covers the full range, so `ParseCentAmount(c.String(), mode)` round-trips every value. |
| `CentAmountFromPtr(ptr *CentAmount, defaultVal CentAmount) CentAmount` | `*ptr`, or `defaultVal` for nil.                   |
| `NullableCentAmountFrom(value int64) NullableCentAmount` | A non-null wrapper. Takes an `int64` cent count.   |
| `NullableCentAmountFromPtr(ptr *CentAmount) NullableCentAmount` | Null for a nil pointer.                            |
| `Ptr() *CentAmount`                                | Pointer to a copy.                                 |
| `(*CentAmount).CentAmountOr(defaultVal) CentAmount`, `CentsOr(defaultVal int64) int64`, `StringOr(defaultVal string) string` | The pointed to value, or the default for nil.      |
| `Cents() int64`                                    | The cent count.                                    |
| `Amount() Amount`                                  | `float64` amount, inexact above 2^53 cents.        |
| `DecimalAmount() DecimalAmount`                    | Exact with scale 2, ±Inf beyond about ±2.88×10^17 cents. |
| `RoundToInt() CentAmount`                          | Half away from zero to whole units: 150 cents is 200, 149 is 100. |
| `WithinOneCent(b CentAmount) bool`                 | The cent counts differ by at most 1, without overflow. |
| `String() string`                                  | Always 2 decimals with a dot: `"1234.50"`.         |
| `FormatSep(thousandsSep, decimalSep rune) string`  | Always 2 decimals, 0 for `decimalSep` means `'.'`: `FormatSep('\'', '.')` is `"-1'234.56"` for -123456. |
| `GoString() string`                                | `money.CentAmount(-150)`.                          |
| `(*CentAmount).Equal(b *CentAmount) bool`          | Both nil, the same pointer, or equal values.       |
| `Sign() int`, `IsZero() bool`, `Signbit() bool`    | -1, 0, +1, zero, negative.                         |
| `Abs()`, `Copysign(sign CentAmount)`               | `math.MinInt64` stays negative because it is out of range. |
| `(*CentAmount).Invert()`, `Inverted()`, `WithPosSign(positive bool)`, `WithNegSign(negative bool)` | Sign changes.                                      |
| `MultipliedByRate(rate Rate)`, `DividedByRate(rate Rate)`, `Percentage(percent float64)` | Computed exactly with the shortest decimal of the `float64` rate and rounded half away from zero to cents: `CentAmount(100).MultipliedByRate(0.145)` is 15, like `Percentage(14.5)`. NaN yields 0, out of range and division by zero clamp to `MinCentAmount`/`MaxCentAmount`. |
| `SplitEqually(count int) []CentAmount`             | Exact parts that sum up to the amount, the indivisible cents go to the last parts: 100 in 3 is `[33 33 34]`. Nil for `count < 1`. |
| `SplitProportionally(weights []CentAmount) []CentAmount` | Exact parts by the largest remainder method with the weight magnitudes, each within one cent of its share and with the sign of the amount. All zero weights put everything in the last part. Computed in 128-bit integers, with `big.Int` if the weight sum overflows `uint64`. |
| `(*CentAmount).ScanString(source string, validate bool) error` | `ParseCentAmount` with `RoundHalfAwayFromZero`. Ignores `validate`, every parsed value is valid. |

Implements: `fmt.Stringer`, `fmt.GoStringer`, `strfmt.Scannable` (pointer). No JSON, SQL, or text methods, deliberately: JSON and SQL use the `int64` kind, see [Serialization](#serialization).

## Currency

```go
type Currency string
const CurrencyNull = ""
```

An ISO 4217 alphabetic code. The package has a typed constant for each of 171 codes, from `AED` to `ZWL`, with the English name as comment. `Currency("").Enums()` returns all of them in alphabetical order and `EnumStrings()` the same as strings. The list contains:

- all current national ISO 4217 codes, including `MRU`, `SLE`, `SSP`, `STN`, `VED`, `VES`, `XCG`, and `ZWG`, plus `XAF`, `XCD`, `XDR`, `XOF`, and `XPF`,
- withdrawn codes kept valid for old data: `ANG`, `BGN`, `CUC`, `HRK`, `MRO`, `SLL`, `STD`, `VEF`, `ZWD`, and `ZWL` (Zimbabwe Dollar 2009 to 2024),
- codes that are not part of ISO 4217: `GGP`, `IMP`, `JEP` (Guernsey, Isle of Man, Jersey pounds), `SPL` (Seborga Luigino), and `TVD` (Tuvalu Dollar).

Fund codes like `BOV` and `CHE`, precious metals like `XAU`, bond units, `XSU`, `XUA`, `XTS`, `XXX`, and crypto codes like `BTC` are not included. To add a code, see [Add an enum code](../docs/howto-add-an-enum-code.md).

`MinorUnits()` returns the ISO 4217 decimal places: 0 for `BIF`, `CLP`, `DJF`, `GNF`, `ISK`, `JPY`, `KMF`, `KRW`, `PYG`, `RWF`, `UGX`, `VND`, `VUV`, `XAF`, `XOF`, and `XPF`, 3 for `BHD`, `IQD`, `JOD`, `KWD`, `LYD`, `OMR`, and `TND`, and 2 for all other codes, for `XDR`, the non-ISO codes, and invalid codes.

`Normalized` trims whitespace and returns the first match of:

1. A symbol, case-sensitive: `€`, `$`, `US$`, `A$`, `Can$`, `C$`, `HK$`, `NZ$`, `S$`, `£`, `GB£`, `₣`, `Ft`, `kn`, `¥`.
2. For more than 3 characters, case-insensitive: the English name of `EnglishName` (`"Canada Dollar"`) or the ISO 4217 name of two or more words (`"Canadian Dollar"`, `"Pound Sterling"`, `"Egyptian Pound"`).
3. For more than 3 characters, a guess by unit: `EUR` for the words `EUR`, `Euro`, or `Euros`, but not for words that only contain `EUR` like `TEUR` (thousands of euros) or `Amateur`. `CHF` for the words `Franc`, `Francs`, or `Franken`, `GBP` for `Pound`, `Pounds`, `Pfund`, or `Sterling`, and `USD` for `Dollar` or `Dollars`. The unit may be qualified with its country, `Swiss`, `Schweizer`, `Schweizerische(r)`, or `suisse(s)` for `CHF`, `British`, `Great British`, or `livre` for `GBP`, `US`, `USA` (`"U.S."`, `"US-Dollar"`), or `United States` for `USD`, or with the ISO code (`"Dollars (USD)"`, `"CHF Francs"`). A unit qualified with anything else is an error: `"American Dollar"`, `"Franc CFA"`. So is a country without a unit: `"Made in USA"`, `"Credit Suisse"`.
4. A valid code in any letter case: `"eur"`.

Single word ISO names like `Yen`, `Rand`, or `Won` are errors, because they are common words.

| Function / Method                                | Description                                        |
|--------------------------------------------------|----------------------------------------------------|
| `NormalizeCurrency(str) (Currency, error)`       | `Currency(str).Normalized()`.                      |
| `StringIsCurrency(str) bool`                     | `NormalizeCurrency` succeeds.                      |
| `Normalized() (Currency, error)`                 | See the rules above. Returns the input unchanged with the error. |
| `NormalizedOrNull() NullableCurrency`            | Normalized, or `CurrencyNull` if invalid.          |
| `Valid() bool`, `Validate() error`               | `Normalized` succeeds, so `"eur"` and `"€"` are valid, `""` is not. |
| `ValidAndNormalized() bool`                      | Valid and already the upper case code.             |
| `(*Currency).ValidPtr() bool`                    | Not nil and valid.                                 |
| `(*Currency).GetOrDefault(defaultVal) Currency`  | The value if the pointer is not nil and valid, else the default. Not normalized. |
| `IsEUR() bool`                                   | Normalizes to `EUR`.                               |
| `NullableCurrency() NullableCurrency`            | Type conversion.                                   |
| `Symbol() string`                                | `€`, `$`, `£`, `₣`, `¥` for EUR, USD, GBP, CHF, JPY, the code for all others. Not normalized. |
| `EnglishName() string`                           | Like `"Euro Member Countries"`, `"Canada Dollar"`, `""` for invalid codes. Not normalized. |
| `MinorUnits() int`                               | See above. Normalizes first: `Currency("jpy").MinorUnits()` is 0. |
| `String() string`                                | Normalized if valid, else unchanged.               |
| `Enums() []Currency`, `EnumStrings() []string`   | All 171 codes, generated by go-enum.               |
| `(*Currency).ScanString(source, validate) error` | Assigns the normalized source. An invalid source is an error with `validate`, else assigned unchanged. |
| `(*Currency).Scan(value) error`, `Value()`       | See [Serialization](#serialization).               |

Implements: `fmt.Stringer`, `sql.Scanner` (pointer), `driver.Valuer`, `json.Marshaler`, `encoding.TextMarshaler`, `strfmt.Scannable` (pointer), `types.NormalizableValidator[Currency]`. No `UnmarshalJSON` and no `JSONSchema`.

## NullableCurrency

```go
type NullableCurrency string
```

A `Currency` where `""` (`CurrencyNull`) is the null value. A whitespace only value is null too, like for `country.NullableCode`. Null counts as valid.

| Method                                             | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `IsNull() bool`, `IsNotNull() bool`                | Null is `""` or whitespace only.                   |
| `Get() Currency`                                   | Panics if null.                                    |
| `GetOr(defaultVal Currency) Currency`              | The value if valid and not null, else the default. |
| `Currency() Currency`                              | Type conversion, also for null.                    |
| `Set(Currency)`, `(*NullableCurrency).SetNull()`   | `Set("")` sets null.                               |
| `StringOr(defaultString) string`                   | Normalized, or the default for null or invalid.    |
| `String() string`                                  | Normalized if valid, else unchanged, `""` for null. |
| `Valid() bool`, `Validate() error`                 | Null or a valid currency.                          |
| `ValidAndNotNull() bool`                           | Not null and valid.                                |
| `ValidAndNormalized() bool`                        | `""`, or valid and normalized. False for a whitespace only null, which normalizes to `""`. |
| `(*NullableCurrency).ValidPtr() bool`              | Nil, null, or valid.                               |
| `Normalized() (NullableCurrency, error)`, `NormalizedOrNull()` | Like `Currency`, null stays null.                  |
| `Symbol() string`, `EnglishName() string`          | Like `Currency`.                                   |
| `(*NullableCurrency).ScanString(source, validate) error` | `""`, whitespace only, `NULL`, `null`, and `nil` set null, other values like `Currency.ScanString`. |
| `(*NullableCurrency).Scan(value) error`, `Value()`, `MarshalJSON()`, `(*NullableCurrency).UnmarshalJSON(data) error` | See [Serialization](#serialization).               |

Implements: `fmt.Stringer`, `sql.Scanner` (pointer), `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler` (pointer), `encoding.TextMarshaler`, `strfmt.Scannable` (pointer), `nullable.Nullable`, `nullable.NullSetable[Currency]` (pointer). No `MinorUnits`, use `Get().MinorUnits()`.

## CurrencyAmount and CurrencyDecimalAmount

```go
type CurrencyAmount struct {
    Currency Currency
    Amount   Amount
}

type CurrencyDecimalAmount struct {
    Currency Currency
    Amount   DecimalAmount
}
```

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NewCurrencyAmount(currency, amount)`, `CurrencyAmountUSD(amount)`, `CurrencyAmountEUR`, `CurrencyAmountCHF`, `CurrencyAmountGBP`, `CurrencyAmountJPY` | Composite literals.                                |
| `NewCurrencyDecimalAmount(currency, amount)`, `CurrencyDecimalAmountUSD(amount)`, `CurrencyDecimalAmountEUR`, `CurrencyDecimalAmountCHF`, `CurrencyDecimalAmountGBP`, `CurrencyDecimalAmountJPY` | The same for `DecimalAmount`.                      |
| `ParseCurrencyAmount(str, acceptedDecimals ...int) (CurrencyAmount, error)` | The text before the first space, separator, sign, or digit is normalized as currency, or else the text after the last one. If neither is a currency, the whole string is parsed as amount without currency. The amount is parsed with `ParseAmount`. `"EUR 1.234,56"`, `"1,234.56 USD"`, `"€12"`, `"CHF1'234.50"`, and `"1234.56"` work. |
| `ParseCurrencyDecimalAmount(str, acceptedDecimals ...int) (CurrencyDecimalAmount, error)` | The same with `ParseDecimalAmount`, exact.         |
| `CurrencyAmount.String() string`                   | Normalized code, space, and the amount rounded to `Currency.MinorUnits()` decimals without grouping: `"EUR 1234.50"`, `"JPY 1235"`, `"KWD 1.235"`. Without currency only the amount with 2 decimals. |
| `CurrencyAmount.FormatSep(currencyFirst bool, thousandsSep, decimalSep rune, precision int) string` | The amount like `Amount.FormatSep`, the normalized code before or after it, no code for an empty currency. |
| `CurrencyDecimalAmount.String() string`            | Normalized code, space, and the exact amount with its scale: `"EUR 1234.50"`. |
| `CurrencyDecimalAmount.FormatSep(currencyFirst bool, thousandsSep, decimalSep rune) string` | The amount like `DecimalAmount.FormatSep`.         |
| `CurrencyDecimalAmount.CurrencyAmount() CurrencyAmount` | Converts the amount to `float64`.                  |
| `GoString() string`                                | `{Currency: "EUR", Amount: money.Amount(1.5)}`.    |
| `(*CurrencyAmount).ScanString(source, validate) error` | `ParseCurrencyAmount` accepting 0, 2, or the currency's minor units decimals: `"EUR 1.234"` is an error, `"KWD 1.234"` is valid. With `validate` NaN and ±Inf are errors. |
| `(*CurrencyDecimalAmount).ScanString(source, validate) error` | `ParseCurrencyDecimalAmount` with any decimals up to 18. With `validate` non-finite values are errors. |
| `Scan(value) error`, `Value()`                     | See [Serialization](#serialization).               |

Implements: `fmt.Stringer`, `fmt.GoStringer`, `sql.Scanner` (pointer), `driver.Valuer`, `strfmt.Scannable` (pointer). No JSON methods.

## Rate

```go
type Rate float64
type NullableRate = nullable.Type[Rate]
```

An exchange rate or a percentage as fraction: 19% is `Rate(0.19)`.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseRate(str, acceptedDecimals ...int) (Rate, error)` | Like `ParseAmount`, and a trailing `%`, also after a space, divides by 100: `"19%"` is 0.19, `"19"` is 19. `acceptedDecimals` is matched before the division. |
| `RateFromPtr(ptr *Rate, defaultVal Rate) Rate`     | `*ptr`, or `defaultVal` for nil.                   |
| `NullableRateFrom(value float64) NullableRate`, `NullableRateFromPtr(ptr *Rate) NullableRate` | Non-null from a `float64`, null for a nil pointer. |
| `Ptr() *Rate`                                      | Pointer to a copy.                                 |
| `(*Rate).RateOr(defaultVal) Rate`, `FloatOr(defaultVal float64) float64` | The pointed to value or the default.               |
| `(*Rate).StringOr(defaultVal string) string`       | `strconv.FormatFloat(r, 'f', -1, 64)` or the default. |
| `RoundToInt() Rate`, `RoundToDecimals(decimals int) Rate` | Like `Amount`.                                     |
| `FormatSep(thousandsSep, decimalSep rune, precision int) string` | Like `Amount.FormatSep`: `Rate(0.19).FormatSep(0, ',', 2)` is `"0,19"`. |
| `GoString() string`                                | The exact decimal expansion without type name, like `0.1000000000000000055511151231257827021181583404541015625`. |
| `BigFloat() *big.Float`                            | A new `big.Float`.                                 |
| `(*Rate).Equal(other *Rate) bool`                  | Both nil, the same pointer, or equal values.       |
| `Signbit()`, `Copysign(sign)`, `Abs()`, `AbsFloat()`, `(*Rate).Invert()`, `Inverted()`, `WithPosSign(bool)`, `WithNegSign(bool)` | Like `Amount`.                                     |
| `Inverse() Rate`                                   | `1 / r`: `Rate(4).Inverse()` is 0.25.              |
| `Valid()`, `ValidAndPositive()`, `ValidAndGreaterZero()`, `ValidAndSmallerZero()`, `ValidAndHasSign(sign)` | Like `Amount`.                                     |
| `(*Rate).ScanString(source, validate) error`       | `ParseRate`, with `validate` NaN and ±Inf are errors. |
| `(*Rate).UnmarshalJSON(j) error`                   | See [Serialization](#serialization).               |

Implements: `fmt.GoStringer`, `json.Unmarshaler` (pointer), `strfmt.Scannable` (pointer). No `String` method, `fmt` prints the `float64`. No `MarshalJSON`, SQL, or text methods.

## Nullable amount types

`NullableAmount`, `NullableDecimalAmount`, `NullableCentAmount`, and `NullableRate` are aliases of `nullable.Type[T]`, whose zero value is null. They have `IsNull`, `IsNotNull`, `Get` (panics for null), `GetOr`, `Set`, `SetNull`, `Ptr`, `Scan`, `Value`, `MarshalJSON`, `UnmarshalJSON`, and `JSONSchema`, see the [`nullable`](../nullable/README.md) package. The constructors are listed with each type above. They have no `ScanString`, [`strfmt`](../strfmt/README.md) `Scan` sets null for a nil string and else the value scanned like the wrapped type, so `"1,5"` is 1.5.

## Parsers and finder

| Identifier                                         | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NewAmountParser(acceptedDecimals ...int) *AmountParser` | `(*AmountParser).Parse(str, langHints...)` parses with `ParseAmount` and returns the amount with a dot, no grouping, and the largest accepted number of decimals, or 2 without `acceptedDecimals`: `NewAmountParser().Parse("1,2378")` is `"1.24"`. |
| `NewCurrencyAmountParser(acceptedDecimals ...int) *CurrencyAmountParser` | `(*CurrencyAmountParser).Parse` parses with `ParseCurrencyAmount` and returns code first: `"1.234,56 eur"` is `"EUR 1234.56"`. |
| `CurrencyParser`                                   | `CurrencyParser{}.Parse(str, langHints...)` returns the normalized code. |
| `AmountFinder`                                     | Variable whose `FindAllIndex(str []byte, n int) [][]int` implements `types.Finder`. Splits the text at whitespace and `:`, trims `.`, `,`, and `;` from the words, and returns the words that `StringIsAmount(word, false)` accepts. `n < 0` returns all, `n == 0` none, `n > 0` at most `n`. |
| `StringIsAmount(str string, acceptInt bool) bool`  | `str` has exactly 2 decimals with `.` or `,`, optional grouping of the integer part with the other one and an optional `-`, like `"1.234,56"` or `"-1234.56"`, or is `NaN`, `Inf`, `+Inf`, `-Inf`. With `acceptInt` also plain integers. `"12.5"` and `"+12.50"` are false. |

The `langHints` of all parsers are ignored. Only the pointer types `*AmountParser` and `*CurrencyAmountParser` implement `strfmt.Parser`.

## Serialization

| Type                                               | SQL `Value()`                                      | SQL `Scan` accepts                                 | JSON written                                       | JSON accepted                                      |
|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|
| `Amount`, `Rate`                                   | No `Valuer`: `database/sql` passes the `float64` to the driver, also NaN and ±Inf. | No `Scanner`: `database/sql` converts `float64`, `int64`, and number strings like `"1234.5"`. NULL is an error. | No `MarshalJSON`: a JSON number of the `float64`, `1234.5`, not rounded. NaN and ±Inf fail with `json: unsupported value`. | Numbers and strings parsed with `ParseAmount`/`ParseRate` after decoding escapes: `"1.234,50"`, `"NaN"`, for `Rate` also `"19%"`. `null` and `""` set 0. |
| `DecimalAmount`                                    | The exact decimal string: `"1234.50"`, `"NaN"`, `"Infinity"`, `"-Infinity"`. Fits `numeric`, the literal also works for `float8` and `text`. | `string` and `[]byte` parsed with `ParseDecimalAmount`, `int64` with scale 0, `float64` as shortest decimal. A `float64` too large is an error, one needing more than 18 decimals is rounded, so `1e-300` is 0. NULL is an error. | A JSON number keeping the scale: `1234.50`, `0.00`. `"NaN"`, `"Inf"`, `"-Inf"` as strings. | Numbers including exponents, strings parsed with `ParseDecimalAmount` after decoding escapes, `null` and `""` as 0. Every written value decodes to the same coefficient and scale. |
| `CentAmount`                                       | No `Valuer`: `database/sql` writes the `int64` cent count, use an integer column. | No `Scanner`: `int64`, integer strings, and `float64` values without fraction. `1.5` and NULL are errors. | No `MarshalJSON`: the cent count, `123450` for 1234.50. | Integers only. `123.45` and `"123.45"` are errors, `null` leaves the value unchanged. |
| `Currency`                                         | `""` as NULL, else normalized if valid, else unchanged: `"eur"` as `"EUR"`. | `string` and `[]byte` unchanged, NULL as `""`, other types are errors. No normalization. | `MarshalJSON` and `MarshalText` like `Value`, `""` as `""`. | Strings unchanged without normalization or validation, `null` leaves the value unchanged. |
| `NullableCurrency`                                 | Null (`""` or whitespace only) as NULL, else like `Currency`. | Like `Currency`, NULL as null.                     | `MarshalJSON`: null (`""` or whitespace only) as `null`, else like `Currency`. `MarshalText`: null as `""`, else like `Currency`. | `UnmarshalJSON`: `null` sets null, strings unchanged without normalization or validation, other JSON values are errors. |
| `CurrencyAmount`                                   | `String()`: `"EUR 1234.50"`, `"JPY 1235"`, `"0.00"` without currency, the currency normalized if valid. Needs a text column. | `string` and `[]byte` parsed with `ParseCurrencyAmount` with any decimals, `float64` as amount without currency. `int64` and NULL are errors. | No JSON methods: `{"Currency":"EUR","Amount":1234.5}`, the currency normalized by `Currency.MarshalJSON`. | The same object.                                   |
| `CurrencyDecimalAmount`                            | `String()`: `"EUR 1234.50"` with the amount's scale, the currency normalized if valid. Needs a text column. | `string` and `[]byte` parsed with `ParseCurrencyDecimalAmount`, `float64` like `DecimalAmount.Scan` without currency. `int64` and NULL are errors. | No JSON methods: `{"Currency":"EUR","Amount":1234.50}`, the currency normalized by `Currency.MarshalJSON`. | The same object.                                   |
| `NullableAmount`, `NullableRate`, `NullableCentAmount`, `NullableDecimalAmount` | NULL for null, else the value of the wrapped type: `float64`, `int64`, or the `DecimalAmount` string. | NULL as null. Other values are converted by the `database/sql` rules into the wrapped type, so `"1,5"` is an error for `NullableAmount` and `"19%"` for `NullableRate`, while `NullableDecimalAmount` uses `DecimalAmount.Scan`. | `null` for null, else like the wrapped type.       | `null` as null, else like the wrapped type: `""` is a non-null 0 for `NullableAmount`, `NullableRate`, and `NullableDecimalAmount`. |

## Gotchas

- **`"1,000"` is 1.00 in every parser.** `ParseAmount`, `ParseDecimalAmount`, `ParseCentAmount`, `ParseRate`, `ParseCurrencyAmount`, the `ScanString` methods, and JSON strings read a single `.` or `,` as decimal separator. Pass `acceptedDecimals` to reject it: `ParseAmount("1,000", 0, 2)` is an error because `"1,000"` has 3 decimals.
- **`DecimalAmount` `==` compares the packed bits.** `NewDecimalAmount(150, 2) == NewDecimalAmount(15, 1)` is false, and so are map keys, `reflect.DeepEqual`, and testify's `assert.Equal` for them. Use `Equal` or `Cmp`, which also treat NaN as equal to NaN.
- **`DecimalAmount` precision is limited.** The coefficient has about 18 digits: at scale 2 the largest value is 2882303761517117.43, about 2.88×10^15. Beyond that `Add` yields ±Inf, `Mul` and `Div` round to fewer decimals until the integer part doesn't fit either, `ParseDecimalAmount` returns an error, and `NewDecimalAmount` panics. Use `CentAmount` for larger two-decimal values. `MultipliedByRate`, `DividedByRate`, and `Percentage` fall back to a `float64` product for a rate with more than 18 decimals.
- **NaN and Inf.** `ParseAmount`, `ParseRate`, and `ParseDecimalAmount` accept `NaN` and `Inf` without error if no `acceptedDecimals` are passed, and reject them if any are. `json.Marshal` fails on such `Amount` and `Rate` values. `Amount.CentAmount` maps NaN to 0 and clamps Inf. Check `Valid()` after parsing untrusted input.
- **`CurrencyAmount` needs a text column.** Its `Value` is `"EUR 1234.50"`. A `numeric` column rejects it, and an `int64` from the database can't be scanned.
- **`CurrencyAmount.String` depends on the currency.** `CurrencyAmountJPY(1234.5)` is `"JPY 1235"` and the stored SQL text is rounded the same way. `String`, `Value` and JSON normalize the currency, so `CurrencyAmount{Currency: "jpy", Amount: 1234.5}` is `"JPY 1235"`. `Symbol` and `EnglishName` don't normalize. Normalize the currency first.
- **`CurrencyAmount.ScanString` is stricter than `Scan`.** `ScanString("EUR 1.234")` is an error, `Scan("EUR 1.234")` assigns 1.234.
- **`ParseCurrencyAmount` needs a one-word currency.** `"12 Canadian Dollar"` is an error, although `NormalizeCurrency("Canadian Dollar")` is `CAD`. A string without currency parses with an empty `Currency`.
- **Decoded currencies are unchecked.** JSON and `Scan` assign `"eur"` or `"XYZ"` unchanged. Call `Normalized` or `Validate` after decoding.
- **JSON `null` keeps the old value** for `CentAmount` and `Currency`, which have no `UnmarshalJSON`, while `Amount`, `Rate`, and `DecimalAmount` set 0 and `NullableCurrency` sets null.
- **`Amount.Cents` is undefined for NaN and ±Inf**, the result depends on the platform. Use `CentAmount(mode)`.
- **`StringIsAmount` and `AmountFinder` only detect 2-decimal amounts.** `"12.5"`, `"(3,00)"`, and `"EUR4,00"` are not found, `Inf` in a text is.
- **`acceptedDecimals` of `ParseRate` counts before the `%` division.** `ParseRate("19%", 2)` is an error because `"19"` has no decimals.
- **`Amount` splits round to cents.** The parts sum up to the amount rounded to cents, not to the unrounded amount: `Amount(0.005).SplitEqually(2)` is `[0.00 0.01]`.

## Examples

Parse amounts and currencies of unknown locale:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/money"
)

func main() {
	a, err := money.ParseAmount("1.234,56")
	fmt.Println(a, err)
	a, err = money.ParseAmount("1,000")
	fmt.Println(a, err)
	_, err = money.ParseAmount("1,000", 0, 2)
	fmt.Println(err != nil)

	ca, err := money.ParseCurrencyAmount("1.234,56 eur")
	fmt.Println(ca, err)

	c, err := money.NormalizeCurrency("Canadian Dollar")
	fmt.Println(c, err)
	_, err = money.NormalizeCurrency("American Dollar")
	fmt.Println(err)

	fmt.Println(money.JPY.MinorUnits(), money.CurrencyAmountJPY(1234.5), money.CurrencyAmount{Currency: money.KWD, Amount: 1.2345})

	// Output:
	// 1234.56 <nil>
	// 1.00 <nil>
	// true
	// EUR 1234.56 <nil>
	// CAD <nil>
	// invalid currency code "AMERICAN DOLLAR"
	// 0 JPY 1235 KWD 1.235
}
```

Calculate exactly and split into cents:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/money"
)

func main() {
	net := money.NewDecimalAmount(1999, 2).MulInt(7)
	vat := net.Mul(money.NewDecimalAmount(19, 2), money.RoundHalfAwayFromZero)
	gross := net.Add(vat)
	fmt.Println(net, vat, gross, gross.RoundToCents(money.RoundHalfAwayFromZero))

	a, b := money.NewDecimalAmount(150, 2), money.NewDecimalAmount(15, 1)
	fmt.Println(a == b, a.Equal(b))

	cents, err := gross.CentAmount(money.RoundHalfAwayFromZero)
	fmt.Println(cents, err, cents.SplitEqually(3))
	fmt.Println(money.Amount(1).SplitEqually(3), money.Amount(0.145).RoundToCents())

	// Output:
	// 139.93 26.5867 166.5167 166.52
	// false true
	// 166.52 <nil> [55.50 55.51 55.51]
	// [0.33 0.33 0.34] 0.15
}
```

JSON and SQL forms:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/money"
)

type Invoice struct {
	Total    money.DecimalAmount    `json:"total"`
	Cents    money.CentAmount       `json:"cents"`
	Amount   money.Amount           `json:"amount"`
	Currency money.NullableCurrency `json:"currency"`
	Discount money.NullableAmount   `json:"discount"`
	Price    money.CurrencyAmount   `json:"price"`
}

func main() {
	inv := Invoice{
		Total:  money.NewDecimalAmount(123450, 2),
		Cents:  123450,
		Amount: 1234.5,
		Price:  money.CurrencyAmountEUR(1234.5),
	}
	j, err := json.Marshal(inv)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))

	var decoded Invoice
	err = json.Unmarshal([]byte(`{"total":"1.234,50","amount":"1.234,50","currency":"EUR","discount":null}`), &decoded)
	fmt.Println(decoded.Total, decoded.Amount, decoded.Currency, decoded.Discount.IsNull(), err)

	total, _ := inv.Total.Value()
	price, _ := inv.Price.Value()
	currency, _ := inv.Currency.Value()
	fmt.Printf("%#v %#v %#v\n", total, price, currency)

	// Output:
	// {"total":1234.50,"cents":123450,"amount":1234.5,"currency":null,"discount":null,"price":{"Currency":"EUR","Amount":1234.5}}
	// 1234.50 1234.50 EUR true <nil>
	// "1234.50" "EUR 1234.50" <nil>
}
```

## Related

- [`float`](../float/README.md): the separator detection and rounding behind every parser.
- [`strfmt`](../strfmt/README.md): scanning and formatting with locale presets, `MoneyFormat`.
- [`nullable`](../nullable/README.md): `nullable.Type`, the base of the nullable amount types.
- [`bank`](../bank/README.md): `bank.Account` and the CAMT.053 structs use `Currency` and `Amount`.
- [Money types and rounding](../docs/explanation-money-types-and-rounding.md): when to use which amount type, and the rounding conventions.
- [Lenient parsing](../docs/explanation-lenient-parsing.md): why `"1,000"` is 1.
- [Scan and format table rows](../docs/howto-scan-and-format-table-rows.md): money values in CSV rows.
- [Extract identifiers from text](../docs/howto-extract-identifiers-from-text.md): `AmountFinder`.
- [Store optional values](../docs/howto-store-optional-values.md): the nullable amount types and `NullableCurrency`.
- [Add an enum code](../docs/howto-add-an-enum-code.md): adding a `Currency` constant.
