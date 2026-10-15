package float

import (
	"math"
	"math/rand/v2"
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
// rounds the value half away from zero like RoundToDecimals and all money
// formatting do, and not to even like strconv does.
//
// The value that is rounded is the shortest decimal representation of the
// float, the decimal a human typed or a JSON number carried, not its binary
// approximation: 1.005 is stored as 1.00499999999999989342 and must still
// format as "1.01", because the same number rounded to "1.01" by one
// function and to "1.00" by another makes a report disagree with itself.
// This test used to pin strconv's ties to even ("2" for 2.5 and "0.12" for
// 0.125), which contradicted RoundToDecimals and the money formatting.
func Test_FormatFloat_precisionRounding(t *testing.T) {
	tests := []struct {
		f            float64
		thousandsSep rune
		precision    int
		want         string
	}{
		{0.5, 0, 0, "1"},
		{1.5, 0, 0, "2"},
		{2.5, 0, 0, "3"},
		{3.5, 0, 0, "4"},
		{0.125, 0, 2, "0.13"},
		{0.375, 0, 2, "0.38"},
		{0.625, 0, 2, "0.63"},
		{0.875, 0, 2, "0.88"},
		{1.0625, 0, 3, "1.063"},
		{1234.5, ',', 0, "1,235"},
		{1234567.875, ',', 2, "1,234,567.88"},
		// Ties of the decimal that are below the tie in binary
		{1.005, 0, 2, "1.01"},
		{0.145, 0, 2, "0.15"},
		{2.675, 0, 2, "2.68"},
		{1234.565, '\'', 2, "1'234.57"},
		// Not a tie
		{0.144999, 0, 2, "0.14"},
		// A carry adds an integer digit that has to be grouped
		{999.995, ',', 2, "1,000"},
		{999999.5, '\'', 0, "1'000'000"},
		// The shortest decimal of 1e23 is 1e23,
		// its binary value 99999999999999991611392
		{1e23, 0, 0, "100000000000000000000000"},
		{1e23, ',', 2, "100,000,000,000,000,000,000,000"},
	}
	for _, tt := range tests {
		got := Format(tt.f, tt.thousandsSep, '.', tt.precision, false)
		assert.Equal(t, tt.want, got, "Format(%v, '%s', '.', %d, false)", tt.f, string(tt.thousandsSep), tt.precision)
		// Half away from zero is symmetric around zero
		got = Format(-tt.f, tt.thousandsSep, '.', tt.precision, false)
		assert.Equal(t, "-"+tt.want, got, "Format(%v, '%s', '.', %d, false)", -tt.f, string(tt.thousandsSep), tt.precision)
	}

	// padPrecision pads the rounded digits
	assert.Equal(t, "1.01", Format(1.005, 0, '.', 2, true))
	assert.Equal(t, "1.0050", Format(1.005, 0, '.', 4, true), "fewer decimals than precision are padded, not rounded")
	assert.Equal(t, "1,000.00", Format(999.995, ',', '.', 2, true))
	// float32 values round their shortest float32 decimal
	assert.Equal(t, "1.01", Format(float32(1.005), 0, '.', 2, false))
	// A value that is zero after rounding is formatted without a sign,
	// unlike by strconv, because "-0.00" is no amount: money.Amount(-0.004)
	// must format like the money.DecimalAmount -0.004, which is "0.00"
	assert.Equal(t, "0.00", Format(-0.001, 0, '.', 2, true))
	assert.Equal(t, "0", Format(-0.4, 0, '.', 0, false))
	assert.Equal(t, "0", Format(math.Copysign(0, -1), 0, '.', -1, false))
	assert.Equal(t, "0,00", Format(math.Copysign(0, -1), '.', ',', 2, true))
	assert.Equal(t, "-0.01", Format(-0.005, 0, '.', 2, true), "rounded away from zero keeps the sign")
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

	// precision rounds the value half away from zero
	// before any trimming or padding.
	assert.Equal(t, "1.3", Format(1.25, 0, '.', 1, false))
	assert.Equal(t, "1.3", Format(1.25, 0, '.', 1, true))
	assert.Equal(t, "1", Format(1.04, 0, '.', 1, false))
	assert.Equal(t, "1.0", Format(1.04, 0, '.', 1, true))

	// precision -1 always yields the shortest form, padPrecision is moot.
	assert.Equal(t, "1.5", Format(1.5, 0, '.', -1, false))
	assert.Equal(t, "1.5", Format(1.5, 0, '.', -1, true))
}

// TestFormat_fastPathMatchesExact protects that the float64 fast path of
// Format never changes a formatted string. It is only allowed to skip
// formatting and rounding the shortest decimal digits where it provably
// writes the identical digits, sign and point, so every input must format
// exactly like rounding the digits does, most of all the decimal ties and
// their float64 neighbors where float64 arithmetic alone gets it wrong.
func TestFormat_fastPathMatchesExact(t *testing.T) {
	check := func(f float64, precision int) {
		t.Helper()
		got := formatRounded(f, 64, precision)
		want := formatRoundedExact(f, 64, precision)
		if got != want {
			t.Fatalf("formatRounded(%v, 64, %d) = %q, exact digit rounding = %q", f, precision, got, want)
		}
	}
	for _, f := range fastPathTargetedFloats() {
		for precision := -1; precision <= 24; precision++ {
			check(f, precision)
		}
	}
	for _, tie := range fastPathTies() {
		for precision := max(tie.decimals-1, 0); precision <= tie.decimals+1; precision++ {
			check(tie.f, precision)
		}
	}
	rng := rand.New(rand.NewPCG(7, 8))
	for range fastPathSweepSize() {
		check(randomFastPathFloat(rng), rng.IntN(13))
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

func BenchmarkFormat(b *testing.B) {
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_ = Format(benchFloats[i%len(benchFloats)], 0, '.', 2, true)
		i++
	}
}
