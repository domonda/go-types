# date

Calendar dates as ISO 8601 strings (`YYYY-MM-DD`) that parse the many ways people write dates: `"25.12.2023"`, `"Dec 25, 2023"`, `"1. Dezember 2016"`, `"2023-12-25T10:00:00Z"`. Use `Date` for required dates, `NullableDate` for optional ones, and `YearMonth` / `YearQuarter` for reporting periods. The types are strings, so they work as map keys, compare cheaply and are readable in logs and SQL.

    import "github.com/domonda/go-types/date"

## Types

| Type                  | Purpose                                            |
|-----------------------|----------------------------------------------------|
| `Date`                | A calendar date, canonical form `YYYY-MM-DD`. Empty and the zero dates are not valid. |
| `NullableDate`        | Like `Date`, but `Null` (`""`) and the zero dates are valid and SQL NULL. |
| `YearMonth`           | A month, `YYYY-MM`.                                |
| `NullableYearMonth`   | `YearMonth` where `""` and whitespace are null.    |
| `YearQuarter`         | A quarter, `YYYY-Qn` with an upper-case `Q`.       |
| `NullableYearQuarter` | `YearQuarter` where `""` and whitespace are null.  |
| `Format`              | Layout plus nil and zero placeholders, for reflection based formatting. |
| `Formatter`           | A `time` layout string that formats and parses dates. |
| `Parser`              | Parses any accepted input to `YYYY-MM-DD`.         |
| `Finder`              | Finds dates in text.                               |

## Date

```go
const (
    Layout    = time.DateOnly // "2006-01-02"
    Regex     = `\d{4}-\d{2}-\d{2}`
    Length    = 10 // len("2006-01-02")
    MinLength = 6  // minimum length of a valid date
    Invalid   Date = ""
)
```

A `Date` holds whatever string you assign. Methods normalize on demand, so `Date("25.12.2023")` is valid and behaves like `"2023-12-25"` in every method except `==`. Reading with `Scan`, JSON, XML and `ScanString` stores the canonical form; normalize a date built from a string yourself, with `Normalize`.

A `Date` is zero, see `IsZero`, if it is empty, whitespace only, or one of the zero dates `"0000-00-00"` and `"0001-01-01"` in any spelling, like `"01.01.0001"` or Go's zero time `"0001-01-01T00:00:00Z"`. A zero date has no practical use as a date: it is not valid, reading stores it as `""`, and `Value` writes it as SQL NULL.

### Parsing

`Normalize(str, lang...)`, `Date.Normalized()` and `Date.NormalizedLang(lang)` accept, case-insensitive and with lenient whitespace:

- ISO dates with `-`, `/`, `.` or space as separator: `"2006-01-02"`, `"2006/01/02"`.
- An ISO date followed by a `T` time (`"2006-01-02T15:04:05Z07:00"`) or a trailing `00:00:00` (`"30.11.2023 00:00:00"`). The time is dropped.
- Numeric day-month-year and month-day-year with 2 or 4 digit years: `"25.12.1975"`, `"25.12.75"`, `"12/25/1975"`.
- English and German month names and abbreviations: `"jan. 24 2012"`, `"1. Dezember 2016"`, `"Oct 7, 1970"`, `"September 17th, 2012"`, `"Oct 7, '70"`, `"2013-Feb-03"`, `"07-Feb-04"`. Common OCR misreadings like `"1an"` or `"apri1"` are accepted too.
- French, Italian and Spanish month names and their usual abbreviations: `"2 janv. 2019"`, `"23/gen/2019"`, `"12 enero 2020"`, `"2 févr. 2019"`, `"5 luglio 2021"`.
- English ordinals and a leading weekday: `"1st of dec. 2020"`, `"Fri, 03 Jul 2015"`.
- CJK separators: `"2014年04月08日"`.

Not supported: year-only or year-month strings (`"2014"`, `"2014-04"`), compact digits (`"20140601"`), dates whose three parts have fewer than 5 characters together (`"8/8/71"`), and times with a zone offset inside the string (`"Mon Jan 02 15:04:05 -0700 2006"`). Use `time.Parse` or `ParseTime` for those.

Rules for numeric dates:

- **Day first by default.** `"05/06/2024"` is 2024-06-05.
- **Language hint.** The first `lang` argument is a hint, normalized with `language.Code.Normalized`, so `language.EN`, `"EN"`, `" en"`, `"en-US"` and `"eng"` all mean English. A hint that can't be normalized, like `"english"`, is ignored. Only English changes the result: it reads month first.
- **The hint only decides ambiguous dates.** It applies only when both numbers are 1 to 12. `"15.06.2024"` and `"06/15/2024"` both give 2024-06-15 with any hint. The separator never matters.
- **Ordinal suffix.** A suffix like `"1st"` marks the day for ambiguous dates and wins over the hint: `"02 1st 2003"` and `"1st 02 2003"` are 2003-02-01, also with `language.EN`. A number with a suffix is never the month, so `"13 1st 2024"` is an error.
- **Two-digit years.** 00 to 44 become 2000 to 2044, 45 to 99 become 1945 to 1999. A single digit is never the year, so `"12.12.5"` is an error, and the year 0 is an error in every order, like `"0000 25 may"`.
- **Date-times.** A `T` after the date is only cut off with the time if a digit follows, so `"2023-12-25T10:00:00Z"` is 2023-12-25 and `"2023-12-25TOTAL"` is an error.

### Functions

| Function                                     | Description                                        |
|----------------------------------------------|----------------------------------------------------|
| `Normalize(str, lang...)`                    | Parses `str` and returns the `YYYY-MM-DD` form, or an error. |
| `Must(str)`                                  | `Normalize` without hint; panics on error.         |
| `StringIsDate(str, lang...)`                 | Reports whether `Normalize` accepts `str`.         |
| `Of(year, month, day)`                       | Date from parts. Out of range values roll over: October 32 is November 1. |
| `OfTime(t)`                                  | Date part of `t` in `t`'s location; `""` for the zero `time.Time`. |
| `OfTimePtr(t)`                               | Like `OfTime`, returns a `NullableDate`: `Null` for nil or the zero time. |
| `OfToday()`, `OfYesterday()`, `OfTomorrow()` | In the local time zone.                            |
| `OfNowInUTC()`                               | Today in UTC.                                      |
| `OfTodayIn(loc)`                             | Today in `loc`.                                    |
| `Parse(layout, value)`                       | `time.Parse`, then the date part. Returns `""` without error if the time is zero. |

### Validation and normalization

| Method                             | Description                                        |
|------------------------------------|----------------------------------------------------|
| `Valid()`                          | True if `Normalized()` succeeds. No hint, so any parseable form is valid. |
| `Validate()`                       | The error of `Normalized()`, or nil.               |
| `Normalized()`                     | The `YYYY-MM-DD` form, or the date unchanged and an error, also for a zero date. Implements `types.Normalizable[Date]`, so it has no order hint. |
| `NormalizedLang(lang)`             | Like `Normalized`, with `lang` as order hint.      |
| `ValidAndNormalized()`             | True only if the date is already a valid `YYYY-MM-DD`, not a zero date. |
| `NormalizedOrUnchanged(lang...)`   | Normalized form, or the date unchanged if it can't be parsed. |
| `NormalizedOrNull(lang...)`        | Normalized form as `NullableDate`, or `Null`.      |
| `IsZero()`                         | True for an empty or whitespace only date and the zero dates `"0000-00-00"` and `"0001-01-01"` in any spelling. |
| `Nullable()`                       | Converts to `NullableDate`; zero dates become `Null`. |
| `ScanString(source, validate)`     | Without `validate` like `Scan`. With `validate` assigns the normalized `source` or returns the error of `Normalized`, also for a blank or zero `source`. |
| `ScanStringWithLang(source, lang)` | Assigns the normalized `source` using the hint `lang`, `""` for a zero date. Returns `wasNormalized` (source was already `YYYY-MM-DD`), `monthMustBeFirst` (only month first is valid, like `"12/25/2023"`) and an error. Unchanged on error. |

`monthMustBeFirst` lets you detect the date order of a whole document from one unambiguous date, then pass `language.EN` for the rest.

### Conversion to time.Time and strings

| Method                            | Description                                        |
|-----------------------------------|----------------------------------------------------|
| `MidnightUTC()`                   | 00:00 UTC of the date. Zero `time.Time` if the date is zero or invalid. |
| `Midnight()`                      | 00:00 in `time.Local`, same rules.                 |
| `MidnightInLocation(loc)`         | 00:00 in `loc`, same rules.                        |
| `Time(hour, minute, second, loc)` | The date at a time of day. Zero `time.Time` for a zero date. An invalid date gives a meaningless time before year 1. |
| `TimeLocal(hour, minute, second)` | `Time` in `time.Local`.                            |
| `TimeUTC(hour, minute, second)`   | `Time` in UTC.                                     |
| `Format(layout)`                  | Formats with a `time` layout. `""` if the date is zero or layout is empty. `Format(Layout)` returns the normalized date, or an invalid date unchanged. |

