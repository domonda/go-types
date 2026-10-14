# Lenient parsing

This page explains why the parsers of this module accept many spellings of the same value, which rules decide between readings, and where that leniency stops. The rules apply to [`float`](../float/README.md), [`money`](../money/README.md), [`date`](../date/README.md), [`bank`](../bank/README.md), [`vat`](../vat/README.md), [`country`](../country/README.md) and [`language`](../language/README.md).

## The problem

The values this module handles come from text that people typed or machines extracted: invoice PDFs, bank CSV exports, OCR output, web forms. The same value shows up in many spellings:

| Value                | Spellings in invoices, exports and forms           |
|----------------------|----------------------------------------------------|
| 1234.50              | `1.234,50`, `1,234.50`, `1 234,50`, `1'234.50`, `1234,5` |
| 2023-12-25           | `25.12.2023`, `12/25/2023`, `Dec 25, 2023`, `25. Dezember 2023`, `25.12.23` |
| ATU10223006          | `atu 1022 3006`, `AT-U 1022.3006`                  |
| AT611904300234573201 | `AT61 1904 3002 3457 3201`                         |
| EUR                  | `€`, `eur`, `Euro`                                 |
| AT                   | ` at `, `AUT`, `Österreich`                        |

A strict parser that accepts one spelling per value pushes the cleanup into every service that reads such input, and every service gets it slightly different. A parser that takes a locale argument needs to know the locale first, and an OCR text or a bank export usually doesn't declare one.

## The approach

Each type parses leniently in one place and reduces every accepted spelling to one canonical form:

```
  "atu 1022 3006"  --+
  "AT-U 1022.3006" --+--> Normalized() --> "ATU10223006" --> store, compare, index
  "ATU10223006"    --+
```

The methods follow one pattern, which [Validation and normalization](explanation-validation-and-normalization.md) explains in detail:

- `Normalized()` returns the canonical form or an error.
- `Valid()` and `Validate()` accept every spelling that normalizes.
- `Scan`, JSON and XML decoding and `ScanString` store a valid value in the canonical form and an invalid one unchanged.
- `Value()` writes the canonical form and refuses an invalid value, `MarshalJSON()` and `MarshalText()` write the canonical form of a valid value.
- `ValidAndNormalized()` accepts only the canonical form.
- `fmt` and `String()` show the value as stored.

So the data is normalized where it enters your system from SQL, JSON or XML. Normalize values that you build in Go from other strings, like `money.Currency(s)`, yourself.

When two readings are possible, a fixed rule decides, and the caller can override it with context: a list of accepted decimal counts for numbers, a language hint for dates. The rules below are the ones that make that decision.

## Numbers

`float.ParseDetails` reads the separators of a number. `float.Parse`, `money.ParseAmount`, `ParseDecimalAmount`, `ParseCentAmount`, `ParseRate`, `ParseCurrencyAmount`, the `ScanString` methods of the money types and the float scanning of `strfmt` all build on it. The rules:

1. **Two different separators.** The last one is the decimal separator, the other one groups thousands: `1.234,56` and `1,234.56` are both 1234.56.
2. **A single `.` or `,` is the decimal separator.** `12,5` is 12.5, and `1,234` is 1.234.
3. **A repeated `.` or `,` groups thousands.** `1.234.567` is 1234567.
4. **`'` and space only ever group thousands.** `1'234` and `1 234` are 1234. One that can't group thousands, like in `1'5`, is an error.
5. **Groups have fixed sizes.** Every thousands separator is followed by exactly 3 digits, and the group in front of the first one has 1 to 3 digits. `1234.567,89` and `1.23.4` are errors.
6. **The sign can be leading or trailing, but not both.** `1.234,56-` is -1234.56, as some bank exports write it. `+1.5-` is an error.
7. **Scientific notation is accepted.** `1.5e3` is 1500, because JSON encoders and PostgreSQL `float8` text output use it. A separator inside the exponent is an error.
8. **Non-finite values are exact tokens.** `NaN`, `Inf`, `+Inf` and `-Inf` parse to the non-finite values, and `ParseDecimalAmount` also reads `Infinity`. `infinity` in lower case is an error, and `ParseCentAmount` rejects all of them because `CentAmount` has no non-finite states. `ParseAmount`, `ParseRate` and `ParseDecimalAmount` reject them too when `acceptedDecimals` are passed.

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/float"
)

func main() {
	for _, s := range []string{"1.234,56", "1,234.56", "12,5", "1,234", "1.234.567", "1'234", "1 234,56", "1.234,56-", "1.5e3", "1'5", "1234.567,89"} {
		f, err := float.Parse(s)
		if err != nil {
			fmt.Printf("%-12s error: %v\n", s, err)
			continue
		}
		fmt.Printf("%-12s %v\n", s, f)
	}

	// Output:
	// 1.234,56     1234.56
	// 1,234.56     1234.56
	// 12,5         12.5
	// 1,234        1.234
	// 1.234.567    1.234567e+06
	// 1'234        1234
	// 1 234,56     1234.56
	// 1.234,56-    -1234.56
	// 1.5e3        1500
	// 1'5          error: thousands separators have to be 3 characters apart: "1'5"
	// 1234.567,89  error: first group before a thousands separator must have 1 to 3 digits: "1234.567,89"
}
```

