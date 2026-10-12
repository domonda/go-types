// Package strfmt converts between strings and typed Go values via reflection,
// for example to read the cells of a CSV file or a spreadsheet into struct
// fields and to write typed values into a localized export.
//
// The package provides:
//   - [Scan]: parse a string into a reflect.Value, controlled by a [ScanConfig]
//     with true, false, and nil strings, time layouts, strict empty string
//     parsing, per-type [Scanner] functions, and a validation function
//   - [Format] and [FormatValue]: format a value as string, controlled by a
//     [FormatConfig] with float, money, date, and time formats, bool labels,
//     a nil placeholder, and per-type [Formatter] functions
//   - [NewFormatConfig], [NewEnglishFormatConfig], and [NewGermanFormatConfig]:
//     locale presets, and [MoneyFormat] for the money types
//   - [Scannable]: the ScanString method that most types of this module
//     implement to parse themselves
//   - [Parser]: the older interface for normalizing a string
//
// A nil config means the defaults: [Scan] uses [DefaultScanConfig] and
// [Format] uses [NewFormatConfig]. The default [ScanConfig] reads back what
// the presets write, like "ja" and "nein" and the preset time layouts.
//
// Scan never modifies the destination when it returns an error. By default
// an empty cell scans as the zero value of a number, use
// [ScanConfig.StrictEmptyStringParsing] to reject it. Floats are parsed with
// float.Parse, so a single separator is a decimal separator and "1,000"
// scans as 1.
//
// See the README.md file of the package for the resolution order of Scan
// and Format, the preset values, and the known pitfalls.
package strfmt
