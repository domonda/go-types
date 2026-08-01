# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
No version has been tagged yet, so everything below is unreleased. See
[TODOS.md](./TODOS.md) for what still blocks a `v1.0.0` cut.

## [Unreleased]

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
- `internal/collections` holding the proposal's abstract interfaces, and
  `internal/collections/settest` holding the abstract set specification as a
  reusable test. Every set type is checked against that specification and by a
  compile-time interface assertion, so a behaviour change in `mapset` is caught
  at each type that delegates to it, not only in `mapset` itself.

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
  which is now `types.Set.SymmetricDifference`. This is the one change of this
  release that does not produce a compile error — review any call site that
  relies on the old meaning.
- **Breaking:** `types.Set.ContainsAll` takes an `iter.Seq[T]` instead of
  variadic values. Use `set.ContainsAll(slices.Values(vals))`.
- `Delete` now reports whether the set changed on all four set types. Existing
  statement calls keep compiling unchanged; only method values typed `func(E)`
  break.
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

| Deprecated                             | Replacement                             |
|----------------------------------------|-----------------------------------------|
| `Add(v)`¹                              | `Insert(v) bool`                        |
| `AddSlice(s)`                          | `InsertAll(slices.Values(s))`           |
| `AddSet(other)`                        | `UnionWith(other)`                      |
| `DeleteSlice(s)`                       | `DeleteAll(slices.Values(s))`           |
| `DeleteSet(other)`                     | `DifferenceWith(other)`                 |

¹ Not deprecated on `email.AddressSet`, where `Add` has different nil semantics
than `Insert`, see above.

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
