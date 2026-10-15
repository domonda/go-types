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
| `Validate() error`          | `types.Validatable`              | If not, why not? `nil` means valid.       |
| `Normalized() (T, error)`   | `types.Normalizable[T]`          | What is the canonical form?               |
| `ValidAndNormalized() bool` | `types.NormalizableValidator[T]` | Is the value valid and already canonical? |

`types.NormalizableValidator[T]` combines `Validator`, `Normalizable[T]` and
`ValidAndNormalized`. These types assert it at compile time: `bank.IBAN`,
`bank.BIC`, `bank.NullableIBAN`, `country.Code`, `date.Date`,
`date.NullableDate`, `date.YearMonth`, `date.NullableYearMonth`,
`date.YearQuarter`, `date.NullableYearQuarter`, `email.Address`,
`email.NullableAddress`,
`email.AddressList`, `email.AddressSet`, `language.Code`, `money.Currency`
and `vat.ID`. The `Normalized` method of the
`date` types takes no language hint; `NormalizedLang(lang language.Code)` takes
the hint that decides between day-first and month-first dates with slashes.

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
and returning its error, and `Valid` agrees with it:

| Value                                      | `Valid()` | `Normalized()`                  | `ValidAndNormalized()` |
|--------------------------------------------|-----------|---------------------------------|------------------------|
| `country.Code(" at ")`                     | `true`    | `AT`                            | `false`                |
| `country.Code("Österreich")`               | `true`    | `AT`                            | `false`                |
| `money.Currency("€")`                      | `true`    | `EUR`                           | `false`                |
| `money.Currency("Canadian Dollar")`        | `true`    | `CAD`                           | `false`                |
| `language.Code("en-US")`                   | `true`    | `en`                            | `false`                |
| `vat.ID("atu 1022 3006")`                  | `true`    | `ATU10223006`                   | `false`                |
| `bank.IBAN("DE89 3704 0044 0532 0130 00")` | `true`    | `DE89370400440532013000`        | `false`                |
| `bank.BIC("BKAUATWW")`                     | `true`    | `BKAUATWWXXX`                   | `false`                |
| `bank.BIC("BKAU AT WW")`                   | `true`    | `BKAUATWWXXX`                   | `false`                |
| `email.Address("John@Example.com")`        | `true`    | `john@example.com`              | `false`                |
| `date.Date("25.12.2023")`                  | `true`    | `2023-12-25`                    | `false`                |
| `country.Code("XX")`                       | `false`   | error                           | `false`                |
| `date.Date("0001-01-01")`                  | `false`   | error, the zero date is no date | `false`                |