### The "1,000" ambiguity

`1,000` is one in a German document and one thousand in an English one, and `1.000` is the other way round. Nothing in the string tells the readings apart. Rule 2 picks the decimal reading, which is wrong for every amount of a thousand or more that has a thousands separator and no decimals.

Money amounts usually have 0 or 2 decimals, so a reading with 3 decimals is a sign of this ambiguity. Pass the decimal counts you expect as `acceptedDecimals`, and the parsers turn the ambiguous case into an error instead of a wrong value:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/money"
)

func main() {
	for _, s := range []string{"1,000", "1.000", "1,000.00", "1.000,00", "1000"} {
		a, err := money.ParseDecimalAmount(s, 0, 2)
		fmt.Println(a, err)
	}

	// Output:
	// 0 parsing "1,000" returned 3 decimals which is not in the accepted list [0 2]
	// 0 parsing "1.000" returned 3 decimals which is not in the accepted list [0 2]
	// 1000.00 <nil>
	// 1000.00 <nil>
	// 1000 <nil>
}
```

The error leaves the decision to code that has more context, like the language of the document or the format of the other amounts in it. `ParseAmount` and `ParseRate` report the same case as `parsing "1,000" returned 3 decimals which is not in accepted list of [0 2]`.

`ParseDecimalAmount` and `ParseCentAmount` match `acceptedDecimals` against the decimals of the result, after an exponent is applied, while `ParseAmount` matches the decimals written in the mantissa. For `1.5e2`, `ParseDecimalAmount` accepts 0 and `ParseAmount` accepts 1.

`strfmt.ScanConfig.AcceptedMoneyAmountDecimals` holds such a list, but `strfmt.Scan` doesn't apply it: a `money.Amount` cell is parsed by `Amount.ScanString`, which accepts any decimal count. Pass the list to `money.ParseAmount` or `money.NewAmountParser` yourself.

## Dates

`date.Normalize(str, lang...)`, `Date.Normalized()` and `Date.NormalizedLang(lang)` split the input into three parts at spaces, `.`, `/`, `-`, `,`, `'` and the CJK date characters, and decide which part is the day, the month and the year:

- **Text around the date.** A leading weekday (`Fri, 03 Jul 2015`), the filler `of` (`1st of dec. 2020`), a trailing `00:00:00` and a `T` time after an ISO date are dropped.
- **Month names.** English and German names and abbreviations are accepted in any case, including common OCR misreadings like `1an` and `apri1`. French, Italian and Spanish month names and their usual abbreviations are accepted too, like `févr.`, `luglio` and `septiembre`.
- **ISO order.** A four digit first part means year, month, day: `2023-12-25`.
- **Day first by default.** `05/06/2024` is 2024-06-05.
- **Month first when nothing else works.** `12/25/2023` is 2023-12-25, because 25 can't be a month.
- **Two-digit years.** 00 to 44 become 2000 to 2044, 45 to 99 become 1945 to 1999.

