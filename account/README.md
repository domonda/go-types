# account

Account number types with validation, numeric conversion, and SQL, JSON, XML, and text integration. Use `Number` for required account numbers and `NullableNumber` for optional ones. By default an account number is alphanumeric, may contain `_ - / : . ; ,`, and must start with `[0-9A-Za-z]`. The format is configurable with `NumberRegexp`.

```
import "github.com/domonda/go-types/account"
```

## Types

| Type             | Purpose                                            |
|------------------|----------------------------------------------------|
| `Number`         | Required account number, a `string` underneath. The empty string is invalid. |
| `NullableNumber` | Optional account number. The empty string `NumberNull` is the null value and valid. |

Both implement `fmt.Stringer`, `driver.Valuer`, `sql.Scanner`, `json.Marshaler`, `json.Unmarshaler`, `xml.Marshaler`, `xml.Unmarshaler`, `encoding.TextUnmarshaler`, and `JSONSchema()`. `NullableNumber` also implements `nullable.NullSetable[Number]`.

## Constants, variables, and errors

```go
const DefaultNumberRegex = `^[0-9A-Za-z][0-9A-Za-z_\-\/:.;,]*$`

var NumberRegexp = regexp.MustCompile(DefaultNumberRegex)

const NumberNull NullableNumber = ""

const (
    ErrInvalidNumber      errs.Sentinel = "invalid account number"
    ErrAlphanumericNumber errs.Sentinel = "account number is alphanumeric"
)
```

`NumberRegexp` is the compiled pattern that `Valid` and `Validate` check, and the pattern that `JSONSchema` publishes. Reassign it once at startup to enforce an application specific format. Reassigning is not safe for concurrent use, so do it before any validation runs.

Errors wrap the sentinels with `%w`, so `errors.Is(err, account.ErrInvalidNumber)` works.

## Constructors

| Function                                          | Description                                        |
|---------------------------------------------------|----------------------------------------------------|
| `NumberFrom(str) (Number, error)`                 | Trims whitespace and validates. `""` is an error.  |
| `NumberFromUint(u uint64) Number`                 | Decimal string, `0` gives `"0"`.                   |
| `NullableNumberFrom(str) (NullableNumber, error)` | Trims whitespace and validates. `""` and whitespace give `NumberNull`. |
| `NullableNumberFromUint(u uint64) NullableNumber` | Decimal string, `0` gives `NumberNull`.            |

## Methods

Both types have these methods. On `NullableNumber` they return `NullableNumber` where `Number` returns `Number`.

| Method                                             | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `Valid() bool`                                     | Matches `NumberRegexp`, or is null for `NullableNumber`. Does not trim. |
| `Validate() error`                                 | Wrapped `ErrInvalidNumber` if not `Valid`.         |
| `IsNumeric() bool`                                 | Only ASCII digits `0-9`. False for `""`.           |
| `ValidateNumeric() error`                          | Wrapped `ErrAlphanumericNumber` if not `IsNumeric`. |
| `Uint() (uint64, error)`                           | Wrapped `ErrAlphanumericNumber` if not numeric, a `strconv` range error above `math.MaxUint64`. |
| `UintPtr() (*uint64, error)`                       | Like `Uint`, nil on error.                         |
| `Int() (int64, error)`                             | Wrapped `ErrAlphanumericNumber` if not numeric, so `-1` and `+1` are errors, a `strconv` range error above `math.MaxInt64`. |
| `HasPrefix(prefix) bool`, `HasSuffix(suffix) bool` | `strings.HasPrefix`, `strings.HasSuffix`.          |
| `Cut(sep) (before, after, found)`                  | `strings.Cut` with typed halves.                   |
| `TrimLeadingZeros()`                               | Removes leading `'0'` characters, `"000"` gives `""`. |
| `String() string`                                  | The string unchanged, `""` for null.               |

`Number` only:

| Method                      | Description      |
|-----------------------------|------------------|
| `Nullable() NullableNumber` | Type conversion. |

`NullableNumber` only:

| Method                               | Description                                  |
|--------------------------------------|----------------------------------------------|
| `IsNull() bool`, `IsNotNull() bool`  | Whether the trimmed number is empty.         |
| `Get() Number`                       | Panics if null.                              |
| `GetOr(defaultNumber Number) Number` | Value or the default.                        |
| `Set(num Number)`, `SetNull()`       | `Set("")` sets null.                         |
| `Uint()`, `UintPtr()`, `Int()`       | `0, nil`, `nil, nil`, and `0, nil` for null. |

## Serialization

Decoding trims whitespace and validates strings with `NumberRegexp`. Numbers (SQL `int64` and `float64`, JSON numbers) only have to be non-negative integers and are formatted as decimal string without a `NumberRegexp` check.

| Value                               | `Value()`    | JSON      | XML element text |
|-------------------------------------|--------------|-----------|------------------|
| `Number("12-34")`                   | `"12-34"`    | `"12-34"` | `12-34`          |
| `Number(" 12-34 ")`                 | `"12-34"`    | `"12-34"` | `12-34`          |
| `Number("")`, `Number(" ")`         | `nil` (NULL) | `null`    | empty            |
| `NullableNumber("12-34")`           | `"12-34"`    | `"12-34"` | `12-34`          |
| `NumberNull`, `NullableNumber(" ")` | `nil` (NULL) | `null`    | empty            |

