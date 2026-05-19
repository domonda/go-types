package float

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type floatInfo struct {
	f            float64
	thousandsSep rune
	decimalSep   rune
	decimals     int
	padPrecision bool
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
		str := Format(info.f, info.thousandsSep, info.decimalSep, info.decimals, info.padPrecision)
		assert.Equal(t, ref, str, "FormatFloat(%#v, '%s', '%s', %#v, %#v)", info.f, string(info.thousandsSep), string(info.decimalSep), info.decimals, info.padPrecision)

		str = Format(-info.f, info.thousandsSep, info.decimalSep, info.decimals, info.padPrecision)
		assert.Equal(t, "-"+ref, str, "FormatFloat(%#v, '%s', '%s', %#v, %#v)", -info.f, string(info.thousandsSep), string(info.decimalSep), info.decimals, info.padPrecision)
	}
}

func Benchmark_FormatFloat_commaDecimalSep(b *testing.B) {
	// Hot path for DE/AT-style formatting: no thousands grouping,
	// comma decimal separator, common precision values.
	cases := []struct {
		name string
		f    float64
		prec int
		pad  bool
	}{
		{"small_noPad", 12.34, -1, false},
		{"small_pad2", 12.3, 2, true},
		{"small_pad5", 12.345, 5, true},
		{"neg_pad4", -987.6, 4, true},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = Format(c.f, 0, ',', c.prec, c.pad)
			}
		})
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
			Format(info.f, info.thousandsSep, info.decimalSep, info.decimals, info.padPrecision)
		})
	}
}
