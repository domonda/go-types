package date

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/language"
)

func Test_Normalize(t *testing.T) {
	dateTable := map[string]Date{
		"2006-01-02": "2006-01-02",
		"2006/01/02": "2006-01-02",
		"2006.01.02": "2006-01-02",
		"2006 01 02": "2006-01-02",

		"25.12.1975": "1975-12-25",
		"25.12.75":   "1975-12-25",

		"12/25/1975": "1975-12-25",
		"12/25/75":   "1975-12-25",

		"01.02.03": "2003-02-01",
		// "1.2.03": "2003-02-01", // too short
		"1.2.2003": "2003-02-01",

		"1st of 02/2003": "2003-02-01",
		"4th of 02/2003": "2003-02-04",

		"jan. 24 2012":    "2012-01-24",
		"Februar 24 89":   "1989-02-24",
		"February 3rd 89": "1989-02-03",

		"1. Dezember 2016": "2016-12-01",
		"1st of dec. 2020": "2020-12-01",

		"2016 Dezember 9th": "2016-12-09",
		"16 Dezember 9th":   "2016-12-09",
		"16. Dezember 98":   "1998-12-16",
		"16th of Dec. 04":   "2004-12-16",

		"2016 25. März":   "2016-03-25",
		"75 1st of march": "1975-03-01",

		// French, Italian, Spanish month abbreviations.
		"2 janv. 2019":  "2019-01-02",
		"30 janv. 2019": "2019-01-30",
		"23/gen/2019":   "2019-01-23",
		// French also abbreviates juillet as juill. and février as fév.
		"5 juill. 2021": "2021-07-05",
		"15 fév. 2020":  "2020-02-15",
		"15 fev 2020":   "2020-02-15",

		// US "Month Day, Year" — comma is now a separator and
		// abbreviated months may end in '.' (also a separator).
		// Test data adapted from github.com/araddon/dateparse.
		"oct 7, 1970":   "1970-10-07",
		"oct 7, '70":    "1970-10-07",
		"Oct 7, '70":    "1970-10-07",
		"Oct. 7, '70":   "1970-10-07",
		"oct. 7, '70":   "1970-10-07",
		"oct. 7, 1970":  "1970-10-07",
		"Sept. 7, '70":  "1970-09-07",
		"sept. 7, 1970": "1970-09-07",
		"Feb 8, 2009":   "2009-02-08",

		"7 oct 70":         "1970-10-07",
		"7 oct 1970":       "1970-10-07",
		"7 May 1970":       "1970-05-07",
		"7 Sep 1970":       "1970-09-07",
		"7 June 1970":      "1970-06-07",
		"7 September 1970": "1970-09-07",

		// "august" ends in the ordinal suffix "st", which must only be
		// stripped after digits, else the month name becomes "augu".
		"17 August 2012":   "2012-08-17",
		"August 17, 2012":  "2012-08-17",
		"1. August 2016":   "2016-08-01",
		"1st August 2020":  "2020-08-01",
		"21st August 2020": "2020-08-21",

		// Leading weekday name is now skipped when present.
		"Mon Jan 02 2006": "2006-01-02",
		"Thu May 08 2009": "2009-05-08",

		// Month dd, yyyy / Month dd yyyy
		"September 17, 2012": "2012-09-17",
		"May 17, 2012":       "2012-05-17",
		"September 17 2012":  "2012-09-17",
		"May 7, 2012":        "2012-05-07",
		"June 7, 2012":       "2012-06-07",
		"June 7 2012":        "2012-06-07",

		// Month dd[th,nd,st,rd] yyyy with and without comma.
		"September 17th, 2012": "2012-09-17",
		"September 17th 2012":  "2012-09-17",
		"September 7th, 2012":  "2012-09-07",
		"September 7th 2012":   "2012-09-07",
		"May 1st 2012":         "2012-05-01",
		"May 1st, 2012":        "2012-05-01",
		"May 21st 2012":        "2012-05-21",
		"May 21st, 2012":       "2012-05-21",
		"May 23rd 2012":        "2012-05-23",
		"May 23rd, 2012":       "2012-05-23",
		"June 2nd, 2012":       "2012-06-02",
		"June 2nd 2012":        "2012-06-02",
		"June 22nd, 2012":      "2012-06-22",
		"June 22nd 2012":       "2012-06-22",

		// "Day, dd Mon yyyy" style (RFC-ish, weekday prefix dropped).
		"Fri, 03 Jul 2015": "2015-07-03",
		"Fri, 3 Jul 2015":  "2015-07-03",
		"Thu, 03 Jul 2017": "2017-07-03",
		"Thu, 3 Jul 2017":  "2017-07-03",
		"Tue, 11 Jul 2017": "2017-07-11",
		"Tue, 5 Jul 2017":  "2017-07-05",

		// RFC850 short year and hyphenated separator.
		"Fri, 03-Jul-15":       "2015-07-03",
		"Fri, 03-Jul 2015":     "2015-07-03",
		"Fri, 3-Jul-15":        "2015-07-03",
		"Wednesday, 07-May-09": "2009-05-07",
		"Wednesday, 28-Feb-18": "2018-02-28",

		"Monday, 02 Jan 2006":    "2006-01-02",
		"Wednesday, 28 Feb 2018": "2018-02-28",
		"Wednesday, 2 Feb 2018":  "2018-02-02",
		//  dd mon yyyy  12 Feb 2006, 19:17:08
		"07 Feb 2004": "2004-02-07",
		"7 Feb 2004":  "2004-02-07",
		//  dd-mon-yyyy   12-Feb-2006 19:17:08
		"07-Feb-2004": "2004-02-07",
		//  dd-mon-yy   12-Feb-2006 19:17:08
		"07-Feb-04": "2004-02-07",
		// yyyy-mon-dd    2013-Feb-03
		"2013-Feb-03": "2013-02-03",
		// 03 February 2013
		"03 February 2013": "2013-02-03",
		"3 February 2013":  "2013-02-03",
		// Chinese 年/月/日 are recognised as separators.
		"2014年04月08日": "2014-04-08",
		//  mm/dd/yyyy
		"03/31/2014": "2014-03-31",
		"3/31/2014":  "2014-03-31",
		//  mm/dd/yy
		"08/08/71": "1971-08-08",
		// "8/8/71":   "1971-08-08", // TODO
		//   yyyy/mm/dd
		"2014/04/02": "2014-04-02",
		"2014/03/31": "2014-03-31",
		"2014/4/2":   "2014-04-02",
		//   yyyy-mm-dd
		"2014-04-02": "2014-04-02",
		"2014-03-31": "2014-03-31",
		"2014-4-2":   "2014-04-02",
		// yyyy-mm
		// "2014-04": "2014-04-01", // TODO
		//   yyyy-mm-dd hh:mm:ss AM
		"2014-04-26": "2014-04-26",
		//   yyyy-mm-dd hh:mm:ss,000
		"2014-05-11": "2014-05-11",
		//   yyyy-mm-dd hh:mm:ss +0000
		"2012-08-03": "2012-08-03",
		"2012-8-03":  "2012-08-03",
		"2012-8-3":   "2012-08-03",

		"2014-05-01": "2014-05-01",
		"2014-5-01":  "2014-05-01",
		"2014-05-1":  "2014-05-01",
		// yyyy.mm
		// "2014.05": "2014-05-01", // TODO
		//   mm.dd.yyyy
		"3.31.2014":  "2014-03-31",
		"3.3.2014":   "2014-03-03",
		"03.31.2014": "2014-03-31",
		//   mm.dd.yy
		"08.21.71": "1971-08-21",
		//  yyyymmdd and similar
		// "2014":     "2014-01-01", // TODO
		// "201412":   "2014-12-01", // TODO
		// "20140601": "2014-06-01", // TODO

		"2006-01-02T15:04:05Z07:00":           "2006-01-02", // RFC3339
		"2006-01-02T15:04:05.999999999Z07:00": "2006-01-02", // RFC3339Nano

		"30.11.2023 00:00:00": "2023-11-30",
	}

	for input, expected := range dateTable {
		t.Run(fmt.Sprintf("Normalize(%s)", input), func(t *testing.T) {
			normalized, err := Normalize(input)
			require.NoError(t, err, "Normalize")
			require.Equal(t, expected, normalized, "Normalize")
		})
	}

	invalidDates := []string{
		"6/12/6",
		"6/12/6,",
		"3:28:00",
	}

	for _, invalidDate := range invalidDates {
		t.Run(fmt.Sprintf("Normalize(%s)", invalidDate), func(t *testing.T) {
			normalized, err := Normalize(invalidDate)
			require.Error(t, err, "Should NOT be valid Normalize(%#v): %#v", invalidDate, normalized)
		})
	}
}

