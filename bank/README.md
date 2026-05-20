# bank

Banking identifiers and statement parsing: IBAN, BIC/SWIFT, bank-account composites, and ISO 20022 CAMT.053 statements.

```
import "github.com/domonda/go-types/bank"
```

## IBAN

International Bank Account Number. 15–32 characters; layout `<CC><check><BBAN>`.

```go
const (
    IBANRegex     = `^([A-Z]{2})(\d{2})([A-Z\d]{8,30})$`
    IBANMinLength = 15
    IBANMaxLength = 32
)
type IBAN string
```

| Method                   | Description                                        |
|--------------------------|----------------------------------------------------|
| `Normalized()`           | Uppercase, strip spaces/punctuation, validate length+regex. |
| `Valid()` / `Validate()` | Pass/error variants.                               |
| `ValidAndNormalized()`   | Already in canonical form.                         |
| `CountryCode()`          | The leading 2-letter `country.Code`.               |
| `Nullable()`             | Convert to `NullableIBAN`.                         |
| `ScanString(src, valid)` | Assign from string, optional validation.           |

Helpers: `NormalizeIBAN(str)`, `StringIsIBAN(str)`, `IBANFinder` (scan text), `IBANParser` (implements `strfmt.Parser`).

## BIC / SWIFT

8 or 11 characters. Empty is treated as SQL NULL on `BIC`; `NullableBIC` keeps empty as a valid null value.

```go
const (
    BICRegex     = `^([A-Z]{4})([A-Z]{2})([A-Z2-9][A-NP-Z0-9])(XXX|[A-WY-Z0-9][A-Z0-9]{2})?$`
    BICMinLength = 8
    BICMaxLength = 11
)
type BIC string
```

Methods: `Normalized()` (strip spaces, append `"XXX"` when 8 chars), `Valid()`, `Validate()`, `Parse()` (returns bank/country/location/branch parts), plus an internal allow-list (`falseBICs`) that rejects known-bad codes.

`BICFinder` scans byte slices for BICs that are surrounded by word boundaries, validated against the country list and the `falseBICs` map.

## Account

Composite type with optional currency and holder.

```go
type Account struct {
    IBAN     IBAN
    BIC      NullableBIC
    Currency money.NullableCurrency
    Holder   nullable.TrimmedString
}
```

Methods: `Valid()`, `Validate()` (joined errors), `Normalize()` (mutates in place), `String()`, plus `sql.Scanner`/`Valuer` using JSON for storage. JSON marshaling round-trips through `UnmarshalJSON`/`MarshalJSON`.

## Bank enums

`AccountType` (`CURRENT`, `SAVINGS`), `TransactionType` (`INCOMING`, `OUTGOING`), `PaymentStatus`. Each implements `sql.Scanner`/`driver.Valuer` with empty-string-as-NULL semantics for the SQL-facing types.

## CAMT.053 — ISO 20022 statements

```go
type CAMT53 struct {
    MessageID            string
    Created              time.Time
    StatementID          string
    ElectronicSequenceNr string
    LegalSequenceNr      string
    FromDate, ToDate     time.Time
    IBAN                 IBAN
    Currency             money.Currency
    BankName             string
    BIC                  BIC
    Balance              []CAMT53Balance
    Entries              []CAMT53Entry
}
```

Plus `CAMT53Amount`, `CAMT53Balance`, `CAMT53Entry`. Struct tags are XML-shaped — unmarshal a CAMT.053 file with `encoding/xml`:

```go
var stmt bank.CAMT53
err := xml.NewDecoder(reader).Decode(&stmt)
```

Entry fields cover debitor/creditor parties, IBANs, BICs, booking + value dates, status (`BOOK`/`PDNG`/`INFO`), credit/debit indicator (`CRDT`/`DBIT`), and the structured reference.

## strfmt integration

`IBANParser{}` and the analogous BIC helpers implement `strfmt.Parser`, suitable for `strfmt.Scanner` pipelines.

## Related

- `country` — leading country code in IBAN/BIC.
- `money` — currency on bank accounts and CAMT.053 amounts.
- `nullable` — used for optional fields on `Account`.
