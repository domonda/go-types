package types

import "testing"

var numberTable = map[string]Number{
	"3":   3,
	"666": 666,

	"123.4":        123.4,
	"123.45":       123.45,
	"1.2345":       1.2345,
	"1000000.8989": 1000000.8989,

	"123,4":        123.4,
	"123,45":       123.45,
	"1,2345":       1.2345,
	"1000000,8989": 1000000.8989,

	"1,234.56":       1234.56,
	"1,234,567.89":   1234567.89,
	"10,234,567.89":  10234567.89,
	"100,234,567.89": 100234567.89,
	"1,234,567":      1234567,

	"1.234,56":       1234.56,
	"1.234.567,89":   1234567.89,
	"10.234.567,89":  10234567.89,
	"100.234.567,89": 100234567.89,
	"1.234.567":      1234567,

	"-3":   -3,
	"-666": -666,

	"-123.4":        -123.4,
	"-123.45":       -123.45,
	"-1.2345":       -1.2345,
	"-1000000.8989": -1000000.8989,

	"-123,4":        -123.4,
	"-123,45":       -123.45,
	"-1,2345":       -1.2345,
	"-1000000,8989": -1000000.8989,

	"-1,234.56":       -1234.56,
	"-1,234,567.89":   -1234567.89,
	"-10,234,567.89":  -10234567.89,
	"-100,234,567.89": -100234567.89,
	"-1,234,567":      -1234567,

	"-1.234,56":       -1234.56,
	"-1.234.567,89":   -1234567.89,
	"-10.234.567,89":  -10234567.89,
	"-100.234.567,89": -100234567.89,
	"-1.234.567":      -1234567,
}

var invalidNumbers = []string{
	".5",
	",5",
	"5.",
	"5,",
	"1,234,56",
	"1000,234,560",
	"10,2340,560",
	"10.2340,560",
}

func Test_ParseNumber(t *testing.T) {
	for str, refNumber := range numberTable {
		number, err := ParseNumber(str)
		if err != nil {
			t.Errorf("Could not parse number %s because of error: '%s'", str, err)
		}
		if number != refNumber {
			t.Errorf("Parsed '%s' number %f != %f", str, number, refNumber)
		}
	}
	for _, str := range invalidNumbers {
		number, err := ParseNumber(str)
		if err == nil {
			t.Errorf("Parsed invalid number '%s' as %f", str, number)
		}
	}
}

func Test_StringIsNumber(t *testing.T) {
	for str := range numberTable {
		if !StringIsNumber(str) {
			t.Errorf("String not detected as number: '%s'", str)
		}
	}
	for _, str := range invalidNumbers {
		if StringIsNumber(str) {
			t.Errorf("Invalid string detected as number: '%s'", str)
		}
	}
}
