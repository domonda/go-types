package date

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/internal/normtest"
)

func TestNullableYearMonth_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ym      NullableYearMonth
		wantErr bool
	}{
		{name: "valid", ym: "2023-03", wantErr: false},
		{name: "null", ym: YearMonthNull, wantErr: false},
		{name: "invalid", ym: "2023-13", wantErr: true},
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

func TestNullableYearMonth_Valid(t *testing.T) {
	tests := []struct {
		name string
		ym   NullableYearMonth
		want bool
	}{
		{name: "valid", ym: "2023-03", want: true},
		{name: "null", ym: YearMonthNull, want: true},
		{name: "invalid", ym: "2023-13", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Valid()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_ValidAndNotNull(t *testing.T) {
	tests := []struct {
		name string
		ym   NullableYearMonth
		want bool
	}{
		{name: "valid", ym: "2023-03", want: true},
		{name: "null", ym: YearMonthNull, want: false},
		{name: "invalid", ym: "2023-13", want: false},
		// Documented as "not Null or Zero": "0001-01" is the year-month of
		// the zero time.Time, a placeholder and not a real reporting period.
		{name: "zero 0001-01", ym: "0001-01", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.ValidAndNotNull()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_IsZero(t *testing.T) {
	tests := []struct {
		name string
		ym   NullableYearMonth
		want bool
	}{
		{name: "empty", ym: "", want: true},
		{name: "0000-00", ym: "0000-00", want: true},
		{name: "0001-01", ym: "0001-01", want: true},
		{name: "valid", ym: "2023-03", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.IsZero()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_IsNull(t *testing.T) {
	tests := []struct {
		name string
		ym   NullableYearMonth
		want bool
	}{
		{name: "null", ym: YearMonthNull, want: true},
		{name: "empty", ym: "", want: true},
		{name: "whitespace only", ym: " \t", want: true},
		// The zero values are null like the zero dates of NullableDate
		{name: "zero 0000-00", ym: "0000-00", want: true},
		{name: "zero 0001-01", ym: "0001-01", want: true},
		{name: "valid", ym: "2023-03", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.IsNull()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_IsNotNull(t *testing.T) {
	tests := []struct {
		name string
		ym   NullableYearMonth
		want bool
	}{
		{name: "null", ym: YearMonthNull, want: false},
		{name: "valid", ym: "2023-03", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.IsNotNull()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_Get(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		ym := NullableYearMonth("2023-03")
		got := ym.Get()
		require.Equal(t, YearMonth("2023-03"), got)
	})

	t.Run("panic on null", func(t *testing.T) {
		ym := NullableYearMonth(YearMonthNull)
		require.Panics(t, func() {
			ym.Get()
		})
	})
}

func TestNullableYearMonth_GetOr(t *testing.T) {
	tests := []struct {
		name     string
		ym       NullableYearMonth
		default_ YearMonth
		want     YearMonth
	}{
		{name: "valid value", ym: "2023-03", default_: "2020-01", want: "2023-03"},
		{name: "null value", ym: YearMonthNull, default_: "2020-01", want: "2020-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.GetOr(tt.default_)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_Set(t *testing.T) {
	var ym NullableYearMonth
	ym.Set(YearMonth("2023-07"))
	require.Equal(t, NullableYearMonth("2023-07"), ym)
	require.False(t, ym.IsNull())
}

func TestNullableYearMonth_SetNull(t *testing.T) {
	ym := NullableYearMonth("2023-07")
	ym.SetNull()
	require.Equal(t, NullableYearMonth(YearMonthNull), ym)
	require.True(t, ym.IsNull())
}

func TestNullableYearMonth_String(t *testing.T) {
	tests := []struct {
		name string
		ym   NullableYearMonth
		want string
	}{
		{name: "valid", ym: "2023-07", want: "2023-07"},
		{name: "null", ym: YearMonthNull, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.String()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_StringOr(t *testing.T) {
	tests := []struct {
		name       string
		ym         NullableYearMonth
		nullString string
		want       string
	}{
		{name: "valid", ym: "2023-07", nullString: "NULL", want: "2023-07"},
		{name: "null", ym: YearMonthNull, nullString: "NULL", want: "NULL"},
		{name: "null with dash", ym: YearMonthNull, nullString: "-", want: "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.StringOr(tt.nullString)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_Compare(t *testing.T) {
	tests := []struct {
		name  string
		ym    NullableYearMonth
		other NullableYearMonth
		want  int
	}{
		{name: "before", ym: "2023-03", other: "2023-07", want: -1},
		{name: "after", ym: "2023-07", other: "2023-03", want: 1},
		{name: "equal", ym: "2023-07", other: "2023-07", want: 0},
		{name: "null vs value", ym: YearMonthNull, other: "2023-03", want: -1},
		{name: "value vs null", ym: "2023-03", other: YearMonthNull, want: 1},
		{name: "both null", ym: YearMonthNull, other: YearMonthNull, want: 0},
		{name: "whitespace is null", ym: "  ", other: YearMonthNull, want: 0},
		{name: "equal when normalized", ym: "2023-07 ", other: "2023-07", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Compare(tt.other)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_Year(t *testing.T) {
	tests := []struct {
		name string
		ym   NullableYearMonth
		want int
	}{
		{name: "valid", ym: "2023-07", want: 2023},
		{name: "different year", ym: "1999-12", want: 1999},
		{name: "null", ym: YearMonthNull, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Year()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_Month(t *testing.T) {
	tests := []struct {
		name string
		ym   NullableYearMonth
		want time.Month
	}{
		{name: "march", ym: "2023-03", want: time.March},
		{name: "january", ym: "2023-01", want: time.January},
		{name: "december", ym: "2023-12", want: time.December},
		{name: "null", ym: YearMonthNull, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Month()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_Date(t *testing.T) {
	tests := []struct {
		name string
		ym   NullableYearMonth
		day  int
		want NullableDate
	}{
		{name: "first day", ym: "2023-03", day: 1, want: NullableDate("2023-03-01")},
		{name: "mid month", ym: "2023-03", day: 15, want: NullableDate("2023-03-15")},
		{name: "null", ym: YearMonthNull, day: 15, want: Null},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Date(tt.day)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_DateRange(t *testing.T) {
	tests := []struct {
		name      string
		ym        NullableYearMonth
		wantFrom  NullableDate
		wantUntil NullableDate
	}{
		{name: "march 2023", ym: "2023-03", wantFrom: NullableDate("2023-03-01"), wantUntil: NullableDate("2023-03-31")},
		{name: "february 2024 (leap)", ym: "2024-02", wantFrom: NullableDate("2024-02-01"), wantUntil: NullableDate("2024-02-29")},
		{name: "null", ym: YearMonthNull, wantFrom: Null, wantUntil: Null},
		// An invalid year-month has no range, like YearMonth.DateRange
		{name: "invalid", ym: "2023-13", wantFrom: Null, wantUntil: Null},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrom, gotUntil := tt.ym.DateRange()
			require.Equal(t, tt.wantFrom, gotFrom)
			// untilDate is the inclusive last day of the month, not "YYYY-MM"
			require.Equal(t, tt.wantUntil, gotUntil)
		})
	}
}

func TestNullableYearMonth_Format(t *testing.T) {
	tests := []struct {
		name   string
		ym     NullableYearMonth
		layout string
		want   string
	}{
		{name: "YearMonthLayout", ym: "2023-07", layout: YearMonthLayout, want: "2023-07"},
		{name: "custom layout", ym: "2023-07", layout: "January 2006", want: "July 2023"},
		{name: "null", ym: YearMonthNull, layout: YearMonthLayout, want: ""},
		{name: "null with custom layout", ym: YearMonthNull, layout: "January 2006", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.Format(tt.layout)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_AddYears(t *testing.T) {
	tests := []struct {
		name  string
		ym    NullableYearMonth
		years int
		want  NullableYearMonth
	}{
		{name: "add 1 year", ym: "2023-03", years: 1, want: "2024-03"},
		{name: "add 5 years", ym: "2020-06", years: 5, want: "2025-06"},
		{name: "subtract 1 year", ym: "2023-03", years: -1, want: "2022-03"},
		{name: "null", ym: YearMonthNull, years: 1, want: YearMonthNull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.AddYears(tt.years)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_AddMonths(t *testing.T) {
	tests := []struct {
		name   string
		ym     NullableYearMonth
		months int
		want   NullableYearMonth
	}{
		{name: "add 1 month", ym: "2023-03", months: 1, want: "2023-04"},
		{name: "add 12 months", ym: "2023-03", months: 12, want: "2024-03"},
		{name: "month overflow", ym: "2023-12", months: 1, want: "2024-01"},
		{name: "subtract 1 month", ym: "2023-03", months: -1, want: "2023-02"},
		{name: "null", ym: YearMonthNull, months: 1, want: YearMonthNull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.AddMonths(tt.months)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_ContainsTime(t *testing.T) {
	ym := NullableYearMonth("2023-07")
	nullYm := NullableYearMonth(YearMonthNull)

	tests := []struct {
		name string
		ym   NullableYearMonth
		t    time.Time
		want bool
	}{
		{name: "first day", ym: ym, t: time.Date(2023, time.July, 1, 0, 0, 0, 0, time.Local), want: true},
		{name: "mid month", ym: ym, t: time.Date(2023, time.July, 15, 12, 30, 0, 0, time.Local), want: true},
		{name: "last day", ym: ym, t: time.Date(2023, time.July, 31, 23, 59, 59, 0, time.Local), want: true},
		{name: "before", ym: ym, t: time.Date(2023, time.June, 30, 23, 59, 59, 0, time.Local), want: false},
		{name: "after", ym: ym, t: time.Date(2023, time.August, 1, 0, 0, 0, 0, time.Local), want: false},
		{name: "null", ym: nullYm, t: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.Local), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.ContainsTime(tt.t)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_ContainsDate(t *testing.T) {
	ym := NullableYearMonth("2023-07")
	nullYm := NullableYearMonth(YearMonthNull)

	tests := []struct {
		name string
		ym   NullableYearMonth
		date Date
		want bool
	}{
		{name: "first day", ym: ym, date: "2023-07-01", want: true},
		{name: "mid month", ym: ym, date: "2023-07-15", want: true},
		{name: "last day", ym: ym, date: "2023-07-31", want: true},
		{name: "before", ym: ym, date: "2023-06-30", want: false},
		{name: "after", ym: ym, date: "2023-08-01", want: false},
		{name: "null", ym: nullYm, date: "2023-07-15", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.ContainsDate(tt.date)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearMonth_UnmarshalJSON(t *testing.T) {
	type doc struct {
		Period NullableYearMonth `json:"period"`
	}
	// JSON null must reset a field that holds a value, like for
	// uu.NullableID and email.NullableAddress, else decoding a partial
	// update with null into an existing struct keeps the old period.
	d := doc{Period: "2024-01"}
	require.NoError(t, json.Unmarshal([]byte(`{"period":null}`), &d))
	require.Equal(t, YearMonthNull, d.Period)

	// A JSON string is stored normalized if it is valid, a blank one as
	// null, and an invalid one unchanged without an error, so one bad
	// period doesn't fail a whole document; Validate reports it.
	for str, want := range map[string]NullableYearMonth{
		"2023-07":   "2023-07",
		" 2023-07 ": "2023-07",
		"2023-13":   "2023-13",
		"bad":       "bad",
		"":          YearMonthNull,
		"  ":        YearMonthNull,
	} {
		d := doc{Period: "2024-01"}
		data, err := json.Marshal(map[string]string{"period": str})
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, &d))
		require.Equal(t, want, d.Period, "json.Unmarshal(%s)", data)
	}

	// Any other JSON value is an error and leaves the value unchanged.
	for _, data := range []string{`202307`, `true`, `{}`, `["2023-07"]`} {
		ym := NullableYearMonth("2024-01")
		require.Error(t, json.Unmarshal([]byte(data), &ym), "json.Unmarshal(%s)", data)
		require.Equal(t, NullableYearMonth("2024-01"), ym, "json.Unmarshal(%s)", data)
	}
}

// A blank value is null, valid and normalizes to YearMonthNull,
// like NullableDate, but it is not normalized itself.
func TestNullableYearMonth_Normalized(t *testing.T) {
	for input, want := range map[NullableYearMonth]NullableYearMonth{
		"":          YearMonthNull,
		" \t":       YearMonthNull,
		"2023-07":   "2023-07",
		" 2023-07 ": "2023-07",
	} {
		got, err := input.Normalized()
		require.NoError(t, err, "%q", input)
		require.Equal(t, want, got, "%q", input)
		require.True(t, input.Valid(), "%q", input)
		require.Equal(t, input == want, input.ValidAndNormalized(), "%q", input)
	}
	got, err := NullableYearMonth("2023-13").Normalized()
	require.Error(t, err)
	require.Equal(t, NullableYearMonth("2023-13"), got, "unchanged")
}

// TestNullableYearMonth_ReadingAndWriting checks the reading and writing
// contract shared by the normalizing types, with a blank or zero
// NullableYearMonth as null like for NullableDate.
func TestNullableYearMonth_ReadingAndWriting(t *testing.T) {
	normtest.Run(t, normtest.Spec[NullableYearMonth]{
		Canonical:    "2023-07",
		Variants:     []NullableYearMonth{" 2023-07 ", "\t2023-07\n"},
		Invalid:      []NullableYearMonth{"2023-13", "2023-7", "0000-05", "07/2023", "bad"},
		Blank:        []NullableYearMonth{"0001-01", "0000-00", " 0001-01 "},
		Nullable:     true,
		StoredString: true,
	})
}

// A date or timestamp column scans as its year-month, like
// NullableDate.Scan scans it as its date, and the zero time as null.
func TestNullableYearMonth_ScanTime(t *testing.T) {
	var ym NullableYearMonth
	require.NoError(t, ym.Scan(time.Date(2023, 7, 15, 10, 0, 0, 0, time.UTC)))
	require.Equal(t, NullableYearMonth("2023-07"), ym)
	require.NoError(t, ym.Scan(time.Time{}))
	require.Equal(t, YearMonthNull, ym)
}
