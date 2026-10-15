package date

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/domonda/go-types/internal/normtest"
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
		{name: "lower case q", yq: "2023-q2", wantErr: false},
		{name: "surrounding whitespace", yq: " 2023-Q3\t", wantErr: false},
		{name: "empty", yq: "", wantErr: true},
		{name: "too short", yq: "2023-Q", wantErr: true},
		{name: "too long", yq: "2023-Q12", wantErr: true},
		{name: "no Q", yq: "2023-01", wantErr: true},
		{name: "invalid quarter 0", yq: "2023-Q0", wantErr: true},
		{name: "invalid quarter 5", yq: "2023-Q5", wantErr: true},
		{name: "invalid year", yq: "abcd-Q1", wantErr: true},
		// Every year of a Date is valid, so that the year-quarter of every
		// valid Date can be written, up to "9999-12-31"
		{name: "year after 3000", yq: "3001-Q1", wantErr: false},
		{name: "last year of Date", yq: "9999-Q4", wantErr: false},
		{name: "5 digit year", yq: "10000-Q1", wantErr: true},
		{name: "year 0", yq: "0000-Q2", wantErr: true},
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
		// Documented to return 0 for invalid values like YearMonth.Year,
		// a value with the right length but an invalid quarter or format
		// must not return a year that callers take as valid
		{name: "invalid quarter 5", yq: "2023-Q5", want: 0},
		{name: "invalid quarter 0", yq: "2023-Q0", want: 0},
		{name: "invalid format", yq: "2023/Q1", want: 0},
		{name: "year only", yq: "2023", want: 0},
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
		// Documented to return 0 for invalid values, a quarter outside
		// 1-4 must not be returned because FirstMonth, DateRange and
		// AddQuarters would compute months that don't exist from it
		{name: "invalid quarter 5", yq: "2023-Q5", want: 0},
		{name: "invalid quarter 9", yq: "2023-Q9", want: 0},
		{name: "invalid format", yq: "2023/Q1", want: 0},
		{name: "invalid year", yq: "abcd-Q1", want: 0},
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
		name           string
		yq             YearQuarter
		monthInQuarter int
		day            int
		want           Date
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
		// The last quarter of the years of Date: adding a month to
		// 9999-12-01 to find the last day gave "0000-12-31"
		{name: "Q4 9999", yq: "9999-Q4", wantFrom: "9999-10-01", wantUntil: "9999-12-31"},
		{name: "Q2 0001", yq: "0001-Q2", wantFrom: "0001-04-01", wantUntil: "0001-06-30"},
		// An invalid quarter must give an empty range,
		// not a range computed from non-existent months
		{name: "invalid quarter 5", yq: "2023-Q5", wantFrom: "", wantUntil: ""},
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
		{name: "to the last year", yq: "2023-Q2", years: 9999 - 2023, want: "9999-Q2"},
		// Past the years of Date the result is invalid, so that Value refuses it
		{name: "after the last year", yq: "9999-Q2", years: 1, want: "10000-Q2"},
		// An invalid year-quarter has no year to add to. It must stay
		// unchanged, so it stays invalid and Validate reports the original
		// value, instead of becoming a bogus year-quarter like "0001-Q0".
		{name: "invalid unchanged", yq: "bad", years: 1, want: "bad"},
		{name: "invalid quarter unchanged", yq: "2023-Q5", years: 1, want: "2023-Q5"},
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
		{name: "last quarter", yq: "9999-Q3", quarters: 1, want: "9999-Q4"},
		// Past the years of Date the result is invalid, so that Value refuses it
		{name: "after the last quarter", yq: "9999-Q4", quarters: 1, want: "10000-Q1"},
		{name: "before the year 1", yq: "0001-Q2", quarters: -2, want: "0000-Q4"},
		{name: "before the year 0", yq: "0001-Q4", quarters: -8, want: "-001-Q4"},
		// Invalid stays unchanged like in AddYears, not "0000-Q1"
		{name: "invalid unchanged", yq: "bad", quarters: 1, want: "bad"},
		{name: "invalid quarter unchanged", yq: "2023-Q5", quarters: 1, want: "2023-Q5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.AddQuarters(tt.quarters)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestYearQuarter_AddQuartersLikeTime protects the quarter arithmetic
// that replaced a loop per year: for every valid year-quarter of the
// years 1 to 9999 and up to two years back and forward, AddQuarters
// must give the quarter of time.Time arithmetic on its first month,
// also past the years 1 and 9999, where the result is invalid.
func TestYearQuarter_AddQuartersLikeTime(t *testing.T) {
	for year := 1; year <= 9999; year++ {
		for quarter := 1; quarter <= 4; quarter++ {
			yq := YearQuarterFrom(year, quarter)
			if yq.IsZero() {
				continue // "0001-Q1" is not valid and returned unchanged
			}
			for n := -9; n <= 9; n++ {
				tm := time.Date(year, time.Month(3*(quarter-1)+1+3*n), 1, 0, 0, 0, 0, time.UTC)
				want := YearQuarterFrom(tm.Year(), (int(tm.Month())-1)/3+1)
				if got := yq.AddQuarters(n); got != want {
					t.Fatalf("%s.AddQuarters(%d) = %s, want %s", yq, n, got, want)
				}
			}
		}
	}
}

// TestYearQuarter_AddQuartersExtremes protects against a denial of
// service from a quarter count read from input: AddQuarters looped once
// per year, which took 400ms for 1<<32 quarters and never returned for
// math.MaxInt, whose sum with the quarter overflowed. The result must
// come at once, match exact big.Int arithmetic and be invalid, so that
// Value refuses it instead of writing a wrapped around year-quarter.
func TestYearQuarter_AddQuartersExtremes(t *testing.T) {
	type call struct {
		yq       YearQuarter
		quarters int
	}
	var calls []call
	for _, yq := range []YearQuarter{"0001-Q2", "2023-Q2", "9999-Q4"} {
		// Every remainder of the division by 4 at both ends of int
		for i := range 4 {
			calls = append(calls, call{yq, math.MaxInt - i}, call{yq, math.MinInt + i})
		}
	}

	results := make(chan []YearQuarter, 1)
	go func() {
		var got []YearQuarter
		for _, c := range calls {
			got = append(got, c.yq.AddQuarters(c.quarters))
		}
		results <- got
	}()
	var got []YearQuarter
	select {
	case got = <-results:
	case <-time.After(10 * time.Second):
		t.Fatal("AddQuarters didn't return for math.MaxInt or math.MinInt")
	}

	for i, c := range calls {
		// Quarters since Q1 of the year 0, floored by the Euclidean DivMod
		total := big.NewInt(int64(c.yq.Year())*4 + int64(c.yq.Quarter()) - 1)
		total.Add(total, big.NewInt(int64(c.quarters)))
		year, quarter := new(big.Int).DivMod(total, big.NewInt(4), new(big.Int))
		want := YearQuarter(fmt.Sprintf("%04d-Q%d", year, quarter.Int64()+1))
		assert.Equal(t, want, got[i], "%s.AddQuarters(%d)", c.yq, c.quarters)
		_, err := got[i].Value()
		assert.Error(t, err, "Value of %s.AddQuarters(%d)", c.yq, c.quarters)
		assert.Equal(t, got[i].Nullable(), c.yq.Nullable().AddQuarters(c.quarters), "NullableYearQuarter")
	}
}

// TestYearQuarter_ValueOfAValidDate protects writing the year-quarter of
// every valid Date: YearQuarter accepted only the years up to 3000, so
// Value failed for the year-quarter of a valid Date like "9999-12-31".
func TestYearQuarter_ValueOfAValidDate(t *testing.T) {
	for d, want := range map[Date]string{"0001-04-01": "0001-Q2", "3001-01-01": "3001-Q1", "9999-12-31": "9999-Q4"} {
		yq := YearQuarterOfTime(d.Midnight())
		v, err := yq.Value()
		assert.NoError(t, err, "YearQuarter(%q).Value()", yq)
		assert.Equal(t, want, v)
		assert.True(t, yq.ContainsDate(d), "ContainsDate(%q)", d)
	}

	yq := YearQuarterFrom(9999, 4)
	assert.Equal(t, YearQuarter("9999-Q4"), yq)
	assert.Equal(t, 9999, yq.Year())
	assert.Equal(t, 4, yq.Quarter())
	assert.True(t, yq.ContainsYearMonth("9999-12"))
	v, err := yq.Nullable().Value()
	assert.NoError(t, err, "NullableYearQuarter.Value")
	assert.Equal(t, "9999-Q4", v)

	// A year-quarter after the years of Date is still refused
	_, err = YearQuarterFrom(10000, 1).Value()
	assert.Error(t, err)
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
		{name: "equal when normalized", yq: " 2023-q3", other: "2023-Q3", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Compare(tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestYearQuarter_CompareFastPath protects the order of Compare: it skips
// normalizing a year-quarter that isNormalized reports as normalized, so
// isNormalized must agree with ValidAndNormalized for every year-quarter,
// or Compare would order an invalid, zero, lower case or padded one as
// it is instead of like Normalized and IsZero do.
func TestYearQuarter_CompareFastPath(t *testing.T) {
	const chars = "0123456789-Qq :"
	for _, year := range []string{"0000", "0001", "2023", "9999", "20a3", " 202", "202 "} {
		for _, q0 := range chars {
			for _, q1 := range chars {
				for _, sep := range []string{"-", "/"} {
					yq := YearQuarter(year + sep + string(q0) + string(q1))
					assert.Equal(t, yq.ValidAndNormalized(), yq.isNormalized(), "%q", yq)
				}
			}
		}
	}
	for _, yq := range []YearQuarter{"", "2023-Q", "10000-Q1", " 2023-Q3", "2023-Q3 "} {
		assert.False(t, yq.isNormalized(), "%q", yq)
	}

	// The fast path and the normalizing path order alike
	assert.Equal(t, -1, YearQuarter("2023-Q3").Compare("2023-Q4"))
	assert.Equal(t, 0, YearQuarter("2023-Q3").Compare("2023-q3"))
	assert.Equal(t, -1, YearQuarter("0001-Q1").Compare("0001-Q2"), "the zero year-quarter compares as \"\"")
	assert.Equal(t, 1, YearQuarter("2023-Q5").Compare("2023-Q4"), "an invalid year-quarter compares as it is")
}

// BenchmarkYearQuarter_Compare measures Compare as a comparator of
// slices.SortFunc, which calls it about n*log(n) times, mostly with
// year-quarters that were read normalized from SQL or JSON.
func BenchmarkYearQuarter_Compare(b *testing.B) {
	for name, operands := range map[string][2]YearQuarter{
		"normalized":     {"2023-Q3", "2023-Q4"},
		"not normalized": {"2023-q3", "2023-Q4"},
		"invalid":        {"2023-Q5", "2023-Q4"},
	} {
		b.Run(name, func(b *testing.B) {
			for b.Loop() {
				_ = operands[0].Compare(operands[1])
			}
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
		// "0001-Q1" is the year-quarter of the zero time.Time, which IsNull
		// reports as null like the zero dates of NullableDate.
		{name: "zero 0001-Q1", yq: "0001-Q1", want: false},
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
		{name: "whitespace only", yq: " \t", want: true},
		// The zero values are null like the zero dates of NullableDate
		{name: "zero 0000-Q0", yq: "0000-Q0", want: true},
		{name: "zero 0001-Q1", yq: "0001-Q1", want: true},
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
		name     string
		yq       NullableYearQuarter
		default_ YearQuarter
		want     YearQuarter
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
	// Like NullableYearMonth.Date, the date of a nullable period is a
	// NullableDate, so a null period gives a null date that is written
	// as SQL and JSON null instead of an empty Date.
	tests := []struct {
		name           string
		yq             NullableYearQuarter
		monthInQuarter int
		day            int
		want           NullableDate
	}{
		{name: "first day", yq: "2023-Q2", monthInQuarter: 1, day: 1, want: "2023-04-01"},
		{name: "mid quarter", yq: "2023-Q2", monthInQuarter: 2, day: 15, want: "2023-05-15"},
		{name: "null", yq: YearQuarterNull, monthInQuarter: 1, day: 15, want: Null},
		{name: "invalid", yq: "2023-Q5", monthInQuarter: 1, day: 15, want: Null},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.yq.Date(tt.monthInQuarter, tt.day)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableYearQuarter_DateRange(t *testing.T) {
	// Like NullableYearMonth.DateRange, the range of a nullable period is
	// a pair of NullableDate, so a null period gives null bounds.
	tests := []struct {
		name      string
		yq        NullableYearQuarter
		wantFrom  NullableDate
		wantUntil NullableDate
	}{
		{name: "Q2 2023", yq: "2023-Q2", wantFrom: "2023-04-01", wantUntil: "2023-06-30"},
		{name: "null", yq: YearQuarterNull, wantFrom: Null, wantUntil: Null},
		{name: "invalid", yq: "2023-Q5", wantFrom: Null, wantUntil: Null},
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
		// Past the years of Date the result is invalid like for YearQuarter
		{name: "after the last quarter", yq: "9999-Q4", quarters: 1, want: "10000-Q1"},
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

// TestYearQuarter_ValidateRejectsYearZero protects that a YearQuarter is only
// valid if its dates are, see TestYearMonth_ValidateRejectsYearZero.
func TestYearQuarter_ValidateRejectsYearZero(t *testing.T) {
	assert.Error(t, YearQuarter("0000-Q2").Validate())
	assert.NoError(t, YearQuarter("0001-Q2").Validate())
}

// Normalized trims the surrounding whitespace and upper-cases a "q",
// which PeriodRange accepts too, and returns an invalid value unchanged
// with an error, like the other normalizing types.
func TestYearQuarter_Normalized(t *testing.T) {
	for input, want := range map[YearQuarter]YearQuarter{
		"2023-Q3":   "2023-Q3",
		"2023-q3":   "2023-Q3",
		" 2023-q3 ": "2023-Q3",
	} {
		got, err := input.Normalized()
		assert.NoError(t, err, "%q", input)
		assert.Equal(t, want, got, "%q", input)
		assert.Equal(t, input == want, input.ValidAndNormalized(), "%q", input)
	}
	for _, input := range []YearQuarter{"", "  ", "2023-Q5", "2023-q0", "0000-Q1", "2023-X1", "Q3/2023"} {
		got, err := input.Normalized()
		assert.Error(t, err, "%q", input)
		assert.Equal(t, input, got, "%q unchanged", input)
		assert.False(t, input.ValidAndNormalized(), "%q", input)
	}
}

// The accessors work on the normalized year-quarter,
// because Valid accepts the spellings that Normalized accepts.
func TestYearQuarter_AccessorsNormalize(t *testing.T) {
	yq := YearQuarter(" 2023-q3 ")
	assert.Equal(t, 2023, yq.Year())
	assert.Equal(t, 3, yq.Quarter())
	from, until := yq.DateRange()
	assert.Equal(t, Date("2023-07-01"), from)
	assert.Equal(t, Date("2023-09-30"), until)
	assert.Equal(t, YearQuarter("2023-Q4"), yq.AddQuarters(1))
	assert.True(t, yq.ContainsYearMonth(" 2023-08"))
	assert.True(t, YearQuarter("0001-q1").IsZero())
}

// Reading JSON normalizes like Date.UnmarshalJSON, so a value read from
// JSON is canonical if it is valid and compares with ==. An invalid value
// is kept without an error, so one bad period doesn't fail a whole
// document, and Validate reports it.
func TestYearQuarter_UnmarshalJSON(t *testing.T) {
	for data, want := range map[string]YearQuarter{
		`"2023-Q3"`:   "2023-Q3",
		`" 2023-q3 "`: "2023-Q3",
		`"2023-Q5"`:   "2023-Q5",
		`"bad"`:       "bad",
		`""`:          "",
		`"  "`:        "",
	} {
		yq := YearQuarter("2024-Q1")
		assert.NoError(t, json.Unmarshal([]byte(data), &yq), "json.Unmarshal(%s)", data)
		assert.Equal(t, want, yq, "json.Unmarshal(%s)", data)
	}

	// JSON null leaves a non-nullable value unchanged like encoding/json
	// does for a string type, any other JSON value is an error
	yq := YearQuarter("2024-Q1")
	assert.NoError(t, json.Unmarshal([]byte(`null`), &yq))
	assert.Equal(t, YearQuarter("2024-Q1"), yq)
	assert.Error(t, json.Unmarshal([]byte(`20233`), &yq))
	assert.Equal(t, YearQuarter("2024-Q1"), yq)
}

// TestYearQuarter_ReadingAndWriting checks the reading and writing contract
// shared by the normalizing types, like for Date: every reading path stores
// a valid year-quarter normalized, a zero one as "" and an invalid one
// unchanged, Value refuses an invalid year-quarter, and fmt prints it as stored.
func TestYearQuarter_ReadingAndWriting(t *testing.T) {
	normtest.Run(t, normtest.Spec[YearQuarter]{
		Canonical:    "2023-Q3",
		Variants:     []YearQuarter{"2023-q3", " 2023-Q3 ", "\t2023-q3\n"},
		Invalid:      []YearQuarter{"2023-Q5", "2023-Q", "0000-Q2", "Q3/2023", "bad"},
		Blank:        []YearQuarter{"0001-Q1", "0001-q1", "0000-Q0", " 0001-Q1 "},
		StoredString: true,
	})
}

// A date or timestamp column scans as its year-quarter, like Date.Scan
// scans it as its date, and the zero time as the zero value "".
func TestYearQuarter_ScanTime(t *testing.T) {
	var yq YearQuarter
	assert.NoError(t, yq.Scan(time.Date(2023, 8, 15, 10, 0, 0, 0, time.UTC)))
	assert.Equal(t, YearQuarter("2023-Q3"), yq)
	assert.NoError(t, yq.Scan(time.Time{}))
	assert.Equal(t, YearQuarter(""), yq)
}
