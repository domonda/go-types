# go-types

[![Go Reference](https://pkg.go.dev/badge/github.com/domonda/go-types.svg)](https://pkg.go.dev/github.com/domonda/go-types) [![Go Report Card](https://goreportcard.com/badge/github.com/domonda/go-types)](https://goreportcard.com/report/github.com/domonda/go-types) [![Go](https://github.com/domonda/go-types/actions/workflows/go.yml/badge.svg)](https://github.com/domonda/go-types/actions/workflows/go.yml) [![Go version](https://img.shields.io/github/go-mod/go-version/domonda/go-types)](https://github.com/domonda/go-types) [![license](https://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://github.com/domonda/go-types/blob/master/LICENSE)

Go types for the data that business documents are made of: money amounts and
currencies, dates, IBANs and BICs, VAT IDs, country and language codes, email
addresses and messages, and UUIDs. Each type parses the many ways people and
other systems write a value, normalizes it to one canonical form, validates
it, and round-trips through `database/sql` and `encoding/json`. Optional
values have `Nullable` variants that map to SQL `NULL` and JSON `null`.

```
go get github.com/domonda/go-types
```

Requires Go 1.26 or later.

## Quick start

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/bank"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/language"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/vat"
)

func main() {
	// Normalize identifiers as they appear on invoices
	vatID, err := vat.ID("atu 1022 3006").Normalized()
	fmt.Println(vatID, err)

	iban, err := bank.IBAN("DE89 3704 0044 0532 0130 00").Normalized()
	fmt.Println(iban, err)

	// Dates are day-first unless a language hint says otherwise
	d, err := date.Normalize("05.06.2024")
	fmt.Println(d, err)
	d, err = date.Normalize("05/06/2024", language.EN)
	fmt.Println(d, err)

	// Exact money amounts with locale-aware parsing
	amount, err := money.ParseDecimalAmount("1.234,50")
	fmt.Println(amount, err)
	currency, err := money.Currency("€").Normalized()
	fmt.Println(currency, err)

	// Integer cents split without losing or inventing a cent
	fmt.Println(money.CentAmount(1000).SplitEqually(3))

	// Output:
	// ATU10223006 <nil>
	// DE89370400440532013000 <nil>
	// 2024-06-05 <nil>
	// 2024-05-06 <nil>
	// 1234.50 <nil>
	// EUR <nil>
	// [3.33 3.33 3.34]
}
```

New here? The [tutorial](docs/tutorial-invoice-record.md) builds an invoice
record from raw strings to validated JSON and SQL values step by step.

## Packages

| Package                                           | What it gives you                                  |
|---------------------------------------------------|----------------------------------------------------|
| [`types`](docs/reference-types.md) (root)         | `Set[T]`, the `Validator`/`ValidatErr` interfaces with `Validate` and `DeepValidate`, `Normalizable`, `LenString`, `KeyMutex`, `Finder`, pointer and iterator helpers |
| [`account`](account/)                             | Alphanumeric account numbers with SQL, JSON and XML support |
| [`bank`](bank/)                                   | IBAN and BIC validation and normalization, finders for document text, CAMT.053 XML structs |
| [`charset`](charset/)                             | BOM detection, UTF-16 and UTF-32, legacy charsets by IANA or MIME label |
| [`country`](country/)                             | ISO 3166-1 alpha-2 country codes, alternative spellings, EU membership |
| [`date`](date/)                                   | ISO 8601 `Date`, `YearMonth`, `YearQuarter`, lenient locale-aware parsing, period ranges |
| [`email`](email/)                                 | Lenient address parsing, address lists and sets, MIME and TNEF message parsing and building |
| [`float`](float/)                                 | Locale-aware float parsing and formatting, decimal rounding, a JSON-tolerant float |
| [`language`](language/)                           | ISO 639-1 language codes from BCP 47, ISO 639-2/3 and locale identifiers |
| [`mapset`](mapset/)                               | Set algebra on `map[K]struct{}` and `map[K]bool`, mirroring the proposed `container/mapset` |
| [`money`](money/)                                 | `Amount` (float64), `CentAmount` (int64 cents), `DecimalAmount` (exact fixed-point), ISO 4217 currencies, rates |
| [`notnull`](notnull/)                             | Never-null SQL/JSON wrappers: nil slices become `'{}'` and `[]` |
| [`nullable`](nullable/)                           | Nullable wrappers that map the zero value to SQL `NULL` and JSON `null`, generic `Type[T]` |
| [`queue`](queue/)                                 | Unbounded goroutine-safe FIFO queue consumed through a channel |
| [`set`](set/)                                     | Helpers for `map[T]struct{}` sets (prefer `mapset` for new code) |
| [`strfmt`](strfmt/)                               | Reflection-driven scanning and formatting of values with locale presets |
| [`strutil`](strutil/)                             | Trimming, rune filtering, transliteration, file names, domain extraction, `StringSet` |
| [`strutil/stringfrom`](strutil/stringfrom/)       | Render pointers and optional values as strings with a text for nil |
| [`uu`](uu/)                                       | UUID `ID`, `NullableID`, `IDSlice`, `IDSet`, v7 by default, context-injectable generation |
| [`vat`](vat/)                                     | European VAT IDs with per-country format and checksum validation |

Every package README is the complete reference for its package: all exported
types and functions, how each value is written to and read from SQL and JSON,
the gotchas, and runnable examples. The same API is on
[pkg.go.dev](https://pkg.go.dev/github.com/domonda/go-types).

## Documentation

The [docs index](docs/README.md) lists everything. By what you need:

- **Learn the library:** [Tutorial: an invoice record](docs/tutorial-invoice-record.md)
- **Get a task done:**
  - [How to validate structs](docs/howto-validate-structs.md)
  - [How to store optional values in SQL and JSON](docs/howto-store-optional-values.md)
  - [How to extract IBANs, VAT IDs, dates and amounts from text](docs/howto-extract-identifiers-from-text.md)
  - [How to scan and format table rows](docs/howto-scan-and-format-table-rows.md)
  - [How to add a country, language or currency code](docs/howto-add-an-enum-code.md)
  - [How to use deterministic IDs in tests](docs/howto-deterministic-ids-in-tests.md)
- **Understand the design:**
  - [Null conventions: `X` and `NullableX`](docs/explanation-null-conventions.md)
  - [Validation and normalization](docs/explanation-validation-and-normalization.md)
  - [Money types and rounding](docs/explanation-money-types-and-rounding.md)
  - [Lenient parsing of numbers and dates](docs/explanation-lenient-parsing.md)
  - [Sets: `types.Set`, `set` and `mapset`](docs/explanation-sets.md)
- **Look something up:** the package READMEs in the table above.

## Conventions shared by all packages

- Identifier types like `country.Code`, `vat.ID`, `bank.IBAN` and `uu.ID`
  have an invalid zero value, and their `NullableX` variant uses that zero
  value as SQL `NULL` and JSON `null`. Number types like `money.Amount`, where
  zero is a valid value, use the generic `nullable.Type[T]` instead. See
  [null conventions](docs/explanation-null-conventions.md).
- `Valid() bool` and `Validate() error` accept the input that `Normalized()`
  accepts, so `country.Code("de").Valid()` and `vat.ID("atu 1022 3006").Valid()`
  are true. `ValidAndNormalized()` reports whether a value is already in
  canonical form. One type is strict: `language.Code.Valid` only accepts a
  canonical code (its `Validate` normalizes first). `Value()` and
  `MarshalJSON()` write the normalized form, so a non-canonical spelling that
  `Validate` accepts doesn't reach SQL or JSON. See
  [validation and normalization](docs/explanation-validation-and-normalization.md).
- How each type is written to and read from SQL and JSON, including its zero
  and null values, is in the Serialization section of its package README.
- `types.Validate` and `types.DeepValidate` work with any type that implements
  `Validate() error` or `Valid() bool`, including your own.

## Design decisions

These rules decide how every type behaves, and how an inconsistency between
types gets resolved. Types that don't follow a rule yet are tracked in
[TODOS.md](TODOS.md).

- **Be generous in accepting input, be canonical in generating output.**
  Parsing, normalizing, validating and scanning accept every common way people
  and other systems write a value: `country.Code("de")`,
  `vat.ID("atu 1022 3006")` and `money.Currency("€")` are valid input. What a
  type writes, its `Value`, its JSON and its `String`, is the canonical form:
  `country.Code("de")` is stored and marshalled as `DE`. Lenient validation
  and canonical output belong together, a value that passes `Validate` must
  never be stored in another spelling. The canonical form is UTF-8 for
  readability, the email types have an extra `ASCII()` for message headers.
  `String` of a set lists its elements as given, because normalizing could
  merge elements.
- **Empty input is null.** For a nullable type, an empty or whitespace-only
  string from SQL, JSON or text is null, because an empty CSV cell or text
  column means "no value".
- **Fail loudly instead of guessing.** Input that contradicts itself, or whose
  meaning depends on something the package can't know, is an error instead of
  a silently wrong value. A number with an ordinal suffix is never read as the
  month, so `"13 1st 2024"` is an error, and `strfmt` rejects a time zone
  abbreviation whose offset is unknown on the machine instead of reading the
  time as UTC. A documented default is no guess: an ambiguous date like
  `05/06/2024` is day first unless a language hint says otherwise.
- **Errors are created with
  [`github.com/domonda/go-errs`](https://github.com/domonda/go-errs)**
  (`errs.New`, `errs.Errorf`) instead of `errors.New` and `fmt.Errorf`.

## Development

```
go test ./...          # all tests
./run-go-enum.sh       # regenerate Enums()/EnumStrings() after adding a code
./run-gosec.sh         # static security analysis
```

`go-enum` and `gosec` are pinned as tool dependencies in `tools/go.mod`, so the
main module does not depend on them. CI builds, vets, checks that the
generated enum methods are up to date and runs the tests on every pull
request.

## Status

No version has been tagged yet. [CHANGELOG.md](CHANGELOG.md) lists every
change, including the breaking ones, and [TODOS.md](TODOS.md) lists what is
left before `v1.0.0`.

## License

MIT, see [LICENSE](LICENSE). The `uu` package is derived from
[github.com/satori/go.uuid](https://github.com/satori/go.uuid).