func Test_Finder(t *testing.T) {

	deFinderData := map[string][][]int{
		"3:28:00":                nil,
		"2006-01-02":             {[]int{0, 10}},
		"2006-01-02, 2017/12/03": {[]int{0, 10}, []int{12, 22}},
		"jan. 24 2012, 2017/12/03 16. Dezember 98": {[]int{0, 12}, []int{14, 24}, []int{25, 40}},
		"Datum: 25.12.1975 Dezember 1975":          {[]int{7, 17}},
	}

	deFinder := NewFinder("de")
	// enFinder := NewFinder("en")

	for str, allIndices := range deFinderData {
		allResult := deFinder.FindAllIndex([]byte(str), -1)
		if len(allResult) != len(allIndices) {
			for _, indices := range allResult {
				fmt.Println("'" + str[indices[0]:indices[1]] + "'")
			}
			t.Errorf("Found %d Dates in %#v, but expected %d", len(allResult), str, len(allIndices))
		} else {
			for i := range allIndices {
				indices := allIndices[i]
				result := allResult[i]
				if len(result) != 2 {
					t.Errorf("Did not find date in %#v", str)
				}
				date := Date(str[result[0]:result[1]])
				if result[0] != indices[0] || result[1] != indices[1] {
					t.Errorf("Found Date %#v at wrong position in %#v. Expected: %v, Result: %v", date, str, indices, result)
				}
				if !date.Valid() {
					t.Errorf("Invalid Date: %#v", date)
				}
			}
		}
	}
}

func Test_PeriodRange(t *testing.T) {
	periodDates := map[string][2]Date{
		"2018-01": {"2018-01-01", "2018-01-31"},
		"2018-06": {"2018-06-01", "2018-06-30"},
		"2018-12": {"2018-12-01", "2018-12-31"},

		"2018-Q1": {"2018-01-01", "2018-03-31"},
		"2018-Q2": {"2018-04-01", "2018-06-30"},
		"2018-Q3": {"2018-07-01", "2018-09-30"},
		"2018-Q4": {"2018-10-01", "2018-12-31"},
		"2018-q1": {"2018-01-01", "2018-03-31"},
		"2018-q2": {"2018-04-01", "2018-06-30"},
		"2018-q3": {"2018-07-01", "2018-09-30"},
		"2018-q4": {"2018-10-01", "2018-12-31"},

		"2018-H1": {"2018-01-01", "2018-06-30"},
		"2018-H2": {"2018-07-01", "2018-12-31"},
		"2018-h1": {"2018-01-01", "2018-06-30"},
		"2018-h2": {"2018-07-01", "2018-12-31"},

		"1900": {"1900-01-01", "1900-12-31"},
		"2018": {"2018-01-01", "2018-12-31"},

		"2019-W1":  {"2018-12-31", "2019-01-06"},
		"2019-W01": {"2018-12-31", "2019-01-06"},
		"2019-W02": {"2019-01-07", "2019-01-13"},
		// January 1st on a Friday or Saturday belongs to the last
		// week of the previous year, so week 1 starts after it.
		"2020-W53": {"2020-12-28", "2021-01-03"},
		"2021-W01": {"2021-01-04", "2021-01-10"},
		"2022-W01": {"2022-01-03", "2022-01-09"},
	}

	for period, expected := range periodDates {
		t.Run(fmt.Sprintf("PeriodRange(%#v)", period), func(t *testing.T) {
			from, until, err := PeriodRange(period)
			if err != nil {
				t.Fatal(err)
			}
			if from != expected[0] {
				t.Errorf("PeriodRange(%#v) expected from to be %#v but got %#v", period, expected[0], from)
			}
			if until != expected[1] {
				t.Errorf("PeriodRange(%#v) expected until to be %#v but got %#v", period, expected[1], until)
			}
		})
	}

	invalidPeriods := []string{
		"18-01",
		"2018-00",
		"2x18-01",
		"2018-13",
		"2018-1",
		"2018 01",
		"Erik",
		"2018-Q0",
		"2018-Q5",
		"2018-H0",
		"2018-H3",
		// 2021 has only 52 ISO weeks, its "week 53" would be 2022-W01
		"2021-W53",
		"2021-W54",
		"2021-W00",
		"2018-H03",
		"2018-W0",
		"2018-W00",
		"2018-W54",
	}

	for _, period := range invalidPeriods {
		t.Run(fmt.Sprintf("PeriodRange(%s)", period), func(t *testing.T) {
			_, _, err := PeriodRange(period)
			require.Error(t, err, "PeriodRange(%#v)", period)
		})
	}

}
func Test_YearRange(t *testing.T) {
	periodDates := map[int][2]Date{
		-333: {"-333-01-01", "-333-12-31"},
		0:    {"0000-01-01", "0000-12-31"},
		325:  {"0325-01-01", "0325-12-31"},
		2018: {"2018-01-01", "2018-12-31"},
	}

	for year, expected := range periodDates {
		t.Run(fmt.Sprintf("YearRange(%#v)", year), func(t *testing.T) {
			from, until := YearRange(year)
			if from != expected[0] {
				t.Errorf("YearRange(%#v) expected from to be %#v but got %#v", year, expected[0], from)
			}
			if until != expected[1] {
				t.Errorf("YearRange(%#v) expected until to be %#v but got %#v", year, expected[1], until)
			}
		})
	}
}

func Test_YearMonthDay(t *testing.T) {
	dates := map[Date]struct {
		year  int
		month time.Month
		day   int
	}{
		// Normalized
		"2010-12-31": {2010, 12, 31},
		"2000-01-01": {2000, 1, 1},

		// Not normalized
		"31.12.2010":     {2010, 12, 31},
		"1st. Jan. 2000": {2000, 1, 1},
	}

	for date, expected := range dates {
		t.Run(fmt.Sprintf("Date(%s).YearMonthDay()", date), func(t *testing.T) {
			year, month, day := date.YearMonthDay()
			require.Equal(t, expected.year, year)
			require.Equal(t, time.Month(expected.month), month)
			require.Equal(t, expected.day, day)
		})
	}
}

func Test_Date_UnmarshalJSON(t *testing.T) {
	sourceJSON := `{
		"empty": "",
		"null": null,
		"notNull": "2012-12-12",
		"invalid": "Not a date!"
	}`
	s := struct {
		Empty   Date `json:"empty"`
		Null    Date `json:"null"`
		NotNull Date `json:"notNull"`
		Invalid Date `json:"invalid"`
	}{}
	err := json.Unmarshal([]byte(sourceJSON), &s)
	require.NoError(t, err, "json.Unmarshal")
	require.Equal(t, Date(""), s.Empty, "empty JSON string is Null")
	require.Equal(t, Date(""), s.Null, "JSON null value as Null")
	require.Equal(t, Date("2012-12-12"), s.NotNull, "valid Date")
	require.Equal(t, Date("Not a date!"), s.Invalid, "invalid Date parsed as is, without error")
	require.False(t, s.Invalid.Valid(), "invalid Date parsed as is, not valid")
}

