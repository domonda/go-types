# float

Parsing and formatting of floating point numbers with locale-dependent separators (`1.234,56`, `1,234.56`, `1'234.56`, `1 234,56`), rounding to decimal places, NaN/Inf checks, and a float type that decodes loose JSON. The `money` and `strfmt` packages parse numbers with this package.

```
import "github.com/domonda/go-types/float"
```

## Generic helpers

All of them work on any type with the underlying type `float32` or `float64`.

| Function                                  | Description                                        |
|-------------------------------------------|----------------------------------------------------|
| `RoundToDecimals[T](f T, decimals int) T` | Rounds half away from zero to `decimals` places, see [Rounding](#rounding). Negative `decimals` round to tens, hundreds, and so on: `RoundToDecimals(1234.5, -2)` is 1200. |
| `DerefOr[T](ptr *T, defaultVal T) T`      | `*ptr`, or `defaultVal` if `ptr` is nil.           |
| `Valid[T](f T) bool`                      | `f` is neither NaN nor ±Inf.                       |
| `ValidAndHasSign[T](f T, sign int) bool`  | `Valid(f)` and `f` has the sign of `sign`. `sign == 0` matches any sign. Zero counts as positive: `ValidAndHasSign(0, +1)` is true, `ValidAndHasSign(0, -1)` is false, also for -0. |

### Rounding

`RoundToDecimals` and `Format` round half away from zero, and they round the shortest decimal representation of the float, which is the decimal literal that parses back to the same float, not its binary value:

| Call                              | Result   | `math.Round`/`strconv` on the binary value |
|-----------------------------------|----------|--------------------------------------------|
| `RoundToDecimals(1.005, 2)`       | 1.01     | 1 (binary 1.00499999999999989342)          |
| `RoundToDecimals(0.145, 2)`       | 0.15     | 0.14                                       |
| `RoundToDecimals(2.5, 0)`         | 3        | 3                                          |
| `Format(2.5, 0, '.', 0, false)`   | `"3"`    | `"2"` (`strconv` rounds ties to even)      |
| `Format(0.125, 0, '.', 2, false)` | `"0.13"` | `"0.12"`                                   |

`float32` values round their shortest `float32` decimal, so `RoundToDecimals(float32(1.005), 2)` is also 1.01. NaN, ±Inf, and values with no more decimal places than requested are returned unchanged. Only a rounded value beyond the range of `T` is ±Inf, like `math.MaxFloat64` rounded to a multiple of 10^307. The `money` package rounds `Amount` and `Rate` the same way.

## Parsing

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `Parse(str string) (float64, error)`               | `ParseDetails` without the details.                |
| `ParseDetails(str string) (f float64, thousandsSep, decimalSep rune, decimals int, err error)` | Parses `str` and returns the detected separators, 0 if not found, and the number of fractional digits before an exponent: `"1.2e6"` has 1 decimal. `decimalSep` is always 0, `'.'`, or `','`. |

Rules of `Parse` and `ParseDetails`:

- Leading and trailing whitespace is trimmed.
- `NaN`, `Inf`, `+Inf`, and `-Inf` are the only non-finite inputs, matched exactly and case-sensitive. `nan`, `Infinity`, and `-NaN` are errors, and so is an out of range value like `1e999`.
- One sign, `+` or `-`, is allowed at the start or the end, also separated by a space: `-12`, `- 12`, `12-`, and `12 -` are all -12.
- Scientific notation with `e` or `E` and an optional exponent sign: `1.5e10`, `2E-3`. Separators in the exponent are errors.
- `.` and `,` are decimal or thousands separators, `'` and space are only thousands separators.
- A single `.` or `,` is the decimal separator: `"1,234"` and `"1.234"` are 1.234, `"1.000"` is 1.
- The same `.` or `,` repeated is the thousands separator: `"1.234.567"` is 1234567.
- After a thousands separator, the other one of `.` and `,` is the decimal separator: `"1.234,5"` and `"1,234.5"` are 1234.5.
- A single `'` or space must be a thousands separator: `"1'234"` and `"1 234"` are 1234, `"1'5"` and `"1 23"` are errors.
- Every thousands separator is followed by exactly 3 digits, and the group in front of the first one has 1 to 3 digits: `"1234.567,89"` and `".234,5"` are errors.
- A space can't follow another kind of thousands separator (`"123,123 123"`), and an apostrophe can't be the decimal separator (`"1.234'5"`).
- Hexadecimal numbers and `_` digit separators are errors.

| Input            | `f`        | `thousandsSep` | `decimalSep` | `decimals` |
|------------------|------------|----------------|--------------|------------|
| `"1.234.567,89"` | 1234567.89 | `'.'`          | `','`        | 2          |
| `"1,234.5"`      | 1234.5     | `','`          | `'.'`        | 1          |
| `"1'234.50"`     | 1234.5     | `'\''`         | `'.'`        | 2          |
| `"1 234,56"`     | 1234.56    | `' '`          | `','`        | 2          |
| `"1,234"`        | 1.234      | 0              | `','`        | 3          |
| `"1'234"`        | 1234       | `'\''`         | 0            | 0          |
| `"158,00 "`      | 158        | 0              | `','`        | 2          |
| `"1,234,567e3"`  | 1234567e3  | `','`          | 0            | 0          |

## Formatting

```go
func Format[T ~float32 | ~float64](
    f T,
    thousandsSep rune, // 0, ',', '.', '\'', or ' '
    decimalSep rune,   // '.' or ','
    precision int,     // -1 for the shortest exact form, otherwise fractional digits
    padPrecision bool, // pad the fraction with zeros to precision digits
) string
```

- `precision >= 0` rounds to that many fractional digits, see [Rounding](#rounding). `precision == -1` writes the fewest digits that parse back to `f` exactly, and `padPrecision` has no effect.
- `padPrecision == true` writes exactly `precision` fractional digits, `false` trims trailing zeros and a dangling decimal separator: `Format(1.50, 0, '.', 2, false)` is `"1.5"`.
- `thousandsSep` groups the integer part in groups of 3 digits. Grouping uses the rounded digits, so `Format(999.995, ',', '.', 2, true)` is `"1,000.00"`.
- The integer part is never written in exponent form: `Format(1e21, ',', '.', -1, false)` is `"1,000,000,000,000,000,000,000"`.
- NaN and ±Inf are returned as `"NaN"`, `"+Inf"`, and `"-Inf"` without grouping or padding.
- `float32` values are formatted with 32-bit precision: `Format(float32(0.1), 0, '.', -1, false)` is `"0.1"`.
- `Format` panics for an invalid `thousandsSep` or `decimalSep`, for `thousandsSep == decimalSep`, and for `precision < -1`.

## FormatDef

```go
type FormatDef struct {
    ThousandsSep rune `json:"thousandsSep,omitempty"`
    DecimalSep   rune `json:"decimalSep"`
    Precision    int  `json:"precision"`
    PadPrecision bool `json:"padPrecision"`
}
```

The arguments of `Format` as a struct, for storing a number format in a configuration. `strfmt.FormatConfig` uses it for floats.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NewFormatDef() *FormatDef`                        | `&FormatDef{DecimalSep: '.', Precision: -1}`: no grouping, dot decimal separator, shortest exact form. |
| `(*FormatDef).Format(f float64) string`            | `Format(f, ThousandsSep, DecimalSep, Precision, PadPrecision)`. |
| `(*FormatDef).Parse(str string, langHints ...language.Code) (string, error)` | Parses `str` with `Parse` and returns it formatted with the definition: `"1.234.567,89"` becomes `"1,234,567.89"` for `{',', '.', 2, true}`. Ignores `langHints`. |
| `FormatDef.MarshalJSON() ([]byte, error)`          | See [Serialization](#serialization).               |
| `(*FormatDef).UnmarshalJSON(j []byte) error`       | See [Serialization](#serialization).               |

Implements: `json.Marshaler`, `json.Unmarshaler` (pointer), `strfmt.Parser` (pointer). The zero `FormatDef` has `DecimalSep` 0, so its `Format` panics: start from `NewFormatDef()` or set `DecimalSep`.

## Tolerant

```go
type Tolerant float64
```

A `float64` for JSON input that has numbers as strings, empty strings, or `null`.

| Method                               | Description                                        |
|--------------------------------------|----------------------------------------------------|
| `(*Tolerant).UnmarshalJSON(j) error` | See [Serialization](#serialization).               |
| `(*Tolerant).AsFloatPtr() *float64`  | The receiver as `*float64`, writes through it change the `Tolerant`. |
| `Valid() bool`                       | Neither NaN nor ±Inf.                              |
| `ValidAndGreaterZero() bool`         | Valid and `> 0`.                                   |
| `ValidAndSmallerZero() bool`         | Valid and `< 0`.                                   |
| `ValidAndHasSign(sign int) bool`     | `ValidAndHasSign(float64(f), sign)`.               |
| `IsNaN() bool`, `IsInf() bool`       | NaN, or +Inf or -Inf.                              |

Implements: `json.Unmarshaler` (pointer). There is no `MarshalJSON`.

## Serialization

The package has no SQL support.

| Type        | JSON written                                       | JSON accepted                                      |
|-------------|----------------------------------------------------|----------------------------------------------------|
| `FormatDef` | `{"thousandsSep":",","decimalSep":".","precision":2,"padPrecision":true}`. Separators are one-character strings, a zero `ThousandsSep` is omitted, a zero `DecimalSep` is `""`. | One-character strings, `""` for a zero separator, and the legacy code point strings like `"44"` for `','` that versions before the JSON methods wrote. Fields missing in the object keep their value. Strings of several characters and JSON numbers are errors. |
| `Tolerant`  | A JSON number via the `float64` kind. NaN and ±Inf fail with `json: unsupported value`. | Numbers, strings parsed with `Parse` after decoding JSON escapes (`"1.234,56"` is 1234.56), and `null` or `""` as 0. `"NaN"`, `"Inf"`, `"+Inf"`, `"-Inf"` as strings. An error leaves the value unchanged. |

## Gotchas

- **`"1,234"` is 1.234.** A single `.` or `,` is always the decimal separator, so `Format(1234, ',', '.', -1, false)` writes `"1,234"`, which `Parse` reads back as 1.234. This affects integers from 1000 to 999999 formatted with a thousands separator and without decimals. Write decimals to make the output unambiguous: `Format(1234, ',', '.', 2, true)` is `"1,234.00"`, which parses back as 1234.
- **`"1.000"` is 1**, not 1000. Parse German style integers with thousands separators only when they have at least two groups, or check the result against the expected magnitude.
- **A detected decimal separator can be 0.** `ParseDetails("1'234")` returns `decimalSep` 0, which `Format` rejects with a panic. Pick a decimal separator that differs from `thousandsSep` before passing the detected separators back to `Format`.
- **`ParseDetails` reports `'.'` for a comma in front of an exponent.** `ParseDetails("1,5e3")` returns 1500 with `decimalSep` `'.'`.
- **The zero `FormatDef` panics in `Format`**, because its `DecimalSep` is 0.
- **Negative values that round to zero keep the sign.** `Format(-0.001, 0, '.', 2, false)` is `"-0"`, like `strconv.FormatFloat`.
- **`Tolerant` can't tell missing from zero.** JSON `null`, `""`, and `0` all decode to 0. Use a pointer field if the difference matters.
- **Unquoted `NaN` needs a direct call.** `Tolerant.UnmarshalJSON([]byte("NaN"))` works, but `json.Unmarshal` rejects `NaN` as invalid JSON before calling it. Only the quoted `"NaN"` works through `encoding/json`.
- **Rounded floats are still binary floats.** `RoundToDecimals(0.145, 2)` returns the `float64` closest to 0.15. Use `money.DecimalAmount` or `money.CentAmount` for exact decimal arithmetic.

## Examples

Parse numbers of unknown locale:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/float"
)

func main() {
	for _, s := range []string{"1.234.567,89", "1,234.5", "1'234", "1 234,56", "1,234", "12 -"} {
		f, thousandsSep, decimalSep, decimals, err := float.ParseDetails(s)
		fmt.Printf("%-14q %v %q %q %d %v\n", s, f, thousandsSep, decimalSep, decimals, err)
	}
	_, err := float.Parse("1'5")
	fmt.Println(err)

	// Output:
	// "1.234.567,89" 1.23456789e+06 '.' ',' 2 <nil>
	// "1,234.5"      1234.5 ',' '.' 1 <nil>
	// "1'234"        1234 '\'' '\x00' 0 <nil>
	// "1 234,56"     1234.56 ' ' ',' 2 <nil>
	// "1,234"        1.234 '\x00' ',' 3 <nil>
	// "12 -"         -12 '\x00' '\x00' 0 <nil>
	// thousands separators have to be 3 characters apart: "1'5"
}
```

Format and round:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/float"
)

func main() {
	fmt.Println(float.Format(1234567.891, '.', ',', 2, true))
	fmt.Println(float.Format(1234.5, ',', '.', 4, true))
	fmt.Println(float.Format(1234.5, '\'', '.', 2, false))
	fmt.Println(float.Format(0.125, 0, '.', 2, false))
	fmt.Println(float.Format(1234.0, ',', '.', -1, false)) // parses back as 1.234
	fmt.Println(float.RoundToDecimals(1.005, 2), float.RoundToDecimals(-2.5, 0))

	// Output:
	// 1.234.567,89
	// 1,234.5000
	// 1'234.5
	// 0.13
	// 1,234
	// 1.01 -3
}
```

Store a number format as JSON and decode loose JSON numbers:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/float"
)

func main() {
	def := float.FormatDef{ThousandsSep: '.', DecimalSep: ',', Precision: 2, PadPrecision: true}
	j, err := json.Marshal(def)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))

	var legacy float.FormatDef
	err = json.Unmarshal([]byte(`{"thousandsSep":"44","decimalSep":"46","precision":-1}`), &legacy)
	fmt.Println(legacy.Format(1234567.5), err)

	var payment struct {
		Amounts []float.Tolerant `json:"amounts"`
	}
	err = json.Unmarshal([]byte(`{"amounts":[12.5,"1.234,56","",null,"NaN"]}`), &payment)
	fmt.Println(payment.Amounts, err)

	// Output:
	// {"thousandsSep":".","decimalSep":",","precision":2,"padPrecision":true}
	// 1,234,567.5 <nil>
	// [12.5 1234.56 0 0 NaN] <nil>
}
```

## Related

- [`money`](../money/README.md): amount and rate parsing builds on `ParseDetails`, `Amount` and `Rate` rounding and formatting on `RoundToDecimals` and `Format`.
- [`strfmt`](../strfmt/README.md): scans floats with `Parse` and formats them with a `FormatDef`.
- [Lenient parsing](../docs/explanation-lenient-parsing.md): why a single separator is the decimal separator.
- [Money types and rounding](../docs/explanation-money-types-and-rounding.md): the shortest decimal rounding convention.
- [Scan and format table rows](../docs/howto-scan-and-format-table-rows.md): `FormatDef` in `strfmt.FormatConfig`.
