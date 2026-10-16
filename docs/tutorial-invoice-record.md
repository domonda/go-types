# Build an invoice record from raw strings

In this tutorial you take the raw strings that a text extraction step found on an invoice and turn them into a typed Go struct. You normalize each value, validate the struct, write it as JSON, look at what it sends to SQL, and split the total into installments.

These are the strings you start from:

| Field         | Raw string                 |
|---------------|----------------------------|
| Vendor VAT ID | `atu 1022 3006`            |
| IBAN          | `AT61 1904 3002 3457 3201` |
| Invoice date  | `25.12.2023`               |
| Due date      | empty                      |
| Total         | `1.234,50`                 |
| Currency      | `€`                        |

## Before you start

You need Go 1.26 or later. Create a module for the tutorial:

```sh
mkdir invoice && cd invoice
go mod init example.com/invoice
go get github.com/domonda/go-types
```

Each step is a complete program. Save it as `main.go`, replacing the previous one, and run it with `go run .`. The `// Output:` comment at the end of each program shows what it prints.

## Step 1: Normalize the VAT ID

The vendor's VAT ID was printed in groups and in lower case. `vat.NormalizeVATID` removes the spaces, upper-cases the letters, and checks the format and the check digit of the country:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/vat"
)

func main() {
	id, err := vat.NormalizeVATID("atu 1022 3006")
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	fmt.Println(id.CountryCode(), id.Number())

	// Output:
	// ATU10223006
	// AT U10223006
}
```

`ATU10223006` is the canonical form: the same ID always gets the same string, so you can store it, compare it and look it up. Change one digit of the input, say `atu 1022 3007`, and the function returns the error `invalid VAT ID: "atu 1022 3007" has an invalid check-sum`.

## Step 2: Normalize the IBAN

`bank.NormalizeIBAN` removes all whitespace and checks the country, the length and the mod-97 check sum:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/bank"
)

func main() {
	iban, err := bank.NormalizeIBAN("AT61 1904 3002 3457 3201")
	if err != nil {
		panic(err)
	}
	fmt.Println(iban)
	fmt.Println(iban.CountryCode())

	// The grouped form is for display only
	grouped, _ := iban.NormalizedWithSpaces()
	fmt.Println(grouped)

	// One wrong digit fails the check sum
	_, err = bank.NormalizeIBAN("AT61 1904 3002 3457 3202")
	fmt.Println(err)

	// Output:
	// AT611904300234573201
	// AT
	// AT61 1904 3002 3457 3201
	// invalid IBAN check sum
}
```

`fmt.Println(grouped)` prints the value as it is stored, because `bank.IBAN` has no normalizing `String` method: what you see is what the variable holds.

## Step 3: Parse the dates

The invoice date is written day first, German style. `date.Normalize` reads it and returns the ISO 8601 form. The due date is empty, which means there is none. An empty `date.Date` is invalid, so use `date.NullableDate` and `date.NormalizeNullable`, which returns `date.Null` for an empty string:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/date"
)

func main() {
	invoiceDate, err := date.Normalize("25.12.2023")
	if err != nil {
		panic(err)
	}
	fmt.Println(invoiceDate, invoiceDate.Weekday())

	dueDate, err := date.NormalizeNullable("")
	if err != nil {
		panic(err)
	}
	fmt.Println(dueDate.IsNull())

	// Output:
	// 2023-12-25 Monday
	// true
}
```

`date.Normalize` also reads `Dec 25, 2023` and `12/25/2023`. A date like `05/06/2024` could be either order: without a language hint it is read month first because of its slashes, while `05.06.2024` with dots is read day first. [Lenient parsing](explanation-lenient-parsing.md#dates) explains the rules.

## Step 4: Parse the total and the currency

`1.234,50` uses a dot to group thousands and a comma for the decimals. `money.ParseDecimalAmount` detects the separators and parses the digits exactly, without going through a `float64`. The `2` says which number of decimals you accept, so a reading with any other number is an error. `money.Currency.Normalized` turns the euro sign into the ISO 4217 code:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/money"
)

func main() {
	total, err := money.ParseDecimalAmount("1.234,50", 2)
	if err != nil {
		panic(err)
	}
	fmt.Println(total, total.Scale())

	currency, err := money.Currency("€").Normalized()
	if err != nil {
		panic(err)
	}
	fmt.Println(currency, currency.MinorUnits())

	// "1.234" has 3 decimals if the dot is a decimal separator
	_, err = money.ParseDecimalAmount("1.234", 2)
	fmt.Println(err)

	// Output:
	// 1234.50 2
	// EUR 2
	// parsing "1.234" returned 3 decimals which is not in the accepted list [2]
}
```

