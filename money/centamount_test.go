package money

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCentAmount(t *testing.T) {
	// Parsing must be exact without a float64 round-trip,
	// that's the whole point of the integer based type.
	valid := map[string]CentAmount{
		"0":               0,
		"22.00":           2200,
		"123.45":          12345,
		"123,45":          12345,
		"0,99":            99,
		"1.234,56":        123456,
		"1,234.56":        123456,
		"100.234.567,89":  10023456789,
		"-123.45":         -12345,
		"-0,99":           -99,
		"-100,234,567.89": -10023456789,

		// Values with more than 2 decimals round half away from zero
		"0.004":  0,
		"0.005":  1,
		"0.006":  1,
		"-0.005": -1,
		"1.0049": 100,
		"1.0050": 101,
	}
	for str, expected := range valid {
		t.Run(str, func(t *testing.T) {
			parsed, err := ParseCentAmount(str, RoundHalfAwayFromZero)
			require.NoError(t, err)
			assert.Equal(t, expected, parsed)
		})
	}

	// CentAmount has no non-finite states, so the non-finite
	// literals that ParseAmount accepts must be rejected here.
	for _, str := range []string{"NaN", "Inf", "+Inf", "-Inf", "Infinity", "EUR 123", "1,234,56"} {
		t.Run(str, func(t *testing.T) {
			_, err := ParseCentAmount(str, RoundHalfAwayFromZero)
			assert.Error(t, err)
		})
	}

	t.Run("acceptedDecimals", func(t *testing.T) {
		parsed, err := ParseCentAmount("123.45", RoundHalfAwayFromZero, 2)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(12345), parsed)

		_, err = ParseCentAmount("123.4", RoundHalfAwayFromZero, 0, 2)
		assert.Error(t, err)
	})

	t.Run("rounding mode is honored", func(t *testing.T) {
		// The mode has to reach the cent conversion, so the same source
		// string must produce different cents per mode.
		// Every RoundingMode the package defines must reach the direct
		// decimal-digit conversion, not just the default.
		modes := map[RoundingMode]CentAmount{
			RoundHalfAwayFromZero: 1,
			RoundHalfToEven:       0,
			RoundHalfUp:           1,
			RoundHalfDown:         0,
			RoundDown:             0,
			RoundUp:               1,
			RoundCeil:             1,
			RoundFloor:            0,
		}
		for mode, expected := range modes {
			parsed, err := ParseCentAmount("0.005", mode)
			require.NoError(t, err, "mode %s", mode)
			assert.Equal(t, expected, parsed, "mode %s", mode)
		}
	})

	t.Run("beyond the CentAmount range", func(t *testing.T) {
		// Only the CentAmount range itself may reject, not the narrower
		// DecimalAmount coefficient range that parsing used to borrow.
		_, err := ParseCentAmount("92233720368547758.08", RoundHalfAwayFromZero)
		assert.Error(t, err)
		_, err = ParseCentAmount("-92233720368547758.09", RoundHalfAwayFromZero)
		assert.Error(t, err)

		// Rounding up past MaxCentAmount is out of range,
		// rounding down to it is not
		cents, err := ParseCentAmount("92233720368547758.07", RoundHalfAwayFromZero)
		require.NoError(t, err)
		assert.Equal(t, MaxCentAmount, cents)
		_, err = ParseCentAmount("92233720368547758.075", RoundHalfAwayFromZero)
		assert.Error(t, err)
		cents, err = ParseCentAmount("92233720368547758.075", RoundDown)
		require.NoError(t, err)
		assert.Equal(t, MaxCentAmount, cents)
		cents, err = ParseCentAmount("-92233720368547758.074", RoundHalfAwayFromZero)
		require.NoError(t, err)
		assert.Equal(t, MinCentAmount, cents)

		// A cent count the DecimalAmount coefficient cannot hold but
		// CentAmount can: 3e17 cents is above maxDecimalAmountCoefficient
		// and well below MaxCentAmount.
		cents, err = ParseCentAmount("3000000000000000.00", RoundHalfAwayFromZero)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(300_000_000_000_000_000), cents)
		assert.Greater(t, int64(cents), int64(maxDecimalAmountCoefficient))
	})

	t.Run("more decimals than DecimalAmount can scale", func(t *testing.T) {
		// The 18 decimal place cap is a DecimalAmount limit; extra decimals
		// only feed the rounding here.
		cents, err := ParseCentAmount("1.0000000000000000000000000", RoundHalfAwayFromZero)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(100), cents)

		cents, err = ParseCentAmount("0.005000000000000000000000001", RoundHalfAwayFromZero)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(1), cents, "just above half must round up")

		cents, err = ParseCentAmount("0.004999999999999999999999999", RoundHalfAwayFromZero)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(0), cents, "just below half must round down")
	})
}

func TestCentAmount_Cents(t *testing.T) {
	assert.Equal(t, int64(12345), CentAmount(12345).Cents())
	assert.Equal(t, int64(-1), CentAmount(-1).Cents())
	assert.Equal(t, int64(math.MinInt64), CentAmount(math.MinInt64).Cents())
}

func TestCentAmount_conversionBoundaries(t *testing.T) {
	// The clamp is an inequality, so the exact edge has to be pinned
	// from both sides or an off-by-one would go unnoticed.
	const maxCents = CentAmount(maxDecimalAmountCoefficient)

	assert.Equal(t, NewDecimalAmount(int64(maxCents), 2), maxCents.DecimalAmount())
	assert.Equal(t, NewDecimalAmount(-int64(maxCents), 2), (-maxCents).DecimalAmount())
	assert.True(t, (maxCents + 1).DecimalAmount().IsInf(1))
	assert.True(t, (-maxCents - 1).DecimalAmount().IsInf(-1))

	// The matching DecimalAmount -> CentAmount edge must succeed
	cents, err := NewDecimalAmount(int64(maxCents), 2).CentAmount(RoundHalfAwayFromZero)
	require.NoError(t, err)
	assert.Equal(t, maxCents, cents)
}

