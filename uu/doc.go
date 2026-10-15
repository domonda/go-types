// Package uu provides UUID types with database/sql, encoding/json and
// text support. It is derived from github.com/satori/go.uuid.
//
// The main types are:
//   - [ID], a UUID as [16]byte. Its zero value [IDNil] is not valid,
//     see [ID.Validate].
//   - [NullableID], a UUID whose zero value [IDNull] is SQL NULL and JSON null.
//   - [IDSlice], an ordered list of IDs for PostgreSQL uuid[] columns.
//   - [IDSet], a set of IDs as map[ID]struct{} with the set algebra of
//     package github.com/domonda/go-types/mapset.
//   - [IDMutex], one mutex per locked ID.
//
// # Generating IDs
//
// [NewID] generates IDs with [IDvDefault], which is [IDv7]: version 7 UUIDs
// with the big-endian Unix time in milliseconds in the first 48 bits as
// specified by RFC 9562, so they sort by creation time as bytes and as
// strings. A test can make NewID deterministic by putting a generator like
// [IDv7DeterministicFunc] into the context with [ContextWithIDFunc].
// [IDv1] to [IDv5] generate the other RFC versions.
//
// # Null and empty values
//
// [IDNil] is written as "00000000-0000-0000-0000-000000000000" to SQL and
// JSON, and [ID.Scan] returns an error for SQL NULL. Use [NullableID] for
// nullable columns: [IDNull] is written as SQL NULL and JSON null.
// A nil [IDSlice] or [IDSet] is SQL NULL and JSON null, an empty non-nil
// one is the empty array '{}' and [].
//
// The parse functions accept the canonical, hex, base64 and URN formats
// but don't validate the version or variant. [IDMust] and [IDFromAny] do.
//
// See the [README] for tables of all functions and the exact
// serialization rules.
//
// [README]: https://github.com/domonda/go-types/blob/master/uu/README.md
package uu
