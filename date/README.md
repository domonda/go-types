# date

Calendar dates as ISO 8601 strings (`YYYY-MM-DD`) that parse the many ways people write dates: `"25.12.2023"`, `"Dec 25, 2023"`, `"1. Dezember 2016"`, `"2023-12-25T10:00:00Z"`. Use `Date` for required dates, `NullableDate` for optional ones, and `YearMonth` / `YearQuarter` for reporting periods. The types are strings, so they work as map keys, compare cheaply and are readable in logs and SQL.

    import "github.com/domonda/go-types/date"

## Types

| Type                  | Purpose                                            |
|-----------------------|----------------------------------------------------|
| `Date`                | A calendar date, canonical form `YYYY-MM-DD`. Empty is not valid. |
| `NullableDate`        | Like `Date`, but `Null` (`""`) and the zero dates are valid and SQL NULL. |
| `YearMonth`           | A month, `YYYY-MM`.                                |
| `NullableYearMonth`   | `YearMonth` where `""` is null.                    |
| `YearQuarter`         | A quarter, `YYYY-Qn` with an upper-case `Q`.       |
| `NullableYearQuarter` | `YearQuarter` where `""` is null.                  |
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

A `Date` holds whatever string you assign. Methods normalize on demand, so `Date("25.12.2023")` is valid and behaves like `"2023-12-25"` in every method except `==`. Normalize once at the boundary (`Normalize`, `ScanString`, `Scan`) to store the canonical form.

### Parsing

`Normalize(str, lang...)` and `Date.Normalized(lang...)` accept, case-insensitive and with lenient whitespace:

- ISO dates with `-`, `/`, `.` or space as separator: `"2006-01-02"`, `"2006/01/02"`.
- An ISO date followed by a `T` time (`"2006-01-02T15:04:05Z07:00"`) or a trailing `00:00:00` (`"30.11.2023 00:00:00"`). The time is dropped.
- Numeric day-month-year and month-day-year with 2 or 4 digit years: `"25.12.1975"`, `"25.12.75"`, `"12/25/1975"`.
- English and German month names and abbreviations: `"jan. 24 2012"`, `"1. Dezember 2016"`, `"Oct 7, 1970"`, `"September 17th, 2012"`, `"Oct 7, '70"`, `"2013-Feb-03"`, `"07-Feb-04"`. Common OCR misreadings like `"1an"` or `"apri1"` are accepted too.
- French, Italian and Spanish names for January only: `"2 janv. 2019"`, `"23/gen/2019"`, `"12 enero 2020"`. `"2 févr. 2019"` fails.
- English ordinals and a leading weekday: `"1st of dec. 2020"`, `"Fri, 03 Jul 2015"`.
- CJK separators: `"2014年04月08日"`.

Not supported: year-only or year-month strings (`"2014"`, `"2014-04"`), compact digits (`"20140601"`), dates whose three parts have fewer than 5 characters together (`"8/8/71"`), and times with a zone offset inside the string (`"Mon Jan 02 15:04:05 -0700 2006"`). Use `time.Parse` or `ParseTime` for those.

Rules for numeric dates:

- **Day first by default.** `"05/06/2024"` is 2024-06-05.
- **Language hint.** The first `lang` argument is a hint. Only its first two letters count, normalized, so `language.EN`, `"EN"`, `"en-US"` and `"eng"` all mean English. Only English changes the result: it reads month first.
- **The hint only decides ambiguous dates.** It applies only when both numbers are 1 to 12. `"15.06.2024"` and `"06/15/2024"` both give 2024-06-15 with any hint. The separator never matters.
- **Ordinal suffix.** A suffix like `"1st"` marks the day for ambiguous dates and wins over the hint: `"02 1st 2003"` and `"1st 02 2003"` are 2003-02-01, also with `language.EN`.
- **Two-digit years.** 00 to 44 become 2000 to 2044, 45 to 99 become 1945 to 1999.

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
| `Normalized(lang...)`              | The `YYYY-MM-DD` form, or the date unchanged and an error. The first `lang` is the order hint. |
| `ValidAndNormalized()`             | True only if the date is already a valid `YYYY-MM-DD`. |
| `NormalizedOrUnchanged(lang...)`   | Normalized form, or the date unchanged if it can't be parsed. |
| `NormalizedOrNull(lang...)`        | Normalized form as `NullableDate`, or `Null`.      |
| `IsZero()`                         | True for `""`, `"0000-00-00"` and `"0001-01-01"`.  |
| `Nullable()`                       | Converts to `NullableDate`; zero dates become `Null`. |
| `String()`                         | Normalized form, or the raw string if invalid.     |
| `ScanString(source, validate)`     | Assigns the normalized `source`. If parsing fails, returns the error when `validate` is true, else assigns `source` as is. |
| `ScanStringWithLang(source, lang)` | Assigns the normalized `source` using the hint `lang`. Returns `wasNormalized` (source was already `YYYY-MM-DD`), `monthMustBeFirst` (only month first is valid, like `"12/25/2023"`) and an error. Unchanged on error. |

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
| `Format(layout)`                  | Formats with a `time` layout. `""` if date or layout is empty. `Format(Layout)` returns `String()`. |

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

