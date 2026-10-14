# Money types and rounding

This page explains why the [`money`](../money/README.md) package has three amount types, how to pick one, and the default rounding rule the whole module follows. For the API of each type, see the package README.

## The problem

Money has to add up to the cent, and `float64` can't hold most cent values exactly. The literal `0.145` is stored as the binary fraction nearest to it:

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	price := 0.145
	fmt.Println(strconv.FormatFloat(price, 'f', 20, 64))
	fmt.Println(strconv.FormatFloat(price, 'f', 2, 64))

	a, b := 0.1, 0.2
	fmt.Println(a + b)

	// Output:
	// 0.14499999999999999001
	// 0.14
	// 0.30000000000000004
}
```

Three problems follow from this:

- **Rounding goes the wrong way.** Someone typed `0.145`, and commercial rounding gives `0.15`. Rounding the binary value gives `0.14`.
- **Sums drift.** Each addition can add binary noise, and the noise shows once a total has to match a bank statement.
- **Splits lose cents.** Divide `100.00` into three parts, round each part, and the parts add up to `99.99`.

The data also arrives in awkward shapes. Invoices carry 0, 2 or 4 decimals, JSON encoders and PostgreSQL `float8` output write `1.5e2`, and float columns can hold `NaN`. A single type can't be exact, wide, cheap and tolerant of all that at once. So the package has three types and one default rounding rule.

## Three amount types

```
                  Amount              DecimalAmount             CentAmount
  Go type         float64             struct{ int64 }           int64
  exact           no (binary)         yes                       yes
  decimals        any (inexact)       0..18, kept per value     always 2
  range (units)   about ±1.8e308      about ±2.88e15 at scale 2 about ±9.2e16
  NaN, ±Inf       yes                 yes                       no
  rounding        half away from zero explicit RoundingMode     half away from zero,
                  (shortest decimal)                            or explicit on parse
  SQL Value       float64 (kind)      "1234.50" (string)        int64 (kind, no Valuer)
  JSON            1234.5              1234.50                   123450
```

### Amount: the float64 type

`Amount` is the original type of the package, a `float64` with money methods. It parses every spelling that [`float.ParseDetails`](../float/README.md) accepts, keeps `NaN` and `±Inf`, and works directly with `Rate`, which is a `float64` too.

Use it where the data is a float already: `float8` columns, JSON APIs that send floats, and code that does `Rate` math. Its limits are the limits of `float64`. It is inexact, so sums can drift, and `json.Marshal` fails for `NaN` with `json: unsupported value: NaN`.

### DecimalAmount: exact fixed point

`DecimalAmount` stores an exact decimal value in a single `int64`:

```
 63                                       5 4       0
+------------------------------------------+---------+
|  coefficient, 59 bit two's complement    |  scale  |
+------------------------------------------+---------+
  value = coefficient × 10^-scale
  scale 0..18 (MaxDecimalAmountScale), scale 31 encodes NaN, +Inf and -Inf