### Parts

| Method                       | Description                                        |
|------------------------------|----------------------------------------------------|
| `Year()`, `Month()`, `Day()` | The parts, 0 if invalid.                           |
| `YearMonthDay()`             | All three parts, zeros if invalid.                 |
| `Weekday()`                  | Day of the week, `time.Sunday` (0) if zero or invalid. |
| `ISOWeek()`                  | ISO 8601 year and week number, zeros if zero or invalid. |

### Comparison

| Method                                             | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `Compare(other)`                                   | -1, 0 or +1 comparing the normalized strings. Invalid dates compare like `""`. |
| `NormalizedEqual(other)`                           | True if both normalize to the same string. False if either is invalid, even for identical strings. |
| `After(other)`, `Before(other)`                    | Strict order of the UTC midnights. False if either date is invalid, also `""`, like `NormalizedEqual`. |
| `EqualOrAfter(other)`, `EqualOrBefore(other)`      | Inclusive variants.                                |
| `WithinIncl(from, until)`                          | `from <= date <= until`. False if any date is invalid. |
| `BetweenExcl(after, before)`                       | `after < date < before`. False if any date is invalid. |
| `AfterTime(t)`, `BeforeTime(t)`                    | Compares midnight of the date in `t`'s location with `t`. False for an invalid date. |
| `IsToday()`, `IsTodayInUTC()`                      | `NormalizedEqual(OfToday())` / `NormalizedEqual(OfNowInUTC())`. |
| `AfterToday()`, `AfterTodayInUTC()`, `BeforeToday()`, `BeforeTodayInUTC()` | `After` / `Before` today in the local time zone or UTC. |

### Arithmetic and period boundaries

| Method                                             | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `AddDate(years, months, days)`                     | Like `time.Time.AddDate`, overflow rolls over: `"2024-01-31"` + 1 month is `"2024-03-02"`. |
| `AddYears(n)`, `AddMonths(n)`, `AddDays(n)`        | Shortcuts for `AddDate`.                           |
| `Add(d)`                                           | Adds a `time.Duration` to midnight UTC and returns the date part. |
| `Sub(other)`                                       | `date - other` as `time.Duration`.                 |
| `BeginningOfWeek(startDay)`, `EndOfWeek(startDay)` | First and last day of the week that starts on `startDay` (`time.Monday` for ISO weeks). |
| `BeginningOfMonth()`, `EndOfMonth()`               | First and last day of the month.                   |
| `BeginningOfQuarter()`, `EndOfQuarter()`           | First and last day of the quarter.                 |
| `BeginningOfYear()`, `EndOfYear()`                 | January 1 and December 31.                         |
| `LastMonday()`                                     | The Monday on or before the date.                  |
| `NextSunday()`                                     | The Sunday on or after the date.                   |

