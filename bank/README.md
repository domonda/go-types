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
| `iban.Nullable() NullableIBAN`                     | Type conversion.                                   |
| `iban.ScanString(source, validate) error`          | Without `validate` like `Scan`. With `validate` the normalized source or the error of `Normalized`, also for a blank source. |
| `iban.Scan(value) error`, `iban.Value()`, JSON and XML | See [Serialization](#serialization).               |
| `IBAN.JSONSchema()`                                | String with the `IBANRegex` pattern.               |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `xml.Unmarshaler`, `xml.UnmarshalerAttr`, `xml.MarshalerAttr`, `strfmt.Scannable`, `types.NormalizableValidator[IBAN]`. No `String` method, `fmt` prints the stored IBAN, and no `UnmarshalText`, so `encoding/json` reads the keys of a `map[IBAN]V` as given.

## NullableIBAN

```go
type NullableIBAN string
const IBANNull NullableIBAN = ""
```

An `IBAN` where `""` is the null value. A value of only the runes that `strutil.TrimSpace` removes (whitespace, control and other non-printable runes, invalid UTF-8) is null too, like for `country.NullableCode`. Null counts as valid.

| Method                                          | Description                                        |
|-------------------------------------------------|----------------------------------------------------|
| `Normalized() (NullableIBAN, error)`            | Like `IBAN.Normalized`, `IBANNull` for null without error. |
| `NormalizedWithSpaces() (NullableIBAN, error)`  | Like `IBAN.NormalizedWithSpaces`, `IBANNull` for null. |
| `NormalizedOrNull() NullableIBAN`               | Normalized, or `IBANNull` if null or invalid.      |
| `Valid() bool`, `Validate() error`              | Null or a valid IBAN.                              |
| `ValidAndNotNull() bool`                        | Not null and valid.                                |
| `ValidAndNormalized() bool`                     | `""`, or valid and in compact form. False for a whitespace only null, which normalizes to `""`. |
| `CountryCode() country.Code`                    | Like `IBAN.CountryCode`, `""` for null or invalid values. |
| `IsNull() bool`, `IsNotNull() bool`             | Null is `""` or only runes that `strutil.TrimSpace` removes. |
| `Get() IBAN`                                    | Panics if null.                                    |
| `GetOr(defaultIBAN IBAN) IBAN`                  | Value or the default.                              |
| `StringOr(defaultString) string`                | As stored, or the default if null.                 |
| `Set(IBAN)`, `SetNull()`                        | `Set("")` sets null.                               |
| `ScanString(source, validate) error`            | `""`, whitespace only, `NULL`, `null`, and `nil` set null, other values like `IBAN.ScanString`. |
| `Scan(value) error`, `Value()`, `MarshalJSON()` | See [Serialization](#serialization).               |
| `JSONSchema()`                                  | `IBANRegex` string or null.                        |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `xml.Unmarshaler`, `xml.UnmarshalerAttr`, `xml.MarshalerAttr`, `strfmt.Scannable`, `nullable.NullSetable[IBAN]`, `types.NormalizableValidator[NullableIBAN]`. No `String` method.

## BIC

```go
const (
    BICRegex     = `^([A-Z]{4})([A-Z]{2})([A-Z2-9][A-NP-Z0-9])(XXX|[A-WY-Z0-9][A-Z0-9]{2})?$`
    BICMinLength = 8
    BICMaxLength = 11
)

type BIC string
```

A BIC has 8 or 11 uppercase characters: a 4 letter institution code, a 2 letter country code at positions 5 and 6 that must be a valid `country.Code`, a 2 character location code, and an optional 3 character branch code where `XXX` means the primary office. `Validate` also rejects an internal deny-list of German words with valid BIC syntax that turn up when scanning documents: `AMTSGERICHT`, `AUTOBANK`, `DEUTSCHLAND`, `DIENSTGEBER`, `DOCUMENT`, `DOKUMENT`, `FACILITY`, `GELISTET`, `GESAMTNETTO`. The 8 character words are also rejected with the `XXX` branch code that `Normalized` appends, like `AUTOBANKXXX`.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NormalizeBIC(str) (BIC, error)`                   | `BIC(str).Normalized()`.                           |
| `StringIsBIC(str) bool`                            | `BIC(str).Valid()`.                                |
| `bic.Validate() error`, `bic.Valid() bool`         | Normalizes first: length 8 or 11 after removing all whitespace, syntax, country code, deny-list. |
| `bic.Normalized() (BIC, error)`                    | Removes all whitespace, validates, and appends `XXX` to an 8 character BIC. Returns the input unchanged with the error. |
| `bic.NormalizedShort() (BIC, error)`               | Like `Normalized`, but removes the `XXX` suffix of an 11 character BIC. |
| `bic.NormalizedOrNull() NullableBIC`               | Normalized, or `BICNull` if invalid.               |
| `bic.ValidAndNormalized() bool`                    | Valid 11 character form.                           |
| `bic.Parse() (bankCode string, countryCode country.Code, branchCode string, isValid bool)` | The institution code, the country code, and the branch code (`""` for 8 characters). All empty and `false` if invalid. There is no location part. Parses the value as given, so whitespace makes it invalid. |
| `bic.CountryCode() country.Code`                   | Characters 5 and 6 of the normalized BIC, `""` if invalid. |
| `bic.TrimBranchCode() BIC`                         | The first 8 characters of the normalized BIC, or of the unchanged BIC if invalid. |
| `bic.IsTestBIC() bool`                             | Valid and the 8th character of the normalized BIC is `0`. |
| `bic.IsPassiveSWIFT() bool`                        | Valid and the 8th character of the normalized BIC is `1`. |
| `bic.ReceiverPaysFees() bool`                      | Valid and the 8th character of the normalized BIC is `2`. |
| `bic.Nullable() NullableBIC`                       | Type conversion.                                   |
| `bic.ScanString(source, validate) error`           | Without `validate` like `Scan`. With `validate` the normalized source or the error of `Normalized`, also for a blank source. |
| `bic.Scan(value) error`, `bic.Value()`, JSON and XML | See [Serialization](#serialization).               |
| `BIC.JSONSchema()`                                 | String with the `BICRegex` pattern.                |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `xml.Unmarshaler`, `xml.UnmarshalerAttr`, `xml.MarshalerAttr`, `strfmt.Scannable`, `types.NormalizableValidator[BIC]`. No `String` method, `fmt` prints the stored BIC, and no `UnmarshalText`, so `encoding/json` reads the keys of a `map[BIC]V` as given.

## NullableBIC

```go
type NullableBIC string
const BICNull NullableBIC = ""
```

A `BIC` where `""` is the null value. A value of only the runes that `strutil.TrimSpace` removes (whitespace, control and other non-printable runes, invalid UTF-8) is null too, like for `country.NullableCode`. Null counts as valid.

| Method                                          | Description                                        |
|-------------------------------------------------|----------------------------------------------------|
| `Normalized() (NullableBIC, error)`             | Like `BIC.Normalized`, `BICNull` for null without error. |
| `NormalizedOrNull() NullableBIC`                | Normalized, or `BICNull` if null or invalid.       |
| `Valid() bool`, `Validate() error`              | Null or a valid BIC.                               |
| `ValidAndNotNull() bool`                        | Not null and valid.                                |
| `ValidAndNormalized() bool`                     | `""`, or the valid 11 character form. False for a whitespace only null, which normalizes to `""`. |
| `IsNull() bool`, `IsNotNull() bool`             | Null is `""` or only runes that `strutil.TrimSpace` removes. |
| `Get() BIC`                                     | Panics if null.                                    |
| `GetOr(defaultBIC BIC) BIC`                     | Value or the default.                              |
| `StringOr(defaultString) string`                | As stored, or the default if null.                 |
| `Set(BIC)`, `SetNull()`                         | `Set("")` sets null.                               |
| `ScanString(source, validate) error`            | `""`, whitespace only, `NULL`, `null`, and `nil` set null, other values like `BIC.ScanString`. |
| `Scan(value) error`, `Value()`, `MarshalJSON()` | See [Serialization](#serialization).               |
| `JSONSchema()`                                  | `BICRegex` string or null.                         |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `xml.Unmarshaler`, `xml.UnmarshalerAttr`, `xml.MarshalerAttr`, `strfmt.Scannable`, `nullable.NullSetable[BIC]`, `types.NormalizableValidator[NullableBIC]`. No `String` method. There is no `CountryCode` or `NormalizedShort`, use `Get()` first.

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
| `String() string`           | Debug form with the fields as stored, like `bank.Account{DE89370400440532013000, BIC: COBADEFFXXX, EUR, "Jane Doe"}`. Panics on a nil pointer. |
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

`IBAN`, `BIC`, and their nullable variants follow the rules of [Reading and writing values](../README.md#reading-and-writing-values): `Scan`, JSON and XML trim and uppercase. Internal spaces stay, and an 8-character BIC is not extended. `fmt` prints the stored value.

| Type                          | `Value()` writes                                   | `Scan` accepts                                     | JSON                                               | Text and XML                                       |
|-------------------------------|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|
| `IBAN`, `BIC`                 | Trimmed and uppercased, blank as NULL. Internal spaces stay. An 8-character BIC stays 8 characters. An invalid value is written, not rejected | `string` and `[]byte` trimmed and uppercased, blank as `""`, NULL as `""`, other types are errors | `MarshalJSON` writes the same prettification, blank as `""`. `UnmarshalJSON` reads like `Scan`, `null` leaves the value unchanged | `MarshalText` like `MarshalJSON`. XML elements and attributes read like `Scan`. No `UnmarshalText`. `ScanString` with validate still calls `Normalized` |
| `NullableIBAN`, `NullableBIC` | Null (see `IsNull`) as NULL, else like `IBAN` and `BIC` | Like `IBAN`, NULL and blank as null                | `MarshalJSON` writes null as `null`, else like `IBAN` and `BIC`. `UnmarshalJSON` reads like `Scan` and sets null for `null`, other JSON values are errors | `MarshalText` writes `""` for null, else like `IBAN` and `BIC`. XML like `Scan` |
| `Account`                     | No `Valuer`, marshal to JSON yourself              | `string` or `[]byte` with JSON or an IBAN, see `UnmarshalJSON`. NULL is an error. | Marshal: `{"iban":"...","bic":"...","currency":"...","holder":"..."}`, `iban` always, the others omitted when empty. Unmarshal: a JSON object is decoded with the field types. A JSON string is trimmed and uppercased and kept even when the IBAN is invalid; `BIC`, `Currency`, and `Holder` become null. `null` leaves the account unchanged. A JSON array is an error. | `UnmarshalText` trims and uppercases a bare IBAN the same way, and still errors on a value that is not an object or a complete string |

## Gotchas

- **IBAN normalization only removes whitespace.** `de89370400440532013000` fails with `invalid IBAN country code` and `DE89-3704-0044-0532-0130-00` with `wrong IBAN length`. Uppercase or clean punctuation yourself.
- **Validation normalizes, parsing a BIC doesn't.** `IBAN("DE89 3704 0044 0532 0130 00").Valid()`, `BIC("BKAU AT WW").Valid()`, `StringIsIBAN` and `StringIsBIC` are true. `IBAN.ScanString`, `BIC.ScanString` and `IBANParser` normalize, while `BICParser` rejects whitespace, because it returns the BIC as given.
- **A value built in Go keeps its spelling until it is read or written.** Reading trims and uppercases and keeps internal spaces, so `bank.IBAN("DE89 3704 0044 0532 0130 00")` stays grouped. It has `ValidAndNormalized() == false` until `Normalized()`.
- **8 vs 11 character BICs.** `BIC("BKAUATWW") == BIC("BKAUATWWXXX")` is false. `Normalized` and `ScanString` with validate append `XXX`. Reading, `Value` and `MarshalJSON` keep 8 characters. `NormalizedShort` returns 8, and `ValidAndNormalized` is only true for the 11 character form. Normalize both sides before comparing.
- **`fmt` prints the stored value.** `fmt` and `%s` print an IBAN with its spaces and the result of `NormalizedShort` with 8 characters, as stored.
- **The BIC country code is at characters 5 and 6**, not at the start like in an IBAN.
- **Empty values are SQL NULL.** `IBAN("").Value()` and `BIC("").Value()` return SQL NULL like the null nullable types, which a `NOT NULL` column rejects. All four types scan NULL as `""`. Only the nullable types write JSON `null`.
- **`Value` writes invalid values.** `IBAN("DE00 1234").Value()` returns `"DE00 1234"`. `MarshalJSON` and `MarshalText` write the same trimmed and uppercased text. `Normalized()` still returns the error.
- **Decoded values are not validated.** `Scan`, JSON and XML decoding, and the JSON object form of `Account` accept anything: `{"iban":"invalid","bic":"nope"}` decodes into an `Account` without error. Call `Validate` after decoding.
- **An IBAN string resets the other `Account` fields.** `UnmarshalJSON` of `"DE89370400440532013000"` sets `BIC`, `Currency`, and `Holder` to null.
- **IBAN lengths.** `IBANMaxLength` is 32, the longest IBAN known to the table, while ISO 13616 allows 34. `GG`, `IM`, and `JE` are accepted as IBAN countries with length 22. `BankAndAccountNumbers` only supports AT, CH, and DE.
- **`IBANFinder` only finds compact uppercase IBANs** and checks no word boundaries: `AT61 1904 3002 3457 3201` and `at611904300234573201` are not found, and `XAT611904300234573201` yields `AT611904300234573201`.
- **`BICFinder` finds uppercase words with BIC syntax.** The deny-list has only 9 entries, so `DEUTSCHE BANK` yields `DEUTSCHE` (country code `SC`). Check the context of finder results.
- **CAMT.053 limits.**
  - The `time.Time` fields are decoded as RFC 3339. A datetime without offset like `2024-01-31T10:00:00` aborts the whole `Decode` with a parse error.
  - The tags match the camt.053.001.02 element names: `<BIC>` (newer versions use `<BICFI>`) and `<Dt>` for dates (a `<DtTm>` alternative is ignored). Unknown elements are skipped without error, so fields stay empty.
  - `IBAN`, `BIC`, and `Currency` are read like by `Scan`: trimmed and uppercased, internal spaces kept, without validation. `date.Date` is read like XML: surrounding whitespace is trimmed, and a date that is not already `YYYY-MM-DD` stays as written. An IBAN in groups of 4 stays grouped.
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
	// true
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

All four types write an empty value as SQL NULL, but only the nullable types write JSON `null`:

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
	// bank.IBAN: <nil>
	// bank.NullableIBAN: <nil>
	// bank.BIC: <nil>
	// bank.NullableBIC: <nil>
	// {"iban":"DE89370400440532013000"}
}
```

## Related

- [`country`](../country/README.md): `country.Code` of IBANs and BICs.
- [`money`](../money/README.md): `money.NullableCurrency` of `Account`, currencies and amounts of CAMT.053 statements.
- [`nullable`](../nullable/README.md): `nullable.TrimmedString` of `Account.Holder`.
- [`strfmt`](../strfmt/README.md): `strfmt.Parser` and `strfmt.Scannable`.
- [Null conventions](../docs/explanation-null-conventions.md): when a value is SQL NULL and JSON `null`.
- [Validation and normalization](../docs/explanation-validation-and-normalization.md): the `Valid`/`Validate`/`Normalized` pattern.
- [Store optional values](../docs/howto-store-optional-values.md): `NullableIBAN` and `NullableBIC` columns.
- [Extract identifiers from text](../docs/howto-extract-identifiers-from-text.md): `IBANFinder` and `BICFinder`.
