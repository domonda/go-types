// Package email provides lenient email address parsing and normalization,
// address lists and sets, and the parsing and building of email messages.
//
// The address types are:
//   - [Address]: one address with an optional display name,
//     parsed more leniently than [net/mail.ParseAddress] by [ParseAddress]
//   - [AddressList]: a comma separated list of addresses
//   - [AddressSet]: a set of addresses, stored in SQL as PostgreSQL text array
//   - [NullableAddress] and [NullableAddressList]: the empty string
//     is the null value and is written as SQL NULL and JSON null
//
// Address and NullableAddress are normalized when read and written, the
// collections AddressList, NullableAddressList and AddressSet are read and
// written as stored, because normalizing rewrites their entries and could
// merge two elements of a set. Value refuses an invalid address or list.
// The collections also store an unquoted display name as written, whose
// words some mail stacks read as recipients: Python's smtplib sends
// `admin, John <ceo@corp.com>` also to admin, so store Normalized, which
// quotes every name, where the text goes into a header of another system.
//
// Normalization lowercases the address part and formats the display name
// like [net/mail.Address.String], but keeps a non-ASCII name as UTF-8,
// so normalized addresses can be compared as strings. The ASCII methods
// and [Message.BuildRawMessage] encode non-ASCII names with RFC 2047. [ParseAddressList] and the SplitNormalized methods are lenient and
// return the valid addresses together with the joined parsing errors,
// while the Validate and Normalized methods fail on any invalid address.
// [SplitAddressList] and the SplitRaw methods split a list without
// parsing it.
//
// The address types are made for the recipients of a message, a form
// field or a database column, not for mailing lists with millions of
// addresses: every Validate, Normalized and Value call parses the whole
// list. Store such a list one address per row or as []string.
//
// [Message] holds the headers, bodies, and [Attachment] values of one email.
// [ParseMessage] detects and parses JSON, Microsoft TNEF (winmail.dat),
// and MIME data, [Message.BuildRawMessage] encodes a message as MIME,
// and [Message.NewReplyMessage] creates a reply from templates.
// [Rule] implementations filter messages.
//
// Errors returned by the message parsing functions are wrapped with
// github.com/domonda/go-errs and include the passed raw message data
// as function parameters in their Error string.
//
// See the README.md file of the package for the complete API reference,
// the serialization rules, and the known pitfalls.
package email
