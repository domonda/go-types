package float

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidAndHasSign(t *testing.T) {
	tests := []struct {
		f    float64
		sign int
		want bool
	}{
		{f: 0, sign: 0, want: true},
		{f: 0, sign: +1, want: true},
		{f: 0, sign: -1, want: false},

		{f: -1, sign: 0, want: true},
		{f: -1, sign: +1, want: false},
		{f: -1, sign: -1, want: true},

		{f: +1, sign: 0, want: true},
		{f: +1, sign: +1, want: true},
		{f: +1, sign: -1, want: false},

		{f: math.NaN(), sign: 0, want: false},
		{f: math.NaN(), sign: +1, want: false},
		{f: math.NaN(), sign: -1, want: false},
		{f: math.Inf(0), sign: 0, want: false},
		{f: math.Inf(+1), sign: +1, want: false},
		{f: math.Inf(-1), sign: -1, want: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%#v_%#v", tt.f, tt.sign), func(t *testing.T) {
			if got := ValidAndHasSign(tt.f, tt.sign); got != tt.want {
				t.Errorf("ValidAndHasSign(%#v, %#v) = %#v, want %#v", tt.f, tt.sign, got, tt.want)
			}
		})
	}
}

// TestRoundToDecimals protects the rounding mode: values exactly halfway
// round away from zero (not banker's rounding), so amounts like 2.5
// become 3 and -2.5 becomes -3, symmetric around zero.
// Negative decimals round to tens, hundreds, etc.
func TestRoundToDecimals(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		decimals int
		want     float64
	}{
		{name: "zero", f: 0, decimals: 2, want: 0},
		{name: "round down", f: 1.2345678, decimals: 2, want: 1.23},
		{name: "round up", f: 1.2345678, decimals: 3, want: 1.235},
		{name: "negative round toward zero", f: -1.2345678, decimals: 2, want: -1.23},
		{name: "already rounded", f: 1.5, decimals: 2, want: 1.5},
		{name: "half away from zero positive", f: 2.5, decimals: 0, want: 3},
		{name: "half away from zero negative", f: -2.5, decimals: 0, want: -3},
		{name: "exactly representable half", f: 0.125, decimals: 2, want: 0.13},
		{name: "negative decimals round to hundreds", f: 1234.5, decimals: -2, want: 1200},
		{name: "negative decimals half up", f: 1250, decimals: -2, want: 1300},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, RoundToDecimals(tt.f, tt.decimals))
		})
	}

	t.Run("float32 keeps its type", func(t *testing.T) {
		got := RoundToDecimals(float32(1.2345678), 2)
		assert.Equal(t, float32(1.23), got)
	})

	t.Run("NaN and Inf pass through", func(t *testing.T) {
		assert.True(t, math.IsNaN(RoundToDecimals(math.NaN(), 2)))
		assert.True(t, math.IsInf(RoundToDecimals(math.Inf(1), 2), 1))
		assert.True(t, math.IsInf(RoundToDecimals(math.Inf(-1), 2), -1))
	})
}

// TestRoundToDecimals_shortestDecimalTies protects the rounding contract
// shared with Format and all money rounding: the value that is rounded is
// the shortest decimal representation of the float, the decimal a human
// typed or a JSON number carried, not its binary approximation.
// 1.005 is stored as 1.00499999999999989342, so rounding the binary value
// yields 1 while every money method and Format yield 1.01 for it.
// Rounding the same number differently depending on the function used
// makes sums and their formatted strings disagree by a cent.
func TestRoundToDecimals_shortestDecimalTies(t *testing.T) {
	tests := []struct {
		f        float64
		decimals int
		want     float64
	}{
		{f: 1.005, decimals: 2, want: 1.01},
		{f: -1.005, decimals: 2, want: -1.01},
		{f: 0.145, decimals: 2, want: 0.15},
		{f: 2.675, decimals: 2, want: 2.68},
		{f: -2.675, decimals: 2, want: -2.68},
		{f: 8.345, decimals: 2, want: 8.35},
		{f: 1234.565, decimals: 2, want: 1234.57},
		{f: 1.0005, decimals: 3, want: 1.001},
		{f: 1.45, decimals: 1, want: 1.5},
		{f: 0.5, decimals: 0, want: 1},
		{f: 9.995, decimals: 2, want: 10},
		{f: 0.144999, decimals: 2, want: 0.14}, // not a tie
		{f: 0.1 + 0.2, decimals: 2, want: 0.3},
		{f: 15, decimals: -1, want: 20},
		{f: -25, decimals: -1, want: -30},
		{f: 4.9, decimals: -1, want: 0},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, RoundToDecimals(tt.f, tt.decimals), "RoundToDecimals(%v, %d)", tt.f, tt.decimals)
	}

	// float32(1.005) is 1.00499999523162841797 in binary,
	// its shortest float32 decimal is still 1.005
	assert.Equal(t, float32(1.01), RoundToDecimals(float32(1.005), 2))
	assert.Equal(t, float32(2.68), RoundToDecimals(float32(2.675), 2))

	// Negative zero keeps its sign like math.Round does
	assert.True(t, math.Signbit(RoundToDecimals(-0.001, 2)), "-0.001 rounds to -0")
}

