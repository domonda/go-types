package float

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_FormatParse_roundTrip verifies that Parse recovers the exact float64
// value produced by Format when the formatted string is unambiguous.
//
// Precision -1 is used so Format emits the shortest string that ParseFloat
// reads back exactly. Pure integers in [1000, 1000000) formatted with a
// thousands separator are intentionally excluded: they produce a single
// separator (e.g. "1,234") that Parse reads as a decimal separator. See
// Test_ParseDetails_singleSeparatorAmbiguity for that case.
func Test_FormatParse_roundTrip(t *testing.T) {
	values := []float64{
		0,
		1,
		42,
		999,
		0.5,
		0.1234,
		1234.56,
		12345.67,
		123456.78,
		1234567.89,
		1234567,
		12345678,
		123456789,
		1000000,
		1000000.5,
	}
	seps := []struct {
		thousandsSep rune
		decimalSep   rune
	}{
		{0, '.'},
		{0, ','},
		{',', '.'},
		{'.', ','},
		{'\'', '.'},
		{' ', '.'},
		{' ', ','},
	}
	for _, v := range values {
		for _, s := range seps {
			formatted := Format(v, s.thousandsSep, s.decimalSep, -1, false)
			parsed, err := Parse(formatted)
			if assert.NoError(t, err, "Parse(%q) from Format(%v)", formatted, v) {
				assert.Equal(t, v, parsed, "round-trip %v via %q", v, formatted)
			}

			if v == 0 {
				continue // -0 is a separate edge case, not covered here
			}
			formattedNeg := Format(-v, s.thousandsSep, s.decimalSep, -1, false)
			parsedNeg, err := Parse(formattedNeg)
			if assert.NoError(t, err, "Parse(%q) from Format(%v)", formattedNeg, -v) {
				assert.Equal(t, -v, parsedNeg, "round-trip %v via %q", -v, formattedNeg)
			}
		}
	}
}

// Test_ParseFormat_roundTrip verifies the other direction: a string Parsed
// into a value and detected separators, then re-Formatted with those same
// separators, yields the original string.
func Test_ParseFormat_roundTrip(t *testing.T) {
	strs := []string{
		"1234.56",
		"1234,56",
		"1,234,567.89",
		"1.234.567,89",
		"1'234'567.89",
		"1 234 567,89",
		"123",
		"0.5",
		"-1,234,567.89",
		"100200300.1234",
	}
	for _, str := range strs {
		f, thousandsSep, decimalSep, decimals, err := ParseDetails(str)
		if !assert.NoError(t, err, "ParseDetails(%q)", str) {
			continue
		}
		// Re-format with the detected separators and decimal count.
		// decimalSep falls back to '.' when the input had no fraction.
		ds := decimalSep
		if ds == 0 {
			ds = '.'
		}
		reformatted := Format(f, thousandsSep, ds, decimals, false)
		assert.Equal(t, str, reformatted, "ParseDetails+Format round-trip of %q", str)
	}
}
