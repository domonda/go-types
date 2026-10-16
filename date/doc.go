// Package date provides calendar dates as ISO 8601 strings (YYYY-MM-DD)
// with lenient parsing of the many ways people write dates.
//
// [Normalize] and [Date.Normalized] turn input like "25.12.2023",
// "Dec 25, 2023" or "1. Dezember 2016" into "2023-12-25".
// A numeric date like "05/06/2024", where both orders are valid and no
// ordinal suffix like "1st" marks the day, is read by its separators:
// day first with dots like in Germany ("05.06.2024") and with any other
// separator but slashes, also with an English language hint. Slashes
// are written in both orders, so for them a language hint decides:
// English (language.EN, "en-US") reads month first, every other
// language day first. Without a hint slashes are read month first like
// in the US. SQL Scan and Value read them that way, because they fully
// normalize and have no hint. JSON and XML only trim surrounding
// whitespace, so a slash date stays as written.
// Two-digit years below 45 are read as 20xx, all others as 19xx.
//
// The package includes:
//   - [Date] and [NullableDate] with SQL Scan and Value methods;
//     only [NullableDate] encodes null as JSON null
//   - [YearMonth], [YearQuarter], [NullableYearMonth] and
//     [NullableYearQuarter] for reporting periods. They trim on every
//     path, and a year-quarter is also uppercased. A zero period is kept.
//   - [PeriodRange] for years, half years, quarters, months and ISO weeks
//   - date arithmetic, comparison and conversion to [time.Time]
//   - [Finder] to find dates in text
//   - [Parser], [Formatter] and [Format] for the strfmt package
//
// Methods other than reading and writing normalize their input, so a
// valid but non-normalized Date like "25.12.2023" works everywhere
// except in == comparisons. SQL Scan stores a valid date normalized and
// a zero date like "0001-01-01" as empty, and Value writes the normalized
// date, SQL NULL for an empty or zero date, and returns an error for an
// invalid date. JSON and XML only trim surrounding whitespace, so
// "25.12.2023" and a zero date stay as written. fmt prints a date as stored.
//
// See the README.md of this package for the full API,
// the serialization rules and the gotchas.
package date
