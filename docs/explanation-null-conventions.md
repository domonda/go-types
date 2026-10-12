# Null conventions: X and NullableX

This page explains how the module represents a missing value in Go, SQL and
JSON, why it uses two different constructions for that, and where the
packages still disagree. The package READMEs list the API of every type;
this page explains the rules behind them.

## The problem

SQL columns and JSON fields can be null. Go values cannot: a `string` is at
least `""`, a `[16]byte` is sixteen zero bytes, a `float64` is `0`. Every type
that crosses a database or JSON boundary has to decide what its Go zero value
means on the other side, and what a NULL coming in turns into.

The common Go answers each have a cost:

- A pointer (`*string`, `*uu.ID`) allocates, needs a nil check before every
  read, and makes `==` compare addresses instead of values. Methods like
  `Validate` or `Normalized` can't be called on a nil pointer.
- `sql.NullString` or `sql.Null[T]` keep a separate `Valid` flag, but have no
  JSON methods, so `encoding/json` writes them as an object like
  `{"String":"","Valid":false}`.
- A plain value type can't tell "no value" from "the zero value". For an ID
  that is harmless, for an amount of `0.00` it loses information.

## The approach: a type pair per value

Every domain value of the module comes as a pair of types. `X` is a value that
must be present, `NullableX` is the same value when it may be absent. The field
type in a struct documents whether the column or JSON field is optional, and
the compiler keeps the two apart: `x.Nullable()` converts to the nullable type,
`n.Get()` converts back (and panics on null), `n.GetOr(def)` converts back with
a default.

Every `NullableX` implements `nullable.NullSetable[X]` (`IsNull`, `Set`,
`SetNull`, `Get`, `GetOr`), and the packages check that at compile time, for
example `var _ nullable.NullSetable[ID] = (*NullableID)(nil)` in
`uu/nullableid.go`. Almost all of them also implement `driver.Valuer`,
`sql.Scanner` and `json.Marshaler` so that null becomes SQL NULL and JSON
`null`.

How `NullableX` is built depends on one question:

```
                 Is the zero value of X a valid X?
                                |
              no ---------------+--------------- yes
              |                                   |
     type NullableX X                  type NullableX = nullable.Type[X]
     the zero value is null            struct{ value X; valid bool }
     uu.NullableID, bank.NullableIBAN, money.NullableAmount,
     country.NullableCode, ...         money.NullableDecimalAmount, ...
```

### The zero value as null

If the zero value of `X` is never valid, nothing is lost by letting it mean
null. `NullableX` is then a defined type with the same underlying type as `X`,
and its zero value is the null value. The third column shows the error that
`X.Validate()` returns for the zero value, which is what frees it up:

| Nullable type              | Null value             | Zero value rejected by `X.Validate()` |
|----------------------------|------------------------|---------------------------------------|
| `uu.NullableID`            | `uu.IDNull`            | `Nil UUID`                            |
| `bank.NullableIBAN`        | `bank.IBANNull`        | `empty IBAN`                          |
| `bank.NullableBIC`         | `bank.BICNull`         | `invalid BIC "" length: 0`            |
| `date.NullableDate`        | `date.Null`            | `too short for a date: ""`            |
| `date.NullableYearMonth`   | `date.YearMonthNull`   | `invalid year month: ""`              |
| `date.NullableYearQuarter` | `date.YearQuarterNull` | `invalid year quarter: ""`            |
| `vat.NullableID`           | `vat.Null`             | `invalid VAT ID: "" is too short`     |
| `money.NullableCurrency`   | `money.CurrencyNull`   | `empty currency code`                 |
| `country.NullableCode`     | `country.Null`         | `invalid country.Code: ''`            |
| `account.NullableNumber`   | `account.NumberNull`   | `invalid account number: ""`          |
| `email.NullableAddress`    | `email.AddressNull`    | `empty email address`                 |

What this buys:

- Converting between `X` and `NullableX` is a plain type conversion. A
  `country.NullableCode` is still a string: it compares with `==`, works as a
  map key and has a constant null value.
- Validating a `NullableX` accepts null, because absence is a valid state of
  an optional value. `ValidAndNotNull()` asks for "present and valid"; the
  bank, country, date, money and vat nullable types have it.
- `X.Nullable()` maps the zero value to null for free, because it is the same
  bit pattern.

The cost is that `NullableX` can't hold the zero value of `X` as a present
value. Since that value is invalid anyway, nothing real is lost.
`date.NullableDate` widens the rule: `IsNull` is also true for `"0001-01-01"`
(the zero `time.Time`) and `"0000-00-00"`, so those write SQL NULL too.