func TestCentAmount_int64Extremes(t *testing.T) {
	// Nothing may overflow or panic at the edges of the underlying int64.
	assert.Equal(t, CentAmount(9223372036854775800), CentAmount(math.MaxInt64).RoundToInt())
	assert.Equal(t, CentAmount(-9223372036854775800), CentAmount(math.MinInt64).RoundToInt())

	assert.Equal(t, "-92,233,720,368,547,758.08", CentAmount(math.MinInt64).FormatSep(',', '.'))

	for _, amount := range []CentAmount{math.MinInt64, math.MaxInt64} {
		for count := 2; count <= 3; count++ {
			var sum CentAmount
			for _, part := range amount.SplitEqually(count) {
				sum += part
			}
			assert.Equal(t, amount, sum, "sum of SplitEqually(%d) of %d", count, int64(amount))
		}
	}

	// Negating math.MinInt64 wraps, so Abs can't make it positive.
	// Documented here so the wart is a decision, not a surprise.
	assert.Equal(t, CentAmount(math.MinInt64), CentAmount(math.MinInt64).Abs())
}

func TestCentAmount_SplitProportionally_weightOverflow(t *testing.T) {
	// Summing the weight magnitudes in a plain uint64 wraps: two MinInt64
	// weights wrapped to a zero total and dumped everything in the last
	// element, and MaxInt64 weights wrapped to a tiny total that made
	// bits.Div64 panic. Both must now split proportionally.
	assert.Equal(t,
		[]CentAmount{5000, 5000},
		CentAmount(10000).SplitProportionally([]CentAmount{math.MinInt64, math.MinInt64}),
	)

	result := CentAmount(10000).SplitProportionally([]CentAmount{math.MaxInt64, math.MaxInt64, 3})
	assert.Equal(t, []CentAmount{5000, 5000, 0}, result)

	// The exact-sum property must survive the shift
	for _, weights := range [][]CentAmount{
		{math.MinInt64, math.MinInt64},
		{math.MaxInt64, math.MaxInt64, 3},
		{math.MaxInt64, 1, math.MaxInt64, 1},
	} {
		var sum CentAmount
		for _, part := range CentAmount(9999).SplitProportionally(weights) {
			sum += part
		}
		assert.Equal(t, CentAmount(9999), sum, "sum for weights %v", weights)
	}
}

func TestCentAmount_SplitProportionally_roundingTie(t *testing.T) {
	// An exact half has to round away from zero, matching every other
	// rounding default in the package.
	assert.Equal(t, []CentAmount{1, 0}, CentAmount(1).SplitProportionally([]CentAmount{1, 1}))
	assert.Equal(t, []CentAmount{-1, 0}, CentAmount(-1).SplitProportionally([]CentAmount{1, 1}))
}

func TestNullableCentAmount_SQL(t *testing.T) {
	// The README claims CentAmount rides on the native int64 handling
	// of database/sql, so the nullable wrapper must round-trip.
	value, err := NullableCentAmountFrom(123).Value()
	require.NoError(t, err)
	assert.Equal(t, int64(123), value)

	var amount NullableCentAmount
	require.NoError(t, amount.Scan(int64(123)))
	assert.Equal(t, CentAmount(123), amount.Get())

	require.NoError(t, amount.Scan(nil))
	assert.True(t, amount.IsNull())
}

func TestCentAmount_String(t *testing.T) {
	// Always exactly two decimals, so a cent amount reads
	// like the money value it represents, not like its cent count.
	table := map[CentAmount]string{
		0:           "0.00",
		1:           "0.01",
		-1:          "-0.01",
		9:           "0.09",
		99:          "0.99",
		100:         "1.00",
		12345:       "123.45",
		-12345:      "-123.45",
		10023456789: "100234567.89",
	}
	for amount, expected := range table {
		assert.Equal(t, expected, amount.String())
	}

	// math.MinInt64 must not overflow while taking the magnitude
	assert.Equal(t, "-92233720368547758.08", CentAmount(math.MinInt64).String())
	assert.Equal(t, "92233720368547758.07", CentAmount(math.MaxInt64).String())
}

func TestCentAmount_FormatSep(t *testing.T) {
	assert.Equal(t, "1.234.567,89", CentAmount(123456789).FormatSep('.', ','))
	assert.Equal(t, "1,234,567.89", CentAmount(123456789).FormatSep(',', '.'))
	assert.Equal(t, "-1.234.567,89", CentAmount(-123456789).FormatSep('.', ','))
	assert.Equal(t, "123456789.00", CentAmount(12345678900).FormatSep(0, 0)) // Zero decimalSep defaults to '.'
	assert.Equal(t, "0,05", CentAmount(5).FormatSep('.', ','))
}

func TestCentAmount_GoString(t *testing.T) {
	assert.Equal(t, "money.CentAmount(12345)", CentAmount(12345).GoString())
	assert.Equal(t, "money.CentAmount(-1)", CentAmount(-1).GoString())
}

func TestCentAmount_Amount(t *testing.T) {
	// Conversion in both directions must round-trip for every
	// amount that is already an exact cent value.
	for cents := CentAmount(-1000); cents <= 1000; cents++ {
		assert.Equal(t, cents, cents.Amount().CentAmount(RoundHalfAwayFromZero), "round-trip of %s", cents)
	}
	assert.Equal(t, Amount(123.45), CentAmount(12345).Amount())
	assert.Equal(t, Amount(-0.01), CentAmount(-1).Amount())
}

func TestCentAmount_DecimalAmount(t *testing.T) {
	assert.Equal(t, NewDecimalAmount(12345, 2), CentAmount(12345).DecimalAmount())
	assert.Equal(t, NewDecimalAmount(-1, 2), CentAmount(-1).DecimalAmount())
	assert.Equal(t, "123.45", CentAmount(12345).DecimalAmount().String())

	// Beyond the DecimalAmount coefficient range the conversion
	// must saturate like every other DecimalAmount overflow
	// instead of panicking in packDecimalAmount.
	assert.True(t, CentAmount(math.MaxInt64).DecimalAmount().IsInf(1))
	assert.True(t, CentAmount(math.MinInt64).DecimalAmount().IsInf(-1))
}

func TestCentAmount_RoundToInt(t *testing.T) {
	table := map[CentAmount]CentAmount{
		0:      0,
		49:     0,
		50:     100,
		149:    100,
		150:    200,
		-49:    0,
		-50:    -100,
		-150:   -200,
		123456: 123500,
	}
	for amount, expected := range table {
		assert.Equal(t, expected, amount.RoundToInt(), "RoundToInt of %s", amount)
	}
}