Implements: `database/sql.Scanner`, `database/sql/driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `xml.Unmarshaler`, `xml.UnmarshalerAttr`, `xml.MarshalerAttr`, `strfmt.Scannable`, `types.Validator`, `types.ValidatErr`, `types.NormalizableValidator[Date]` and `JSONSchema()` (`{"type":"string","format":"date"}`). `Date` has no `String` method, so `fmt` prints the stored date, and no `UnmarshalText`, so `encoding/json` reads the keys of a `map[Date]V` as given.

## NullableDate

```go
const Null NullableDate = ""
```

`NullableDate` has the same string form as `Date`. `Null`, a value of only the runes `strutil.TrimSpace` removes (like for `country.NullableCode`), the zero dates `"0000-00-00"` and `"0001-01-01"`, and a zero date in another format like `"01.01.0001"` are null, valid, and written as SQL NULL and JSON `null`.

| Function                          | Description                                        |
|-----------------------------------|----------------------------------------------------|
| `NormalizeNullable(str, lang...)` | Like `Normalize`; `Null` without error for `""` and zero dates. |
| `MustNullable(str)`               | `NormalizeNullable` without hint; panics on error. |
| `OfTimePtr(t)`                    | See `Date` functions.                              |

| Method                                             | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `IsNull()`                                         | True for `Null`, only `strutil.TrimSpace` runes and the zero dates, also in another format like `"01.01.0001"`. |
| `IsZero()`                                         | True for `Null` and the zero dates.                |
| `IsNotNull()`                                      | `!IsNull()`.                                       |
| `Get()`                                            | The `Date`; panics if null.                        |
| `GetOr(defaultDate)`, `DateOr(useIfNull)`          | The `Date`, or the argument if null.               |
| `Date()`                                           | Plain conversion to `Date`, no null check.         |
| `Set(d)`, `SetNull()`                              | Assign a `Date` or `Null`.                         |
| `Valid()`, `Validate()`                            | Null is valid, otherwise like `Date`.              |
| `ValidAndNotNull()`                                | Valid and not null, so false for the zero dates.   |
| `ValidAndNormalized()`                             | True for `Null` and for a valid `YYYY-MM-DD` that is not a zero date. |
| `Normalized()`, `NormalizedLang(lang)`             | `Null` for null input, otherwise like `Date`.      |
| `NormalizedOrNull(lang...)`, `NormalizedOrUnchanged(lang...)` | As for `Date`.                                     |
| `StringOr(nullString)`                             | As stored, or `nullString` if null.                |
| `ScanString(source, validate)`                     | `""`, `"NULL"`, `"null"`, `"nil"` and whitespace only set `Null`; otherwise like `Date`. |
| `ScanStringWithLang(source, lang)`                 | A null `source` sets `Null`; otherwise like `Date`. |
| `MidnightUTC()`, `Midnight()`, `MidnightInLocation(loc)` | Return `nullable.Time`, null for null or invalid dates. |

These methods mirror `Date` and add null rules:

| Methods                                            | Null rule                                          |
|----------------------------------------------------|----------------------------------------------------|
| `AddDate`, `AddYears`, `AddMonths`, `AddDays`, `Add`, `BeginningOf*`, `EndOf*`, `LastMonday`, `NextSunday` | Return `NullableDate`; `Null` stays `Null`.        |
| `Year`, `Month`, `Day`, `YearMonthDay`, `Weekday`, `ISOWeek` | Zeros for null.                                    |
| `Format(layout)`                                   | `""` for `Null`.                                   |
| `Compare`, `NormalizedEqual`, `After`, `Before`, `EqualOrAfter`, `EqualOrBefore` | Null is before every valid non-null date. An invalid date is neither before nor after any date, also not `Null`, but `Compare` sorts it like `Null`. |
| `AfterTime(t)`                                     | False for null and invalid dates. True for a valid non-null date if `t` is zero. |
| `BeforeTime(t)`                                    | For null: true unless `t` is zero. False for an invalid date. |
| `Sub(other)`                                       | 0 if either is null.                               |
| `IsToday`, `IsTodayInUTC`, `AfterToday`, `AfterTodayInUTC` | False for null.                                    |
| `BeforeToday`, `BeforeTodayInUTC`                  | True for null.                                     |

`NullableDate` has no `Time`, `TimeLocal`, `TimeUTC`, `WithinIncl`, `BetweenExcl` or `Nullable`.

Implements: `database/sql.Scanner`, `database/sql/driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, `xml.Unmarshaler`, `xml.UnmarshalerAttr`, `xml.MarshalerAttr`, `strfmt.Scannable`, `types.NormalizableValidator[NullableDate]`, `nullable.NullSetable[Date]` and `JSONSchema()` (string with format `date`, or null). No `String` method, `fmt` prints the stored date.

## Periods

| Function                                   | Description                                        |
|--------------------------------------------|----------------------------------------------------|
| `PeriodRange(period)`                      | First and last day of `"YYYY"`, `"YYYY-MM"`, `"YYYY-Qn"`, `"YYYY-Hn"` (half year) or `"YYYY-Wnn"` (ISO week). The letters may be lower case. |
| `YearRange(year)`                          | `"YYYY-01-01"` and `"YYYY-12-31"`.                 |
| `YearWeekMonday(year, week)`               | Monday of an ISO 8601 week.                        |
| `YearWeekRange(year, week)`                | Monday and Sunday of an ISO 8601 week.             |
| `FromUntilFromYearAndMonths(year, months)` | First and last day of a month range: `("2024", "3-9")` gives 2024-03-01 and 2024-09-30. Empty `months` means `"1-12"`; empty `year` returns empty dates and no error. |

An ISO week runs Monday to Sunday and can start or end in the neighbouring year: `PeriodRange("2021-W01")` is 2021-01-04 to 2021-01-10, `PeriodRange("2019-W01")` is 2018-12-31 to 2019-01-06. Week 53 is only accepted for years that have 53 ISO weeks: `PeriodRange("2020-W53")` is 2020-12-28 to 2021-01-03, `PeriodRange("2021-W53")` is an error.