Not supported: year-only and year-month strings (`2014`, `2014-04`), compact digits (`20140601`), dates whose three parts have fewer than 5 characters together (`8/8/71`) and a time with a zone offset inside the string. Use `time.Parse` for those.

### Ambiguous dates and the language hint

A numeric date with the year last is ambiguous when both numbers can be the month:

```
  first  second   reading
  1-12   13-31    month first, monthMustBeFirst = true    12/25/2023 -> 2023-12-25
  13-31  1-12     day first                               25.12.2023 -> 2023-12-25
  1-12   1-12     ambiguous: day first,                   05/06/2024 -> 2024-06-05
                  month first with an English hint        05/06/2024 -> 2024-05-06 (EN)
                  or an ordinal suffix on the second      02 1st 2003 -> 2003-02-01
```

The hint is the first `lang` argument, normalized with `language.Code.Normalized`, so `language.EN`, `"en-US"` and `"eng"` all mean English, a hint that can't be normalized is ignored, and only English changes the result. An ordinal suffix decides before the hint, so `1st 02 2003` is 2003-02-01 also with `language.EN`. The hint and an ordinal suffix only decide ambiguous dates, but a number with an ordinal suffix is never read as the month, so `13 1st 2024` is an error and not 2024-01-13. A date with one valid reading, like `15.06.2024`, gives 2024-06-15 with any hint, and the separator never matters.

### Detecting the date order of a whole document

One ambiguous date can't tell you the order, but a document usually writes all its dates the same way. `Date.ScanStringWithLang` returns `monthMustBeFirst`, which is true when a date can only be read month first. Find one such date, and you know how to read the ambiguous ones:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/language"
)

func main() {
	found := []string{"05/06/2024", "12/25/2023", "07/08/2024"}

	var hint language.Code
	for _, s := range found {
		var d date.Date
		_, monthMustBeFirst, err := d.ScanStringWithLang(s, "")
		if err == nil && monthMustBeFirst {
			hint = language.EN
			break
		}
	}

	for _, s := range found {
		d, err := date.Normalize(s, hint)
		fmt.Println(s, d, err)
	}

	// Output:
	// 05/06/2024 2024-05-06 <nil>
	// 12/25/2023 2023-12-25 <nil>
	// 07/08/2024 2024-07-08 <nil>
}
```

Without `12/25/2023` in the list, the other two dates stay day first, 2024-06-05 and 2024-08-07. A list without any date that only reads month first can't tell the two orders apart, so the default applies.

## Identifiers

Identifiers have check digits and fixed formats, so their leniency is about whitespace, punctuation and case, not about choosing between readings. Each type removes a different set of characters:

| Type             | Removes or folds                                   | Then checks                                        |
|------------------|----------------------------------------------------|----------------------------------------------------|
| `bank.IBAN`      | all Unicode whitespace, also no-break and zero width spaces. No case folding, no punctuation removal | country code, length of that country, characters, mod-97 check sum |
| `bank.BIC`       | all Unicode whitespace, appends `XXX` to 8 characters. No case folding | syntax, country code at characters 5 and 6, deny-list of German words |
| `vat.ID`         | whitespace and Unicode punctuation, upper-cases    | length, country prefix, country format, check digit for most countries |
| `country.Code`   | surrounding whitespace, upper-cases                | ISO 3166-1 table, then `AltCodes` like `SUI`, `D` and German names like `Österreich` |
| `money.Currency` | surrounding whitespace, upper-cases                | symbols like `€`, codes, English and ISO names like `Canadian Dollar`, then guesses for an unqualified `Dollar`, `Franc` or `Pound` |
| `language.Code`  | surrounding whitespace, case, and the script, region and other subtags of a BCP-47 tag | ISO 639-1, 639-2 and 639-3 codes, so `EN`, `eng` and `en-US` are `en` |

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/bank"
	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/vat"
)

func main() {
	fmt.Println(vat.NormalizeVATID("AT-U 1022.3006"))
	fmt.Println(bank.NormalizeIBAN("AT61 1904 3002 3457 3201"))
	fmt.Println(bank.NormalizeIBAN("at61 1904 3002 3457 3201"))
	fmt.Println(bank.NormalizeBIC("GIBA AT WW"))
	fmt.Println(country.Code("Österreich").Normalized())
	fmt.Println(money.Currency("€").Normalized())
	fmt.Println(money.Currency("Atlantean Dollar").Normalized())

	// Output:
	// ATU10223006 <nil>
	// AT611904300234573201 <nil>
	// at61 1904 3002 3457 3201 invalid IBAN country code
	// GIBAATWWXXX <nil>
	// AT <nil>
	// EUR <nil>
	// Atlantean Dollar invalid currency code "ATLANTEAN DOLLAR"
}
```