func TestCentAmount_WithinOneCent(t *testing.T) {
	assert.True(t, CentAmount(100).WithinOneCent(100))
	assert.True(t, CentAmount(100).WithinOneCent(101))
	assert.True(t, CentAmount(100).WithinOneCent(99))
	assert.False(t, CentAmount(100).WithinOneCent(102))
	assert.False(t, CentAmount(100).WithinOneCent(98))
}

func TestCentAmount_signs(t *testing.T) {
	assert.Equal(t, -1, CentAmount(-1).Sign())
	assert.Equal(t, 0, CentAmount(0).Sign())
	assert.Equal(t, +1, CentAmount(1).Sign())

	assert.True(t, CentAmount(0).IsZero())
	assert.False(t, CentAmount(1).IsZero())

	assert.True(t, CentAmount(-1).Signbit())
	assert.False(t, CentAmount(0).Signbit())

	assert.Equal(t, CentAmount(5), CentAmount(-5).Abs())
	assert.Equal(t, CentAmount(5), CentAmount(5).Abs())

	assert.Equal(t, CentAmount(-5), CentAmount(5).Copysign(-1))
	assert.Equal(t, CentAmount(5), CentAmount(-5).Copysign(+1))
	assert.Equal(t, CentAmount(0), CentAmount(0).Copysign(-1)) // No negative zero for int64

	assert.Equal(t, CentAmount(-5), CentAmount(5).Inverted())
	amount := CentAmount(5)
	amount.Invert()
	assert.Equal(t, CentAmount(-5), amount)

	assert.Equal(t, CentAmount(5), CentAmount(-5).WithPosSign(true))
	assert.Equal(t, CentAmount(-5), CentAmount(5).WithPosSign(false))
	assert.Equal(t, CentAmount(-5), CentAmount(5).WithNegSign(true))
	assert.Equal(t, CentAmount(5), CentAmount(-5).WithNegSign(false))
}

func TestCentAmount_rateMethods(t *testing.T) {
	assert.Equal(t, CentAmount(23800), CentAmount(20000).MultipliedByRate(1.19))
	assert.Equal(t, CentAmount(20000), CentAmount(23800).DividedByRate(1.19))
	assert.Equal(t, CentAmount(3800), CentAmount(20000).Percentage(19))

	// Results are rounded to whole cents half away from zero
	assert.Equal(t, CentAmount(2), CentAmount(1).MultipliedByRate(1.5))
	assert.Equal(t, CentAmount(-2), CentAmount(-1).MultipliedByRate(1.5))

	// A zero or NaN rate must not yield a platform dependent int64
	assert.Equal(t, MaxCentAmount, CentAmount(1).DividedByRate(0))
	assert.Equal(t, MinCentAmount, CentAmount(-1).DividedByRate(0))
	assert.Equal(t, CentAmount(0), CentAmount(1).MultipliedByRate(Rate(math.NaN())))
}

// TestCentAmount_rateMethodsRoundShortestDecimal protects the rounding
// convention of the module, that a float is rounded by its shortest decimal
// representation half away from zero, for the CentAmount rate methods.
// They rounded the binary float64 result instead, so a tie like the 14.5
// cents of CentAmount(100).MultipliedByRate(0.145), whose float64 product is
// 14.499999999999998, was rounded down to 14, while Percentage(14.5) and
// Amount(1).MultipliedByRate(0.145).CentAmount gave 15 for the same price.
// The rate is taken as its shortest decimal, 0.145 and not
// 0.14499999999999999, and the result is rounded exactly once.
func TestCentAmount_rateMethodsRoundShortestDecimal(t *testing.T) {
	tests := []struct {
		name string
		got  CentAmount
		want CentAmount
	}{
		{name: "100 * 0.145 = 14.5", got: CentAmount(100).MultipliedByRate(0.145), want: 15},
		{name: "-100 * 0.145 = -14.5", got: CentAmount(-100).MultipliedByRate(0.145), want: -15},
		{name: "300 * 0.145 = 43.5", got: CentAmount(300).MultipliedByRate(0.145), want: 44},
		{name: "10000 * 0.00145 = 14.5", got: CentAmount(10000).MultipliedByRate(0.00145), want: 15},
		{name: "1 * 0.5 = 0.5", got: CentAmount(1).MultipliedByRate(0.5), want: 1},
		{name: "-1 * 0.5 = -0.5", got: CentAmount(-1).MultipliedByRate(0.5), want: -1},
		{name: "700 * 0.0035 = 2.45", got: CentAmount(700).MultipliedByRate(0.0035), want: 2},
		{name: "17 / 0.272 = 62.5", got: CentAmount(17).DividedByRate(0.272), want: 63},
		{name: "-17 / 0.272 = -62.5", got: CentAmount(-17).DividedByRate(0.272), want: -63},
		{name: "1 / 3 = 0.333", got: CentAmount(1).DividedByRate(3), want: 0},
		{name: "2 / 3 = 0.667", got: CentAmount(2).DividedByRate(3), want: 1},
		{name: "100 * 14.5% = 14.5", got: CentAmount(100).Percentage(14.5), want: 15},
		{name: "2500 * 1.14% = 28.5", got: CentAmount(2500).Percentage(1.14), want: 29},
		{name: "-2500 * 1.14% = -28.5", got: CentAmount(-2500).Percentage(1.14), want: -29},
		{name: "1250 * 2.28% = 28.5", got: CentAmount(1250).Percentage(2.28), want: 29},
		{name: "like Amount", got: CentAmount(100).MultipliedByRate(0.145), want: Amount(1).MultipliedByRate(0.145).CentAmount(RoundHalfAwayFromZero)},
		{name: "like Percentage", got: CentAmount(100).MultipliedByRate(0.145), want: CentAmount(100).Percentage(14.5)},
		{name: "MaxCentAmount * 2 clamps", got: MaxCentAmount.MultipliedByRate(2), want: MaxCentAmount},
		{name: "MinCentAmount * 2 clamps", got: MinCentAmount.MultipliedByRate(2), want: MinCentAmount},
		{name: "MaxCentAmount * -1", got: MaxCentAmount.MultipliedByRate(-1), want: -MaxCentAmount},
		{name: "1 * 1e300 clamps", got: CentAmount(1).MultipliedByRate(1e300), want: MaxCentAmount},
		{name: "1 * +Inf clamps", got: CentAmount(1).MultipliedByRate(Rate(math.Inf(1))), want: MaxCentAmount},
		{name: "0 * +Inf is NaN", got: CentAmount(0).MultipliedByRate(Rate(math.Inf(1))), want: 0},
		{name: "1 / +Inf", got: CentAmount(1).DividedByRate(Rate(math.Inf(1))), want: 0},
		{name: "0 / 0 is NaN", got: CentAmount(0).DividedByRate(0), want: 0},
		{name: "-1 / -0", got: CentAmount(-1).DividedByRate(Rate(math.Copysign(0, -1))), want: MaxCentAmount},
		{name: "1 * NaN%", got: CentAmount(1).Percentage(math.NaN()), want: 0},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.got, tt.name)
	}
}

