package money

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecimalAmount_packing(t *testing.T) {
	cases := []struct {
		coeff int64
		scale int
	}{
		{0, 0},
		{123456, 2},
		{-123456, 2},
		{1, 18},
		{-1, 18},
		{maxDecimalAmountCoefficient, 0},
		{minDecimalAmountCoefficient, 0},
		{maxDecimalAmountCoefficient, 18},
		{minDecimalAmountCoefficient, 18},
	}
	for _, c := range cases {
		a := DecimalAmountFromCoefficient(c.coeff, c.scale)
		assert.Equal(t, c.coeff, a.Coefficient(), "coefficient for %d/%d", c.coeff, c.scale)
		assert.Equal(t, c.scale, a.Scale(), "scale for %d/%d", c.coeff, c.scale)
	}
}

func TestDecimalAmount_zeroValue(t *testing.T) {
	var a DecimalAmount
	assert.Equal(t, int64(0), a.Coefficient())
	assert.Equal(t, 0, a.Scale())
	assert.True(t, a.IsZero())
	assert.Equal(t, "0", a.String())
}

func TestDecimalAmountFromCoefficient_panics(t *testing.T) {
	assert.Panics(t, func() { DecimalAmountFromCoefficient(1, -1) })
	assert.Panics(t, func() { DecimalAmountFromCoefficient(1, MaxDecimalAmountScale+1) })
	assert.Panics(t, func() { DecimalAmountFromCoefficient(maxDecimalAmountCoefficient+1, 0) })
	assert.Panics(t, func() { DecimalAmountFromCoefficient(minDecimalAmountCoefficient-1, 0) })
}

func TestParseDecimalAmount(t *testing.T) {
	cases := []struct {
		str   string
		coeff int64
		scale int
	}{
		{"0", 0, 0},
		{"1234", 1234, 0},
		{"1234.56", 123456, 2},
		{"-99.99", -9999, 2},
		{"0.005", 5, 3},
		{"-0.005", -5, 3},
		{"007.50", 750, 2},
		{"1,234.56", 123456, 2}, // English grouping
		{"1.234,56", 123456, 2}, // German grouping
		{"1'234.56", 123456, 2}, // Swiss grouping
		{"1,234,567.89", 123456789, 2},
		{"-0.00", 0, 2},
		// Exactness: this loses its last cent through a float64 round-trip
		// but must be preserved here.
		{"99999999999999.99", 9999999999999999, 2},
	}
	for _, c := range cases {
		a, err := ParseDecimalAmount(c.str)
		require.NoError(t, err, "ParseDecimalAmount(%q)", c.str)
		assert.Equal(t, c.coeff, a.Coefficient(), "coefficient of %q", c.str)
		assert.Equal(t, c.scale, a.Scale(), "scale of %q", c.str)
	}
}

func TestParseDecimalAmount_exactBeatsFloat(t *testing.T) {
	const s = "99999999999999.99"
	a, err := ParseDecimalAmount(s)
	require.NoError(t, err)
	assert.Equal(t, s, a.String())

	// The float64 path used by the plain Amount type loses the last cent.
	f, err := ParseAmount(s)
	require.NoError(t, err)
	assert.NotEqual(t, s, f.String())
}

func TestParseDecimalAmount_acceptedDecimals(t *testing.T) {
	_, err := ParseDecimalAmount("1.234", 0, 2)
	assert.Error(t, err, "3 decimals not in {0,2}")

	a, err := ParseDecimalAmount("1.23", 0, 2)
	require.NoError(t, err)
	assert.Equal(t, "1.23", a.String())
}

func TestParseDecimalAmount_errors(t *testing.T) {
	for _, s := range []string{
		"",
		"1.5e3",                 // scientific notation rejected
		"1.1234567890123456789", // 19 decimals > MaxDecimalAmountScale
		"999999999999999999",    // 18 nines ~1e18 exceeds coefficient range
		"abc",
	} {
		_, err := ParseDecimalAmount(s)
		assert.Error(t, err, "expected error parsing %q", s)
	}
}

func TestParseDecimalAmount_nonFinite(t *testing.T) {
	for _, s := range []string{"NaN"} {
		a, err := ParseDecimalAmount(s)
		require.NoError(t, err)
		assert.True(t, a.IsNaN(), "%q should parse to NaN", s)
	}
	for _, s := range []string{"Inf", "+Inf"} {
		a, err := ParseDecimalAmount(s)
		require.NoError(t, err)
		assert.True(t, a.IsInf(1), "%q should parse to +Inf", s)
	}
	a, err := ParseDecimalAmount("-Inf")
	require.NoError(t, err)
	assert.True(t, a.IsInf(-1))
}

func TestDecimalAmount_String(t *testing.T) {
	cases := map[string]DecimalAmount{
		"0":        DecimalAmountFromCoefficient(0, 0),
		"0.00":     DecimalAmountFromCoefficient(0, 2),
		"1234.56":  DecimalAmountFromCoefficient(123456, 2),
		"-1234.56": DecimalAmountFromCoefficient(-123456, 2),
		"0.005":    DecimalAmountFromCoefficient(5, 3),
		"-0.005":   DecimalAmountFromCoefficient(-5, 3),
		"1.50":     DecimalAmountFromCoefficient(150, 2),
		"1000":     DecimalAmountFromCoefficient(1000, 0),
	}
	for want, a := range cases {
		assert.Equal(t, want, a.String(), "String of coeff=%d scale=%d", a.Coefficient(), a.Scale())
	}
}

func TestDecimalAmount_GoString(t *testing.T) {
	assert.Equal(t, "money.DecimalAmountFromCoefficient(123456, 2)", DecimalAmountFromCoefficient(123456, 2).GoString())
	assert.Equal(t, "money.DecimalAmountFromCoefficient(123456, 2)", fmt.Sprintf("%#v", DecimalAmountFromCoefficient(123456, 2)))
	// Non-finite sentinels produce constructor-call source (round-trips as Go).
	assert.Equal(t, "money.DecimalAmountNaN()", DecimalAmountNaN().GoString())
	assert.Equal(t, "money.DecimalAmountInf(-1)", DecimalAmountInf(-1).GoString())
}

func TestRoundingMode_String(t *testing.T) {
	// Every mode has a distinct name and an unknown mode is reported by value;
	// the strings match the constant identifiers so they are safe to log/parse.
	cases := map[RoundingMode]string{
		RoundHalfAwayFromZero: "RoundHalfAwayFromZero",
		RoundHalfToEven:       "RoundHalfToEven",
		RoundHalfUp:           "RoundHalfUp",
		RoundHalfDown:         "RoundHalfDown",
		RoundDown:             "RoundDown",
		RoundUp:               "RoundUp",
		RoundFloor:            "RoundFloor",
		RoundCeil:             "RoundCeil",
	}
	for mode, want := range cases {
		assert.Equal(t, want, mode.String())
	}
	assert.Equal(t, "Unknown rounding mode: 99", RoundingMode(99).String())
}

