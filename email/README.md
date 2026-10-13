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
| `ParseAddress(str) (*mail.Address, error)`         | Lenient parser. Lowercases the address part, decodes RFC 2047 names in the charsets of `golang.org/x/net/html/charset` (ISO 8859-2, Windows-1250, ...), accepts quoted names, `addr (Name)` comments, and umlauts. A name equal to the address is dropped. |
| `NormalizedAddress(str) (Address, error)`          | `Address(str).Normalized()`.                       |
| `AddressFrom(*mail.Address) Address`               | `""` for nil, the bare address without a name, else `mail.Address.String()` (`"Name" <addr>`). |
| `FindAllAddresses(text) []Address`                 | All matches of `AddressRegexp` in text, in order, not normalized. |
| `UniqueNormalizedAddressSlice([]Address) []Address` | New sorted slice of the unique address parts. Skips invalid entries, returns nil if none is valid, does not modify the input. |
| `AddressRegexp`                                    | Compiled, unanchored regular expression for an address without name. |
| `a.Parse() (*mail.Address, error)`                 | `ParseAddress(string(a))`.                         |
| `a.Normalized() (Address, error)`                  | Parsed and re-formatted with `AddressFrom`. Returns `a` unchanged with the error. |
| `a.Validate() error` / `a.Valid() bool`            | Whether `Parse` succeeds.                          |
| `a.ValidAndNormalized() bool`                      | Valid and equal to its normalized form.            |
| `a.NamePart() (string, error)`                     | Decoded display name, `""` if there is none.       |
| `a.AddressPart() (Address, error)`                 | Normalized `local@domain` without the display name. |
| `a.AddressPartString() (string, error)`            | Same as `AddressPart` as string.                   |
| `a.LocalPart() (string, error)`                    | Normalized part before the `@`.                    |
| `a.DomainPart() string`                            | Raw text after the last `@` with trailing `>` and whitespace trimmed, `""` without `@`. Not parsed or lowercased. |
| `a.AsList() AddressList`                           | List containing only `a`.                          |
| `a.Nullable() NullableAddress`                     | Type conversion.                                   |

Implements: `types.NormalizableValidator[Address]`, `JSONSchema()` (`string`, format `email`). No JSON or SQL methods, it is encoded as a plain string.

## AddressList

```go
type AddressList string // "a@example.com, John <j@example.com>"
```

