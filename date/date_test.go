package date

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

		// TODO
		// "2 janv. 2019",  // french
		// "30 janv. 2019", // french
		// "23/gen/2019",   // italian

		// Test data from https://raw.githubusercontent.com/araddon/dateparse/master/parseany_test.go
		// "oct 7, 1970":   "1970-10-07", // TODO
		// "oct 7, '70":    "1970-10-07", // TODO
		// "Oct 7, '70":    "1970-10-07", // TODO
		// "Oct. 7, '70":   "1970-10-07", // TODO
		// "oct. 7, '70":   "1970-10-07", // TODO
		// "oct. 7, 1970": "1970-10-07", // TODO
		// "Sept. 7, '70":  "1970-09-07", // TODO
		// "sept. 7, 1970": "1970-09-07", // TODO
		// "Feb 8, 2009":      "2009-02-08", // TODO
		"7 oct 70":         "1970-10-07",
		"7 oct 1970":       "1970-10-07",
		"7 May 1970":       "1970-05-07",
		"7 Sep 1970":       "1970-09-07",
		"7 June 1970":      "1970-06-07",
		"7 September 1970": "1970-09-07",
		// RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
		// "Mon Jan 02 2006": "2006-01-02", // TODO
		// "Thu May 08 2009": "2009-05-08", // TODO
		// Month dd, yyyy at time
		// "September 17, 2012": "2012-09-17", // TODO
		// "May 17, 2012":       "2012-05-17", // TODO
		// Month dd yyyy time
		// "September 17 2012": "2012-09-17", // TODO
		// Month dd, yyyy
		// "May 7, 2012":  "2012-05-07", // TODO
		// "June 7, 2012": "2012-06-07", // TODO
		// "June 7 2012":  "2012-06-07", // TODO
		// Month dd[th,nd,st,rd] yyyy
		// "September 17th, 2012": "2012-09-17", // TODO
		// "September 17th 2012":  "2012-09-17", // TODO
		// "September 7th, 2012":  "2012-09-07", // TODO
		// "September 7th 2012":   "2012-09-07", // TODO
		// "May 1st 2012":         "2012-05-01", // TODO
		// "May 1st, 2012":        "2012-05-01", // TODO
		// "May 21st 2012":        "2012-05-21", // TODO
		// "May 21st, 2012":       "2012-05-21", // TODO
		// "May 23rd 2012":        "2012-05-23", // TODO
		// "May 23rd, 2012":       "2012-05-23", // TODO
		// "June 2nd, 2012":       "2012-06-02", // TODO
		// "June 2nd 2012":        "2012-06-02", // TODO
		// "June 22nd, 2012":      "2012-06-22", // TODO
		// "June 22nd 2012":       "2012-06-22", // TODO
		// ?
		// "Fri, 03 Jul 2015": "2015-07-03", // TODO
		// "Fri, 3 Jul 2015":  "2015-07-03", // TODO
		// "Thu, 03 Jul 2017": "2017-07-03", // TODO
		// "Thu, 3 Jul 2017":  "2017-07-03", // TODO
		// "Tue, 11 Jul 2017": "2017-07-11", // TODO
		// "Tue, 5 Jul 2017":  "2017-07-05", // TODO
		// "Fri, 03-Jul-15":   "2015-07-03", // TODO
		// "Fri, 03-Jul 2015": "2015-07-03", // TODO
		// "Fri, 3-Jul-15":    "2015-07-03", // TODO
		// RFC850    = "Monday, 02-Jan-06 15:04:05 MST"
		// "Wednesday, 07-May-09": "2009-05-07", // TODO
		// "Wednesday, 28-Feb-18": "2018-02-28", // TODO
		// with offset then with variations on non-zero filled stuff
		// "Monday, 02 Jan 2006":    "2006-01-02", // TODO
		// "Wednesday, 28 Feb 2018": "2018-02-28", // TODO
		// "Wednesday, 2 Feb 2018":  "2018-02-02", // TODO
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
		// Chinese 2014年04月18日
		// "2014年04月08日": "2014-04-08", // TODO
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
	}

	for input, expected := range dateTable {
		t.Run(fmt.Sprintf("Normalize(%s)", input), func(t *testing.T) {
			normalized, err := Normalize(input)
			assert.NoError(t, err, "Normalize")
			assert.Equal(t, expected, normalized, "Normalize")
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
			assert.Error(t, err, "Should NOT be valid Normalize(%#v): %#v", invalidDate, normalized)
		})
	}
}

