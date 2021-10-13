package float

import (
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
	}

	testFunc := func(str string, refFloat float64, refThousandsSep, refDecimalSep rune, refDecimals int) func(*testing.T) {
		return func(t *testing.T) {
			parsed, thousandsSep, decimalSep, decimals, err := ParseDetails(str)
			assert.NoError(t, err)
			assert.Equal(t, refFloat, parsed, "ParseFloatDetails(%#v)", str)
			assert.Equal(t, string(refThousandsSep), string(thousandsSep), "ParseFloatDetails(%#v)", str)
			assert.Equal(t, string(refDecimalSep), string(decimalSep), "ParseFloatDetails(%#v)", str)
			assert.Equal(t, refDecimals, decimals, "ParseFloatDetails(%#v)", str)
		}
	}

	for str, ref := range validDecimalFloats {
		t.Run("no sign", testFunc(str, ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))

		t.Run("plus in front", testFunc("+"+str, ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		t.Run("plus in front with space", testFunc("+ "+str, ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		t.Run("plus on end", testFunc(str+"+", ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		t.Run("plus on end with space", testFunc(str+" +", ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))

		t.Run("minus in front", testFunc("-"+str, -ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		t.Run("minus in front with space", testFunc("- "+str, -ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		t.Run("minus on end", testFunc(str+"-", -ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
		t.Run("minus on end with space", testFunc(str+" -", -ref.f, ref.thousandsSep, ref.decimalSep, ref.decimals))
	}
}

func Test_ParseFloat_invalid(t *testing.T) {
	invalidDecimalFloats := []string{
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
	}

	for _, s := range invalidDecimalFloats {
		_, err := Parse(s)
		assert.Error(t, err, "ParseFloat(%#v)", s)
	}
}