A comma separated list. `Validate` and `Valid` accept the empty list, only `Scan` rejects it. Use `NullableAddressList` for lists that can be empty.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseAddressList(list) ([]*mail.Address, error)`  | Lenient. Skips a malformed entry up to the next comma and keeps parsing, returns all parsed addresses together with an `errors.Join` error. `""` and the `undisclosed-recipients` / `withheld-recipients` placeholders (also with a space instead of the dash) give nil without error. |
| `NormalizeAddressList(list) ([]string, error)`     | Strict. Normalized addresses, de-duplicated by address part, first spelling wins. Any bad entry returns `nil, err`, an empty list `nil, nil`. |
| `AddressListJoin(addrs ...Address) AddressList`    | Joins with `", "`, skips empty addresses.          |
| `AddressListJoinStrings(addrs ...string) AddressList` | Same for strings.                                  |
| `l.Append(addrs ...Address) AddressList`           | Appends with `", "`, skips empty addresses.        |
| `l.Parse() ([]*mail.Address, error)`               | `ParseAddressList(string(l))`, lenient.            |
| `l.Split() ([]Address, error)`                     | Lenient. Parsed addresses as `Address` values (via `AddressFrom`), nil for an empty list. |
| `l.Normalized() (AddressList, error)`              | Strict. Every address normalized and re-joined with `", "`. Does not de-duplicate. Returns `l` unchanged with the error. |
| `l.UniqueAddressParts() (AddressSet, error)`       | Strict. Set of the address parts, `nil, nil` for an empty list. |
| `l.Validate() error` / `l.Valid() bool`            | Strict, any bad entry is an error.                 |
| `l.ValidAndNormalized() bool`                      | Valid and equal to its normalized form.            |
| `l.Nullable() NullableAddressList`                 | Type conversion.                                   |
| `l.Scan(value) error`                              | See [Serialization](#serialization).               |

Implements: `sql.Scanner`, `types.NormalizableValidator[AddressList]`. No `driver.Valuer` and no JSON methods, it is written as a plain string.

## AddressSet

```go
type AddressSet map[Address]struct{}
```

A set of addresses, stored in SQL as a PostgreSQL text array. It implements the abstract set interface of the Go 1.28 collections proposal ([go.dev/issue/80590](https://go.dev/issue/80590)).

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
| `AddressList() AddressList`                        | Sorted elements joined with `", "`.                |
| `GetOne() Address`                                 | Any element, `""` for an empty set.                |
| `String() string`                                  | Like `AddressList`, but `"<nil>"` for a nil set.   |
| `Add`, `AddSet`, `AddNormalized`, `AddAddressPart` | Pointer receivers that allocate the map, safe on a nil set. `AddNormalized` and `AddAddressPart` return the parse error. |
| `Insert`, `InsertAll`, `UnionWith`, `SymmetricDifferenceWith` | Value receivers, panic on a nil set when they have to add an element. |
| `Delete`, `DeleteAll`, `DeleteFunc`, `Clear`, `IntersectionWith`, `DifferenceWith` | Remove elements, safe on a nil set. `Delete*` report whether the set changed. |
| `Clone`, `Union`, `Intersection`, `Difference`, `SymmetricDifference` | New sets. `Clone` of nil is nil. `Difference` is asymmetric. |
| `Normalized() (AddressSet, error)`                 | New set of the normalized addresses, nil for nil, the receiver with the error if any address is invalid. |
| `Validate`, `Valid`, `ValidAndNormalized`          | Validation of all elements.                        |
| `Scan`, `Value`                                    | See [Serialization](#serialization).               |

Implements: `sql.Scanner`, `driver.Valuer`, `fmt.Stringer`, `types.NormalizableValidator[AddressSet]`, the collections proposal set interface. No JSON methods.

## NullableAddress

```go
type NullableAddress string
const AddressNull NullableAddress = ""
```

An `Address` where `""` is SQL NULL and JSON null. A whitespace only value is null too, like for `country.NullableCode`. Null counts as valid.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `NullableAddressFrom(*mail.Address) NullableAddress` | Like `AddressFrom`, nil gives `AddressNull`.       |
| `NormalizedNullableAddress(str) (NullableAddress, error)` | `""` and whitespace only give `AddressNull`, otherwise like `NormalizedAddress`. |
| `n.Parse() (*mail.Address, error)`                 | `nil, nil` for null.                               |
| `n.Normalized()`, `n.Validate()`, `n.ValidAndNormalized()` | Like `Address`, null is valid and normalizes to null. |
| `n.NamePart()`, `n.AddressPart()`, `n.LocalPart()`, `n.DomainPart()` | Like `Address`, `""` for null.                     |
| `n.IsNull()`, `n.IsNotNull()`                      | Null is `""` or whitespace only.                   |
| `n.Get() Address`                                  | Panics if null.                                    |
| `n.GetOr(def Address) Address`, `n.StringOr(nullString) string` | Value or the default.                              |
| `n.Set(Address)`, `n.SetNull()`                    | `Set("")` sets null.                               |
| `n.AsList() NullableAddressList`                   | List containing only `n`, the null list if `n` is null. |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextUnmarshaler`, `JSONSchema()` (email string or null), `nullable.NullSetable[Address]`. There is no `Valid()` method, use `Validate() == nil`.

## NullableAddressList

```go
type NullableAddressList string
```

An `AddressList` where `""` is the empty list, SQL NULL, and JSON null. Null counts as valid.