// TestCentAmount_rateFastPathMatchesExact protects that the float64 fast path
// of MultipliedByRate, DividedByRate and Percentage never changes a result.
// It is only allowed to skip the exact big.Rat arithmetic where it provably
// rounds to the same cents, so every input must round exactly like the exact
// arithmetic with the shortest decimal of the rate does, most of all the
// decimal ties and their float64 neighbors where float64 arithmetic alone
// gets it wrong, and the amounts beyond 2^53 cents that float64 can't hold.
func TestCentAmount_rateFastPathMatchesExact(t *testing.T) {
	check := func(c CentAmount, rate float64) {
		t.Helper()
		if got, want := c.MultipliedByRate(Rate(rate)), c.multipliedByRateExact(Rate(rate)); got != want {
			t.Fatalf("CentAmount(%d).MultipliedByRate(%v) = %d, exact = %d", c, rate, got, want)
		}
		if got, want := c.DividedByRate(Rate(rate)), c.dividedByRateExact(Rate(rate)); got != want {
			t.Fatalf("CentAmount(%d).DividedByRate(%v) = %d, exact = %d", c, rate, got, want)
		}
		if got, want := c.Percentage(rate), c.percentageExact(rate); got != want {
			t.Fatalf("CentAmount(%d).Percentage(%v) = %d, exact = %d", c, rate, got, want)
		}
	}
	// checkSigns checks c with both signs times rate and its float64 neighbors
	checkSigns := func(c CentAmount, rate float64) {
		t.Helper()
		for _, r := range []float64{rate, math.Nextafter(rate, 0), math.Nextafter(rate, math.Inf(1))} {
			check(c, r)
			check(-c, r)
		}
	}

	targetedCents := []CentAmount{
		0, 1, 7, 100, 1000, 12345, 999999, 1e15,
		1<<53 - 1, 1 << 53, 1<<53 + 1, MaxCentAmount,
	}
	targetedRates := []float64{
		0, 0.5, 1.5, 0.145, 1.005, 2.675, 0.0025, 0.0035, 0.272, 14.5, 1.14,
		1.0845, 0.92137, 0.1 + 0.2, 3, 1e-300, 5e-324, 1e300, math.MaxFloat64,
		math.Inf(1), math.NaN(),
	}
	for _, c := range targetedCents {
		for _, rate := range targetedRates {
			checkSigns(c, rate)
			checkSigns(c, -rate)
		}
	}
	check(MinCentAmount, 0.5)
	check(MinCentAmount, -1)

	rng := rand.New(rand.NewPCG(1, 2))
	for range 2000 {
		// A rate of up to 15 digits ending in a 5 at the decimal place
		// decimals, like 0.145 or 1.005, is a tie of its shortest decimal
		// multiplied by every odd multiple of 10^(decimals-1) cents, and
		// a percent like it of every odd multiple of 10^(decimals+1) cents.
		decimals := 1 + rng.IntN(12)
		digits := rng.Int64N(int64(math.Pow10(rng.IntN(16 - decimals))))
		tie, err := strconv.ParseFloat(fmt.Sprintf("%d5e-%d", digits, decimals), 64)
		require.NoError(t, err)
		odd := 2*rng.Int64N(1000) + 1
		checkSigns(CentAmount(odd*int64(math.Pow10(decimals-1))), tie)
		checkSigns(CentAmount(odd*int64(math.Pow10(decimals+1))), tie)

		// c divided by the rate 2c/5^a, which is exactly 2c·2^a/10^a,
		// is the tie 5^a/2
		c := 1 + rng.Int64N(1e9)
		a := 1 + rng.IntN(8)
		rate, err := strconv.ParseFloat(fmt.Sprintf("%de-%d", 2*c<<a, a), 64)
		require.NoError(t, err)
		checkSigns(CentAmount(c), rate)
	}

	for range rateSweepSize() {
		check(randomCents(rng), randomRate(rng))
	}
}

// TestCentAmount_rateFastPathTaken protects the reason for the fast path:
// converting amounts with real exchange rates and tax percentages must not
// take the exact big.Rat arithmetic, which is an order of magnitude slower.
// Only results within a tiny relative distance of a tie may take it.
func TestCentAmount_rateFastPathTaken(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	exact := 0
	for range 100_000 {
		c := CentAmount(rng.Int64N(2e10) - 1e10)
		rate := []float64{1.0845, 0.92137, 7.4602, 25.317, 0.0067}[rng.IntN(5)]
		percent := []float64{19, 20, 7.7, 2.5, 33.333}[rng.IntN(5)]
		for _, result := range []float64{float64(c) * rate, float64(c) / rate, float64(c) * percent / 100} {
			if _, ok := roundCentsFast(c, result); !ok {
				exact++
			}
		}
	}
	// Only the exact arithmetic can round exact ties, like 20 cents times
	// 2.5% or 50 cents times 19%. They are 760 of the 300,000 results here.
	assert.Less(t, exact, 1000, "results that took the exact path")
}

