package strfmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type floatInfo struct {
	f          float64
	groupSep   byte
	decimalSep byte
}

func Test_ParseFloat(t *testing.T) {
	// Variations with leading + and - are created automatically, don't put them here
	var validDecimalFloats = map[string]floatInfo{
		"100":              {100, 0, 0},
		"100.9":            {100.9, 0, '.'},
		"1e6":              {1e6, 0, 0},
		"1.2e6":            {1.2e6, 0, '.'},
		",1":               {0.1, 0, ','},
		".1":               {0.1, 0, '.'},
		"1,":               {1.0, 0, ','},
		"1.":               {1.0, 0, '.'},
		"123.456":          {123.456, 0, '.'},
		"123,456":          {123.456, 0, ','},
		"100 200 300.1234": {100200300.1234, ' ', '.'},
		"100 200 300,1234": {100200300.1234, ' ', ','},
		"100,200,300.1234": {100200300.1234, ',', '.'},
		"100.200.300,1234": {100200300.1234, '.', ','},
		"1,200,300.1234":   {1200300.1234, ',', '.'},
		"1.200.300,1234":   {1200300.1234, '.', ','},
	}

	for str, ref := range validDecimalFloats {
		parsed, groupSep, decimalSep, err := ParseFloatInfo(str)
		if err != nil {
			assert.NoError(t, err)
		}
		assert.Equal(t, ref.f, parsed)
		assert.Equal(t, string(ref.groupSep), string(groupSep))
		assert.Equal(t, string(ref.decimalSep), string(decimalSep))

		parsed, groupSep, decimalSep, err = ParseFloatInfo("+" + str)
		if err != nil {
			assert.NoError(t, err)
		}
		assert.Equal(t, +ref.f, parsed)
		assert.Equal(t, string(ref.groupSep), string(groupSep))
		assert.Equal(t, string(ref.decimalSep), string(decimalSep))

		parsed, groupSep, decimalSep, err = ParseFloatInfo("-" + str)
		if err != nil {
			assert.NoError(t, err)
		}
		assert.Equal(t, -ref.f, parsed)
		assert.Equal(t, string(ref.groupSep), string(groupSep))
		assert.Equal(t, string(ref.decimalSep), string(decimalSep))
	}

	var invalidDecimalFloats = []string{
		"xxx",
		"e3",
		"1ee6",
		",1,1",
		"9,1,1",
		"10.000.00,00",
		"123.456.789",
		"123,456,789",
		"10,2340,560",
	}

	for _, s := range invalidDecimalFloats {
		_, err := ParseFloat(s)
		if err != nil {
			assert.Error(t, err)
		}
	}
}

func Test_FormatFloat(t *testing.T) {
	var formatFloatValues = map[floatInfo]string{
		{1234, 0, '.'}: "1234",
		{1234, 0, ','}: "1234",

		{1234, ',', '.'}:      "1234",
		{1234, ',', '.'}:      "1,234",
		{12345, ',', '.'}:     "12,345",
		{123456, ',', '.'}:    "123,456",
		{1234567, ',', '.'}:   "1,234,567",
		{12345678, ',', '.'}:  "12,345,678",
		{123456789, ',', '.'}: "123,456,789",

		{1234, '.', ','}:      "1234",
		{1234, '.', ','}:      "1.234",
		{12345, '.', ','}:     "12.345",
		{123456, '.', ','}:    "123.456",
		{1234567, '.', ','}:   "1.234.567",
		{12345678, '.', ','}:  "12.345.678",
		{123456789, '.', ','}: "123.456.789",

		{1234, ' ', '.'}:      "1234",
		{1234, ' ', '.'}:      "1 234",
		{12345, ' ', '.'}:     "12 345",
		{123456, ' ', '.'}:    "123 456",
		{1234567, ' ', '.'}:   "1 234 567",
		{12345678, ' ', '.'}:  "12 345 678",
		{123456789, ' ', '.'}: "123 456 789",

		{0.1234, ',', '.'}: "0.1234",
		{0.1234, 0, '.'}:   "0.1234",

		{1234.01, ',', '.'}:      "1234.01",
		{1234.01, ',', '.'}:      "1,234.01",
		{12345.01, ',', '.'}:     "12,345.01",
		{123456.01, ',', '.'}:    "123,456.01",
		{1234567.01, ',', '.'}:   "1,234,567.01",
		{12345678.01, ',', '.'}:  "12,345,678.01",
		{123456789.01, ',', '.'}: "123,456,789.01",

		{1234.01, '.', ','}:      "1234,01",
		{1234.01, '.', ','}:      "1.234,01",
		{12345.01, '.', ','}:     "12.345,01",
		{123456.01, '.', ','}:    "123.456,01",
		{1234567.01, '.', ','}:   "1.234.567,01",
		{12345678.01, '.', ','}:  "12.345.678,01",
		{123456789.01, '.', ','}: "123.456.789,01",

		{1234.01, 0, '.'}:      "1234.01",
		{1234.01, 0, '.'}:      "1234.01",
		{12345.01, 0, '.'}:     "12345.01",
		{123456.01, 0, '.'}:    "123456.01",
		{1234567.01, 0, '.'}:   "1234567.01",
		{12345678.01, 0, '.'}:  "12345678.01",
		{123456789.01, 0, '.'}: "123456789.01",

		{1234.01, 0, ','}:      "1234,01",
		{1234.01, 0, ','}:      "1234,01",
		{12345.01, 0, ','}:     "12345,01",
		{123456.01, 0, ','}:    "123456,01",
		{1234567.01, 0, ','}:   "1234567,01",
		{12345678.01, 0, ','}:  "12345678,01",
		{123456789.01, 0, ','}: "123456789,01",
	}

	for info, ref := range formatFloatValues {
		str := FormatFloat(info.f, info.groupSep, info.decimalSep, -1)
		assert.Equal(t, ref, str)

		str = FormatFloat(-info.f, info.groupSep, info.decimalSep, -1)
		assert.Equal(t, "-"+ref, str)
	}
}
