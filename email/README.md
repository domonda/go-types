# email

Email addresses (lenient parsing, normalization, lists and sets), full message structures, and MIME/TNEF parsing.

```
import "github.com/domonda/go-types/email"
```

## Address

```go
type Address string // "user@example.com" or "John Doe <user@example.com>"
```

Lenient parser that fixes common malformations encountered in real-world mail and lowercases the address part. More permissive than `net/mail.ParseAddress`.

| Function / Method                           | Description                                        |
|---------------------------------------------|----------------------------------------------------|
| `NormalizedAddress(str)`                    | Parse + normalize.                                 |
| `AddressFrom(*mail.Address)`                | Convert a stdlib address.                          |
| `a.Normalized()` / `Validate()` / `Valid()` | Validation variants.                               |
| `a.Parse()`                                 | To `*mail.Address` via the package's lenient parser. |
| `ParseAddress(str)`                         | Underlying parser; supports umlauts, RFC 2047 encoded-words, quoted names. |

`AddressRegexp` is the compiled regex for the email-only form. It tolerates a wide range of umlaut/diacritic characters since real-world mail commonly violates strict RFC 2821/2822.

## AddressList

```go
type AddressList string // comma-separated
```

`NormalizeAddressList(str)` parses, normalizes, and de-duplicates by normalized address part. Methods: `Normalized()`, `Validate()`, plus `ParseAddressList(str)` for the parsed slice.

## AddressSet

```go
type AddressSet map[Address]struct{}

email.MakeAddressSet(a, b, c)
email.NormalizedAddressSet(a, b, c) // normalizes each entry
```

Set semantics over `Address` values. Sorted iteration and JSON serialization included.

## NullableAddress / NullableAddressList

SQL/JSON-friendly cousins where empty string is the null sentinel.

## Message

```go
type Header = textproto.MIMEHeader

type Message struct {
    ProviderID          nullable.TrimmedString
    InReplyToProviderID nullable.TrimmedString
    ProviderLabels      []string
    // ... headers, body, attachments
}
```

Parsed message headers tracked separately (`Message-Id`, `In-Reply-To`, `References`, `Date`, `From`, `Reply-To`, `To`, `Delivered-To`, `Cc`, `Bcc`, `Subject`). Anything else is exposed via `Header`. Helpers: `IsParsedHeader(key)`, `IsExtraHeader(key)`.

## Parsing wire-format mail

```go
import "os"

f, _ := os.Open("mail.eml")
msg, err := email.ParseMIMEMessage(f)
```

Backed by [`jhillyerd/enmime`](https://github.com/jhillyerd/enmime). For Microsoft TNEF (`winmail.dat`) attachments, see `parsetnef.go`.

## Attachment

```go
type Attachment struct {
    PartID, ContentID, ContentType string
    Inline, OtherPart              bool
    Filename                       string
    Content                        []byte
}

email.NewAttachment(partID, filename, content)
```

`Attachment` implements `fs.FileReader`. `ContentType` is auto-detected via `http.DetectContentType` when constructed.

## Rule

```go
type Rule interface {
    AppliesToMessage(*Message) bool
}
```

Compose rules for filtering: `RuleFunc(func(*Message) bool)`, `BoolRule(true|false)`, `AllRule{r1, r2}` (AND), `AnyRule{r1, r2}` (OR).

## Config

`ProviderDomains()` returns a set of well-known consumer mail-provider domains (gmail, yahoo, gmx, web.de, …). Useful for distinguishing personal from business addresses.

## Related

- `nullable.TrimmedString` — used heavily for optional header fields.
- `notnull` — `Message.ProviderLabels` and similar list fields.
- `uu` — `Attachment.ContentID` defaults to a v4 UUID hex.