The reasoning: a valid value in another spelling is still the same value, so
rejecting it would only reject real-world input. Leniency is safe because
reading stores and writing writes the normalized form, see
[Reading and writing](#reading-and-writing). A value built in Go with a
conversion like `money.Currency("eur")` keeps its spelling, so where a
contract requires the canonical form, like a map key or a comparison with
`==`, check `ValidAndNormalized` or keep the result of `Normalized`.

Each type decides how far its leniency goes. `bank.IBAN` removes whitespace
but doesn't change the letter case, so `de89370400440532013000` fails with
`invalid IBAN country code`. `country.Code` accepts German names and ITU codes
from `AltCodes` but not English names; `country.ParseCode` adds those.
`bank.BIC` removes whitespace, but `BIC.Parse` and `BICParser` read the BIC
as given, so whitespace makes it invalid there.
The package READMEs and [Lenient parsing](explanation-lenient-parsing.md) list
the details.

## Why `types.Validate` prefers `Validatable`

`types.Validate(v any) error` and `types.TryValidate(v any) (error, bool)`
check whether `v` implements `Validatable` or `Validator`. If it implements
both, only `Validate()` is called. There are two reasons:

1. **The error says why.** A `false` from `Valid()` carries no reason, so
   `types.Validate` can only build a generic message from it.
2. **A type can pair a lenient `Validate` with a strict `Valid`.** Then
   `Validate` decides, because it is the method the type wrote to explain its
   rules. `types.DeepValidate` and the scanning in `strfmt` prefer `Validate`
   for the same reason. The types of this module don't do that, their `Valid`
   and `Validate` agree, but your own types can.

The test case `strictValidLenientValidate` in
[validator_test.go](../validator_test.go) pins this precedence.

```go
package main

import (
	"fmt"
	"strings"

	"github.com/domonda/go-types"
)

// Status accepts any letter case in Validate,
// but Valid only the canonical lower case
type Status string

func (s Status) Valid() bool {
	return s == "open" || s == "closed"
}

func (s Status) Validate() error {
	switch strings.ToLower(string(s)) {
	case "open", "closed":
		return nil
	}
	return fmt.Errorf("invalid Status: %q", string(s))
}

type Ticket struct {
	Status Status
}

func main() {
	status := Status("OPEN")
	fmt.Println(status.Valid(), status.Validate())

	// types.Validate prefers the lenient Validate over the strict Valid
	fmt.Println(types.Validate(status))
	fmt.Println(types.DeepValidate(Ticket{Status: "pending"}))

	// Output:
	// false <nil>
	// <nil>
	// [struct field Status: invalid Status: "pending"]
}
```

What `types.Validate` returns:

- The error of `Validate()` comes back unchanged, the same error value, not
  wrapped. `errors.Is` with the type's own sentinel works directly.
- For a `Validator` whose `Valid()` is `false` it builds an error from `%T` and
  `%#v`, like `invalid main.Status value: "unknown"`. For a pointer, `%#v`
  prints the address, like `invalid *main.Status value:
  (*main.Status)(0x...)`. The value ends up in the message, so a type
  that holds secrets should implement `Validatable` with a message that leaves
  the value out.
- For a nil pointer whose element type has the method with a value receiver,
  there is no value to call it on. `types.Validate` returns `nil` and
  `TryValidate` returns `nil, true`, the same way `DeepValidate` skips nil
  pointers. A method with a pointer receiver is called with the nil pointer and
  can handle it itself.

## No sentinel error for "invalid"

The module used to export `types.ErrInvalidValue`, and `types.Validate`
returned it when `Valid()` was `false`. The error of a `Validatable` never
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
  `Validatable` or `Validator` is validated and the levels below are skipped,
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
The writing methods rely on it: they write the result of `Normalized`
whether it succeeded or not. A zero date like `date.Date("0001-01-01")` has
no canonical form, its `Normalized` returns an error, and
`date.NullableDate` normalizes it to null.

When an invalid value should simply become null, use `NormalizedOrNull`. It
exists on `bank.IBAN`, `bank.BIC`, their nullable types, `country.NullableCode`,
`date.Date`, `date.NullableDate`, `money.Currency`, `money.NullableCurrency`,
`vat.ID` and `vat.NullableID`.

## Reading and writing

Because `Validate` accepts every spelling that `Normalized` accepts, the
types normalize where a value crosses the boundary of the program, so
`"eur"` and `"EUR"` can't reach a database column, an API response or a
comparison in Go as two different currencies. The
[README](../README.md#reading-and-writing-values) has the tables for every
format.

- **Reading** with `Scan`, `UnmarshalJSON`, `UnmarshalXML`,
  `UnmarshalXMLAttr` and `ScanString` stores a valid value normalized, a
  blank value as the empty value and an invalid value unchanged, without an
  error. One bad value doesn't fail a whole query or document, and
  `Validate` reports it, with its field path from `DeepValidate`. A value
  read from SQL, JSON or XML is canonical if it is valid, so `==` and map
  lookups work on it.
- **`Value`** writes the normalized value, SQL `NULL` for a blank value, and
  returns the error of `Normalized` for an invalid value, because the
  database is where an invalid value would stay.
- **`MarshalJSON` and `MarshalText`**, which `encoding/xml` and `strfmt`
  use, write the normalized value and an invalid value unchanged, so a
  response or log that contains one still renders. A null value of a
  nullable type is JSON `null` and the empty text.
- **`fmt` and `String`** show the stored value. Most of the types have no
  `String` method, so `fmt` prints the underlying string, and
  `country.Code`, `language.Code` and `money.Currency` have the `String`
  that go-enum generates, which returns the stored string. A normalizing
  `String` would show a non-canonical or invalid value as if it were the
  canonical one. `StringOr(alternative)` of the nullable types returns the
  stored value, or the alternative for null. The normalized form is UTF-8,
  so a display name like `"Jürgen Müller" <j@x.de>` stays readable; the
  email types have an extra `ASCII()` method that RFC 2047 encodes
  non-ASCII names for message headers.

Two inputs stay as given, because the types have no `UnmarshalText`:
`encoding/json` reads and writes the keys of a map with a
string key type as they are, so `map[money.Currency]int{"eur": 1, "EUR": 2}`
keeps both keys, and `encoding/xml` reads a `,chardata` field only through
`UnmarshalText`, so such a field keeps the text as given.

A nil pointer attribute field, like `Currency *money.Currency` with the tag
`xml:"currency,attr"`, is left out, also without `omitempty`. The types
implement `MarshalXMLAttr` with a pointer receiver for it, because
`encoding/xml` would call the value receiver `MarshalText` through the nil
pointer and panic, like it does for a `*time.Time` attribute.

### Collections keep their elements as given

`types.Set`, `email.AddressSet`, `email.AddressList` and
`email.NullableAddressList` read and write their elements unchanged in
every format, because normalizing could merge two elements:
`{"John@Example.com", "john@example.com"}` would be written as one address
and read back as a set with one element. `Normalized()` and
`email.AddressSet.NormalizedAddressList()` normalize explicitly. `Value` of
the email collections still refuses an invalid address, and
`AddressSet.MarshalText` also one that would not read back as one entry of
an address list. The `String()` of `email.AddressSet`, `types.Set`,
`uu.IDSet` and `strutil.StringSet` lists the elements as given too.

## Related

- [How to validate structs](howto-validate-structs.md)
- [Null conventions](explanation-null-conventions.md): why a null `NullableX`
  is valid.
- [Lenient parsing](explanation-lenient-parsing.md)
- Reference: [root package](../README.md), [country](../country/README.md),
  [language](../language/README.md), [money](../money/README.md),
  [bank](../bank/README.md), [vat](../vat/README.md),
  [email](../email/README.md), [date](../date/README.md).