```

These choices follow from the layout:

- **Exact values.** `0.10` is ten hundredths, not a binary fraction near it.
- **The scale is kept.** `1.50` (scale 2) and `1.5` (scale 1) are the same value with different representations. `Equal` and `Cmp` compare values, while `==` compares the packed bits, so `NewDecimalAmount(150, 2) == NewDecimalAmount(15, 1)` is false.
- **Exact arithmetic.** `Add` and `Sub` keep the larger scale, and `Mul` adds the scales. `Mul` and `Div` compute in 128 bit integers without heap allocation. They round only when the exact result needs more than 18 decimal places or more digits than the coefficient holds, and then with the `RoundingMode` you pass.
- **No panics on data.** Overflow gives `±Inf`, `x/0` gives `±Inf` and `0/0` gives `NaN`, and `NaN` propagates like in IEEE floating point. Only misuse panics: a scale or decimals argument out of range, or a coefficient out of range passed to `NewDecimalAmount`.
- **Lossless storage.** `Value` writes the exact decimal string (`"1234.50"`), so a `numeric` column keeps every digit. `Scan` reads `numeric` text, `int64` and `float64`. JSON writes a number that keeps the scale (`1234.50`) and writes the non-finite states as `"NaN"`, `"Inf"` and `"-Inf"`.

The cost is range. The coefficient holds ±(2^58-1), about 2.88×10^17, which is about 2.88×10^15 currency units at scale 2.

### CentAmount: exact integer cents

`CentAmount` is an `int64` that counts cents. It is the simplest exact type, and these properties follow from it:

- **The full `int64` range.** It reaches ±92,233,720,368,547,758.07 units, 32 times the `DecimalAmount` coefficient range.
- **No `NaN` or infinity.** Every value from `MinCentAmount` to `MaxCentAmount` is a valid amount.
- **Negation never overflows.** `MinCentAmount` is `math.MinInt64 + 1`, so `-c` is defined for every valid amount.
- **Conversions must decide what to do with non-finite input.** `Amount.CentAmount(mode)` maps `NaN` to 0 and clamps `±Inf`. `DecimalAmount.CentAmount(mode)` returns an error, and `ParseCentAmount("NaN", mode)` returns `can't parse "NaN" as money.CentAmount which has no non-finite values`.
- **Exact splits.** `SplitEqually` and `SplitProportionally` use the largest remainder method, see [Splits](#splits).
- **No `driver.Valuer`.** `database/sql` stores it through its `int64` kind, so a `bigint` column holds the cent count. JSON is the cent count too (`123450`), and a JSON string or a fractional number is rejected.

The cost is the fixed two decimals. `CentAmount` can't represent the fils of the Kuwaiti dinar (3 decimals), and for the yen (0 decimals) its cents have no meaning.

### Side by side

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/money"
)

