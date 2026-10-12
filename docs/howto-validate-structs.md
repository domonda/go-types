# How to validate structs

This guide shows how to make your own types validatable, validate a nested
struct in one call with `types.DeepValidate`, and work with the errors it
returns. For the reasoning behind the interfaces, see
[Validation and normalization](explanation-validation-and-normalization.md).

## Prerequisites

- Go 1.26 or later, the minimum version of the module.
- The module in your `go.mod`:

  ```sh
  go get github.com/domonda/go-types
  ```

- The root package imported as `types`:

  ```go
  import "github.com/domonda/go-types"
  ```

## Steps

### 1. Implement `Validator` on types that can only say yes or no

`types.Validator` has one method, `Valid() bool`. Use a value receiver, so
that values, pointers and copies all have the method:

```go
// Status can only say whether it is valid: types.Validator.
type Status string

func (s Status) Valid() bool {
	return s == "open" || s == "paid"
}
```

When `Valid()` returns `false`, the module builds the error for you from the
type and the value: `invalid main.Status value: "unknown"`.

### 2. Implement `ValidatErr` on types that can say why

`types.ValidatErr` has one method, `Validate() error`. Any non-nil error means
invalid, so return an error that describes the problem. Return a custom error
type or a sentinel if callers need to tell problems apart:

```go
// QuantityError tells the caller which quantity is wrong.
type QuantityError struct {
	Quantity int
}

func (e *QuantityError) Error() string {
	return fmt.Sprintf("quantity %d must be positive", e.Quantity)
}

// Line can say why it is invalid: types.ValidatErr.
type Line struct {
	SKU      string
	Quantity int
}

func (l Line) Validate() error {
	if l.Quantity <= 0 {
		return &QuantityError{Quantity: l.Quantity}
	}
	return nil
}
```

If a type implements both interfaces, `Validate()` wins and `Valid()` is not
called.

A pointer receiver (`func (l *Line) Validate() error`) also works with
`types.DeepValidate`, which validates through a pointer when only the pointer
type has the method. `types.Validate(line)` with a value does not find a
pointer receiver method; pass `&line`.

### 3. Build the struct from validatable fields

The domain types of this module already implement both interfaces, so a
`vat.ID` or `money.Currency` field is validated without extra code. Put only
the rules that span several fields into the struct's own `Validate` method:

```go
type Invoice struct {
	Status   Status
	VATID    vat.ID
	Currency money.Currency
	Lines    []Line
	Note     *string // nil pointers are skipped
}

// Validate checks only the rules that span several fields.
// DeepValidate validates the fields itself.
func (inv Invoice) Validate() error {
	if inv.Status == "paid" && len(inv.Lines) == 0 {
		return errors.New("a paid invoice needs at least one line")
	}
	return nil
}
```

### 4. Call `types.DeepValidate`

`types.DeepValidate` returns a `[]error` with one entry per invalid value it
found, or nil if everything is valid:

```go
errs := types.DeepValidate(&inv)
for _, err := range errs {
	fmt.Println(err)
}
```

It validates the value itself, follows pointers and interfaces, and descends
into exported struct fields, slice and array elements, and map values.

### 5. Read the path prefixes

Every error from a nested value starts with the path to it:

```
struct field Status: invalid main.Status value: "unknown"
struct field VATID: invalid VAT ID: "ATU12345678" has an invalid check-sum
struct field Lines -> element [1]: quantity 0 must be positive
```

The path elements are `struct field <Name>`, `element [<index>]` and
`map value [<key>]`, joined with ` -> `. An error of the top-level value
itself, like the one from `Invoice.Validate`, has no prefix.

### 6. Join the errors into one

Use `errors.Join` when an API expects a single `error`. It returns nil for an
empty slice, so it doubles as the validity check:

```go
err := errors.Join(errs...)
```

### 7. Find a specific error with `errors.As` or `errors.Is`

The path prefix wraps the original error with `%w`, and `errors.Join` keeps
every error, so both functions see through to the error your type returned:

```go
var qe *QuantityError
if errors.As(err, &qe) {
	fmt.Println("bad quantity:", qe.Quantity)
}
fmt.Println("invalid VAT ID:", errors.Is(err, vat.ErrInvalidID))
```

### 8. Combine validators

The root package has helpers to combine checks without writing a new type:

| Helper                                             | Result                                             |
|----------------------------------------------------|----------------------------------------------------|
| `types.Validators{a, b}`                           | a `Validator` that is valid if all of them are     |
| `types.CombinedValidator(a, b)`                    | the same as `types.Validators{a, b}`               |
| `types.ValidatErrs{a, b}`                          | a `ValidatErr` that calls all of them and joins their errors |
| `types.CombinedValidatErr(a, b)`                   | the same as `types.ValidatErrs{a, b}`              |
| `types.ValidatorAsValidatErr{Validator: v}`        | a `ValidatErr` with the error `types.Validate` would build |
| `types.ValidatorFunc(f)`, `types.ValidatErrFunc(f)` | a function as validator                            |
| `types.StaticValidator(b)`, `types.StaticValidatErr{Err: err}` | a precomputed result                               |

`types.ValidatErrs` calls every element, also after one returned an error,
and joins all errors with `errors.Join`:

```go
check := types.CombinedValidatErr(
	inv.VATID,
	types.ValidatorAsValidatErr{Validator: inv.Status},
)
fmt.Println(check.Validate())
fmt.Println(types.Validators{inv.Status, inv.Currency}.Valid())
```

### The complete program

