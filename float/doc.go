// Package float parses and formats floating point numbers with
// locale-dependent separators and rounds them to decimal places.
//
// The package provides:
//   - [Parse] and [ParseDetails]: parse numbers like "1.234,56",
//     "1,234.56", "1'234.56", or "1 234,56" and report the detected
//     thousands and decimal separators
//   - [Format]: format a float with a thousands and a decimal separator,
//     rounded to a precision
//   - [FormatDef]: a format definition for [Format] with a JSON form,
//     and a [FormatDef.Parse] method implementing strfmt.Parser
//   - [Tolerant]: a float64 that decodes JSON numbers, strings, and null
//   - [RoundToDecimals], [Valid], [ValidAndHasSign], and [DerefOr]:
//     generic helpers for float32 and float64
//
// A single '.' or ',' is always the decimal separator, so "1,234"
// parses as 1.234 and not as 1234. The apostrophe and the space are only
// thousands separators, so "1'234" parses as 1234 and "1'5" is an error.
//
// [RoundToDecimals] and [Format] round half away from zero, and they round
// the shortest decimal representation of a float instead of its binary
// value, so 1.005 rounds to 1.01 and 2.5 to 3.
//
// See the README.md file of the package for the complete API reference,
// the JSON forms, and the known pitfalls.
package float
