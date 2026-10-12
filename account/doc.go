// Package account provides account number types with validation,
// numeric conversion, and SQL, JSON, XML, and text integration.
//
// The package provides:
//   - [Number]: a required account number, the empty string is invalid
//   - [NullableNumber]: an optional account number,
//     the empty string [NumberNull] is the null value
//   - [NumberRegexp]: the pattern that Valid and Validate check,
//     [DefaultNumberRegex] unless reassigned at startup
//   - [ErrInvalidNumber] and [ErrAlphanumericNumber]: error sentinels
//     wrapped by the validation and conversion errors
//
// By default an account number consists of ASCII letters, digits,
// and the characters _ - / : . ; , and starts with a letter or digit.
// [Number.Uint] and [Number.Int] only convert numbers of ASCII digits.
//
// Both types trim whitespace when they are written and decoded.
// An empty Number is written as SQL NULL and JSON null like a null
// NullableNumber, but only NullableNumber can read those values back.
//
// See the README.md file of the package for the complete API reference,
// the serialization rules, and the known pitfalls.
package account