func main() {
	// 0.10 + 0.20 in each type
	fmt.Println(float64(money.Amount(0.1) + money.Amount(0.2)))
	fmt.Println(money.NewDecimalAmount(10, 2).Add(money.NewDecimalAmount(20, 2)))
	fmt.Println(money.CentAmount(10) + money.CentAmount(20))

	// Division by zero
	fmt.Println(money.NewDecimalAmount(1, 0).Div(money.NewDecimalAmount(0, 0), money.RoundHalfAwayFromZero))

	// Scale is kept, == compares the representation
	a, b := money.NewDecimalAmount(150, 2), money.NewDecimalAmount(15, 1)
	fmt.Println(a, b, a == b, a.Equal(b))

	// Output:
	// 0.30000000000000004
	// 0.30
	// 0.30
	// Inf
	// 1.50 1.5 false true
}
```

## The rounding convention

**Round the shortest decimal representation, half away from zero.**

Every `float64` has a shortest decimal string that parses back to exactly that float: `strconv.FormatFloat(f, 'f', -1, 64)`. For a value that came from a decimal literal with up to 15 significant digits, like `0.145` typed in code, read from an invoice or decoded from JSON, that string is the literal itself. The module rounds those digits, not the binary value behind them. So `Amount(0.145)` rounds to `0.15`, the result a person expects, and not to the `0.14` of its binary value.

| Value    | `strconv.FormatFloat(f, 'f', 2, 64)` | `Amount(f).RoundToCents()` |
|----------|--------------------------------------|----------------------------|
| `0.145`  | `0.14`                               | `0.15`                     |
| `2.675`  | `2.67`                               | `2.68`                     |
| `1.005`  | `1.00`                               | `1.01`                     |
| `0.125`  | `0.12`                               | `0.13`                     |
| `-0.145` | `-0.14`                              | `-0.15`                    |

Half away from zero is the zero value of `RoundingMode`, `RoundHalfAwayFromZero`, the usual default for commercial rounding.

Where the rule applies without a rounding argument:

- `Amount`: `Cents`, `RoundToCents`, `RoundToDecimals`, `String`, `FormatSep`, and the `Amount` splits.
- `Rate`: `RoundToDecimals` and `FormatSep`.
- `CurrencyAmount`: `String` and `FormatSep`.
- `float`: `RoundToDecimals` and `Format`, which back all of the above.
- `CentAmount`: `ScanString`, `MultipliedByRate`, `DividedByRate` and `Percentage`.
- `DecimalAmount`: `MultipliedByRate`, `DividedByRate` and `Percentage`, for a result with more than 18 decimals like a quotient that doesn't terminate.
- `strfmt`: `MoneyFormat.FormatDecimalAmount`, `FormatCentAmount` and `FormatCurrencyDecimalAmount`.

`Amount.RoundToInt` uses `math.Round`, which gives the same result because a tie like `2.5` is exact in `float64`.

Where you choose the rounding:

- every other `DecimalAmount` method that may round: `RoundToDecimals`, `RoundToCents`, `RoundToInt`, `Mul`, `Div` and `CentAmount`,
- `Amount.CentAmount(mode)` and `Amount.DecimalAmount(scale, mode)`, which also round the shortest decimal,
- `ParseCentAmount(str, mode, ...)` for input with more than two decimals.

The eight modes differ only on ties and on the direction of truncation. For `2.5` and `-2.5` rounded to an integer:

| Mode                    | `2.5` | `-2.5` |
|-------------------------|-------|--------|
| `RoundHalfAwayFromZero` | 3     | -3     |
| `RoundHalfToEven`       | 2     | -2     |
| `RoundHalfUp`           | 3     | -2     |
| `RoundHalfDown`         | 2     | -3     |
| `RoundDown`             | 2     | -2     |
| `RoundUp`               | 3     | -3     |
| `RoundFloor`            | 2     | -3     |
| `RoundCeil`             | 3     | -2     |

### What the convention can't fix

The rule recovers the decimal a value was written as. It can't recover the decimal a float calculation was meant to produce. `3 × 0.145` is `0.43499999999999994` in `float64`, and that is also its shortest decimal, so the `float64` `Amount` rounds it down. `CentAmount` and `DecimalAmount` take the shortest decimal of the rate, `0.145`, and multiply exactly:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/money"
)

func main() {
	// Rate math on a float64 Amount lands just below the tie
	fmt.Println(money.Amount(3).MultipliedByRate(0.145).CentAmount(money.RoundHalfAwayFromZero))

	// CentAmount and DecimalAmount compute with the rate's decimal exactly
	fmt.Println(money.CentAmount(300).MultipliedByRate(0.145))
	fmt.Println(money.NewDecimalAmount(300, 2).MultipliedByRate(0.145))

	// Output:
	// 0.43
	// 0.44
	// 0.435
}
```

On `Amount`, the methods that take a `Rate` or a percentage, `MultipliedByRate`, `DividedByRate` and `Percentage`, compute in `float64`. On `CentAmount` and `DecimalAmount` they convert the rate to its shortest decimal and compute exactly; `CentAmount` rounds the result once, half away from zero. `DecimalAmount` falls back to a `float64` product only for a rate with more than 18 decimals or beyond the coefficient range. When a percentage or exchange rate must give an exact cent result, keep the amount in a `CentAmount` or `DecimalAmount`, or put the factor in a `DecimalAmount` and use `Mul` or `Div`.

## Calculate exactly, round once

Rounding is the one step that loses information. Round every intermediate result, and the lost fractions add up to whole cents. `DecimalAmount` arithmetic keeps the full precision of each result, so you can round once, at the end, with the mode your accounting rules require:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/money"
)