### `nullable.Type[T]` when zero is a real value

`money.Amount(0)` is a valid amount. If the zero value meant null, a zero
invoice total would turn into an unknown one. For these types the module wraps
the value in [`nullable.Type[T]`](../nullable/README.md), a struct with a
separate validity flag. The zero `Type[T]` is null, `nullable.TypeFrom(v)` is
present even when `v` is zero. The money package declares its nullable types as
aliases of it:

| Alias                         | Definition                           |
|-------------------------------|--------------------------------------|
| `money.NullableAmount`        | `nullable.Type[money.Amount]`        |
| `money.NullableDecimalAmount` | `nullable.Type[money.DecimalAmount]` |
| `money.NullableRate`          | `nullable.Type[money.Rate]`          |
| `money.NullableCentAmount`    | `nullable.Type[money.CentAmount]`    |

Because they are aliases, they have exactly the methods of `Type[T]`. They are
created with `nullable.TypeFrom`, `nullable.TypeFromPtr`, or helpers like
`money.NullableAmountFrom(float64)` and `DecimalAmount.Nullable()`, which
always returns a present value.

`Type[T].Value` calls the `driver.Valuer` of `T` if it has one
(`DecimalAmount` writes a decimal string), and otherwise converts with
`driver.DefaultParameterConverter` (`Amount` and `Rate` become `float64`,
`CentAmount` becomes `int64`). `Scan` uses a copy of the conversion code of
`database/sql`, so `T` needs no `Scan` method.

The trade-offs of the wrapper:

- It is a struct, so it can't be compared to a constant, and every read goes
  through `Get`, `GetOr` or `Ptr`.
- `encoding/json`'s `omitempty` never omits a struct; use `omitzero`, see below.
- `Type[T]` works for any `T`, including your own types. For a struct `T`
  without a `Value` method, `Value` returns the error of
  `driver.DefaultParameterConverter`, which rejects structs.

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/uu"
)

func main() {
	// uu.ID has no null: its zero value is written as the Nil UUID.
	// uu.NullableID uses the same 16 zero bytes as SQL NULL.
	idValue, _ := uu.IDNil.Value()
	nullValue, _ := uu.IDNull.Value()
	fmt.Println(idValue, nullValue)
	fmt.Println(uu.IDNil.Validate(), uu.IDNull.Validate())

	// nullable.Type[T] keeps a present zero apart from null.
	var unknown money.NullableAmount
	zero := nullable.TypeFrom(money.Amount(0))
	a, _ := json.Marshal(unknown)
	b, _ := json.Marshal(zero)
	fmt.Println(string(a), string(b))

	// Output:
	// 00000000-0000-0000-0000-000000000000 <nil>
	// Nil UUID <nil>
	// null 0
}
```

### Alternatives and why the module does not use them

- **Pointers everywhere.** They work with `database/sql` and `encoding/json`
  out of the box, but bring the costs listed in the problem section, and the
  domain types carry methods like `Validate` and `Normalized` that need a
  value.
- **`nullable.Type[T]` for every type.** It would work, but for string and
  array types it adds a flag that carries no information and gives up
  comparison, map keys and constant null values.
- **`sql.Null[T]`.** It has no JSON methods, so the JSON side would need a
  second type anyway.

## `nullable` and `notnull`

For strings, arrays and raw JSON the module has two packages that use the same
type names. They differ in what a nil or empty Go value turns into:

| Go value                                | [`nullable`](../nullable/README.md) | [`notnull`](../notnull/README.md) |
|-----------------------------------------|-------------------------------------|-----------------------------------|
| nil slice                               | SQL `NULL`, JSON `null`             | SQL `{}`, JSON `[]`               |
| empty slice                             | SQL `{}`, JSON `[]`                 | SQL `{}`, JSON `[]`               |
| nil `JSON`                              | SQL `NULL`, JSON `null`             | SQL `{}`, JSON `{}`               |
| `TrimmedString("")`                     | SQL `NULL`, JSON `null`             | SQL `''`, JSON `""`               |
| SQL `NULL` scanned into a slice         | nil slice                           | empty non-nil slice               |
| SQL `NULL` scanned into `JSON`          | nil                                 | `JSON("{}")`                      |
| SQL `NULL` scanned into `TrimmedString` | null (`""`)                         | error                             |

`notnull` exists for columns like `text[] NOT NULL DEFAULT '{}'`. With a
`nullable` array, a struct field that was never assigned holds nil and writes
NULL, which violates the constraint. With a `notnull` array the same nil writes
`{}`, and reading never returns nil, so code can range over and append to the
value without a nil check. The price is that `notnull` can't express "unknown":
SQL NULL and the empty array read back the same.

Both packages have `StringArray`, `IntArray`, `FloatArray`, `NullBoolArray`,
`TrimmedString` and `JSON`. Only `nullable` has `NonEmptyString`, `Time`,
`Type[T]`, `BoolArray`, `NullIntArray` and `NullFloatArray`.

## Three states for sets and slices

A list of IDs or tags has three meaningful states: unknown, known to be empty,
and known members. The module maps them to three distinct wire values:

| Go value       | SQL `Value()` | JSON        |
|----------------|---------------|-------------|
| nil            | `NULL`        | `null`      |
| empty, non-nil | `{}`          | `[]`        |
| with elements  | `{"a","b"}`   | `["a","b"]` |

Decoding reverses each row: SQL NULL and JSON `null` give nil, `{}` and `[]`
give an empty non-nil value. The rule holds for:

- `uu.IDSlice` and `uu.IDSet`, for SQL and JSON;
- the `nullable` arrays, for SQL, and for JSON through `encoding/json`'s
  default slice encoding;
- `types.Set[T]`, for JSON (it has no SQL methods);
- `email.AddressSet`, for SQL (its JSON is Go's default map encoding).

Before the 2026-09-02 release, `uu.IDSet.Value` wrote NULL for an empty set
and silently nulled the column; see [CHANGELOG.md](../CHANGELOG.md).
[explanation-sets.md](explanation-sets.md) compares all set types.

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/notnull"
	"github.com/domonda/go-types/uu"
)

func main() {
	id := uu.IDMust("0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d")
	for _, ids := range []uu.IDSlice{nil, {}, {id}} {
		sqlValue, _ := ids.Value()
		jsonValue, _ := json.Marshal(ids)
		fmt.Printf("%#v %s\n", sqlValue, jsonValue)
	}

	// notnull collapses nil into the empty array.
	var tags notnull.StringArray
	sqlValue, _ := tags.Value()
	jsonValue, _ := json.Marshal(tags)
	fmt.Printf("%#v %s\n", sqlValue, jsonValue)

	// Output:
	// <nil> null
	// "{}" []
	// "{\"0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d\"}" ["0e5f8a45-5f2c-4d6c-9e57-6a8d1e1b2c3d"]
	// "{}" []
}
```