| Function / Method                                  | Description                                      |
|----------------------------------------------------|--------------------------------------------------|
| `NullableAddressListJoin(addrs ...Address)`, `NullableAddressListJoinStrings(addrs ...string)` | Like the `AddressList` joins.                    |
| `n.Append(addrs ...Address)`                       | Like `AddressList.Append`.                       |
| `n.Parse()`, `n.Split()`                           | Lenient like `AddressList`, `nil, nil` for null. |
| `n.Normalized()`, `n.Validate()`, `n.ValidAndNormalized()` | Strict like `AddressList`, null is valid.        |
| `n.IsNull()`, `n.IsNotNull()`                      | `n == ""`, `n != ""`.                            |
| `n.Get() AddressList`                              | Panics if null.                                  |
| `n.GetOr(def AddressList)`, `n.StringOr(nullString)` | Value or the default.                            |
| `n.Set(AddressList)`, `n.SetNull()`                | `Set("")` sets null.                             |

Implements: `sql.Scanner`, `driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextUnmarshaler`, `nullable.NullSetable[AddressList]`. No `Valid()` and no `JSONSchema`.

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

`BuildRawMessage` writes `text/plain` and `text/html` bodies (as `multipart/alternative` if both exist), attachments with `Inline` set as inline parts with their Content-ID inside `multipart/related`, and all other attachments with an attachment disposition inside `multipart/mixed`. `From`, `Reply-To`, `To`, `Cc`, and `Bcc` are written normalized, so a non-ASCII display name is RFC 2047 encoded without the address. `To`, `Cc`, and `Bcc` are each written once with all addresses comma separated, `Delivered-To` once per address, then all `ExtraHeader` values.

