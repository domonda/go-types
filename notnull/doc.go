// Package notnull provides never-null counterparts of some types of package
// github.com/domonda/go-types/nullable, for NOT NULL columns and JSON fields
// that must not be null.
//
// The zero value of every type is written as a value:
//   - [StringArray], [IntArray], [FloatArray] and [NullBoolArray] write a nil
//     slice as the empty PostgreSQL array '{}' and the JSON array [].
//     Scan and UnmarshalJSON read SQL NULL and JSON null as a non-nil empty
//     slice, so the arrays can be used without nil checks.
//   - [JSON] writes nil as the empty JSON object {}.
//   - [TrimmedString] trims every value and has no null state,
//     the empty string is a value.
//
// Use the nullable types instead where SQL NULL and JSON null must stay
// distinct from an empty value. Note that switching a field between the two
// packages changes what a nil slice is written as.
//
// [SplitArray], [SplitArrayValues] and [SQLArrayLiteral] parse and build
// PostgreSQL array literals and never return nil or NULL for an empty array.
//
// See the [README] for tables of all functions and the exact
// serialization rules.
//
// [README]: https://github.com/domonda/go-types/blob/master/notnull/README.md
package notnull
