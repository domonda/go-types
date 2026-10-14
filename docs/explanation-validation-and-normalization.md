# Validation and normalization

This page explains how the types of this module decide whether a value is
valid, how they turn accepted input into one canonical spelling, and how
`types.Validate` and `types.DeepValidate` apply both to whole structs. The
steps for your own types are in
[How to validate structs](howto-validate-structs.md).

## The problem

Real input arrives in many spellings. A user types `de`, an OCR engine reads
` DE `, a German form says `Deutschland`, and all three mean `country.DE`. A
bank statement prints an IBAN in groups of four characters. Code that receives
such a value has three separate questions:

1. Can the value be used at all?
2. If not, why not?
3. What is its one canonical spelling, for storage and comparison?

A single method can't answer all three well. A strict check rejects real
input that is only spelled differently, and a lenient check lets
non-canonical text reach the database, where `"de"` and `"DE"` then compare
as different countries.

## Four methods for four questions

| Method                      | Interface                        | Question it answers                       |
|-----------------------------|----------------------------------|-------------------------------------------|
| `Valid() bool`              | `types.Validator`                | Can the value be used?                    |
| `Validate() error`          | `types.ValidatErr`               | If not, why not? `nil` means valid.       |
| `Normalized() (T, error)`   | `types.Normalizable[T]`          | What is the canonical form?               |
| `ValidAndNormalized() bool` | `types.NormalizableValidator[T]` | Is the value valid and already canonical? |

`types.NormalizableValidator[T]` combines `Validator`, `Normalizable[T]` and
`ValidAndNormalized`. These types assert it at compile time: `bank.IBAN`,
`bank.BIC`, `bank.NullableIBAN`, `country.Code`, `date.Date`,
`date.NullableDate`, `email.Address`, `email.AddressList`, `email.AddressSet`,
`language.Code`, `money.Currency` and `vat.ID`. The `Normalized` method of the
`date` types takes no language hint; `NormalizedLang(lang language.Code)` takes
the hint that decides between day-first and month-first dates.

```
   input "atu 1022 3006"
           |
           |   Valid(), Validate(): can Normalized() succeed?
           v
     Normalized() ------------> "ATU10223006"
           |                    ValidAndNormalized() is true for this one
           |
           +--> error: the input comes back unchanged,
                the error message names it
```

## Validity is lenient

`Validate` doesn't ask "is this already canonical?" but "can `Normalized`
turn this into a valid value?". It is implemented by calling `Normalized`
and returning its error, and `Valid` agrees with it for every type but
`language.Code`:

| Value                                      | `Valid()` | `Normalized()`           | `ValidAndNormalized()` |
|--------------------------------------------|-----------|--------------------------|------------------------|
| `country.Code(" at ")`                     | `true`    | `AT`                     | `false`                |
| `country.Code("Österreich")`               | `true`    | `AT`                     | `false`                |
| `money.Currency("€")`                      | `true`    | `EUR`                    | `false`                |
| `money.Currency("Canadian Dollar")`        | `true`    | `CAD`                    | `false`                |
| `vat.ID("atu 1022 3006")`                  | `true`    | `ATU10223006`            | `false`                |
| `bank.IBAN("DE89 3704 0044 0532 0130 00")` | `true`    | `DE89370400440532013000` | `false`                |
| `bank.BIC("BKAUATWW")`                     | `true`    | `BKAUATWWXXX`            | `false`                |
| `bank.BIC("BKAU AT WW")`                   | `true`    | `BKAUATWWXXX`            | `false`                |
| `email.Address("John@Example.com")`        | `true`    | `john@example.com`       | `false`                |
| `date.Date("25.12.2023")`                  | `true`    | `2023-12-25`             | `false`                |
| `country.Code("XX")`                       | `false`   | error                    | `false`                |

The reasoning: a valid value in another spelling is still the same value, so
rejecting it would only reject real-world input. Leniency is safe because
`Value` and `MarshalJSON` write the normalized form, see
[Normalized output](#normalized-output). Where a contract requires the
canonical form in Go, like a map key or a comparison with `==`, check
`ValidAndNormalized` or keep the result of `Normalized`.

Each type decides how far its leniency goes. `bank.IBAN` removes whitespace
but doesn't change the letter case, so `de89370400440532013000` fails with
`invalid IBAN country code`. `country.Code` accepts German names and ITU codes
from `AltCodes` but not English names; `country.ParseCode` adds those.
`bank.BIC` removes whitespace, but `BIC.Parse` and `BICParser` read the BIC
as given, so whitespace makes it invalid there.
The package READMEs and [Lenient parsing](explanation-lenient-parsing.md) list
the details.

## Where the methods disagree

### `language.Code`: strict `Valid`, lenient `Validate`

`language.Code.Valid` only accepts the canonical lower-case ISO 639-1 code,
while `Validate` accepts every spelling that `Normalized` accepts:

| Value                    | `Valid()` | `Validate()` | `Normalized()` |
|--------------------------|-----------|--------------|----------------|
| `language.Code("EN")`    | `false`   | `nil`        | `en`           |
| `language.Code("en-US")` | `false`   | `nil`        | `en`           |

`types.Validate` and `types.DeepValidate` call `Validate`, so they accept
`"EN"`, see below. Check `Valid` or `ValidAndNormalized` to require the
canonical code.

## Why `types.Validate` prefers `ValidatErr`

`types.Validate(v any) error` and `types.TryValidate(v any) (error, bool)`
check whether `v` implements `ValidatErr` or `Validator`. If it implements
both, only `Validate()` is called. There are two reasons:

1. **The error says why.** A `false` from `Valid()` carries no reason, so
   `types.Validate` can only build a generic message from it.
2. **A type can pair a lenient `Validate` with a strict `Valid`.** Then
   `Validate` decides, because it is the method the type wrote to explain its
   rules. `types.DeepValidate` and the scanning in `strfmt` prefer `Validate`
   for the same reason.

The test case `strictValidLenientValidate` in
[validator_test.go](../validator_test.go) pins this precedence.

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/language"
)

