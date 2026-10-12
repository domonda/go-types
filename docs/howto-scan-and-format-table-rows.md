# How to scan and format table rows

This guide shows how to export a slice of structs with `go-types` fields to a localized CSV file with [`strfmt`](../strfmt/README.md), and how to scan the cells of a CSV file or spreadsheet back into typed struct fields.

## Prerequisites

- Go 1.26 or later and the module: `go get github.com/domonda/go-types`.
- Rows as `[]string` per record, for example from `encoding/csv` or a spreadsheet library.
- A struct with one exported field per column, in column order.

## Steps

The steps build one program. Step 6 shows it complete, with its output.

1. Define the row type. Use the `go-types` type that matches each column, and a nullable type for optional columns:

   ```go
   type Payment struct {
   	Payee   string
   	Amount  money.Amount
   	Booked  date.Date
   	Paid    bool
   	DueDate date.NullableDate
   	Note    nullable.TrimmedString
   }
   ```

2. Pick a format preset and a placeholder for null values. `strfmt.NewEnglishFormatConfig()` and `strfmt.NewGermanFormatConfig()` set the money separators, the date layout and the bool labels. `Nil` is the text written for nil pointers, null values of nullable types, a zero `time.Time` and an empty or null `date.Date` or `date.NullableDate`, and it scans back as null:

   | Preset                     | Money      | Date         | Bool           | Float    |
   |----------------------------|------------|--------------|----------------|----------|
   | `NewFormatConfig()`        | `1,234.50` | `2023-12-25` | `true`/`false` | `1234.5` |
   | `NewEnglishFormatConfig()` | `1,234.50` | `25/12/2023` | `yes`/`no`     | `1234.5` |
   | `NewGermanFormatConfig()`  | `1.234,50` | `25.12.2023` | `ja`/`nein`    | `1234,5` |

   ```go
   formatConfig := strfmt.NewGermanFormatConfig()
   formatConfig.Nil = "-"
   ```

3. Format each field of a row with `strfmt.Format`:

   ```go
   // formatRow formats every field of the struct row with config.
   func formatRow(row any, config *strfmt.FormatConfig) []string {
   	v := reflect.ValueOf(row)
   	cells := make([]string, v.NumField())
   	for i := range cells {
   		cells[i] = strfmt.Format(v.Field(i).Interface(), config)
   	}
   	return cells
   }
   ```

   Write the cells with `encoding/csv`. Use `;` as field separator for the German preset, because its decimal separator is the comma.

4. Configure scanning. `strfmt.NewScanConfig()` accepts the bool labels and time layouts of all three presets. Add your `Nil` placeholder to `NilStrings`, and turn on `StrictEmptyStringParsing`, so an empty cell in a column that can't be null is an error instead of a zero value:

   ```go
   scanConfig := strfmt.NewScanConfig()
   scanConfig.NilStrings = append(scanConfig.NilStrings, "-")
   scanConfig.StrictEmptyStringParsing = true
   ```

   `NewScanConfig` sets `ValidateFunc` to `types.Validate`, so a cell that parses but is invalid, like the date `30.02.2023`, is an error too.

5. Scan each cell into its field with `strfmt.Scan`. It takes the field as a settable `reflect.Value`, so pass a pointer to the row:

   ```go
   // scanRow scans cells into the fields of the struct that dest points to
   // and returns the name of the field that failed.
   func scanRow(dest any, cells []string, config *strfmt.ScanConfig) (string, error) {
   	v := reflect.ValueOf(dest).Elem()
   	for i, cell := range cells {
   		if err := strfmt.Scan(v.Field(i), cell, config); err != nil {
   			return v.Type().Field(i).Name, err
   		}
   	}
   	return "", nil
   }
   ```

   A failed `Scan` leaves the field unchanged. Its error carries a call stack and the arguments of the call. `errs.UnwrapCallStack(err)` from `github.com/domonda/go-errs` returns the error without them, for display.