## YearMonth and NullableYearMonth

```go
const YearMonthLayout = "2006-01"
const YearMonthNull NullableYearMonth = ""
```

| Function                     | Description                                    |
|------------------------------|------------------------------------------------|
| `YearMonthFrom(year, month)` | Formats as `YYYY-MM` without validation.       |
| `YearMonthOfTime(t)`         | Year and month of `t`; `""` for the zero time. |
| `YearMonthOfToday()`         | In the local time zone.                        |

| `YearMonth` method                   | Description                                        |
|--------------------------------------|----------------------------------------------------|
| `Normalized()`                       | Trims the surrounding whitespace; the value unchanged and an error if invalid. |
| `Validate()`, `Valid()`              | `Normalized` succeeds: `YYYY-MM`, year 1 to 3000, month 1 to 12. |
| `ValidAndNormalized()`               | Exactly `YYYY-MM` and valid.                       |
| `IsZero()`                           | True for `""`, `"0000-00"` and `"0001-01"`, also with whitespace. |
| `String()`                           | The raw string.                                    |
| `Year()`, `Month()`                  | The parts, 0 if invalid.                           |
| `Date(day)`                          | `"YYYY-MM-DD"` for `day`, without validation (`Date(30)` of February is invalid). |
| `DateRange()`                        | First and last day of the month; `""`, `""` if invalid. |
| `Format(layout)`                     | `""` if empty; `YearMonthLayout` returns the normalized value, an invalid one as is; other layouts format the 1st of the month. |
| `AddYears(n)`, `AddMonths(n)`        | Arithmetic; months roll over into years. Invalid values are returned unchanged. |
| `ContainsTime(t)`, `ContainsDate(d)` | True if `t` or `d` falls in the month (local time zone midnights). |
| `Compare(other)`                     | Compares the normalized values, invalid ones as stored. |
| `UnmarshalJSON(data)`                | Reads a JSON string like `Date`, see Reading and writing. |
| `Nullable()`                         | Plain conversion to `NullableYearMonth`.           |

`NullableYearMonth` has the same methods (null returns 0, `""`, `false` or null; `Normalized` returns `YearMonthNull` for null) plus `IsNull()` (`""` and whitespace only; `"0000-00"` is zero but not null), `IsNotNull()`, `ValidAndNotNull()` (false for null and zero values), `Get()` (panics if null), `GetOr(defaultYearMonth)`, `Set(ym)`, `SetNull()` and `StringOr(nullString)`. Its `Date` and `DateRange` return `NullableDate`. Implements `nullable.NullSetable[YearMonth]`, `types.NormalizableValidator` and `json.Unmarshaler` (JSON `null` sets null, a string is read like by `YearMonth`, a blank one as null).

## YearQuarter and NullableYearQuarter

```go
const YearQuarterNull NullableYearQuarter = ""
```

| Function                         | Description                              |
|----------------------------------|------------------------------------------|
| `YearQuarterFrom(year, quarter)` | Formats as `YYYY-Qn` without validation. |
| `YearQuarterOfTime(t)`           | Quarter of `t`; `""` for the zero time.  |
| `YearQuarterOfToday()`           | In the local time zone.                  |

| `YearQuarter` method                               | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `Normalized()`                                     | Trims the surrounding whitespace and upper-cases `q`; the value unchanged and an error if invalid. |
| `Validate()`, `Valid()`                            | `Normalized` succeeds: `YYYY-Qn`, year 1 to 3000, quarter 1 to 4. |
| `ValidAndNormalized()`                             | Exactly `YYYY-Qn` with upper-case `Q` and valid.   |
| `IsZero()`                                         | True for `""`, `"0000-Q0"` and `"0001-Q1"`, also with whitespace or `q`. |
| `String()`                                         | The raw string.                                    |
| `Year()`, `Quarter()`                              | The parts, 0 if invalid.                           |
| `FirstMonth()`                                     | January, April, July or October; 0 if invalid.     |
| `Date(monthInQuarter, day)`                        | Date in the quarter via `Of`, so values roll over; `""` if invalid. |
| `DateRange()`                                      | First and last day of the quarter; `""`, `""` if invalid. |
| `AddYears(n)`, `AddQuarters(n)`                    | Arithmetic; quarters roll over into years. Invalid values are returned unchanged. |
| `ContainsTime(t)`, `ContainsDate(d)`, `ContainsYearMonth(ym)` | True if the value falls in the quarter.            |
| `Compare(other)`                                   | Compares the normalized values, invalid ones as stored. |
| `Nullable()`                                       | Plain conversion to `NullableYearQuarter`.         |
| `UnmarshalJSON(data)`                              | Reads a JSON string like `Date`, see Reading and writing. |