func main() {
	// Three invoice lines of 0.3333 each, for example a unit price of
	// 0.03333 times 10 pieces
	line := money.NewDecimalAmount(3333, 4)

	roundedOnce := line.MulInt(3).RoundToCents(money.RoundHalfAwayFromZero)
	roundedPerLine := line.RoundToCents(money.RoundHalfAwayFromZero).MulInt(3)
	fmt.Println(roundedOnce, roundedPerLine)

	// 20% VAT on 19.99 keeps all four decimals until the final rounding
	tax := money.NewDecimalAmount(1999, 2).Mul(money.NewDecimalAmount(20, 2), money.RoundHalfAwayFromZero)
	fmt.Println(tax, tax.RoundToCents(money.RoundHalfToEven))

	// Output:
	// 1.00 0.99
	// 3.9980 4.00
}
```

The rounding mode passed to `Mul` and `Div` only matters when the exact result doesn't fit, like the quotient of a division that doesn't terminate. It is not a hidden rounding to cents.

## Splits

Splitting a total into parts, like installments or cost shares, has to meet three conditions:

1. The parts add up to the total exactly.
2. Each part is within one cent of its exact share.
3. No part has the opposite sign of the total.

Rounding each share on its own breaks condition 1. Giving the whole remainder to the last part meets condition 1 but can break the others. Split 1.00 into 200 parts that way, and 199 parts of 0.01 leave -0.99 for the last one.

`CentAmount.SplitEqually` and `CentAmount.SplitProportionally` use the largest remainder method instead. Each share is truncated to whole cents, which leaves fewer than one cent per part undistributed. Those leftover cents then go one at a time to the parts with the largest truncated fractions. For 100.00 split by the weights 11, 17 and 37:

```
  part  weight  exact share (cents)  truncated  fraction  +1 cent  result
  1     11      1692.307...          1692       0.307              16.92
  2     17      2615.384...          2615       0.384     yes      26.16
  3     37      5692.307...          5692       0.307              56.92
                                     ----                          -----
                                     9999 (1 cent left)            100.00
```

`SplitEqually` has equal fractions everywhere, so the extra cents go to the last parts. `SplitProportionally` uses only the magnitudes of the weights, gives every part the sign of the total, and puts the whole amount in the last part if all weights are zero. It computes the shares in 128 bit integers, so `amount × weight` can't overflow.

The `Amount` splits round the total to cents and then use the same `CentAmount` splits, so they meet the same three conditions. For `SplitProportionally` the one cent bound holds up to the `float64` precision of the weight ratios. Both split without rounding for `NaN`, `±Inf` and amounts beyond the `CentAmount` range. `DecimalAmount` has no split methods: convert with `CentAmount(mode)` first.

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/money"
)

func main() {
	fmt.Println(money.CentAmount(10000).SplitEqually(3))
	fmt.Println(money.CentAmount(-10000).SplitEqually(3))
	fmt.Println(money.CentAmount(10000).SplitProportionally([]money.CentAmount{11, 17, 37}))

	parts := money.Amount(1).SplitEqually(200)
	fmt.Println(parts[0], parts[99], parts[100], parts[199])

	// Output:
	// [33.33 33.33 33.34]
	// [-33.33 -33.33 -33.34]
	// [16.92 26.16 56.92]
	// 0.00 0.00 0.01 0.01
}
```

## Currency minor units

Not every currency has cents. ISO 4217 defines the number of decimals of each currency's minor unit, and `Currency.MinorUnits()` returns it:

| Minor units | Currencies                                         |
|-------------|----------------------------------------------------|
| 0           | BIF, CLP, DJF, GNF, ISK, JPY, KMF, KRW, PYG, RWF, UGX, VND, VUV, XAF, XOF, XPF |
| 3           | BHD, IQD, JOD, KWD, LYD, OMR, TND                  |
| 2           | every other code, including codes without an ISO minor unit like XDR, non-ISO codes like GGP, and invalid codes |

`MinorUnits` normalizes the currency first, so `Currency("jpy").MinorUnits()` is 0 like for `"JPY"`.

