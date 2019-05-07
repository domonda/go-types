package strfmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type floatInfo struct {
	f            float64
	thousandsSep byte
	decimalSep   byte
	decimals     int
	padPrecision bool
}

func Test_ParseFloat(t *testing.T) {
	// Variations with leading + and - are created automatically, don't put them here
	validDecimalFloats := map[string]floatInfo{
		"100":              {100, 0, 0, 0, false},
		"100.9":            {100.9, 0, '.', 1, false},
		"1e6":              {1e6, 0, 0, 0, false},
		"1.2e6":            {1.2e6, 0, '.', 1, false},
		",1":               {0.1, 0, ',', 1, false},
		".1":               {0.1, 0, '.', 1, false},
		"1,":               {1.0, 0, ',', 0, false},
		"1.":               {1.0, 0, '.', 0, false},
		"123.456":          {123.456, 0, '.', 3, false},
		"123,456":          {123.456, 0, ',', 3, false},
		"100 200 300.1234": {100200300.1234, ' ', '.', 4, false},
		"100 200 300,1234": {100200300.1234, ' ', ',', 4, false},
		"100,200,300.1234": {100200300.1234, ',', '.', 4, false},
		"100.200.300,1234": {100200300.1234, '.', ',', 4, false},
		"100'200'300.1234": {100200300.1234, '\'', '.', 4, false},
		"100'200'300,1234": {100200300.1234, '\'', ',', 4, false},
		"1,200,300.1234":   {1200300.1234, ',', '.', 4, false},
		"1.200.300,1234":   {1200300.1234, '.', ',', 4, false},
		"1'200'300,1234":   {1200300.1234, '\'', ',', 4, false},
		"1.234.567":        {1234567, '.', 0, 0, false},
		"1,234,567":        {1234567, ',', 0, 0, false},
		"123.456.789":      {123456789, '.', 0, 0, false},
		"123,456,789":      {123456789, ',', 0, 0, false},
		"123 456 789":      {123456789, ' ', 0, 0, false},
		"1000000.8989":     {1000000.8989, 0, '.', 4, false},
		"1000000,8989":     {1000000.8989, 0, ',', 4, false},
		"158,00 ":          {158, 0, ',', 2, false},
	}

	testFunc := func(str string, refFloat float64, refThousandsSep, refDecimalSep byte, refDecimals int) func(*testing.T) {
		return func(t *testing.T) {
			parsed, thousandsSep, decimalSep, decimals, err := ParseFloatDetails(str)
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
	}

	for _, s := range invalidDecimalFloats {
		_, err := ParseFloat(s)
		assert.Error(t, err, "ParseFloat(%#v)", s)
	}
}

func Test_FormatFloat(t *testing.T) {
	formatFloatValues := map[floatInfo]string{
		{1234, 0, '.', -1, false}: "1234",
		{1234, 0, ',', -1, false}: "1234",

		{1234, ',', '.', -1, false}:      "1234",
		{1234, ',', '.', -1, false}:      "1,234",
		{12345, ',', '.', -1, false}:     "12,345",
		{123456, ',', '.', -1, false}:    "123,456",
		{1234567, ',', '.', -1, false}:   "1,234,567",
		{12345678, ',', '.', -1, false}:  "12,345,678",
		{123456789, ',', '.', -1, false}: "123,456,789",

		{1234, '.', ',', -1, false}:      "1234",
		{1234, '.', ',', -1, false}:      "1.234",
		{12345, '.', ',', -1, false}:     "12.345",
		{123456, '.', ',', -1, false}:    "123.456",
		{1234567, '.', ',', -1, false}:   "1.234.567",
		{12345678, '.', ',', -1, false}:  "12.345.678",
		{123456789, '.', ',', -1, false}: "123.456.789",

		{1234, ' ', '.', -1, false}:      "1234",
		{1234, ' ', '.', -1, false}:      "1 234",
		{12345, ' ', '.', -1, false}:     "12 345",
		{123456, ' ', '.', -1, false}:    "123 456",
		{1234567, ' ', '.', -1, false}:   "1 234 567",
		{12345678, ' ', '.', -1, false}:  "12 345 678",
		{123456789, ' ', '.', -1, false}: "123 456 789",

		{0.1234, ',', '.', -1, false}: "0.1234",
		{0.1234, 0, '.', -1, false}:   "0.1234",

		{1234.01, ',', '.', -1, false}:      "1234.01",
		{1234.01, ',', '.', -1, false}:      "1,234.01",
		{12345.01, ',', '.', -1, false}:     "12,345.01",
		{123456.01, ',', '.', -1, false}:    "123,456.01",
		{1234567.01, ',', '.', -1, false}:   "1,234,567.01",
		{12345678.01, ',', '.', -1, false}:  "12,345,678.01",
		{123456789.01, ',', '.', -1, false}: "123,456,789.01",

		{1234.01, '\'', '.', -1, false}:      "1234.01",
		{1234.01, '\'', '.', -1, false}:      "1'234.01",
		{12345.01, '\'', '.', -1, false}:     "12'345.01",
		{123456.01, '\'', '.', -1, false}:    "123'456.01",
		{1234567.01, '\'', '.', -1, false}:   "1'234'567.01",
		{12345678.01, '\'', '.', -1, false}:  "12'345'678.01",
		{123456789.01, '\'', '.', -1, false}: "123'456'789.01",

		{1234.01, '.', ',', -1, false}:      "1234,01",
		{1234.01, '.', ',', -1, false}:      "1.234,01",
		{12345.01, '.', ',', -1, false}:     "12.345,01",
		{123456.01, '.', ',', -1, false}:    "123.456,01",
		{1234567.01, '.', ',', -1, false}:   "1.234.567,01",
		{12345678.01, '.', ',', -1, false}:  "12.345.678,01",
		{123456789.01, '.', ',', -1, false}: "123.456.789,01",

		{1234.01, 0, '.', -1, false}:      "1234.01",
		{1234.01, 0, '.', -1, false}:      "1234.01",
		{12345.01, 0, '.', -1, false}:     "12345.01",
		{123456.01, 0, '.', -1, false}:    "123456.01",
		{1234567.01, 0, '.', -1, false}:   "1234567.01",
		{12345678.01, 0, '.', -1, false}:  "12345678.01",
		{123456789.01, 0, '.', -1, false}: "123456789.01",

		{1234.01, 0, ',', -1, false}:      "1234,01",
		{1234.01, 0, ',', -1, false}:      "1234,01",
		{12345.01, 0, ',', -1, false}:     "12345,01",
		{123456.01, 0, ',', -1, false}:    "123456,01",
		{1234567.01, 0, ',', -1, false}:   "1234567,01",
		{12345678.01, 0, ',', -1, false}:  "12345678,01",
		{123456789.01, 0, ',', -1, false}: "123456789,01",

		{1234.01, 0, '.', -1, true}: "1234.01",
		{1234.01, 0, '.', 0, true}:  "1234",
		{1234.01, 0, '.', 1, true}:  "1234.0",
		{1234.01, 0, '.', 2, true}:  "1234.01",
		{1234.01, 0, '.', 3, true}:  "1234.010",
		{1234.01, 0, '.', 4, true}:  "1234.0100",
		{1234.01, 0, '.', 5, true}:  "1234.01000",

		{1234.01, 0, ',', -1, true}: "1234,01",
		{1234.01, 0, ',', 0, true}:  "1234",
		{1234.01, 0, ',', 1, true}:  "1234,0",
		{1234.01, 0, ',', 2, true}:  "1234,01",
		{1234.01, 0, ',', 3, true}:  "1234,010",
		{1234.01, 0, ',', 4, true}:  "1234,0100",
		{1234.01, 0, ',', 5, true}:  "1234,01000",

		{1234.01, ',', '.', -1, true}: "1,234.01",
		{1234.01, ',', '.', 0, true}:  "1,234",
		{1234.01, ',', '.', 1, true}:  "1,234.0",
		{1234.01, ',', '.', 2, true}:  "1,234.01",
		{1234.01, ',', '.', 3, true}:  "1,234.010",
		{1234.01, ',', '.', 4, true}:  "1,234.0100",
		{1234.01, ',', '.', 5, true}:  "1,234.01000",

		{1234.01, '.', ',', -1, true}: "1.234,01",
		{1234.01, '.', ',', 0, true}:  "1.234",
		{1234.01, '.', ',', 1, true}:  "1.234,0",
		{1234.01, '.', ',', 2, true}:  "1.234,01",
		{1234.01, '.', ',', 3, true}:  "1.234,010",
		{1234.01, '.', ',', 4, true}:  "1.234,0100",
		{1234.01, '.', ',', 5, true}:  "1.234,01000",

		{1234.01, ' ', ',', -1, true}: "1 234,01",
		{1234.01, ' ', ',', 0, true}:  "1 234",
		{1234.01, ' ', ',', 1, true}:  "1 234,0",
		{1234.01, ' ', ',', 2, true}:  "1 234,01",
		{1234.01, ' ', ',', 3, true}:  "1 234,010",
		{1234.01, ' ', ',', 4, true}:  "1 234,0100",
		{1234.01, ' ', ',', 5, true}:  "1 234,01000",
	}

	for info, ref := range formatFloatValues {
		str := FormatFloat(info.f, info.thousandsSep, info.decimalSep, info.decimals, info.padPrecision)
		assert.Equal(t, ref, str, "FormatFloat(%#v, '%s', '%s', %#v, %#v)", info.f, string(info.thousandsSep), string(info.decimalSep), info.decimals, info.padPrecision)

		str = FormatFloat(-info.f, info.thousandsSep, info.decimalSep, info.decimals, info.padPrecision)
		assert.Equal(t, "-"+ref, str, "FormatFloat(%#v, '%s', '%s', %#v, %#v)", -info.f, string(info.thousandsSep), string(info.decimalSep), info.decimals, info.padPrecision)
	}
}

func Test_FormatFloat_invalid(t *testing.T) {
	var formatFloatValues = []floatInfo{
		{0, 0, 0, 0, false},
		{0, 'x', '.', 0, false},
		{0, 0, 'x', 0, false},
		{0, '.', '.', 0, false},
		{0, 0, '.', -2, false},
	}

	for _, info := range formatFloatValues {
		assert.Panics(t, func() {
			FormatFloat(info.f, info.thousandsSep, info.decimalSep, info.decimals, info.padPrecision)
		})
	}
}