// TestRoundToDecimals_overflow protects that a huge number of decimals or a
// huge value never turns a finite float into NaN or Inf. Scaling by 10^decimals
// overflowed: 10^400 is Inf and f*Inf/Inf is NaN, and 1e300*10^10 is Inf.
// A value that has no more decimals than requested is returned unchanged,
// and rounding to a power of ten beyond the value yields zero.
func TestRoundToDecimals_overflow(t *testing.T) {
	assert.Equal(t, 1.5, RoundToDecimals(1.5, 400))
	assert.Equal(t, 1.5, RoundToDecimals(1.5, 309))
	assert.Equal(t, 1e300, RoundToDecimals(1e300, 10))
	assert.Equal(t, -1e300, RoundToDecimals(-1e300, 10))
	assert.Equal(t, math.MaxFloat64, RoundToDecimals(math.MaxFloat64, 2))
	assert.Equal(t, math.SmallestNonzeroFloat64, RoundToDecimals(math.SmallestNonzeroFloat64, 400))
	assert.Equal(t, 1.2345e-300, RoundToDecimals(1.2345e-300, 310))
	assert.Equal(t, 0.0, RoundToDecimals(1.5, -400))
	assert.Equal(t, 0.0, RoundToDecimals(1e300, -400))
	// Rounding to a power of ten more than ten times the value is zero,
	// while one ten times the value or less rounds to it
	assert.Zero(t, RoundToDecimals(4.0, -2))
	assert.Zero(t, RoundToDecimals(-4.0, -2))
	assert.Equal(t, 100.0, RoundToDecimals(51.0, -2))
	assert.Equal(t, float32(math.MaxFloat32), RoundToDecimals(float32(math.MaxFloat32), 2))
	assert.Equal(t, float32(1.5), RoundToDecimals(float32(1.5), 400))

	// Only a rounded value that is itself beyond the float range is ±Inf,
	// like any other float64 operation overflowing:
	// 1.7976931348623157e308 rounded to a multiple of 10^307 is 1.8e308
	assert.True(t, math.IsInf(RoundToDecimals(math.MaxFloat64, -307), 1))
	assert.True(t, math.IsInf(RoundToDecimals(-math.MaxFloat64, -307), -1))
}

// TestRoundToDecimals_fastPathMatchesExact protects that the float64 fast
// path of RoundToDecimals never changes a result. It is only allowed to skip
// formatting and parsing the digits where it provably returns the identical
// float64, so every input must round exactly like rounding the digits of the
// shortest decimal representation does, most of all the decimal ties and
// their float64 neighbors where float64 arithmetic alone gets it wrong.
func TestRoundToDecimals_fastPathMatchesExact(t *testing.T) {
	check := func(f float64, decimals int) {
		t.Helper()
		got := RoundToDecimals(f, decimals)
		want := roundToDecimalsExact(f, 64, decimals)
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("RoundToDecimals(%v, %d) = %v, exact digit rounding = %v", f, decimals, got, want)
		}
	}
	for _, f := range fastPathTargetedFloats() {
		for decimals := -24; decimals <= 24; decimals++ {
			check(f, decimals)
		}
	}
	for _, tie := range fastPathTies() {
		for decimals := tie.decimals - 1; decimals <= tie.decimals+1; decimals++ {
			check(tie.f, decimals)
		}
	}
	rng := rand.New(rand.NewPCG(3, 4))
	for range fastPathSweepSize() {
		check(randomFastPathFloat(rng), rng.IntN(20)-5)
	}
}

// TestRoundToDecimals_fastPathTaken protects the reason for the fast path:
// rounding computed amounts to cents must not format and parse their digits,
// which is an order of magnitude slower. Only values within a tiny relative
// distance of a tie may take the exact path.
func TestRoundToDecimals_fastPathTaken(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	exact := 0
	for range 100_000 {
		amount := float64(rng.Int64N(2e10)-1e10) / 100 * (rng.Float64() * 2)
		if _, ok := roundScaledFast(amount, 2); !ok {
			exact++
		}
	}
	assert.Less(t, exact, 10, "amounts that took the exact path")
}

// fastPathTie is a float64 whose shortest decimal representation ends
// in a 5 right after the passed number of decimals.
type fastPathTie struct {
	f        float64
	decimals int
}