`DecimalAmount` keeps the scale it was parsed with, so the total stays `1234.50` and not `1234.5`. `MinorUnits` reports that the euro has 2 decimals. [Money types and rounding](explanation-money-types-and-rounding.md) explains why there are three amount types and when to use which.

## Step 5: Build and validate the invoice

Now put the values into a struct. `parseInvoice` runs the parsers from steps 1 to 4 and joins their errors, so one call reports every bad field. `types.DeepValidate` checks a struct you got from somewhere else, like a JSON request or a database row, which didn't go through the parsers:

```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/bank"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/vat"
)

type Invoice struct {
	VendorVATID vat.ID              `json:"vendorVatId"`
	IBAN        bank.IBAN           `json:"iban"`
	InvoiceDate date.Date           `json:"invoiceDate"`
	DueDate     date.NullableDate   `json:"dueDate"`
	Total       money.DecimalAmount `json:"total"`
	Currency    money.Currency      `json:"currency"`
}

func parseInvoice(vatID, iban, invoiceDate, dueDate, total, currency string) (Invoice, error) {
	var inv Invoice
	var parseErrs [6]error
	inv.VendorVATID, parseErrs[0] = vat.NormalizeVATID(vatID)
	inv.IBAN, parseErrs[1] = bank.NormalizeIBAN(iban)
	inv.InvoiceDate, parseErrs[2] = date.Normalize(invoiceDate)
	inv.DueDate, parseErrs[3] = date.NormalizeNullable(dueDate)
	inv.Total, parseErrs[4] = money.ParseDecimalAmount(total, 2)
	inv.Currency, parseErrs[5] = money.Currency(currency).Normalized()
	return inv, errors.Join(parseErrs[:]...)
}

func main() {
	inv, err := parseInvoice("atu 1022 3006", "AT61 1904 3002 3457 3201", "25.12.2023", "", "1.234,50", "€")
	if err != nil {
		panic(err)
	}
	fmt.Println(inv.VendorVATID, inv.IBAN, inv.InvoiceDate, inv.DueDate.IsNull(), inv.Total, inv.Currency)
	fmt.Println("errors:", len(types.DeepValidate(inv)))

	// JSON decoding stores the strings as they are, without any check
	var received Invoice
	err = json.Unmarshal([]byte(`{
		"vendorVatId": "ATU10223006",
		"iban": "AT61 1904 3002 3457 3202",
		"invoiceDate": "30.02.2023",
		"total": 1234.50,
		"currency": "EUR"
	}`), &received)
	if err != nil {
		panic(err)
	}
	for _, err := range types.DeepValidate(received) {
		fmt.Println(err)
	}

	// Output:
	// ATU10223006 AT611904300234573201 2023-12-25 true 1234.50 EUR
	// errors: 0
	// struct field IBAN: invalid IBAN check sum
	// struct field InvoiceDate: parsing time "2023-02-30": day out of range
}
```

`DeepValidate` calls the `Validate` or `Valid` method of every exported field it can reach, also in nested structs, slices and maps, and prefixes each error with the path to the field. It returns a slice of errors. Pass it to `errors.Join` if you need a single error.

The validation methods normalize before they check, so `DeepValidate` accepts `atu 1022 3006` as a valid VAT ID. It tells you whether a value is valid, not whether it is in canonical form. That's why `parseInvoice` stores the normalized values. Reading does not do that, except `date.Date` on SQL. JSON and XML only trim, or trim and uppercase. Call `Normalized()` when the canonical form is what you store.

## Step 6: Write the invoice as JSON