func TestDecimalAmount_FormatSep(t *testing.T) {
	a := DecimalAmountFromCoefficient(123456789, 2) // 1234567.89
	assert.Equal(t, "1234567.89", a.FormatSep(0, '.'))
	assert.Equal(t, "1,234,567.89", a.FormatSep(',', '.'))
	assert.Equal(t, "1.234.567,89", a.FormatSep('.', ','))
	assert.Equal(t, "-1.234.567,89", a.Neg().FormatSep('.', ','))
	// zero decimalSep defaults to '.'
	assert.Equal(t, "1,234,567.89", a.FormatSep(',', 0))
}

func TestDecimalAmount_Format(t *testing.T) {
	a := DecimalAmountFromCoefficient(123456, 2) // 1234.56
	cases := map[string]string{
		"%s":     "1234.56",
		"%v":     "1234.56",
		"%q":     `"1234.56"`,
		"%f":     "1234.56",   // default precision = scale
		"%.1f":   "1234.6",    // rounds half away from zero
		"%.4f":   "1234.5600", // pads
		"%.0f":   "1235",
		"%d":     "1235",
		"%#f":    "1,234.56", // grouping
		"%#d":    "1,235",
		"%+.2f":  "+1234.56",
		"% .2f":  " 1234.56",
		"%8.1f":  "  1234.6", // width padding
		"%-8.1f": "1234.6  ",
		"%08.1f": "001234.6",
	}
	for format, want := range cases {
		assert.Equal(t, want, fmt.Sprintf(format, a), "format %q", format)
	}
	// Negative zero-padding keeps the sign leftmost.
	assert.Equal(t, "-01234.6", fmt.Sprintf("%08.1f", a.Neg()))
	// Unknown verb reports the type.
	assert.Equal(t, "%!x(money.DecimalAmount=1234.56)", fmt.Sprintf("%x", a))
	// Formatting must not panic even for the largest value at high precision.
	big := DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0)
	assert.NotPanics(t, func() { _ = fmt.Sprintf("%.18f", big) })
}

func TestDecimalAmount_CmpEqual(t *testing.T) {
	// Same value, different scale.
	a := DecimalAmountFromCoefficient(150, 2) // 1.50
	b := DecimalAmountFromCoefficient(15, 1)  // 1.5
	assert.NotEqual(t, a, b, "different packed representation")
	assert.True(t, a.Equal(b), "equal value")
	assert.Equal(t, 0, a.Cmp(b))

	assert.Equal(t, -1, DecimalAmountFromCoefficient(149, 2).Cmp(b))
	assert.Equal(t, +1, DecimalAmountFromCoefficient(151, 2).Cmp(b))
	assert.Equal(t, -1, DecimalAmountFromCoefficient(-1, 0).Cmp(DecimalAmountFromCoefficient(1, 0)))
	// Equal scale, differing coefficients exercise both fast-path branches.
	assert.Equal(t, +1, DecimalAmountFromCoefficient(200, 2).Cmp(DecimalAmountFromCoefficient(100, 2)))
	assert.Equal(t, -1, DecimalAmountFromCoefficient(100, 2).Cmp(DecimalAmountFromCoefficient(200, 2)))
	// Large cross-scale comparison must not overflow.
	assert.Equal(t, +1, DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0).Cmp(DecimalAmountFromCoefficient(1, 18)))
}

func TestDecimalAmount_AddSub(t *testing.T) {
	a := DecimalAmountFromCoefficient(123, 2)  // 1.23
	b := DecimalAmountFromCoefficient(4567, 4) // 0.4567
	sum := a.Add(b)
	assert.Equal(t, "1.6867", sum.String())
	assert.Equal(t, 4, sum.Scale())

	diff := a.Sub(b)
	assert.Equal(t, "0.7733", diff.String())

	// Adding opposite values yields zero at the larger scale.
	z := DecimalAmountFromCoefficient(5, 2).Add(DecimalAmountFromCoefficient(-5, 2))
	assert.True(t, z.IsZero())

	// Overflow of the integer part yields +Inf (or -Inf for negative operands).
	assert.True(t, DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0).Add(DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0)).IsInf(1))
	assert.True(t, DecimalAmountFromCoefficient(minDecimalAmountCoefficient, 0).Add(DecimalAmountFromCoefficient(minDecimalAmountCoefficient, 0)).IsInf(-1))
}

// TestDecimalAmount_AddScaleMismatch guards the regression where aligning a
// representable sum to the larger scale overflowed to ±Inf.
func TestDecimalAmount_AddScaleMismatch(t *testing.T) {
	// x + 0 is identity even when the zero carries a large scale (used to
	// return +Inf because 2.5 is not representable at scale 18).
	x := DecimalAmountFromCoefficient(25, 1) // 2.5
	sum := x.Add(DecimalAmountFromCoefficient(0, 18))
	assert.True(t, sum.IsFinite(), "x + 0 must stay finite")
	assert.True(t, x.Equal(sum), "x + 0 must equal x, got %s", sum)

	// Near-cancellation of large opposite-sign operands yields the small sum.
	got := DecimalAmountFromCoefficient(28823037615171175, 0).Add(DecimalAmountFromCoefficient(-288230376151711743, 1))
	assert.Equal(t, "0.7", got.String())

	// Sub inherits the fix.
	assert.True(t, x.Equal(x.Sub(DecimalAmountFromCoefficient(0, 18))))
}

// TestDecimalAmount_AddOracle fuzzes Add against a big.Rat reference, including
// the representability boundary for the ±Inf result.
func TestDecimalAmount_AddOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	randDec := func() DecimalAmount {
		return DecimalAmountFromCoefficient(rng.Int63n(2*maxDecimalAmountCoefficient+1)-maxDecimalAmountCoefficient, rng.Intn(MaxDecimalAmountScale+1))
	}
	toRat := func(d DecimalAmount) *big.Rat {
		r := new(big.Rat).SetInt64(d.Coefficient())
		return r.Quo(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d.Scale())), nil)))
	}
	maxc := big.NewInt(maxDecimalAmountCoefficient)
	representable := func(want *big.Rat) bool {
		for s := 0; s <= MaxDecimalAmountScale; s++ {
			scaled := new(big.Rat).Mul(want, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s)), nil)))
			if scaled.IsInt() {
				return scaled.Num().CmpAbs(maxc) <= 0
			}
		}
		return false
	}
	for range 20000 {
		a, b := randDec(), randDec()
		got := a.Add(b)
		want := new(big.Rat).Add(toRat(a), toRat(b))
		if got.IsInf(0) {
			require.Falsef(t, representable(want), "Add(%s,%s)=Inf but %s is representable", a, b, want.FloatString(20))
			continue
		}
		require.Zerof(t, toRat(got).Cmp(want), "Add(%s,%s)=%s, want %s", a, b, got, want.FloatString(20))
	}
}