A guess stops where it could pick the wrong currency. An unqualified `Dollar` becomes USD, but a dollar qualified with an unknown country is an error, because the dollar of another country is a different currency.

English country names are not part of `country.Code.Normalized`. `country.ParseCode("Germany")` reads them, and so does `language.ParseCode("German")` for language names.

## Formatting and parsing don't always round-trip

`Normalized` and `String` write the canonical form, and the canonical form always parses back to itself. A localized form can lose the information the parser needs:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/float"
)

func main() {
	// A thousands separator without decimals reads back as a decimal separator
	s := float.Format(1234.0, ',', '.', -1, false)
	f, _ := float.Parse(s)
	fmt.Println(s, f)

	// With decimals the separators are unambiguous
	s = float.Format(1234.0, ',', '.', 2, true)
	f, _ = float.Parse(s)
	fmt.Println(s, f)

	// A month first layout reads back day first without a hint
	s = date.Date("2024-05-06").Format("01/02/2006")
	fmt.Println(s, date.Must(s))

	// Output:
	// 1,234 1.234
	// 1,234.00 1234
	// 05/06/2024 2024-06-05
}
```

The `strfmt` presets avoid both traps. Their money formats always write two decimals, or the three of a currency like KWD, so `1,234.50`, `1.234,50` and `KWD 1.234,000` read back exactly. Their float formats don't group thousands. `NewEnglishFormatConfig` writes dates as `02/01/2006`, day first, which is the default reading. If you define your own formats, write decimals whenever you group thousands, and pass `language.EN` when you read back dates you wrote month first. For data exchanged between programs, use the canonical forms.

## Trade-offs

- **Leniency against ambiguity.** Every extra spelling accepted is a chance that a string has two readings. The parsers resolve them with fixed rules: a single separator is decimal, numeric dates are day first. When the rule is wrong for your input, the caller has to supply the context. Use `acceptedDecimals` for numbers, a language hint for dates, `ScanStringWithLang` to detect the date order of a whole document, and `ParseCode` instead of `Normalized` for country and language names.
- **Leniency against false matches.** The finders that search free text, like `bank.BICFinder` and `money.AmountFinder`, can't accept as much as the parsers, or ordinary words and numbers would match. They require upper-case BICs and amounts with exactly two decimals, and still find words like `RECHNUNG` that happen to have BIC syntax. See [Extract identifiers from text](howto-extract-identifiers-from-text.md).
- **Offline checks only.** A valid VAT ID or IBAN has the right format and check digit. Whether it is registered or belongs to the vendor needs an online check like VIES.
- **Unvalidated input.** Reading never fails because of the text itself: `Scan`, JSON and XML decoding store an invalid value unchanged, so one bad row doesn't fail a whole query. Validate after decoding, see the [tutorial](tutorial-invoice-record.md); `Value` refuses an invalid value when it is written.

## Related

- [`float`](../float/README.md): `ParseDetails` and `Format`.
- [`money`](../money/README.md): the amount parsers and `Currency.Normalized`.
- [`date`](../date/README.md): the full list of accepted date shapes.
- [`bank`](../bank/README.md), [`vat`](../vat/README.md), [`country`](../country/README.md), [`language`](../language/README.md): the identifier rules.
- [`strfmt`](../strfmt/README.md): scanning and formatting table cells with these parsers.
- [Validation and normalization](explanation-validation-and-normalization.md): the `Valid`, `Validate` and `Normalized` methods.
- [Money types and rounding](explanation-money-types-and-rounding.md): what happens to an amount after parsing.
- [Scan and format table rows](howto-scan-and-format-table-rows.md): a CSV round trip with the `strfmt` presets.