func TestDate_Normalized(t *testing.T) {
	tests := []struct {
		name    string
		date    Date
		lang    []language.Code
		want    Date
		wantErr bool
	}{
		// Valid:
		{name: "earliest", date: "0001-01-01", lang: nil, want: "0001-01-01"},
		{name: "3000-01-01", date: "3000-01-01", lang: nil, want: "3000-01-01"},
		{name: "3000/01/01", date: "3000/01/01", lang: nil, want: "3000-01-01"},
		{name: "3000.01.01", date: "3000.01.01", lang: nil, want: "3000-01-01"},
		// Invalid, returned unchanged with the error:
		{name: "empty", date: "", lang: nil, want: "", wantErr: true},
		{name: "invalid year", date: "0000-01-01", lang: nil, want: "0000-01-01", wantErr: true},
		{name: "invalid month", date: "2020-00-01", lang: nil, want: "2020-00-01", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.date.Normalized(tt.lang...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Date.Normalized() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Date.Normalized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDate_NormalizedErrorReturnsReceiver(t *testing.T) {
	// The normalization contract of the module (see
	// docs/explanation-validation-and-normalization.md): on error the
	// receiver comes back unchanged, so a caller that ignores or logs the
	// error still has the input instead of an empty date.
	for _, d := range []Date{"foo", "", "  ", "2020-13-01", "0000-00-00"} {
		t.Run(string(d), func(t *testing.T) {
			got, err := d.Normalized()
			require.Error(t, err)
			assert.Equal(t, d, got)

			got, err = d.Normalized(language.EN)
			require.Error(t, err)
			assert.Equal(t, d, got, "with language hint")

			got, err = Normalize(string(d))
			require.Error(t, err)
			assert.Equal(t, d, got, "Normalize")
		})
	}
}

func TestYearWeekRange(t *testing.T) {
	type args struct {
		year int
		week int
	}
	tests := []struct {
		name       string
		args       args
		wantMonday Date
		wantSunday Date
	}{
		// Weeks outside of 1..53 count from the Monday of ISO week 1 (2021-01-04),
		// so week 0 of 2021 is ISO week 2020-W53.
		{name: "2021/-1", args: args{year: 2021, week: -1}, wantMonday: "2020-12-21", wantSunday: "2020-12-27"},
		{name: "2021/0", args: args{year: 2021, week: 0}, wantMonday: "2020-12-28", wantSunday: "2021-01-03"},
		{name: "2021/1", args: args{year: 2021, week: 1}, wantMonday: "2021-01-04", wantSunday: "2021-01-10"},
		{name: "2020/52", args: args{year: 2020, week: 52}, wantMonday: "2020-12-21", wantSunday: "2020-12-27"},
		{name: "2020/53", args: args{year: 2020, week: 53}, wantMonday: "2020-12-28", wantSunday: "2021-01-03"},
		{name: "2020/54", args: args{year: 2020, week: 54}, wantMonday: "2021-01-04", wantSunday: "2021-01-10"},
		{name: "2020/55", args: args{year: 2020, week: 55}, wantMonday: "2021-01-11", wantSunday: "2021-01-17"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMonday, gotSunday := YearWeekRange(tt.args.year, tt.args.week)
			if gotMonday != tt.wantMonday {
				t.Errorf("YearWeekRange() gotMonday = %v, want %v", gotMonday, tt.wantMonday)
			}
			if gotSunday != tt.wantSunday {
				t.Errorf("YearWeekRange() gotSunday = %v, want %v", gotSunday, tt.wantSunday)
			}
		})
	}
}

func TestDate_AddMonths(t *testing.T) {
	type args struct {
		months int
	}
	tests := []struct {
		date Date
		args args
		want Date
	}{
		{date: "2020-01-01", args: args{months: 1}, want: "2020-02-01"},
		{date: "2020-01-01", args: args{months: 2}, want: "2020-03-01"},
		{date: "2020-12-01", args: args{months: 1}, want: "2021-01-01"},
		{date: "2020-01-01", args: args{months: 12}, want: "2021-01-01"},
		{date: "2020-01-01", args: args{months: -1}, want: "2019-12-01"},
		// TODO
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s + %d", tt.date, tt.args.months), func(t *testing.T) {
			if got := tt.date.AddMonths(tt.args.months); got != tt.want {
				t.Errorf("Date.AddMonths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDate_YearMonthDay(t *testing.T) {
	tests := []struct {
		name      string
		date      Date
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{name: "valid date", date: "2023-12-25", wantYear: 2023, wantMonth: time.December, wantDay: 25},
		{name: "leap year", date: "2020-02-29", wantYear: 2020, wantMonth: time.February, wantDay: 29},
		{name: "invalid date", date: "invalid", wantYear: 0, wantMonth: 0, wantDay: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotYear, gotMonth, gotDay := tt.date.YearMonthDay()
			if gotYear != tt.wantYear {
				t.Errorf("Date.YearMonthDay() gotYear = %v, want %v", gotYear, tt.wantYear)
			}
			if gotMonth != tt.wantMonth {
				t.Errorf("Date.YearMonthDay() gotMonth = %v, want %v", gotMonth, tt.wantMonth)
			}
			if gotDay != tt.wantDay {
				t.Errorf("Date.YearMonthDay() gotDay = %v, want %v", gotDay, tt.wantDay)
			}
		})
	}
}

// Constructor function tests
func TestMust(t *testing.T) {
	t.Run("valid date", func(t *testing.T) {
		d := Must("2023-12-25")
		require.Equal(t, Date("2023-12-25"), d)
	})

	t.Run("panic on invalid", func(t *testing.T) {
		require.Panics(t, func() {
			Must("invalid")
		})
	})
}

func TestOf(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month time.Month
		day   int
		want  Date
	}{
		{name: "regular date", year: 2023, month: time.December, day: 25, want: "2023-12-25"},
		{name: "overflow day", year: 2023, month: time.October, day: 32, want: "2023-11-01"},
		{name: "overflow month", year: 2023, month: 13, day: 1, want: "2024-01-01"},
		{name: "zero day", year: 2023, month: time.March, day: 0, want: "2023-02-28"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Of(tt.year, tt.month, tt.day)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestOfTime(t *testing.T) {
	t.Run("valid time", func(t *testing.T) {
		tm := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)
		d := OfTime(tm)
		require.Equal(t, Date("2023-12-25"), d)
	})

	t.Run("zero time", func(t *testing.T) {
		d := OfTime(time.Time{})
		require.Equal(t, Date(""), d)
	})
}

func TestOfTimePtr(t *testing.T) {
	t.Run("valid time", func(t *testing.T) {
		tm := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)
		d := OfTimePtr(&tm)
		require.Equal(t, NullableDate("2023-12-25"), d)
	})

	t.Run("nil pointer", func(t *testing.T) {
		d := OfTimePtr(nil)
		require.Equal(t, Null, d)
	})

	t.Run("zero time", func(t *testing.T) {
		tm := time.Time{}
		d := OfTimePtr(&tm)
		require.Equal(t, Null, d)
	})
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		layout  string
		value   string
		want    Date
		wantErr bool
	}{
		{name: "RFC3339", layout: time.RFC3339, value: "2023-12-25T15:30:00Z", want: "2023-12-25"},
		{name: "DateOnly", layout: time.DateOnly, value: "2023-12-25", want: "2023-12-25"},
		{name: "invalid", layout: time.DateOnly, value: "invalid", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.layout, tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

// Validation tests
func TestDate_IsZero(t *testing.T) {
	tests := []struct {
		name string
		date Date
		want bool
	}{
		{name: "empty string", date: "", want: true},
		{name: "0000-00-00", date: "0000-00-00", want: true},
		{name: "0001-01-01", date: "0001-01-01", want: true},
		{name: "valid date", date: "2023-12-25", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.date.IsZero())
		})
	}
}

func TestDate_Valid(t *testing.T) {
	tests := []struct {
		name string
		date Date
		want bool
	}{
		{name: "valid normalized", date: "2023-12-25", want: true},
		{name: "valid unnormalized", date: "25.12.2023", want: true},
		{name: "invalid", date: "invalid", want: false},
		{name: "empty", date: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.date.Valid())
		})
	}
}

func TestDate_ValidAndNormalized(t *testing.T) {
	tests := []struct {
		name string
		date Date
		want bool
	}{
		{name: "valid normalized", date: "2023-12-25", want: true},
		{name: "valid unnormalized", date: "25.12.2023", want: false},
		{name: "invalid", date: "invalid", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.date.ValidAndNormalized())
		})
	}
}

// Comparison tests
func TestDate_Compare(t *testing.T) {
	tests := []struct {
		name  string
		date1 Date
		date2 Date
		want  int
	}{
		{name: "before", date1: "2023-12-24", date2: "2023-12-25", want: -1},
		{name: "after", date1: "2023-12-26", date2: "2023-12-25", want: 1},
		{name: "equal", date1: "2023-12-25", date2: "2023-12-25", want: 0},
		{name: "equal non-normalized", date1: "25.12.2023", date2: "2023-12-25", want: 0},
		// Compare is a sort order, so invalid dates must still be ordered
		// consistently: like the empty string, before every valid date and
		// equal to each other. Compare must not use the invalid input that
		// Normalized returns with its error.
		{name: "invalid before valid", date1: "bad", date2: "0001-01-01", want: -1},
		{name: "valid after invalid", date1: "2023-12-25", date2: "bad", want: 1},
		{name: "invalid equals invalid", date1: "bad", date2: "worse", want: 0},
		{name: "invalid equals empty", date1: "zzz", date2: "", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.date1.Compare(tt.date2))
		})
	}
}