`NullableYearQuarter` has the same methods plus the same null helpers as `NullableYearMonth`. Its `Date` and `DateRange` return `NullableDate`. Implements `nullable.NullSetable[YearQuarter]`, `types.NormalizableValidator` and `json.Unmarshaler` like `NullableYearMonth`.

## Format, Formatter and Parser

These plug dates into the [`strfmt`](../strfmt/README.md) package, which calls `Parse(str string, langHints ...language.Code) (normalized string, err error)`.

| Type / method                                | Description                                        |
|----------------------------------------------|----------------------------------------------------|
| `Parser{}.Parse(str, langHints...)`          | `Normalize`, returns `YYYY-MM-DD`.                 |
| `Formatter(layout).Format(d)`                | `d.Format(layout)`.                                |
| `Formatter(layout).Parse(str, langHints...)` | `Normalize`, then formats with the layout.         |
| `Format{Layout, NilString, ZeroString}`      | Struct with JSON tags `layout`, `nilString`, `zeroString`. |
| `(*Format).Format(d)`                        | `d.Format(f.Layout)`.                              |
| `(*Format).Parse(str, langHints...)`         | `Normalize`, then formats with `Layout`. Pointer receiver: only `*Format` implements `strfmt.Parser`. |
| `(*Format).AssignString(dest, source)`       | Parses `source` (with `Normalize` if `Layout` is empty, else `time.Parse(Layout)`) and assigns it to the addressable `dest` of type `Date`, `NullableDate`, `time.Time`, `*time.Time` or `nullable.Time`. Whitespace-only `source` assigns the empty value. A parsed zero time is an error. |
| `(*Format).FormatString(val)`                | `NilString` for an invalid `reflect.Value` or a nil pointer, `ZeroString` for zero values with `Format`/`IsZero` methods (`Date`, `NullableDate`, `time.Time`), else `Format(Layout)`, which returns an invalid date like `"bad"` unchanged. Other types are an error. |

`Parser` and `Formatter` implement `strfmt.Parser` as values, `Format` only as `*Format`. None of them implements `strfmt.Formatter`.

## Finder

| Identifier                       | Description                                        |
|----------------------------------|----------------------------------------------------|
| `NewFinder(lang...)`             | Returns a `*Finder` with the normalized first `lang` as hint, no hint if it can't be normalized. |
| `Finder.LangHint`                | The language hint field.                           |
| `(*Finder).FindAllIndex(str, n)` | `[start, end)` byte offsets of dates in `str`. At most `n` results; `n < 0` returns all, `n == 0` none. |

`*Finder` implements `types.Finder`. It tries spans of one to three whitespace-separated words and trims surrounding punctuation, so `"25.12.2023,"` is found as `"25.12.2023"`. It reports the shortest span that parses at each start word. A span is found exactly if `Normalize` accepts it, so an ISO 8601 date-time like `"2023-12-25T10:00:00Z"` is found as a whole. The exception are the Italian and Spanish month abbreviations that are common words in other languages, `set`, `ago`, `mag`, `lug`, `ott` and `dic`: they need `language.IT` or `language.ES` as hint, so that `"Game set 10 2020"` has no date.

## time.Time helpers

| Identifier                   | Description                                        |
|------------------------------|----------------------------------------------------|
| `ParseTime(str, layouts...)` | Tries `time.Parse` with each layout and returns the first success. Defaults to `ParseTimeDefaultLayouts`. |
| `ParseTimeDefaultLayouts`    | `time.RFC3339Nano`, `time.RFC3339`, `"2006-01-02 15:04:05"`. A package variable. |

## Serialization

