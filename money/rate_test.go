package money

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRate(t *testing.T) {
	type args struct {
		str              string
		acceptedDecimals []int
	}
	tests := []struct {
		name    string
		args    args
		want    Rate
		wantErr bool
	}{
		{name: "15", args: args{str: "15"}, want: 15},
		{name: "15%", args: args{str: "15%"}, want: 0.15},
		{name: "15 %", args: args{str: "15 %"}, want: 0.15},
		{name: " 15 % ", args: args{str: " 15 % "}, want: 0.15},
		{name: "8,382.00", args: args{str: "8,382.00"}, want: 8382},
		{name: "NaN", args: args{str: `NaN`}, want: Rate(math.NaN())},
		{name: "Inf", args: args{str: `Inf`}, want: Rate(math.Inf(1))},
		{name: "+Inf", args: args{str: `+Inf`}, want: Rate(math.Inf(1))},
		{name: "-Inf", args: args{str: `-Inf`}, want: Rate(math.Inf(-1))},

		// German invoices write percentages with a decimal comma
		{name: "19,5%", args: args{str: "19,5%"}, want: 0.195},
		{name: "-2,5%", args: args{str: "-2,5%"}, want: -0.025},
		// The decimal count is checked on the written number,
		// not on the fraction after the division by 100
		{name: "1.5 % with 1 decimal", args: args{str: "1.5 %", acceptedDecimals: []int{1}}, want: 0.015},
		{name: "1.5 % with 2 decimals", args: args{str: "1.5 %", acceptedDecimals: []int{2}}, wantErr: true},
		{name: "20% with 0 decimals", args: args{str: "20%", acceptedDecimals: []int{0}}, want: 0.2},
		// Non-finite values have no decimals to check
		{name: "NaN with 2 decimals", args: args{str: "NaN", acceptedDecimals: []int{2}}, want: Rate(math.NaN())},
		{name: "empty", args: args{str: ""}, wantErr: true},
		{name: "only percent sign", args: args{str: "%"}, wantErr: true},
		{name: "not a number", args: args{str: "abc"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRate(tt.args.str, tt.args.acceptedDecimals...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !equalInclNaN(float64(got), float64(tt.want)) {
				t.Errorf("ParseRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseRate_errorQuotesOriginalInput protects the error message users
// see: the percent sign and surrounding whitespace are stripped before the
// number is parsed, so without wrapping the error would quote a different
// string than the one that was actually passed in.
func TestParseRate_errorQuotesOriginalInput(t *testing.T) {
	_, err := ParseRate("abc %")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"abc %"`)
}

// TestRate_UnmarshalJSON pins that JSON APIs may send a rate as a number,
// as a numeric string, or as a percentage string like "19%", and that all of
// them decode to the same fraction, while null and "" reset the rate to zero.
func TestRate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Rate
		wantErr bool
	}{
		{name: "null resets to zero", json: `null`, want: 0},
		{name: "empty string resets to zero", json: `""`, want: 0},
		{name: "number", json: `0.19`, want: 0.19},
		{name: "numeric string", json: `"0.19"`, want: 0.19},
		{name: "percent string", json: `"19%"`, want: 0.19},
		{name: "percent string with space", json: `"19 %"`, want: 0.19},
		{name: "percent string with decimal comma", json: `"2,5%"`, want: 0.025},
		{name: "quoted NaN", json: `"NaN"`, want: Rate(math.NaN())},
		{name: "empty input", json: ``, wantErr: true},
		{name: "boolean", json: `true`, wantErr: true},
		{name: "non numeric string", json: `"abc"`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rate(666) // Different from the zero value to detect resets
			err := r.UnmarshalJSON([]byte(tt.json))
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, Rate(666), r, "a failed unmarshal must not modify the rate")
				return
			}
			require.NoError(t, err)
			assert.True(t, equalInclNaN(float64(tt.want), float64(r)), "got %v, want %v", r, tt.want)
		})
	}
}

// TestRate_JSONRoundTrip pins that Rate, which has no MarshalJSON method,
// is encoded as a plain JSON number that decodes back to the identical rate,
// so exchange rates don't drift when passed through an API, and that a
// percentage string in a struct field goes through Rate.UnmarshalJSON.
func TestRate_JSONRoundTrip(t *testing.T) {
	for _, r := range []Rate{0, 0.19, 1.0857, -0.025, 1e-9, 123456.789} {
		data, err := json.Marshal(r)
		require.NoError(t, err)
		var got Rate
		require.NoError(t, json.Unmarshal(data, &got), "json.Unmarshal(%s)", data)
		assert.Equal(t, r, got, "round-trip of %s", data)
	}

	var s struct{ VAT Rate }
	require.NoError(t, json.Unmarshal([]byte(`{"VAT":"20%"}`), &s))
	assert.Equal(t, Rate(0.2), s.VAT)
}

// TestNullableRate pins that NullableRate distinguishes a zero rate
// (for example 0% VAT) from a missing rate in Go, JSON and SQL.
func TestNullableRate(t *testing.T) {
	t.Run("constructors", func(t *testing.T) {
		zero := NullableRateFrom(0)
		assert.True(t, zero.IsNotNull(), "a zero rate is a valid rate, not NULL")
		assert.Equal(t, Rate(0), zero.Get())

		assert.True(t, NullableRateFromPtr(nil).IsNull())

		r := Rate(0.19)
		fromPtr := NullableRateFromPtr(&r)
		r = 0.5
		assert.Equal(t, Rate(0.19), fromPtr.Get(), "NullableRateFromPtr must copy the pointed to value")
	})

	t.Run("JSON", func(t *testing.T) {
		for _, tt := range []struct {
			rate NullableRate
			json string
		}{
			{rate: NullableRate{}, json: `null`},
			{rate: NullableRateFrom(0), json: `0`},
			{rate: NullableRateFrom(0.19), json: `0.19`},
		} {
			data, err := json.Marshal(tt.rate)
			require.NoError(t, err)
			assert.Equal(t, tt.json, string(data))

			var got NullableRate
			require.NoError(t, json.Unmarshal(data, &got))
			assert.Equal(t, tt.rate, got, "round-trip of %s", data)
		}

		// JSON null must clear a previously set rate
		got := NullableRateFrom(0.19)
		require.NoError(t, json.Unmarshal([]byte(`null`), &got))
		assert.True(t, got.IsNull())

		// The percent notation of Rate.UnmarshalJSON must also work
		// through the nullable wrapper
		require.NoError(t, json.Unmarshal([]byte(`"19%"`), &got))
		assert.Equal(t, NullableRateFrom(0.19), got)
	})

	t.Run("SQL", func(t *testing.T) {
		v, err := NullableRate{}.Value()
		require.NoError(t, err)
		assert.Nil(t, v, "NULL must be written to SQL as nil, not as 0")

		// Drivers only accept the basic driver.Value types,
		// so the named Rate type must arrive as a plain float64
		v, err = NullableRateFrom(0.19).Value()
		require.NoError(t, err)
		assert.Equal(t, float64(0.19), v)

		got := NullableRateFrom(1)
		require.NoError(t, got.Scan(nil))
		assert.True(t, got.IsNull(), "SQL NULL must scan as NULL")

		require.NoError(t, got.Scan(float64(0.19)))
		assert.Equal(t, NullableRateFrom(0.19), got)

		// Postgres NUMERIC columns arrive as text bytes
		require.NoError(t, got.Scan([]byte("0.0825")))
		assert.Equal(t, NullableRateFrom(0.0825), got)
	})
}

// TestRate_ScanString pins that ScanString accepts both decimal separators,
// thousands grouping and a trailing percent sign like ParseRate does, and
// that validate rejects the non-finite rates that the parser accepts.
// A percentage like "15%" has to scan as the same fraction 0.15 that
// ParseRate and UnmarshalJSON return for it, because a rate read from a
// CSV or form must not depend on which of the three entry points read it.
func TestRate_ScanString(t *testing.T) {
	tests := []struct {
		source   string
		validate bool
		want     Rate
		wantErr  bool
	}{
		{source: "0.19", validate: true, want: 0.19},
		{source: "0,19", validate: true, want: 0.19},
		{source: " 0.19 ", validate: true, want: 0.19},
		{source: "1.234,5", validate: true, want: 1234.5},
		{source: "1,234.5", validate: true, want: 1234.5},
		{source: "-1,5", validate: true, want: -1.5},
		{source: "15%", validate: true, want: 0.15},
		{source: " 19,5 % ", validate: true, want: 0.195},
		{source: "-2.5%", validate: false, want: -0.025},
		{source: "NaN", validate: false, want: Rate(math.NaN())},
		{source: "Inf", validate: false, want: Rate(math.Inf(1))},
		{source: "NaN", validate: true, wantErr: true},
		{source: "-Inf", validate: true, wantErr: true},
		{source: "", validate: false, wantErr: true},
		{source: "abc", validate: false, wantErr: true},
		{source: "%", validate: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q validate=%t", tt.source, tt.validate), func(t *testing.T) {
			r := Rate(666)
			err := r.ScanString(tt.source, tt.validate)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, Rate(666), r, "a failed scan must not modify the rate")
				return
			}
			require.NoError(t, err)
			assert.True(t, equalInclNaN(float64(tt.want), float64(r)), "got %v, want %v", r, tt.want)

			parsed, err := ParseRate(tt.source)
			require.NoError(t, err)
			assert.True(t, equalInclNaN(float64(parsed), float64(r)), "ScanString %v must equal ParseRate %v", r, parsed)
		})
	}
}

// TestRate_Inverse pins the exchange rate inversion: the inverse of the
// EUR->USD rate is the USD->EUR rate, so converting an amount with a rate
// and back with its inverse must return the original amount.
func TestRate_Inverse(t *testing.T) {
	assert.Equal(t, Rate(0.8), Rate(1.25).Inverse())
	assert.Equal(t, Rate(0.25), Rate(4).Inverse())
	assert.Equal(t, Rate(-2), Rate(-0.5).Inverse())

	for _, rate := range []Rate{1.25, 1.0857, 0.9213, 157.32} {
		for _, amount := range []Amount{100, 1234.56, -99.99} {
			converted := amount.MultipliedByRate(rate)
			assert.Equal(t, amount, converted.MultipliedByRate(rate.Inverse()).RoundToCents(),
				"%v converted with %v and back", amount, rate)
			assert.Equal(t, amount, converted.DividedByRate(rate).RoundToCents(),
				"%v converted with %v and back", amount, rate)
		}
	}

	// A zero rate has no inverse, the result is infinite
	// and must be detectable via Valid before it is used
	assert.True(t, math.IsInf(float64(Rate(0).Inverse()), 1))
	assert.True(t, math.IsInf(float64(Rate(math.Copysign(0, -1)).Inverse()), -1))
	assert.False(t, Rate(0).Inverse().Valid())
}

// TestRate_Sign pins the sign helpers that are used to give a rate
// a direction independent of the sign it was parsed with,
// including negative zero which must not leak a sign into results.
func TestRate_Sign(t *testing.T) {
	negZero := Rate(math.Copysign(0, -1))

	assert.False(t, Rate(0).Signbit())
	assert.True(t, negZero.Signbit())
	assert.True(t, Rate(-0.19).Signbit())

	assert.Equal(t, Rate(-1.5), Rate(1.5).Copysign(-1))
	assert.Equal(t, Rate(1.5), Rate(-1.5).Copysign(1))

	assert.Equal(t, Rate(0.19), Rate(-0.19).Abs())
	assert.Equal(t, Rate(math.Inf(1)), Rate(math.Inf(-1)).Abs())
	assert.True(t, math.IsNaN(float64(Rate(math.NaN()).Abs())))
	assert.False(t, negZero.Abs().Signbit())
	assert.Equal(t, 0.19, Rate(-0.19).AbsFloat())

	assert.Equal(t, Rate(-0.19), Rate(0.19).Inverted())
	assert.Equal(t, Rate(0.19), Rate(-0.19).Inverted())
	r := Rate(0.19)
	r.Invert()
	assert.Equal(t, Rate(-0.19), r, "Invert must change the rate in place")

	for _, r := range []Rate{0.19, -0.19} {
		assert.Equal(t, Rate(0.19), r.WithPosSign(true), "%v.WithPosSign(true)", r)
		assert.Equal(t, Rate(-0.19), r.WithPosSign(false), "%v.WithPosSign(false)", r)
		assert.Equal(t, Rate(-0.19), r.WithNegSign(true), "%v.WithNegSign(true)", r)
		assert.Equal(t, Rate(0.19), r.WithNegSign(false), "%v.WithNegSign(false)", r)
	}
	assert.False(t, negZero.WithPosSign(true).Signbit())
}

// TestRate_Rounding pins half away from zero rounding, which is how
// published exchange rates and percentages are commonly rounded,
// and that negative decimals round to tens, hundreds, etc.
func TestRate_Rounding(t *testing.T) {
	assert.Equal(t, Rate(3), Rate(2.5).RoundToInt())
	assert.Equal(t, Rate(-3), Rate(-2.5).RoundToInt())
	assert.Equal(t, Rate(2), Rate(2.4).RoundToInt())

	tests := []struct {
		rate     Rate
		decimals int
		want     Rate
	}{
		{rate: 1.23456, decimals: 4, want: 1.2346},
		{rate: 1.23454, decimals: 4, want: 1.2345},
		{rate: -1.23456, decimals: 4, want: -1.2346},
		{rate: 0.5, decimals: 0, want: 1},
		{rate: -0.5, decimals: 0, want: -1},
		{rate: 1234.5, decimals: -2, want: 1200},
		{rate: 1250, decimals: -2, want: 1300},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.rate.RoundToDecimals(tt.decimals), "%v.RoundToDecimals(%d)", tt.rate, tt.decimals)
	}
}

// TestRate_FormatSep pins the localized formatting of rates with the
// German and English separator conventions and an explicit precision,
// because rates like 0.19 and 1.0857 need different digit counts.
func TestRate_FormatSep(t *testing.T) {
	tests := []struct {
		rate         Rate
		thousandsSep rune
		decimalSep   rune
		precision    int
		want         string
	}{
		{rate: 1234.5, thousandsSep: '.', decimalSep: ',', precision: 4, want: "1.234,5000"},
		{rate: 1234.5, thousandsSep: ',', decimalSep: '.', precision: 2, want: "1,234.50"},
		{rate: 0.0825, thousandsSep: 0, decimalSep: ',', precision: -1, want: "0,0825"},
		{rate: -0.19, thousandsSep: 0, decimalSep: '.', precision: 2, want: "-0.19"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.rate.FormatSep(tt.thousandsSep, tt.decimalSep, tt.precision))
	}
}

// TestRate_GoString pins that the debug representation shows the exact
// binary value of the rate, which the short representation hides:
// 0.15 is not exactly representable as float64.
func TestRate_GoString(t *testing.T) {
	assert.Equal(t, exactDecimalExpansion(0.15), Rate(0.15).GoString())
	assert.Equal(t, exactDecimalExpansion(1.0857), fmt.Sprintf("%#v", Rate(1.0857)))
	assert.Equal(t, "1", Rate(1).GoString())
	assert.Equal(t, "-2.5", Rate(-2.5).GoString())
	assert.Equal(t, "0", Rate(0).GoString())
}

// TestRate_PtrHelpers pins the nil handling of the pointer helpers
// used for optional rates, and that StringOr keeps the full precision
// because rates, unlike amounts, are not rounded to cents.
func TestRate_PtrHelpers(t *testing.T) {
	var nilRate *Rate
	r := Rate(0.0825)

	assert.Equal(t, Rate(1), RateFromPtr(nil, 1))
	assert.Equal(t, r, RateFromPtr(&r, 1))

	ptr := r.Ptr()
	*ptr = 0.5
	assert.Equal(t, Rate(0.0825), r, "Ptr must return a pointer to a copy")

	assert.Equal(t, "n/a", nilRate.StringOr("n/a"))
	assert.Equal(t, "0.0825", r.Ptr().StringOr("n/a"))
	assert.Equal(t, 1.5, nilRate.FloatOr(1.5))
	assert.Equal(t, 0.0825, r.Ptr().FloatOr(1.5))
	assert.Equal(t, Rate(1.5), nilRate.RateOr(1.5))
	assert.Equal(t, r, r.Ptr().RateOr(1.5))
}

// TestRate_Equal pins that Equal compares the pointed to values
// (not the pointer identity) and treats nil as only equal to nil.
func TestRate_Equal(t *testing.T) {
	a, b, c := Rate(0.19), Rate(0.19), Rate(0.2)
	nan1, nan2 := Rate(math.NaN()), Rate(math.NaN())
	var nilRate *Rate

	assert.True(t, nilRate.Equal(nil))
	assert.False(t, a.Ptr().Equal(nil))
	assert.False(t, nilRate.Equal(&a))
	assert.True(t, (&a).Equal(&b), "different pointers to equal values")
	assert.False(t, (&a).Equal(&c))
	assert.True(t, (&nan1).Equal(&nan1), "the same pointer is always equal")
	assert.False(t, (&nan1).Equal(&nan2), "NaN values are never equal")
}

// TestRate_Valid pins that NaN and ±Inf, which ParseRate accepts and
// Inverse of zero produces, are rejected by all Valid methods,
// so they can't sneak into a conversion.
func TestRate_Valid(t *testing.T) {
	tests := []struct {
		name        string
		rate        Rate
		valid       bool
		positive    bool
		greaterZero bool
		smallerZero bool
	}{
		{name: "positive", rate: 0.19, valid: true, positive: true, greaterZero: true},
		{name: "zero", rate: 0, valid: true, positive: true},
		{name: "negative zero", rate: Rate(math.Copysign(0, -1)), valid: true, positive: true},
		{name: "negative", rate: -0.19, valid: true, smallerZero: true},
		{name: "NaN", rate: Rate(math.NaN())},
		{name: "+Inf", rate: Rate(math.Inf(1))},
		{name: "-Inf", rate: Rate(math.Inf(-1))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.rate.Valid(), "Valid")
			assert.Equal(t, tt.positive, tt.rate.ValidAndPositive(), "ValidAndPositive")
			assert.Equal(t, tt.greaterZero, tt.rate.ValidAndGreaterZero(), "ValidAndGreaterZero")
			assert.Equal(t, tt.smallerZero, tt.rate.ValidAndSmallerZero(), "ValidAndSmallerZero")
			assert.Equal(t, tt.valid, tt.rate.ValidAndHasSign(0), "ValidAndHasSign(0) accepts any sign")
		})
	}

	assert.True(t, Rate(0.19).ValidAndHasSign(1))
	assert.False(t, Rate(0.19).ValidAndHasSign(-1))
	assert.True(t, Rate(-0.19).ValidAndHasSign(-1))
	assert.False(t, Rate(-0.19).ValidAndHasSign(1))
	assert.False(t, Rate(math.Inf(1)).ValidAndHasSign(1))
	assert.False(t, Rate(math.Inf(-1)).ValidAndHasSign(-1))
}

// TestRate_BigFloat pins that BigFloat carries the exact float64 value,
// so high precision calculations start from the identical rate.
func TestRate_BigFloat(t *testing.T) {
	f, acc := Rate(0.19).BigFloat().Float64()
	assert.Equal(t, 0.19, f)
	assert.Equal(t, big.Exact, acc)
}