```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/domonda/go-types/bank"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/vat"
)

type Invoice struct {
	VendorVATID vat.ID              `json:"vendorVatId"`
	IBAN        bank.IBAN           `json:"iban"`
	InvoiceDate date.Date           `json:"invoiceDate"`
	DueDate     date.NullableDate   `json:"dueDate"`
	Total       money.DecimalAmount `json:"total"`
	Currency    money.Currency      `json:"currency"`
}

func parseInvoice(vatID, iban, invoiceDate, dueDate, total, currency string) (Invoice, error) {
	var inv Invoice
	var parseErrs [6]error
	inv.VendorVATID, parseErrs[0] = vat.NormalizeVATID(vatID)
	inv.IBAN, parseErrs[1] = bank.NormalizeIBAN(iban)
	inv.InvoiceDate, parseErrs[2] = date.Normalize(invoiceDate)
	inv.DueDate, parseErrs[3] = date.NormalizeNullable(dueDate)
	inv.Total, parseErrs[4] = money.ParseDecimalAmount(total, 2)
	inv.Currency, parseErrs[5] = money.Currency(currency).Normalized()
	return inv, errors.Join(parseErrs[:]...)
}

func main() {
	inv, err := parseInvoice("atu 1022 3006", "AT61 1904 3002 3457 3201", "25.12.2023", "", "1.234,50", "€")
	if err != nil {
		panic(err)
	}
	j, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))

	// Decoding gives back the same values, the total with its scale
	var decoded Invoice
	if err := json.Unmarshal(j, &decoded); err != nil {
		panic(err)
	}
	fmt.Println(decoded == inv)

	// Output:
	// {
	//   "vendorVatId": "ATU10223006",
	//   "iban": "AT611904300234573201",
	//   "invoiceDate": "2023-12-25",
	//   "dueDate": null,
	//   "total": 1234.50,
	//   "currency": "EUR"
	// }
	// true
}
```

Three things to notice:

- `dueDate` is `null`, because `date.NullableDate` writes `null` for `date.Null`. A plain `date.Date` would write `""`.
- `total` is the JSON number `1234.50` with both decimals. A `money.Amount` (a `float64`) would write `1234.5`.
- The identifier types trim, and uppercase where that is how they are written, like their `Value` method. `money.Currency("eur")` is written as `"EUR"`. A symbol such as `"€"` stays `"€"`. An invalid value is written after that prettification. `date.Date` JSON only trims. Call `Normalized()` for the canonical form.

## Step 7: See what goes to SQL

`database/sql` calls the `Value` method of each argument that implements `driver.Valuer`. This program prints what each field would send for an insert statement, without a database:

```go
package main

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/domonda/go-types/bank"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/vat"
)

type Invoice struct {
	VendorVATID vat.ID              `json:"vendorVatId"`
	IBAN        bank.IBAN           `json:"iban"`
	InvoiceDate date.Date           `json:"invoiceDate"`
	DueDate     date.NullableDate   `json:"dueDate"`
	Total       money.DecimalAmount `json:"total"`
	Currency    money.Currency      `json:"currency"`
}

func parseInvoice(vatID, iban, invoiceDate, dueDate, total, currency string) (Invoice, error) {
	var inv Invoice
	var parseErrs [6]error
	inv.VendorVATID, parseErrs[0] = vat.NormalizeVATID(vatID)
	inv.IBAN, parseErrs[1] = bank.NormalizeIBAN(iban)
	inv.InvoiceDate, parseErrs[2] = date.Normalize(invoiceDate)
	inv.DueDate, parseErrs[3] = date.NormalizeNullable(dueDate)
	inv.Total, parseErrs[4] = money.ParseDecimalAmount(total, 2)
	inv.Currency, parseErrs[5] = money.Currency(currency).Normalized()
	return inv, errors.Join(parseErrs[:]...)
}

const insertInvoice = /*sql*/ `
	INSERT INTO invoice (vendor_vat_id, iban, invoice_date, due_date, total, currency)
	VALUES ($1, $2, $3, $4, $5, $6)`

func main() {
	inv, err := parseInvoice("atu 1022 3006", "AT61 1904 3002 3457 3201", "25.12.2023", "", "1.234,50", "€")
	if err != nil {
		panic(err)
	}

	// The arguments you would pass to db.ExecContext(ctx, insertInvoice, ...)
	args := []driver.Valuer{inv.VendorVATID, inv.IBAN, inv.InvoiceDate, inv.DueDate, inv.Total, inv.Currency}
	for i, arg := range args {
		v, err := arg.Value()
		if err != nil {
			panic(err)
		}
		fmt.Printf("$%d %-19T %#v\n", i+1, arg, v)
	}

	// Output:
	// $1 vat.ID              "ATU10223006"
	// $2 bank.IBAN           "AT611904300234573201"
	// $3 date.Date           "2023-12-25"
	// $4 date.NullableDate   <nil>
	// $5 money.DecimalAmount "1234.50"
	// $6 money.Currency      "EUR"
}
```

A table that fits these values:

```sql
CREATE TABLE invoice (
    vendor_vat_id text    NOT NULL,
    iban          text    NOT NULL,
    invoice_date  date    NOT NULL,
    due_date      date,
    total         numeric NOT NULL,
    currency      text    NOT NULL
);
```

