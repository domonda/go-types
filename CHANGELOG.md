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

- **Breaking:** `strfmt.Scan` resolves a nil source string (as reported by
  `ScanConfig.IsNil`, now compared with surrounding whitespace trimmed) before
  dispatching to any scanning method, so that every destination type gets the
  same handling. A `Scanner` registered in `ScanConfig.TypeScanners` is no
  longer consulted first for a nil source string.
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

### Performance

- `strutil.TrimSpace` and `strutil.TrimSpaceBytes` gained an ASCII fast path
  that avoids decoding runes and calling through a function value, falling
  back to the rune based predicate at the first non-ASCII byte.
