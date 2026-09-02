# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
No version has been tagged yet, so releases are dated rather than numbered.
See [TODOS.md](./TODOS.md) for what still blocks a `v1.0.0` cut, including
picking the semver baseline.

## [Unreleased]

## 2026-09-02

### Added

- `strfmt.ScanConfig.StrictEmptyStringParsing` controls what a nil source
  string means for a destination that can neither hold nil, be set to null,
  nor hold the source string itself, like a number, a bool or `time.Time`.
  When `false` (the default) such a destination is assigned its zero value,
  so a spreadsheet row with empty cells stays scannable into a column of any
  type. When `true` it is an error, unless the type has a `ScanString` or
  `UnmarshalText` method that gives a nil source string its own meaning.
- `strfmt.Scan` sets a destination with a `SetNull()` method to null for a
  nil source string, so every nullable type behaves the same way for an
  empty cell without its scanning method having to repeat it.
- `strfmt.Scan` supports multiple levels of pointer indirection.
- A `nullable.Time` scanner is registered by default in `strfmt.NewScanConfig`,
  so it parses every configured `TimeFormats` layout instead of only the
  RFC 3339 format that its `UnmarshalText` method accepts.
- `strfmt.Scan` accepts a `time.Duration` without a unit suffix as plain
  nanoseconds, which is how a duration is represented by its underlying
  `int64` kind and how it is marshalled as JSON.

