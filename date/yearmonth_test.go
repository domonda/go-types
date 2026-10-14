package date

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/internal/normtest"
)

// Constructor tests

func TestYearMonthFrom(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month time.Month
		want  YearMonth
	}{
		{name: "regular month", year: 2023, month: time.March, want: "2023-03"},
		{name: "january", year: 2020, month: time.January, want: "2020-01"},
		{name: "december", year: 2021, month: time.December, want: "2021-12"},
		{name: "year padding", year: 999, month: time.June, want: "0999-06"},
		{name: "month padding", year: 2023, month: time.May, want: "2023-05"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := YearMonthFrom(tt.year, tt.month)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestYearMonthOfTime(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want YearMonth
	}{
		{name: "regular time", t: time.Date(2023, time.July, 15, 12, 30, 0, 0, time.UTC), want: "2023-07"},
		{name: "first day", t: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), want: "2020-01"},
		{name: "last day", t: time.Date(2021, time.December, 31, 23, 59, 59, 0, time.UTC), want: "2021-12"},
		{name: "zero time", t: time.Time{}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := YearMonthOfTime(tt.t)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestYearMonthOfToday(t *testing.T) {
	result := YearMonthOfToday()
	require.True(t, result.Valid())
	require.False(t, result.IsZero())

	// Verify it matches current time
	now := time.Now()
	expected := YearMonthOfTime(now)
	require.Equal(t, expected, result)
}

// Validation tests

func TestYearMonth_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ym      YearMonth
		wantErr bool
	}{
		{name: "valid", ym: "2023-03", wantErr: false},
		{name: "valid january", ym: "2020-01", wantErr: false},
		{name: "valid december", ym: "2021-12", wantErr: false},
		{name: "surrounding whitespace", ym: " 2023-03\t", wantErr: false},
		{name: "whitespace only", ym: "  ", wantErr: true},
		{name: "empty", ym: "", wantErr: true},
		{name: "too short", ym: "2023-3", wantErr: true},
		{name: "too long", ym: "2023-123", wantErr: true},
		{name: "no dash", ym: "202303", wantErr: true},
		{name: "invalid month 0", ym: "2023-00", wantErr: true},
		{name: "invalid month 13", ym: "2023-13", wantErr: true},
		{name: "invalid year", ym: "abcd-03", wantErr: true},
		{name: "year too large", ym: "3001-03", wantErr: true},
		{name: "invalid format", ym: "2023/03", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ym.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestYearMonth_Valid(t *testing.T) {
	tests := []struct {
		name string
		ym   YearMonth
		want bool
	}{
		{name: "valid", ym: "2023-03", want: true},
		{name: "invalid month", ym: "2023-13", want: false},
		{name: "empty", ym: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Valid()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestYearMonth_IsZero(t *testing.T) {
	tests := []struct {
		name string
		ym   YearMonth
		want bool
	}{
		{name: "empty string", ym: "", want: true},
		{name: "0000-00", ym: "0000-00", want: true},
		{name: "0001-01", ym: "0001-01", want: true},
		{name: "valid date", ym: "2023-03", want: false},
		{name: "0001-02", ym: "0001-02", want: false},
		{name: "whitespace only", ym: "  ", want: true},
		{name: "padded 0001-01", ym: " 0001-01 ", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.IsZero()
			require.Equal(t, tt.want, got)
		})
	}
}

// String and conversion tests

func TestYearMonth_String(t *testing.T) {
	ym := YearMonth("2023-07")
	require.Equal(t, "2023-07", ym.String())
}

func TestYearMonth_Year(t *testing.T) {
	tests := []struct {
		name string
		ym   YearMonth
		want int
	}{
		{name: "regular year", ym: "2023-07", want: 2023},
		{name: "different year", ym: "1999-12", want: 1999},
		{name: "zero year", ym: "0000-00", want: 0},
		{name: "invalid", ym: "invalid", want: 0},
		// Documented to return 0 for invalid values, so calling it
		// on unvalidated input must not panic on a short string.
		{name: "empty", ym: "", want: 0},
		{name: "too short", ym: "2023-1", want: 0},
		{name: "year only", ym: "2023", want: 0},
		{name: "invalid month", ym: "2023-13", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Year()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestYearMonth_Month(t *testing.T) {
	tests := []struct {
		name string
		ym   YearMonth
		want time.Month
	}{
		{name: "march", ym: "2023-03", want: time.March},
		{name: "january", ym: "2023-01", want: time.January},
		{name: "december", ym: "2023-12", want: time.December},
		{name: "invalid", ym: "invalid", want: 0},
		// Documented to return 0 for invalid values, so calling it
		// on unvalidated input must not panic on a short string.
		{name: "empty", ym: "", want: 0},
		{name: "too short", ym: "2023-1", want: 0},
		{name: "invalid month", ym: "2023-13", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Month()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestYearMonth_Date(t *testing.T) {
	tests := []struct {
		name string
		ym   YearMonth
		day  int
		want Date
	}{
		{name: "first day", ym: "2023-03", day: 1, want: "2023-03-01"},
		{name: "mid month", ym: "2023-03", day: 15, want: "2023-03-15"},
		{name: "last day", ym: "2023-03", day: 31, want: "2023-03-31"},
		{name: "day padding", ym: "2023-03", day: 5, want: "2023-03-05"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Date(tt.day)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestYearMonth_Nullable(t *testing.T) {
	ym := YearMonth("2023-07")
	nullable := ym.Nullable()
	require.Equal(t, NullableYearMonth("2023-07"), nullable)
	require.False(t, nullable.IsNull())
}

func TestYearMonth_DateRange(t *testing.T) {
	tests := []struct {
		name      string
		ym        YearMonth
		wantFrom  Date
		wantUntil Date
	}{
		{name: "march 2023", ym: "2023-03", wantFrom: "2023-03-01", wantUntil: "2023-03-31"},
		{name: "february 2020 (leap)", ym: "2020-02", wantFrom: "2020-02-01", wantUntil: "2020-02-29"},
		{name: "february 2021", ym: "2021-02", wantFrom: "2021-02-01", wantUntil: "2021-02-28"},
		{name: "april", ym: "2023-04", wantFrom: "2023-04-01", wantUntil: "2023-04-30"},
		{name: "december", ym: "2023-12", wantFrom: "2023-12-01", wantUntil: "2023-12-31"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrom, gotUntil := tt.ym.DateRange()
			require.Equal(t, tt.wantFrom, gotFrom)
			// Both ends of the range are used as inclusive Date bounds
			// (e.g. in SQL BETWEEN), so untilDate must be a full valid Date
			// of the last day of the month, not the "YYYY-MM" year-month.
			require.Equal(t, tt.wantUntil, gotUntil)
			require.True(t, gotUntil.ValidAndNormalized())
		})
	}
}

func TestYearMonth_DateRange_Invalid(t *testing.T) {
	// An invalid year-month has no month to span, so like
	// YearQuarter.DateRange it must give an empty range instead of
	// dates computed from non-existent months that look like real bounds.
	for _, ym := range []YearMonth{"2023-13", "2023-00", "bad", ""} {
		t.Run(string(ym), func(t *testing.T) {
			gotFrom, gotUntil := ym.DateRange()
			require.Equal(t, Date(""), gotFrom)
			require.Equal(t, Date(""), gotUntil)
		})
	}
}

func TestYearMonth_Format(t *testing.T) {
	tests := []struct {
		name   string
		ym     YearMonth
		layout string
		want   string
	}{
		{name: "YearMonthLayout", ym: "2023-07", layout: YearMonthLayout, want: "2023-07"},
		{name: "custom layout", ym: "2023-07", layout: "January 2006", want: "July 2023"},
		{name: "empty layout", ym: "2023-07", layout: "", want: ""},
		{name: "empty yearmonth", ym: "", layout: YearMonthLayout, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Format(tt.layout)
			require.Equal(t, tt.want, got)
		})
	}
}

// Arithmetic tests

func TestYearMonth_AddYears(t *testing.T) {
	tests := []struct {
		name  string
		ym    YearMonth
		years int
		want  YearMonth
	}{
		{name: "add 1 year", ym: "2023-03", years: 1, want: "2024-03"},
		{name: "add 5 years", ym: "2020-06", years: 5, want: "2025-06"},
		{name: "subtract 1 year", ym: "2023-03", years: -1, want: "2022-03"},
		{name: "add 0 years", ym: "2023-03", years: 0, want: "2023-03"},
		// An invalid year-month has no year to add to. It must stay
		// unchanged, so it stays invalid and Validate reports the original
		// value, instead of becoming a bogus year-month like "0001-00".
		{name: "invalid unchanged", ym: "bad", years: 1, want: "bad"},
		{name: "invalid month unchanged", ym: "2023-13", years: 1, want: "2023-13"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.AddYears(tt.years)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestYearMonth_AddMonths(t *testing.T) {
	tests := []struct {
		name   string
		ym     YearMonth
		months int
		want   YearMonth
	}{
		{name: "add 1 month", ym: "2023-03", months: 1, want: "2023-04"},
		{name: "add 12 months", ym: "2023-03", months: 12, want: "2024-03"},
		{name: "month overflow", ym: "2023-12", months: 1, want: "2024-01"},
		{name: "subtract 1 month", ym: "2023-03", months: -1, want: "2023-02"},
		{name: "month underflow", ym: "2023-01", months: -1, want: "2022-12"},
		{name: "add 0 months", ym: "2023-03", months: 0, want: "2023-03"},
		// Invalid stays unchanged like in AddYears, not "0001-02"
		{name: "invalid unchanged", ym: "bad", months: 1, want: "bad"},
		{name: "invalid month unchanged", ym: "2023-13", months: 1, want: "2023-13"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.AddMonths(tt.months)
			require.Equal(t, tt.want, got)
		})
	}
}

// Comparison and contains tests

func TestYearMonth_ContainsTime(t *testing.T) {
	ym := YearMonth("2023-07")

	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{name: "first day", t: time.Date(2023, time.July, 1, 0, 0, 0, 0, time.Local), want: true},
		{name: "mid month", t: time.Date(2023, time.July, 15, 12, 30, 0, 0, time.Local), want: true},
		{name: "last day", t: time.Date(2023, time.July, 31, 23, 59, 59, 0, time.Local), want: true},
		{name: "before", t: time.Date(2023, time.June, 30, 23, 59, 59, 0, time.Local), want: false},
		{name: "after", t: time.Date(2023, time.August, 1, 0, 0, 0, 0, time.Local), want: false},
		{name: "different year", t: time.Date(2022, time.July, 15, 0, 0, 0, 0, time.Local), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ym.ContainsTime(tt.t)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestYearMonth_ContainsDate(t *testing.T) {
	ym := YearMonth("2023-07")

	tests := []struct {
		name string
		date Date
		want bool
	}{
		{name: "first day", date: "2023-07-01", want: true},
		{name: "mid month", date: "2023-07-15", want: true},
		{name: "last day", date: "2023-07-31", want: true},
		{name: "before", date: "2023-06-30", want: false},
		{name: "after", date: "2023-08-01", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ym.ContainsDate(tt.date)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestYearMonth_Compare(t *testing.T) {
	tests := []struct {
		name  string
		ym    YearMonth
		other YearMonth
		want  int
	}{
		{name: "before", ym: "2023-03", other: "2023-07", want: -1},
		{name: "after", ym: "2023-07", other: "2023-03", want: 1},
		{name: "equal", ym: "2023-07", other: "2023-07", want: 0},
		{name: "different years", ym: "2022-12", other: "2023-01", want: -1},
		{name: "equal when normalized", ym: " 2023-07", other: "2023-07", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Compare(tt.other)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestYearMonth_ValidateRejectsYearZero protects that a YearMonth is only
// valid if its dates are: Date has no year 0, so "0000-05" would add months
// through an invalid date and give "0001-02" for AddMonths(1).
func TestYearMonth_ValidateRejectsYearZero(t *testing.T) {
	require.Error(t, YearMonth("0000-05").Validate())
	require.NoError(t, YearMonth("0001-05").Validate())
	require.Equal(t, YearMonth("0000-05"), YearMonth("0000-05").AddMonths(1), "invalid stays unchanged")
}

// Normalized trims the surrounding whitespace, which a typed or OCR read
// period often has, and returns an invalid value unchanged with an error,
// like the other normalizing types, so the error can name the input.
func TestYearMonth_Normalized(t *testing.T) {
	for input, want := range map[YearMonth]YearMonth{
		"2023-07":    "2023-07",
		" 2023-07\t": "2023-07",
	} {
		got, err := input.Normalized()
		require.NoError(t, err, "%q", input)
		require.Equal(t, want, got, "%q", input)
		require.Equal(t, input == want, input.ValidAndNormalized(), "%q", input)
	}
	for _, input := range []YearMonth{"", "  ", "2023-13", "0000-05", "2023-7", "07/2023"} {
		got, err := input.Normalized()
		require.Error(t, err, "%q", input)
		require.Equal(t, input, got, "%q unchanged", input)
		require.False(t, input.ValidAndNormalized(), "%q", input)
	}
}

// The accessors work on the normalized year-month,
// because Valid accepts the spellings that Normalized accepts.
func TestYearMonth_AccessorsNormalize(t *testing.T) {
	ym := YearMonth(" 2023-07 ")
	require.Equal(t, 2023, ym.Year())
	require.Equal(t, time.July, ym.Month())
	require.Equal(t, Date("2023-07-15"), ym.Date(15))
	require.Equal(t, "2023-07", ym.Format(YearMonthLayout))
	require.Equal(t, YearMonth("2023-09"), ym.AddMonths(2))
}

// Reading JSON normalizes like Date.UnmarshalJSON, so a value read from
// JSON is canonical if it is valid and compares with ==. An invalid value
// is kept without an error, so one bad period doesn't fail a whole
// document, and Validate reports it.
func TestYearMonth_UnmarshalJSON(t *testing.T) {
	for data, want := range map[string]YearMonth{
		`"2023-07"`:   "2023-07",
		`" 2023-07 "`: "2023-07",
		`"2023-13"`:   "2023-13",
		`"bad"`:       "bad",
		`""`:          "",
		`"  "`:        "",
	} {
		ym := YearMonth("2024-01")
		require.NoError(t, json.Unmarshal([]byte(data), &ym), "json.Unmarshal(%s)", data)
		require.Equal(t, want, ym, "json.Unmarshal(%s)", data)
	}

	// JSON null leaves a non-nullable value unchanged like encoding/json
	// does for a string type, any other JSON value is an error
	ym := YearMonth("2024-01")
	require.NoError(t, json.Unmarshal([]byte(`null`), &ym))
	require.Equal(t, YearMonth("2024-01"), ym)
	require.Error(t, json.Unmarshal([]byte(`202307`), &ym))
	require.Equal(t, YearMonth("2024-01"), ym)
}

// TestYearMonth_ReadingAndWriting checks the reading and writing contract
// shared by the normalizing types, like for Date: every reading path stores
// a valid year-month normalized, a zero one as "" and an invalid one
// unchanged, Value refuses an invalid year-month, and fmt prints it as stored.
func TestYearMonth_ReadingAndWriting(t *testing.T) {
	normtest.Run(t, normtest.Spec[YearMonth]{
		Canonical:    "2023-07",
		Variants:     []YearMonth{" 2023-07 ", "\t2023-07\n"},
		Invalid:      []YearMonth{"2023-13", "2023-7", "0000-05", "07/2023", "bad"},
		Blank:        []YearMonth{"0001-01", "0000-00", " 0001-01 "},
		StoredString: true,
	})
}

// A date or timestamp column scans as its year-month, like Date.Scan
// scans it as its date, and the zero time as the zero value "".
func TestYearMonth_ScanTime(t *testing.T) {
	var ym YearMonth
	require.NoError(t, ym.Scan(time.Date(2023, 7, 15, 10, 0, 0, 0, time.UTC)))
	require.Equal(t, YearMonth("2023-07"), ym)
	require.NoError(t, ym.Scan(time.Time{}))
	require.Equal(t, YearMonth(""), ym)
}

// A zero year-month is blank like a zero Date, so Format returns ""
// for every layout, like Date.Format and NullableYearMonth.Format.
func TestYearMonth_FormatZero(t *testing.T) {
	for _, ym := range []YearMonth{"", "  ", "0001-01", "0000-00", " 0001-01 "} {
		require.Equal(t, "", ym.Format(YearMonthLayout), "%q", ym)
		require.Equal(t, "", ym.Format("01/2006"), "%q", ym)
	}
}

// Every spelling of the blank value is the same "no value" for IsZero,
// Value and MarshalJSON, so Compare must not order them apart,
// like Date.Compare doesn't order the zero dates apart.
func TestYearMonthAndYearQuarter_CompareBlank(t *testing.T) {
	require.Equal(t, 0, YearMonth("").Compare("0001-01"))
	require.Equal(t, 0, YearMonth(" ").Compare(""))
	require.Equal(t, 0, YearMonth("0000-00").Compare("0001-01"))
	require.Equal(t, 0, YearQuarter("0001-q1").Compare("0001-Q1"))
	require.Equal(t, 0, YearQuarter("0000-Q0").Compare(""))
	require.Equal(t, 0, YearQuarter(" ").Compare(""))
	require.Equal(t, -1, YearMonth("").Compare("2023-01"))
	require.Equal(t, -1, YearQuarter("0000-Q0").Compare("2023-Q1"))
}

// Like Date.Scan, a time.Time is read in its own location, not in UTC,
// so a TIMESTAMPTZ returned in the session time zone gives the period
// of that zone.
func TestYearMonthAndYearQuarter_ScanTimeLocation(t *testing.T) {
	loc := time.FixedZone("UTC-5", -5*60*60)
	tm := time.Date(2023, 7, 1, 2, 0, 0, 0, time.UTC).In(loc)
	var ym YearMonth
	require.NoError(t, ym.Scan(tm))
	require.Equal(t, YearMonth("2023-06"), ym)
	var yq YearQuarter
	require.NoError(t, yq.Scan(tm))
	require.Equal(t, YearQuarter("2023-Q2"), yq)
}

// The schemas document the canonical form that MarshalJSON writes,
// and the nullable ones also null, which is their default,
// like TestNullableDate_JSONSchema checks for NullableDate.
func TestYearMonthAndYearQuarter_JSONSchema(t *testing.T) {
	re := regexp.MustCompile(YearMonth("").JSONSchema().Pattern)
	require.True(t, re.MatchString("2023-07"))
	require.False(t, re.MatchString("2023-13"))
	re = regexp.MustCompile(YearQuarter("").JSONSchema().Pattern)
	require.True(t, re.MatchString("2023-Q3"))
	require.False(t, re.MatchString("2023-Q5"))

	for name, schema := range map[string]interface {
		JSONSchema() *jsonschema.Schema
	}{
		"NullableYearMonth":   NullableYearMonth(""),
		"NullableYearQuarter": NullableYearQuarter(""),
	} {
		s := schema.JSONSchema()
		require.Len(t, s.OneOf, 2, name)
		require.Equal(t, "null", s.OneOf[1].Type, "%s: null must be allowed", name)
		data, err := json.Marshal(s.Default)
		require.NoError(t, err)
		require.Equal(t, "null", string(data), "%s: the default is JSON null", name)
		re := regexp.MustCompile(s.OneOf[0].Pattern)
		require.True(t, re.MatchString(map[string]string{"NullableYearMonth": "2023-07", "NullableYearQuarter": "2023-Q3"}[name]), name)
	}
}
