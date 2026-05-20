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
