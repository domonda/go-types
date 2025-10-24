package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Constructor tests

func TestYearQuarterFrom(t *testing.T) {
	tests := []struct {
		name    string
		year    int
		quarter int
		want    YearQuarter
	}{
		{name: "Q1", year: 2023, quarter: 1, want: "2023-Q1"},
		{name: "Q2", year: 2023, quarter: 2, want: "2023-Q2"},
		{name: "Q3", year: 2023, quarter: 3, want: "2023-Q3"},
		{name: "Q4", year: 2023, quarter: 4, want: "2023-Q4"},
		{name: "year padding", year: 999, quarter: 2, want: "0999-Q2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := YearQuarterFrom(tt.year, tt.quarter)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarterOfTime(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want YearQuarter
	}{
		{name: "Q1 January", t: time.Date(2023, time.January, 15, 12, 30, 0, 0, time.UTC), want: "2023-Q1"},
		{name: "Q1 March", t: time.Date(2023, time.March, 31, 0, 0, 0, 0, time.UTC), want: "2023-Q1"},
		{name: "Q2 April", t: time.Date(2023, time.April, 1, 0, 0, 0, 0, time.UTC), want: "2023-Q2"},
		{name: "Q3 July", t: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC), want: "2023-Q3"},
		{name: "Q4 December", t: time.Date(2023, time.December, 31, 23, 59, 59, 0, time.UTC), want: "2023-Q4"},
		{name: "zero time", t: time.Time{}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := YearQuarterOfTime(tt.t)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarterOfToday(t *testing.T) {
	result := YearQuarterOfToday()
	assert.True(t, result.Valid())
	assert.False(t, result.IsZero())

	// Verify it matches current time
	now := time.Now()
	expected := YearQuarterOfTime(now)
	assert.Equal(t, expected, result)
}

// Validation tests

func TestYearQuarter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		yq      YearQuarter
		wantErr bool
	}{
		{name: "valid Q1", yq: "2023-Q1", wantErr: false},
		{name: "valid Q4", yq: "2023-Q4", wantErr: false},
		{name: "empty", yq: "", wantErr: true},
		{name: "too short", yq: "2023-Q", wantErr: true},
		{name: "too long", yq: "2023-Q12", wantErr: true},
		{name: "no Q", yq: "2023-01", wantErr: true},
		{name: "invalid quarter 0", yq: "2023-Q0", wantErr: true},
		{name: "invalid quarter 5", yq: "2023-Q5", wantErr: true},
		{name: "invalid year", yq: "abcd-Q1", wantErr: true},
		{name: "year too large", yq: "3001-Q1", wantErr: true},
		{name: "invalid format", yq: "2023/Q1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.yq.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestYearQuarter_Valid(t *testing.T) {
	tests := []struct {
		name string
		yq   YearQuarter
		want bool
	}{
		{name: "valid", yq: "2023-Q1", want: true},
		{name: "invalid quarter", yq: "2023-Q5", want: false},
		{name: "empty", yq: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Valid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarter_IsZero(t *testing.T) {
	tests := []struct {
		name string
		yq   YearQuarter
		want bool
	}{
		{name: "empty string", yq: "", want: true},
		{name: "0000-Q0", yq: "0000-Q0", want: true},
		{name: "0001-Q1", yq: "0001-Q1", want: true},
		{name: "valid quarter", yq: "2023-Q2", want: false},
		{name: "0001-Q2", yq: "0001-Q2", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.IsZero()
			assert.Equal(t, tt.want, got)
		})
	}
}

// String and conversion tests

func TestYearQuarter_String(t *testing.T) {
	yq := YearQuarter("2023-Q3")
	assert.Equal(t, "2023-Q3", yq.String())
}

func TestYearQuarter_Year(t *testing.T) {
	tests := []struct {
		name string
		yq   YearQuarter
		want int
	}{
		{name: "regular year", yq: "2023-Q3", want: 2023},
		{name: "different year", yq: "1999-Q4", want: 1999},
		{name: "zero year", yq: "0000-Q0", want: 0},
		{name: "invalid", yq: "invalid", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Year()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarter_Quarter(t *testing.T) {
	tests := []struct {
		name string
		yq   YearQuarter
		want int
	}{
		{name: "Q1", yq: "2023-Q1", want: 1},
		{name: "Q2", yq: "2023-Q2", want: 2},
		{name: "Q3", yq: "2023-Q3", want: 3},
		{name: "Q4", yq: "2023-Q4", want: 4},
		{name: "invalid", yq: "invalid", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Quarter()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarter_FirstMonth(t *testing.T) {
	tests := []struct {
		name string
		yq   YearQuarter
		want time.Month
	}{
		{name: "Q1", yq: "2023-Q1", want: time.January},
		{name: "Q2", yq: "2023-Q2", want: time.April},
		{name: "Q3", yq: "2023-Q3", want: time.July},
		{name: "Q4", yq: "2023-Q4", want: time.October},
		{name: "invalid", yq: "invalid", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.FirstMonth()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarter_Date(t *testing.T) {
	tests := []struct {
		name            string
		yq              YearQuarter
		monthInQuarter int
		day             int
		want            Date
	}{
		{name: "Q1 first month", yq: "2023-Q1", monthInQuarter: 1, day: 1, want: "2023-01-01"},
		{name: "Q1 second month", yq: "2023-Q1", monthInQuarter: 2, day: 15, want: "2023-02-15"},
		{name: "Q2 third month", yq: "2023-Q2", monthInQuarter: 3, day: 30, want: "2023-06-30"},
		{name: "Q4 first month", yq: "2023-Q4", monthInQuarter: 1, day: 1, want: "2023-10-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Date(tt.monthInQuarter, tt.day)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarter_DateRange(t *testing.T) {
	tests := []struct {
		name      string
		yq        YearQuarter
		wantFrom  Date
		wantUntil Date
	}{
		{name: "Q1 2023", yq: "2023-Q1", wantFrom: "2023-01-01", wantUntil: "2023-03-31"},
		{name: "Q2 2023", yq: "2023-Q2", wantFrom: "2023-04-01", wantUntil: "2023-06-30"},
		{name: "Q3 2023", yq: "2023-Q3", wantFrom: "2023-07-01", wantUntil: "2023-09-30"},
		{name: "Q4 2023", yq: "2023-Q4", wantFrom: "2023-10-01", wantUntil: "2023-12-31"},
		{name: "Q1 2020 (leap)", yq: "2020-Q1", wantFrom: "2020-01-01", wantUntil: "2020-03-31"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrom, gotUntil := tt.yq.DateRange()
			assert.Equal(t, tt.wantFrom, gotFrom)
			assert.Equal(t, tt.wantUntil, gotUntil)
		})
	}
}

func TestYearQuarter_Nullable(t *testing.T) {
	yq := YearQuarter("2023-Q3")
	nullable := yq.Nullable()
	assert.Equal(t, NullableYearQuarter("2023-Q3"), nullable)
	assert.False(t, nullable.IsNull())
}

// Arithmetic tests

func TestYearQuarter_AddYears(t *testing.T) {
	tests := []struct {
		name  string
		yq    YearQuarter
		years int
		want  YearQuarter
	}{
		{name: "add 1 year", yq: "2023-Q2", years: 1, want: "2024-Q2"},
		{name: "add 5 years", yq: "2020-Q3", years: 5, want: "2025-Q3"},
		{name: "subtract 1 year", yq: "2023-Q2", years: -1, want: "2022-Q2"},
		{name: "add 0 years", yq: "2023-Q2", years: 0, want: "2023-Q2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.AddYears(tt.years)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarter_AddQuarters(t *testing.T) {
	tests := []struct {
		name     string
		yq       YearQuarter
		quarters int
		want     YearQuarter
	}{
		{name: "add 1 quarter", yq: "2023-Q2", quarters: 1, want: "2023-Q3"},
		{name: "add 4 quarters", yq: "2023-Q2", quarters: 4, want: "2024-Q2"},
		{name: "quarter overflow", yq: "2023-Q4", quarters: 1, want: "2024-Q1"},
		{name: "subtract 1 quarter", yq: "2023-Q2", quarters: -1, want: "2023-Q1"},
		{name: "quarter underflow", yq: "2023-Q1", quarters: -1, want: "2022-Q4"},
		{name: "add 0 quarters", yq: "2023-Q2", quarters: 0, want: "2023-Q2"},
		{name: "add 2 years worth", yq: "2023-Q2", quarters: 8, want: "2025-Q2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.AddQuarters(tt.quarters)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Contains tests

func TestYearQuarter_ContainsTime(t *testing.T) {
	yq := YearQuarter("2023-Q3")

	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{name: "first day", t: time.Date(2023, time.July, 1, 0, 0, 0, 0, time.Local), want: true},
		{name: "mid quarter", t: time.Date(2023, time.August, 15, 12, 30, 0, 0, time.Local), want: true},
		{name: "last day", t: time.Date(2023, time.September, 30, 23, 59, 59, 0, time.Local), want: true},
		{name: "before", t: time.Date(2023, time.June, 30, 23, 59, 59, 0, time.Local), want: false},
		{name: "after", t: time.Date(2023, time.October, 1, 0, 0, 0, 0, time.Local), want: false},
		{name: "different year", t: time.Date(2022, time.August, 15, 0, 0, 0, 0, time.Local), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yq.ContainsTime(tt.t)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarter_ContainsDate(t *testing.T) {
	yq := YearQuarter("2023-Q3")

	tests := []struct {
		name string
		date Date
		want bool
	}{
		{name: "first day", date: "2023-07-01", want: true},
		{name: "mid quarter", date: "2023-08-15", want: true},
		{name: "last day", date: "2023-09-30", want: true},
		{name: "before", date: "2023-06-30", want: false},
		{name: "after", date: "2023-10-01", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yq.ContainsDate(tt.date)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarter_ContainsYearMonth(t *testing.T) {
	yq := YearQuarter("2023-Q3")

	tests := []struct {
		name string
		ym   YearMonth
		want bool
	}{
		{name: "first month", ym: "2023-07", want: true},
		{name: "second month", ym: "2023-08", want: true},
		{name: "third month", ym: "2023-09", want: true},
		{name: "before", ym: "2023-06", want: false},
		{name: "after", ym: "2023-10", want: false},
		{name: "different year", ym: "2022-07", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yq.ContainsYearMonth(tt.ym)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestYearQuarter_Compare(t *testing.T) {
	tests := []struct {
		name  string
		yq    YearQuarter
		other YearQuarter
		want  int
	}{
		{name: "before", yq: "2023-Q1", other: "2023-Q3", want: -1},
		{name: "after", yq: "2023-Q3", other: "2023-Q1", want: 1},
		{name: "equal", yq: "2023-Q3", other: "2023-Q3", want: 0},
		{name: "different years", yq: "2022-Q4", other: "2023-Q1", want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Compare(tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

// NullableYearQuarter tests

func TestNullableYearQuarter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		yq      NullableYearQuarter
		wantErr bool
	}{
		{name: "valid", yq: "2023-Q2", wantErr: false},
		{name: "null", yq: YearQuarterNull, wantErr: false},
		{name: "invalid", yq: "2023-Q5", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.yq.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNullableYearQuarter_Valid(t *testing.T) {
	tests := []struct {
		name string
		yq   NullableYearQuarter
		want bool
	}{
		{name: "valid", yq: "2023-Q2", want: true},
		{name: "null", yq: YearQuarterNull, want: true},
		{name: "invalid", yq: "2023-Q5", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Valid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_ValidAndNotNull(t *testing.T) {
	tests := []struct {
		name string
		yq   NullableYearQuarter
		want bool
	}{
		{name: "valid", yq: "2023-Q2", want: true},
		{name: "null", yq: YearQuarterNull, want: false},
		{name: "invalid", yq: "2023-Q5", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.ValidAndNotNull()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_IsZero(t *testing.T) {
	tests := []struct {
		name string
		yq   NullableYearQuarter
		want bool
	}{
		{name: "empty", yq: "", want: true},
		{name: "0000-Q0", yq: "0000-Q0", want: true},
		{name: "0001-Q1", yq: "0001-Q1", want: true},
		{name: "valid", yq: "2023-Q2", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.IsZero()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_IsNull(t *testing.T) {
	tests := []struct {
		name string
		yq   NullableYearQuarter
		want bool
	}{
		{name: "null", yq: YearQuarterNull, want: true},
		{name: "empty", yq: "", want: true},
		{name: "valid", yq: "2023-Q2", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.IsNull()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_IsNotNull(t *testing.T) {
	tests := []struct {
		name string
		yq   NullableYearQuarter
		want bool
	}{
		{name: "null", yq: YearQuarterNull, want: false},
		{name: "valid", yq: "2023-Q2", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.IsNotNull()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_Get(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		yq := NullableYearQuarter("2023-Q2")
		got := yq.Get()
		assert.Equal(t, YearQuarter("2023-Q2"), got)
	})

	t.Run("panic on null", func(t *testing.T) {
		yq := NullableYearQuarter(YearQuarterNull)
		assert.Panics(t, func() {
			yq.Get()
		})
	})
}

func TestNullableYearQuarter_GetOr(t *testing.T) {
	tests := []struct {
		name    string
		yq      NullableYearQuarter
		default_ YearQuarter
		want    YearQuarter
	}{
		{name: "valid value", yq: "2023-Q2", default_: "2020-Q1", want: "2023-Q2"},
		{name: "null value", yq: YearQuarterNull, default_: "2020-Q1", want: "2020-Q1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.GetOr(tt.default_)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_Set(t *testing.T) {
	var yq NullableYearQuarter
	yq.Set(YearQuarter("2023-Q3"))
	assert.Equal(t, NullableYearQuarter("2023-Q3"), yq)
	assert.False(t, yq.IsNull())
}

func TestNullableYearQuarter_SetNull(t *testing.T) {
	yq := NullableYearQuarter("2023-Q3")
	yq.SetNull()
	assert.Equal(t, NullableYearQuarter(YearQuarterNull), yq)
	assert.True(t, yq.IsNull())
}

func TestNullableYearQuarter_String(t *testing.T) {
	tests := []struct {
		name string
		yq   NullableYearQuarter
		want string
	}{
		{name: "valid", yq: "2023-Q3", want: "2023-Q3"},
		{name: "null", yq: YearQuarterNull, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_StringOr(t *testing.T) {
	tests := []struct {
		name       string
		yq         NullableYearQuarter
		nullString string
		want       string
	}{
		{name: "valid", yq: "2023-Q3", nullString: "NULL", want: "2023-Q3"},
		{name: "null", yq: YearQuarterNull, nullString: "NULL", want: "NULL"},
		{name: "null with dash", yq: YearQuarterNull, nullString: "-", want: "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.StringOr(tt.nullString)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_Compare(t *testing.T) {
	tests := []struct {
		name  string
		yq    NullableYearQuarter
		other NullableYearQuarter
		want  int
	}{
		{name: "before", yq: "2023-Q1", other: "2023-Q3", want: -1},
		{name: "after", yq: "2023-Q3", other: "2023-Q1", want: 1},
		{name: "equal", yq: "2023-Q3", other: "2023-Q3", want: 0},
		{name: "null vs value", yq: YearQuarterNull, other: "2023-Q2", want: -1},
		{name: "value vs null", yq: "2023-Q2", other: YearQuarterNull, want: 1},
		{name: "both null", yq: YearQuarterNull, other: YearQuarterNull, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Compare(tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_Year(t *testing.T) {
	tests := []struct {
		name string
		yq   NullableYearQuarter
		want int
	}{
		{name: "valid", yq: "2023-Q3", want: 2023},
		{name: "different year", yq: "1999-Q4", want: 1999},
		{name: "null", yq: YearQuarterNull, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Year()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_Quarter(t *testing.T) {
	tests := []struct {
		name string
		yq   NullableYearQuarter
		want int
	}{
		{name: "Q1", yq: "2023-Q1", want: 1},
		{name: "Q4", yq: "2023-Q4", want: 4},
		{name: "null", yq: YearQuarterNull, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Quarter()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_FirstMonth(t *testing.T) {
	tests := []struct {
		name string
		yq   NullableYearQuarter
		want time.Month
	}{
		{name: "Q1", yq: "2023-Q1", want: time.January},
		{name: "Q2", yq: "2023-Q2", want: time.April},
		{name: "null", yq: YearQuarterNull, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.FirstMonth()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_Date(t *testing.T) {
	tests := []struct {
		name            string
		yq              NullableYearQuarter
		monthInQuarter int
		day             int
		want            Date
	}{
		{name: "first day", yq: "2023-Q2", monthInQuarter: 1, day: 1, want: "2023-04-01"},
		{name: "mid quarter", yq: "2023-Q2", monthInQuarter: 2, day: 15, want: "2023-05-15"},
		{name: "null", yq: YearQuarterNull, monthInQuarter: 1, day: 15, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Date(tt.monthInQuarter, tt.day)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_DateRange(t *testing.T) {
	tests := []struct {
		name      string
		yq        NullableYearQuarter
		wantFrom  Date
		wantUntil Date
	}{
		{name: "Q2 2023", yq: "2023-Q2", wantFrom: "2023-04-01", wantUntil: "2023-06-30"},
		{name: "null", yq: YearQuarterNull, wantFrom: "", wantUntil: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrom, gotUntil := tt.yq.DateRange()
			assert.Equal(t, tt.wantFrom, gotFrom)
			assert.Equal(t, tt.wantUntil, gotUntil)
		})
	}
}

func TestNullableYearQuarter_AddYears(t *testing.T) {
	tests := []struct {
		name  string
		yq    NullableYearQuarter
		years int
		want  NullableYearQuarter
	}{
		{name: "add 1 year", yq: "2023-Q2", years: 1, want: "2024-Q2"},
		{name: "add 5 years", yq: "2020-Q3", years: 5, want: "2025-Q3"},
		{name: "subtract 1 year", yq: "2023-Q2", years: -1, want: "2022-Q2"},
		{name: "null", yq: YearQuarterNull, years: 1, want: YearQuarterNull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.AddYears(tt.years)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_AddQuarters(t *testing.T) {
	tests := []struct {
		name     string
		yq       NullableYearQuarter
		quarters int
		want     NullableYearQuarter
	}{
		{name: "add 1 quarter", yq: "2023-Q2", quarters: 1, want: "2023-Q3"},
		{name: "add 4 quarters", yq: "2023-Q2", quarters: 4, want: "2024-Q2"},
		{name: "quarter overflow", yq: "2023-Q4", quarters: 1, want: "2024-Q1"},
		{name: "subtract 1 quarter", yq: "2023-Q2", quarters: -1, want: "2023-Q1"},
		{name: "null", yq: YearQuarterNull, quarters: 1, want: YearQuarterNull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.AddQuarters(tt.quarters)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_ContainsTime(t *testing.T) {
	yq := NullableYearQuarter("2023-Q3")
	nullYq := NullableYearQuarter(YearQuarterNull)

	tests := []struct {
		name string
		yq   NullableYearQuarter
		t    time.Time
		want bool
	}{
		{name: "first day", yq: yq, t: time.Date(2023, time.July, 1, 0, 0, 0, 0, time.Local), want: true},
		{name: "mid quarter", yq: yq, t: time.Date(2023, time.August, 15, 12, 30, 0, 0, time.Local), want: true},
		{name: "last day", yq: yq, t: time.Date(2023, time.September, 30, 23, 59, 59, 0, time.Local), want: true},
		{name: "before", yq: yq, t: time.Date(2023, time.June, 30, 23, 59, 59, 0, time.Local), want: false},
		{name: "after", yq: yq, t: time.Date(2023, time.October, 1, 0, 0, 0, 0, time.Local), want: false},
		{name: "null", yq: nullYq, t: time.Date(2023, time.August, 15, 0, 0, 0, 0, time.Local), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.ContainsTime(tt.t)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_ContainsDate(t *testing.T) {
	yq := NullableYearQuarter("2023-Q3")
	nullYq := NullableYearQuarter(YearQuarterNull)

	tests := []struct {
		name string
		yq   NullableYearQuarter
		date Date
		want bool
	}{
		{name: "first day", yq: yq, date: "2023-07-01", want: true},
		{name: "mid quarter", yq: yq, date: "2023-08-15", want: true},
		{name: "last day", yq: yq, date: "2023-09-30", want: true},
		{name: "before", yq: yq, date: "2023-06-30", want: false},
		{name: "after", yq: yq, date: "2023-10-01", want: false},
		{name: "null", yq: nullYq, date: "2023-08-15", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.ContainsDate(tt.date)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_ContainsYearMonth(t *testing.T) {
	yq := NullableYearQuarter("2023-Q3")
	nullYq := NullableYearQuarter(YearQuarterNull)

	tests := []struct {
		name string
		yq   NullableYearQuarter
		ym   YearMonth
		want bool
	}{
		{name: "first month", yq: yq, ym: "2023-07", want: true},
		{name: "second month", yq: yq, ym: "2023-08", want: true},
		{name: "third month", yq: yq, ym: "2023-09", want: true},
		{name: "before", yq: yq, ym: "2023-06", want: false},
		{name: "after", yq: yq, ym: "2023-10", want: false},
		{name: "null", yq: nullYq, ym: "2023-07", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.ContainsYearMonth(tt.ym)
			assert.Equal(t, tt.want, got)
		})
	}
}