type Profile struct {
	Language language.Code
}

func main() {
	lang := language.Code("EN")
	fmt.Println(lang.Valid(), lang.Validate())

	// types.Validate prefers the lenient Validate over the strict Valid
	fmt.Println(types.Validate(lang))
	fmt.Println(types.DeepValidate(Profile{Language: "xx"}))

	// Output:
	// false <nil>
	// <nil>
	// [struct field Language: invalid language.Code: "xx"]
}
```

What `types.Validate` returns:

- The error of `Validate()` comes back unchanged, the same error value, not
  wrapped. `errors.Is` with the type's own sentinel works directly.
- For a `Validator` whose `Valid()` is `false` it builds an error from `%T` and
  `%#v`, like `invalid main.Status value: "unknown"`. For a pointer, `%#v`
  prints the address, like `invalid *main.Status value:
  (*main.Status)(0x...)`. The value ends up in the message, so a type
  that holds secrets should implement `ValidatErr` with a message that leaves
  the value out.
- For a nil pointer whose element type has the method with a value receiver,
  there is no value to call it on. `types.Validate` returns `nil` and
  `TryValidate` returns `nil, true`, the same way `DeepValidate` skips nil
  pointers. A method with a pointer receiver is called with the nil pointer and
  can handle it itself.

## No sentinel error for "invalid"

The module used to export `types.ErrInvalidValue`, and `types.Validate`
returned it when `Valid()` was `false`. The error of a `ValidatErr` never
wrapped it, so `errors.Is(err, types.ErrInvalidValue)` returned `false` for
most invalid values: a check written against it passed invalid data. It was
removed (see the Removed section of [CHANGELOG.md](../CHANGELOG.md)), and the
rule is now simple: any non-nil error means the value is invalid.

Packages that want callers to classify errors define their own sentinels and
wrap them with `%w`: `vat.ErrInvalidID`, `account.ErrInvalidNumber`,
`account.ErrAlphanumericNumber`, `uu.ErrNilID`, `uu.ErrInvalidVariant`, and
the error type `uu.ErrInvalidVersion`.

## How `DeepValidate` walks a value

`types.DeepValidate(v any) []error` walks `v` with reflection and calls
`types.Validate` on `v` and on every value reachable from it. Each error gets
the path to the value that caused it as a prefix:

```
DeepValidate(&invoice)
  invoice itself ............ "<error>"                       no prefix
  .Status ................... "struct field Status: <error>"
  .Lines[1] ................. "struct field Lines -> element [1]: <error>"
  .Meta["k"] ................ "struct field Meta -> map value [\"k\"]: <error>"
```

The rules of the walk, from the doc comment of `DeepValidate` and its tests:

- **Pointers and interfaces** are followed to any depth. A nil pointer,
  interface or func stops the walk without an error: there is nothing to
  validate.
- **Pointer receivers** are found: if only `*T` implements the interface, like
  `types.LenString`, the value is validated through a pointer to it. A value
  that is not addressable, like a map value, is copied first, so validation
  never writes into the map. Only the value passed to `DeepValidate` is not
  copied, so a pointer receiver `Validate` method calling
  `types.DeepValidate(*receiver)` doesn't call itself again. Such a nested
  call has its own cycle detection, and an outer `DeepValidate` reports the
  field errors a second time while descending, so check only rules that
  span several fields in `Validate`.
- **Struct fields:** exported fields, including the exported fields promoted
  from an unexported embedded struct. Other unexported fields are skipped,
  reflection can't call methods on them.
- **Slices, arrays and maps:** every element and every map value. Map keys
  are not validated, but sorted with `types.CompareReflectValue`, so the
  errors come in the same order on every run.
- **Cycles:** a pointer, map or slice that is already on the current path is
  not entered again, so cyclic data ends the walk instead of overflowing the
  stack. Data shared by two paths without a cycle is validated on each path.