Three places lose the third state:

- `omitempty` drops an empty slice or map, so `[]` turns into an absent field
  and decodes as nil. `omitzero` only drops nil.
- `uu.IDSlice.MarshalText` writes `[]` for both nil and empty, and
  `UnmarshalText` reads `[]` back as nil.
- `uu.IDSet{}.AsSlice()` returns nil, which writes NULL, while
  `uu.IDSlice(nil).AsSet()` returns an empty set, which writes `{}`.

A nil set or map panics on `Insert`, exactly like a nil Go map. A set read from
SQL NULL or JSON `null` is nil, so allocate it before inserting.
`email.AddressSet` has pointer-receiver `Add` and `AddSet` methods that
allocate for you.

## `uu.IDNil` and `uu.IDNull`

Both are sixteen zero bytes. The type decides what they mean:

| Behaviour           | `uu.IDNil` (type `uu.ID`)                | `uu.IDNull` (type `uu.NullableID`) |
|---------------------|------------------------------------------|------------------------------------|
| `Value()`           | `"00000000-0000-0000-0000-000000000000"` | `nil`, SQL NULL                    |
| `MarshalJSON`       | `"00000000-0000-0000-0000-000000000000"` | `null`                             |
| `Scan(nil)`         | error `cannot convert <nil> to uu.ID`    | `uu.IDNull`                        |
| JSON `null` decoded | no-op, the field keeps its value         | `uu.IDNull`                        |
| `Valid()`           | `false`, `Validate()` returns `Nil UUID` | `true`                             |

`uu.IDNil` is the "not assigned yet" state of a required ID. Validation
rejects it, so it never ends up in a primary key by accident. A column that may
be empty is `uuid NULL` in SQL and `uu.NullableID` in Go, with `uu.IDNull` as
its null value.

## JSON: `omitzero`, not `omitempty`

`omitempty` decides by kind: it omits `false`, `0`, nil pointers and
interfaces, and arrays, slices, maps and strings of length zero. Go 1.24 added
`omitzero`, which omits the zero value of the type, or a value whose
`IsZero()` method returns true. For the null conventions above that makes a
difference:

| Field type                                         | `omitempty` omits    | `omitzero` omits                     |
|----------------------------------------------------|----------------------|--------------------------------------|
| string based `NullableX`, `nullable.NonEmptyString` | null (`""`)          | null (`""`)                          |
| `uu.NullableID` (an array, never empty)            | never, writes `null` | `uu.IDNull`                          |
| `nullable.Type[T]` and the money aliases           | never, writes `null` | null, but not a present zero         |
| `nullable.Time` (a struct)                         | never, writes `null` | the zero time                        |
| `uu.IDSlice`, `nullable` arrays                    | nil and empty        | nil only                             |
| `date.NullableDate`                                | `""`                 | `""`, `"0001-01-01"`, `"0000-00-00"` |
| `money.DecimalAmount`, `money.CentAmount`          | `CentAmount(0)` only | a zero amount                        |

Tag optional fields with `omitzero`. It omits exactly the null value of every
`NullableX` and keeps the empty set apart from the nil set. Be careful with
`omitzero` on non-nullable types that have an `IsZero` method:
`money.DecimalAmount` and `money.CentAmount` drop a real zero amount, and
`date.Date` drops `"0001-01-01"`.

## Summary: Go zero value to SQL and JSON

What the zero value of each type writes, and what scanning SQL NULL into it
does. "kind" means the type has no `Value` method and `database/sql` converts
it by its underlying kind.

| Type                                               | Zero value | `Value()`          | JSON            | `Scan(nil)`    |
|----------------------------------------------------|------------|--------------------|-----------------|----------------|
| `uu.ID`                                            | `IDNil`    | Nil UUID string    | Nil UUID string | error          |
| `uu.NullableID`                                    | `IDNull`   | NULL               | `null`          | `IDNull`       |
| `uu.IDSlice`, `uu.IDSet`                           | nil        | NULL               | `null`          | nil            |
| `bank.IBAN`, `bank.BIC`                            | `""`       | `''`               | `""`            | `""`           |
| `bank.NullableIBAN`, `bank.NullableBIC`            | `""`       | NULL               | `null`          | null           |
| `date.Date`                                        | `""`       | NULL               | `""`            | `""`           |
| `date.NullableDate`                                | `""`       | NULL               | `null`          | null           |
| `date.YearMonth`, `date.YearQuarter`               | `""`       | no SQL methods     | `""`            | no SQL methods |
| `date.NullableYearMonth`, `date.NullableYearQuarter` | `""`       | no SQL methods     | `""`            | no SQL methods |
| `vat.ID`                                           | `""`       | `''`               | `""`            | error          |
| `vat.NullableID`                                   | `""`       | NULL               | `null`          | null           |
| `money.Currency`                                   | `""`       | NULL               | `""`            | `""`           |
| `money.NullableCurrency`                           | `""`       | NULL               | `null`          | null           |
| `country.Code`                                     | `""`       | NULL               | `""`            | `""`           |
| `country.NullableCode`                             | `""`       | NULL               | `null`          | null           |
| `language.Code`                                    | `""`       | NULL               | `""`            | `""`           |
| `account.Number`                                   | `""`       | NULL               | `null`          | error          |
| `account.NullableNumber`                           | `""`       | NULL               | `null`          | null           |
| `email.Address`                                    | `""`       | no SQL methods     | `""`            | no SQL methods |
| `email.NullableAddress`                            | `""`       | NULL               | `null`          | null           |
| `email.AddressSet`                                 | nil        | NULL               | `null`          | nil            |
| `money.Amount`, `money.Rate`                       | `0`        | `float64` 0 (kind) | `0`             | error          |
| `money.CentAmount`                                 | `0`        | `int64` 0 (kind)   | `0`             | error          |
| `money.DecimalAmount`                              | `0`        | `"0"`              | `0`             | error          |
| `money.Nullable*` (`nullable.Type[T]`)             | null       | NULL               | `null`          | null           |
| `nullable.NonEmptyString`, `nullable.TrimmedString` | `""`       | NULL               | `null`          | null           |
| `notnull.TrimmedString`                            | `""`       | `''`               | `""`            | error          |
| `nullable.Time`                                    | zero time  | NULL               | `null`          | null           |
| `nullable.JSON`                                    | nil        | NULL               | `null`          | nil            |
| `notnull.JSON`                                     | nil        | `{}`               | `{}`            | `{}`           |
| `nullable.StringArray` and the other arrays        | nil        | NULL               | `null`          | nil            |
| `notnull.StringArray` and the other arrays         | nil        | `{}`               | `[]`            | empty slice    |
| `types.Set[T]`, `strutil.StringSet`                | nil        | no SQL methods     | `null`          | no SQL methods |