func Test_Finder(t *testing.T) {

	deFinderData := map[string][][]int{
		"3:28:00":                nil,
		"2006-01-02":             [][]int{[]int{0, 10}},
		"2006-01-02, 2017/12/03": [][]int{[]int{0, 10}, []int{12, 22}},
		"jan. 24 2012, 2017/12/03 16. Dezember 98": [][]int{[]int{0, 12}, []int{14, 24}, []int{25, 40}},
		"Datum: 25.12.1975 Dezember 1975":          [][]int{[]int{7, 17}},
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

func Test_RangeOfPeriod(t *testing.T) {
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
	}

	for period, expected := range periodDates {
		t.Run(fmt.Sprintf("RangeOfPeriod(%#v)", period), func(t *testing.T) {
			from, until, err := RangeOfPeriod(period)
			if err != nil {
				t.Fatal(err)
			}
			if from != expected[0] {
				t.Errorf("RangeOfPeriod(%#v) expected from to be %#v but got %#v", period, expected[0], from)
			}
			if until != expected[1] {
				t.Errorf("RangeOfPeriod(%#v) expected until to be %#v but got %#v", period, expected[1], until)
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
		"2018-H03",
		"2018-W0",
		"2018-W00",
		"2018-W54",
	}

	for _, period := range invalidPeriods {
		t.Run(fmt.Sprintf("RangeOfPeriod(%s)", period), func(t *testing.T) {
			_, _, err := RangeOfPeriod(period)
			assert.Error(t, err, "RangeOfPeriod(%#v)", period)
		})
	}

}
func Test_RangeOfYear(t *testing.T) {
	periodDates := map[int][2]Date{
		-333: {"-333-01-01", "-333-12-31"},
		0:    {"0000-01-01", "0000-12-31"},
		325:  {"0325-01-01", "0325-12-31"},
		2018: {"2018-01-01", "2018-12-31"},
	}

	for year, expected := range periodDates {
		t.Run(fmt.Sprintf("RangeOfYear(%#v)", year), func(t *testing.T) {
			from, until := RangeOfYear(year)
			if from != expected[0] {
				t.Errorf("RangeOfYear(%#v) expected from to be %#v but got %#v", year, expected[0], from)
			}
			if until != expected[1] {
				t.Errorf("RangeOfYear(%#v) expected until to be %#v but got %#v", year, expected[1], until)
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
		"2010-12-31": {2010, 12, 31},
		"2000-01-01": {2000, 1, 1},
	}

	for date, expected := range dates {
		t.Run(fmt.Sprintf("Date(%s).YearMonthDay()", date), func(t *testing.T) {
			year, month, day := date.YearMonthDay()
			assert.Equal(t, expected.year, year)
			assert.Equal(t, time.Month(expected.month), month)
			assert.Equal(t, expected.day, day)
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
	assert.NoError(t, err, "json.Unmarshal")
	assert.Equal(t, Date(""), s.Empty, "empty JSON string is Null")
	assert.Equal(t, Date(""), s.Null, "JSON null value as Null")
	assert.Equal(t, Date("2012-12-12"), s.NotNull, "valid Date")
	assert.Equal(t, Date("Not a date!"), s.Invalid, "invalid Date parsed as is, without error")
	assert.False(t, s.Invalid.Valid(), "invalid Date parsed as is, not valid")
}