// TestDecimalAmount_MulDivOracle verifies the max-precision contract of Mul
// and Div against big.Rat: the result is exact whenever the true value is
// representable, rounding (half away from zero here) only touches the last
// representable digit, and ±Inf occurs only for a true integer-part overflow.
func TestDecimalAmount_MulDivOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	randDec := func() DecimalAmount {
		// Mix small and full-range coefficients so exact results, forced
		// rounding and overflow all occur.
		coeff := rng.Int63n(2*maxDecimalAmountCoefficient+1) - maxDecimalAmountCoefficient
		if rng.Intn(2) == 0 {
			coeff = rng.Int63n(20001) - 10000
		}
		return DecimalAmountFromCoefficient(coeff, rng.Intn(MaxDecimalAmountScale+1))
	}
	toRat := func(d DecimalAmount) *big.Rat {
		r := new(big.Rat).SetInt64(d.Coefficient())
		return r.Quo(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d.Scale())), nil)))
	}
	maxc := big.NewInt(maxDecimalAmountCoefficient)
	check := func(op string, a, b, got DecimalAmount, want *big.Rat) {
		t.Helper()
		if got.IsInf(0) {
			// ±Inf only when even the integer part exceeds the coefficient range.
			require.Truef(t, new(big.Rat).Abs(want).Cmp(new(big.Rat).SetInt(maxc)) > 0,
				"%s(%s,%s)=Inf but |%s| fits the coefficient range", op, a, b, want.FloatString(25))
			return
		}
		if toRat(got).Cmp(want) == 0 {
			return // exact result
		}
		// Rounded result: the error is at most half an ulp of the result scale
		// (RoundHalfAwayFromZero) ...
		diff := new(big.Rat).Sub(toRat(got), want)
		diff.Abs(diff)
		halfUlp := big.NewRat(1, 2)
		halfUlp.Quo(halfUlp, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(got.Scale())), nil)))
		require.Truef(t, diff.Cmp(halfUlp) <= 0,
			"%s(%s,%s)=%s deviates from %s by more than half an ulp", op, a, b, got, want.FloatString(25))
		// ... and only happens because one more decimal place cannot fit:
		// rounding |want| half away from zero at scale+1 would exceed the
		// coefficient range, i.e. |want|·10^(scale+1) ≥ max + 1/2.
		if got.Scale() < MaxDecimalAmountScale {
			scaled := new(big.Rat).Abs(want)
			scaled.Mul(scaled, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(got.Scale()+1)), nil)))
			limit := new(big.Rat).Add(new(big.Rat).SetInt(maxc), big.NewRat(1, 2))
			require.Truef(t, scaled.Cmp(limit) >= 0,
				"%s(%s,%s)=%s was rounded although scale %d is not maximal", op, a, b, got, got.Scale())
		}
	}
	for range 20000 {
		a, b := randDec(), randDec()
		check("Mul", a, b, a.Mul(b, RoundHalfAwayFromZero), new(big.Rat).Mul(toRat(a), toRat(b)))
		if !b.IsZero() {
			check("Div", a, b, a.Div(b, RoundHalfAwayFromZero), new(big.Rat).Quo(toRat(a), toRat(b)))
		}
	}
}

func TestDecimalAmount_ScanStringValidate(t *testing.T) {
	var d DecimalAmount
	// validate=true rejects non-finite input (mirrors Amount.ScanString).
	assert.Error(t, d.ScanString("NaN", true))
	assert.Error(t, d.ScanString("Inf", true))
	assert.Error(t, d.ScanString("-Infinity", true))
	// validate=false still accepts it as a sentinel.
	require.NoError(t, d.ScanString("NaN", false))
	assert.True(t, d.IsNaN())
	// Finite values pass validation.
	require.NoError(t, d.ScanString("1.23", true))
	assert.Equal(t, "1.23", d.String())

	var ca CurrencyDecimalAmount
	assert.Error(t, ca.ScanString("EUR Inf", true))
	require.NoError(t, ca.ScanString("EUR 1.23", true))
	assert.Equal(t, "1.23", ca.Amount.String())
}

func TestDecimalAmount_MulInt(t *testing.T) {
	a := DecimalAmountFromCoefficient(199, 2) // 1.99
	assert.Equal(t, "5.97", a.MulInt(3).String())
	assert.Equal(t, "-3.98", a.MulInt(-2).String())
	assert.Equal(t, "0.00", a.MulInt(0).String())
	assert.True(t, DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0).MulInt(2).IsInf(1))

	// MulInt64 takes the wider int64 and MulInt delegates to it.
	assert.Equal(t, "5.97", a.MulInt64(3).String())
	assert.Equal(t, a.MulInt(3), a.MulInt64(3))
	assert.True(t, DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0).MulInt64(2).IsInf(1))
	assert.True(t, DecimalAmountFromCoefficient(minDecimalAmountCoefficient, 0).MulInt(2).IsInf(-1))

	// A coefficient overflow with exact trailing zeros reduces the scale
	// instead of overflowing to Inf.
	assert.Equal(t,
		DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0),
		DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 2).MulInt(100))
}

func TestDecimalAmount_Mul(t *testing.T) {
	cases := []struct {
		a, b DecimalAmount
		want string
	}{
		// The exact product keeps the full precision: the result scale is the
		// sum of the operand scales.
		{DecimalAmountFromCoefficient(150, 2), DecimalAmountFromCoefficient(150, 2), "2.2500"},     // 1.50 * 1.50
		{DecimalAmountFromCoefficient(10, 1), DecimalAmountFromCoefficient(10, 1), "1.00"},         // 1.0 * 1.0
		{DecimalAmountFromCoefficient(3, 0), DecimalAmountFromCoefficient(4, 0), "12"},             // 3 * 4
		{DecimalAmountFromCoefficient(-2, 0), DecimalAmountFromCoefficient(150, 2), "-3.00"},       // -2 * 1.50
		{DecimalAmountFromCoefficient(100, 2), DecimalAmountFromCoefficient(119, 2), "1.1900"},     // 1.00 * 1.19
		{DecimalAmountFromCoefficient(11, 1), DecimalAmountFromCoefficient(11, 1), "1.21"},         // 1.1 * 1.1 stays exact
		{DecimalAmountFromCoefficient(105, 4), DecimalAmountFromCoefficient(105, 4), "0.00011025"}, // 0.0105²
	}
	for _, c := range cases {
		got := c.a.Mul(c.b, RoundHalfAwayFromZero)
		assert.Equal(t, c.want, got.String(), "%s * %s", c.a, c.b)
		assert.Equal(t, c.a.Scale()+c.b.Scale(), got.Scale())
	}

	// Large product whose exact 128-bit intermediate exceeds 64 bits and whose
	// coefficient at scale 18 exceeds the coefficient range: the scale is
	// reduced by stripping exact trailing zeros, losing no data.
	// 5.000000000 * 5.000000000 = 25 with 16 remaining zero decimals.
	x := DecimalAmountFromCoefficient(5_000_000_000, 9)
	assert.Equal(t, "25.0000000000000000", x.Mul(x, RoundHalfAwayFromZero).String())
	assert.Equal(t, 16, x.Mul(x, RoundHalfAwayFromZero).Scale())

	// A product needing more than the representable precision is rounded with
	// the given mode at the largest scale that fits (this is the only case
	// where Mul rounds): 0.111111111111111111 * 0.5 has 19 decimal places.
	y := DecimalAmountFromCoefficient(111_111_111_111_111_111, 18)
	assert.Equal(t, "0.055555555555555556", y.Mul(DecimalAmountFromCoefficient(5, 1), RoundHalfAwayFromZero).String())
	assert.Equal(t, "0.055555555555555555", y.Mul(DecimalAmountFromCoefficient(5, 1), RoundDown).String())

	// Overflow yields ±Inf by the product sign.
	assert.True(t, DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0).Mul(DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0), RoundHalfAwayFromZero).IsInf(1))
	assert.True(t, DecimalAmountFromCoefficient(minDecimalAmountCoefficient, 0).Mul(DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0), RoundHalfAwayFromZero).IsInf(-1))
}