func TestDate_After(t *testing.T) {
	require.True(t, Date("2023-12-26").After("2023-12-25"))
	require.False(t, Date("2023-12-24").After("2023-12-25"))
	require.False(t, Date("2023-12-25").After("2023-12-25"))
}

func TestDate_Before(t *testing.T) {
	require.True(t, Date("2023-12-24").Before("2023-12-25"))
	require.False(t, Date("2023-12-26").Before("2023-12-25"))
	require.False(t, Date("2023-12-25").Before("2023-12-25"))
}

func TestDate_WithinIncl(t *testing.T) {
	tests := []struct {
		name  string
		date  Date
		from  Date
		until Date
		want  bool
	}{
		{name: "within", date: "2023-12-25", from: "2023-12-20", until: "2023-12-30", want: true},
		{name: "equal to from", date: "2023-12-20", from: "2023-12-20", until: "2023-12-30", want: true},
		{name: "equal to until", date: "2023-12-30", from: "2023-12-20", until: "2023-12-30", want: true},
		{name: "before", date: "2023-12-19", from: "2023-12-20", until: "2023-12-30", want: false},
		{name: "after", date: "2023-12-31", from: "2023-12-20", until: "2023-12-30", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.date.WithinIncl(tt.from, tt.until))
		})
	}
}

func TestDate_BetweenExcl(t *testing.T) {
	tests := []struct {
		name   string
		date   Date
		after  Date
		before Date
		want   bool
	}{
		{name: "between", date: "2023-12-25", after: "2023-12-20", before: "2023-12-30", want: true},
		{name: "equal to after", date: "2023-12-20", after: "2023-12-20", before: "2023-12-30", want: false},
		{name: "equal to before", date: "2023-12-30", after: "2023-12-20", before: "2023-12-30", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.date.BetweenExcl(tt.after, tt.before))
		})
	}
}

func TestDate_ComparisonWithInvalidDate(t *testing.T) {
	// An invalid date has no position in time, so every comparison with
	// it must be false in both directions, like NormalizedEqual. Treating
	// it as the zero time made "2020-01-01" equal to or after "bad", so an
	// unparsable deadline passed every "is it reached yet" check.
	valid := Date("2020-01-01")
	tm := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	for _, invalid := range []Date{"bad", "", "0000-00-00", "2020-13-01"} {
		t.Run(string(invalid), func(t *testing.T) {
			assert.False(t, valid.After(invalid), "valid.After")
			assert.False(t, valid.Before(invalid), "valid.Before")
			assert.False(t, invalid.After(valid), "invalid.After")
			assert.False(t, invalid.Before(valid), "invalid.Before")
			assert.False(t, valid.EqualOrAfter(invalid), "valid.EqualOrAfter")
			assert.False(t, valid.EqualOrBefore(invalid), "valid.EqualOrBefore")
			assert.False(t, invalid.EqualOrAfter(valid), "invalid.EqualOrAfter")
			assert.False(t, invalid.EqualOrBefore(valid), "invalid.EqualOrBefore")
			assert.False(t, invalid.EqualOrAfter(invalid), "invalid.EqualOrAfter(invalid)")

			assert.False(t, valid.WithinIncl(invalid, "2030-01-01"), "WithinIncl invalid from")
			assert.False(t, valid.WithinIncl("2010-01-01", invalid), "WithinIncl invalid until")
			assert.False(t, invalid.WithinIncl("0001-01-01", "9999-12-31"), "invalid.WithinIncl")
			assert.False(t, valid.BetweenExcl(invalid, "2030-01-01"), "BetweenExcl invalid after")
			assert.False(t, valid.BetweenExcl("2010-01-01", invalid), "BetweenExcl invalid before")
			assert.False(t, invalid.BetweenExcl("0001-01-01", "9999-12-31"), "invalid.BetweenExcl")

			assert.False(t, invalid.AfterTime(tm), "AfterTime")
			assert.False(t, invalid.BeforeTime(tm), "BeforeTime")
			assert.False(t, invalid.AfterToday(), "AfterToday")
			assert.False(t, invalid.BeforeToday(), "BeforeToday")
		})
	}
	// The zero time.Time date "0001-01-01" is valid and stays comparable
	assert.True(t, Date("0001-01-01").Before(valid))
	assert.True(t, valid.After("0001-01-01"))
	assert.True(t, Date("0001-01-01").BeforeTime(tm))
}

// Arithmetic tests
func TestDate_AddDate(t *testing.T) {
	d := Date("2023-12-25")
	require.Equal(t, Date("2024-01-26"), d.AddDate(0, 1, 1))
	require.Equal(t, Date("2024-12-25"), d.AddDate(1, 0, 0))
}

func TestDate_AddYears(t *testing.T) {
	d := Date("2023-12-25")
	require.Equal(t, Date("2024-12-25"), d.AddYears(1))
	require.Equal(t, Date("2022-12-25"), d.AddYears(-1))
}

func TestDate_AddDays(t *testing.T) {
	d := Date("2023-12-25")
	require.Equal(t, Date("2023-12-26"), d.AddDays(1))
	require.Equal(t, Date("2023-12-24"), d.AddDays(-1))
}

func TestDate_Sub(t *testing.T) {
	d1 := Date("2023-12-26")
	d2 := Date("2023-12-25")
	require.Equal(t, 24*time.Hour, d1.Sub(d2))
}

// Component extraction tests
func TestDate_Year(t *testing.T) {
	require.Equal(t, 2023, Date("2023-12-25").Year())
	require.Equal(t, 0, Date("invalid").Year())
}

func TestDate_Month(t *testing.T) {
	require.Equal(t, time.December, Date("2023-12-25").Month())
	require.Equal(t, time.Month(0), Date("invalid").Month())
}

func TestDate_Day(t *testing.T) {
	require.Equal(t, 25, Date("2023-12-25").Day())
	require.Equal(t, 0, Date("invalid").Day())
}

func TestDate_Weekday(t *testing.T) {
	require.Equal(t, time.Monday, Date("2023-12-25").Weekday())
}

func TestDate_ISOWeek(t *testing.T) {
	year, week := Date("2023-12-25").ISOWeek()
	require.Equal(t, 2023, year)
	require.Equal(t, 52, week)
}

