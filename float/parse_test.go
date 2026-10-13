package float

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseFloat(t *testing.T) {
	// Variations with leading + and - are created automatically, don't put them here
	validDecimalFloats := map[string]floatInfo{
		"100":                  {100, 0, 0, 0, false},
		"100.9":                {100.9, 0, '.', 1, false},
		"1e6":                  {1e6, 0, 0, 0, false},
		"1.2e6":                {1.2e6, 0, '.', 1, false},
		"1e-6":                 {1e-6, 0, 0, 0, false},
		"1.2e-6":               {1.2e-6, 0, '.', 1, false},
		"1e+6":                 {1e6, 0, 0, 0, false},
		"1.2e+6":               {1.2e+6, 0, '.', 1, false},
		"2.48689957516035e14":  {2.48689957516035e14, 0, '.', 14, false},
		"2.48689957516035E14":  {2.48689957516035e14, 0, '.', 14, false},
		"2.48689957516035e-14": {2.48689957516035e-14, 0, '.', 14, false},
		"2.48689957516035E-14": {2.48689957516035e-14, 0, '.', 14, false},
		"2.48689957516035e+14": {2.48689957516035e+14, 0, '.', 14, false},
		"2.48689957516035E+14": {2.48689957516035e+14, 0, '.', 14, false},
		",1":                   {0.1, 0, ',', 1, false},
		".1":                   {0.1, 0, '.', 1, false},
		"1,":                   {1.0, 0, ',', 0, false},
		"1.":                   {1.0, 0, '.', 0, false},
		"123.456":              {123.456, 0, '.', 3, false},
		"123,456":              {123.456, 0, ',', 3, false},
		"100 200 300.1234":     {100200300.1234, ' ', '.', 4, false},
		"100 200 300,1234":     {100200300.1234, ' ', ',', 4, false},
		"100,200,300.1234":     {100200300.1234, ',', '.', 4, false},
		"100.200.300,1234":     {100200300.1234, '.', ',', 4, false},
		"100'200'300.1234":     {100200300.1234, '\'', '.', 4, false},
		"100'200'300,1234":     {100200300.1234, '\'', ',', 4, false},
		"1,200,300.1234":       {1200300.1234, ',', '.', 4, false},
		"1.200.300,1234":       {1200300.1234, '.', ',', 4, false},
		"1'200'300,1234":       {1200300.1234, '\'', ',', 4, false},
		"1.234.567":            {1234567, '.', 0, 0, false},
		"1,234,567":            {1234567, ',', 0, 0, false},
		"123.456.789":          {123456789, '.', 0, 0, false},
		"123,456,789":          {123456789, ',', 0, 0, false},
		"123 456 789":          {123456789, ' ', 0, 0, false},
		"1000000.8989":         {1000000.8989, 0, '.', 4, false},
		"1000000,8989":         {1000000.8989, 0, ',', 4, false},
		"158,00 ":              {158, 0, ',', 2, false},
		"NaN":                  {math.NaN(), 0, 0, 0, false},  // No sign prepending in test
		"Inf":                  {math.Inf(1), 0, 0, 0, false}, // +Inf and -Inf will generated in test by prepending sign
	}

	testFunc := func(str string, refFloat float64, refThousandsSep, refDecimalSep rune, refDecimals int) func(*testing.T) {
		return func(t *testing.T) {
			parsed, thousandsSep, decimalSep, decimals, err := ParseDetails(str)
			assert.NoError(t, err)
			if math.IsNaN(refFloat) {
				assert.True(t, math.IsNaN(parsed), "ParseFloatDetails(%#v)", str)
			} else {
				assert.Equal(t, refFloat, parsed, "ParseFloatDetails(%#v)", str)
			}
			assert.Equal(t, string(refThousandsSep), string(thousandsSep), "ParseFloatDetails(%#v)", str)
			assert.Equal(t, string(refDecimalSep), string(decimalSep), "ParseFloatDetails(%#v)", str)
			assert.Equal(t, refDecimals, decimals, "ParseFloatDetails(%#v)", str)
		}
	}

	for str, ref := range validDecimalFloats {
		t.Run("no sign", testFunc(str, ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		if str == "NaN" {
			continue
		}
		t.Run("plus in front", testFunc("+"+str, ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		if str != "Inf" {
			t.Run("plus in front with space", testFunc("+ "+str, ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
			t.Run("plus on end", testFunc(str+"+", ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
			t.Run("plus on end with space", testFunc(str+" +", ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		}
		t.Run("minus in front", testFunc("-"+str, -ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		if str != "Inf" {
			t.Run("minus in front with space", testFunc("- "+str, -ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
			t.Run("minus on end", testFunc(str+"-", -ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
			t.Run("minus on end with space", testFunc(str+" -", -ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		}
	}
}

// Test_ParseFloat_invalid protects that malformed numbers are errors
// instead of a guessed value. A separator in the exponent must be an
// error, because dropping it like a thousands separator turns "1e3,5"
// into 1e35, an amount 32 orders of magnitude off.
func Test_ParseFloat_invalid(t *testing.T) {
	invalidDecimalFloats := []string{
		"",
		"xxx",
		"e3",
		"--1",
		"1--",
		"++1",
		"1++",
		"-+1",
		"1-+",
		"+-1",
		"1+-",
		"1-2",
		"1+2",
		"1/2",
		"1ee6",
		"123+456",
		"12-3456",
		",1,1",
		"9,1,1",
		"10.2340 560",
		"123.12.1.0",
		"10.000.00,00",
		"10,2340,560",
		"10.2340,560",
		"10.23,560",
		"1.234.56",
		"1,234,56",
		"123-456",
		"123.45.67.890",
		"123.45.67.890,0",
		"1234.567.890,0",
		"-1234.567.890,0",
		"123,456.789,000",
		"123,456.789 000",
		"123,123 123 123",
		"123,123 23 123",
		"123,1234,123.99",
		"123 1234 123.99",
		"1234567,123,456", // first grouped block longer than 6 digits
		"1234567 123 456", // first grouped block longer than 6 digits (space)
		"1e",              // dangling exponent without digits
		"1e999",           // out of float64 range must error, not silently become +Inf
		"1e3,5",           // separator in the exponent, not 1e35
		"1e3.5",           // separator in the exponent, not 1e35
		"1e3'5",           // separator in the exponent, not 1e35
		"1e3 5",           // separator in the exponent, not 1e35
	}

	for _, s := range invalidDecimalFloats {
		_, err := Parse(s)
		assert.Error(t, err, "ParseFloat(%#v)", s)
	}
}

// Test_Parse exercises the Parse wrapper, which returns only the float value.
func Test_Parse(t *testing.T) {
	f, err := Parse("1.234.567,89")
	assert.NoError(t, err)
	assert.Equal(t, 1234567.89, f)

	f, err = Parse(" -158,00 ")
	assert.NoError(t, err)
	assert.Equal(t, -158.0, f)

	_, err = Parse("not a number")
	assert.Error(t, err)
}

// Test_ParseFloat_decimals locks the decimals return value of ParseDetails,
// which reports how many fractional digits the input carried.
func Test_ParseFloat_decimals(t *testing.T) {
	tests := []struct {
		str          string
		wantDecimals int
	}{
		{"100", 0},
		{"100.", 0},
		{"100.9", 1},
		{"100.90", 2},
		{"123.456", 3},
		{"1,234.5678", 4},
		{"1.2e6", 1},
		{"2.48689957516035e14", 14},
	}
	for _, tt := range tests {
		_, _, _, decimals, err := ParseDetails(tt.str)
		assert.NoError(t, err, "ParseDetails(%q)", tt.str)
		assert.Equal(t, tt.wantDecimals, decimals, "ParseDetails(%q) decimals", tt.str)
	}
}

// Test_ParseDetails_singleSeparatorAmbiguity locks the current interpretation
// of a number with exactly one separator and a 1-3 digit trailing group.
// "1,234" is ambiguous: it could be the integer 1234 (US grouping) or the
// decimal 1.234. ParseDetails resolves a lone separator as the decimal
// separator. This is asymmetric with Format, which writes
// Format(1234, ',', '.', -1, false) == "1,234" — so that specific integer
// does not round-trip. See Test_FormatParse_roundTrip for the cases that do.
func Test_ParseDetails_singleSeparatorAmbiguity(t *testing.T) {
	tests := []struct {
		str        string
		wantFloat  float64
		wantDecSep rune
	}{
		{"1,234", 1.234, ','},
		{"1.234", 1.234, '.'},
		{"12,345", 12.345, ','},
		{"123,456", 123.456, ','},
		{"123.456", 123.456, '.'},
	}
	for _, tt := range tests {
		f, thousandsSep, decimalSep, _, err := ParseDetails(tt.str)
		assert.NoError(t, err, "ParseDetails(%q)", tt.str)
		assert.Equal(t, tt.wantFloat, f, "ParseDetails(%q) value", tt.str)
		assert.Equal(t, rune(0), thousandsSep, "ParseDetails(%q) thousandsSep", tt.str)
		assert.Equal(t, string(tt.wantDecSep), string(decimalSep), "ParseDetails(%q) decimalSep", tt.str)
	}
}

// Test_ParseDetails_apostropheAndSpaceOnlyGroupThousands protects the rule
// that the Swiss apostrophe and the SI/French space are only ever thousands
// separators, unlike '.' and ',' which are the decimal separator when they
// occur once. "1'234" is 1234 in Switzerland and "1 234" is 1234 in France,
// reading them as 1.234 silently divides an amount by a thousand.
// An apostrophe or a space that can't group thousands, like the one of
// "1'5", is an error instead of a guess at a decimal separator that no
// locale uses.
func Test_ParseDetails_apostropheAndSpaceOnlyGroupThousands(t *testing.T) {
	valid := []struct {
		str              string
		want             float64
		wantThousandsSep rune
		wantDecimalSep   rune
		wantDecimals     int
	}{
		{"1'234", 1234, '\'', 0, 0},
		{"1 234", 1234, ' ', 0, 0},
		{"-12'345", -12345, '\'', 0, 0},
		{"123 456 -", -123456, ' ', 0, 0},
		{"1'234.50", 1234.5, '\'', '.', 2},
		{"1'234,5", 1234.5, '\'', ',', 1},
		{"1 234,56", 1234.56, ' ', ',', 2},
		{"1'234e3", 1234e3, '\'', 0, 0},
	}
	for _, tt := range valid {
		f, thousandsSep, decimalSep, decimals, err := ParseDetails(tt.str)
		if assert.NoError(t, err, "ParseDetails(%q)", tt.str) {
			assert.Equal(t, tt.want, f, "ParseDetails(%q) value", tt.str)
			assert.Equal(t, string(tt.wantThousandsSep), string(thousandsSep), "ParseDetails(%q) thousandsSep", tt.str)
			assert.Equal(t, string(tt.wantDecimalSep), string(decimalSep), "ParseDetails(%q) decimalSep", tt.str)
			assert.Equal(t, tt.wantDecimals, decimals, "ParseDetails(%q) decimals", tt.str)
		}
	}

	invalid := []string{
		"1'5",
		"1 5",
		"1'23",
		"1 23",
		"1'2345",
		"1 2345",
		"1'",
		"'234",
		"1'234'5",
		"1.234'5",   // apostrophe after a thousands separator would be the decimal one
		"1,234'567", // apostrophe after a thousands separator would be the decimal one
		"1'2e3",
	}
	for _, str := range invalid {
		f, _, _, _, err := ParseDetails(str)
		assert.Error(t, err, "ParseDetails(%q) = %v", str, f)
	}
}

// Test_ParseDetails_firstThousandsGroup protects the digit grouping rule
// that the group before the first thousands separator has 1 to 3 digits.
// "1234.567,89" is no valid grouping in any locale and accepting it as
// 1234567.89 turns a typo or a misread column into a plausible amount.
// The rule must not depend on whether the thousands separator is followed
// by another one, which "1234.567.890,0" always was rejected for,
// or by the decimal separator.
func Test_ParseDetails_firstThousandsGroup(t *testing.T) {
	valid := map[string]float64{
		"1.567,89":     1567.89,
		"12.567,89":    12567.89,
		"123.567,89":   123567.89,
		"123,567.89":   123567.89,
		"-1'567.89":    -1567.89,
		"123 567,89":   123567.89,
		"1.234.567,89": 1234567.89,
	}
	for str, want := range valid {
		f, err := Parse(str)
		if assert.NoError(t, err, "Parse(%q)", str) {
			assert.Equal(t, want, f, "Parse(%q)", str)
		}
	}

	invalid := []string{
		"1234.567,89",
		"1234,567.89",
		"-1234.567,89",
		"1234.567,89-",
		"1234'567.89",
		"1234 567,89",
		"12345'678",
		"12345 678",
		".234,5",
		",234.5",
		"'234.5",
	}
	for _, str := range invalid {
		f, err := Parse(str)
		assert.Error(t, err, "Parse(%q) = %v", str, f)
	}
}

// Test_ParseDetails_exponentAfterThousands protects that a thousands
// grouped number keeps all of its integer digits in front of an exponent.
// Before, the last group was taken as fractional digits,
// so "1,234,567e3" parsed as 1234.567e3 instead of 1234567e3.
func Test_ParseDetails_exponentAfterThousands(t *testing.T) {
	tests := []struct {
		str              string
		want             float64
		wantThousandsSep rune
	}{
		{"1,234,567e3", 1234567e3, ','},
		{"1.234.567e-3", 1234567e-3, '.'},
		{"1 234 567E3", 1234567e3, ' '},
		{"1'234e3", 1234e3, '\''},
	}
	for _, tt := range tests {
		f, thousandsSep, decimalSep, decimals, err := ParseDetails(tt.str)
		if assert.NoError(t, err, "ParseDetails(%q)", tt.str) {
			assert.Equal(t, tt.want, f, "ParseDetails(%q) value", tt.str)
			assert.Equal(t, string(tt.wantThousandsSep), string(thousandsSep), "ParseDetails(%q) thousandsSep", tt.str)
			assert.Equal(t, rune(0), decimalSep, "ParseDetails(%q) decimalSep", tt.str)
			assert.Equal(t, 0, decimals, "ParseDetails(%q) decimals", tt.str)
		}
	}

	// A single '.' or ',' in front of the exponent stays the decimal separator,
	// and ParseDetails reports that rune, not always '.', so that callers
	// can format the number back in the style it was read in.
	for _, sep := range []rune{',', '.'} {
		str := "1" + string(sep) + "5e3"
		f, thousandsSep, decimalSep, decimals, err := ParseDetails(str)
		if assert.NoError(t, err, "ParseDetails(%q)", str) {
			assert.Equal(t, 1500.0, f, "ParseDetails(%q) value", str)
			assert.Equal(t, rune(0), thousandsSep, "ParseDetails(%q) thousandsSep", str)
			assert.Equal(t, string(sep), string(decimalSep), "ParseDetails(%q) decimalSep", str)
			assert.Equal(t, 1, decimals, "ParseDetails(%q) decimals", str)
		}
	}
	f, err := Parse("1,5e3")
	assert.NoError(t, err)
	assert.Equal(t, 1500.0, f)
	// A last group of other than 3 digits can't be grouped thousands
	_, err = Parse("1,234,56e3")
	assert.Error(t, err)
}

// Test_ParseDetails_separatorsFormatRoundTrip protects that the separators
// ParseDetails detects can be passed back into Format, which panics for a
// separator it doesn't support, to write a parsed number back in the style
// it was read in. A zero decimalSep means the input had no decimal
// separator, so any decimal separator that differs from thousandsSep can be
// passed instead. The value must survive the round-trip.
func Test_ParseDetails_separatorsFormatRoundTrip(t *testing.T) {
	strs := []string{
		"0",
		"-1",
		"1.5",
		"1,5",
		",5",
		"1.",
		"1,234",
		"1.234",
		"1 234",
		"1'234",
		"12 345 678",
		"12'345'678",
		"1,234,567",
		"1.234.567",
		"1,234.5",
		"1.234,5",
		"1 234.5",
		"1 234,5",
		"1'234.5",
		"1'234,5",
		"1.234.567,0",
		"-1'234'567.891",
		"1.2e6",
		"1'234e3",
		"1000000.8989",
	}
	for _, str := range strs {
		f, thousandsSep, decimalSep, decimals, err := ParseDetails(str)
		if !assert.NoError(t, err, "ParseDetails(%q)", str) {
			continue
		}
		assert.Contains(t, []rune{0, '.', ',', ' ', '\''}, thousandsSep, "ParseDetails(%q) thousandsSep", str)
		assert.Contains(t, []rune{0, '.', ','}, decimalSep, "ParseDetails(%q) decimalSep", str)
		if decimalSep == 0 {
			decimalSep = '.'
			if thousandsSep == '.' {
				decimalSep = ','
			}
		}
		var formatted string
		if !assert.NotPanics(t, func() { formatted = Format(f, thousandsSep, decimalSep, decimals, true) }, "Format(ParseDetails(%q))", str) {
			continue
		}
		parsed, err := Parse(formatted)
		if assert.NoError(t, err, "Parse(%q) of Format(ParseDetails(%q))", formatted, str) {
			assert.Equal(t, f, parsed, "Parse(%q) of Format(ParseDetails(%q))", formatted, str)
		}
	}
}

// Test_ParseDetails_oneSign protects the documented rule that a number has
// at most one sign, either as its first or as its last character.
// A sign at both ends is a malformed or misread value: taking the trailing
// one, like "+1.5-" as -1.5, or the leading one, like "-1+" as -1,
// silently picks one of two contradicting signs, and an agreeing doubled
// sign like "+1+" or "-1-" is no number format of any locale either.
func Test_ParseDetails_oneSign(t *testing.T) {
	invalid := []string{
		"+1.5-",
		"-1.5+",
		"-1+",
		"+1-",
		"- 1 +",
		"+ 1.234,56 -",
		"+1e5-",
		"-1.5e-3+",
		"+1+",
		"-1-",
		"-1.5-",
		"+ 1 +",
	}
	for _, str := range invalid {
		f, _, _, _, err := ParseDetails(str)
		assert.Error(t, err, "ParseDetails(%q) = %v", str, f)
	}
}