func TestDecimalAmount_Div(t *testing.T) {
	cases := []struct {
		a, b DecimalAmount
		want string
	}{
		// Terminating quotients are exact, with trailing zeros stripped down
		// to (at least) the preferred scale a.Scale()-b.Scale().
		{DecimalAmountFromCoefficient(1000, 2), DecimalAmountFromCoefficient(400, 2), "2.5"},           // 10.00 / 4.00
		{DecimalAmountFromCoefficient(1000, 2), DecimalAmountFromCoefficient(500, 2), "2"},             // 10.00 / 5.00
		{DecimalAmountFromCoefficient(1000, 2), DecimalAmountFromCoefficient(5, 0), "2.00"},            // 10.00 / 5 keeps the preferred 2 decimals
		{DecimalAmountFromCoefficient(1000, 3), DecimalAmountFromCoefficient(8, 0), "0.125"},           // 1.000 / 8
		{DecimalAmountFromCoefficient(100, 2), DecimalAmountFromCoefficient(8, 0), "0.125"},            // 1.00 / 8 extends beyond the preferred scale
		{DecimalAmountFromCoefficient(1, 0), DecimalAmountFromCoefficient(3, 5), "33333.333333333333"}, // 1 / 0.00003
		// Non-terminating quotients are rounded at the largest representable scale.
		{DecimalAmountFromCoefficient(1000, 2), DecimalAmountFromCoefficient(3, 0), "3.3333333333333333"},   // 10.00 / 3
		{DecimalAmountFromCoefficient(10, 0), DecimalAmountFromCoefficient(3, 0), "3.3333333333333333"},     // 10 / 3
		{DecimalAmountFromCoefficient(-1000, 2), DecimalAmountFromCoefficient(3, 0), "-3.3333333333333333"}, // -10.00 / 3
	}
	for _, c := range cases {
		got := c.a.Div(c.b, RoundHalfAwayFromZero)
		assert.Equal(t, c.want, got.String(), "%s / %s", c.a, c.b)
	}

	// Division by zero yields ±Inf, and 0/0 yields NaN.
	assert.True(t, DecimalAmountFromCoefficient(1, 0).Div(DecimalAmountFromCoefficient(0, 2), RoundHalfAwayFromZero).IsInf(1))
	assert.True(t, DecimalAmountFromCoefficient(-1, 0).Div(DecimalAmountFromCoefficient(0, 2), RoundHalfAwayFromZero).IsInf(-1))
	assert.True(t, DecimalAmountFromCoefficient(0, 0).Div(DecimalAmountFromCoefficient(0, 2), RoundHalfAwayFromZero).IsNaN())

	// A quotient whose integer part overflows the coefficient range yields ±Inf.
	huge := DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0)
	tiny := DecimalAmountFromCoefficient(1, 18)
	assert.True(t, huge.Div(tiny, RoundHalfAwayFromZero).IsInf(1))
	assert.True(t, huge.Neg().Div(tiny, RoundHalfAwayFromZero).IsInf(-1))
}

func TestDecimalAmount_MulDivRoundingModes(t *testing.T) {
	// Mul and Div round only when the exact result needs more precision than
	// representable. 1/3 fills all 17 significant digits, so the mode decides
	// the last digit.
	assert.Equal(t, "0.33333333333333333", DecimalAmountFromCoefficient(100, 2).Div(DecimalAmountFromCoefficient(3, 0), RoundDown).String())
	assert.Equal(t, "0.33333333333333334", DecimalAmountFromCoefficient(100, 2).Div(DecimalAmountFromCoefficient(3, 0), RoundUp).String())

	// 0.111111111111111111 / 2 = 0.0555555555555555555 needs 19 decimal
	// places, one more than representable, so the dropped digit 5 is an exact
	// tie. The truncated quotient ...555 is odd, so RoundHalfToEven rounds up.
	tie := func(sign int64, mode RoundingMode) string {
		return DecimalAmountFromCoefficient(sign*111_111_111_111_111_111, 18).
			Div(DecimalAmountFromCoefficient(2, 0), mode).String()
	}
	const down, up = "0.055555555555555555", "0.055555555555555556"
	assert.Equal(t, up, tie(1, RoundHalfAwayFromZero))
	assert.Equal(t, up, tie(1, RoundHalfToEven)) // odd quotient rounds up to even
	assert.Equal(t, up, tie(1, RoundHalfUp))     // ties toward +Inf
	assert.Equal(t, down, tie(1, RoundHalfDown)) // ties toward -Inf
	assert.Equal(t, "-"+down, tie(-1, RoundHalfUp))
	assert.Equal(t, "-"+up, tie(-1, RoundHalfDown))
	assert.Equal(t, "-"+up, tie(-1, RoundHalfAwayFromZero))

	// Directed modes.
	assert.Equal(t, down, tie(1, RoundFloor))
	assert.Equal(t, up, tie(1, RoundCeil))
	assert.Equal(t, "-"+up, tie(-1, RoundFloor))
	assert.Equal(t, "-"+down, tie(-1, RoundCeil))
	assert.Equal(t, down, tie(1, RoundDown))
	assert.Equal(t, up, tie(1, RoundUp))

	assert.Equal(t, "RoundHalfToEven", RoundHalfToEven.String())
}