// Format tests
func TestDate_Format(t *testing.T) {
	d := Date("2023-12-25")
	require.Equal(t, "2023-12-25", d.Format(Layout))
	require.Equal(t, "12/25/2023", d.Format("01/02/2006"))
	require.Equal(t, "", d.Format(""))
}

// Database tests
func TestDate_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    Date
		wantErr bool
	}{
		{name: "string", value: "2023-12-25", want: "2023-12-25"},
		// Several SQL drivers return DATE and text columns as []byte,
		// which must scan exactly like the same string.
		{name: "[]byte", value: []byte("2023-12-25"), want: "2023-12-25"},
		{name: "[]byte is normalized", value: []byte("25.12.2023"), want: "2023-12-25"},
		{name: "invalid []byte", value: []byte("invalid"), wantErr: true},
		{name: "time.Time", value: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), want: "2023-12-25"},
		{name: "nil", value: nil, want: ""},
		{name: "invalid type", value: 123, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			err := d.Scan(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, d)
			}
		})
	}
}

func TestDate_Value(t *testing.T) {
	tests := []struct {
		name    string
		date    Date
		want    any
		wantErr bool
	}{
		{name: "valid date", date: "2023-12-25", want: "2023-12-25"},
		{name: "zero date", date: "", want: nil},
		{name: "invalid date", date: "invalid", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.date.Value()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

// MarshalJSON writes a Date normalized like Value, because Validate accepts
// other formats like "25.12.2023": without normalization API consumers
// would get dates that are not in the "date" format of the JSON schema.
// A date that can't be normalized is written unchanged instead of failing
// the write.
func TestDate_MarshalJSON(t *testing.T) {
	tests := []struct {
		date Date
		want string
	}{
		{date: "2023-12-25", want: `"2023-12-25"`},
		{date: "25.12.2023", want: `"2023-12-25"`},
		{date: "December 25, 2023", want: `"2023-12-25"`},
		{date: "invalid", want: `"invalid"`},
		{date: "", want: `""`},
	}
	for _, tt := range tests {
		j, err := json.Marshal(tt.date)
		require.NoError(t, err)
		assert.Equal(t, tt.want, string(j), "json.Marshal(Date(%q))", string(tt.date))
	}
}

// MarshalText must write the normalized date like Value and MarshalJSON,
// because encoding/xml uses it: without it XML output would carry
// whatever spelling Validate accepted, like "25.12.2023". A date that
// can't be normalized is written unchanged instead of failing the write.
// Value: protects=Date.MarshalText and encoding/xml write the normalized date, an invalid date unchanged; fails_when=MarshalText returns []byte(date) instead of []byte(date.String()); why_new=MarshalText is a new method without any test or caller in the module; seam=none; no_claim=strfmt output, it formats dates through TypeFormatters
func TestDate_MarshalText(t *testing.T) {
	for _, tt := range []struct {
		date Date
		want string
	}{
		{date: "25.12.2023", want: "2023-12-25"},
		{date: "invalid", want: "invalid"},
	} {
		text, err := tt.date.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, tt.want, string(text), "Date(%q).MarshalText()", string(tt.date))
	}

	type row struct{ D Date }
	data, err := xml.Marshal(row{D: "25.12.2023"})
	require.NoError(t, err)
	assert.Equal(t, `<row><D>2023-12-25</D></row>`, string(data))
}

// Additional constructor function tests
func TestOfToday(t *testing.T) {
	d := OfToday()
	assert.True(t, d.Valid())
	assert.True(t, d.IsToday())
}

func TestOfNowInUTC(t *testing.T) {
	d := OfNowInUTC()
	assert.True(t, d.Valid())
}

func TestOfTodayIn(t *testing.T) {
	loc := time.UTC
	d := OfTodayIn(loc)
	assert.True(t, d.Valid())
}

func TestOfYesterday(t *testing.T) {
	d := OfYesterday()
	assert.True(t, d.Valid())
	assert.True(t, d.BeforeToday())
}

func TestOfTomorrow(t *testing.T) {
	d := OfTomorrow()
	assert.True(t, d.Valid())
	assert.True(t, d.AfterToday())
}

func TestYearWeekMonday(t *testing.T) {
	// ISO 8601 week 1 is the week containing January 4th. Depending on the
	// weekday of January 1st its Monday is in December of the previous year
	// or in the first days of January. A wrong Monday shifts every week
	// period of the year (YearWeekRange, PeriodRange("YYYY-Wnn")) by a week.
	tests := []struct {
		name string
		year int
		week int
		want Date
	}{
		{name: "2018 week 1, Jan 1 is a Monday", year: 2018, week: 1, want: "2018-01-01"},
		{name: "2019 week 1, Jan 1 is a Tuesday", year: 2019, week: 1, want: "2018-12-31"},
		{name: "2020 week 1, Jan 1 is a Wednesday", year: 2020, week: 1, want: "2019-12-30"},
		{name: "2015 week 1, Jan 1 is a Thursday", year: 2015, week: 1, want: "2014-12-29"},
		{name: "2021 week 1, Jan 1 is a Friday", year: 2021, week: 1, want: "2021-01-04"},
		{name: "2022 week 1, Jan 1 is a Saturday", year: 2022, week: 1, want: "2022-01-03"},
		{name: "2023 week 1, Jan 1 is a Sunday", year: 2023, week: 1, want: "2023-01-02"},
		{name: "2020 week 53", year: 2020, week: 53, want: "2020-12-28"},
		{name: "2021 week 52", year: 2021, week: 52, want: "2021-12-27"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := YearWeekMonday(tt.year, tt.week)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, time.Monday, got.Weekday())
			year, week := got.MidnightUTC().ISOWeek()
			assert.Equal(t, tt.year, year, "time.Time.ISOWeek year")
			assert.Equal(t, tt.week, week, "time.Time.ISOWeek week")
		})
	}

	// Cross-check every week of a range of years covering all weekdays
	// of January 1st (and leap years) against the standard library.
	for year := 1999; year <= 2040; year++ {
		// December 28th is always in the last ISO week of its year
		_, lastWeek := time.Date(year, 12, 28, 0, 0, 0, 0, time.UTC).ISOWeek()
		for week := 1; week <= lastWeek; week++ {
			monday := YearWeekMonday(year, week)
			gotYear, gotWeek := monday.MidnightUTC().ISOWeek()
			if gotYear != year || gotWeek != week || monday.Weekday() != time.Monday {
				t.Errorf("YearWeekMonday(%d, %d) = %s which is a %s in ISO week %d-W%02d", year, week, monday, monday.Weekday(), gotYear, gotWeek)
			}
		}
	}
}

func TestStringIsDate(t *testing.T) {
	assert.True(t, StringIsDate("2023-12-25"))
	assert.True(t, StringIsDate("25.12.2023"))
	assert.False(t, StringIsDate("invalid"))
	assert.False(t, StringIsDate(""))
}

func TestFromUntilFromYearAndMonths(t *testing.T) {
	tests := []struct {
		name      string
		year      string
		months    string
		wantFrom  Date
		wantUntil Date
		wantErr   bool
	}{
		{name: "full year", year: "2023", months: "", wantFrom: "2023-01-01", wantUntil: "2023-12-31"},
		{name: "single month", year: "2023", months: "06", wantFrom: "2023-06-01", wantUntil: "2023-06-30"},
		{name: "month range", year: "2023", months: "06-09", wantFrom: "2023-06-01", wantUntil: "2023-09-30"},
		{name: "empty year", year: "", months: "", wantFrom: "", wantUntil: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrom, gotUntil, err := FromUntilFromYearAndMonths(tt.year, tt.months)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantFrom, gotFrom)
				assert.Equal(t, tt.wantUntil, gotUntil)
			}
		})
	}
}

// String and conversion tests
func TestDate_String(t *testing.T) {
	assert.Equal(t, "2023-12-25", Date("2023-12-25").String())
	assert.Equal(t, "2023-12-25", Date("25.12.2023").String())
	assert.Equal(t, "invalid", Date("invalid").String())
}