- **`mapset` package** — set operations on `map[K]struct{}` and `map[K]bool`,
  mirroring the `container/mapset` package proposed for Go 1.28
  ([go.dev/issue/77052](https://go.dev/issue/77052), CL 724420) function for
  function. It is the shared implementation behind the set methods of
  `types.Set[T]`, `uu.IDSet`, `strutil.StringSet` and `email.AddressSet`, and
  can be replaced by the standard library package with an import path change
  once Go 1.28 ships.
- The methods of the abstract set interface of the Go collections proposal
  ([go.dev/issue/80590](https://go.dev/issue/80590)) on all four set types:
  `All`, `ContainsAll`, `DeleteAll`, `DeleteFunc`, `Difference`,
  `DifferenceWith`, `Insert`, `InsertAll`, `Intersection`, `IntersectionWith`,
  `Intersects`, `SymmetricDifference`, `SymmetricDifferenceWith`, `Union` and
  `UnionWith`. They now behave like the `container/set.Set` type planned for
  Go 1.28, which is likewise represented as `map[E]struct{}`.
- `strutil.StringSet.Len` and `email.AddressSet.Equal`, which were missing.
- `ContainsAny` on `uu.IDSet` and `email.AddressSet`, `IsEmpty` on
  `strutil.StringSet` and `Sorted` on `uu.IDSet`, so that all four set types
  now offer `ContainsAny`, `ContainsAll`, `IsEmpty`, `Len` and `Sorted` under
  the same names. `strutil.StringSet` still has no `IsNull`, being the one
  set type that does not distinguish a nil from an empty set.
- `internal/collections` holding the proposal's abstract interfaces, and
  `internal/collections/settest` holding the abstract set specification as a
  reusable test. Every set type is checked against that specification and by a
  compile-time interface assertion, so a behaviour change in `mapset` is caught
  at each type that delegates to it, not only in `mapset` itself.
- `nullable.SplitArrayValues` and `notnull.SplitArrayValues`, which split an
  SQL or JSON array into its top level elements like `SplitArray` and return
  the value of every double quoted string element with the quotes removed and
  the escape sequences of the parsed syntax undone. `SplitArray` returns the
  raw text of its elements, which is required for a JSON array of objects but
  is the wrong thing for a string array, and because it was the only thing on
  offer every known caller hand-rolled the unquoting and got it wrong. SQL and
  JSON unescape differently — a backslash escapes the following character
  whatever it is in a PostgreSQL array literal, while JSON interprets `\n`,
  `\t` and `\uXXXX` — so the new functions unescape with the syntax they
  parsed, and a PostgreSQL element is unescaped whether it is quoted or not.
  Elements that are not double quoted strings, like the objects of a JSON array
  of objects, are returned unchanged. A quoted element of a JSON array that is
  not a valid JSON string is an error, because returning it unchanged would
  hand out the value with the quotes these functions exist to remove, which is
  not distinguishable from a value that really has them.

### Fixed

- `strfmt.FormatValue` returned the text of `encoding.TextMarshaler` only when
  `MarshalText` returned an **error**, and fell through to the generic
  conversions when it succeeded. The condition was inverted; a successfully
  marshalled text is now used, as the documentation always stated.
- `strfmt.Scan` silently truncated an integer that did not fit the destination
  type: `"300"` scanned into an `int8` assigned `44`. Integers are now parsed
  with the bit size of the destination type, and floats are range checked, so
  both report an out-of-range value as an error.
- `strfmt.Scan` panicked inside the `reflect` package for a destination that is
  invalid or not settable, like the value of an unexported struct field. Those
  are reported as errors now. A pointer that is not settable itself, like
  `reflect.ValueOf(&x)`, is scanned by scanning the value it points to.
- `strfmt.Scan` replaced an already allocated pointer destination with a newly
  allocated value, silently dropping every field the scanning method did not
  assign and leaving other holders of the original pointer with a stale value.
  An allocated pointer is now scanned in place and keeps its identity.
- `strfmt.Scan` could leave a destination partially assigned when it returned
  an error, because a scanning method can assign before it fails and
  `ValidateFunc` runs after the value was assigned. A failed scan no longer
  modifies the destination at all.
- `strfmt.Scan` used the scanner registered for a type only for the type
  itself, so a `*time.Time` destination fell through to the RFC 3339 only
  `UnmarshalText` method of `time.Time`. A pointer destination now parses the
  same strings as its value equivalent.
- `strfmt.FormatValue` panicked in `reflect.Value.Interface` for the value of
  an unexported struct field. An addressable one is now formatted like an
  exported field, so a reflection driven caller walking a struct gets
  `2026-01-02T15:04:05Z` instead of a crash or `{0 63902963045 <nil>}`.
- `email.AddressSet`, `email.AddressList` and `email.NullableAddressList` kept
  the quotes of a quoted PostgreSQL array element, so scanning the `{"a@x.com"}`
  literal that `AddressSet.Value` itself writes produced the address
  `"a@x.com"` including the quote characters and `Value` did not survive `Scan`.
  All three now decode the array with the same PostgreSQL array codec that
  `Value` writes with, which also unescapes an element containing a comma or an
  escaped quote. An SQL `NULL` array element is an error now instead of the
  address `NULL`.

  Two deliberate limits of that decoding: every element is trimmed, so a member
  stored with leading or trailing whitespace (which `Validate` rejects anyway)
  does not survive a `Value`/`Scan` round trip, and an element that is empty
  after trimming is dropped. The trim is what keeps a hand-written literal like
  `{ a@x.com , b@x.com }` working, which the previous decoder accepted.
- **Breaking:** `nullable.SQLArrayLiteral` and `notnull.SQLArrayLiteral` escaped
  the quote but not the backslash of an element, so the literal `{"a\b"}`
  written for the value `a\b` is read back by PostgreSQL as `ab`: a silent data
  loss for every value containing a backslash. Both characters are escaped now,
  and a test pins the literal to be byte identical to the one written by the
  vendored `pq.StringArray.Value` and to round trip back through its parser. The
  literal was additionally checked against a live PostgreSQL 16 during
  development, which the test suite itself does not do, having no database.
  Nothing about the changed output produces a compile error: review stored
  literals, golden tests and anything comparing the output. A caller that pairs
  `SQLArrayLiteral` with `SplitArray` and strips the quotes itself has to move
  to `SplitArrayValues` now: the literal written for `a\b` is `{"a\\b"}`
  instead of `{"a\b"}`, so stripping only the quotes yields `a\\b`. It fails
  silently, which is the same trap that made this function worth fixing.
- **Breaking:** `SplitArray` reported an unclosed quote for a quoted element
  whose value ends with a backslash, like the `{"a\\"}` literal PostgreSQL
  outputs for the value `a\`, because it took the escaped backslash before the
  closing quote for an escape of that quote. Escape sequences are tracked now
  instead of looking at the previous character only. The tracking cuts the other
  way for a malformed literal that PostgreSQL never emits: `{"a\\"","b"}` used
  to split into two elements and is an `invalid rune` error now, which is what
  PostgreSQL reports for it too.
- **Breaking:** the `SplitArray` functions honoured the backslash escaping of a
  PostgreSQL array literal in a quoted element only, so the literal `{a\,b}`,
  which PostgreSQL reads as the single value `a,b`, was split into the two
  elements `a\` and `b`. A backslash escapes the following character in an
  unquoted element too, verified against PostgreSQL 16, and is now undone by
  `SplitArrayValues` for every element of an SQL array. Note that the vendored
  `lib/pq` parser behind `StringArray.Scan` implements the escaping for quoted
  elements only, so it still reads `{a\,b}` as two elements; `SplitArrayValues`
  follows PostgreSQL, not `lib/pq`. A JSON array is unaffected, a backslash
  outside a quoted string is not JSON syntax.
- **Breaking:** the `SplitArray` functions dropped the last element of a literal
  with a trailing comma: `{a,}` returned `{"a"}` and lost the element the comma
  announced without reporting anything. That comma now yields an empty string as
  last element. PostgreSQL rejects the literal outright, this parser is the
  tolerant one and reports what it read rather than silently returning a shorter
  array. An empty element between two commas, `{a,,b}`, stays an error.
- The `SplitArray` functions counted the `{`, `[`, `}` and `]` of a nested
  object or array without noticing that those runes can appear inside a string
  value of it, so a JSON array of objects failed to split with a `has too many
  '}'` error as soon as one of its string values contained a closing brace or
  bracket, like `[{"a":"}"}]`. Nesting deeper than one level failed the same
  way, because only the opening rune of an element was counted and never the
  ones within it, so `[{"a":{"b":1}}]` and `[[1,[2]]]` were rejected too. Both
  are tracked now, like the escapes of a quoted element. Every well-formed
  literal that split before splits into the same elements; only malformed input
  whose braces happened to balance out, like `[{a{b}]`, is an error now. An
  unmatched `{` or `[` at the top level of an unquoted element stays literal
  text, so a hand written `{in{o@example.com}` still parses.
- The `SplitArray` functions could not tell the value of a quoted string
  element, which is what `SplitArrayValues` was added for. Their documentation
  now says so, and also that an unquoted `NULL` element of an SQL array (`null`
  in a JSON array) is returned as the string `NULL` (`null`) and is
  indistinguishable from a quoted `"NULL"` (`"null"`) element for
  `SplitArrayValues`. `SplitArray` keeps the quotes that tell the two apart and
  its elements correspond to those of `SplitArrayValues` by index.

### Changed

- `strfmt.Scan` resolves a nil source string (as reported by `ScanConfig.IsNil`,
  now compared with surrounding whitespace trimmed) before dispatching to a
  `Scannable` or `encoding.TextUnmarshaler` destination, so that every type gets
  the same handling instead of every scanning method repeating it. A `Scanner`
  registered in `ScanConfig.TypeScanners` keeps taking precedence over all
  built-in scanning logic, including this handling, so it can give a nil source
  string a meaning of its own. The scanners registered by `NewScanConfig` apply
  the standard handling themselves.
- **Breaking:** with the default `StrictEmptyStringParsing == false`, scanning
  an empty or `"NULL"` cell into a number, a bool, `time.Time` or a
  `Scannable` type such as `money.Amount` assigns the zero value instead of
  returning a parse error. Set `StrictEmptyStringParsing = true` to restore the
  previous behavior.
- **Breaking:** `strfmt.FormatValue` now prefers `encoding.TextMarshaler` over
  the conversions based on the value's kind, as documented. Types with a
  registered `FormatConfig.TypeFormatters` entry are unaffected, since those
  are still resolved first. Observable for `notnull.TrimmedString` and
  `nullable.TrimmedString`, which now format their normalized text.
- A value assigned for a nil source string is not passed to
  `ScanConfig.ValidateFunc`, because the absence of a value is not a value to
  validate. A `string` destination assigned the source string is validated
  like any other scanned value.

- **Breaking:** the minimum Go version is now 1.26.0, raised from 1.25.0, in
  `go.mod` and `tools/go.mod`; CI builds, tests and gosec run on 1.26. This
  lets `internal/collections` declare the abstract interfaces with the
  self-referential constraint the collections proposal specifies, as in
  `Collection[E any, C Collection[E, C]]`, which the Go 1.25 type checker
  rejected with "invalid recursive type".
- **Breaking:** `types.Set.Difference` now returns the asymmetric difference
  (the elements of the receiver that are not in the argument), as specified by
  the collections proposal. It previously returned the *symmetric* difference,
  which is now `types.Set.SymmetricDifference`. Like the array literal and
  parser changes above, this does not produce a compile error — review any
  call site that relies on the old meaning.
- **Breaking:** `types.Set.ContainsAll` takes an `iter.Seq[T]` instead of
  variadic values. Use `set.ContainsAll(slices.Values(vals))`.
- `types.Set.UnmarshalJSON` had a dead branch: it assigned nil for JSON
  `null` without returning, then fell through to `json.Unmarshal` and
  re-allocated in the following nil check, so the assignment could never be
  observed and `null` produced an allocated empty set. It now returns
  directly, so `null` yields the nil set as its documentation implies.
- **Fixed:** a set-map collapsed the nil and the empty set into the same
  wire value. `uu.IDSet.Value` wrote SQL `NULL` for an allocated empty set
  instead of the empty array `{}`, so storing an empty set silently nulled
  the column; `uu.IDSet.MarshalJSON` wrote JSON `null` for it instead of `[]`;
  and `types.Set.MarshalJSON` did the same. All three delegated to
  `AsSortedSlice`/`Sorted`, which return a nil slice for an empty set.
  `email.AddressSet.Value` already got this right.

  The nil, empty and populated set are now three distinct wire values, in
  both directions:

  | Go value      | SQL         | JSON        | Types                          |
  |---------------|-------------|-------------|--------------------------------|
  | nil map       | `NULL`      | `null`      | SQL: `uu.IDSet`,               |
  | empty map     | `{}`        | `[]`        | `email.AddressSet`.            |
  | populated map | `{"a","b"}` | `["a","b"]` | JSON: `uu.IDSet`, `types.Set`. |

  Only the types listed have the codec in question. `email.AddressSet` and
  `strutil.StringSet` declare no `MarshalJSON`, so `encoding/json` uses Go's
  default map encoding for them (`{"a@example.com":{}}`) — unchanged by this
  release, but do not assume the JSON column of the table above applies to
  them. `types.Set` and `strutil.StringSet` have no SQL codec at all.

  `null`/`NULL` yields a nil map, which panics when inserted into exactly
  like any nil Go map: nil-check before `Insert`, or use the pointer-receiver
  `Add` where one exists.
- `email.AddressSet.String` returns `"<nil>"` for a nil set instead of the
  empty string, matching `types.Set.String`. An allocated empty set still
  renders as the empty string, so the two are now distinguishable in a log
  line. `email.AddressSet.AddressList` is unchanged and still returns the
  joined list, which is empty for both.
- **Breaking:** `Delete` now reports whether the set changed on all four set
  types, so its signature goes from `Delete(E)` to `Delete(E) bool`. Existing
  statement calls keep compiling unchanged, but three things do not:
  an interface declaring `Delete(E)` is no longer satisfied by these types,
  a method value typed `func(E)` no longer type checks, and neither does a
  method expression such as `uu.IDSet.Delete` typed `func(uu.IDSet, uu.ID)`.
  All three are compile errors, not silent behaviour changes.
- Nil sets are documented as valid empty sets for every read and for removals.
  `Insert`, `InsertAll`, `UnionWith` and `SymmetricDifferenceWith` panic on a
  nil set, exactly like an assignment to a nil Go map.
- `email.AddressSet` keeps its pointer-receiver `Add`, `AddSet`,
  `AddNormalized` and `AddAddressPart` methods, which allocate the underlying
  map and therefore work on a nil variable or struct field. They are the
  nil-safe counterparts of the value-receiver `Insert` and `UnionWith` that the
  abstract interface requires.

### Deprecated

Still present and working, scheduled for removal before `v1.0.0`:

| Deprecated              | Replacement                               |
|-------------------------|-------------------------------------------|
| `Add(v)`¹               | `Insert(v) bool`                          |
| `AddSlice(s)`           | `InsertAll(slices.Values(s))`             |
| `AddSet(other)`¹        | `UnionWith(other)`                        |
| `DeleteSlice(s)`        | `DeleteAll(slices.Values(s))`             |
| `DeleteSet(other)`      | `DifferenceWith(other)`                   |
| `IDSet.AddIDs(ids)`     | `InsertAll(slices.Values(ids.AsSlice()))` |
| `IDSet.AsSortedSlice()` | `Sorted()`                                |

¹ Neither is deprecated on `email.AddressSet`. Its pointer-receiver `Add` and
`AddSet` allocate the underlying map, so they have different nil semantics than
the value-receiver `Insert` and `UnionWith` and are kept as their nil-safe
counterparts, see above.

The `set` package is superseded by `mapset` for new code but is unchanged and
not deprecated.

### Removed

- **Breaking:** `uu.IDSet.Diff` and `strutil.StringSet.Diff`, which returned the
  symmetric difference. Use `SymmetricDifference` — or `Difference` if the
  asymmetric difference was what was meant.

### Performance

- `strutil.TrimSpace` and `strutil.TrimSpaceBytes` gained an ASCII fast path
  that avoids decoding runes and calling through a function value, falling
  back to the rune based predicate at the first non-ASCII byte.