6. Put it together. This program writes the payments with both presets, reads each file back and compares the result with the original rows:

   ```go
   package main

   import (
   	"bytes"
   	"encoding/csv"
   	"fmt"
   	"reflect"
   	"slices"

   	"github.com/domonda/go-errs"
   	"github.com/domonda/go-types/date"
   	"github.com/domonda/go-types/money"
   	"github.com/domonda/go-types/nullable"
   	"github.com/domonda/go-types/strfmt"
   )

   type Payment struct {
   	Payee   string
   	Amount  money.Amount
   	Booked  date.Date
   	Paid    bool
   	DueDate date.NullableDate
   	Note    nullable.TrimmedString
   }

   // formatRow formats every field of the struct row with config.
   func formatRow(row any, config *strfmt.FormatConfig) []string {
   	v := reflect.ValueOf(row)
   	cells := make([]string, v.NumField())
   	for i := range cells {
   		cells[i] = strfmt.Format(v.Field(i).Interface(), config)
   	}
   	return cells
   }

   // scanRow scans cells into the fields of the struct that dest points to
   // and returns the name of the field that failed.
   func scanRow(dest any, cells []string, config *strfmt.ScanConfig) (string, error) {
   	v := reflect.ValueOf(dest).Elem()
   	for i, cell := range cells {
   		if err := strfmt.Scan(v.Field(i), cell, config); err != nil {
   			return v.Type().Field(i).Name, err
   		}
   	}
   	return "", nil
   }

   func writeCSV(payments []Payment, config *strfmt.FormatConfig, comma rune) []byte {
   	var buf bytes.Buffer
   	w := csv.NewWriter(&buf)
   	w.Comma = comma
   	for _, p := range payments {
   		if err := w.Write(formatRow(p, config)); err != nil {
   			panic(err)
   		}
   	}
   	w.Flush()
   	return buf.Bytes()
   }

   func readCSV(data []byte, config *strfmt.ScanConfig, comma rune) []Payment {
   	r := csv.NewReader(bytes.NewReader(data))
   	r.Comma = comma
   	r.FieldsPerRecord = reflect.TypeFor[Payment]().NumField()
   	records, err := r.ReadAll()
   	if err != nil {
   		panic(err)
   	}
   	var payments []Payment
   	for i, record := range records {
   		var p Payment
   		field, err := scanRow(&p, record, config)
   		if err != nil {
   			fmt.Printf("row %d, %s: %v\n", i+1, field, errs.UnwrapCallStack(err))
   			continue
   		}
   		payments = append(payments, p)
   	}
   	return payments
   }

   func main() {
   	payments := []Payment{
   		{Payee: "Muster GmbH", Amount: 1234.5, Booked: "2023-12-25", Paid: true, DueDate: date.Null},
   		{Payee: "ACME Ltd.", Amount: 99, Booked: "2024-01-08", DueDate: "2024-02-01", Note: "second reminder"},
   	}

   	scanConfig := strfmt.NewScanConfig()
   	scanConfig.NilStrings = append(scanConfig.NilStrings, "-")
   	scanConfig.StrictEmptyStringParsing = true

   	english := strfmt.NewEnglishFormatConfig()
   	english.Nil = "-"
   	german := strfmt.NewGermanFormatConfig()
   	german.Nil = "-"

   	for _, f := range []struct {
   		config *strfmt.FormatConfig
   		comma  rune
   	}{{english, ','}, {german, ';'}} {
   		data := writeCSV(payments, f.config, f.comma)
   		fmt.Print(string(data))
   		scanned := readCSV(data, scanConfig, f.comma)
   		fmt.Println("round trip equal:", slices.Equal(payments, scanned))
   	}

   	// A row with an invalid date and a row without an amount
   	readCSV([]byte("Muster GmbH;1.234,50;30.02.2023;ja;;-\nACME Ltd.;;08.01.2024;nein;;-\n"), scanConfig, ';')

   	// Output:
   	// Muster GmbH,"1,234.50",25/12/2023,yes,-,-
   	// ACME Ltd.,99.00,08/01/2024,no,01/02/2024,second reminder
   	// round trip equal: true
   	// Muster GmbH;1.234,50;25.12.2023;ja;-;-
   	// ACME Ltd.;99,00;08.01.2024;nein;01.02.2024;second reminder
   	// round trip equal: true
   	// row 1, Booked: parsing time "2023-02-30": day out of range
   	// row 2, Amount: empty string can't be parsed as float
   }
   ```

## Verification