// randomCents returns random cents for the rate sweeps,
// from small amounts up to beyond 2^53 cents.
func randomCents(rng *rand.Rand) CentAmount {
	var c CentAmount
	switch rng.IntN(4) {
	case 0:
		c = CentAmount(rng.Int64N(1e5))
	case 1:
		c = CentAmount(rng.Int64N(1e13))
	case 2:
		c = CentAmount(1<<53 - 100 + rng.Int64N(200))
	default:
		c = CentAmount(rng.Int64N(int64(MaxCentAmount)))
	}
	if rng.IntN(2) == 0 {
		c = -c
	}
	return c
}

// randomRate returns a random finite rate or percent for the rate sweeps:
// a short decimal like a stored rate, the product of one with a random
// factor, which has a long binary fraction, the float64 neighbor of one,
// or any mantissa with an exponent from -80 to 80. The extreme exponents
// are among the targeted rates, here they would only slow down the
// big.Rat arithmetic.
func randomRate(rng *rand.Rand) float64 {
	decimal := float64(rng.Int64N(int64(math.Pow10(1+rng.IntN(15))))) / math.Pow10(rng.IntN(13))
	if rng.IntN(2) == 0 {
		decimal = -decimal
	}
	switch rng.IntN(4) {
	case 0:
		return decimal
	case 1:
		return decimal * (rng.Float64() * 2)
	case 2:
		return math.Nextafter(decimal, math.Inf(1-2*rng.IntN(2)))
	default:
		return math.Ldexp(1+rng.Float64(), rng.IntN(161)-80)
	}
}

// rateSweepSize returns the number of random inputs
// for the rate sweeps, fewer with -short.
func rateSweepSize() int {
	if testing.Short() {
		return 10_000
	}
	return 100_000
}

func TestCentAmount_SplitEqually(t *testing.T) {
	type input struct {
		amount CentAmount
		count  int
	}
	data := map[input][]CentAmount{
		{amount: 10000, count: 0}:  nil,
		{amount: 10000, count: 1}:  {10000},
		{amount: 10000, count: 3}:  {3333, 3333, 3334},
		{amount: 1, count: 5}:      {0, 0, 0, 0, 1},
		{amount: 5, count: 5}:      {1, 1, 1, 1, 1},
		{amount: 100, count: 17}:   {5, 5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6},
		{amount: -10000, count: 3}: {-3333, -3333, -3334},
		{amount: -1, count: 5}:     {0, 0, 0, 0, -1},
		{amount: -100, count: 17}:  {-5, -5, -6, -6, -6, -6, -6, -6, -6, -6, -6, -6, -6, -6, -6, -6, -6},
	}
	for input, expected := range data {
		result := input.amount.SplitEqually(input.count)
		assert.Equal(t, expected, result, "SplitEqually(%d) of %s", input.count, input.amount)
	}

	// The defining property: the parts must sum up to the initial amount exactly
	for amount := CentAmount(-500); amount <= 500; amount++ {
		for count := 1; count <= 17; count++ {
			var sum CentAmount
			for _, part := range amount.SplitEqually(count) {
				sum += part
			}
			assert.Equal(t, amount, sum, "sum of SplitEqually(%d) of %s", count, amount)
		}
	}
}

func TestCentAmount_SplitProportionally(t *testing.T) {
	data := []struct {
		amount   CentAmount
		weights  []CentAmount
		expected []CentAmount
	}{
		{0, nil, nil},
		{0, []CentAmount{1}, []CentAmount{0}},
		{10000, []CentAmount{1}, []CentAmount{10000}},
		{10000, []CentAmount{1000}, []CentAmount{10000}},
		{10000, []CentAmount{1, 2}, []CentAmount{3333, 6667}},
		{10000, []CentAmount{1, 2, 3}, []CentAmount{1667, 3333, 5000}},
		{10000, []CentAmount{1, 2, 3, 4}, []CentAmount{1000, 2000, 3000, 4000}},
		// Only the weight magnitudes matter, the sign comes from the amount
		{10000, []CentAmount{-1, +2, -3, +4}, []CentAmount{1000, 2000, 3000, 4000}},
		{10000, []CentAmount{1100, 1700, 3700}, []CentAmount{1692, 2616, 5692}},
		{100, []CentAmount{1100, 1700, 3700}, []CentAmount{17, 26, 57}},
		{10000, []CentAmount{1, 0, 1}, []CentAmount{5000, 0, 5000}},
		{25000, []CentAmount{33333, 8333, 1}, []CentAmount{20000, 5000, 0}},

		{-10000, []CentAmount{1, 2}, []CentAmount{-3333, -6667}},
		{-10000, []CentAmount{-1, +2, -3, +4}, []CentAmount{-1000, -2000, -3000, -4000}},
		{-100, []CentAmount{1100, 1700, 3700}, []CentAmount{-17, -26, -57}},

		// All weights zero: nothing to distribute proportionally,
		// so the last element absorbs the whole amount.
		{10000, []CentAmount{0, 0, 0}, []CentAmount{0, 0, 10000}},
	}
	for _, test := range data {
		result := test.amount.SplitProportionally(test.weights)
		assert.Equal(t, test.expected, result, "SplitProportionally(%v) of %s", test.weights, test.amount)
	}

	// The defining property: the parts must sum up to the initial amount exactly
	weights := []CentAmount{11, 17, 37, 0, 4711}
	for amount := CentAmount(-500); amount <= 500; amount++ {
		var sum CentAmount
		for _, part := range amount.SplitProportionally(weights) {
			sum += part
		}
		assert.Equal(t, amount, sum, "sum of SplitProportionally of %s", amount)
	}

	// Weights and amount large enough that amount*weight
	// overflows int64 must still split exactly.
	large := CentAmount(4_000_000_000)
	largeWeights := []CentAmount{3_000_000_000, 6_000_000_000}
	assert.Equal(t, []CentAmount{1_333_333_333, 2_666_666_667}, large.SplitProportionally(largeWeights))
}

