# email

Email addresses, address lists and sets, and email messages. The address parser is more lenient than `net/mail`, because mail found in the wild breaks the RFC rules (umlauts in addresses, unquoted commas in names, `addr (Name)` comments). Normalization lowercases the address part, so normalized addresses can be compared as strings. Messages are parsed from MIME (`.eml`), Microsoft TNEF (`winmail.dat`), or the package's own JSON form, and can be encoded back to MIME.

```
import "github.com/domonda/go-types/email"
```

## Address

```go
type Address string // "user@example.com" or "John Doe <user@example.com>"
```

A non-normalized address with an optional display name. All parsing goes through `ParseAddress`.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseAddress(str) (*mail.Address, error)`         | Lenient parser. Lowercases the address part, decodes RFC 2047 names in the charsets of `golang.org/x/net/html/charset` (ISO 8859-2, Windows-1250, ...), accepts quoted names, `addr (Name)` comments, and umlauts. A comment can nest: `addr (Name (Dept))` has the name `Name (Dept)`, and `\)` doesn't end it. A comment after the address that isn't closed is an error. A name equal to the address is dropped. |
| `NormalizedAddress(str) (Address, error)`          | `Address(str).Normalized()`.                       |
| `AddressFrom(*mail.Address) Address`               | `""` for nil, the bare address without a name, else `"Name" <addr>` like `mail.Address.String()`, but with a non-ASCII name as UTF-8 instead of RFC 2047 encoded. A name containing `=?` is RFC 2047 encoded. |
| `FindAllAddresses(text) []Address`                 | All matches of `AddressRegexp` in text, in order, not normalized. |
| `UniqueNormalizedAddressSlice([]Address) []Address` | New sorted slice of the unique address parts. Skips invalid entries, returns nil if none is valid, does not modify the input. |
| `AddressRegexp`                                    | Compiled, unanchored regular expression for an address without name. |
| `a.Parse() (*mail.Address, error)`                 | `ParseAddress(string(a))`.                         |
| `a.Normalized() (Address, error)`                  | Parsed and re-formatted with `AddressFrom`. Returns `a` unchanged with the error. |
| `a.Validate() error` / `a.Valid() bool`            | Whether `Parse` succeeds.                          |
| `a.ValidAndNormalized() bool`                      | Valid and equal to its normalized form.            |
| `a.NamePart() (string, error)`                     | Decoded display name, `""` if there is none.       |
| `a.AddressPart() (Address, error)`                 | Normalized `local@domain` without the display name, a local part that needs quotes quoted like `"a b"@example.com`. |
| `a.AddressPartString() (string, error)`            | Same as `AddressPart` as string.                   |
| `a.LocalPart() (string, error)`                    | Normalized part before the `@`, quoted like `"a b"` if it needs quotes. |
| `a.DomainPart() string`                            | Raw text after the last `@` with trailing `>` and whitespace trimmed, `""` without `@`. Not parsed or lowercased. |
| `a.AsList() AddressList`                           | List containing only `a`.                          |
| `a.Nullable() NullableAddress`                     | Type conversion.                                   |
| `a.ASCII() string`                                 | Normalized, but a non-ASCII name RFC 2047 encoded for message headers, or unchanged if invalid. The address part is not encoded. |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `xml.Unmarshaler`, `xml.UnmarshalerAttr`, `xml.MarshalerAttr`, `strfmt.Scannable`, `types.NormalizableValidator[Address]`, `JSONSchema()` (`string`, format `email`). No `String` method, `fmt` prints the stored address, and no `UnmarshalText`, so `encoding/json` reads the keys of a `map[Address]V` as given.

## AddressList

```go
type AddressList string // "a@example.com, John <j@example.com>"
```

A comma separated list. `Validate` and `Valid` accept the empty list. Use `NullableAddressList` for lists that can be empty. `AddressList` is a collection: it is read and written as stored in every format, see [Serialization](#serialization).

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseAddressList(list) ([]*mail.Address, error)`  | Lenient. Skips a malformed entry up to the next comma and keeps parsing, returns the parsed addresses together with an `errors.Join` error. A comma in a comment doesn't separate entries, like for `SplitAddressList`, and a comment that isn't closed holds the rest of the list: it is an error and parsing stops at it. Parsing stops after 100 errors, so addresses after them are not returned. `""` and a whole `undisclosed-recipients` / `withheld-recipients` placeholder (also with a space instead of the dash, or as empty group like `undisclosed-recipients:;`) give nil without error. Text after a placeholder is parsed like any list. |
| `SplitAddressList(list) []Address`                 | Splits at the commas between entries, outside double quotes, angle brackets and comments, and returns the trimmed entries unchanged: not parsed, normalized or validated. Like `ParseAddressList`, entries without `@` are joined with a following `Name <address>` entry, so the unquoted name of `Doe, John <john@example.com>` stays one entry. A blank list and the placeholders give nil. Never fails. |
| `NormalizeAddressList(list) ([]string, error)`     | Strict. Normalized addresses, de-duplicated by address part, first spelling wins. Any bad entry returns `nil, err`, an empty list `nil, nil`. |
| `AddressListJoin(addrs ...Address) AddressList`    | Joins with `", "`, skips empty addresses.          |
| `AddressListJoinStrings(addrs ...string) AddressList` | Same for strings.                                  |
| `l.Append(addrs ...Address) AddressList`           | Appends with `", "`, skips empty addresses.        |
| `l.Parse() ([]*mail.Address, error)`               | `ParseAddressList(string(l))`, lenient.            |
| `l.SplitNormalized() ([]Address, error)`           | Lenient. Parsed and normalized addresses (via `AddressFrom`), nil for an empty list. |
| `l.SplitRaw() []Address`                           | The entries as stored, `SplitAddressList(string(l))`. |
| `l.Normalized() (AddressList, error)`              | Strict. Every address normalized and re-joined with `", "`. Does not de-duplicate. Returns `l` unchanged with the error. |
| `l.UniqueAddressParts() (AddressSet, error)`       | Strict. Set of the address parts, `nil, nil` for an empty list. |
| `l.Validate() error` / `l.Valid() bool`            | Strict, any bad entry is an error.                 |
| `l.ValidAndNormalized() bool`                      | Valid and equal to its normalized form.            |
| `l.Nullable() NullableAddressList`                 | Type conversion.                                   |
| `l.ASCII() string`                                 | Like `Address.ASCII` for every address, `""` for a list without addresses. |
| `l.Scan(value) error`, `l.Value()`, `l.MarshalJSON()`, `l.UnmarshalJSON(data)`, `l.MarshalText()` | See [Serialization](#serialization).               |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `xml.MarshalerAttr`, `types.NormalizableValidator[AddressList]`. No `String` method, `fmt` prints the stored list.

## AddressSet

```go
type AddressSet map[Address]struct{}
```

A set of addresses, stored in SQL as a PostgreSQL text array; write `AddressList()` to store it in a plain string column. Like the lists it is a collection: its addresses are read and written as stored. It implements the abstract set interface of the Go 1.28 collections proposal ([go.dev/issue/80590](https://go.dev/issue/80590)).

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `MakeAddressSet(addrs ...Address) AddressSet`      | Set of the passed addresses as they are.           |
| `NormalizedAddressSet(addrs ...Address) (AddressSet, error)` | Set of the normalized addresses (display names kept). Strict: the first invalid address returns `nil, err`. |
| `NormalizedAddressPartSet(addrs ...Address) (AddressSet, error)` | Set of the address parts without names. Strict like `NormalizedAddressSet`. |

| Method                                             | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `Len`, `IsEmpty`, `IsNull`                         | Size, empty or nil, nil only.                      |
| `Contains`, `ContainsAny(addrs...)`, `ContainsAll(seq)` | Membership. `ContainsAny()` is false, `ContainsAll` of an empty sequence true. |
| `Equal`, `Intersects`                              | Set comparison.                                    |
| `All() iter.Seq[Address]`                          | Iterates in undefined order.                       |
| `Sorted() []Address`, `Strings() []string`         | Sorted elements.                                   |
| `AddressList() AddressList`                        | Sorted elements as stored joined with `", "`.      |
| `NormalizedAddressList() (AddressList, error)`     | Normalized, merged and sorted addresses joined with `", "`, or the error of an invalid address. |
| `GetOne() Address`                                 | Any element, `""` for an empty set.                |
| `String() string`                                  | Like `AddressList`, but `"<nil>"` for a nil set. Not normalized, because normalizing can merge elements. |
| `Add`, `AddSet`, `AddNormalized`, `AddAddressPart` | Pointer receivers that allocate the map, safe on a nil set. `AddNormalized` and `AddAddressPart` return the parse error. |
| `Insert`, `InsertAll`, `UnionWith`, `SymmetricDifferenceWith` | Value receivers, panic on a nil set when they have to add an element. |
| `Delete`, `DeleteAll`, `DeleteFunc`, `Clear`, `IntersectionWith`, `DifferenceWith` | Remove elements, safe on a nil set. `Delete*` report whether the set changed. |
| `Clone`, `Union`, `Intersection`, `Difference`, `SymmetricDifference` | New sets. `Clone` of nil is nil. `Difference` is asymmetric. |
| `Normalized() (AddressSet, error)`                 | New set of the normalized addresses, nil for nil, the receiver with the error if any address is invalid. |
| `Validate`, `Valid`, `ValidAndNormalized`          | Validation of all elements.                        |
| `Scan`, `Value`, `MarshalJSON`, `UnmarshalJSON`, `MarshalText`, `UnmarshalText` | See [Serialization](#serialization).               |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `xml.MarshalerAttr`, `fmt.Stringer`, `types.NormalizableValidator[AddressSet]`, the collections proposal set interface.

## NullableAddress

```go
type NullableAddress string
const AddressNull NullableAddress = ""
```

An `Address` where `""` is SQL NULL and JSON null. A value of only the runes that `strutil.TrimSpace` removes (whitespace, control and other non-printable runes, invalid UTF-8) is null too, like for `country.NullableCode`. Null counts as valid.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NullableAddressFrom(*mail.Address) NullableAddress` | Like `AddressFrom`, nil gives `AddressNull`.       |
| `NormalizedNullableAddress(str) (NullableAddress, error)` | `""` and values that are empty after `strutil.TrimSpace` give `AddressNull`, otherwise like `NormalizedAddress`. |
| `n.Parse() (*mail.Address, error)`                 | `nil, nil` for null.                               |
| `n.Normalized()`, `n.Valid()`, `n.Validate()`, `n.ValidAndNormalized()` | Like `Address`, null is valid and normalizes to null. |
| `n.NamePart()`, `n.AddressPart()`, `n.LocalPart()`, `n.DomainPart()` | Like `Address`, `""` for null.                     |
| `n.IsNull()`, `n.IsNotNull()`                      | Null is `""` or only runes that `strutil.TrimSpace` removes. |
| `n.Get() Address`                                  | Panics if null.                                    |
| `n.GetOr(def Address) Address`                     | Value or the default.                              |
| `n.ASCII()`                                        | Like `Address`, `""` for null.                     |
| `n.StringOr(nullString) string`                    | As stored, or the default if null.                 |
| `n.Set(Address)`, `n.SetNull()`                    | `Set("")` sets null.                               |
| `n.AsList() NullableAddressList`                   | List containing only `n`, the null list if `n` is null. |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `xml.Unmarshaler`, `xml.UnmarshalerAttr`, `xml.MarshalerAttr`, `strfmt.Scannable`, `JSONSchema()` (email string or null), `nullable.NullSetable[Address]`, `types.NormalizableValidator[NullableAddress]`. No `String` method, `fmt` prints the stored address, and no `UnmarshalText`, so `encoding/json` reads the keys of a `map[NullableAddress]V` as given.

## NullableAddressList

```go
type NullableAddressList string
```

An `AddressList` where `""` and whitespace only are the empty list, SQL NULL, and JSON null. Null counts as valid.

| Function / Method                                  | Description                                      |
|----------------------------------------------------|--------------------------------------------------|
| `NullableAddressListJoin(addrs ...Address)`, `NullableAddressListJoinStrings(addrs ...string)` | Like the `AddressList` joins.                    |
| `n.Append(addrs ...Address)`                       | Like `AddressList.Append`.                       |
| `n.Parse()`, `n.SplitNormalized()`                 | Lenient like `AddressList`, `nil, nil` for null. |
| `n.SplitRaw()`                                     | Like `AddressList.SplitRaw`, nil for null.       |
| `n.Normalized()`, `n.Validate()`, `n.ValidAndNormalized()` | Strict like `AddressList`, null is valid.        |
| `n.IsNull()`, `n.IsNotNull()`                      | Null is `""` or only runes that `strutil.TrimSpace` removes. |
| `n.Get() AddressList`                              | Panics if null.                                  |
| `n.GetOr(def AddressList)`                         | Value or the default.                            |
| `n.ASCII()`                                        | Like `AddressList`, `""` for null and for a list without addresses. |
| `n.StringOr(nullString)`                           | As stored, or the default if null.               |
| `n.Set(AddressList)`, `n.SetNull()`                | `Set("")` sets null.                             |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `xml.MarshalerAttr`, `nullable.NullSetable[AddressList]`. No `String` method, no `Valid()` and no `JSONSchema`.

## Message

`Message` holds the parsed headers, bodies, and attachments of one email. It has no custom JSON methods, the JSON keys come from the struct tags and empty fields are omitted.

| Field                 | Type                     | JSON key              | Source                                             |
|-----------------------|--------------------------|-----------------------|----------------------------------------------------|
| `ProviderID`          | `nullable.TrimmedString` | `providerID`          | ID at the mail provider (e.g. Gmail), not parsed from MIME |
| `InReplyToProviderID` | `nullable.TrimmedString` | `inReplyToProviderID` | `ProviderID` of the parent message                 |
| `ProviderLabels`      | `[]string`               | `providerLabels`      | Provider labels, not parsed from MIME              |
| `MessageID`           | `nullable.TrimmedString` | `messageID`           | `Message-Id` header                                |
| `InReplyTo`           | `nullable.TrimmedString` | `inReplyTo`           | `In-Reply-To` header                               |
| `References`          | `nullable.TrimmedString` | `references`          | `References` header                                |
| `Date`                | `*time.Time`             | `date`                | `Date` header, nil if missing or unparsable        |
| `From`                | `Address`                | `from`                | `From` header, normalized                          |
| `ReplyTo`             | `NullableAddress`        | `replyTo`             | `Reply-To` header, normalized                      |
| `To`                  | `AddressList`            | `to`                  | `To` headers, normalized, unparsable entries dropped |
| `DeliveredTo`         | `NullableAddressList`    | `deliveredTo`         | `Delivered-To` headers, like `To`                  |
| `Cc`                  | `NullableAddressList`    | `cc`                  | `Cc` headers, like `To`                            |
| `Bcc`                 | `NullableAddressList`    | `bcc`                 | `Bcc` headers, like `To`                           |
| `ExtraHeader`         | `Header`                 | `extraHeader`         | All other headers, RFC 2047 decoded                |
| `Subject`             | `string`                 | `subject`             | `Subject` header, trimmed                          |
| `Body`                | `string`                 | `body`                | Plain text body                                    |
| `BodyHTML`            | `nullable.TrimmedString` | `bodyHTML`            | HTML body                                          |
| `Attachments`         | `[]*Attachment`          | `attachments`         | Attachments, inline parts, other parts             |

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `type Header = textproto.MIMEHeader`               | Header map type of `ExtraHeader`.                  |
| `IsParsedHeader(key) bool`                         | True for the 11 headers with own fields (`Message-Id`, `In-Reply-To`, `References`, `Date`, `From`, `Reply-To`, `To`, `Delivered-To`, `Cc`, `Bcc`, `Subject`), case-insensitive. |
| `IsExtraHeader(key) bool`                          | `!IsParsedHeader(key)`.                            |
| `NewMessage(from, to, subject, body, bodyHTML) *Message` | New message with an allocated `ExtraHeader`.       |
| `msg.AddAttachment(partID, filename, content)`     | Appends `NewAttachment(partID, filename, content)`. |
| `msg.Recipients() []string`                        | Address parts of `To`, `Cc`, and `Bcc` in that order, de-duplicated. Unparsable entries are dropped without error. |
| `msg.ReplyToAddress() Address`                     | `ReplyTo` if set, else `From`.                     |
| `msg.ReferencesMessageIDs() []string`              | `References` split at whitespace and commas, empty entries skipped. |
| `msg.IsAutoSubmitted() bool`                       | `Auto-Submitted` header set and not `no` (RFC 3834). |
| `msg.AutoResponseSuppress() bool`                  | An `X-Auto-Response-Suppress` value is `DR`, `AutoReply`, or `All`. |
| `msg.ListID() string`, `msg.FeedbackID() string`   | `List-Id` and `Feedback-Id` from `ExtraHeader`.    |
| `msg.NewReplyMessage(from, replyText, replyHTML, keepAttachments, textTempl, htmlTempl)` | Builds a reply, see below.                         |
| `msg.BuildRawMessage() ([]byte, error)`            | Encodes the message as RFC 5322 MIME, see below.   |
| `msg.String() string`                              | Short debug description.                           |
| `ExtractBodyURLs(msg) strutil.StringSet`           | URLs starting with `http` found in `Body` and `BodyHTML`. |
| `HTMLToPlaintext(html []byte, delimiter string) (string, error)` | Trimmed, non-empty text nodes joined with `delimiter`. On a parse error it returns the text so far with the error. |

`NewReplyMessage` returns a message with `InReplyTo` set to the parent's `MessageID`, `InReplyToProviderID` to its `ProviderID`, `References` to the parent's References followed by its Message-ID (space separated), where a parent without References but with a single In-Reply-To ID contributes that ID, and `To` to the parent's `ReplyToAddress()`. The subject gets a `"Re: "` prefix unless it already starts with `Re:` in any case. `Body` is rendered from `textTempl` with `text/template`, `BodyHTML` from `htmlTempl` with `html/template`, the latter only if the parent has an HTML body or `replyHTML` is not empty. Both templates get a `ReplyTemplateData`:

```go
type ReplyTemplateData struct {
    Message                 // the parent message
    Date      string        // parent Date as RFC 1123Z, the current time if nil
    BodyLines []string      // parent Body split at "\n"
    BodyHTML  template.HTML // parent BodyHTML, or Body with "\n" replaced by "<br>"
    ReplyText string
    ReplyHTML template.HTML // replyHTML, or replyText with "\n" replaced by "<br>"
}
```

`BuildRawMessage` writes `text/plain` and `text/html` bodies (as `multipart/alternative` if both exist), attachments with `Inline` set as inline parts with their Content-ID inside `multipart/related`, and all other attachments with an attachment disposition inside `multipart/mixed`. The addresses are written normalized, with a non-ASCII display name RFC 2047 encoded without the address like `mail.Address.String()` does. If an address part is not ASCII, the names are written as UTF-8 and the MIME encoder encodes the whole header value, the address included. `To`, `Cc`, and `Bcc` are each written once with all addresses comma separated, `Delivered-To` once per address, then all `ExtraHeader` values.

## Parsing messages

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseMessage(data []byte) (*Message, error)`      | Detects the format: JSON if the data without surrounding whitespace starts with `{` and ends with `}`, TNEF if `ParseTNEFMessageBytes` succeeds, MIME otherwise. Empty data is an error. |
| `ParseMIMEMessage(r io.Reader) (*Message, error)`  | Parses MIME with [`jhillyerd/enmime`](https://github.com/jhillyerd/enmime). Returns a message with all usable data and the joined header errors, nil only if the envelope can't be read. |
| `ParseMIMEMessageBytes(data []byte) (*Message, error)` | `ParseMIMEMessage` of a byte slice.                |
| `ParseTNEFMessageBytes(data []byte) (*Message, error)` | Parses Microsoft TNEF (`winmail.dat`). Data shorter than 4 bytes is an error. |

How `ParseMIMEMessage` treats the headers:

- The address headers (`From`, `Reply-To`, `To`, `Cc`, `Bcc`, `Delivered-To`) are parsed raw, so only display names are RFC 2047 decoded and an encoded comma or `@` in a name can't add a recipient. A value that is only encoded-words, which MIME encoders write for a non-ASCII address part, is decoded as a whole. A name that can't be decoded is dropped, the address is kept. A second address is only taken from angle brackets: `attacker@evil.com ceo@corp.com` is an error, not `ceo@corp.com`.
- `From` is normalized. If it is missing or unparsable, `From` is `""` and an error is collected.
- `To` entries are parsed leniently. The parsable addresses are kept and an error is collected for the others.
- `Reply-To`, `Delivered-To`, `Cc`, and `Bcc` errors are ignored. The parsable addresses are kept, an unparsable `Reply-To` is dropped.
- `Date` is matched against a fixed list of layouts. On failure `Date` is nil and an error is collected.
- Every header without an own field goes to `ExtraHeader`, RFC 2047 decoded like display names. A value that fails to decode is kept as is.
- Attachments come first, then inline parts (`Inline` set), then other parts (`OtherPart` set). `ContentType` and `ContentID` are taken from the MIME part.

`ParseTNEFMessageBytes` sets `Body`, `BodyHTML`, and one attachment per TNEF attachment with `PartID` `TNEF0`, `TNEF1`, ... and `Filename` from the attachment title. `ContentType` and `ContentID` stay empty. If the plain body has at most one byte and the HTML body is longer, `Body` is set to the text converted from the HTML.

`ParseMIMEMessage` does not unpack a `winmail.dat` attachment. Pass its `Content` to `ParseTNEFMessageBytes` yourself.

## Attachment

```go
type Attachment struct {
    PartID      string `json:",omitempty"` // MIME part ID
    ContentID   string `json:",omitempty"` // Content-ID, referenced from HTML bodies
    ContentType string `json:",omitempty"` // MIME media type
    Inline      bool   `json:",omitempty"` // inline part referenced from the body
    OtherPart   bool   `json:",omitempty"` // neither attachment nor inline part
    Filename    string
    Content     []byte
}
```

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NewAttachment(partID, filename, content) *Attachment` | Sets `ContentID` to a random UUID v4 in hex and `ContentType` with `http.DetectContentType(content)`. |
| `a.String() string`                                | Short debug description with part ID, filename, and size. |

## Rule

`Rule` is a predicate over messages, for example to filter inbound mail.

| Type       | Description                                        |
|------------|----------------------------------------------------|
| `Rule`     | Interface with `AppliesToMessage(msg *Message) bool`. |
| `RuleFunc` | `func(*Message) bool` as `Rule`.                   |
| `BoolRule` | Constant result.                                   |
| `AllRule`  | `[]Rule` that applies if all rules apply. An empty `AllRule` applies. |
| `AnyRule`  | `[]Rule` that applies if any rule applies. An empty `AnyRule` does not apply. |

## Config

`ProviderDomains() map[string]struct{}` returns a new map of well-known consumer mail provider domains (`gmail.com`, `gmx.at`, `web.de`, ...) on every call, so callers can modify it. Use it with `DomainPart` of a normalized address to tell personal from business addresses.

## Serialization

`Address` and `NullableAddress` follow the rules of [Reading and writing values](../README.md#reading-and-writing-values): reading normalizes, `Value` refuses an invalid address, and `fmt` prints the stored address. The collections `AddressList`, `NullableAddressList` and `AddressSet` keep their addresses as stored in every format, because normalizing rewrites every entry and could merge two elements of a set. Their `Value` still refuses an invalid list or address.

| Type                  | `Value()` writes                                   | `Scan` accepts                                     | JSON                                               | Text and XML                                       |
|-----------------------|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|
| `Address`             | Normalized, blank as NULL, an invalid address is an error | `string` and `[]byte` normalized if valid, blank as `""`, else unchanged without error. NULL as `""` | `MarshalJSON` normalized, an invalid address unchanged, blank as `""`. `UnmarshalJSON` like `Scan`, `null` leaves the value unchanged | `MarshalText` like `MarshalJSON`. XML elements and attributes read like `Scan`. `ScanString` for `strfmt` reads like `Scan`, with `validate` an invalid or blank address is an error |
| `NullableAddress`     | Null (see `IsNull`) as NULL, else like `Address`   | Like `Address`, NULL and blank as null             | Null marshals as `null`, else like `Address`. `null` and blank unmarshal to null | `MarshalText` `""` for null, else like `Address`. XML like `Scan`. `ScanString` like `Address`, but `NULL`, `null`, `nil` and blank as null |
| `AddressList`         | The list as stored, trimmed. Blank as NULL, an invalid list is an error, also one with whitespace other than spaces, control characters or invalid UTF-8 that the parser ignores | A plain list as stored, trimmed. A PostgreSQL array or JSON array: the elements joined with `", "`. A JSON object: the keys sorted and joined. A JSON string: its content. NULL, `""` and whitespace only as `""` | `MarshalJSON` the list as stored as JSON string. `UnmarshalJSON`: a string as stored, an array or object joined like `Scan`, `null` leaves the value unchanged. Round-trips every list | `MarshalText` as stored. XML reads the plain string as given |
| `NullableAddressList` | Null (see `IsNull`) as NULL, else like `AddressList` | Like `AddressList`, NULL and blank as null         | Null marshals as `null`, else like `AddressList`. `null` unmarshals to null | `MarshalText` `""` for null, else as stored. `UnmarshalText` assigns the text as is |
| `AddressSet`          | nil: NULL. Empty: `{}`. Else the sorted PostgreSQL array `{"a@example.com","b@example.com"}` of the addresses as stored, valid ones trimmed. An invalid address is an error, also one with whitespace other than spaces, control characters or invalid UTF-8 that the parser ignores | NULL, `""` and whitespace only: nil set. A PostgreSQL array: elements trimmed, empty ones skipped, a `NULL` element or a nested array is an error. A JSON array: elements exactly. A JSON object: the keys. A plain string or a JSON string: split with `SplitAddressList`. `[]byte` like `string` | `MarshalJSON` a sorted array of the addresses as stored, also invalid ones, nil is `null`, empty `[]`. `UnmarshalJSON` an array exactly, an object's keys, a string split with `SplitAddressList`, `null` and `""` as nil. Round-trips every set | `MarshalText` the sorted addresses as stored joined with `", "`, valid ones trimmed, an error for an invalid address, one with runes that the parser ignores, or one that would not read back as one entry. `UnmarshalText` splits with `SplitAddressList`, also for XML |

The arrays written by `AddressSet` and read by the `Scan` methods of `AddressSet`, `AddressList`, and `NullableAddressList` need PostgreSQL or an array-compatible database like CockroachDB or YugabyteDB. `AddressList.Value` and `NullableAddressList.Value` write a plain string that works with any database, so write `set.AddressList()` to store a set in a plain string column. `AddressSet.Scan` reads it back.

## Gotchas

- **Lenient vs strict.** `ParseAddressList`, `AddressList.SplitNormalized`, and `NullableAddressList.SplitNormalized`/`Parse` return the valid entries plus an error. `Validate`, `Normalized`, `UniqueAddressParts`, `NormalizeAddressList`, `NormalizedAddressSet`, and `NormalizedAddressPartSet` fail on the first bad entry. `Message.Recipients` drops bad entries silently. `"a@example.com, not an address"` gives `[a@example.com]` and an error from `SplitNormalized`, and only an error from `Normalized`. `SplitRaw` and `SplitAddressList` don't parse at all and return `[a@example.com not an address]`.
- **Normalization is lossy.** The local part is lowercased (`John.Doe@Example.com` becomes `john.doe@example.com`). A quoted local part keeps its content and quotes: `"Doe, John"@Example.com` becomes `"doe, john"@example.com`, so it stays the same mailbox. A local part is written without quotes only if `net/mail` and this parser both read it unquoted, so `doe..john@x.com` becomes `"doe..john"@x.com` and `"'abc"@x.com` stays quoted. Display names lose quotes, backslashes and runes that are not graphic, like bidi overrides, because the parser reads no escapes: `"Say \"Hi\"" <a@b.com>` becomes `"Say Hi" <a@b.com>`. A display name equal to the address is dropped, `<a@b.com>` becomes `a@b.com`, and `mike@doe.com (Mike Doe)` becomes `"Mike Doe" <mike@doe.com>`.
- **Display names are quoted UTF-8.** `John Doe <j@example.com>` normalizes to `"John Doe" <j@example.com>`, so `ValidAndNormalized` is false for the unquoted form. A non-ASCII name stays UTF-8 (RFC 6532), also when the input was RFC 2047 encoded: `Jürgen Müller <J@X.de>` and `=?utf-8?q?J=C3=BCrgen_M=C3=BCller?= <j@x.de>` both normalize to `"Jürgen Müller" <j@x.de>`. Only `BuildRawMessage` and the `ASCII` methods RFC 2047 encode non-ASCII names, because MIME headers must be ASCII, and a name containing `=?` is always RFC 2047 encoded, also by `Normalized` and `MarshalText`, because the parser decodes encoded-words in names, also within quotes, and would read another name.
- **A top level domain of two or more ASCII letters is required.** `user@example` and IP literals like `user@[192.0.2.1]` are invalid.
- **`AddressRegexp` is not anchored.** `AddressRegexp.MatchString("xx a@b.com yy")` is true. Use `Address.Validate` to validate.
- **`DomainPart` does not parse.** `Address("John <J@Example.COM>").DomainPart()` is `Example.COM`. Call it on a normalized address before looking it up in `ProviderDomains()`.
- **Empty strings in SQL.** `AddressList.Scan` reads `""` and whitespace only as `""`, `AddressSet.Scan` as nil set, and the nullable types as null. `Value` writes them as SQL NULL, which a `NOT NULL` column rejects.
- **The collections refuse what the parser ignores.** The parser reads an address with line breaks and other whitespace turned into spaces, and without control and format characters (like NUL and bidi overrides) and invalid UTF-8. So `Validate` and `Valid` accept `"x\r\nBcc: evil@example.com" <a@example.com>`, but the `Value` methods of `AddressList`, `NullableAddressList` and `AddressSet` and `AddressSet.MarshalText`, which write the text as stored, return an error for it, also for a tab or a no-break space within the text. Store `Normalized()` instead, and build message headers from the parsed addresses, like `BuildRawMessage` does, not from the stored text.
- **The collections are not normalized.** `AddressList`, `NullableAddressList` and `AddressSet` keep `John@Example.com` and `john@example.com` as written, a set as two elements. Normalize with `Normalized` or `NormalizedAddressList` where you compare or deduplicate addresses.
- **Not for mailing lists with millions of addresses.** The address types and `ParseAddressList` are made for the recipients of a message, a form field or a database column. Every `Validate`, `Normalized` and `Value` call parses the whole list, and parsing millions of chained addresses like `<a@x.com> <b@x.com> ...` can exhaust the goroutine stack, which crashes the process. Store a huge list one address per row or as `[]string` and parse the addresses one at a time.
- **`AddressSet` iterates in random order.** `All` and `range` iterate in random order. Use `Sorted`, `Strings`, or `AddressList` for a stable order; `MarshalJSON`, `MarshalText` and `Value` sort.
- **A nil `AddressSet` panics on `Insert`, `InsertAll`, `UnionWith`, and `SymmetricDifferenceWith`**, like a nil map assignment. Use `Add` and `AddSet` on fields that may be nil.
- **Errors carry the raw message.** `ParseMessage`, `ParseMIMEMessage`, `ParseMIMEMessageBytes`, `ParseTNEFMessageBytes`, `NewReplyMessage`, and `BuildRawMessage` wrap errors with `errs.WrapWithFuncParams` from [go-errs](https://github.com/domonda/go-errs). `err.Error()` contains the call stack with the passed arguments, for the parse functions the raw message bytes, cut at `errs.FormatParamMaxLen` (5000 bytes by default) per argument and repeated for every wrapping function. This also applies to the joined header errors returned together with a usable message. `ParseAddressList` errors contain at most 100 bytes of the list from each failing entry on, for at most 100 errors. Don't log these errors where message content must not appear.
- **Named time zones in `Date`.** Zone abbreviations are parsed with `time.Parse`, which only knows `UTC` and the abbreviations of the local time zone. Others get offset zero: `Mon, 2 Jan 2006 15:04:05 EST` parses as `15:04:05 +0000 EST`.
- **`ParseMessage` falls back to MIME.** Data that is not detected as JSON or TNEF is parsed as MIME. A truncated JSON document like `{"subject":"x"` returns a mostly empty, non-nil message with a `From` error instead of a JSON syntax error.
- **`BuildRawMessage` writes `Bcc`.** Remove `Bcc` before encoding a message for the visible recipients. A missing `Date` is written as the current time, and an unparsable `From` or `ReplyTo`, an unparsable entry in `To`, `Cc`, `Bcc`, or `DeliveredTo`, or a nil attachment fails the whole build.
- **Reply templates are not escaped.** `ReplyTemplateData.BodyHTML` and `ReplyHTML` are `template.HTML`, so `html/template` inserts them as is, including a plain text body converted with `<br>`.
- **`NewReplyMessage` with `keepAttachments` shares** the parent's `Attachments` slice and its `*Attachment` values.
- **TNEF attachments have no `ContentType` or `ContentID`.** Only `NewAttachment` detects the content type and creates a Content-ID.

## Examples

Normalize an address and compare lenient with strict list parsing:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/email"
)

func main() {
	addr, err := email.NormalizedAddress("John Doe <John.Doe@Example.com>")
	if err != nil {
		panic(err)
	}
	fmt.Println(addr)
	part, _ := addr.AddressPart()
	fmt.Println(part)

	list := email.AddressList("a@example.com, not an address, B@Example.com")
	// Lenient: the valid entries are returned together with the error
	addrs, err := list.SplitNormalized()
	fmt.Println(addrs, err != nil)
	// Strict: any invalid entry is an error
	_, err = list.Normalized()
	fmt.Println(err != nil)

	// Output:
	// "John Doe" <john.doe@example.com>
	// john.doe@example.com
	// [a@example.com b@example.com] true
	// true
}
```

Collect recipients in a possibly nil `AddressSet` and write it as PostgreSQL array:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/email"
)

func main() {
	var set email.AddressSet // nil, Add allocates the map
	set.Add("b@example.com")
	err := set.AddAddressPart(`"Alice" <Alice@Example.com>`)
	if err != nil {
		panic(err)
	}
	fmt.Println(set.Sorted())

	value, err := set.Value()
	if err != nil {
		panic(err)
	}
	fmt.Println(value)

	// Output:
	// [alice@example.com b@example.com]
	// {"alice@example.com","b@example.com"}
}
```

Use the partial result of a message with a broken `To` header:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/email"
)

func main() {
	raw := "From: John Doe <John@Example.com>\r\n" +
		"To: jane@example.com, broken@\r\n" +
		"Subject: Invoice\r\n" +
		"\r\n" +
		"Please find the invoice attached.\r\n"

	msg, err := email.ParseMessage([]byte(raw))
	if msg == nil {
		panic(err)
	}
	fmt.Println(msg.From)
	fmt.Println(msg.To)
	fmt.Println(msg.Subject)
	fmt.Println("has error:", err != nil)

	// Output:
	// "John Doe" <john@example.com>
	// jane@example.com
	// Invoice
	// has error: true
}
```

## Related

- [`nullable`](../nullable/README.md): `nullable.TrimmedString` for the optional `Message` fields.
- [`strutil`](../strutil/README.md): `strutil.StringSet` returned by `ExtractBodyURLs`.
- [`charset`](../charset/README.md): encodings for message bodies outside UTF-8.
- [Null conventions](../docs/explanation-null-conventions.md): why `""` is the null value of `NullableAddress`.
- [Validation and normalization](../docs/explanation-validation-and-normalization.md): the `Valid`/`Validate`/`Normalized` pattern.
- [Lenient parsing](../docs/explanation-lenient-parsing.md): partial results with joined errors.
- [Store optional values](../docs/howto-store-optional-values.md): `NullableAddress` and `NullableAddressList` columns.
- [Extract identifiers from text](../docs/howto-extract-identifiers-from-text.md): `FindAllAddresses`.
