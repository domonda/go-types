# go-types documentation

The documentation follows the [Diataxis](https://diataxis.fr/) structure:
pick the section that matches what you need right now.

## Tutorial: learn by building

| Document                                          | You will                                           |
|---------------------------------------------------|----------------------------------------------------|
| [Build an invoice record from raw strings](tutorial-invoice-record.md) | Turn extracted invoice text into validated `vat.ID`, `bank.IBAN`, `date.Date` and money values, marshal them to JSON and SQL, and split the total into installments |

## How-to guides: get a task done

| Document                                          | Task                                               |
|---------------------------------------------------|----------------------------------------------------|
| [How to validate structs](howto-validate-structs.md) | Make your own types validatable and check nested structs with `types.DeepValidate` |
| [How to store optional values](howto-store-optional-values.md) | Pick the Go type for a nullable SQL column or optional JSON field |
| [How to extract identifiers, dates and amounts from text](howto-extract-identifiers-from-text.md) | Find IBANs, BICs, VAT IDs, dates and amounts in OCR or PDF text |
| [How to scan and format table rows](howto-scan-and-format-table-rows.md) | Read CSV cells into typed fields and write localized CSV with `strfmt` |
| [How to add a country, language or currency code](howto-add-an-enum-code.md) | Extend the code tables and regenerate the enum methods |
| [How to use deterministic IDs in tests](howto-deterministic-ids-in-tests.md) | Inject predictable UUIDs through the context with `uu.NewID` |

## Explanations: understand the design

| Document                                          | Question it answers                                |
|---------------------------------------------------|----------------------------------------------------|
| [Null conventions: X and NullableX](explanation-null-conventions.md) | Why most types come in pairs, and what each writes to SQL and JSON |
| [Validation and normalization](explanation-validation-and-normalization.md) | How `Valid`, `Validate`, `Normalized` and `DeepValidate` relate |
| [Money types and rounding](explanation-money-types-and-rounding.md) | When to use `Amount`, `DecimalAmount` or `CentAmount`, and how rounding works |
| [Lenient parsing](explanation-lenient-parsing.md) | Which spellings of numbers, dates and identifiers parse, and how ambiguity is resolved |
| [Sets](explanation-sets.md)                       | Why there are several set types and which to use   |

## Reference: look something up

Each package README is the complete reference for its package, with every
exported identifier, the SQL and JSON serialization, gotchas and runnable
examples.

| Package                                           | Reference                                          |
|---------------------------------------------------|----------------------------------------------------|
| `types` (root)                                    | [reference-types.md](reference-types.md)           |
| `account`                                         | [account/README.md](../account/README.md)          |
| `bank`                                            | [bank/README.md](../bank/README.md)                |
| `charset`                                         | [charset/README.md](../charset/README.md)          |
| `charset/findencoding`                            | [charset/findencoding/README.md](../charset/findencoding/README.md) |
| `country`                                         | [country/README.md](../country/README.md)          |
| `date`                                            | [date/README.md](../date/README.md)                |
| `email`                                           | [email/README.md](../email/README.md)              |
| `float`                                           | [float/README.md](../float/README.md)              |
| `language`                                        | [language/README.md](../language/README.md)        |
| `language/gen-lang-maps`                          | [language/gen-lang-maps/README.md](../language/gen-lang-maps/README.md) |
| `mapset`                                          | [mapset/README.md](../mapset/README.md)            |
| `money`                                           | [money/README.md](../money/README.md)              |
| `notnull`                                         | [notnull/README.md](../notnull/README.md)          |
| `nullable`                                        | [nullable/README.md](../nullable/README.md)        |
| `queue`                                           | [queue/README.md](../queue/README.md)              |
| `set`                                             | [set/README.md](../set/README.md)                  |
| `strfmt`                                          | [strfmt/README.md](../strfmt/README.md)            |
| `strutil`                                         | [strutil/README.md](../strutil/README.md)          |
| `strutil/stringfrom`                              | [strutil/stringfrom/README.md](../strutil/stringfrom/README.md) |
| `uu`                                              | [uu/README.md](../uu/README.md)                    |
| `vat`                                             | [vat/README.md](../vat/README.md)                  |

The API documentation generated from the Go doc comments is on
[pkg.go.dev](https://pkg.go.dev/github.com/domonda/go-types).

Changes, including breaking ones, are listed in the
[CHANGELOG](../CHANGELOG.md). Open questions before `v1.0.0` are in
[TODOS](../TODOS.md).
