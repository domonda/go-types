package float

import (
	"math"
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

		{1234, ',', '.', -1, false}:      "1,234",
		{12345, ',', '.', -1, false}:     "12,345",
		{123456, ',', '.', -1, false}:    "123,456",
		{1234567, ',', '.', -1, false}:   "1,234,567",
		{12345678, ',', '.', -1, false}:  "12,345,678",
		{123456789, ',', '.', -1, false}: "123,456,789",

		{1234, '.', ',', -1, false}:      "1.234",
		{12345, '.', ',', -1, false}:     "12.345",
		{123456, '.', ',', -1, false}:    "123.456",
		{1234567, '.', ',', -1, false}:   "1.234.567",
		{12345678, '.', ',', -1, false}:  "12.345.678",
		{123456789, '.', ',', -1, false}: "123.456.789",

		{1234, ' ', '.', -1, false}:      "1 234",
		{12345, ' ', '.', -1, false}:     "12 345",
		{123456, ' ', '.', -1, false}:    "123 456",
		{1234567, ' ', '.', -1, false}:   "1 234 567",
		{12345678, ' ', '.', -1, false}:  "12 345 678",
		{123456789, ' ', '.', -1, false}: "123 456 789",

		{0.1234, ',', '.', -1, false}: "0.1234",
		{0.1234, 0, '.', -1, false}:   "0.1234",

		{1234.01, ',', '.', -1, false}:      "1,234.01",
		{12345.01, ',', '.', -1, false}:     "12,345.01",
		{123456.01, ',', '.', -1, false}:    "123,456.01",
		{1234567.01, ',', '.', -1, false}:   "1,234,567.01",
		{12345678.01, ',', '.', -1, false}:  "12,345,678.01",
		{123456789.01, ',', '.', -1, false}: "123,456,789.01",

		{1234.01, '\'', '.', -1, false}:      "1'234.01",
		{12345.01, '\'', '.', -1, false}:     "12'345.01",
		{123456.01, '\'', '.', -1, false}:    "123'456.01",
		{1234567.01, '\'', '.', -1, false}:   "1'234'567.01",
		{12345678.01, '\'', '.', -1, false}:  "12'345'678.01",
		{123456789.01, '\'', '.', -1, false}: "123'456'789.01",

		{1234.01, '.', ',', -1, false}:      "1.234,01",
		{12345.01, '.', ',', -1, false}:     "12.345,01",
		{123456.01, '.', ',', -1, false}:    "123.456,01",
		{1234567.01, '.', ',', -1, false}:   "1.234.567,01",
		{12345678.01, '.', ',', -1, false}:  "12.345.678,01",
		{123456789.01, '.', ',', -1, false}: "123.456.789,01",

		{1234.01, 0, '.', -1, false}:      "1234.01",
		{12345.01, 0, '.', -1, false}:     "12345.01",
		{123456.01, 0, '.', -1, false}:    "123456.01",
		{1234567.01, 0, '.', -1, false}:   "1234567.01",
		{12345678.01, 0, '.', -1, false}:  "12345678.01",
		{123456789.01, 0, '.', -1, false}: "123456789.01",

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

// Test_FormatFloat_precisionRounding checks that a non-negative precision
// rounds the value. strconv uses round-half-to-even; every input below is an
// exact float64 value so the expected output is deterministic.
func Test_FormatFloat_precisionRounding(t *testing.T) {
	tests := []struct {
		f            float64
		thousandsSep rune
		precision    int
		want         string
	}{
		{0.5, 0, 0, "0"},
		{1.5, 0, 0, "2"},
		{2.5, 0, 0, "2"},
		{3.5, 0, 0, "4"},
		{0.125, 0, 2, "0.12"},
		{0.375, 0, 2, "0.38"},
		{0.625, 0, 2, "0.62"},
		{0.875, 0, 2, "0.88"},
		{1.0625, 0, 3, "1.062"},
		{1234.5, ',', 0, "1,234"},
		{1234567.875, ',', 2, "1,234,567.88"},
	}
	for _, tt := range tests {
		got := Format(tt.f, tt.thousandsSep, '.', tt.precision, false)
		assert.Equal(t, tt.want, got, "Format(%v, '%s', '.', %d, false)", tt.f, string(tt.thousandsSep), tt.precision)
	}
}

// Test_FormatFloat_zero checks formatting of zero across separator and
// padding combinations.
func Test_FormatFloat_zero(t *testing.T) {
	assert.Equal(t, "0", Format(0.0, 0, '.', -1, false))
	assert.Equal(t, "0", Format(0.0, ',', '.', -1, false))
	assert.Equal(t, "0", Format(0.0, '.', ',', -1, false))
	assert.Equal(t, "0.00", Format(0.0, 0, '.', 2, true))
	assert.Equal(t, "0,00", Format(0.0, '.', ',', 2, true))
}

// Test_FormatFloat_padPrecision contrasts the two padPrecision modes for a
// non-negative precision: true keeps exactly precision fractional digits,
// false trims trailing zeros to the shortest form of the rounded value.
func Test_FormatFloat_padPrecision(t *testing.T) {
	// padPrecision == true: exactly precision fractional digits.
	assert.Equal(t, "1.500", Format(1.5, 0, '.', 3, true))
	assert.Equal(t, "2.000", Format(2.0, 0, '.', 3, true))
	assert.Equal(t, "1.250000", Format(1.25, 0, '.', 6, true))
	assert.Equal(t, "1,234.5000", Format(1234.5, ',', '.', 4, true))
	assert.Equal(t, "1.234.567,000", Format(1234567.0, '.', ',', 3, true))

	// padPrecision == false: trailing fractional zeros trimmed.
	assert.Equal(t, "1.5", Format(1.5, 0, '.', 3, false))
	assert.Equal(t, "2", Format(2.0, 0, '.', 3, false))
	assert.Equal(t, "1.25", Format(1.25, 0, '.', 6, false))
	assert.Equal(t, "1,234.5", Format(1234.5, ',', '.', 4, false))
	assert.Equal(t, "1.234.567", Format(1234567.0, '.', ',', 3, false))

	// precision rounds the value before any trimming or padding.
	assert.Equal(t, "1.2", Format(1.25, 0, '.', 1, false))
	assert.Equal(t, "1.2", Format(1.25, 0, '.', 1, true))

	// precision -1 always yields the shortest form, padPrecision is moot.
	assert.Equal(t, "1.5", Format(1.5, 0, '.', -1, false))
	assert.Equal(t, "1.5", Format(1.5, 0, '.', -1, true))
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

func Test_FormatFloat_float32AndSpecials(t *testing.T) {
	// float32 values format with float32 shortest-roundtrip precision,
	// not the longer float64 expansion of the same value.
	assert.Equal(t, "1.2", Format(float32(1.2), 0, '.', -1, false))
	assert.Equal(t, "0.1", Format(float32(0.1), 0, '.', -1, false))

	// NaN and ±Inf are returned as-is, never grouped or padded.
	assert.Equal(t, "NaN", Format(math.NaN(), ',', '.', -1, false))
	assert.Equal(t, "+Inf", Format(math.Inf(1), ',', '.', -1, false))
	assert.Equal(t, "-Inf", Format(math.Inf(-1), ',', '.', -1, false))
	assert.Equal(t, "+Inf", Format(math.Inf(1), ',', '.', 2, true))
}