`Date` and `NullableDate` follow the rules of [Reading and writing values](../README.md#reading-and-writing-values), with the zero dates as blank: reading normalizes, `Value` refuses an invalid date, and `fmt` prints the stored date.

| Value                                              | SQL `Value()`                              | JSON (`json.Marshal`)         | `MarshalText`, XML | `fmt`          |
|----------------------------------------------------|--------------------------------------------|-------------------------------|--------------------|----------------|
| `Date("")`, `Date("  ")`                           | `NULL`                                     | `""`                          | `""`               | as stored      |
| `Date("0001-01-01")`, `Date("0000-00-00")`, `Date("01.01.0001")` | `NULL`                                     | `""`                          | `""`               | as stored      |
| `Date("25.12.2023")` (valid, not normalized)       | `"2023-12-25"`                             | `"2023-12-25"`                | `"2023-12-25"`     | `25.12.2023`   |
| `Date("2023-12-25")`                               | `"2023-12-25"`                             | `"2023-12-25"`                | `"2023-12-25"`     | `2023-12-25`   |
| `Date("bad")` (invalid)                            | error                                      | `"bad"`                       | `"bad"`            | `bad`          |
| `NullableDate` null or zero date                   | `NULL`                                     | `null`                        | `""`               | as stored      |
| `NullableDate("25.12.2023")`                       | `"2023-12-25"`                             | `"2023-12-25"`                | `"2023-12-25"`     | `25.12.2023`   |
| `NullableDate("bad")`                              | error                                      | `"bad"`                       | `"bad"`            | `bad`          |
| `YearMonth`, `YearQuarter` and their nullable types | the raw string, `""` for null (not `NULL`) | the raw string, `""` for null | the raw string     | the raw string |

Reading values:

| Input                                              | `Date`                                             | `NullableDate`                       |
|----------------------------------------------------|----------------------------------------------------|--------------------------------------|
| `Scan(nil)`                                        | `""`                                               | `Null`                               |
| `Scan` of a string or `[]byte`, JSON string, XML element or attribute | normalized if valid, `""` if blank or a zero date, else as is, without error | same, `Null` if blank or a zero date |
| `Scan(time.Time)`                                  | its date, `""` if zero                             | its date, `Null` if zero             |
| `Scan` of other types                              | error                                              | error                                |
| JSON `null`                                        | field left unchanged                               | `Null`                               |
| other JSON values                                  | error                                              | error                                |

An invalid date is read without error, so one bad row doesn't fail a whole query, and kept for `Validate` to report. `YearMonth` and `YearQuarter` read a JSON string like `Date`: normalized if valid, `""` if blank, else as is, without an error, and leave the value unchanged for JSON `null`. `NullableYearMonth` and `NullableYearQuarter` read a JSON string the same way, a blank one as null, and set null for JSON `null`. `YearMonth`, `YearQuarter`, `NullableYearMonth` and `NullableYearQuarter` have no SQL methods: `database/sql` writes them as plain strings and fails to scan NULL into them (`converting NULL to string is unsupported`); scan into a `*YearMonth` to accept NULL.

## Gotchas

- **Valid, Validate, Normalized, ValidAndNormalized.** `Valid` and `Validate` accept any parseable form: `Date("25.12.2023").Valid()` is true. `Normalized` returns the canonical form. `ValidAndNormalized` is true only for an exact `YYYY-MM-DD`. `Date("")` is invalid, `NullableDate("")` is valid.
- **`==` compares raw strings.** `Date("25.12.2023") != Date("2023-12-25")`. Use `NormalizedEqual` or `Compare`, or normalize on input. Dates read with `Scan`, JSON or XML are already normalized.
- **Zero dates are no dates.** `"0001-01-01"` and `"0000-00-00"`, also in another spelling like `"01.01.0001"`, are `IsZero()` and not valid. Reading stores them as `""`, `Value` writes SQL NULL, `Date` encodes them in JSON as `""` and `NullableDate` as `null`, and the comparisons treat them like an invalid date.
- **Day first.** `"05/06/2024"` is June 5 unless the hint is English. The hint never changes a date that has only one valid reading. Only the first `lang` argument is used.
- **Two-digit years** pivot at 45: `"25.12.44"` is 2044-12-25, `"25.12.45"` is 1945-12-25.
- **Invalid and zero dates become the zero time in arithmetic.** `Date("").AddDays(1)` and `Date("bad").AddDays(1)` return `"0001-01-02"`. Check `Valid` (or `IsZero`) before doing arithmetic. `Format` returns an invalid date unchanged with every layout: `Date("bad").Format("02.01.2006")` is `"bad"`. Comparisons don't use the zero time: `After`, `Before`, `WithinIncl`, `BetweenExcl`, `AfterTime` and `BeforeTime` are false for an invalid date, so `Date("bad").Before("2024-01-01")` and `Date("2024-01-01").EqualOrAfter("bad")` are false.
- **Reading doesn't validate.** `Scan`, JSON and XML decoding store a valid date as `YYYY-MM-DD` and an invalid one as is, without an error. An invalid date is an error for `Value` and written unchanged by `MarshalJSON`.
- **`fmt` prints the stored date.** `fmt.Println(date.Date("25.12.2023"))` prints `25.12.2023`. Use `Format(date.Layout)` or `MarshalText` for the normalized form.
- **Null ordering.** `Null` is before every date: `Null.Before(d)` and `d.After(Null)` are true, `Null.BeforeToday()` is true.
- **Nullable period types differ from `NullableDate`.** The `IsNull` of `NullableYearMonth` and `NullableYearQuarter` is only `""` and whitespace, while `NullableDate.IsNull` also covers the zero dates.
- **`Format.AssignString` with empty input** assigns a non-nil pointer to the zero time to a `*time.Time` destination, not nil.
- **`ParseTimeDefaultLayouts` is a shared package variable.** Pass layouts to `ParseTime` instead of changing it.

## Examples

Parse dates in different formats:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/language"
)

