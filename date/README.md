# date

ISO 8601 calendar dates (`YYYY-MM-DD`) with lenient locale-aware parsing, period arithmetic, year-month and year-quarter types, and SQL/JSON integration.

```
import "github.com/domonda/go-types/date"
```

## Date

```go
const (
    Layout    = time.DateOnly // "2006-01-02"
    Regex     = `\d{4}-\d{2}-\d{2}`
    Length    = 10
    MinLength = 6
    Invalid   Date = ""
)

type Date string
```

Empty string, `"0000-00-00"`, and `"0001-01-01"` round-trip as SQL NULL. For a real nullable, see `NullableDate`.

### Constructors

| Function                                         | Returns                                          |
|--------------------------------------------------|--------------------------------------------------|
| `Of(year, month, day)`                           | Date normalized through `time.Date`.             |
| `OfTime(t)`                                      | Date part of a `time.Time` (empty if zero).      |
| `OfTimePtr(*t)`                                  | `NullableDate`; nil/zero → null.                 |
| `OfToday()` / `OfYesterday()` / `OfTomorrow()`   | Local timezone shortcuts.                        |
| `OfNowInUTC()` / `OfTodayIn(loc)`                | Timezone-specific now.                           |
| `Parse(layout, value)`                           | `time.Parse` then take date part.                |
| `Must(str)`                                      | Normalized or panic.                             |
| `Normalize(str, lang...)`                        | Lenient parse with language hint.                |

`StringIsDate(str, lang...)` is the boolean shortcut.

### Methods (selected)

`Normalized(lang...)`, `Valid()`, `Validate()`, `Format(layout)`, `Time(loc)`, `MidnightInUTC()`, `MidnightInLocal()`, plus arithmetic — `AddDate`, `Before`, `After`, `Equal`, `Compare`. `Nullable()` converts to `NullableDate`.

### PeriodRange

```go
from, until, err := date.PeriodRange("2024-Q1")
// from = "2024-01-01", until = "2024-03-31"
```

Accepts:

| Format                                           | Example                                          | Meaning                                          |
|--------------------------------------------------|--------------------------------------------------|--------------------------------------------------|
| `YYYY`                                           | `2024`                                           | Full year.                                       |
| `YYYY-MM`                                        | `2024-06`                                        | Single month.                                    |
| `YYYY-Qn`                                        | `2024-Q3`                                        | Quarter.                                         |
| `YYYY-Hn`                                        | `2024-H2`                                        | Half year.                                       |
| `YYYY-Wnn`                                       | `2024-W23`                                       | ISO 8601 week.                                   |

## YearMonth

```go
type YearMonth string // YYYY-MM
const YearMonthLayout = "2006-01"
```

Constructors: `YearMonthFrom(year, month)`, `YearMonthOfTime(t)`, `YearMonthOfToday()`. Methods: `Validate()`, `Valid()`, plus a `NullableYearMonth` variant.

## YearQuarter

```go
type YearQuarter string // YYYY-Q#
```

Constructors: `YearQuarterFrom(year, quarter)`, `YearQuarterOfTime(t)`, `YearQuarterOfToday()`. Plus `NullableYearQuarter`.

## Format & Formatter

```go
type Format struct {
    Layout     string // time.Time-style layout
    NilString  string
    ZeroString string
}

type Formatter string // shorthand — layout as the value
```

Both implement `strfmt.Parser`. `Format` additionally supports reflective `AssignString(dest, source)` for scanning into `time.Time`, `Date`, `NullableDate`, etc.

## Parser

`Parser{}` implements `strfmt.Parser` using `Normalize`.

## Finder

```go
f := date.NewFinder(language.DE)
for _, idx := range f.FindAllIndex(text, -1) {
    fmt.Println(string(text[idx[0]:idx[1]]))
}
```

Scans byte slices for date-shaped substrings between word boundaries; uses the language hint for parsers like `"15.06.2024"` vs `"06/15/2024"`.

## time.Time helpers

```go
date.ParseTime(str, layouts...)
// Defaults to ParseTimeDefaultLayouts:
//   time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"
```

Returns the first matching layout.

## Related

- `language` — language hints for lenient parsing.
- `strfmt` — `Parser`/`Formatter` interfaces implemented by `Parser`, `Formatter`, and `Format`.
- `nullable.Time` — for full `time.Time` nullability.
