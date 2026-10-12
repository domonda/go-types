// Package nullable provides nullable wrapper types for database/sql and
// encoding/json, and the interfaces that describe nullable values across
// this module.
//
// Every type has one null state that is written as SQL NULL and JSON null
// and read back from them, without pointer fields:
//   - [Type] wraps any type with a valid flag, so its zero value
//     Type[T]{} is null while TypeFrom(0) is a valid zero.
//   - [NonEmptyString] is null when empty, [TrimmedString] when empty
//     after trimming.
//   - [Time] is null for the zero time.Time.
//   - [JSON] holds raw JSON text and is null when nil.
//   - [StringArray], [IntArray], [FloatArray], [BoolArray] and the
//     element-nullable [NullIntArray], [NullFloatArray] and [NullBoolArray]
//     are PostgreSQL arrays where only a nil slice is NULL, so NULL and the
//     empty array '{}' stay distinct. Package
//     github.com/domonda/go-types/notnull writes a nil slice as '{}' instead.
//
// The interfaces [Nullable] and [NullSetable] are implemented by these types
// and by the NullableX types of the domain packages, [NullableValidator] by
// the domain types whose non-null values can be invalid, and [Zeroable]
// matches types with an IsZero method like time.Time. [ReflectIsNull] and
// [IsNull] report null values of any type, including nil pointers.
//
// See the [README] for tables of all functions and the exact
// serialization rules.
//
// [README]: https://github.com/domonda/go-types/blob/master/nullable/README.md
package nullable