func TestDate_Nullable(t *testing.T) {
	d := Date("2023-12-25")
	n := d.Nullable()
	assert.Equal(t, NullableDate("2023-12-25"), n)
	assert.False(t, n.IsNull())

	d2 := Date("")
	n2 := d2.Nullable()
	assert.True(t, n2.IsNull())
}

func TestDate_ScanString(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		validate bool
		want     Date
		wantErr  bool
	}{
		{name: "valid with validation", source: "2023-12-25", validate: true, want: "2023-12-25"},
		{name: "valid no validation", source: "2023-12-25", validate: false, want: "2023-12-25"},
		{name: "invalid with validation", source: "invalid", validate: true, wantErr: true},
		{name: "invalid no validation", source: "invalid", validate: false, want: "invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			err := d.ScanString(tt.source, tt.validate)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, d)
			}
		})
	}
}

// Time conversion tests
func TestDate_Time(t *testing.T) {
	d := Date("2023-12-25")
	tm := d.Time(15, 30, 45, time.UTC)
	assert.Equal(t, 2023, tm.Year())
	assert.Equal(t, time.December, tm.Month())
	assert.Equal(t, 25, tm.Day())
	assert.Equal(t, 15, tm.Hour())
	assert.Equal(t, 30, tm.Minute())
	assert.Equal(t, 45, tm.Second())
}

func TestDate_TimeLocal(t *testing.T) {
	d := Date("2023-12-25")
	tm := d.TimeLocal(15, 30, 45)
	assert.Equal(t, 2023, tm.Year())
	assert.Equal(t, time.Local, tm.Location())
}

func TestDate_TimeUTC(t *testing.T) {
	d := Date("2023-12-25")
	tm := d.TimeUTC(15, 30, 45)
	assert.Equal(t, 2023, tm.Year())
	assert.Equal(t, time.UTC, tm.Location())
}

func TestDate_MidnightUTC(t *testing.T) {
	d := Date("2023-12-25")
	tm := d.MidnightUTC()
	assert.Equal(t, 2023, tm.Year())
	assert.Equal(t, time.December, tm.Month())
	assert.Equal(t, 25, tm.Day())
	assert.Equal(t, 0, tm.Hour())
	assert.Equal(t, time.UTC, tm.Location())

	// Invalid date returns zero time
	d2 := Date("invalid")
	tm2 := d2.MidnightUTC()
	assert.True(t, tm2.IsZero())

	// time.Parse accepts the year 0 that Valid rejects,
	// an invalid date must still return the zero time.
	assert.False(t, Date("0000-01-01").Valid())
	assert.True(t, Date("0000-01-01").MidnightUTC().IsZero(), "year 0")
	assert.True(t, Date("0000-01-01").MidnightInLocation(time.Local).IsZero(), "year 0 local")
	assert.True(t, NullableDate("0000-05-05").MidnightUTC().IsNull(), "NullableDate year 0")
}

// TestDate_ComparisonsMatchValid protects that the comparison methods,
// which parse each date only once, return false for exactly the dates
// that Valid rejects and compare the same instants as MidnightUTC.
func TestDate_ComparisonsMatchValid(t *testing.T) {
	dates := []Date{
		"", "invalid", "0000-00-00", "0000-01-01", "0001-01-01", "0001-01-02",
		"2023-02-30", "2023-12-24", "2023-12-25", "25.12.2023", "12/26/2023",
		" 2023-12-25 ", "1st August 2020",
	}
	for _, a := range dates {
		for _, b := range dates {
			valid := a.Valid() && b.Valid()
			ta, tb := a.MidnightUTC(), b.MidnightUTC()
			assert.Equal(t, valid && ta.After(tb), a.After(b), "Date(%q).After(%q)", a, b)
			assert.Equal(t, valid && ta.Before(tb), a.Before(b), "Date(%q).Before(%q)", a, b)
			assert.Equal(t, a.Valid() && ta.After(tb), a.AfterTime(tb), "Date(%q).AfterTime(%s)", a, tb)
			assert.Equal(t, a.Valid() && ta.Before(tb), a.BeforeTime(tb), "Date(%q).BeforeTime(%s)", a, tb)
			for _, c := range dates {
				tc := c.MidnightUTC()
				valid3 := valid && c.Valid()
				assert.Equal(t, valid3 && !ta.Before(tb) && !ta.After(tc), a.WithinIncl(b, c), "Date(%q).WithinIncl(%q, %q)", a, b, c)
				assert.Equal(t, valid3 && ta.After(tb) && ta.Before(tc), a.BetweenExcl(b, c), "Date(%q).BetweenExcl(%q, %q)", a, b, c)
			}
		}
	}
}

func TestDate_Midnight(t *testing.T) {
	d := Date("2023-12-25")
	tm := d.Midnight()
	assert.Equal(t, 2023, tm.Year())
	assert.Equal(t, 0, tm.Hour())
	assert.Equal(t, time.Local, tm.Location())
}

func TestDate_MidnightInLocation(t *testing.T) {
	d := Date("2023-12-25")
	tm := d.MidnightInLocation(time.UTC)
	assert.Equal(t, 2023, tm.Year())
	assert.Equal(t, time.UTC, tm.Location())
}

// Normalization method tests
func TestDate_NormalizedOrUnchanged(t *testing.T) {
	assert.Equal(t, Date("2023-12-25"), Date("25.12.2023").NormalizedOrUnchanged())
	assert.Equal(t, Date("invalid"), Date("invalid").NormalizedOrUnchanged())
}

func TestDate_NormalizedOrNull(t *testing.T) {
	assert.Equal(t, NullableDate("2023-12-25"), Date("25.12.2023").NormalizedOrNull())
	assert.Equal(t, Null, Date("invalid").NormalizedOrNull())
}

func TestDate_NormalizedEqual(t *testing.T) {
	assert.True(t, Date("2023-12-25").NormalizedEqual("25.12.2023"))
	assert.False(t, Date("2023-12-25").NormalizedEqual("2023-12-26"))
	// An invalid date has no normalized value, so it must not be equal
	// to anything: two different garbage strings are not the same date,
	// and neither is an identical invalid string (like NaN != NaN).
	assert.False(t, Date("bad").NormalizedEqual("worse"))
	assert.False(t, Date("bad").NormalizedEqual("bad"))
	assert.False(t, Date("bad").NormalizedEqual("2023-12-25"))
	assert.False(t, Date("2023-12-25").NormalizedEqual("bad"))
	// "" is Invalid for Date, see NullableDate for a valid empty date
	assert.False(t, Date("").NormalizedEqual(""))
}

// Additional comparison tests
func TestDate_EqualOrAfter(t *testing.T) {
	assert.True(t, Date("2023-12-26").EqualOrAfter("2023-12-25"))
	assert.True(t, Date("2023-12-25").EqualOrAfter("2023-12-25"))
	assert.False(t, Date("2023-12-24").EqualOrAfter("2023-12-25"))
}

func TestDate_EqualOrBefore(t *testing.T) {
	assert.True(t, Date("2023-12-24").EqualOrBefore("2023-12-25"))
	assert.True(t, Date("2023-12-25").EqualOrBefore("2023-12-25"))
	assert.False(t, Date("2023-12-26").EqualOrBefore("2023-12-25"))
}

func TestDate_AfterTime(t *testing.T) {
	d := Date("2023-12-25")
	tm := time.Date(2023, 12, 24, 23, 59, 59, 0, time.UTC)
	assert.True(t, d.AfterTime(tm))

	tm2 := time.Date(2023, 12, 25, 0, 0, 1, 0, time.UTC)
	assert.False(t, d.AfterTime(tm2))
}

func TestDate_BeforeTime(t *testing.T) {
	d := Date("2023-12-25")
	tm := time.Date(2023, 12, 26, 0, 0, 1, 0, time.UTC)
	assert.True(t, d.BeforeTime(tm))

	tm2 := time.Date(2023, 12, 24, 23, 59, 59, 0, time.UTC)
	assert.False(t, d.BeforeTime(tm2))
}