```go
package main

import (
	"errors"
	"fmt"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/vat"
)

// Status can only say whether it is valid: types.Validator.
type Status string

func (s Status) Valid() bool {
	return s == "open" || s == "paid"
}

// QuantityError tells the caller which quantity is wrong.
type QuantityError struct {
	Quantity int
}

func (e *QuantityError) Error() string {
	return fmt.Sprintf("quantity %d must be positive", e.Quantity)
}

// Line can say why it is invalid: types.ValidatErr.
type Line struct {
	SKU      string
	Quantity int
}

func (l Line) Validate() error {
	if l.Quantity <= 0 {
		return &QuantityError{Quantity: l.Quantity}
	}
	return nil
}

type Invoice struct {
	Status   Status
	VATID    vat.ID
	Currency money.Currency
	Lines    []Line
	Note     *string // nil pointers are skipped
}

// Validate checks only the rules that span several fields.
// DeepValidate validates the fields itself.
func (inv Invoice) Validate() error {
	if inv.Status == "paid" && len(inv.Lines) == 0 {
		return errors.New("a paid invoice needs at least one line")
	}
	return nil
}

func main() {
	inv := Invoice{
		Status:   "unknown",
		VATID:    "ATU12345678",
		Currency: "eur",
		Lines: []Line{
			{SKU: "A-1", Quantity: 2},
			{SKU: "B-2", Quantity: 0},
		},
	}

	errs := types.DeepValidate(&inv)
	for _, err := range errs {
		fmt.Println(err)
	}

	err := errors.Join(errs...)
	var qe *QuantityError
	if errors.As(err, &qe) {
		fmt.Println("bad quantity:", qe.Quantity)
	}
	fmt.Println("invalid VAT ID:", errors.Is(err, vat.ErrInvalidID))

	// An error of the top-level value has no path prefix
	paid := Invoice{Status: "paid", VATID: "ATU10223006", Currency: "EUR"}
	fmt.Println(types.DeepValidate(paid))

	check := types.CombinedValidatErr(
		inv.VATID,
		types.ValidatorAsValidatErr{Validator: inv.Status},
	)
	fmt.Println(check.Validate())
	fmt.Println(types.Validators{inv.Status, inv.Currency}.Valid())

	// Output:
	// struct field Status: invalid main.Status value: "unknown"
	// struct field VATID: invalid VAT ID: "ATU12345678" has an invalid check-sum
	// struct field Lines -> element [1]: quantity 0 must be positive
	// bad quantity: 0
	// invalid VAT ID: true
	// [a paid invoice needs at least one line]
	// invalid VAT ID: "ATU12345678" has an invalid check-sum
	// invalid main.Status value: "unknown"
	// false
}
```

## Verification

Save the program as `main.go` in a module that requires
`github.com/domonda/go-types` and run it:

```sh
go run .
```

The output matches the `// Output:` block. Three things in it confirm that
validation works as intended:

- `Currency: "eur"` produced no error: `money.Currency` validation accepts
  every spelling that it can normalize.
- `Note` is nil and produced no error: nil pointers are skipped.
- The paid invoice without lines reports one error without a path prefix:
  `Invoice.Validate` checks only its cross-field rule, so no field problem
  is reported twice.

## Troubleshooting

**The error only says `invalid main.Status value: "unknown"`.**
The type implements only `Validator`, whose `Valid()` can't say why. Add a
`Validate() error` method; it takes precedence over `Valid()`.

**The error contains an address, like `invalid *main.Status value:
(*main.Status)(0x...)`.**
A pointer to a `Validator`-only type was validated, and `%#v` of a pointer is
its address. Implement `ValidatErr`, or pass the value instead of the pointer.
The same `%#v` puts the value into the message, so don't rely on `Valid()`
for types that hold secrets.

**`Validate` is never called.**
Check, in this order:

- The method has a pointer receiver and you called `types.Validate(v)` with a
  value. `types.TryValidate(v)` then returns `nil, false`. Pass `&v`, or use
  `types.DeepValidate`, which finds pointer receivers.
- The field is unexported. `DeepValidate` skips unexported fields, except the
  exported fields promoted from an unexported embedded struct.
- The value is a map key. Only map values are validated.
- The value sits behind a nil pointer, nil interface or nil func. There is
  nothing to validate, so `DeepValidate` skips it.

**The same problem is reported twice.**
The struct's own `Validate` checks its fields, and `DeepValidate` validates
the fields again. One error comes without a path prefix, one with. Keep only
cross-field rules in the struct's `Validate`, as in step 3.

**`err == ErrSomething` is false although the error is in the output.**
`DeepValidate` wraps every error with its path, and `types.ValidatErrs` and
`types.CombinedValidatErr` wrap even a single error with `errors.Join`.
Compare with `errors.Is(err, ErrSomething)`.

**A `language.Code` field with the value `"EN"` passes, but `Valid()` is
false.**
`Validate()` normalizes first and accepts `"EN"`, `"eng"` and `"en-US"`, while
`Valid()` is a strict lookup. Validation does not change the value: store the
result of `Normalized()` if you need `"en"`.

**A `bank.BIC` with spaces fails with `invalid BIC "BKAU AT WW" length: 10`.**
`bank.BIC.Validate` does not remove whitespace, `bank.BIC.Normalized` does.
Normalize the BIC before validating it.

## Related

- [Validation and normalization](explanation-validation-and-normalization.md):
  how `Valid`, `Validate` and `Normalized` relate, and the full traversal rules
  of `DeepValidate`.
- [Root package reference](../README.md)