func TestDecimalAmount_Rounding(t *testing.T) {
	cases := []struct {
		in       DecimalAmount
		decimals int
		want     string
	}{
		{DecimalAmountFromCoefficient(12345, 3), 2, "12.35"}, // 12.345 -> 12.35 (half away)
		{DecimalAmountFromCoefficient(12344, 3), 2, "12.34"},
		{DecimalAmountFromCoefficient(-12345, 3), 2, "-12.35"},
		{DecimalAmountFromCoefficient(125, 2), 1, "1.3"},    // 1.25 -> 1.3 (half away from zero)
		{DecimalAmountFromCoefficient(150, 2), 4, "1.5000"}, // padding
		{DecimalAmountFromCoefficient(19, 1), 0, "2"},       // 1.9 -> 2
		{DecimalAmountFromCoefficient(-15, 1), 0, "-2"},     // -1.5 -> -2 (half away)
	}
	for _, c := range cases {
		got := c.in.RoundToDecimals(c.decimals, RoundHalfAwayFromZero).String()
		assert.Equal(t, c.want, got, "%s round to %d", c.in, c.decimals)
	}
	assert.Equal(t, "1.25", DecimalAmountFromCoefficient(12500, 4).RoundToCents(RoundHalfAwayFromZero).String())
	assert.Equal(t, "2", DecimalAmountFromCoefficient(150, 2).RoundToInt(RoundHalfAwayFromZero).String())

	// RoundToDecimals honors the rounding mode.
	assert.Equal(t, "12.34", DecimalAmountFromCoefficient(12345, 3).RoundToDecimals(2, RoundHalfToEven).String()) // 12.345 -> 12.34 (even)
	assert.Equal(t, "12.35", DecimalAmountFromCoefficient(12350, 3).RoundToDecimals(2, RoundUp).String())
	assert.Equal(t, "12.34", DecimalAmountFromCoefficient(12345, 3).RoundToDecimals(2, RoundDown).String())
	assert.Equal(t, "-12.35", DecimalAmountFromCoefficient(-12341, 3).RoundToDecimals(2, RoundFloor).String())
}

func TestDecimalAmount_FloatAndAmount(t *testing.T) {
	a := DecimalAmountFromCoefficient(123456, 2)
	assert.InDelta(t, 1234.56, a.Float(), 1e-9)
	assert.InDelta(t, 1234.56, float64(a.Amount()), 1e-9)
}

func TestDecimalAmount_JSON(t *testing.T) {
	a := DecimalAmountFromCoefficient(123456, 2)
	data, err := json.Marshal(a)
	require.NoError(t, err)
	assert.Equal(t, "1234.56", string(data)) // unquoted number

	var got DecimalAmount
	require.NoError(t, json.Unmarshal([]byte("1234.56"), &got))
	assert.True(t, a.Equal(got))

	// Accept quoted string too.
	got = DecimalAmount{}
	require.NoError(t, json.Unmarshal([]byte(`"1234.56"`), &got))
	assert.Equal(t, "1234.56", got.String())

	// A quoted string carrying JSON escape sequences must be decoded through
	// the standard library, not by stripping the outer quotes. The token here
	// spells the digit 1 as a Unicode escape (u+0031); a conformant encoder
	// may emit that and it must still decode to 1.23. byte(92) is the
	// backslash, written numerically so the source carries no literal escape.
	got = DecimalAmount{}
	escaped := string([]byte{'"', byte(92), 'u', '0', '0', '3', '1', '.', '2', '3', '"'})
	require.NoError(t, json.Unmarshal([]byte(escaped), &got))
	assert.Equal(t, "1.23", got.String())

	// null and "" decode to zero.
	got = DecimalAmountFromCoefficient(1, 0)
	require.NoError(t, json.Unmarshal([]byte("null"), &got))
	assert.True(t, got.IsZero())

	got = DecimalAmountFromCoefficient(1, 0)
	require.NoError(t, json.Unmarshal([]byte(`""`), &got))
	assert.True(t, got.IsZero())

	// Round-trip inside a struct.
	type wrap struct {
		Price DecimalAmount `json:"price"`
	}
	b, err := json.Marshal(wrap{Price: DecimalAmountFromCoefficient(99999, 2)})
	require.NoError(t, err)
	assert.JSONEq(t, `{"price":999.99}`, string(b))
}

func TestDecimalAmount_Text(t *testing.T) {
	a := DecimalAmountFromCoefficient(-5, 3)
	text, err := a.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "-0.005", string(text))

	var got DecimalAmount
	require.NoError(t, got.UnmarshalText(text))
	assert.Equal(t, a, got)

	got = DecimalAmountFromCoefficient(1, 0)
	require.NoError(t, got.UnmarshalText(nil))
	assert.True(t, got.IsZero())
}

func TestDecimalAmount_Binary(t *testing.T) {
	for _, a := range []DecimalAmount{
		DecimalAmountFromCoefficient(0, 0),
		DecimalAmountFromCoefficient(123456, 2),
		DecimalAmountFromCoefficient(-123456, 2),
		DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 18),
		DecimalAmountFromCoefficient(minDecimalAmountCoefficient, 18),
	} {
		data, err := a.MarshalBinary()
		require.NoError(t, err)
		assert.Len(t, data, 8)

		var got DecimalAmount
		require.NoError(t, got.UnmarshalBinary(data))
		assert.Equal(t, a, got, "binary round-trip of %s", a)
	}

	var got DecimalAmount
	assert.Error(t, got.UnmarshalBinary([]byte{1, 2, 3}))
	// A non-canonical scale (19..30 is never emitted) is rejected as corrupt.
	badScale := make([]byte, 8)
	badScale[7] = 19
	assert.Error(t, got.UnmarshalBinary(badScale))
	// An out-of-range coefficient (-2^58, one below the valid minimum) is
	// rejected so it can never break the Neg/Abs invariant. 0x80.. decodes to
	// scale 0, coefficient -2^58.
	assert.Error(t, got.UnmarshalBinary([]byte{0x80, 0, 0, 0, 0, 0, 0, 0}))
}

func TestDecimalAmount_SQL(t *testing.T) {
	a := DecimalAmountFromCoefficient(123456, 2)
	v, err := a.Value()
	require.NoError(t, err)
	assert.Equal(t, "1234.56", v)
	assert.IsType(t, driver.Value(""), v)

	cases := []struct {
		src  any
		want string
	}{
		{"1234.56", "1234.56"},
		{[]byte("1.234,56"), "1234.56"},
		{int64(42), "42"},
		{float64(12.5), "12.5"},
	}
	for _, c := range cases {
		var got DecimalAmount
		require.NoError(t, got.Scan(c.src), "scan %#v", c.src)
		assert.Equal(t, c.want, got.String(), "scan %#v", c.src)
	}

	var got DecimalAmount
	assert.Error(t, got.Scan(true))
	// An int64 outside the coefficient range is rejected rather than silently
	// truncated (the coefficient must round-trip Neg/Abs).
	assert.Error(t, got.Scan(int64(math.MaxInt64)))
}

func TestNullableDecimalAmount(t *testing.T) {
	// Non-null round-trips through JSON as a number.
	n := DecimalAmountFromCoefficient(123456, 2).Nullable()
	data, err := json.Marshal(n)
	require.NoError(t, err)
	assert.Equal(t, "1234.56", string(data))

	// Null marshals to JSON null.
	var null NullableDecimalAmount
	data, err = json.Marshal(null)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))

	// FromPtr with nil is null.
	assert.True(t, NullableDecimalAmountFromPtr(nil).IsNull())
	a := DecimalAmountFromCoefficient(1, 0)
	assert.True(t, NullableDecimalAmountFromPtr(&a).IsNotNull())

	// SQL Value/Scan through the nullable wrapper.
	v, err := n.Value()
	require.NoError(t, err)
	assert.Equal(t, "1234.56", v)

	var scanned NullableDecimalAmount
	require.NoError(t, scanned.Scan("9.99"))
	assert.True(t, scanned.IsNotNull())
	assert.Equal(t, "9.99", scanned.Get().String())

	require.NoError(t, scanned.Scan(nil))
	assert.True(t, scanned.IsNull())
}