func TestCentAmount_pointerHelpers(t *testing.T) {
	assert.Equal(t, CentAmount(123), *new(CentAmount(123)))
	assert.Equal(t, CentAmount(123), *CentAmount(123).Ptr())

	assert.Equal(t, CentAmount(9), CentAmountFromPtr(nil, 9))
	assert.Equal(t, CentAmount(123), CentAmountFromPtr(new(CentAmount(123)), 9))

	var nilPtr *CentAmount
	assert.Equal(t, "n/a", nilPtr.StringOr("n/a"))
	assert.Equal(t, "1.23", new(CentAmount(123)).StringOr("n/a"))
	assert.Equal(t, int64(9), nilPtr.CentsOr(9))
	assert.Equal(t, int64(123), new(CentAmount(123)).CentsOr(9))
	assert.Equal(t, CentAmount(9), nilPtr.CentAmountOr(9))
	assert.Equal(t, CentAmount(123), new(CentAmount(123)).CentAmountOr(9))

	assert.True(t, nilPtr.Equal(nil))
	assert.False(t, nilPtr.Equal(new(CentAmount(0))))
	assert.True(t, new(CentAmount(123)).Equal(new(CentAmount(123))))
	assert.False(t, new(CentAmount(123)).Equal(new(CentAmount(124))))
}

func TestCentAmount_ScanString(t *testing.T) {
	var amount CentAmount
	require.NoError(t, amount.ScanString("1.234,56", false))
	assert.Equal(t, CentAmount(123456), amount)

	assert.Error(t, amount.ScanString("not a number", true))
	assert.Equal(t, CentAmount(123456), amount, "value must not change on error")
}

func TestCentAmount_JSON(t *testing.T) {
	// CentAmount marshals as its plain cent count, matching the
	// underlying int64 rather than the decimal String representation.
	marshalled, err := json.Marshal(CentAmount(12345))
	require.NoError(t, err)
	assert.Equal(t, `12345`, string(marshalled))

	var amount CentAmount
	require.NoError(t, json.Unmarshal([]byte(`12345`), &amount))
	assert.Equal(t, CentAmount(12345), amount)

	// A fractional JSON number is not a cent count and must be rejected
	assert.Error(t, json.Unmarshal([]byte(`123.45`), &amount))
}

func TestNullableCentAmount(t *testing.T) {
	assert.True(t, NullableCentAmountFromPtr(nil).IsNull())
	assert.Equal(t, CentAmount(123), NullableCentAmountFrom(123).Get())
	assert.Equal(t, CentAmount(123), NullableCentAmountFromPtr(new(CentAmount(123))).Get())
}

func TestCentAmount_SplitEqually_noSignFlip(t *testing.T) {
	// Handing the whole remainder to one amount used to produce
	// 199 parts of +0.01 and a last part of -0.99: an ordinary input
	// that charged one participant instead of paying them.
	parts := CentAmount(100).SplitEqually(200)
	var sum CentAmount
	for _, part := range parts {
		assert.False(t, part.Signbit(), "no part of a positive amount may be negative")
		sum += part
	}
	assert.Equal(t, CentAmount(100), sum)

	// No two amounts may differ by more than one cent
	for _, parts := range [][]CentAmount{
		CentAmount(1500).SplitEqually(1000),
		CentAmount(-1500).SplitEqually(1000),
		CentAmount(100).SplitEqually(17),
	} {
		lowest, highest := parts[0], parts[0]
		for _, part := range parts {
			lowest, highest = min(lowest, part), max(highest, part)
		}
		assert.LessOrEqual(t, int64(highest-lowest), int64(1), "spread of %v", parts)
	}
}

func TestCentAmount_SplitProportionally_noSignFlip(t *testing.T) {
	// A zero weight used to be charged a cent because the last element
	// absorbed the whole rounding difference.
	assert.Equal(t,
		[]CentAmount{1, 0, 0},
		CentAmount(1).SplitProportionally([]CentAmount{1, 1, 0}),
	)

	// Brute force the sign contract: no result amount may have
	// a sign opposite to the split amount.
	weightSets := [][]CentAmount{{1, 1, 0}, {-3, -3, 0}, {1, 2, 3}, {0, 0, 1}, {7, 1, 1, 1}}
	for amount := CentAmount(-20); amount <= 20; amount++ {
		for _, weights := range weightSets {
			var sum CentAmount
			for _, part := range amount.SplitProportionally(weights) {
				if amount > 0 {
					assert.False(t, part.Signbit(), "amount %s weights %v", amount, weights)
				}
				if amount < 0 {
					assert.LessOrEqual(t, int64(part), int64(0), "amount %s weights %v", amount, weights)
				}
				sum += part
			}
			assert.Equal(t, amount, sum, "sum for amount %s weights %v", amount, weights)
		}
	}
}

func TestCentAmount_WithinOneCent_int64Extremes(t *testing.T) {
	// The int64 difference of the two extremes wraps to -1, which must not
	// be mistaken for a one cent tolerance: these amounts are as far apart
	// as CentAmount can express.
	assert.False(t, MaxCentAmount.WithinOneCent(MinCentAmount))
	assert.False(t, MinCentAmount.WithinOneCent(MaxCentAmount))
	assert.False(t, CentAmount(math.MaxInt64).WithinOneCent(math.MinInt64))

	// Real neighbours at the edges must still be within one cent
	assert.True(t, MaxCentAmount.WithinOneCent(MaxCentAmount-1))
	assert.True(t, MinCentAmount.WithinOneCent(MinCentAmount+1))
	assert.False(t, MinCentAmount.WithinOneCent(MinCentAmount+2))
}

func TestCentAmount_signs_rangeExtremes(t *testing.T) {
	// Every sign method negates, so they all share the same overflow edge.
	// Within the valid range negation is total, which is why MinCentAmount
	// stops one above math.MinInt64.
	assert.Equal(t, MaxCentAmount, MinCentAmount.Abs())
	assert.Equal(t, MaxCentAmount, MinCentAmount.Copysign(+1))
	assert.Equal(t, MaxCentAmount, MinCentAmount.WithPosSign(true))
	assert.Equal(t, MaxCentAmount, MinCentAmount.Inverted())
	assert.Equal(t, MinCentAmount, MaxCentAmount.WithNegSign(true))

	// math.MinInt64 itself is out of range and documented to stay negative
	assert.Equal(t, CentAmount(math.MinInt64), CentAmount(math.MinInt64).Abs())
}