## Remaining inconsistencies

The convention is not applied evenly yet. These differences exist in the
current code:

- **`Scan("")` of nullable strings.** `nullable.NonEmptyString.Scan("")` and
  `email.NullableAddress.Scan("")` return an error (`can't scan empty string as
  nullable.NonEmptyString`), while `nullable.TrimmedString.Scan("")` and the
  other string based `NullableX` types store null.
- **What a non-nullable zero value writes.** `date.Date`, `country.Code`,
  `language.Code`, `money.Currency` and `account.Number` write SQL NULL.
  `bank.IBAN`, `bank.BIC`, `vat.ID` and `notnull.TrimmedString` write `''`.
  `uu.ID` writes the Nil UUID.
- **What a non-nullable type does with SQL NULL.** `bank.IBAN`, `bank.BIC`,
  `date.Date`, `country.Code`, `language.Code` and `money.Currency` scan it as
  their zero value. `uu.ID`, `vat.ID`, `account.Number`,
  `notnull.TrimmedString` and `money.DecimalAmount` return an error. So
  `account.Number("")` writes a NULL it can't read back.
- **`account.Number("")` marshals to JSON `null`**, while unmarshalling JSON
  `null` into an `account.Number` returns an error.
- **`date.NullableYearMonth` and `date.NullableYearQuarter`** have no SQL
  methods and no `MarshalJSON`, so their null is written as the string `""`
  instead of NULL or `null`.
- **`language.Code` has no nullable type.** `language.Null` writes SQL NULL,
  but `language.Null.Validate()` returns an error, so an optional language
  field fails `types.DeepValidate`.
- **Whitespace-only input.** `"  "` is null for `country.NullableCode`,
  `bank.NullableIBAN`, `bank.NullableBIC`, `money.NullableCurrency`,
  `vat.NullableID`, `email.NullableAddress`, `date.NullableDate` and
  `nullable.TrimmedString`. `email.NullableAddressList`,
  `account.NullableNumber` (although its `MarshalJSON` writes `null`),
  `nullable.NonEmptyString`, `date.NullableYearMonth` and
  `date.NullableYearQuarter` treat it as a non-null value.
- **JSON `""` into a nullable type.** For the string based `NullableX` types
  and `nullable.NonEmptyString` it means null. For `money.NullableAmount`,
  `NullableDecimalAmount` and `NullableRate` it decodes as a present zero,
  because `Amount`, `DecimalAmount` and `Rate` read `""` as `0`. For
  `money.NullableCentAmount` and `uu.NullableID` it is an error.
- **JSON `null` into a field that holds a value.** `country.NullableCode`,
  `bank.NullableIBAN`, `bank.NullableBIC`, `date.NullableDate`,
  `vat.NullableID` and `money.NullableCurrency` have no `UnmarshalJSON`, so
  `encoding/json` treats `null` as a no-op and the field keeps its value.
  `uu.NullableID`, `email.NullableAddress`, `account.NullableNumber`, the
  `nullable` strings and `nullable.Type[T]` set it to null.
- **`Scan("")` of ID collections.** `uu.IDSlice.Scan("")` gives a nil slice,
  `uu.IDSet.Scan("")` an empty non-nil set.
- **Names.** `money.Currency` converts with `NullableCurrency()`, every other
  `X` with `Nullable()`. `vat.Null` is a variable while the other string null
  values are typed constants, and `money.CurrencyNull` is an untyped constant.
- **`nullable.NullableValidator`** (`ValidAndNotNull`) is implemented by the
  bank, country, date, money and vat nullable types, but not by
  `uu.NullableID`, `account.NullableNumber` or `email.NullableAddress`.

Changing any of these changes stored data or wire formats, so each needs its
own breaking-change decision. They are tracked in [TODOS.md](../TODOS.md).

## Related

- [How to store optional values](howto-store-optional-values.md): pick the Go
  type for a nullable column or optional JSON field.
- [Sets](explanation-sets.md): the set types and their codecs.
- [Validation and normalization](explanation-validation-and-normalization.md):
  why `NullableX.Valid()` accepts null.
- Reference: [nullable](../nullable/README.md), [notnull](../notnull/README.md),
  [uu](../uu/README.md), [money](../money/README.md),
  [date](../date/README.md), [country](../country/README.md),
  [bank](../bank/README.md), [vat](../vat/README.md),
  [email](../email/README.md), [account](../account/README.md),
  [language](../language/README.md).
