// Package date provides calendar dates as ISO 8601 strings (YYYY-MM-DD)
// with lenient parsing of the many ways people write dates.
//
// [Normalize] and [Date.Normalized] turn input like "25.12.2023",
// "Dec 25, 2023" or "1. Dezember 2016" into "2023-12-25".
// Numeric dates are read day first. An English language hint
// (language.EN, "en-US") switches to month first, but only for dates
// like "05/06/2024" where both orders are valid and no ordinal suffix
// like "1st" marks the day.
// Two-digit years below 45 are read as 20xx, all others as 19xx.
//
// The package includes:
//   - [Date] and [NullableDate] with SQL Scan and Value methods;
//     only [NullableDate] encodes null as JSON null
//   - [YearMonth], [YearQuarter], [NullableYearMonth] and
//     [NullableYearQuarter] for reporting periods
//   - [PeriodRange] for years, half years, quarters, months and ISO weeks
//   - date arithmetic, comparison and conversion to [time.Time]
//   - [Finder] to find dates in text
//   - [Parser], [Formatter] and [Format] for the strfmt package
//
// Methods normalize their input, so a valid but non-normalized Date
// like "25.12.2023" works everywhere except in == comparisons.
// No type in this package implements json.Unmarshaler:
// decoded JSON strings are stored as they are.
//
// See the README.md of this package for the full API,
// the serialization rules and the gotchas.
package date