func TestCentAmount_ScanString_rounding(t *testing.T) {
	// ScanString has no rounding mode parameter, so the hard-coded
	// RoundHalfAwayFromZero is part of its contract and must be pinned.
	var amount CentAmount
	require.NoError(t, amount.ScanString("0.005", false))
	assert.Equal(t, CentAmount(1), amount)

	require.NoError(t, amount.ScanString("-0.005", false))
	assert.Equal(t, CentAmount(-1), amount)

	require.NoError(t, amount.ScanString("1.0049", false))
	assert.Equal(t, CentAmount(100), amount)
}

func TestCentAmount_SQLParameter(t *testing.T) {
	// CentAmount deliberately implements no driver.Valuer: database/sql has to
	// convert it through the default converter's int64 handling, which is what
	// the README promises. Adding a Valuer or changing the underlying type
	// would break SQL usage silently without this.
	value, err := driver.DefaultParameterConverter.ConvertValue(CentAmount(123))
	require.NoError(t, err)
	assert.Equal(t, int64(123), value)

	value, err = driver.DefaultParameterConverter.ConvertValue(MinCentAmount)
	require.NoError(t, err)
	assert.Equal(t, int64(MinCentAmount), value)
}

func TestCentAmount_stringRoundTripsFullRange(t *testing.T) {
	// Routing parsing through DecimalAmount capped it at the DecimalAmount
	// coefficient range, so ParseCentAmount(c.String()) failed for everything
	// above roughly 2.88e15 currency units, including MaxCentAmount itself.
	for _, amount := range []CentAmount{
		MinCentAmount, MinCentAmount + 1, -1_000_000_000_000_000_000, -12345, -1, 0,
		1, 12345, 1_000_000_000_000_000_000, MaxCentAmount - 1, MaxCentAmount,
		CentAmount(maxDecimalAmountCoefficient), CentAmount(maxDecimalAmountCoefficient) + 1,
	} {
		parsed, err := ParseCentAmount(amount.String(), RoundHalfAwayFromZero)
		require.NoError(t, err, "parsing %s", amount)
		assert.Equal(t, amount, parsed, "round-trip of %s", amount)
	}
}

func TestAmount_CentAmount_largeFiniteAmounts(t *testing.T) {
	// Amount(3e15) is an ordinary finite float whose cent count fits
	// CentAmount; it used to saturate to MaxCentAmount because the
	// intermediate DecimalAmount overflowed at scale 2.
	assert.Equal(t, CentAmount(300_000_000_000_000_000), Amount(3e15).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, CentAmount(-300_000_000_000_000_000), Amount(-3e15).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, CentAmount(1_000_000_000_000_000_000), Amount(1e16).CentAmount(RoundHalfAwayFromZero))

	// Only values genuinely beyond the range may clamp
	assert.Equal(t, MaxCentAmount, Amount(1e30).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, MinCentAmount, Amount(-1e30).CentAmount(RoundHalfAwayFromZero))
}

func TestCentAmount_SplitProportionally_hugeWeightFairness(t *testing.T) {
	// Right-shifting each weight to make their sum fit uint64 changed their
	// ratios and pushed a part more than a cent off its exact share. The
	// big.Int fallback keeps every part within one cent.
	weights := []CentAmount{7937857609592500785, 8248106595770569066, 3007080212677340397}
	parts := MaxCentAmount.SplitProportionally(weights)

	total := new(big.Int)
	for _, weight := range weights {
		total.Add(total, big.NewInt(int64(weight)))
	}
	var sum CentAmount
	for i, part := range parts {
		exact := new(big.Rat).SetFrac(
			new(big.Int).Mul(big.NewInt(int64(MaxCentAmount)), big.NewInt(int64(weights[i]))),
			total,
		)
		diff := new(big.Rat).Sub(new(big.Rat).SetInt64(int64(part)), exact)
		assert.LessOrEqual(t, diff.Abs(diff).Cmp(big.NewRat(1, 1)), 0,
			"part %d is more than one cent from its exact share", i)
		sum += part
	}
	assert.Equal(t, MaxCentAmount, sum)
}

func TestDecimalAmount_CentAmount_fullRange(t *testing.T) {
	// Rounding through RoundToCents forced the scale 2 result into the
	// narrower DecimalAmount coefficient range, so conversions failed well
	// before the CentAmount range was exhausted.
	cents, err := NewDecimalAmount(30_000_000_000_000_000, 0).CentAmount(RoundHalfAwayFromZero)
	require.NoError(t, err)
	assert.Equal(t, CentAmount(3_000_000_000_000_000_000), cents)

	cents, err = NewDecimalAmount(-30_000_000_000_000_000, 0).CentAmount(RoundHalfAwayFromZero)
	require.NoError(t, err)
	assert.Equal(t, CentAmount(-3_000_000_000_000_000_000), cents)

	// The largest amount that still fits: MaxCentAmount cents at scale 2
	cents, err = NewDecimalAmount(maxDecimalAmountCoefficient, 2).CentAmount(RoundHalfAwayFromZero)
	require.NoError(t, err)
	assert.Equal(t, CentAmount(maxDecimalAmountCoefficient), cents)

	// Only genuinely unrepresentable values may fail
	_, err = NewDecimalAmount(maxDecimalAmountCoefficient, 0).CentAmount(RoundHalfAwayFromZero)
	assert.Error(t, err, "2.88e17 currency units is 2.88e19 cents, beyond MaxCentAmount")
}

// TestParseCentAmount_acceptedDecimalsRejectNonFinite pins that ParseCentAmount
// rejects NaN and ±Inf with acceptedDecimals like ParseAmount, ParseRate and
// ParseDecimalAmount do, and also without them, because a CentAmount has no
// non-finite values to return.
func TestParseCentAmount_acceptedDecimalsRejectNonFinite(t *testing.T) {
	for _, str := range []string{"NaN", "Inf", "+Inf", "-Inf", "Infinity", "-Infinity"} {
		c, err := ParseCentAmount(str, RoundHalfAwayFromZero, 0, 2)
		assert.Error(t, err, "ParseCentAmount(%q, 0, 2) = %v", str, c)
		c, err = ParseCentAmount(str, RoundHalfAwayFromZero)
		assert.Error(t, err, "ParseCentAmount(%q) = %v", str, c)
	}
}

