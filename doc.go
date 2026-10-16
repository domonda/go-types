// Package types is the root of the github.com/domonda/go-types module:
// a collection of Go packages providing enhanced type definitions,
// validation, and utility functions for common data and domain types.
//
// The root package itself provides foundational utilities used by the
// rest of the module:
//
//   - [Set]: generic set type backed by map[T]struct{} with set algebra
//     and JSON marshaling. See also [SetToSortedSlice], [SetToRandomizedSlice],
//     [ReduceSet], [ReduceSlice], [SliceContainsAll], [SliceContainsAny].
//   - [Validator] / [Validatable]: interfaces for boolean and error-based
//     validation, with [Validators] / [Validatables] composition helpers,
//     [CombinedValidator] / [CombinedValidatable], and [Validate],
//     [TryValidate], and [DeepValidate] entry points for recursive
//     reflection-based validation.
//   - [Normalizable] / [NormalizableValidator]: interfaces for one
//     canonical form via Normalized. SQL, JSON and XML reading and
//     writing do not call it, except date.Date on SQL.
//   - [FromPtr], [FromPtrOr]: generic pointer helpers.
//   - [Yield], [Yield2], [YieldErr], [Seq2NilError]: adapters for the
//     iter package introduced in Go 1.23.
//   - [LenString]: string with min/max length constraints.
//   - [Finder]: interface for finding sub-strings in text, implemented by
//     bank.IBANFinder, bank.BICFinder, *date.Finder, vat.IDFinder,
//     and money.AmountFinder.
//   - [KeyMutex]: generic per-key mutex pool whose zero value is usable.
//   - [CanMarshalJSON]: static check whether encoding/json can encode
//     a type. It does not inspect field, element, or map value types,
//     so it can report true for a type that fails to marshal.
//   - [CompareReflectValue]: three-way comparison of two [reflect.Value]
//     arguments, used to sort map keys in [DeepValidate].
//
// # Subpackages
//
// Domain value types with validation, normalization, and SQL/JSON
// integration:
//
//   - [github.com/domonda/go-types/account]: account numbers
//   - [github.com/domonda/go-types/bank]: IBAN, BIC, CAMT.053
//   - [github.com/domonda/go-types/country]: ISO 3166-1 country codes
//   - [github.com/domonda/go-types/date]: ISO 8601 dates, year-month, year-quarter
//   - [github.com/domonda/go-types/email]: addresses, lists, MIME parsing
//   - [github.com/domonda/go-types/language]: ISO 639-1 language codes
//   - [github.com/domonda/go-types/money]: amounts, currencies, rates
//   - [github.com/domonda/go-types/uu]: UUID type with nullable and slice variants
//   - [github.com/domonda/go-types/vat]: European VAT identification numbers
//
// Wrappers for SQL/JSON null semantics:
//
//   - [github.com/domonda/go-types/nullable]: empty/zero round-trips as NULL
//   - [github.com/domonda/go-types/notnull]: nil/empty round-trips as '{}' / []
//
// String and numeric utilities:
//
//   - [github.com/domonda/go-types/charset]: encoding, BOM, UTF-16/32
//   - [github.com/domonda/go-types/float]: locale-aware parsing and formatting
//   - [github.com/domonda/go-types/strfmt]: reflection-driven scan/format
//   - [github.com/domonda/go-types/strutil]: string manipulation and sets
//   - [github.com/domonda/go-types/strutil/stringfrom]: pointers and values
//     to strings with a default for nil
//
// Generic collection helpers:
//
//   - [github.com/domonda/go-types/mapset]: set functions over
//     map[K]struct{} and map[K]bool, mirroring the proposed container/mapset
//   - [github.com/domonda/go-types/queue]: concurrent FIFO queue
//   - [github.com/domonda/go-types/set]: set helpers over map[T]struct{}
//
// Most types in this module follow consistent conventions: a non-nullable
// type (e.g. [github.com/domonda/go-types/money.Currency]) where the empty
// value is invalid, paired with a NullableX variant where the empty value
// is a valid NULL representation that round-trips through SQL and JSON.
// Some NullableX types, like [github.com/domonda/go-types/money.NullableAmount],
// are aliases of [github.com/domonda/go-types/nullable.Type] instead,
// because the zero value of the wrapped type is valid.
// SQL is supported via [database/sql.Scanner] and
// [database/sql/driver.Valuer]. JSON works through encoding/json, natively
// for string based types or via [encoding/json.Marshaler] and
// [encoding.TextMarshaler] methods. Many types also provide a
// JSONSchema method compatible with github.com/invopop/jsonschema.
package types
