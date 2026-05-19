// Package types is the root of the github.com/domonda/go-types module:
// a collection of Go packages providing enhanced type definitions,
// validation, and utility functions for common data and domain types.
//
// The root package itself provides foundational utilities used by the
// rest of the module:
//
//   - [Set] — generic set type backed by map[T]struct{} with set algebra
//     and JSON marshaling. See also [SetToSortedSlice], [SetToRandomizedSlice],
//     [ReduceSet], [ReduceSlice], [SliceContainsAll], [SliceContainsAny].
//   - [Validator] / [ValidatErr] — interfaces for boolean and error-based
//     validation, with [Validators] / [ValidatErrs] composition helpers,
//     [CombinedValidator] / [CombinedValidatErr], and [Validate],
//     [TryValidate], and [DeepValidate] entry points for recursive
//     reflection-based validation.
//   - [Normalizable] / [NormalizableValidator] — interfaces implemented
//     by types that round-trip through a canonical normalized form.
//   - [Ptr], [FromPtr], [FromPtrOr] — generic pointer helpers.
//   - [Yield], [Yield2], [YieldErr], [Seq2NilError] — adapters for the
//     iter package introduced in Go 1.23.
//   - [LenString] — string with min/max length constraints.
//   - [Finder] — interface for sub-string finders implemented by the
//     date, bank, vat, and email packages.
//   - [KeyMutex] — generic per-key mutex pool.
//   - [CanMarshalJSON] — reflection-based check whether a type produces
//     valid JSON output.
//
// # Subpackages
//
// Domain value types with validation, normalization, and SQL/JSON
// integration:
//
//   - [github.com/domonda/go-types/account] — account numbers
//   - [github.com/domonda/go-types/bank] — IBAN, BIC, CAMT.053
//   - [github.com/domonda/go-types/country] — ISO 3166-1 country codes
//   - [github.com/domonda/go-types/date] — ISO 8601 dates, year-month, year-quarter
//   - [github.com/domonda/go-types/email] — addresses, lists, MIME parsing
//   - [github.com/domonda/go-types/language] — ISO 639-1 language codes
//   - [github.com/domonda/go-types/money] — amounts, currencies, rates
//   - [github.com/domonda/go-types/uu] — UUID type with nullable and slice variants
//   - [github.com/domonda/go-types/vat] — European VAT identification numbers
//
// Wrappers for SQL/JSON null semantics:
//
//   - [github.com/domonda/go-types/nullable] — empty/zero round-trips as NULL
//   - [github.com/domonda/go-types/notnull] — nil/empty round-trips as '{}' / []
//
// String and numeric utilities:
//
//   - [github.com/domonda/go-types/charset] — encoding, BOM, UTF-16/32
//   - [github.com/domonda/go-types/float] — locale-aware parsing and formatting
//   - [github.com/domonda/go-types/strfmt] — reflection-driven scan/format
//   - [github.com/domonda/go-types/strutil] — string manipulation and sets
//
// Generic collection helpers:
//
//   - [github.com/domonda/go-types/deref] — safe pointer dereferencing
//   - [github.com/domonda/go-types/queue] — concurrent FIFO queue
//   - [github.com/domonda/go-types/set] — set helpers over map[T]struct{}
//
// Most types in this module follow consistent conventions: a non-nullable
// type (e.g. [github.com/domonda/go-types/money.Currency]) where the empty
// value is invalid, paired with a NullableX variant where the empty value
// is a valid NULL representation that round-trips through SQL and JSON.
// SQL is supported via [database/sql.Scanner] and
// [database/sql/driver.Valuer]; JSON via [encoding/json.Marshaler] and
// [encoding/json.Unmarshaler]; most types additionally implement
// [encoding.TextMarshaler] / [encoding.TextUnmarshaler] and provide a
// JSONSchema method compatible with github.com/invopop/jsonschema.
package types