// TestParseCentAmount_scientificNotation covers the exponent forms that JSON
// encoders and PostgreSQL float8 text output produce. CentAmount and
// DecimalAmount are read from the same columns and the same payloads, so a
// literal one of them accepts must not be an error in the other.
//
// The exponent moves the split between the integer and the fractional digits
// instead of going through a float64, so the result stays exact: digits the
// exponent pushes past the second decimal place still only decide the rounding.
func TestParseCentAmount_scientificNotation(t *testing.T) {
	valid := map[string]CentAmount{
		"1e0":      100,
		"1e2":      10000,
		"1.5e2":    15000,
		"1.5E+2":   15000,
		"-2.5e1":   -2500,
		"1e-2":     1,
		"12.345e1": 12345,
		"12.345e2": 123450,
		// The exponent composes with the locale-aware separator detection.
		"1.234,56e1": 1234560,
		// A zero has nothing for the exponent to shift.
		"0e-1000": 0,
		"0e1000":  0,
		// The full CentAmount range is reachable through an exponent, which is
		// the range DecimalAmount cannot represent.
		"92233720368547758.07e0": MaxCentAmount,
		"9223372036854775807e-2": MaxCentAmount,
	}
	for str, expected := range valid {
		t.Run(str, func(t *testing.T) {
			parsed, err := ParseCentAmount(str, RoundHalfAwayFromZero)
			require.NoError(t, err)
			assert.Equal(t, expected, parsed)
		})
	}

	// Rounding still applies to whatever lands past the second decimal place
	// after the shift, and the mode still reaches it.
	t.Run("rounding after the shift", func(t *testing.T) {
		// 1.005e0 == 1.005 rounds to 101 cents, while the same digits shifted
		// one place left land exactly on 10.05 and need no rounding at all.
		parsed, err := ParseCentAmount("1.005e0", RoundHalfAwayFromZero)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(101), parsed)

		parsed, err = ParseCentAmount("1.005e0", RoundDown)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(100), parsed)

		parsed, err = ParseCentAmount("1.005e1", RoundDown)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(1005), parsed)
	})

	// acceptedDecimals matches the decimal places of the value after the
	// exponent is applied, so a caller validating cent precision sees what the
	// value actually is, not how it happened to be spelled.
	t.Run("acceptedDecimals after the shift", func(t *testing.T) {
		parsed, err := ParseCentAmount("1.5e2", RoundHalfAwayFromZero, 0)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(15000), parsed)

		_, err = ParseCentAmount("1.5e2", RoundHalfAwayFromZero, 1)
		assert.Error(t, err)

		parsed, err = ParseCentAmount("12.345e1", RoundHalfAwayFromZero, 2)
		require.NoError(t, err)
		assert.Equal(t, CentAmount(12345), parsed)
	})

	// Unrepresentable or ambiguous exponent forms fail loud rather than
	// producing an amount assembled from the wrong digits.
	for _, str := range []string{
		"1e19",   // 1e19 units exceeds MaxCentAmount cents
		"1e1000", // far out of range
		"-1e19",  // out of range on the negative side
		"1e1e1",  // repeated exponent
		"1 e3",   // whitespace before the exponent marker
		"1e 3",   // whitespace after the exponent marker
		"1e",     // marker without digits
		"1e1001", // exponent beyond the accepted magnitude
		"NaNe1",  // non-finite with an exponent
	} {
		t.Run(str, func(t *testing.T) {
			_, err := ParseCentAmount(str, RoundHalfAwayFromZero)
			assert.Error(t, err)
		})
	}

	// The error names the type being parsed, not the one that happens to own
	// the shared exponent helper.
	_, err := ParseCentAmount("1e 3", RoundHalfAwayFromZero)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "money.CentAmount")
}

// TestParseCentAmount_exponentMatchesDecimalAmount pins that the two exact
// money parsers agree on every exponent literal both can represent. They read
// the same JSON payloads and the same SQL columns, so a value that scans into
// one must scan into the other with the same amount.
func TestParseCentAmount_exponentMatchesDecimalAmount(t *testing.T) {
	for _, str := range []string{
		"1e0", "1e2", "1.5e2", "1.5E+2", "-2.5e1", "1e-2", "12.345e2",
		"1e-5", "0e-30", "0e30", "1.234,56e1", "2.5e-1", "-1e3",
	} {
		t.Run(str, func(t *testing.T) {
			cents, err := ParseCentAmount(str, RoundHalfAwayFromZero)
			require.NoError(t, err, "ParseCentAmount(%q)", str)

			decimal, err := ParseDecimalAmount(str)
			require.NoError(t, err, "ParseDecimalAmount(%q)", str)

			fromDecimal, err := decimal.CentAmount(RoundHalfAwayFromZero)
			require.NoError(t, err)
			assert.Equal(t, fromDecimal, cents, "%q parsed differently by the two parsers", str)
		})
	}
}

// benchCents, benchRates and benchPercents are realistic inputs for the
// benchmarks of the rate methods, varied so that the branch predictor
// can't learn a single input.
var (
	benchCents    = makeBenchCents()
	benchRates    = []Rate{1.0845, 0.92137, 1.5, 0.145, 7.4602, 0.0067, 25.317, 1}
	benchPercents = []float64{19, 20, 7.7, 2.5, 10, 13, 0.5, 33.333}
)

func makeBenchCents() []CentAmount {
	rng := rand.New(rand.NewPCG(1, 2))
	cents := make([]CentAmount, 1024)
	for i := range cents {
		cents[i] = CentAmount(rng.Int64N(2e8) - 1e8)
	}
	return cents
}

func BenchmarkCentAmount_MultipliedByRate(b *testing.B) {
	i := 0
	for b.Loop() {
		benchCents[i%len(benchCents)].MultipliedByRate(benchRates[i%len(benchRates)])
		i++
	}
}

func BenchmarkCentAmount_DividedByRate(b *testing.B) {
	i := 0
	for b.Loop() {
		benchCents[i%len(benchCents)].DividedByRate(benchRates[i%len(benchRates)])
		i++
	}
}

func BenchmarkCentAmount_Percentage(b *testing.B) {
	i := 0
	for b.Loop() {
		benchCents[i%len(benchCents)].Percentage(benchPercents[i%len(benchPercents)])
		i++
	}
}