Implements: `fmt.Stringer`, `database/sql.Scanner`, `database/sql/driver.Valuer`, `strfmt.Scannable`, `types.Validator`, `types.ValidatErr` and `JSONSchema()` (`{"type":"string","format":"date"}`). `Date` has no JSON methods. It is not a `types.NormalizableValidator` because `Normalized` takes optional language hints.

## NullableDate

```go
const Null NullableDate = ""
```

`NullableDate` has the same string form as `Date`. `Null`, whitespace only (like for `country.NullableCode`) and the zero dates `"0000-00-00"` and `"0001-01-01"` are null, valid, and written as SQL NULL and JSON `null`.

| Function                          | Description                                        |
|-----------------------------------|----------------------------------------------------|
| `NormalizeNullable(str, lang...)` | Like `Normalize`; `Null` without error for `""` and zero dates. |
| `MustNullable(str)`               | `NormalizeNullable` without hint; panics on error. |
| `OfTimePtr(t)`                    | See `Date` functions.                              |

| Method                                             | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `IsNull()`                                         | True for `Null`, whitespace only and the zero dates. |
| `IsZero()`                                         | True for `Null` and the zero dates.                |
| `IsNotNull()`                                      | `!IsNull()`.                                       |
| `Get()`                                            | The `Date`; panics if null.                        |
| `GetOr(defaultDate)`, `DateOr(useIfNull)`          | The `Date`, or the argument if null.               |
| `Date()`                                           | Plain conversion to `Date`, no null check.         |
| `Set(d)`, `SetNull()`                              | Assign a `Date` or `Null`.                         |
| `Valid()`, `Validate()`                            | Null is valid, otherwise like `Date`.              |
| `ValidAndNotNull()`                                | Valid and not null, so false for the zero dates.   |
| `ValidAndNormalized()`                             | True for `Null` and for a valid `YYYY-MM-DD` that is not a zero date. |
| `Normalized(lang...)`                              | `Null` for null input, otherwise like `Date`.      |
| `NormalizedOrNull(lang...)`, `NormalizedOrUnchanged(lang...)` | As for `Date`.                                     |
| `String()`                                         | `""` for null, else like `Date`.                   |
| `StringOr(nullString)`                             | `String()`, or `nullString` if null.               |
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

Implements: `fmt.Stringer`, `database/sql.Scanner`, `database/sql/driver.Valuer`, `json.Marshaler`, `json.Unmarshaler`, `strfmt.Scannable`, `nullable.NullSetable[Date]` and `JSONSchema()` (string with format `date`, or null).

## Periods

| Function                                   | Description                                        |
|--------------------------------------------|----------------------------------------------------|
| `PeriodRange(period)`                      | First and last day of `"YYYY"`, `"YYYY-MM"`, `"YYYY-Qn"`, `"YYYY-Hn"` (half year) or `"YYYY-Wnn"` (ISO week). The letters may be lower case. |
| `YearRange(year)`                          | `"YYYY-01-01"` and `"YYYY-12-31"`.                 |
| `YearWeekMonday(year, week)`               | Monday of an ISO 8601 week.                        |
| `YearWeekRange(year, week)`                | Monday and Sunday of an ISO 8601 week.             |
| `FromUntilFromYearAndMonths(year, months)` | First and last day of a month range: `("2024", "3-9")` gives 2024-03-01 and 2024-09-30. Empty `months` means `"1-12"`; empty `year` returns empty dates and no error. |

An ISO week runs Monday to Sunday and can start or end in the neighbouring year: `PeriodRange("2021-W01")` is 2021-01-04 to 2021-01-10, `PeriodRange("2019-W01")` is 2018-12-31 to 2019-01-06. Week numbers 1 to 53 are accepted; week 53 of a year with only 52 weeks is week 1 of the next year.

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
| `Validate()`, `Valid()`              | Exactly `YYYY-MM`, year 0 to 3000, month 1 to 12.  |
| `IsZero()`                           | True for `""`, `"0000-00"` and `"0001-01"`.        |
| `String()`                           | The raw string.                                    |
| `Year()`, `Month()`                  | The parts, 0 if invalid.                           |
| `Date(day)`                          | `"YYYY-MM-DD"` for `day`, without validation (`Date(30)` of February is invalid). |
| `DateRange()`                        | First and last day of the month; `""`, `""` if invalid. |
| `Format(layout)`                     | `""` if empty; `YearMonthLayout` returns the string as is; other layouts format the 1st of the month. |
| `AddYears(n)`, `AddMonths(n)`        | Arithmetic; months roll over into years. Invalid values are returned unchanged. |
| `ContainsTime(t)`, `ContainsDate(d)` | True if `t` or `d` falls in the month (local time zone midnights). |
| `Compare(other)`                     | String comparison.                                 |
| `Nullable()`                         | Plain conversion to `NullableYearMonth`.           |