// Arithmetic tests
func TestDate_Add(t *testing.T) {
	d := Date("2023-12-25")
	d2 := d.Add(24 * time.Hour)
	assert.Equal(t, Date("2023-12-26"), d2)
}

// Period boundary tests
func TestDate_BeginningOfWeek(t *testing.T) {
	d := Date("2023-12-27") // Wednesday

	// Test with Monday as week start (ISO 8601)
	beginning := d.BeginningOfWeek(time.Monday)
	assert.True(t, beginning.Valid())
	assert.Equal(t, time.Monday, beginning.Weekday())
	assert.Equal(t, Date("2023-12-25"), beginning) // Monday before Wednesday

	// Test with Sunday as week start (US style)
	beginningSunday := d.BeginningOfWeek(time.Sunday)
	assert.True(t, beginningSunday.Valid())
	assert.Equal(t, time.Sunday, beginningSunday.Weekday())
	assert.Equal(t, Date("2023-12-24"), beginningSunday) // Sunday before Wednesday
}

func TestDate_BeginningOfMonth(t *testing.T) {
	d := Date("2023-12-25")
	assert.Equal(t, Date("2023-12-01"), d.BeginningOfMonth())
}

func TestDate_BeginningOfQuarter(t *testing.T) {
	d := Date("2023-11-15")
	assert.Equal(t, Date("2023-10-01"), d.BeginningOfQuarter())
}

func TestDate_BeginningOfYear(t *testing.T) {
	d := Date("2023-12-25")
	assert.Equal(t, Date("2023-01-01"), d.BeginningOfYear())
}

func TestDate_EndOfWeek(t *testing.T) {
	d := Date("2023-12-25") // Monday

	// Test with Monday as week start (ISO 8601) - ends on Sunday
	end := d.EndOfWeek(time.Monday)
	assert.True(t, end.Valid())
	assert.Equal(t, time.Sunday, end.Weekday())
	assert.Equal(t, Date("2023-12-31"), end) // Sunday after Monday

	// Test with Sunday as week start (US style) - ends on Saturday
	endSaturday := d.EndOfWeek(time.Sunday)
	assert.True(t, endSaturday.Valid())
	assert.Equal(t, time.Saturday, endSaturday.Weekday())
	assert.Equal(t, Date("2023-12-30"), endSaturday) // Saturday after Monday
}

func TestDate_EndOfMonth(t *testing.T) {
	d := Date("2023-12-25")
	assert.Equal(t, Date("2023-12-31"), d.EndOfMonth())
}

func TestDate_EndOfQuarter(t *testing.T) {
	d := Date("2023-11-15")
	assert.Equal(t, Date("2023-12-31"), d.EndOfQuarter())
}

func TestDate_EndOfYear(t *testing.T) {
	d := Date("2023-06-15")
	assert.Equal(t, Date("2023-12-31"), d.EndOfYear())
}

func TestDate_LastMonday(t *testing.T) {
	d := Date("2023-12-27") // Wednesday
	monday := d.LastMonday()
	assert.True(t, monday.Valid())
	assert.Equal(t, time.Monday, monday.Weekday())
	assert.Equal(t, Date("2023-12-25"), monday) // Monday before Wednesday
	assert.True(t, monday.EqualOrBefore(d))
}

func TestDate_NextSunday(t *testing.T) {
	d := Date("2023-12-25") // Monday
	sunday := d.NextSunday()
	assert.True(t, sunday.Valid())
	assert.Equal(t, time.Sunday, sunday.Weekday())
	assert.Equal(t, Date("2023-12-31"), sunday) // Sunday after Monday
	assert.True(t, sunday.EqualOrAfter(d))
}

// Today comparison tests
func TestDate_IsToday(t *testing.T) {
	today := OfToday()
	assert.True(t, today.IsToday())
	assert.False(t, Date("2020-01-01").IsToday())
	// Like After and Before, IsToday compares the date, not its spelling:
	// a valid but non-normalized today is still today.
	assert.True(t, Date(today.Format("02.01.2006")).IsToday())
	assert.False(t, Date("invalid").IsToday())
}

func TestDate_IsTodayInUTC(t *testing.T) {
	today := OfNowInUTC()
	assert.True(t, today.IsTodayInUTC())
	// A valid but non-normalized today is still today, see TestDate_IsToday
	assert.True(t, Date(today.Format("02.01.2006")).IsTodayInUTC())
	// Normalizing first must not make other days or a parse error count
	// as today, like IsToday in TestDate_IsToday.
	// Value: protects=IsTodayInUTC is false for another day, the next UTC day and an invalid date; fails_when=IsTodayInUTC returns err == nil without comparing to OfNowInUTC; why_new=the test only had true cases, so an always-true-for-valid result passed; seam=none; no_claim=behavior across a UTC midnight between OfNowInUTC calls
	assert.False(t, Date("2020-01-01").IsTodayInUTC())
	assert.False(t, Date("invalid").IsTodayInUTC())
	assert.False(t, today.AddDays(1).IsTodayInUTC())
}

func TestDate_AfterToday(t *testing.T) {
	tomorrow := OfTomorrow()
	assert.True(t, tomorrow.AfterToday())
	assert.False(t, OfYesterday().AfterToday())
}

func TestDate_AfterTodayInUTC(t *testing.T) {
	future := OfNowInUTC().AddDays(1)
	assert.True(t, future.AfterTodayInUTC())
}

func TestDate_BeforeToday(t *testing.T) {
	yesterday := OfYesterday()
	assert.True(t, yesterday.BeforeToday())
	assert.False(t, OfTomorrow().BeforeToday())
}

func TestDate_BeforeTodayInUTC(t *testing.T) {
	past := OfNowInUTC().AddDays(-1)
	assert.True(t, past.BeforeTodayInUTC())
}

// Schema test
func TestDate_JSONSchema(t *testing.T) {
	schema := Date("").JSONSchema()
	assert.NotNil(t, schema)
	assert.Equal(t, "Date", schema.Title)
	assert.Equal(t, "string", schema.Type)
	assert.Equal(t, "date", schema.Format)
}

func TestDate_ScanStringWithLang(t *testing.T) {
	tests := []struct {
		name                 string
		source               string
		lang                 language.Code
		want                 Date
		wantNormalized       bool
		wantMonthMustBeFirst bool
		wantErr              bool
	}{
		{name: "already normalized", source: "2023-12-25", want: "2023-12-25", wantNormalized: true},
		{name: "German day first", source: "25.12.2023", lang: language.DE, want: "2023-12-25"},
		// 25 can't be a month, so the input proves month-first ordering
		// which callers use to detect the date format of a whole document.
		{name: "unambiguous US order", source: "12/25/2023", lang: language.DE, want: "2023-12-25", wantMonthMustBeFirst: true},
		{name: "ambiguous with EN hint", source: "05/06/2024", lang: language.EN, want: "2024-05-06"},
		{name: "ambiguous with DE hint", source: "05/06/2024", lang: language.DE, want: "2024-06-05"},
		{name: "ambiguous with upper case EN hint", source: "05/06/2024", lang: "EN", want: "2024-05-06"},
		// The EN hint must not reinterpret a day that can't be a month
		{name: "unambiguous day first with EN hint", source: "15.06.2024", lang: language.EN, want: "2024-06-15"},
		{name: "unambiguous US order with EN hint", source: "06/15/2024", lang: language.EN, want: "2024-06-15", wantMonthMustBeFirst: true},
		{name: "empty", source: "", wantErr: true},
		{name: "invalid", source: "invalid", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Date("2000-01-01")
			wasNormalized, monthMustBeFirst, err := d.ScanStringWithLang(tt.source, tt.lang)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, Date("2000-01-01"), d, "unchanged on error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, d)
			assert.Equal(t, tt.wantNormalized, wasNormalized, "wasNormalized")
			assert.Equal(t, tt.wantMonthMustBeFirst, monthMustBeFirst, "monthMustBeFirst")
		})
	}
}