- Both files print `round trip equal: true`: every field was written and read back to the same Go value.
- The amounts have the separators of the preset, `1,234.50` and `1.234,50`. The CSV writer quotes the English amount because it contains the field separator.
- The null `DueDate` and the null `Note` are written as the placeholder `-`, which is in `NilStrings`, so both scan back as null.
- The two broken rows report the column and the reason, and are left out of the result.

## Troubleshooting

The messages below are the errors `strfmt.Scan` returns, after `errs.UnwrapCallStack`.

**`can't scan nil string "" into non-optional destination type bool with strict empty string parsing`.** The cell is empty or a `NilStrings` entry, and the field can't hold null. Make the field a pointer, like `*bool`, which scans an empty cell as nil. Or keep `StrictEmptyStringParsing` false, which scans it as the zero value.

**`empty string can't be parsed as float`.** The same case for a `money.Amount` field: with strict parsing, `Amount.ScanString` gets the empty cell and fails. Use `*money.Amount` for an optional amount.

**`parsing time "2023-02-30": day out of range`.** The date has a valid shape but doesn't exist. With `ValidateFunc` set to nil, `strfmt.Scan` passes `validate = false` to `ScanString`, and `date.Date` stores such a cell unchanged instead of failing. Keep the default `ValidateFunc` when you scan untrusted files.

**`can't scan "maybe" as bool`.** The cell is not in `TrueStrings` or `FalseStrings`. Append the labels your files use, like `"x"`.

**`can't scan "1.000" as int because strconv.ParseInt: parsing "1.000": invalid syntax`.** Integer fields are parsed as plain base 10 numbers, without separators. Use a `float64` or money field for numbers with thousands separators.

**A German amount `1.000` scans as 1, or a US date `05/06/2024` scans as 2024-06-05, without an error.** A single separator is a decimal separator, see [The "1,000" ambiguity](explanation-lenient-parsing.md#the-1000-ambiguity), and `strfmt.Scan` doesn't apply `AcceptedMoneyAmountDecimals`. `date.Date.ScanString` has no language hint, so ambiguous dates are read day first. For both, register a scanner for the type. A registered scanner owns its type completely, also for empty cells, so it has to handle those itself:

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/language"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/strfmt"
)

func main() {
	scanConfig := strfmt.NewScanConfig()

	// Read dates month first
	scanConfig.SetTypeScanner(reflect.TypeFor[date.Date](), strfmt.ScannerFunc(
		func(dest reflect.Value, str string, config *strfmt.ScanConfig) error {
			d, err := date.Normalize(str, language.EN)
			if err != nil {
				return err
			}
			dest.Set(reflect.ValueOf(d))
			return nil
		},
	))

	// Reject amounts with 3 decimals like "1.000"
	scanConfig.SetTypeScanner(reflect.TypeFor[money.Amount](), strfmt.ScannerFunc(
		func(dest reflect.Value, str string, config *strfmt.ScanConfig) error {
			a, err := money.ParseAmount(str, config.AcceptedMoneyAmountDecimals...)
			if err != nil {
				return err
			}
			dest.Set(reflect.ValueOf(a))
			return nil
		},
	))

	var d date.Date
	err := strfmt.Scan(reflect.ValueOf(&d).Elem(), "05/06/2024", scanConfig)
	fmt.Println(d, err)

	var a money.Amount
	err = strfmt.Scan(reflect.ValueOf(&a).Elem(), "1.000", scanConfig)
	fmt.Println(a, errs.UnwrapCallStack(err))

	// Output:
	// 2024-05-06 <nil>
	// 0.00 parsing "1.000" returned 3 decimals which is not in accepted list of [0 2 4]
}
```

Register `date.NullableDate` separately if you have such columns. A scanner registered for a type is also used for pointers to it.

## Related

- [`strfmt`](../strfmt/README.md): `Scan`, `Format`, `ScanConfig`, `FormatConfig` and the presets.
- [`money`](../money/README.md), [`date`](../date/README.md), [`nullable`](../nullable/README.md): the field types and their `ScanString` methods.
- [Lenient parsing](explanation-lenient-parsing.md): how cells like `1.234,50` and `25.12.2023` are read.
- [Null conventions](explanation-null-conventions.md): nullable types and empty cells.
- [Store optional values](howto-store-optional-values.md): choosing a type for optional columns.