`NullableYearMonth` has the same methods (null returns 0, `""`, `false` or null) plus `IsNull()` (only `""`; `"0000-00"` is zero but not null), `IsNotNull()`, `ValidAndNotNull()` (false for null and zero values), `Get()` (panics if null), `GetOr(defaultYearMonth)`, `Set(ym)`, `SetNull()` and `StringOr(nullString)`. Its `Date` and `DateRange` return `NullableDate`. Implements `nullable.NullSetable[YearMonth]` and `json.Unmarshaler` (JSON `null` sets null, a string is assigned as is).

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
| `Validate()`, `Valid()`                            | Exactly `YYYY-Qn` with upper-case `Q`, year 0 to 3000, quarter 1 to 4. |
| `IsZero()`                                         | True for `""`, `"0000-Q0"` and `"0001-Q1"`.        |
| `String()`                                         | The raw string.                                    |
| `Year()`, `Quarter()`                              | The parts, 0 if invalid.                           |
| `FirstMonth()`                                     | January, April, July or October; 0 if invalid.     |
| `Date(monthInQuarter, day)`                        | Date in the quarter via `Of`, so values roll over; `""` if invalid. |
| `DateRange()`                                      | First and last day of the quarter; `""`, `""` if invalid. |
| `AddYears(n)`, `AddQuarters(n)`                    | Arithmetic; quarters roll over into years. Invalid values are returned unchanged. |
| `ContainsTime(t)`, `ContainsDate(d)`, `ContainsYearMonth(ym)` | True if the value falls in the quarter.            |
| `Compare(other)`                                   | String comparison.                                 |
| `Nullable()`                                       | Plain conversion to `NullableYearQuarter`.         |

`NullableYearQuarter` has the same methods plus the same null helpers as `NullableYearMonth`. Its `Date` and `DateRange` return `NullableDate`. Implements `nullable.NullSetable[YearQuarter]` and `json.Unmarshaler` like `NullableYearMonth`.

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
| `(*Format).FormatString(val)`                | `NilString` for nil or invalid values, `ZeroString` for zero values with `Format`/`IsZero` methods (`Date`, `NullableDate`, `time.Time`), else `Format(Layout)`. Other types are an error. |

`Parser` and `Formatter` implement `strfmt.Parser` as values, `Format` only as `*Format`. None of them implements `strfmt.Formatter`.

## Finder

| Identifier                       | Description                                        |
|----------------------------------|----------------------------------------------------|
| `NewFinder(lang...)`             | Returns a `*Finder` with the first two letters of the first `lang` as hint. |
| `Finder.LangHint`                | The language hint field.                           |
| `(*Finder).FindAllIndex(str, n)` | `[start, end]` byte offsets of dates in `str`. At most `n` results; `n < 0` returns all, `n == 0` none. |

`*Finder` implements `types.Finder`. It tries spans of one to three whitespace-separated words and trims surrounding punctuation, so `"25.12.2023,"` is found as `"25.12.2023"`. It reports the shortest span that parses at each start word. A span is found exactly if `Normalize` accepts it, so an ISO 8601 date-time like `"2023-12-25T10:00:00Z"` is found as a whole.

## time.Time helpers

| Identifier                   | Description                                        |
|------------------------------|----------------------------------------------------|
| `ParseTime(str, layouts...)` | Tries `time.Parse` with each layout and returns the first success. Defaults to `ParseTimeDefaultLayouts`. |
| `ParseTimeDefaultLayouts`    | `time.RFC3339Nano`, `time.RFC3339`, `"2006-01-02 15:04:05"`. A package variable. |

## Serialization

SQL `Value()` and JSON encoding:

| Value                                              | SQL `Value()`                              | JSON (`json.Marshal`)         | `String()`       |
|----------------------------------------------------|--------------------------------------------|-------------------------------|------------------|
| `Date("")`                                         | `NULL`                                     | `""`                          | `""`             |
| `Date("0001-01-01")`, `Date("0000-00-00")`         | `NULL`                                     | the string as is              | the string as is |
| `Date("25.12.2023")` (valid, not normalized)       | `"2023-12-25"`                             | `"25.12.2023"`                | `"2023-12-25"`   |
| `Date("2023-12-25")`                               | `"2023-12-25"`                             | `"2023-12-25"`                | `"2023-12-25"`   |
| `Date("bad")` (invalid)                            | error                                      | `"bad"`                       | `"bad"`          |
| `NullableDate` null or zero date                   | `NULL`                                     | `null`                        | `""`             |
| `NullableDate("25.12.2023")`                       | `"2023-12-25"`                             | `"25.12.2023"`                | `"2023-12-25"`   |
| `NullableDate("bad")`                              | error                                      | `"bad"`                       | `"bad"`          |
| `YearMonth`, `YearQuarter` and their nullable types | the raw string, `""` for null (not `NULL`) | the raw string, `""` for null | the raw string   |

Reading values:

| Input                                      | `Date`                    | `NullableDate`              |
|--------------------------------------------|---------------------------|-----------------------------|
| `Scan(nil)`                                | `""`                      | `Null`                      |
| `Scan("")`, `"0000-00-00"`, `"0001-01-01"` | stored as is              | `Null`                      |
| `Scan` whitespace only                     | an error                  | `Null`                      |
| `Scan` other `string`                      | normalized, or an error   | normalized, or an error     |
| `Scan([]byte)`                             | like `string`             | like `string`               |
| `Scan(time.Time)`                          | `OfTime(t)`, `""` if zero | `OfTime(t)`, `Null` if zero |
| `Scan` of other types                      | error                     | error                       |
| JSON string                                | stored as is, not checked | stored as is, not checked   |
| JSON `null`                                | field left unchanged      | `Null`                      |
| other JSON values                          | error                     | error                       |

`Date`, `YearMonth` and `YearQuarter` have no `UnmarshalJSON`, so `encoding/json` stores any JSON string as is and ignores JSON `null`. The `UnmarshalJSON` of `NullableDate`, `NullableYearMonth` and `NullableYearQuarter` also stores a JSON string as is, but sets null for JSON `null`. Call `Validate` or `Normalized` after decoding. `YearMonth`, `YearQuarter`, `NullableYearMonth` and `NullableYearQuarter` have no SQL methods: `database/sql` writes them as plain strings and fails to scan NULL into them (`converting NULL to string is unsupported`); scan into a `*YearMonth` to accept NULL.

## Gotchas

- **Valid, Validate, Normalized, ValidAndNormalized.** `Valid` and `Validate` accept any parseable form: `Date("25.12.2023").Valid()` is true. `Normalized` returns the canonical form. `ValidAndNormalized` is true only for an exact `YYYY-MM-DD`. `Date("")` is invalid, `NullableDate("")` is valid.
- **`==` compares raw strings.** `Date("25.12.2023") != Date("2023-12-25")`. Use `NormalizedEqual` or `Compare`, or normalize on input.
- **Zero dates.** `"0001-01-01"` is `Valid()` and `IsZero()`. `"0000-00-00"` is `IsZero()` but not valid. Both are written as SQL NULL and read back as `""`. `Date` encodes them in JSON as they are, `NullableDate` as `null`.
- **Day first.** `"05/06/2024"` is June 5 unless the hint is English. The hint never changes a date that has only one valid reading. Only the first `lang` argument is used.
- **Two-digit years** pivot at 45: `"25.12.44"` is 2044-12-25, `"25.12.45"` is 1945-12-25.
- **Invalid and zero dates become the zero time.** `Date("bad").Format("02.01.2006")` returns `"01.01.0001"` and `Date("").AddDays(1)` returns `"0001-01-02"`. Check `Valid` (or `IsZero`) before formatting or doing arithmetic. `Format(Layout)` instead returns an invalid date unchanged. Comparisons don't use the zero time: `After`, `Before`, `WithinIncl`, `BetweenExcl`, `AfterTime` and `BeforeTime` are false for an invalid date, so `Date("bad").Before("2024-01-01")` and `Date("2024-01-01").EqualOrAfter("bad")` are false.
- **JSON is not normalized.** `Date` and `NullableDate` encode the raw string, and decoding does not validate. Normalize before encoding if consumers expect `YYYY-MM-DD`.
- **Null ordering.** `Null` is before every date: `Null.Before(d)` and `d.After(Null)` are true, `Null.BeforeToday()` is true.
- **Upper-case `Q`.** `YearQuarter("2024-q1").Valid()` is false, while `PeriodRange("2024-q1")` works.
- **Nullable period types differ from `NullableDate`.** The `IsNull` of `NullableYearMonth` and `NullableYearQuarter` is only `""`, while `NullableDate.IsNull` also covers the zero dates.
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
	b, _ := json.Marshal(Invoice{Issued: "2024-03-01"})
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
	// {"issued":"2024-03-01","paid":null}
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