func TestNormalize_HintOnlyDecidesAmbiguousOrder(t *testing.T) {
	// A language hint (or an ordinal suffix marking the day) can only
	// choose between dd/mm and mm/dd when both numbers are valid months.
	// If only one ordering is a valid date, that date must be returned
	// no matter the language hint, a hint must never turn a valid date
	// into an error. An ordinal suffix is no guess but part of the input,
	// a number with one is never read as the month.
	tests := []struct {
		name    string
		str     string
		lang    []language.Code
		want    Date
		wantErr bool
	}{
		{name: "EN hint, unambiguous day first", str: "15.06.2024", lang: []language.Code{language.EN}, want: "2024-06-15"},
		{name: "EN hint, unambiguous day first, short year", str: "15.06.24", lang: []language.Code{language.EN}, want: "2024-06-15"},
		{name: "EN hint, unambiguous month first", str: "06/15/2024", lang: []language.Code{language.EN}, want: "2024-06-15"},
		{name: "EN hint, ambiguous", str: "05/06/2024", lang: []language.Code{language.EN}, want: "2024-05-06"},
		{name: "DE hint, unambiguous day first", str: "15.06.2024", lang: []language.Code{language.DE}, want: "2024-06-15"},
		{name: "DE hint, unambiguous month first", str: "06/15/2024", lang: []language.Code{language.DE}, want: "2024-06-15"},
		{name: "DE hint, ambiguous", str: "05/06/2024", lang: []language.Code{language.DE}, want: "2024-06-05"},
		{name: "no hint, ambiguous", str: "05/06/2024", want: "2024-06-05"},
		// language.Code.Validate accepts these spellings of English,
		// so they must work as hint too and not be cut to " e" or "EN"
		{name: "padded EN hint, ambiguous", str: "05/06/2024", lang: []language.Code{" en"}, want: "2024-05-06"},
		{name: "BCP-47 EN hint, ambiguous", str: "05/06/2024", lang: []language.Code{"EN-us"}, want: "2024-05-06"},
		{name: "ordinal day, ambiguous", str: "02 1st 2003", want: "2003-02-01"},
		// The ordinal suffix is written into the input and marks the day
		// explicitly, while the language hint is only a guess about the
		// whole document, so the suffix must win over the hint in both orders.
		{name: "ordinal day first, ambiguous", str: "1st 02 2003", want: "2003-02-01"},
		{name: "ordinal day first, ambiguous, EN hint", str: "1st 02 2003", lang: []language.Code{language.EN}, want: "2003-02-01"},
		{name: "ordinal day second, ambiguous, EN hint", str: "02 1st 2003", lang: []language.Code{language.EN}, want: "2003-02-01"},
		{name: "ordinal day first, ambiguous, DE hint", str: "1st 02 2003", lang: []language.Code{language.DE}, want: "2003-02-01"},
		{name: "ordinal day second, ambiguous, DE hint", str: "02 1st 2003", lang: []language.Code{language.DE}, want: "2003-02-01"},
		// The ordinal suffix marks "1st" and "5th" as the day, so 13 would
		// have to be the month. Reading the ordinal as the month instead
		// contradicts the input, so there is no valid date.
		{name: "ordinal day, other number no month", str: "13 1st 2024", wantErr: true},
		{name: "ordinal day, other number no month, 5th", str: "13 5th 2020", wantErr: true},
		{name: "ordinal day first, other number no month", str: "1st 13 2024", wantErr: true},
		{name: "ordinal day, other number no month, EN hint", str: "13 1st 2024", lang: []language.Code{language.EN}, wantErr: true},
		{name: "EN hint, no valid ordering", str: "13/13/2024", lang: []language.Code{language.EN}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.str, tt.lang...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDate_NonNormalizedValidInput(t *testing.T) {
	// A valid Date that is not in the normalized YYYY-MM-DD form must
	// behave like its normalized form in every method. Year, Month, Day
	// and Compare already normalize, the methods converting to time.Time
	// must agree with them instead of silently using the zero time.Time.
	d := Date("25.12.2023") // Monday
	norm := Date("2023-12-25")

	assert.Equal(t, norm.MidnightUTC(), d.MidnightUTC(), "MidnightUTC")
	assert.Equal(t, norm.Midnight(), d.Midnight(), "Midnight")
	assert.Equal(t, norm.MidnightInLocation(time.UTC), d.MidnightInLocation(time.UTC), "MidnightInLocation")

	assert.Equal(t, "25.12.2023", d.Format("02.01.2006"), "Format")
	assert.Equal(t, "2023-12-25", d.Format(Layout), "Format(Layout)")

	assert.True(t, d.After("2023-12-24"), "After")
	assert.True(t, Date("2023-12-26").After(d), "After non-normalized other")
	assert.True(t, d.Before("2023-12-26"), "Before")
	assert.True(t, d.EqualOrAfter("2023-12-24"), "EqualOrAfter")
	assert.True(t, d.EqualOrBefore("2023-12-26"), "EqualOrBefore")
	assert.True(t, d.WithinIncl("2023-12-01", "31.12.2023"), "WithinIncl")
	assert.True(t, d.BetweenExcl("2023-12-24", "2023-12-26"), "BetweenExcl")
	assert.True(t, d.AfterTime(time.Date(2023, 12, 24, 12, 0, 0, 0, time.UTC)), "AfterTime")
	assert.True(t, d.BeforeTime(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)), "BeforeTime")

	assert.Equal(t, Date("2023-12-26"), d.AddDays(1), "AddDays")
	assert.Equal(t, Date("2024-01-25"), d.AddMonths(1), "AddMonths")
	assert.Equal(t, Date("2024-12-25"), d.AddYears(1), "AddYears")
	assert.Equal(t, Date("2024-01-26"), d.AddDate(0, 1, 1), "AddDate")
	assert.Equal(t, Date("2023-12-26"), d.Add(24*time.Hour), "Add")
	assert.Equal(t, 24*time.Hour, d.Sub("2023-12-24"), "Sub")

	assert.Equal(t, time.Monday, d.Weekday(), "Weekday")
	year, week := d.ISOWeek()
	assert.Equal(t, 2023, year, "ISOWeek year")
	assert.Equal(t, 52, week, "ISOWeek week")
	assert.Equal(t, Date("2023-12-25"), d.LastMonday(), "LastMonday")
	assert.Equal(t, Date("2023-12-31"), d.NextSunday(), "NextSunday")
	assert.Equal(t, Date("2023-12-24"), d.BeginningOfWeek(time.Sunday), "BeginningOfWeek")
	assert.Equal(t, Date("2023-12-31"), d.EndOfWeek(time.Monday), "EndOfWeek")

	// A non-normalized zero date still has no time
	assert.True(t, Date("01.01.0001").MidnightUTC().IsZero(), "zero date MidnightUTC")
	assert.True(t, Date("01.01.0001").MidnightInLocation(time.Local).IsZero(), "zero date MidnightInLocation")
}

func TestDate_ZeroDateHasNoTime(t *testing.T) {
	// A zero Date represents "no date" and must not be converted
	// into the midnight of year 1 or any other concrete point in time.
	for _, zero := range []Date{"", "0000-00-00", "0001-01-01"} {
		t.Run(string(zero), func(t *testing.T) {
			assert.True(t, zero.Time(12, 0, 0, time.UTC).IsZero(), "Time")
			assert.True(t, zero.MidnightUTC().IsZero(), "MidnightUTC")
			assert.True(t, zero.MidnightInLocation(time.UTC).IsZero(), "MidnightInLocation")
			assert.Equal(t, time.Weekday(0), zero.Weekday(), "Weekday")
			year, week := zero.ISOWeek()
			assert.Equal(t, 0, year, "ISOWeek year")
			assert.Equal(t, 0, week, "ISOWeek week")
		})
	}
}
