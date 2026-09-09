package money

import (
	"database/sql/driver"
	"encoding/json"
	"math"
	"math/big"
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

		// A cent count the DecimalAmount coefficient cannot hold but
		// CentAmount can: 3e17 cents is above maxDecimalAmountCoefficient
		// and well below MaxCentAmount.
		cents, err := ParseCentAmount("3000000000000000.00", RoundHalfAwayFromZero)
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