func main() {
	for _, s := range []string{"25.12.2023", "Dec 25, 2023", "1. Dezember 2016", "2023-12-25T10:00:00Z"} {
		d, err := date.Normalize(s)
		fmt.Println(d, err)
	}

	// Both orders are valid: day first by default, month first for English
	fmt.Println(date.Normalize("05/06/2024"))
	fmt.Println(date.Normalize("05/06/2024", language.EN))

	// Only one order is valid, so the hint changes nothing
	fmt.Println(date.Normalize("15.06.2024", language.EN))

	// Output:
	// 2023-12-25 <nil>
	// 2023-12-25 <nil>
	// 2016-12-01 <nil>
	// 2023-12-25 <nil>
	// 2024-06-05 <nil>
	// 2024-05-06 <nil>
	// 2024-06-15 <nil>
}
```

Reporting periods:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/date"
)

func main() {
	from, until, err := date.PeriodRange("2024-Q1")
	fmt.Println(from, until, err)

	// ISO week 1 of 2021 starts on Monday, January 4th
	from, until, _ = date.PeriodRange("2021-W01")
	fmt.Println(from, until)

	fmt.Println(date.YearMonth("2024-02").DateRange())

	d := date.Date("10.02.2024")
	fmt.Println(d.BeginningOfQuarter(), d.EndOfMonth(), d.AddMonths(1))

	// Output:
	// 2024-01-01 2024-03-31 <nil>
	// 2021-01-04 2021-01-10
	// 2024-02-01 2024-02-29
	// 2024-01-01 2024-02-29 2024-03-10
}
```

Optional dates in JSON and SQL, and finding dates in text:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/language"
)

type Invoice struct {
	Issued date.Date         `json:"issued"`
	Paid   date.NullableDate `json:"paid"`
}

func main() {
	b, _ := json.Marshal(Invoice{Issued: "2024-03-01", Paid: "15.3.2024"})
	fmt.Println(string(b))

	v, _ := date.Null.Value()
	fmt.Println(v)
	v, _ = date.Date("01.03.2024").Value()
	fmt.Println(v)

	text := []byte("Rechnung vom 25.12.2023, fällig am 8. Jänner 2024")
	for _, idx := range date.NewFinder(language.DE).FindAllIndex(text, -1) {
		fmt.Println(string(text[idx[0]:idx[1]]))
	}

	// Output:
	// {"issued":"2024-03-01","paid":"2024-03-15"}
	// <nil>
	// 2024-03-01
	// 25.12.2023
	// 8. Jänner 2024
}
```

## Related

- [`language`](../language/README.md): the language codes used as parsing hints.
- [`strfmt`](../strfmt/README.md): uses `Parser`, `Formatter` and `Format`, and formats `Date` and `NullableDate` with its `FormatConfig.Date` layout.
- [`nullable`](../nullable/README.md): `nullable.Time` returned by the `NullableDate` midnight methods.
- [Lenient parsing](../docs/explanation-lenient-parsing.md): why dates are parsed day first and how hints work.
- [Validation and normalization](../docs/explanation-validation-and-normalization.md): `Valid` vs `Validate` vs `Normalized` across the module.
- [Null conventions](../docs/explanation-null-conventions.md): `Date` vs `NullableDate` and the zero dates.
- [Store optional values](../docs/howto-store-optional-values.md)
- [Extract identifiers from text](../docs/howto-extract-identifiers-from-text.md): using `Finder`.