Writing trims like `strutil.TrimSpace` (whitespace, control and other non-printable runes, invalid UTF-8), and a value that is empty after trimming becomes SQL NULL, JSON `null`, or an empty XML element, for both types.

| Input                    | `Number`                                           | `NullableNumber`                                 |
|--------------------------|----------------------------------------------------|--------------------------------------------------|
| SQL NULL                 | Error                                              | `NumberNull`                                     |
| SQL `string` or `[]byte` | Trimmed and validated, `""` is an error            | Trimmed and validated, `""` is `NumberNull`      |
| SQL `int64`              | Decimal string, negative is an error               | Like `Number`, `0` is `NumberNull`               |
| SQL `float64`            | Non-negative integer only, `-0` gives `"0"`, NaN, infinities, and fractions are errors | Like `Number`, `0` and `-0` are `NumberNull`     |
| Other SQL types          | Error                                              | Error                                            |
| JSON string              | Trimmed and validated, `""` is an error            | Trimmed and validated, `""` is `NumberNull`      |
| JSON number              | Non-negative integer as decimal string, else error | Like `Number`, but `0` is `NumberNull`           |
| JSON `null`              | Error                                              | `NumberNull`                                     |
| XML element text         | Trimmed and validated, empty is an error           | Trimmed and validated, empty is `NumberNull`     |
| `UnmarshalText`          | Trimmed and validated, `""` is an error            | Trimmed and validated, `""` is `NumberNull`      |

Neither type has a `MarshalText` method. `JSONSchema` describes `Number` as a string with the `NumberRegexp` pattern and `NullableNumber` as that string or null.

## Gotchas

- **A `Number` can be written but not read back empty.** `Number("").Value()` writes NULL and `json.Marshal(Number(""))` writes `null`, but `Scan(nil)`, `Scan("")`, JSON `null`, and JSON `""` are errors for `Number`. Use `NullableNumber` for columns and fields that can be empty.
- **Zero is null for `NullableNumber`.** `NullableNumberFromUint(0)` and `NullableNumber.Scan(int64(0))` give `NumberNull`, while `NumberFromUint(0)` and `Number.Scan(int64(0))` give `"0"`. The JSON number `0` gives `NumberNull` too, while `Number` reads it as `"0"`.
- **`Valid` does not trim, writing does.** `Number(" 123").Valid()` is false, but `Value()`, `MarshalJSON`, and `MarshalXML` write `"123"`. Construct values with `NumberFrom` or `NullableNumberFrom` to get trimmed, validated values.
- **`Int` never returns a negative number.** The digit check rejects `-1` and `+1` with `ErrAlphanumericNumber`.
- **`TrimLeadingZeros` can produce an invalid `Number`.** `Number("000").TrimLeadingZeros()` is `""`, which `Validate` rejects.
- **`NumberRegexp` is a package variable.** Reassigning it changes `Valid`, `Validate`, the decoding of strings, and `JSONSchema` for the whole program.

## Examples

Validate and convert account numbers:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/domonda/go-types/account"
)

func main() {
	n, err := account.NumberFrom(" AT-12345/67 ")
	if err != nil {
		panic(err)
	}
	fmt.Println(n, n.IsNumeric())

	u, err := account.Number("00123").Uint()
	fmt.Println(u, err)

	_, err = account.Number("-1").Int()
	fmt.Println(errors.Is(err, account.ErrAlphanumericNumber))

	_, err = account.NumberFrom("#1")
	fmt.Println(errors.Is(err, account.ErrInvalidNumber))

	// Output:
	// AT-12345/67 false
	// 123 <nil>
	// true
	// true
}
```

Store optional account numbers with `NullableNumber`:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/account"
)

type Payment struct {
	Account   account.Number         `json:"account"`
	Reference account.NullableNumber `json:"reference"`
}

func main() {
	j, err := json.Marshal(Payment{Account: "12345"})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))

	var p Payment
	err = json.Unmarshal(j, &p)
	fmt.Println(p.Reference.IsNull(), err)

	value, _ := p.Reference.Value()
	fmt.Println(value)

	var scanned account.NullableNumber
	err = scanned.Scan(int64(0))
	fmt.Println(scanned.IsNull(), err)

	// Output:
	// {"account":"12345","reference":null}
	// true <nil>
	// <nil>
	// true <nil>
}
```

Enforce an application specific format at startup:

```go
package main

import (
	"fmt"
	"regexp"

	"github.com/domonda/go-types/account"
)

func init() {
	// Only 6 to 10 digits, set once before any validation runs
	account.NumberRegexp = regexp.MustCompile(`^[0-9]{6,10}$`)
}

func main() {
	fmt.Println(account.Number("123456").Valid())
	fmt.Println(account.Number("AT-12345/67").Valid())
	fmt.Println(account.Number("").JSONSchema().Pattern)

	// Output:
	// true
	// false
	// ^[0-9]{6,10}$
}
```

## Related

- [`bank`](../bank/README.md): IBAN and BIC types for bank accounts.
- [`nullable`](../nullable/README.md): the `nullable.NullSetable` interface implemented by `NullableNumber`.
- [Null conventions](../docs/explanation-null-conventions.md): why the empty string is the null value.
- [Validation and normalization](../docs/explanation-validation-and-normalization.md): the `Valid`/`Validate` pattern.
- [Store optional values](../docs/howto-store-optional-values.md): when to use `NullableNumber` instead of `Number`.