func TestDecimalAmount_ScanString(t *testing.T) {
	var a DecimalAmount
	require.NoError(t, a.ScanString("1.234,56", true))
	assert.Equal(t, "1234.56", a.String())
	assert.Error(t, a.ScanString("not a number", false))
}

func TestDecimalAmountFrom(t *testing.T) {
	// Integers convert exactly with scale 0.
	assert.Equal(t, DecimalAmountFromCoefficient(42, 0), DecimalAmountFrom(42))
	assert.Equal(t, DecimalAmountFromCoefficient(-42, 0), DecimalAmountFrom(int8(-42)))
	assert.Equal(t, DecimalAmountFromCoefficient(65535, 0), DecimalAmountFrom(uint16(65535)))
	assert.Equal(t, DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0), DecimalAmountFrom(int64(maxDecimalAmountCoefficient)))
	// An integer outside the coefficient range maps to ±Inf instead of panicking.
	assert.True(t, DecimalAmountFrom(int64(math.MaxInt64)).IsInf(1))
	assert.True(t, DecimalAmountFrom(int64(math.MinInt64)).IsInf(-1))
	assert.True(t, DecimalAmountFrom(uint64(math.MaxUint64)).IsInf(1))

	// Floats convert via their shortest round-tripping decimal representation.
	assert.Equal(t, "1234.567", DecimalAmountFrom(1234.567).String())
	assert.Equal(t, "0.1", DecimalAmountFrom(0.1).String())
	assert.Equal(t, "0.5", DecimalAmountFrom(0.5).String())
	// float32 uses the float32 shortest representation, not the float64 one.
	assert.Equal(t, "0.1", DecimalAmountFrom(float32(0.1)).String())
	// Amount and Rate convert like float64.
	assert.Equal(t, "1234.56", DecimalAmountFrom(Amount(1234.56)).String())
	assert.Equal(t, "1.19", DecimalAmountFrom(Rate(1.19)).String())
	// Non-finite floats map to the sentinels.
	assert.True(t, DecimalAmountFrom(math.NaN()).IsNaN())
	assert.True(t, DecimalAmountFrom(math.Inf(-1)).IsInf(-1))
	// A float needing more than 18 fractional digits is rounded at scale 18.
	assert.True(t, DecimalAmountFrom(1e-19).IsZero())
	// A float too large for the coefficient range maps to +Inf.
	assert.True(t, DecimalAmountFrom(1e30).IsInf(1))
}

// TestDecimalAmount_MarshalJSONValidNumber guards that MarshalJSON always emits
// a token the standard library re-parses as the same number.
func TestDecimalAmount_MarshalJSONValidNumber(t *testing.T) {
	for _, a := range []DecimalAmount{
		DecimalAmountFromCoefficient(0, 0),
		DecimalAmountFromCoefficient(-5, 3),
		DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 2),
	} {
		data, err := json.Marshal(a)
		require.NoError(t, err)
		assert.True(t, json.Valid(data), "invalid JSON %q", data)
		assert.False(t, bytes.HasPrefix(data, []byte(`"`)), "should be a number, got %q", data)
	}
}

func TestDecimalAmount_nonFiniteNoPanic(t *testing.T) {
	nan := DecimalAmountNaN()
	posInf := DecimalAmountInf(1)
	negInf := DecimalAmountInf(-1)

	// Float()/Amount() must not panic on non-finite values (regression).
	assert.True(t, math.IsNaN(nan.Float()))
	assert.True(t, math.IsInf(posInf.Float(), 1))
	assert.True(t, math.IsInf(negInf.Float(), -1))
	assert.True(t, math.IsNaN(float64(nan.Amount())))
	assert.True(t, math.IsInf(float64(posInf.Amount()), 1))
	assert.True(t, math.IsInf(float64(CurrencyDecimalAmountEUR(posInf).CurrencyAmount().Amount), 1))

	// Rate methods on a non-finite receiver propagate instead of panicking.
	assert.True(t, posInf.MultipliedByRate(Rate(2)).IsInf(1))
	assert.True(t, negInf.DividedByRate(Rate(2)).IsInf(-1))
	assert.True(t, nan.Percentage(50).IsNaN())

	// fmt: non-finite tokens are space-padded (never zero-padded) and honor the
	// sign flag on +Inf, matching the standard library.
	assert.Equal(t, "     Inf", fmt.Sprintf("%08.2f", posInf))
	assert.Equal(t, "    -Inf", fmt.Sprintf("%8.2f", negInf))
	assert.Equal(t, "     NaN", fmt.Sprintf("%08.2f", nan))
	assert.Equal(t, "+Inf", fmt.Sprintf("%+f", posInf))
	assert.Equal(t, "Inf   ", fmt.Sprintf("%-6.2f", posInf))
	// The padded token still round-trips.
	rt, err := ParseDecimalAmount("Inf")
	require.NoError(t, err)
	assert.True(t, rt.IsInf(1))

	// Scan(float64) rounds a value that needs >18 decimals instead of erroring.
	var scanned DecimalAmount
	require.NoError(t, scanned.Scan(1e-19))
	assert.True(t, scanned.IsZero())
	require.NoError(t, scanned.Scan(12.5))
	assert.Equal(t, "12.5", scanned.String())
}

// floatToDecimalOracle is an independent big.Rat reference for
// decimalFromFloatRounded, used to validate the int128 implementation.
func floatToDecimalOracle(f float64, scale int, mode RoundingMode) DecimalAmount {
	switch {
	case math.IsNaN(f):
		return DecimalAmountNaN()
	case math.IsInf(f, 0):
		return decimalInf(math.Signbit(f))
	}
	r := new(big.Rat).SetFloat64(f)
	pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	r.Mul(r, new(big.Rat).SetInt(pow))
	num, den := r.Num(), r.Denom()
	quo, rem := new(big.Int), new(big.Int)
	quo.QuoRem(num, den, rem)
	negative := num.Sign() < 0
	if rem.Sign() != 0 {
		twiceRem := new(big.Int).Lsh(new(big.Int).Abs(rem), 1)
		if roundAwayFromZero(mode, negative, quo.Bit(0) == 1, twiceRem.Cmp(den)) {
			if negative {
				quo.Sub(quo, big.NewInt(1))
			} else {
				quo.Add(quo, big.NewInt(1))
			}
		}
	}
	if !quo.IsInt64() {
		return decimalInf(negative)
	}
	c := quo.Int64()
	if c < minDecimalAmountCoefficient || c > maxDecimalAmountCoefficient {
		return decimalInf(negative)
	}
	return packDecimalAmount(c, scale)
}