// fastPathTies returns ties of the shortest decimal representation for every
// number of decimals from -6 to 16 with both signs, plus their float64
// neighbors two ulps up and down, for the tests of the fast paths.
// The digits are at most 15, so they are the shortest decimal of the parsed
// float64.
func fastPathTies() []fastPathTie {
	rng := rand.New(rand.NewPCG(1, 2))
	var ties []fastPathTie
	for decimals := -6; decimals <= 16; decimals++ {
		for range 300 {
			// The digits before the 5, from none to 14 of them
			digits := rng.Int64N(int64(math.Pow10(rng.IntN(15))))
			tie, err := strconv.ParseFloat(fmt.Sprintf("%d5e%d", digits, -decimals-1), 64)
			if err != nil {
				panic(err)
			}
			for _, f := range withNeighbors(tie) {
				ties = append(ties, fastPathTie{f, decimals}, fastPathTie{-f, decimals})
			}
		}
	}
	return ties
}

// fastPathTargetedFloats returns ties, boundaries of the fast paths, and
// special values with both signs and their float64 neighbors.
func fastPathTargetedFloats() []float64 {
	var floats []float64
	for _, f := range []float64{
		0, 0.5, 1.5, 2.5, 0.125, 0.145, 1.005, 2.675, 8.345, 1234.565,
		1.45, 9.995, 1.0005, 0.005, 15, 25, 1250, 0.1 + 0.2, 0.3, 1.0845,
		5e-324, 2.2250738585072014e-308, 1e-300, 1e-20, 0.4999999999999999,
		0x1p47, 0x1p47 + 0.5, 0x1p48, 0x1p48 + 0.5, 0x1p52, 0x1p52 + 0.5,
		0x1p53, 0x1p53 + 2, 1e15, 1e15 + 0.5, 1e16, 1e22, 1e23, 1e300,
		math.MaxFloat64,
	} {
		for _, g := range withNeighbors(f) {
			floats = append(floats, g, -g)
		}
	}
	return floats
}

// withNeighbors returns f and its finite float64 neighbors
// up to two ulps away.
func withNeighbors(f float64) []float64 {
	up1 := math.Nextafter(f, math.Inf(1))
	down1 := math.Nextafter(f, math.Inf(-1))
	floats := []float64{f}
	for _, g := range []float64{up1, math.Nextafter(up1, math.Inf(1)), down1, math.Nextafter(down1, math.Inf(-1))} {
		if !math.IsInf(g, 0) {
			floats = append(floats, g)
		}
	}
	return floats
}

// randomFastPathFloat returns a random finite float64 for the fast path
// sweeps: a short decimal like a stored amount, the product of one with a
// rate, which has a long binary fraction, or any bit pattern.
func randomFastPathFloat(rng *rand.Rand) float64 {
	decimal := float64(rng.Int64N(int64(math.Pow10(1+rng.IntN(15))))) / math.Pow10(rng.IntN(13))
	if rng.IntN(2) == 0 {
		decimal = -decimal
	}
	switch rng.IntN(4) {
	case 0:
		return decimal
	case 1, 2:
		return decimal * (rng.Float64() * 2)
	default:
		for {
			f := math.Float64frombits(rng.Uint64())
			if !math.IsNaN(f) && !math.IsInf(f, 0) {
				return f
			}
		}
	}
}

// fastPathSweepSize returns the number of random inputs
// for the fast path sweeps, fewer with -short.
func fastPathSweepSize() int {
	if testing.Short() {
		return 100_000
	}
	return 1_000_000
}

// TestDerefOr protects the nil-safety contract: a nil pointer yields the
// default, while a non-nil pointer to zero must yield zero (not the default),
// because zero is a legitimate value distinct from "not set".
func TestDerefOr(t *testing.T) {
	assert.Equal(t, 7.5, DerefOr(nil, 7.5), "nil pointer returns default")

	zero := 0.0
	assert.Equal(t, 0.0, DerefOr(&zero, 7.5), "pointer to zero must not be replaced by default")

	val := -1.25
	assert.Equal(t, -1.25, DerefOr(&val, 7.5))

	f32 := float32(3.5)
	assert.Equal(t, float32(3.5), DerefOr(&f32, 1))
	assert.Equal(t, float32(1), DerefOr[float32](nil, 1))
}

func TestValid(t *testing.T) {
	assert.True(t, Valid(0.0))
	assert.True(t, Valid(-1.5))
	assert.True(t, Valid(math.MaxFloat64))
	assert.True(t, Valid(float32(1)))
	assert.False(t, Valid(math.NaN()))
	assert.False(t, Valid(math.Inf(1)))
	assert.False(t, Valid(math.Inf(-1)))
	assert.False(t, Valid(float32(math.Inf(1))))
}

// benchFloats are realistic money values for the rounding benchmark:
// every other one has two decimals as stored, the rest are rate products
// with a long binary fraction.
var benchFloats = func() []float64 {
	rng := rand.New(rand.NewPCG(1, 2))
	floats := make([]float64, 1024)
	for i := range floats {
		floats[i] = float64(rng.Int64N(2e8)-1e8) / 100
		if i%2 == 1 {
			floats[i] *= 1.0845
		}
	}
	return floats
}()

func BenchmarkRoundToDecimals(b *testing.B) {
	i := 0
	for b.Loop() {
		RoundToDecimals(benchFloats[i%len(benchFloats)], 2)
		i++
	}
}