Who uses it:

- `CurrencyAmount.String` and `CurrencyAmount.Value` write two decimals, or the three of a currency like KWD, and never fewer: `CurrencyAmountJPY(1234.5)` is `JPY 1234.50`, because a JPY amount can still have a fraction, and a KWD amount of 1.2345 is `KWD 1.235`. `Value` writes the currency normalized and returns an error for an invalid one, `String` shows it as stored. The default `CurrencyAmountParser` normalizes the same way.
- `CurrencyAmount.ScanString` accepts 0 or 2 decimals, or the minor units of the currency, so `"KWD 1.234"` and `"KWD 1,234"` scan as 1.234, while `"EUR 1.234"` is rejected as an ambiguous thousands grouping.
- `CurrencyDecimalAmount.String` prints its own scale and doesn't need the table.
- The `strfmt` money formats raise a precision of 2 or more to the minor units of the currency of a `CurrencyAmount` or `CurrencyDecimalAmount`: the German preset writes `KWD 1.234,567`, and, with its 2 decimals, `JPY 1.234,00`. A precision of 0 or 1 rounds below the cents on purpose and is kept.
- `CentAmount` always uses two decimals.

For a currency with 0 or 3 decimals, store a `DecimalAmount` and round it with `RoundToDecimals(currency.MinorUnits(), mode)`.

## Choosing a type

| Use case                                           | Type                         | SQL column                    | JSON                                  |
|----------------------------------------------------|------------------------------|-------------------------------|---------------------------------------|
| Ledger and invoice totals with 2 decimals, exact sums and splits | `CentAmount`                 | `bigint`                      | `123450`                              |
| Prices with more decimals, VAT and exchange calculations, 0 or 3 decimal currencies | `DecimalAmount`              | `numeric`                     | `1234.50`                             |
| Existing float columns, float JSON APIs, `Rate` math | `Amount`                     | `double precision`            | `1234.5`                              |
| Optional amount of any of the above                | `NullableDecimalAmount` etc. | the same type, `NULL` allowed | `null` for null                       |
| Currency and amount in one text column             | `CurrencyDecimalAmount`      | `text`, like `EUR 1234.50`    | `{"Currency":"EUR","Amount":1234.50}` |

`CurrencyAmount` is the `float64` counterpart of `CurrencyDecimalAmount`. It writes text like `EUR 1234.50` to SQL too, with two decimals or the three of a currency like KWD, and `{"Currency":"EUR","Amount":1234.5}` to JSON.

If you are not sure, `DecimalAmount` is the safe default. It is exact, keeps whatever scale the input has, and converts to `CentAmount` for a split with one call that reports an error instead of clamping.

### Alternatives

- **`float64` everywhere, rounded at the edges.** That is `Amount`. It works for display and interop, and fails for sums that must match to the cent.
- **`math/big`.** A `*big.Float` or `*big.Int` is a pointer to a heap allocated value: `==` compares addresses, and every result allocates. `DecimalAmount` is a single `int64`, so it is a plain comparable value, and its 128 bit `Mul` and `Div` don't allocate.
- **Plain `int64` cents.** That is `CentAmount`, with parsing, formatting, splits and the `MinCentAmount` guard added.

## Related

- [`money`](../money/README.md): the API of all amount, currency and rate types.
- [`float`](../float/README.md): `float.ParseDetails`, `float.Format` and `float.RoundToDecimals`, which implement the parsing and the rounding rule.
- [`strfmt`](../strfmt/README.md): the `MoneyFormat` presets.
- [Lenient parsing](explanation-lenient-parsing.md): how `"1.234,50"` and `"1,234.50"` both become 1234.50.
- [Null conventions](explanation-null-conventions.md): the `Nullable` amount types.
- [Tutorial: build an invoice record](tutorial-invoice-record.md): `DecimalAmount` and `CentAmount.SplitEqually` in context.