## Parsing messages

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `ParseMessage(data []byte) (*Message, error)`      | Detects the format: JSON if the data without surrounding whitespace starts with `{` and ends with `}`, TNEF if `ParseTNEFMessageBytes` succeeds, MIME otherwise. Empty data is an error. |
| `ParseMIMEMessage(r io.Reader) (*Message, error)`  | Parses MIME with [`jhillyerd/enmime`](https://github.com/jhillyerd/enmime). Returns a message with all usable data and the joined header errors, nil only if the envelope can't be read. |
| `ParseMIMEMessageBytes(data []byte) (*Message, error)` | `ParseMIMEMessage` of a byte slice.                |
| `ParseTNEFMessageBytes(data []byte) (*Message, error)` | Parses Microsoft TNEF (`winmail.dat`). Data shorter than 4 bytes is an error. |

How `ParseMIMEMessage` treats the headers:

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

None of the `Scan` or `Unmarshal*` methods validate or normalize.

| Type                  | `Value()` writes                                   | `Scan` accepts                                     | JSON                                               | Text                                               |
|-----------------------|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|----------------------------------------------------|
| `Address`             | No `Valuer`, `database/sql` writes the string, `""` as `""` | No `Scanner`, `database/sql` assigns `string` and `[]byte`, NULL is an error | Plain string, `""` as `""`                         | No methods                                         |
| `AddressList`         | No `Valuer`, `database/sql` writes the string      | `string` or `[]byte`: a `{...}` PostgreSQL array literal is decoded, elements trimmed, empty ones skipped, joined with `", "` (so `{}` gives `""`). Any other string trimmed. NULL, `""` and whitespace only are errors. | Plain string                                       | No methods                                         |
| `AddressSet`          | nil: NULL. Empty: `{}`. Else sorted array literal `{"a@example.com","b@example.com"}` | NULL: nil set. `{...}` array literal: elements trimmed, empty ones skipped, a `NULL` element or a nested array is an error. Any other string: a single element set, trimmed. `""` and whitespace only are errors. `[]byte` like `string`. | No methods: nil is `null`, a set is a JSON object `{"a@example.com":{}}`, which also unmarshals | No methods                                         |
| `NullableAddress`     | Null (`""` or whitespace only): NULL. Else the string | NULL, `""` and whitespace only: null. Any other `string` or `[]byte` as is. | Null marshals as `null`. `null` and `""` unmarshal to null | `UnmarshalText` assigns the text as is, no `MarshalText` |
| `NullableAddressList` | `""`: NULL. Else the string                        | NULL, `""` and whitespace only: null. Other `string` and `[]byte` like `AddressList`. | `""` marshals as `null`. `null` and `""` unmarshal to null | `UnmarshalText` assigns the text as is, no `MarshalText` |

The array literals written by `AddressSet` and read by the `Scan` methods of `AddressSet`, `AddressList`, and `NullableAddressList` need PostgreSQL or an array-compatible database like CockroachDB or YugabyteDB. `NullableAddressList.Value` writes a plain string that works with any database.

## Gotchas

- **Lenient vs strict.** `ParseAddressList`, `AddressList.Split`, and `NullableAddressList.Split`/`Parse` return the valid entries plus an error. `Validate`, `Normalized`, `UniqueAddressParts`, `NormalizeAddressList`, `NormalizedAddressSet`, and `NormalizedAddressPartSet` fail on the first bad entry. `Message.Recipients` drops bad entries silently. `"a@example.com, not an address"` gives `[a@example.com]` and an error from `Split`, and only an error from `Normalized`.
- **Normalization is lossy.** The local part is lowercased (`John.Doe@Example.com` becomes `john.doe@example.com`). Quotes in the local part are removed and spaces and commas become dots: `"Doe, John"@example.com` becomes `doe..john@example.com` (which `net/mail` writes quoted). A display name equal to the address is dropped, `<a@b.com>` becomes `a@b.com`, and `mike@doe.com (Mike Doe)` becomes `"Mike Doe" <mike@doe.com>`.
- **Display names follow `net/mail`.** ASCII names are quoted: `John Doe <j@example.com>` normalizes to `"John Doe" <j@example.com>`, so `ValidAndNormalized` is false for the unquoted form. Non-ASCII names are RFC 2047 encoded: `Jürgen Müller <J@X.de>` normalizes to `=?utf-8?q?J=C3=BCrgen_M=C3=BCller?= <j@x.de>`.
- **A top level domain of two or more ASCII letters is required.** `user@example` and IP literals like `user@[192.0.2.1]` are invalid.
- **`AddressRegexp` is not anchored.** `AddressRegexp.MatchString("xx a@b.com yy")` is true. Use `Address.Validate` to validate.
- **`DomainPart` does not parse.** `Address("John <J@Example.COM>").DomainPart()` is `Example.COM`. Call it on a normalized address before looking it up in `ProviderDomains()`.
- **Empty strings in SQL.** `AddressList.Scan` and `AddressSet.Scan` reject `""` and whitespace only strings, `NullableAddress.Scan` and `NullableAddressList.Scan` read them as null. Store NULL instead of `""` (the nullable `Value` methods do).
- **`AddressSet` is a plain map for JSON and iteration.** `json.Marshal` writes `{"a@example.com":{}}`, `All` and `range` iterate in random order. Use `Sorted`, `Strings`, or `AddressList` for a stable order.
- **A nil `AddressSet` panics on `Insert`, `InsertAll`, `UnionWith`, and `SymmetricDifferenceWith`**, like a nil map assignment. Use `Add` and `AddSet` on fields that may be nil.
- **Errors carry the raw message.** `ParseMessage`, `ParseMIMEMessage`, `ParseMIMEMessageBytes`, `ParseTNEFMessageBytes`, `NewReplyMessage`, and `BuildRawMessage` wrap errors with `errs.WrapWithFuncParams` from [go-errs](https://github.com/domonda/go-errs). `err.Error()` contains the call stack with the passed arguments, for the parse functions the raw message bytes, cut at `errs.FormatParamMaxLen` (5000 bytes by default) per argument and repeated for every wrapping function. This also applies to the joined header errors returned together with a usable message. `ParseAddressList` errors contain the whole list. Don't log these errors where message content must not appear.
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
	addrs, err := list.Split()
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
