# bank

Banking identifiers: IBAN and BIC (SWIFT code) with validation and normalization, nullable variants for optional columns, an `Account` that combines them with currency and holder, finders that extract IBANs and BICs from text, and XML structs for ISO 20022 CAMT.053 bank statements.

```
import "github.com/domonda/go-types/bank"
```

## IBAN

```go
const (
    IBANRegex     = `^([A-Z]{2})(\d{2})([A-Z\d]{8,30})$`
    IBANMinLength = 15 // Norway
    IBANMaxLength = 32 // Saint Lucia, ISO 13616 allows 34
)

type IBAN string
```

`Normalized` checks an IBAN in this order and returns the input unchanged together with the first error:

1. Remove all Unicode whitespace: spaces, tabs, line breaks, no-break and zero width spaces, also leading ones.
2. `empty IBAN` if nothing is left, `IBAN too short` below `IBANMinLength`.
3. `invalid IBAN country code` if the first two characters are not an uppercase country code of the IBAN length table.
4. `wrong IBAN length` if the length differs from the length of that country.
5. `invalid IBAN characters` if `IBANRegex` doesn't match.
6. `invalid IBAN check sum` if the ISO 7064 mod-97 check fails.

There is no uppercasing and no removal of dashes or other punctuation.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NormalizeIBAN(str) (IBAN, error)`                 | `IBAN(str).Normalized()`.                          |
| `StringIsIBAN(str) bool`                           | `NormalizeIBAN` succeeds, so whitespace is accepted. |
| `iban.Normalized() (IBAN, error)`                  | Compact form without whitespace, see the rules above. |
| `iban.NormalizedWithSpaces() (IBAN, error)`        | Normalized in groups of 4 characters: `DE89 3704 0044 0532 0130 00`. |
| `iban.NormalizedOrNull() NullableIBAN`             | Normalized, or `IBANNull` if invalid.              |
| `iban.Valid() bool`, `iban.Validate() error`       | `Normalized` succeeds.                             |
| `iban.ValidAndNormalized() bool`                   | Valid and already in compact form.                 |
| `iban.CountryCode() country.Code`                  | Country of the normalized IBAN, `country.Invalid` if invalid. |
| `iban.BankAndAccountNumbers() (bankNo, accountNo string, err error)` | Bank and account number of a normalized AT (5 + 11 digits), CH (5 + 12), or DE (8 + 10) IBAN. Error for invalid IBANs and other countries. |
| `iban.String() string`                             | Normalized if valid, else unchanged.               |
| `iban.Nullable() NullableIBAN`                     | Type conversion.                                   |
| `iban.ScanString(source, validate) error`          | Assigns the normalized source. An invalid source is an error if `validate` is true, else assigned unchanged. |
| `iban.Scan(value) error`, `iban.Value()`           | See [Serialization](#serialization).               |
| `IBAN.JSONSchema()`                                | String with the `IBANRegex` pattern.               |

Implements: `sql.Scanner`, `driver.Valuer`, `fmt.Stringer`, `strfmt.Scannable`, `types.NormalizableValidator[IBAN]`. No JSON or text methods.

## NullableIBAN

```go
type NullableIBAN string
const IBANNull NullableIBAN = ""
```

An `IBAN` where `""` is the null value. Null counts as valid.

| Method                                          | Description                                        |
|-------------------------------------------------|----------------------------------------------------|
| `Normalized() (NullableIBAN, error)`            | Like `IBAN.Normalized`, null stays null without error. |
| `NormalizedWithSpaces() (NullableIBAN, error)`  | Like `IBAN.NormalizedWithSpaces`, null stays null. |
| `NormalizedOrNull() NullableIBAN`               | Normalized, or `IBANNull` if null or invalid.      |
| `Valid() bool`, `Validate() error`              | Null or a valid IBAN.                              |
| `ValidAndNotNull() bool`                        | Not null and valid.                                |
| `ValidAndNormalized() bool`                     | Null, or valid and in compact form.                |
| `CountryCode() country.Code`                    | First two characters of a valid value, `""` for null or invalid values. With leading whitespace this is not a country code, normalize first. |
| `IsNull() bool`, `IsNotNull() bool`             | `iban == ""`, `iban != ""`.                        |
| `Get() IBAN`                                    | Panics if null.                                    |
| `GetOr(defaultIBAN IBAN) IBAN`                  | Value or the default.                              |
| `StringOr(defaultString) string`                | Value unchanged, or the default if null.           |
| `String() string`                               | Normalized if valid, else unchanged, `""` for null. |
| `Set(IBAN)`, `SetNull()`                        | `Set("")` sets null.                               |
| `ScanString(source, validate) error`            | `""`, `NULL`, `null`, and `nil` set null, other values like `IBAN.ScanString`. |
| `Scan(value) error`, `Value()`, `MarshalJSON()` | See [Serialization](#serialization).               |
| `JSONSchema()`                                  | `IBANRegex` string or null.                        |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `fmt.Stringer`, `strfmt.Scannable`, `nullable.NullSetable[IBAN]`, `types.NormalizableValidator[NullableIBAN]`. No `json.Unmarshaler`.

## BIC

```go
const (
    BICRegex     = `^([A-Z]{4})([A-Z]{2})([A-Z2-9][A-NP-Z0-9])(XXX|[A-WY-Z0-9][A-Z0-9]{2})?$`
    BICMinLength = 8
    BICMaxLength = 11
)