- **Once per value:** in a chain of pointers, the first level that implements
  `ValidatErr` or `Validator` is validated and the levels below are skipped,
  so a stateful validator is called once.
- **Structs are validated and descended into.** If a struct's own `Validate`
  also checks its fields, `DeepValidate` reports the same problem twice: once
  from the struct, without a path, and once from the field, with a path.

Each prefixed error is built with `fmt.Errorf("%s: %w", path, err)`, so
`errors.Is` and `errors.As` see through the path to the original error. The
result is a slice, so callers can count the errors or show one per field;
`errors.Join(types.DeepValidate(v)...)` turns it into a single error, and into
`nil` when the slice is empty.

## The normalization contract

The doc comment of `types.Normalizable` asks for three things: return the
canonical value, return an error if there is none, and be idempotent, so that
normalizing a normalized value returns it unchanged. On top of that the module
follows a convention, introduced for all `Normalized` methods at once in
January 2025: **on error, return the receiver unchanged together with the
error**. The error message names the original input, like
`invalid country.Code: 'XX'`. `date.Date("foo").Normalized()` returns `"foo"`
and `money.Currency("  ").Normalized()` returns `"  "`, each with its error.

When an invalid value should simply become null, use `NormalizedOrNull`. It
exists on `bank.IBAN`, `bank.BIC`, their nullable types, `country.NullableCode`,
`date.Date`, `date.NullableDate`, `money.Currency`, `money.NullableCurrency`,
`vat.ID` and `vat.NullableID`.

## Normalized output

Because `Validate` accepts every spelling that `Normalized` accepts, the
`Value` and `MarshalJSON` methods of these types write the normalized form,
so `"eur"` and `"EUR"` can't reach a database column or an API response as
two different currencies:

| Type                                       | `Value()`                    | `MarshalJSON()`     |
|--------------------------------------------|------------------------------|---------------------|
| `country.Code`, `country.NullableCode`     | normalized if valid          | normalized if valid |
| `vat.ID`, `vat.NullableID`                 | normalized if valid          | normalized if valid |
| `money.Currency`, `NullableCurrency`       | normalized if valid          | normalized if valid |
| `language.Code`                            | normalized if valid          | normalized if valid |
| `bank.IBAN`, `bank.NullableIBAN`           | normalized if valid          | normalized if valid |
| `bank.BIC`, `bank.NullableBIC`             | normalized if valid          | normalized if valid |
| `date.Date`, `date.NullableDate`           | normalized, error if invalid | normalized if valid |
| `email.Address`, `NullableAddress`         | normalized if valid          | normalized if valid |
| `email.AddressList`, `NullableAddressList` | normalized if valid          | normalized if valid |
| `email.AddressSet`                         | normalized if valid          | normalized if valid |

A value that can't be normalized is written unchanged, except by the `Value`
methods of the `date` types, which return an error. A null value of a
nullable type is written as SQL NULL and JSON `null`.

`String()` of these types returns the normalized form too, or the original
if it can't be normalized. A valid `email.AddressList` without addresses,
like `"Undisclosed recipients"`, normalizes to `""`, but its `String`,
`Value`, `MarshalJSON` and `MarshalText` write the placeholder
`undisclosed-recipients:;`, because `Scan` rejects the empty string.
The normalized form is UTF-8, so a display name
like `"Jürgen Müller" <j@x.de>` stays readable; the email types have an
extra `ASCII()` method that RFC 2047 encodes non-ASCII names for message
headers. `StringOr(alternative)` of the nullable types returns the
normalized form, or the alternative for a null value and for a value that
can't be normalized. The set types are the exception: the `String()` of
`email.AddressSet`, `types.Set`, `uu.IDSet` and `strutil.StringSet` lists
the elements as given, because normalizing could merge two elements and
the string would no longer show every element of the set.

`MarshalText` returns `String()`, so `strfmt` and `encoding/xml` write the
canonical form too. `encoding/xml` calls it through a nil pointer in an
attribute field and panics, like for a `*time.Time` attribute, so tag such
fields `omitempty`. One output stays raw: `encoding/json` writes the keys
of a map with a string key type as they are, so
`map[money.Currency]int{"eur": 1}` marshals as `{"eur":1}`; normalize such
keys before building the map.

Reading doesn't normalize: `Scan` and JSON decoding store the text without
checking it, except the `Scan` methods of `date.Date` and
`date.NullableDate`, which normalize and reject invalid dates. So a Go value can hold `"eur"` until it is written. To compare values
in Go, call `Normalized()` where input enters your program, handle the error,
and keep the result. Everything after that point can rely on
`ValidAndNormalized()` being true.

## Related

- [How to validate structs](howto-validate-structs.md)
- [Null conventions](explanation-null-conventions.md): why a null `NullableX`
  is valid.
- [Lenient parsing](explanation-lenient-parsing.md)
- Reference: [root package](../README.md), [country](../country/README.md),
  [language](../language/README.md), [money](../money/README.md),
  [bank](../bank/README.md), [vat](../vat/README.md),
  [email](../email/README.md), [date](../date/README.md).