func TestDecimalAmount_FromFloatOracle(t *testing.T) {
	values := []float64{
		0, math.Copysign(0, -1), 1, -1, 0.5, -0.5, 0.25, 0.125, 0.0625,
		1.5, 2.5, 12.5, 0.375, 2.675, 1.005, 100.0 / 3, 1.0 / 3, 2.0 / 3,
		1234.56, -99.99, 0.1, 0.2, 0.3, 9.99999999, 123456789.123456,
		1e-9, 1e-15, 1e-18, 1e-19, 1e-20, 5e-10,
		1e6, 1e12, 1e15, 1e17, 2.8e17, 1e18, 1e19, 1e30,
		math.SmallestNonzeroFloat64, math.MaxFloat64,
		math.Nextafter(0.5, 1), math.Nextafter(0.5, 0),
	}
	modes := []RoundingMode{
		RoundHalfAwayFromZero, RoundHalfToEven, RoundHalfUp, RoundHalfDown,
		RoundDown, RoundUp, RoundFloor, RoundCeil,
	}
	rng := rand.New(rand.NewSource(1))
	for range 4000 {
		// Add random floats across a wide exponent range, both signs.
		values = append(values, math.Ldexp(rng.Float64()-0.5, rng.Intn(200)-100))
	}
	for _, f := range values {
		for scale := 0; scale <= MaxDecimalAmountScale; scale++ {
			for _, mode := range modes {
				want := floatToDecimalOracle(f, scale, mode)
				got, err := decimalFromFloatRounded(f, scale, mode)
				require.NoError(t, err)
				if want != got {
					t.Fatalf("fromFloat(%v, %d, %s) = %s (packed %d), want %s (packed %d)",
						f, scale, mode, got, got, want, want)
				}
			}
		}
	}
}

func TestDecimalAmount_CmpInt128(t *testing.T) {
	// Exercise the 128-bit cross-scale comparison path with signs and magnitudes.
	cases := []struct {
		a, b DecimalAmount
		want int
	}{
		{DecimalAmountFromCoefficient(150, 2), DecimalAmountFromCoefficient(15, 1), 0},    // 1.50 == 1.5
		{DecimalAmountFromCoefficient(151, 2), DecimalAmountFromCoefficient(15, 1), 1},    // 1.51 > 1.5
		{DecimalAmountFromCoefficient(-151, 2), DecimalAmountFromCoefficient(-15, 1), -1}, // -1.51 < -1.5
		{DecimalAmountFromCoefficient(-1, 0), DecimalAmountFromCoefficient(1, 5), -1},     // negative < positive
		{DecimalAmountFromCoefficient(0, 0), DecimalAmountFromCoefficient(0, 8), 0},       // zeros equal across scales
		{DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0), DecimalAmountFromCoefficient(1, 18), 1},
		{DecimalAmountFromCoefficient(minDecimalAmountCoefficient, 0), DecimalAmountFromCoefficient(-1, 18), -1},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, c.a.Cmp(c.b), "%s cmp %s", c.a, c.b)
		assert.Equal(t, -c.want, c.b.Cmp(c.a), "%s cmp %s (reversed)", c.b, c.a)
	}
}

func TestDecimalAmount_sentinels(t *testing.T) {
	nan := DecimalAmountNaN()
	posInf := DecimalAmountInf(1)
	negInf := DecimalAmountInf(-1)
	fin := DecimalAmountFromCoefficient(150, 2)

	// Predicates.
	for _, s := range []DecimalAmount{nan, posInf, negInf} {
		assert.False(t, s.IsFinite())
		assert.False(t, s.Valid())
		assert.False(t, s.IsZero())
	}
	assert.True(t, fin.IsFinite())
	assert.True(t, nan.IsNaN())
	assert.True(t, posInf.IsInf(1))
	assert.True(t, posInf.IsInf(0))
	assert.False(t, posInf.IsInf(-1))
	assert.True(t, negInf.IsInf(-1))
	assert.False(t, nan.IsInf(0))

	// Sign/Signbit.
	assert.Equal(t, 1, posInf.Sign())
	assert.Equal(t, -1, negInf.Sign())
	assert.Equal(t, 0, nan.Sign())
	assert.True(t, negInf.Signbit())

	// Neg/Abs.
	assert.True(t, posInf.Neg().IsInf(-1))
	assert.True(t, negInf.Abs().IsInf(1))
	assert.True(t, nan.Neg().IsNaN())
	assert.True(t, nan.Abs().IsNaN())

	// Strings.
	assert.Equal(t, "NaN", nan.String())
	assert.Equal(t, "Inf", posInf.String())
	assert.Equal(t, "-Inf", negInf.String())
	assert.Equal(t, "money.DecimalAmountInf(1)", fmt.Sprintf("%#v", posInf))
	assert.Equal(t, "Inf", fmt.Sprintf("%.2f", posInf))
	assert.Equal(t, "  -Inf", fmt.Sprintf("%6s", negInf))

	// Total order via Cmp: -Inf < finite < +Inf < NaN.
	assert.Equal(t, -1, negInf.Cmp(fin))
	assert.Equal(t, -1, fin.Cmp(posInf))
	assert.Equal(t, -1, posInf.Cmp(nan))
	assert.Equal(t, 0, nan.Cmp(DecimalAmountNaN()))
	assert.True(t, posInf.Equal(DecimalAmountInf(1)))

	// Propagation.
	assert.True(t, posInf.Add(fin).IsInf(1))
	assert.True(t, posInf.Add(negInf).IsNaN()) // +Inf + -Inf
	assert.True(t, nan.Add(fin).IsNaN())
	assert.True(t, posInf.Mul(DecimalAmountFromCoefficient(0, 0), RoundHalfAwayFromZero).IsNaN()) // Inf * 0
	assert.True(t, nan.Mul(fin, RoundHalfAwayFromZero).IsNaN())
	assert.True(t, posInf.Mul(negInf, RoundHalfAwayFromZero).IsInf(-1)) // +Inf * -Inf
	assert.True(t, negInf.Mul(fin, RoundHalfAwayFromZero).IsInf(-1))    // -Inf * positive finite
	assert.True(t, posInf.Div(posInf, RoundHalfAwayFromZero).IsNaN())   // Inf / Inf
	assert.True(t, fin.Div(posInf, RoundHalfAwayFromZero).IsZero())     // finite / Inf
	assert.True(t, negInf.Div(fin, RoundHalfAwayFromZero).IsInf(-1))    // -Inf / positive finite
	assert.True(t, nan.Div(fin, RoundHalfAwayFromZero).IsNaN())
	assert.True(t, nan.RoundToCents(RoundHalfAwayFromZero).IsNaN())

	// MulInt on non-finite: NaN propagates, Inf × 0 is NaN, sign follows n.
	assert.True(t, nan.MulInt(5).IsNaN())
	assert.True(t, posInf.MulInt(0).IsNaN())
	assert.True(t, posInf.MulInt(3).IsInf(1))
	assert.True(t, posInf.MulInt(-3).IsInf(-1))

	// JSON round-trips through quoted strings.
	for _, s := range []DecimalAmount{nan, posInf, negInf} {
		data, err := json.Marshal(s)
		require.NoError(t, err)
		assert.True(t, json.Valid(data), "invalid JSON %q", data)
		var got DecimalAmount
		require.NoError(t, json.Unmarshal(data, &got))
		assert.Equal(t, s, got, "JSON round-trip of %s", s)
	}

	// Binary round-trips.
	for _, s := range []DecimalAmount{nan, posInf, negInf} {
		data, err := s.MarshalBinary()
		require.NoError(t, err)
		var got DecimalAmount
		require.NoError(t, got.UnmarshalBinary(data))
		assert.Equal(t, s, got, "binary round-trip of %s", s)
	}

	// SQL uses PostgreSQL numeric literals and round-trips.
	for _, tc := range []struct {
		in  DecimalAmount
		sql string
	}{
		{posInf, "Infinity"},
		{negInf, "-Infinity"},
		{nan, "NaN"},
	} {
		v, err := tc.in.Value()
		require.NoError(t, err)
		assert.Equal(t, tc.sql, v)
		var scanned DecimalAmount
		require.NoError(t, scanned.Scan(tc.sql))
		assert.Equal(t, tc.in, scanned, "SQL round-trip of %s", tc.in)
	}
}