type BIC string
```

A BIC has 8 or 11 uppercase characters: a 4 letter institution code, a 2 letter country code at positions 5 and 6 that must be a valid `country.Code`, a 2 character location code, and an optional 3 character branch code where `XXX` means the primary office. `Validate` also rejects an internal deny-list of German words with valid BIC syntax that turn up when scanning documents: `AMTSGERICHT`, `AUTOBANK`, `DEUTSCHLAND`, `DIENSTGEBER`, `DOCUMENT`, `DOKUMENT`, `FACILITY`, `GELISTET`, `GESAMTNETTO`.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NormalizeBIC(str) (BIC, error)`                   | `BIC(str).Normalized()`.                           |
| `StringIsBIC(str) bool`                            | `BIC(str).Valid()`, so whitespace is not accepted. |
| `bic.Validate() error`, `bic.Valid() bool`         | Strict check of the raw value: length 8 or 11, syntax, country code, deny-list. |
| `bic.Normalized() (BIC, error)`                    | Removes all whitespace, validates, and appends `XXX` to an 8 character BIC. Returns the input unchanged with the error. |
| `bic.NormalizedShort() (BIC, error)`               | Like `Normalized`, but removes the `XXX` suffix of an 11 character BIC. |
| `bic.NormalizedOrNull() NullableBIC`               | Normalized, or `BICNull` if invalid.               |
| `bic.ValidAndNormalized() bool`                    | Valid 11 character form.                           |
| `bic.Parse() (bankCode string, countryCode country.Code, branchCode string, isValid bool)` | The institution code, the country code, and the branch code (`""` for 8 characters). All empty and `false` if invalid. There is no location part. |
| `bic.CountryCode() country.Code`                   | Characters 5 and 6, `""` if invalid.               |
| `bic.TrimBranchCode() BIC`                         | The first 8 characters.                            |
| `bic.IsTestBIC() bool`                             | Valid and the 8th character is `0`.                |
| `bic.IsPassiveSWIFT() bool`                        | Valid and the 8th character is `1`.                |
| `bic.ReceiverPaysFees() bool`                      | Valid and the 8th character is `2`.                |
| `bic.String() string`                              | Normalized (11 characters) if valid, else unchanged. |
| `bic.Nullable() NullableBIC`                       | Type conversion.                                   |
| `bic.ScanString(source, validate) error`           | Assigns the source unchanged, without normalization. With `validate` an invalid source is an error. |
| `bic.Scan(value) error`, `bic.Value()`             | See [Serialization](#serialization).               |
| `BIC.JSONSchema()`                                 | String with the `BICRegex` pattern.                |

Implements: `sql.Scanner`, `driver.Valuer`, `fmt.Stringer`, `strfmt.Scannable`, `types.NormalizableValidator[BIC]`. No JSON or text methods.

## NullableBIC

```go
type NullableBIC string
const BICNull NullableBIC = ""
```

A `BIC` where `""` is the null value. Null counts as valid.

| Method                                          | Description                                        |
|-------------------------------------------------|----------------------------------------------------|
| `Normalized() (NullableBIC, error)`             | Like `BIC.Normalized`, null stays null without error. |
| `NormalizedOrNull() NullableBIC`                | Normalized, or `BICNull` if null or invalid.       |
| `Valid() bool`, `Validate() error`              | Null or a valid BIC.                               |
| `ValidAndNotNull() bool`                        | Not null and valid.                                |
| `ValidAndNormalized() bool`                     | Null, or the valid 11 character form.              |
| `IsNull() bool`, `IsNotNull() bool`             | `bic == ""`, `bic != ""`.                          |
| `Get() BIC`                                     | Panics if null.                                    |
| `GetOr(defaultBIC BIC) BIC`                     | Value or the default.                              |
| `StringOr(defaultString) string`                | Value unchanged, or the default if null.           |
| `String() string`                               | Normalized if valid, else unchanged, `""` for null. |
| `Set(BIC)`, `SetNull()`                         | `Set("")` sets null.                               |
| `ScanString(source, validate) error`            | `""`, `NULL`, `null`, and `nil` set null, other values like `BIC.ScanString`. |
| `Scan(value) error`, `Value()`, `MarshalJSON()` | See [Serialization](#serialization).               |
| `JSONSchema()`                                  | `BICRegex` string or null.                         |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `fmt.Stringer`, `strfmt.Scannable`, `nullable.NullSetable[BIC]`. No `json.Unmarshaler`. There is no `CountryCode` or `NormalizedShort`, use `Get()` first.

## Account

```go
type Account struct {
    IBAN     IBAN                   `json:"iban"`
    BIC      NullableBIC            `json:"bic,omitempty"`
    Currency money.NullableCurrency `json:"currency,omitempty"`
    Holder   nullable.TrimmedString `json:"holder,omitempty"`
}
```

| Method                      | Description                                        |
|-----------------------------|----------------------------------------------------|
| `Valid() bool`              | Not nil, and `IBAN`, `BIC`, and `Currency` are valid. |
| `Validate() error`          | The validation errors of `IBAN`, `BIC`, and `Currency` joined with `errors.Join`, an error for a nil pointer. |
| `Normalize() error`         | Normalizes `IBAN`, `BIC`, and `Currency` in place and returns their joined errors. Invalid fields stay unchanged. |
| `String() string`           | Debug form like `bank.Account{DE89370400440532013000, BIC: COBADEFFXXX, EUR, "Jane Doe"}`. Panics on a nil pointer. |
| `Scan(value) error`         | `string` or `[]byte` decoded with `UnmarshalJSON`. |
| `UnmarshalJSON(j) error`    | A JSON object, or a JSON string with an IBAN. See [Serialization](#serialization). |
| `UnmarshalText(text) error` | Trims whitespace, then like `UnmarshalJSON`, also accepting an unquoted IBAN. |

Implements (pointer receiver): `sql.Scanner`, `json.Unmarshaler`, `encoding.TextUnmarshaler`, `fmt.Stringer`. No `driver.Valuer` and no `MarshalJSON`, `encoding/json` writes the struct with its tags.

## Finders and parsers

| Identifier   | Description                                        |
|--------------|----------------------------------------------------|
| `IBANFinder` | Variable whose `FindAllIndex(str []byte, n int) [][]int` implements `types.Finder`. Checks every position for a country code of the IBAN length table followed by the IBAN length of that country, and returns the candidates that pass `IBAN.Valid`. The search continues after the end of a match. |
| `BICFinder`  | Variable whose `FindAllIndex(str []byte, n int) [][]int` implements `types.Finder`. Returns uppercase BIC candidates with a valid country code that are not on the deny-list and have a word separator (whitespace, punctuation, mark, or symbol) or the text boundary on both sides. |
| `IBANParser` | `Parse(str string, langHints ...language.Code) (string, error)` returns the normalized IBAN. Implements `strfmt.Parser`. |
| `BICParser`  | `Parse(str string, langHints ...language.Code) (string, error)` returns `str` unchanged if it is a valid BIC, without normalization. Implements `strfmt.Parser`. |

For both finders `n < 0` returns all matches, `n == 0` none, and `n > 0` at most `n` valid matches. The `ScanString` methods implement `strfmt.Scannable`, which `strfmt.Scan` uses to assign parsed strings.

## CAMT.053 structs

`CAMT53`, `CAMT53Balance`, `CAMT53Entry`, and `CAMT53Amount` are plain structs with `encoding/xml` tags for ISO 20022 CAMT.053 bank-to-customer statements. The package has no parsing logic of its own, and the structs have no tests:

```go
var stmt bank.CAMT53
err := xml.NewDecoder(reader).Decode(&stmt)
```

The tag paths start below the root element, which is not checked.

| `CAMT53` field         | Type              | XML path                                      |
|------------------------|-------------------|-----------------------------------------------|
| `MessageID`            | `string`          | `BkToCstmrStmt>GrpHdr>MsgId`                  |
| `Created`              | `time.Time`       | `BkToCstmrStmt>GrpHdr>CreDtTm`                |
| `StatementID`          | `string`          | `BkToCstmrStmt>Stmt>Id`                       |
| `ElectronicSequenceNr` | `string`          | `BkToCstmrStmt>Stmt>ElctrncSeqNb`             |
| `LegalSequenceNr`      | `string`          | `BkToCstmrStmt>Stmt>LglSeqNb`                 |
| `FromDate`, `ToDate`   | `time.Time`       | `BkToCstmrStmt>Stmt>FrToDt>FrDtTm`, `ToDtTm`  |
| `IBAN`                 | `IBAN`            | `BkToCstmrStmt>Stmt>Acct>Id>IBAN`             |
| `Currency`             | `money.Currency`  | `BkToCstmrStmt>Stmt>Acct>Ccy`                 |
| `BankName`             | `string`          | `BkToCstmrStmt>Stmt>Acct>Svcr>FinInstnId>Nm`  |
| `BIC`                  | `BIC`             | `BkToCstmrStmt>Stmt>Acct>Svcr>FinInstnId>BIC` |
| `Balance`              | `[]CAMT53Balance` | `BkToCstmrStmt>Stmt>Bal`                      |
| `Entries`              | `[]CAMT53Entry`   | `BkToCstmrStmt>Stmt>Ntry`                     |

| `CAMT53Balance` field (below `Bal`) | Type           | XML path                                 |
|-------------------------------------|----------------|------------------------------------------|
| `Type`                              | `string`       | `Tp>CdOrPrtry>Cd`, like `PRCD` or `CLBD` |
| `Proprietary`                       | `string`       | `Tp>CdOrPrtry>Prtry`                     |
| `Amount`                            | `CAMT53Amount` | `Amt`                                    |
| `CreditOrDebit`                     | `string`       | `CdtDbtInd`, `CRDT` or `DBIT`            |
| `Date`                              | `date.Date`    | `Dt>Dt`                                  |

| `CAMT53Entry` field (below `Ntry`) | Type           | XML path                                           |
|------------------------------------|----------------|----------------------------------------------------|
| `Amount`                           | `CAMT53Amount` | `Amt`                                              |
| `CreditOrDebit`                    | `string`       | `CdtDbtInd`, `CRDT` or `DBIT`                      |
| `Status`                           | `string`       | `Sts`, like `BOOK`, `PDNG`, or `INFO`              |
| `BookingDate`, `ValueDate`         | `date.Date`    | `BookgDt>Dt`, `ValDt>Dt`                           |
| `ReferenceCode`                    | `string`       | `AcctSvcrRef`                                      |
| `DebitorName`, `CreditorName`      | `string`       | `NtryDtls>TxDtls>RltdPties>Dbtr>Nm`, `Cdtr>Nm`     |
| `DebitorAddr`, `CreditorAddr`      | `[]string`     | `NtryDtls>TxDtls>RltdPties>Dbtr>PstlAdr>AdrLine`, `Cdtr>PstlAdr>AdrLine` |
| `DebitorIBAN`, `CreditorIBAN`      | `IBAN`         | `NtryDtls>TxDtls>RltdPties>DbtrAcct>Id>IBAN`, `CdtrAcct>Id>IBAN` |
| `DebitorBIC`, `CreditorBIC`        | `BIC`          | `NtryDtls>TxDtls>RltdAgts>DbtrAgt>FinInstnId>BIC`, `CdtrAgt>FinInstnId>BIC` |
| `Reference`                        | `string`       | `NtryDtls>TxDtls>RmtInf>Strd>CdtrRefInf>Ref`       |

`CAMT53Amount` has the `Amount money.Amount` from the element text and the `Currency money.Currency` from the `Ccy` attribute.

## Serialization

`Scan` and JSON decoding of `IBAN`, `BIC`, and their nullable variants never normalize or validate.

| Type                          | `Value()` writes                             | `Scan` accepts                                     | JSON                                               | Text                                               |
|-------------------------------|----------------------------------------------|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|
| `IBAN`, `BIC`                 | The string unchanged, `""` as `""`, not NULL | `string` and `[]byte` unchanged, NULL as `""`, other types are errors | No methods: plain string, `""` as `""`, `null` leaves the value unchanged | No text methods, `ScanString` for `strfmt`         |
| `NullableIBAN`, `NullableBIC` | `""` as NULL, else the string unchanged      | Like `IBAN`, NULL is null                          | `MarshalJSON` writes `""` as `null`, else the string unchanged. No `UnmarshalJSON`: strings are assigned unchanged, `null` leaves the value unchanged | No text methods, `ScanString` for `strfmt`         |
| `Account`                     | No `Valuer`, marshal to JSON yourself        | `string` or `[]byte` with JSON or an IBAN, see `UnmarshalJSON`. NULL is an error. | Marshal: `{"iban":"...","bic":"...","currency":"...","holder":"..."}`, `iban` always, the others omitted when empty. Unmarshal: a JSON object is assigned without normalization or validation, a JSON string must be a valid IBAN and is normalized, `BIC`, `Currency`, and `Holder` become null. `null` leaves the account unchanged. | `UnmarshalText` like `UnmarshalJSON` after trimming, also for an unquoted IBAN |

## Gotchas

- **IBAN normalization only removes whitespace.** `de89370400440532013000` fails with `invalid IBAN country code` and `DE89-3704-0044-0532-0130-00` with `wrong IBAN length`. Uppercase or clean punctuation yourself.
- **IBAN validation normalizes, BIC validation doesn't.** `IBAN("DE89 3704 0044 0532 0130 00").Valid()` and `StringIsIBAN` are true, while `BIC("BKAU AT WW").Valid()` and `StringIsBIC` are false, although `NormalizeBIC("BKAU AT WW")` returns `BKAUATWWXXX`. Likewise `IBAN.ScanString` and `IBANParser` normalize, `BIC.ScanString` and `BICParser` don't.
- **Store the normalized form.** A valid IBAN with spaces has `ValidAndNormalized() == false`, and `Value()` writes it with the spaces.
- **8 vs 11 character BICs.** `BIC("BKAUATWW") == BIC("BKAUATWWXXX")` is false. `Normalized` and `String` produce 11 characters, `NormalizedShort` 8, and `ValidAndNormalized` is only true for the 11 character form. Normalize both sides before comparing.
- **`String` normalizes.** `fmt` and `%s` print an IBAN without spaces and a BIC with 11 characters, also the results of `NormalizedWithSpaces` and `NormalizedShort`. Convert with `string(iban)` to print the value as it is.
- **The BIC country code is at characters 5 and 6**, not at the start like in an IBAN.
- **Only the nullable types write SQL NULL.** `IBAN("").Value()` and `BIC("").Value()` return `""`. All four types scan NULL as `""`.
- **Decoded values are unchecked.** `Scan`, JSON decoding, and the JSON object form of `Account` accept anything: `{"iban":"invalid","bic":"nope"}` decodes into an `Account` without error. Call `Normalize` or `Validate` after decoding.
- **An IBAN string resets the other `Account` fields.** `UnmarshalJSON` of `"DE89370400440532013000"` sets `BIC`, `Currency`, and `Holder` to null.
- **JSON `null` doesn't reset `NullableIBAN` and `NullableBIC`**, because they have no `UnmarshalJSON`. Decoding `null` into a field that already has a value keeps the old value.
- **IBAN lengths.** `IBANMaxLength` is 32, the longest IBAN known to the table, while ISO 13616 allows 34. `GG`, `IM`, and `JE` are accepted as IBAN countries with length 22. `BankAndAccountNumbers` only supports AT, CH, and DE.
- **`IBANFinder` only finds compact uppercase IBANs** and checks no word boundaries: `AT61 1904 3002 3457 3201` and `at611904300234573201` are not found, and `XAT611904300234573201` yields `AT611904300234573201`.
- **`BICFinder` finds uppercase words with BIC syntax.** The deny-list has only 9 entries, so `DEUTSCHE BANK` yields `DEUTSCHE` (country code `SC`). Check the context of finder results.
- **CAMT.053 limits.**
  - The `time.Time` fields are decoded as RFC 3339. A datetime without offset like `2024-01-31T10:00:00` aborts the whole `Decode` with a parse error.
  - The tags match the camt.053.001.02 element names: `<BIC>` (newer versions use `<BICFI>`) and `<Dt>` for dates (a `<DtTm>` alternative is ignored). Unknown elements are skipped without error, so fields stay empty.
  - `IBAN`, `BIC`, `Currency`, and the `date.Date` fields are copied as they are, without normalization or validation. An IBAN in groups of 4 keeps its spaces.
  - With several `<Stmt>` elements, the scalar fields hold the values of the last statement, while `Balance` and `Entries` collect those of all statements. The same applies to several `<TxDtls>` within one entry.
  - Only the structured creditor reference is captured. `Ustrd`, `EndToEndId`, and `BkTxCd` are not.

## Examples

Normalize user input:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/bank"
)

func main() {
	iban, err := bank.NormalizeIBAN(" DE89 3704 0044 0532 0130 00\n")
	if err != nil {
		panic(err)
	}
	fmt.Println(iban)
	spaced, _ := iban.NormalizedWithSpaces()
	fmt.Println(string(spaced)) // fmt would call String, which removes the spaces
	bankNo, accountNo, err := iban.BankAndAccountNumbers()
	fmt.Println(iban.CountryCode(), bankNo, accountNo, err)

	bic, err := bank.NormalizeBIC("BKAU AT WW")
	fmt.Println(bic, err)
	short, _ := bic.NormalizedShort()
	fmt.Println(string(short)) // String would append XXX again
	fmt.Println(bank.StringIsBIC("BKAU AT WW"))

	// Output:
	// DE89370400440532013000
	// DE89 3704 0044 0532 0130 00
	// DE 37040044 0532013000 <nil>
	// BKAUATWWXXX <nil>
	// BKAUATWW
	// false
}
```

Find IBANs and BICs in text:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/bank"
)

func main() {
	text := []byte("Please pay to IBAN AT611904300234573201 (BIC GIBAATWWXXX) by Friday.")

	for _, loc := range bank.IBANFinder.FindAllIndex(text, -1) {
		fmt.Println("IBAN:", string(text[loc[0]:loc[1]]))
	}
	for _, loc := range bank.BICFinder.FindAllIndex(text, -1) {
		fmt.Println("BIC:", string(text[loc[0]:loc[1]]))
	}

	// Output:
	// IBAN: AT611904300234573201
	// BIC: GIBAATWWXXX
}
```

Only the nullable types write SQL NULL and JSON `null`:

```go
package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/bank"
)

func main() {
	for _, v := range []driver.Valuer{bank.IBAN(""), bank.IBANNull, bank.BIC(""), bank.BICNull} {
		value, _ := v.Value()
		fmt.Printf("%T: %#v\n", v, value)
	}

	j, err := json.Marshal(bank.Account{IBAN: "DE89370400440532013000", BIC: bank.BICNull})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))

	// Output:
	// bank.IBAN: ""
	// bank.NullableIBAN: <nil>
	// bank.BIC: ""
	// bank.NullableBIC: <nil>
	// {"iban":"DE89370400440532013000"}
}
```

## Related

- [`country`](../country/README.md): `country.Code` of IBANs and BICs.
- [`money`](../money/README.md): `money.NullableCurrency` of `Account`, currencies and amounts of CAMT.053 statements.
- [`nullable`](../nullable/README.md): `nullable.TrimmedString` of `Account.Holder`.
- [`strfmt`](../strfmt/README.md): `strfmt.Parser` and `strfmt.Scannable`.
- [Null conventions](../docs/explanation-null-conventions.md): why only the nullable types write NULL.
- [Validation and normalization](../docs/explanation-validation-and-normalization.md): the `Valid`/`Validate`/`Normalized` pattern.
- [Store optional values](../docs/howto-store-optional-values.md): `NullableIBAN` and `NullableBIC` columns.
- [Extract identifiers from text](../docs/howto-extract-identifiers-from-text.md): `IBANFinder` and `BICFinder`.
