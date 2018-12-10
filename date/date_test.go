package date

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var dateTable = []string{
	"2006-01-02", "2006-01-02",
	"2006/01/02", "2006-01-02",
	"2006.01.02", "2006-01-02",
	"2006 01 02", "2006-01-02",

	"25.12.1975", "1975-12-25",
	"25.12.75", "1975-12-25",

	"12/25/1975", "1975-12-25",
	"12/25/75", "1975-12-25",

	"01.02.03", "2003-02-01",
	// "1.2.03", "2003-02-01", // too short
	"1.2.2003", "2003-02-01",

	"1st of 02/2003", "2003-02-01",
	"4th of 02/2003", "2003-02-04",

	"jan. 24 2012", "2012-01-24",
	"Februar 24 89", "1989-02-24",
	"February 3rd 89", "1989-02-03",

	"1. Dezember 2016", "2016-12-01",
	"1st of dec. 2020", "2020-12-01",

	"2016 Dezember 9th", "2016-12-09",
	"16 Dezember 9th", "2016-12-09",
	"16. Dezember 98", "1998-12-16",
	"16th of Dec. 04", "2004-12-16",

	"2016 25. März", "2016-03-25",
	"75 1st of march", "1975-03-01",
}

var invalidDates = []string{
	"6/12/6",
	"6/12/6,",
	"3:28:00",
}

func Test_Normalize(t *testing.T) {
	for i := 0; i < len(dateTable); i += 2 {
		normalized, err := Normalize(dateTable[i])
		if err != nil {
			t.Errorf("Normalize(%s): %s", dateTable[i], err.Error())
			continue
		}
		if string(normalized) != dateTable[i+1] {
			t.Errorf("Normalize(%s): %s != %s", dateTable[i], normalized, dateTable[i+1])
		}
	}

	for _, invalidDate := range invalidDates {
		normalized, err := Normalize(invalidDate)
		if err == nil {
			t.Errorf("Should NOT be valid Normalize(%s): %s", invalidDate, normalized)
		}
	}
}

var deFinderData = map[string][][]int{
	"3:28:00":                nil,
	"2006-01-02":             [][]int{[]int{0, 10}},
	"2006-01-02, 2017/12/03": [][]int{[]int{0, 10}, []int{12, 22}},
	"jan. 24 2012, 2017/12/03 16. Dezember 98": [][]int{[]int{0, 12}, []int{14, 24}, []int{25, 40}},
	"Datum: 25.12.1975 Dezember 1975":          [][]int{[]int{7, 17}},
}

func Test_Finder(t *testing.T) {
	deFinder := NewFinder("de")
	// enFinder := NewFinder("en")

	for str, allIndices := range deFinderData {
		allResult := deFinder.FindAllIndex([]byte(str), -1)
		if len(allResult) != len(allIndices) {
			for _, indices := range allResult {
				fmt.Println("'" + str[indices[0]:indices[1]] + "'")
			}
			t.Errorf("Found %d Dates in '%s', but expected %d", len(allResult), str, len(allIndices))
		} else {
			for i := range allIndices {
				indices := allIndices[i]
				result := allResult[i]
				if len(result) != 2 {
					t.Errorf("Did not find date in '%s'", str)
				}
				date := Date(str[result[0]:result[1]])
				if result[0] != indices[0] || result[1] != indices[1] {
					t.Errorf("Found Date '%s' at wrong position in '%s'. Expected: %v, Result: %v", date, str, indices, result)
				}
				if !date.Valid() {
					t.Errorf("Invalid Date: %s", date)
				}
			}
		}
	}
}

var periodDates = map[string][2]Date{
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
}

var invalidPeriods = []string{
	"18-01",
	"2018-00",
	"2x18-01",
	"2018-13",
	"2018-1",
	"2018 01",
	"Erik",
	"2018-Q0",
	"2018-Q5",
	"2018-Q04",
	"2018-H0",
	"2018-H3",
	"2018-H03",
}

func Test_RangeOfPeriod(t *testing.T) {
	for period, expected := range periodDates {
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
	}

	for _, period := range invalidPeriods {
		_, _, err := RangeOfPeriod(period)
		assert.Error(t, err, "RangeOfPeriod(%#v)", period)
	}
}

func Test_YearMonthDay(t *testing.T) {
	year, month, day := Date("2010-12-31").YearMonthDay()
	assert.Equal(t, 2010, year)
	assert.Equal(t, time.Month(12), month)
	assert.Equal(t, 31, day)
}