func TestDecimalAmount_accessors(t *testing.T) {
	assert.Equal(t, "42", DecimalAmountFrom(42).String())

	pos := DecimalAmountFromCoefficient(150, 2)
	neg := DecimalAmountFromCoefficient(-150, 2)
	zero := DecimalAmount{}

	assert.Equal(t, +1, pos.Sign())
	assert.Equal(t, -1, neg.Sign())
	assert.Equal(t, 0, zero.Sign())

	assert.False(t, pos.Signbit())
	assert.True(t, neg.Signbit())

	assert.Equal(t, pos, pos.Abs())
	assert.Equal(t, pos, neg.Abs())
	assert.Equal(t, neg, pos.Neg())

	assert.Equal(t, pos, *pos.Ptr())

	// JSONSchema allows a number (finite) or a string (non-finite tokens).
	schema := DecimalAmount{}.JSONSchema()
	require.Len(t, schema.OneOf, 2)
	assert.Equal(t, "number", schema.OneOf[0].Type)
	assert.Equal(t, "string", schema.OneOf[1].Type)
}

func TestDecimalAmount_AmountInterop(t *testing.T) {
	d := DecimalAmountFromCoefficient(123456, 2)
	assert.InDelta(t, 1234.56, float64(d.Amount()), 1e-9)

	// DecimalAmountFrom converts an Amount via its shortest exact decimal;
	// round explicitly for a fixed scale.
	got := DecimalAmountFrom(Amount(1234.567)).RoundToDecimals(2, RoundHalfAwayFromZero)
	assert.Equal(t, "1234.57", got.String())

	// Amount.DecimalAmount rounds the exact binary float64 value directly.
	assert.Equal(t, "1234.57", Amount(1234.567).DecimalAmount(2, RoundHalfAwayFromZero).String())
	// A non-finite Amount maps to the matching sentinel instead of panicking.
	assert.True(t, Amount(math.Inf(1)).DecimalAmount(2, RoundHalfAwayFromZero).IsInf(1))
	assert.True(t, Amount(math.NaN()).DecimalAmount(2, RoundHalfAwayFromZero).IsNaN())

	// Round-trip Amount -> DecimalAmount -> Amount.
	back := DecimalAmountFrom(d.Amount())
	assert.True(t, d.Equal(back))
}

func TestDecimalAmount_RateInterop(t *testing.T) {
	price := DecimalAmountFromCoefficient(10000, 2) // 100.00

	// The result is the shortest decimal that round-trips the float64
	// product exactly; round to the final precision explicitly at the end.
	product := price.MultipliedByRate(Rate(1.19))
	assert.Equal(t, "119.00", product.RoundToCents(RoundHalfAwayFromZero).String())
	// A float64 product carrying binary noise keeps that noise until rounded:
	// 0.10 * 3 in float64 is 0.30000000000000004.
	noisy := DecimalAmountFromCoefficient(10, 2).MultipliedByRate(Rate(3))
	assert.Equal(t, "0.30000000000000004", noisy.String())
	assert.Equal(t, "0.30", noisy.RoundToCents(RoundHalfAwayFromZero).String())
	// Reverse: 119.00 / 1.19 -> 100.00 after rounding to cents.
	assert.Equal(t, "100.00", DecimalAmountFromCoefficient(11900, 2).DividedByRate(Rate(1.19)).RoundToCents(RoundHalfAwayFromZero).String())
	// 19% of 100.00 -> 19.00 (exact in float64, shortest representation "19").
	assert.Equal(t, "19.00", price.Percentage(19).RoundToCents(RoundHalfAwayFromZero).String())
	// An exact float64 product converts without any rounding needed.
	assert.Equal(t, "250", price.MultipliedByRate(Rate(2.5)).String())

	// Dividing by a zero rate yields +Inf instead of panicking.
	assert.True(t, price.DividedByRate(Rate(0)).IsInf(1))
}

func TestDecimalAmount_CentAmount(t *testing.T) {
	tests := []struct {
		decimal  DecimalAmount
		rounding RoundingMode
		expected CentAmount
	}{
		{DecimalAmountFromCoefficient(12345, 2), RoundHalfAwayFromZero, 12345},
		{DecimalAmountFromCoefficient(123, 0), RoundHalfAwayFromZero, 12300},
		{DecimalAmountFromCoefficient(1235, 1), RoundHalfAwayFromZero, 12350},
		{DecimalAmountFromCoefficient(123455, 3), RoundHalfAwayFromZero, 12346},
		{DecimalAmountFromCoefficient(123455, 3), RoundHalfToEven, 12346},
		{DecimalAmountFromCoefficient(123465, 3), RoundHalfToEven, 12346},
		{DecimalAmountFromCoefficient(123455, 3), RoundDown, 12345},
		{DecimalAmountFromCoefficient(-123455, 3), RoundHalfAwayFromZero, -12346},
	}
	for _, test := range tests {
		t.Run(test.decimal.String(), func(t *testing.T) {
			cents, err := test.decimal.CentAmount(test.rounding)
			require.NoError(t, err)
			assert.Equal(t, test.expected, cents)
		})
	}

	// Non-finite and out of range values must fail loud
	// because CentAmount can't represent them.
	for _, decimal := range []DecimalAmount{
		DecimalAmountNaN(),
		DecimalAmountInf(1),
		DecimalAmountInf(-1),
		DecimalAmountFromCoefficient(maxDecimalAmountCoefficient, 0),
	} {
		t.Run(decimal.String(), func(t *testing.T) {
			_, err := decimal.CentAmount(RoundHalfAwayFromZero)
			assert.Error(t, err)
		})
	}
}