- The null due date is sent as SQL `NULL`.
- The total is sent as the decimal string `"1234.50"`, so a `numeric` column stores it exactly. A `float64` would be rounded to binary.
- `Value` trims and uppercases. `bank.IBAN("AT61 1904 3002 3457 3201")` is sent with its spaces, and `money.Currency("€")` is sent as `"€"`. An invalid value is written that way. `date.Date.Value` still writes `YYYY-MM-DD` and returns an error when the date cannot be normalized.

## Step 8: Split the total into installments

The vendor allows payment in four monthly installments. Splitting 1234.50 by 4 doesn't give whole cents. `CentAmount.SplitEqually` divides an amount of whole cents so that the parts add up to the total exactly:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/domonda/go-types/bank"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/vat"
)

type Invoice struct {
	VendorVATID vat.ID              `json:"vendorVatId"`
	IBAN        bank.IBAN           `json:"iban"`
	InvoiceDate date.Date           `json:"invoiceDate"`
	DueDate     date.NullableDate   `json:"dueDate"`
	Total       money.DecimalAmount `json:"total"`
	Currency    money.Currency      `json:"currency"`
}

func parseInvoice(vatID, iban, invoiceDate, dueDate, total, currency string) (Invoice, error) {
	var inv Invoice
	var parseErrs [6]error
	inv.VendorVATID, parseErrs[0] = vat.NormalizeVATID(vatID)
	inv.IBAN, parseErrs[1] = bank.NormalizeIBAN(iban)
	inv.InvoiceDate, parseErrs[2] = date.Normalize(invoiceDate)
	inv.DueDate, parseErrs[3] = date.NormalizeNullable(dueDate)
	inv.Total, parseErrs[4] = money.ParseDecimalAmount(total, 2)
	inv.Currency, parseErrs[5] = money.Currency(currency).Normalized()
	return inv, errors.Join(parseErrs[:]...)
}

func main() {
	inv, err := parseInvoice("atu 1022 3006", "AT61 1904 3002 3457 3201", "25.12.2023", "", "1.234,50", "€")
	if err != nil {
		panic(err)
	}

	// The total has 2 decimals, so converting it to cents doesn't round
	cents, err := inv.Total.CentAmount(money.RoundHalfAwayFromZero)
	if err != nil {
		panic(err)
	}

	installments := cents.SplitEqually(4)
	var sum money.CentAmount
	for i, part := range installments {
		fmt.Println(inv.InvoiceDate.AddMonths(i+1), part, inv.Currency)
		sum += part
	}
	fmt.Println("sum:", sum, sum == cents)

	// Output:
	// 2024-01-25 308.62 EUR
	// 2024-02-25 308.62 EUR
	// 2024-03-25 308.63 EUR
	// 2024-04-25 308.63 EUR
	// sum: 1234.50 true
}
```

123450 cents divided by 4 leaves 2 cents, which go to the last two installments. No installment differs from another by more than one cent, and the sum is exactly the total. `DecimalAmount.CentAmount` returns an error for an amount that is `NaN`, infinite or beyond the `CentAmount` range, so such a value can't turn into a wrong installment. [Splits](explanation-money-types-and-rounding.md#splits) explains the method, also for splits by weight.

## What you built

You turned six raw strings into an `Invoice` struct with one canonical value per field:

- [`vat.ID`](../vat/README.md) and [`bank.IBAN`](../bank/README.md) normalized and checked with their check digits.
- [`date.Date`](../date/README.md) parsed from a German date, and `date.NullableDate` for the missing due date.
- [`money.DecimalAmount`](../money/README.md) parsed exactly with its scale, and `money.Currency` from a symbol.
- Validation of untrusted structs with [`types.DeepValidate`](reference-types.md#validation).
- JSON and SQL output that keeps `null`, the decimals of the total and the canonical identifiers.
- Installments that add up to the cent with `money.CentAmount`.

Where to go next:

- [Extract identifiers from text](howto-extract-identifiers-from-text.md): find these values in the full text of an invoice in the first place.
- [Scan and format table rows](howto-scan-and-format-table-rows.md): read and write such records as CSV.
- [Lenient parsing](explanation-lenient-parsing.md): which spellings the parsers accept and how they decide between readings.
- [Money types and rounding](explanation-money-types-and-rounding.md): `Amount`, `DecimalAmount` and `CentAmount` compared.
- [Validation and normalization](explanation-validation-and-normalization.md) and [Null conventions](explanation-null-conventions.md): the patterns behind `Normalized`, `Validate` and the nullable types.
