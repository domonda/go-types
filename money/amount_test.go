package money

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var amountTable2Decimals = map[string]Amount{
	"22.00":  22.00,
	"123.45": 123.45,
	"123,45": 123.45,
	"0,99":   0.99,

	"1,234.56":       1234.56,
	"1,234,567.89":   1234567.89,
	"10,234,567.89":  10234567.89,
	"100,234,567.89": 100234567.89,

	"1.234,56":       1234.56,
	"1.234.567,89":   1234567.89,
	"10.234.567,89":  10234567.89,
	"100.234.567,89": 100234567.89,

	"-22.00":  -22.00,
	"-123.45": -123.45,
	"-123,45": -123.45,
	"-0,99":   -0.99,

	"-1,234.56":       -1234.56,
	"-1,234,567.89":   -1234567.89,
	"-10,234,567.89":  -10234567.89,
	"-100,234,567.89": -100234567.89,

	"-1.234,56":       -1234.56,
	"-1.234.567,89":   -1234567.89,
	"-10.234.567,89":  -10234567.89,
	"-100.234.567,89": -100234567.89,

	"NaN":  Amount(math.NaN()),
	"Inf":  Amount(math.Inf(1)),
	"+Inf": Amount(math.Inf(1)),
	"-Inf": Amount(math.Inf(-1)),
}

var invalidAmounts = []string{
	"EUR 123",
	"1234.123.12,0",
	"1,234,56",
	"10,2340,560",
	// "1000,234,560",
}

var nonStandardFormatted = map[string]Amount{
	"3":            3,
	"666":          666,
	"123,4":        123.4,
	"123.4":        123.4,
	"1.2345":       1.2345,
	"1,2345":       1.2345,
	"1000000.8989": 1000000.8989,
	"1000000,8989": 1000000.8989,
	"1,234,567":    1234567,
	"1.234.567":    1234567,
	"1845.1800":    1845.1800,
	"-153.8900":    -153.8900,
	"-91.1000":     -91.1000,

	// Commonly produced by PHP instead of zero
	"3.5527136788005E-15": 3.5527136788005e-15,
	"3.5527136788005e-15": 3.5527136788005e-15,
}

func Test_ParseAmount(t *testing.T) {
	for str, refAmount := range amountTable2Decimals {
		amount, err := ParseAmount(str, 2)
		if err != nil {
			t.Errorf("Could not parse amount %s because of error: '%s'", str, err)
		}
		if !equalInclNaN(float64(amount), float64(refAmount)) {
			t.Errorf("Parsed '%s' amount %f != %f", str, amount, refAmount)
		}
	}
	for _, str := range invalidAmounts {
		amount, err := ParseAmount(str, 2)
		if err == nil {
			t.Errorf("Parsed invalid amount '%s' as %f", str, amount)
		}
	}
	for str, refAmount := range nonStandardFormatted {
		amount, err := ParseAmount(str)
		if err != nil {
			t.Errorf("Could not parse amount %s because of error: '%s'", str, err)
		}
		if !equalInclNaN(float64(amount), float64(refAmount)) {
			t.Errorf("Parsed '%s' amount %f != %f", str, amount, refAmount)
		}
	}
}

// TestParseAmount_oneSign protects that an amount with a sign at both ends
// is an error instead of one of the signs winning: reading "+1.5-" as -1.5
// or "-1+" as -1 books a misread value with a silently chosen sign.
// One leading or trailing sign stays valid.
func TestParseAmount_oneSign(t *testing.T) {
	for _, str := range []string{"+1.5-", "-1.5+", "-1+", "- 1.234,56 +", "+1+", "-1.5-"} {
		amount, err := ParseAmount(str)
		assert.Error(t, err, "ParseAmount(%q) = %v", str, amount)
	}
	for str, want := range map[string]Amount{"-1.5": -1.5, "1.5-": -1.5, "+1.5": 1.5, "1.5+": 1.5} {
		amount, err := ParseAmount(str)
		if assert.NoError(t, err, "ParseAmount(%q)", str) {
			assert.Equal(t, want, amount, "ParseAmount(%q)", str)
		}
	}
}

var stringTable = map[Amount]string{
	0:           "0.00",
	0.99:        "0.99",
	-0.99:       "-0.99",
	1:           "1.00",
	-1:          "-1.00",
	20:          "20.00",
	166:         "166.00",
	1.01:        "1.01",
	1.05:        "1.05",
	123456:      "123456.00",
	123456.789:  "123456.79",
	-123456.789: "-123456.79",
	0.055123:    "0.06",
	0.054123:    "0.05",
}

func Test_Amount_String(t *testing.T) {
	for amount, refstr := range stringTable {
		str := amount.String()
		if str != refstr {
			t.Errorf("%v to string is '%s' but should be '%s'", float64(amount), str, refstr)
		}
	}
}

func Test_Amount_RoundToCents(t *testing.T) {
	withoutRoudingError := Amount(137.89)
	withRoundingError := Amount(137.89000000000001)
	assert.NotEqual(t, withoutRoudingError, withRoundingError)

	// Example from production code:
	var amount Amount = 137.89
	var discountPercent int // 0
	var fee Amount          // 0
	total := (amount - (amount * (Amount(discountPercent) / 100)) + fee)
	assert.Equal(t, total, total.RoundToCents())

	// Create always the same pseude random list of integers
	r := rand.New(rand.NewSource(9371))
	testIntegers := make([]int, 1000)
	for i := range testIntegers {
		testIntegers[i] = r.Intn(100000)
		if i&1 == 1 {
			testIntegers[i] = -testIntegers[i]
		}
	}
	// Create all possible cent amounts
	allPossibleCents := make([]int, 100)
	for i := range allPossibleCents {
		allPossibleCents[i] = i
	}

	for _, integer := range testIntegers {
		for _, cents := range allPossibleCents {
			refAmount := Amount(integer) + Amount(cents)*Amount(0.01)
			assert.Equal(t, refAmount, refAmount.RoundToCents())
		}
	}

	roundToCentsTable := map[Amount]Amount{
		137.89000000000001: 137.89,
		0.001:              0,
		0.004:              0,
		0.005:              0.01,
		0.009:              0.01,
		0.099:              0.1,
		9999999.001:        9999999,
		9999999.004:        9999999,
		9999999.005:        9999999.01,
		9999999.009:        9999999.01,
		19999999.55:        19999999.55,
		99.999:             100,
		89.99999:           90,
		189.99999:          190,
		123456789.9999:     123456790,
	}
	for testAmount, refAmount := range roundToCentsTable {
		assert.Equal(t, refAmount, testAmount.RoundToCents())
	}
}

func Test_Amount_SplitEqually(t *testing.T) {
	type input struct {
		amount     Amount
		numAmounts int
	}
	data := map[input][]Amount{
		{amount: 100, numAmounts: 0}:     nil,
		{amount: 100.005, numAmounts: 1}: {100.005}, // Single amount won't be rounded to cents
		{amount: 100.005, numAmounts: 2}: {50, 50.01},
		{amount: 100, numAmounts: 3}:     {33.33, 33.33, 33.34},
		{amount: 0.01, numAmounts: 5}:    {0, 0, 0, 0, 0.01},
		{amount: 0.05, numAmounts: 5}:    {0.01, 0.01, 0.01, 0.01, 0.01},
		// The last amount used to get the whole rounding difference:
		// 16 times 0.06 and 0.04, which is 1.88 cents off its fair share
		// of 5.88 cents. The 15 cents that can't be divided evenly
		// now go to the last 15 amounts like CentAmount.SplitEqually does.
		{amount: 1, numAmounts: 17}: {0.05, 0.05, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06},

		{amount: -100, numAmounts: 0}:     nil,
		{amount: -100.005, numAmounts: 1}: {-100.005}, // Single amount won't be rounded to cents
		{amount: -100.005, numAmounts: 2}: {-50, -50.01},
		{amount: -100, numAmounts: 3}:     {-33.33, -33.33, -33.34},
		{amount: -0.01, numAmounts: 5}:    {-0, -0, -0, -0, -0.01},
		{amount: -0.05, numAmounts: 5}:    {-0.01, -0.01, -0.01, -0.01, -0.01},
		{amount: -1, numAmounts: 17}:      {-0.05, -0.05, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06},
	}
	for input, expected := range data {
		result := input.amount.SplitEqually(input.numAmounts)
		assert.Equal(t, expected, result)
	}
}

func Test_Amount_SplitProportionally(t *testing.T) {
	data := []struct {
		amount   Amount
		weights  []Amount
		expected []Amount
	}{
		{0, nil, nil},
		{0, []Amount{1}, []Amount{0}},
		{100, []Amount{1}, []Amount{100}},
		{100, []Amount{50.4444345}, []Amount{100}},
		{100, []Amount{100}, []Amount{100}},
		{100, []Amount{1000}, []Amount{100}},
		{100, []Amount{1, 2}, []Amount{33.33, 66.67}},
		{100, []Amount{1, 2, 3}, []Amount{16.67, 33.33, 50}},
		{100, []Amount{1, 2, 3, 4}, []Amount{10, 20, 30, 40}},
		{100, []Amount{-1, +2, -3, +4}, []Amount{10, 20, 30, 40}},
		// The exact shares are 16.923, 26.154 and 56.923. The last amount used
		// to get the whole rounding difference (26.15 and 56.93), now the cent
		// lost to truncation goes to the largest truncated fraction of 26.154
		// like CentAmount.SplitProportionally does.
		{100, []Amount{11, 17, 37}, []Amount{16.92, 26.16, 56.92}},
		{1, []Amount{11, 17, 37}, []Amount{0.17, 0.26, 0.57}},
		{100, []Amount{1, 0, 1}, []Amount{50, 0, 50}},
		{250.00, []Amount{333.33, 83.33, 0.01}, []Amount{200, 50, 0}}, // Note that the 1 cent amount got rounded to 0

		{-0, nil, nil},
		{-0, []Amount{-1}, []Amount{-0}},
		{-100, []Amount{-1}, []Amount{-100}},
		{-100, []Amount{-50.4444345}, []Amount{-100}},
		{-100, []Amount{-100}, []Amount{-100}},
		{-100, []Amount{-1000}, []Amount{-100}},
		{-100, []Amount{-1, 2}, []Amount{-33.33, -66.67}},
		{-100, []Amount{-1, 2, 3}, []Amount{-16.67, -33.33, -50}},
		{-100, []Amount{-1, 2, 3, 4}, []Amount{-10, -20, -30, -40}},
		{-100, []Amount{-1, +2, -3, +4}, []Amount{-10, -20, -30, -40}},
		{-100, []Amount{-11, 17, 37}, []Amount{-16.92, -26.16, -56.92}},
		{-1, []Amount{-11, 17, 37}, []Amount{-0.17, -0.26, -0.57}},
		{-100, []Amount{1, 0, 1}, []Amount{-50, 0, -50}},
	}
	for _, test := range data {
		result := test.amount.SplitProportionally(test.weights)
		assert.Equal(t, test.expected, result)
	}
}

// checkAmountSplit asserts the apportionment contract shared by the Amount
// and the CentAmount splits for the cents of the split amount and the
// exact shares of them: the parts sum to the cents of the split amount, no
// part is a cent or more away from its exact share, and no part has the
// opposite sign of the split amount.
func checkAmountSplit(t *testing.T, amount Amount, shares []float64, parts []Amount, msg string) {
	t.Helper()

	if !assert.Len(t, parts, len(shares), msg) {
		return
	}
	var sum int64
	for i, part := range parts {
		cents := part.Cents()
		sum += cents
		assert.Equal(t, part, part.RoundToCents(), "%s: part %d is rounded to cents", msg, i)
		assert.Less(t, math.Abs(float64(cents)-shares[i]), 1.0, "%s: part %d = %v is less than one cent from its share of %v cents", msg, i, part, shares[i])
		assert.False(t, (amount > 0 && cents < 0) || (amount < 0 && cents > 0), "%s: part %d = %v has the sign of %v", msg, i, part, amount)
	}
	assert.Equal(t, amount.Cents(), sum, "%s: parts sum to the cents of the amount", msg)
}

// TestAmount_SplitEqually_apportionment protects that the parts of an equal
// split sum to the cents of the amount, are less than one cent from their
// fair share and never flip the sign. Handing the whole rounding difference
// to the last part broke the last two: Amount(1).SplitEqually(200) rounded
// each half cent share up to 0.01, so the last part was 1 - 1.99 = -0.99
// for an amount that only has positive shares.
func TestAmount_SplitEqually_apportionment(t *testing.T) {
	for _, amount := range []Amount{1, -1, 0.01, 100, 100.005, -100.005, 1234.56, 0.07, 99999.99} {
		for _, count := range []int{2, 3, 7, 17, 200, 1000} {
			parts := amount.SplitEqually(count)
			share := float64(amount.Cents()) / float64(count)
			shares := make([]float64, count)
			for i := range shares {
				shares[i] = share
			}
			msg := fmt.Sprintf("Amount(%v).SplitEqually(%d)", amount, count)
			checkAmountSplit(t, amount, shares, parts, msg)

			// Exactly the cents of CentAmount.SplitEqually
			for i, cents := range amount.CentAmount(RoundHalfAwayFromZero).SplitEqually(count) {
				assert.Equal(t, cents.Amount(), parts[i], "%s: part %d like CentAmount", msg, i)
			}
		}
	}

	parts := Amount(1).SplitEqually(200)
	assert.Equal(t, Amount(0), parts[0], "a half cent share rounds down for the first parts")
	assert.Equal(t, Amount(0.01), parts[199], "and up for the last ones")

	// Without cents to apportion the amount is divided without rounding
	for _, part := range Amount(math.NaN()).SplitEqually(3) {
		assert.True(t, part.IsNaN())
	}
	assert.Equal(t, []Amount{Amount(math.Inf(-1)), Amount(math.Inf(-1))}, Amount(math.Inf(-1)).SplitEqually(2))
	assert.Equal(t, []Amount{1e20 / 4, 1e20 / 4, 1e20 / 4, 1e20 / 4}, Amount(1e20).SplitEqually(4), "beyond the CentAmount range")
}

// TestAmount_SplitProportionally_apportionment protects that the parts of a
// proportional split sum to the cents of the amount, are less than one cent
// from their exact share and never flip the sign, because the cents lost to
// rounding go to the largest truncated fractions. Handing the whole rounding
// difference to the last part flipped its sign for many small shares that
// all rounded up, like 200 equal weights of one euro.
func TestAmount_SplitProportionally_apportionment(t *testing.T) {
	equalWeights := make([]Amount, 200)
	for i := range equalWeights {
		equalWeights[i] = 1
	}
	weightSets := [][]Amount{
		equalWeights,
		{1, 2},
		{11, 17, 37},
		{-11, 17, -37},
		{333.33, 83.33, 0.01},
		{0.001, 0.002, 0.003},
		{1e-300, 1, 1e300},
		{0, 0, 5},
		{7, 0, 0, 3},
		{0.3, 0.3, 0.3, 0.1},
	}
	for _, amount := range []Amount{1, -1, 0.01, 100, -100, 1234.56, 0.07, 99999.99, 250} {
		for _, weights := range weightSets {
			parts := amount.SplitProportionally(weights)
			var totalWeight float64
			for _, weight := range weights {
				totalWeight += weight.AbsFloat()
			}
			shares := make([]float64, len(weights))
			for i, weight := range weights {
				shares[i] = float64(amount.Cents()) * weight.AbsFloat() / totalWeight
			}
			checkAmountSplit(t, amount, shares, parts, fmt.Sprintf("Amount(%v).SplitProportionally(%v)", amount, weights))
		}
	}

	// A half cent share rounds up for one half of the parts
	// and down for the other half
	parts := Amount(1).SplitProportionally(equalWeights)
	numCents := 0
	for _, part := range parts {
		if part == 0.01 {
			numCents++
		}
	}
	assert.Equal(t, 100, numCents, "parts of 0.01 in %v", parts)

	// Weights below a cent of the amount are still proportional
	assert.Equal(t, []Amount{16.67, 33.33, 50}, Amount(100).SplitProportionally([]Amount{0.001, 0.002, 0.003}))

	// If all weights are zero the whole amount ends up in the last part
	// like for CentAmount.SplitProportionally instead of NaN parts
	assert.Equal(t, []Amount{0, 0, 100.01}, Amount(100.005).SplitProportionally([]Amount{0, 0, 0}))

	// Without cents to apportion the amount is split without rounding
	for _, part := range Amount(1).SplitProportionally([]Amount{1, Amount(math.NaN())}) {
		assert.True(t, part.IsNaN())
	}
	assert.Equal(t, []Amount{2.5e19, 7.5e19}, Amount(1e20).SplitProportionally([]Amount{1, 3}), "beyond the CentAmount range")
}

func TestAmount_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		json       []byte
		wantAmount Amount
		wantErr    bool
	}{
		{name: "nil", json: nil, wantErr: true},
		{name: "JSON boolean", json: []byte(`true`), wantErr: true},
		{name: "JSON non number string", json: []byte(`"true"`), wantErr: true},

		{name: "JSON null", json: []byte(`null`), wantAmount: 0},
		{name: "empty JSON string", json: []byte(`""`), wantAmount: 0},
		{name: "0", json: []byte(`0`), wantAmount: 0},
		{name: "-0.12345", json: []byte(`-0.12345`), wantAmount: -0.12345},
		{name: "naked NaN", json: []byte(`NaN`), wantAmount: Amount(math.NaN())},
		{name: "quoted NaN", json: []byte(`"NaN"`), wantAmount: Amount(math.NaN())},
		{name: "naked Inf", json: []byte(`Inf`), wantAmount: Amount(math.Inf(1))},
		{name: "quoted Inf", json: []byte(`"Inf"`), wantAmount: Amount(math.Inf(1))},
		{name: "naked +Inf", json: []byte(`+Inf`), wantAmount: Amount(math.Inf(1))},
		{name: "quoted +Inf", json: []byte(`"+Inf"`), wantAmount: Amount(math.Inf(1))},
		{name: "naked -Inf", json: []byte(`-Inf`), wantAmount: Amount(math.Inf(-1))},
		{name: "quoted -Inf", json: []byte(`"-Inf"`), wantAmount: Amount(math.Inf(-1))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var amount Amount = 666 // Init with value different from default 0
			err := amount.UnmarshalJSON(tt.json)
			if (err != nil) != tt.wantErr {
				t.Errorf("Amount.UnmarshalJSON(%s) error = %v, wantErr %v", tt.json, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalInclNaN(float64(amount), float64(tt.wantAmount)) {
				t.Errorf("Amount.UnmarshalJSON(%s) got %f, want %f", tt.json, amount, tt.wantAmount)
			}
		})
	}
}

// TestAmount_UnmarshalJSON_stringEscapes protects that a JSON string is
// decoded as JSON before it is parsed as amount, like float.Tolerant and
// DecimalAmount do. A JSON encoder may escape any rune, so "1\u002c5" is
// the same JSON value as "1,5" and must decode to the same amount instead
// of failing on the backslash of the escape.
func TestAmount_UnmarshalJSON_stringEscapes(t *testing.T) {
	valid := map[string]Amount{
		"\"1\\u002c5\"":           1.5,
		"\"\\u0031.234,56\"":      1234.56,
		"\"\\u002d1,5\"":          -1.5,
		"\"1.5\\t\"":              1.5,
		"\"\\u004e\\u0061N\"":     Amount(math.NaN()),
		"\"1\\u0027234\\u002e5\"": 1234.5,
	}
	for j, want := range valid {
		var amount Amount
		if assert.NoError(t, json.Unmarshal([]byte(j), &amount), "json.Unmarshal(%s)", j) {
			assert.True(t, equalInclNaN(float64(want), float64(amount)), "json.Unmarshal(%s) = %v, want %v", j, amount, want)
		}
	}

	var s struct{ A Amount }
	require.NoError(t, json.Unmarshal([]byte("{\"A\":\"1\\u002c5\"}"), &s))
	assert.Equal(t, Amount(1.5), s.A)

	for _, j := range []string{"\"1\\u002c5", "\"1\\x2c5\"", "\"1\\u002c5\\u002c5\""} {
		amount := Amount(666)
		assert.Error(t, amount.UnmarshalJSON([]byte(j)), "Amount.UnmarshalJSON(%s)", j)
		assert.Equal(t, Amount(666), amount, "a failed Amount.UnmarshalJSON(%s) must not modify the amount", j)
	}
}

func equalInclNaN(a, b float64) bool {
	if math.IsNaN(a) {
		return math.IsNaN(b)
	}
	if math.IsNaN(b) {
		return false
	}
	return a == b
}

func TestAmount_GoString(t *testing.T) {
	// The Go source representation must compile back to the same
	// value, so the float literal keeps the full float64 precision
	// and the non-finite values use their math package constructors.
	tests := []struct {
		amount   Amount
		expected string
	}{
		{0, "money.Amount(0)"},
		{5, "money.Amount(5)"},
		{-5, "money.Amount(-5)"},
		{0.5, "money.Amount(0.5)"},
		{123.45, "money.Amount(123.4500000000000028421709430404007434844970703125)"},
		{Amount(math.NaN()), "money.Amount(math.NaN())"},
		{Amount(math.Inf(1)), "money.Amount(math.Inf(1))"},
		{Amount(math.Inf(-1)), "money.Amount(math.Inf(-1))"},
		// The Go literal -0 is the untyped constant zero, so a bare "-0"
		// would silently turn negative zero into positive zero.
		{Amount(math.Copysign(0, -1)), "money.Amount(math.Copysign(0, -1))"},
		// Carried over from the pre-existing test of the old bare-number
		// format, which this change replaces
		{0.001, "money.Amount(0.001000000000000000020816681711721685132943093776702880859375)"},
		{0.001000000000000000020816681711721685132943093776702880859375, "money.Amount(0.001000000000000000020816681711721685132943093776702880859375)"},
		{-100000.99, "money.Amount(-100000.990000000005238689482212066650390625)"},

		// Subnormals and other tiny values need far more than the
		// 200 decimal places that used to silently round them to zero
		{math.SmallestNonzeroFloat64, "money.Amount(" + exactDecimalExpansion(math.SmallestNonzeroFloat64) + ")"},
		{1e-250, "money.Amount(" + exactDecimalExpansion(1e-250) + ")"},
	}
	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.amount.GoString())
			assert.Equal(t, test.expected, fmt.Sprintf("%#v", test.amount))
		})
	}
}

func TestAmount_GoString_fullPrecision(t *testing.T) {
	// The defining property of the full precision representation:
	// the printed literal must parse back to the identical float64,
	// for the extremes of the float64 range as well as for random values.
	amounts := []Amount{
		0,
		Amount(math.Copysign(0, -1)),
		math.SmallestNonzeroFloat64,
		-math.SmallestNonzeroFloat64,
		math.MaxFloat64,
		-math.MaxFloat64,
		1e-250,
		2.5e-300,
		Amount(math.Float64frombits(1)),  // smallest subnormal
		Amount(math.Float64frombits(42)), // arbitrary subnormal
	}
	r := rand.New(rand.NewSource(9371))
	for range 1000 {
		amounts = append(amounts, Amount(r.NormFloat64()*math.Pow10(r.Intn(600)-300)))
	}
	for _, amount := range amounts {
		literal := strings.TrimSuffix(strings.TrimPrefix(amount.GoString(), "money.Amount("), ")")
		if strings.HasPrefix(literal, "math.") {
			// Negative zero and the non-finite values can't be written as a
			// float literal at all, so they carry a constructor instead and
			// are pinned by their exact string in TestAmount_GoString.
			continue
		}
		parsed, err := strconv.ParseFloat(literal, 64)
		require.NoError(t, err, "GoString of %g must be a parsable float literal", float64(amount))
		assert.Equal(t, float64(amount), parsed, "GoString of %g lost precision", float64(amount))
		// float64 equality treats -0 and +0 as equal, so the sign bit
		// needs its own assertion to catch a lost negative zero.
		assert.Equal(t, amount.Signbit(), math.Signbit(parsed), "GoString of %g lost the sign bit", float64(amount))
	}
}

// exactDecimalExpansion returns the exact decimal value of f
// computed independently of Amount.GoString via math/big.
func exactDecimalExpansion(f float64) string {
	exact := new(big.Rat).SetFloat64(f).FloatString(1074)
	return strings.TrimRight(strings.TrimRight(exact, "0"), ".")
}

func TestAmount_CentAmount(t *testing.T) {
	// The rounding mode is applied to the shortest decimal representation
	// of the float, not to float64(a)*100, so a decimal literal rounds the
	// way the literal reads. Amount(0.145)*100 is 14.499999999999998 and
	// would round to 14 the naive way.
	assert.Equal(t, CentAmount(12345), Amount(123.45).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, CentAmount(-12345), Amount(-123.45).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, CentAmount(12346), Amount(123.456).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, CentAmount(1), Amount(0.005).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, CentAmount(15), Amount(0.145).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, CentAmount(14), Amount(0.145).CentAmount(RoundDown))
	assert.Equal(t, CentAmount(14), Amount(0.145).CentAmount(RoundHalfToEven))

	// Non-finite amounts have no CentAmount representation,
	// but must not produce a platform dependent int64.
	assert.Equal(t, CentAmount(0), Amount(math.NaN()).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, MaxCentAmount, Amount(math.Inf(1)).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, MinCentAmount, Amount(math.Inf(-1)).CentAmount(RoundHalfAwayFromZero))
	assert.Equal(t, MaxCentAmount, Amount(1e30).CentAmount(RoundHalfAwayFromZero))
}

// TestNullableAmount pins that NullableAmount distinguishes a zero amount
// (for example a free item) from a missing amount in Go, JSON and SQL.
func TestNullableAmount(t *testing.T) {
	t.Run("constructors", func(t *testing.T) {
		zero := NullableAmountFrom(0)
		assert.True(t, zero.IsNotNull(), "a zero amount is a valid amount, not NULL")
		assert.Equal(t, Amount(0), zero.Get())

		assert.True(t, NullableAmountFromPtr(nil).IsNull())

		a := Amount(12.34)
		fromPtr := NullableAmountFromPtr(&a)
		a = 0
		assert.Equal(t, Amount(12.34), fromPtr.Get(), "NullableAmountFromPtr must copy the pointed to value")
	})

	t.Run("JSON", func(t *testing.T) {
		data, err := json.Marshal(NullableAmount{})
		require.NoError(t, err)
		assert.Equal(t, `null`, string(data))

		data, err = json.Marshal(NullableAmountFrom(0))
		require.NoError(t, err)
		assert.Equal(t, `0`, string(data))

		got := NullableAmountFrom(1)
		require.NoError(t, json.Unmarshal([]byte(`null`), &got))
		assert.True(t, got.IsNull(), "JSON null must clear a previously set amount")

		// The tolerant string parsing of Amount.UnmarshalJSON
		// must also work through the nullable wrapper
		require.NoError(t, json.Unmarshal([]byte(`"1.234,56"`), &got))
		assert.Equal(t, NullableAmountFrom(1234.56), got)
	})

	t.Run("SQL", func(t *testing.T) {
		v, err := NullableAmount{}.Value()
		require.NoError(t, err)
		assert.Nil(t, v, "NULL must be written to SQL as nil, not as 0")

		v, err = NullableAmountFrom(12.34).Value()
		require.NoError(t, err)
		assert.Equal(t, float64(12.34), v, "drivers need a plain float64, not the named Amount type")

		got := NullableAmountFrom(1)
		require.NoError(t, got.Scan(nil))
		assert.True(t, got.IsNull(), "SQL NULL must scan as NULL")

		// Postgres NUMERIC columns arrive as text bytes
		require.NoError(t, got.Scan([]byte("1234.56")))
		assert.Equal(t, NullableAmountFrom(1234.56), got)
	})
}

// TestAmount_ScanString pins that amounts from CSV and bank exports in the
// common real-world notations (German decimal comma, English decimal point,
// French space and Swiss apostrophe grouping, leading or trailing sign) all
// scan to the same value, and that validate rejects NaN and ±Inf.
func TestAmount_ScanString(t *testing.T) {
	tests := []struct {
		source   string
		validate bool
		want     Amount
		wantErr  bool
	}{
		{source: "1.234,56", validate: true, want: 1234.56},
		{source: "1,234.56", validate: true, want: 1234.56},
		{source: "1 234,56", validate: true, want: 1234.56},
		{source: "1'234.56", validate: true, want: 1234.56},
		{source: "1234,56", validate: true, want: 1234.56},
		{source: " 1234.56 ", validate: true, want: 1234.56},
		{source: "-12,34", validate: true, want: -12.34},
		{source: "12,34-", validate: true, want: -12.34}, // trailing minus of accounting exports
		{source: "+12,34", validate: true, want: 12.34},
		{source: "1.234.567", validate: true, want: 1234567}, // repeated group separator is not a decimal separator
		{source: "NaN", validate: false, want: Amount(math.NaN())},
		{source: "-Inf", validate: false, want: Amount(math.Inf(-1))},
		{source: "NaN", validate: true, wantErr: true},
		{source: "Inf", validate: true, wantErr: true},
		{source: "", validate: false, wantErr: true},
		{source: "12,34 EUR", validate: false, wantErr: true},
		{source: "1,23,45", validate: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q validate=%t", tt.source, tt.validate), func(t *testing.T) {
			a := Amount(666)
			err := a.ScanString(tt.source, tt.validate)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, Amount(666), a, "a failed scan must not modify the amount")
				return
			}
			require.NoError(t, err)
			assert.True(t, equalInclNaN(float64(tt.want), float64(a)), "got %v, want %v", float64(a), float64(tt.want))
		})
	}
}

// TestAmount_Cents pins that Cents rounds half away from zero
// symmetrically for positive and negative amounts.
func TestAmount_Cents(t *testing.T) {
	tests := []struct {
		amount Amount
		want   int64
	}{
		{amount: 0, want: 0},
		{amount: 1234.56, want: 123456},
		{amount: -1234.56, want: -123456},
		{amount: 0.125, want: 13}, // exactly representable half cent
		{amount: -0.125, want: -13},
		{amount: 0.1 + 0.2, want: 30}, // binary noise must not change the cents
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.amount.Cents(), "Amount(%v).Cents()", float64(tt.amount))
	}
}

// TestAmount_RoundingConventionAgrees pins that every way of rounding or
// formatting an Amount to cents gives the same result for half cent ties.
//
// The contract is to round the shortest decimal representation of the
// float64, which is the decimal a human typed or a JSON number literal
// carried, half away from zero. The binary value of most such literals is
// slightly off: 0.145 is stored as 0.14499999999999999001, 2.675 as
// 2.67499999999999982236. Rounding the binary value (or multiplying it by
// 100 first) turns those into 0.14 and 2.67, so the same invoice amount
// would have 15 cents in one place and 14 in another, depending on which
// method computed it.
func TestAmount_RoundingConventionAgrees(t *testing.T) {
	tests := []struct {
		amount Amount
		cents  int64
		str    string
	}{
		{amount: 0.145, cents: 15, str: "0.15"},
		{amount: -0.145, cents: -15, str: "-0.15"},
		{amount: 0.125, cents: 13, str: "0.13"}, // exactly representable tie
		{amount: -0.125, cents: -13, str: "-0.13"},
		{amount: 1.005, cents: 101, str: "1.01"},
		{amount: 2.675, cents: 268, str: "2.68"},
		{amount: -2.675, cents: -268, str: "-2.68"},
		{amount: 0.285, cents: 29, str: "0.29"},
		{amount: 1.115, cents: 112, str: "1.12"},
		{amount: 8.345, cents: 835, str: "8.35"},
		{amount: 1234.565, cents: 123457, str: "1234.57"},
		{amount: 0.144999, cents: 14, str: "0.14"}, // not a tie
		{amount: 0.1 + 0.2, cents: 30, str: "0.30"},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			a := tt.amount
			rounded := Amount(tt.cents) / 100 // the float64 nearest to the rounded decimal
			assert.Equal(t, tt.cents, a.Cents(), "Cents")
			assert.Equal(t, CentAmount(tt.cents), a.CentAmount(RoundHalfAwayFromZero), "CentAmount")
			assert.Equal(t, rounded, a.RoundToCents(), "RoundToCents")
			assert.Equal(t, rounded, a.RoundToDecimals(2), "RoundToDecimals(2)")
			assert.Equal(t, NewDecimalAmount(tt.cents, 2), a.DecimalAmount(2, RoundHalfAwayFromZero), "DecimalAmount(2)")
			assert.Equal(t, NewDecimalAmount(tt.cents, 2), DecimalAmountFrom(a).RoundToCents(RoundHalfAwayFromZero), "DecimalAmountFrom")
			assert.Equal(t, tt.str, a.String(), "String")
			assert.Equal(t, tt.str, a.FormatSep(0, '.', 2), "FormatSep")
			assert.Equal(t, strings.Replace(tt.str, ".", ",", 1), a.FormatSep(0, ',', 2), "FormatSep decimal comma")
			assert.Equal(t, "EUR "+tt.str, CurrencyAmount{Currency: EUR, Amount: a}.String(), "CurrencyAmount.String")
			assert.Equal(t, tt.str+" EUR", CurrencyAmount{Currency: EUR, Amount: a}.FormatSep(false, 0, '.', 2), "CurrencyAmount.FormatSep")
			assert.Equal(t, Rate(rounded), Rate(a).RoundToDecimals(2), "Rate.RoundToDecimals(2)")
			assert.Equal(t, tt.str, Rate(a).FormatSep(0, '.', 2), "Rate.FormatSep")
		})
	}
}

// TestAmount_RoundingOtherDecimals pins that RoundToDecimals, FormatSep
// and DecimalAmount round the shortest decimal representation at any
// number of decimal places, not only at cents, and that the non-finite
// values pass through every method unchanged.
func TestAmount_RoundingOtherDecimals(t *testing.T) {
	// 1.0005 is stored as 1.000499999999999989, 1.45 as 1.4499999999999999556
	assert.Equal(t, Amount(1.001), Amount(1.0005).RoundToDecimals(3))
	assert.Equal(t, "1.001", Amount(1.0005).FormatSep(0, '.', 3))
	assert.Equal(t, NewDecimalAmount(1001, 3), Amount(1.0005).DecimalAmount(3, RoundHalfAwayFromZero))
	assert.Equal(t, Amount(1.5), Amount(1.45).RoundToDecimals(1))
	assert.Equal(t, "1.5", Amount(1.45).FormatSep(0, '.', 1))
	assert.Equal(t, "2", Amount(1.5).FormatSep(0, '.', 0), "FormatSep must round half away from zero, not to even")
	assert.Equal(t, "-3", Amount(-2.5).FormatSep(0, '.', 0))
	assert.Equal(t, "1,234.57", Amount(1234.565).FormatSep(',', '.', 2))
	// Values with fewer decimals are padded, not rounded
	assert.Equal(t, "0.1450", Amount(0.145).FormatSep(0, '.', 4))
	assert.Equal(t, "0.145", Amount(0.145).FormatSep(0, '.', -1))
	// Rounding modes other than half away from zero also see the decimal
	assert.Equal(t, NewDecimalAmount(14, 2), Amount(0.145).DecimalAmount(2, RoundHalfToEven))
	assert.Equal(t, NewDecimalAmount(15, 2), Amount(0.141).DecimalAmount(2, RoundUp))
	assert.Equal(t, NewDecimalAmount(-15, 2), Amount(-0.141).DecimalAmount(2, RoundFloor))
	assert.Equal(t, NewDecimalAmount(0, 2), Amount(0.001).DecimalAmount(2, RoundDown))
	// Negative zero keeps its sign like math.Round does
	assert.True(t, Amount(-0.001).RoundToCents().Signbit(), "-0.001 rounds to -0")

	nan, posInf, negInf := Amount(math.NaN()), Amount(math.Inf(1)), Amount(math.Inf(-1))
	assert.True(t, nan.RoundToCents().IsNaN())
	assert.True(t, nan.RoundToDecimals(3).IsNaN())
	assert.Equal(t, posInf, posInf.RoundToCents())
	assert.Equal(t, negInf, negInf.RoundToDecimals(-2))
	assert.Equal(t, "NaN", nan.String())
	assert.Equal(t, "+Inf", posInf.String())
	assert.Equal(t, "-Inf", negInf.FormatSep(',', '.', 2))
	assert.True(t, nan.DecimalAmount(2, RoundHalfAwayFromZero).IsNaN())
	assert.True(t, negInf.DecimalAmount(2, RoundHalfAwayFromZero).IsInf(-1))
	// Beyond the DecimalAmount range a finite Amount maps to ±Inf
	assert.True(t, Amount(1e300).DecimalAmount(0, RoundHalfAwayFromZero).IsInf(1))
	assert.True(t, Amount(-1e17).DecimalAmount(2, RoundHalfAwayFromZero).IsInf(-1))
	// 2^58 has the shortest decimal 288230376151711740, which still fits the
	// coefficient maximum of 2^58-1, the next float64 288230376151711800 not
	assert.Equal(t, NewDecimalAmount(288230376151711740, 0), Amount(1<<58).DecimalAmount(0, RoundHalfAwayFromZero))
	assert.Equal(t, NewDecimalAmount(-288230376151711740, 0), Amount(-1<<58).DecimalAmount(0, RoundHalfAwayFromZero))
	assert.True(t, Amount(math.Nextafter(1<<58, math.Inf(1))).DecimalAmount(0, RoundHalfAwayFromZero).IsInf(1))
	assert.Panics(t, func() { Amount(1).DecimalAmount(MaxDecimalAmountScale+1, RoundHalfAwayFromZero) })
	assert.Panics(t, func() { Amount(1).DecimalAmount(-1, RoundHalfAwayFromZero) })
}

// TestAmount_WithinOneCent pins the tolerance used to compare amounts
// from different sources (e.g. invoice total vs. sum of line items)
// where rounding of intermediate results can differ by one cent.
func TestAmount_WithinOneCent(t *testing.T) {
	tests := []struct {
		a, b Amount
		want bool
	}{
		{a: 10, b: 10, want: true},
		{a: 10, b: 10.01, want: true},
		{a: 10, b: 9.99, want: true},
		{a: 10, b: 10.02, want: false},
		{a: 10, b: 9.98, want: false},
		{a: 0.1 + 0.2, b: 0.3, want: true},
		{a: -5, b: -5.01, want: true},
		{a: 0.01, b: -0.01, want: false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.a.WithinOneCent(tt.b), "Amount(%v).WithinOneCent(%v)", float64(tt.a), float64(tt.b))
		assert.Equal(t, tt.want, tt.b.WithinOneCent(tt.a), "WithinOneCent must be symmetric")
	}
}

// TestAmount_Rounding pins half away from zero rounding for RoundToInt
// and RoundToDecimals, that RoundToDecimals(2) agrees with RoundToCents,
// and that negative decimals round to tens, hundreds, etc.
func TestAmount_Rounding(t *testing.T) {
	assert.Equal(t, Amount(3), Amount(2.5).RoundToInt())
	assert.Equal(t, Amount(-3), Amount(-2.5).RoundToInt())
	assert.Equal(t, Amount(1), Amount(0.5).RoundToInt())
	assert.Equal(t, Amount(2), Amount(2.4999).RoundToInt())

	tests := []struct {
		amount   Amount
		decimals int
		want     Amount
	}{
		{amount: 1.23456, decimals: 3, want: 1.235},
		{amount: -1.23456, decimals: 3, want: -1.235},
		{amount: 1.23456, decimals: 0, want: 1},
		{amount: 123456, decimals: -3, want: 123000},
		{amount: 123500, decimals: -3, want: 124000},
		{amount: -123500, decimals: -3, want: -124000},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.amount.RoundToDecimals(tt.decimals), "Amount(%v).RoundToDecimals(%d)", float64(tt.amount), tt.decimals)
	}

	for _, a := range []Amount{0.004, 0.005, 1.999, -1.995, 123456.789, 9999999.005} {
		assert.Equal(t, a.RoundToCents(), a.RoundToDecimals(2), "Amount(%v)", float64(a))
		assert.Equal(t, a.RoundToInt(), a.RoundToDecimals(0), "Amount(%v)", float64(a))
	}
}

// TestAmount_PtrHelpers pins the nil handling of the pointer helpers
// used for optional amounts, and that StringOr formats like String,
// rounded to cents.
func TestAmount_PtrHelpers(t *testing.T) {
	var nilAmount *Amount
	a := Amount(1234.567)

	assert.Equal(t, Amount(1), AmountFromPtr(nil, 1))
	assert.Equal(t, a, AmountFromPtr(&a, 1))

	ptr := a.Ptr()
	*ptr = 0
	assert.Equal(t, Amount(1234.567), a, "Ptr must return a pointer to a copy")

	assert.Equal(t, "n/a", nilAmount.StringOr("n/a"))
	assert.Equal(t, "1234.57", a.Ptr().StringOr("n/a"))
	assert.Equal(t, 1.5, nilAmount.FloatOr(1.5))
	assert.Equal(t, 1234.567, a.Ptr().FloatOr(1.5))
	assert.Equal(t, Amount(1.5), nilAmount.AmountOr(1.5))
	assert.Equal(t, a, a.Ptr().AmountOr(1.5))
}

// TestAmount_Equal pins that Equal compares the pointed to values
// (not the pointer identity) and treats nil as only equal to nil.
func TestAmount_Equal(t *testing.T) {
	a, b, c := Amount(12.34), Amount(12.34), Amount(12.35)
	nan1, nan2 := Amount(math.NaN()), Amount(math.NaN())
	var nilAmount *Amount

	assert.True(t, nilAmount.Equal(nil))
	assert.False(t, a.Ptr().Equal(nil))
	assert.False(t, nilAmount.Equal(&a))
	assert.True(t, (&a).Equal(&b), "different pointers to equal values")
	assert.False(t, (&a).Equal(&c))
	assert.True(t, (&nan1).Equal(&nan1), "the same pointer is always equal")
	assert.False(t, (&nan1).Equal(&nan2), "NaN values are never equal")
}

// TestAmount_Sign pins the sign helpers that are used to book an amount
// as debit or credit regardless of the sign it was parsed with
// (e.g. credit notes with negative or positive totals).
func TestAmount_Sign(t *testing.T) {
	negZero := Amount(math.Copysign(0, -1))

	assert.Equal(t, Amount(12.34), Amount(-12.34).Abs())
	assert.Equal(t, Amount(math.Inf(1)), Amount(math.Inf(-1)).Abs())
	assert.True(t, Amount(math.NaN()).Abs().IsNaN())
	assert.False(t, negZero.Abs().Signbit())
	assert.Equal(t, 12.34, Amount(-12.34).AbsFloat())

	assert.Equal(t, Amount(-12.34), Amount(12.34).Inverted())
	assert.Equal(t, Amount(12.34), Amount(-12.34).Inverted())
	a := Amount(12.34)
	a.Invert()
	assert.Equal(t, Amount(-12.34), a, "Invert must change the amount in place")

	for _, a := range []Amount{12.34, -12.34} {
		assert.Equal(t, Amount(12.34), a.WithPosSign(true), "%v.WithPosSign(true)", float64(a))
		assert.Equal(t, Amount(-12.34), a.WithPosSign(false), "%v.WithPosSign(false)", float64(a))
		assert.Equal(t, Amount(-12.34), a.WithNegSign(true), "%v.WithNegSign(true)", float64(a))
		assert.Equal(t, Amount(12.34), a.WithNegSign(false), "%v.WithNegSign(false)", float64(a))
	}
	assert.False(t, negZero.WithPosSign(true).Signbit())
}

// TestAmount_RateArithmetic pins the typical tax and currency calculations:
// 20% VAT of a net amount via a rate or a percentage gives the same result,
// and dividing by a rate is the inverse of multiplying by it.
func TestAmount_RateArithmetic(t *testing.T) {
	assert.Equal(t, Amount(40), Amount(200).MultipliedByRate(0.2))
	assert.Equal(t, Amount(40), Amount(200).Percentage(20))
	assert.Equal(t, Amount(80), Amount(100).DividedByRate(1.25))
	assert.Equal(t, Amount(-50), Amount(-250).Percentage(20))
	assert.Equal(t, Amount(22.8), Amount(119.99).Percentage(19).RoundToCents())

	// Dividing by zero must produce a non-finite amount that Valid rejects
	assert.False(t, Amount(100).DividedByRate(0).Valid())
}

// TestAmount_Valid pins that NaN and ±Inf, which ParseAmount accepts
// without error, are rejected by all Valid methods.
func TestAmount_Valid(t *testing.T) {
	tests := []struct {
		name        string
		amount      Amount
		valid       bool
		positive    bool
		greaterZero bool
		smallerZero bool
		nan         bool
		inf         bool
	}{
		{name: "positive", amount: 12.34, valid: true, positive: true, greaterZero: true},
		{name: "zero", amount: 0, valid: true, positive: true},
		{name: "negative zero", amount: Amount(math.Copysign(0, -1)), valid: true, positive: true},
		{name: "negative", amount: -12.34, valid: true, smallerZero: true},
		{name: "NaN", amount: Amount(math.NaN()), nan: true},
		{name: "+Inf", amount: Amount(math.Inf(1)), inf: true},
		{name: "-Inf", amount: Amount(math.Inf(-1)), inf: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.amount.Valid(), "Valid")
			assert.Equal(t, tt.positive, tt.amount.ValidAndPositive(), "ValidAndPositive")
			assert.Equal(t, tt.greaterZero, tt.amount.ValidAndGreaterZero(), "ValidAndGreaterZero")
			assert.Equal(t, tt.smallerZero, tt.amount.ValidAndSmallerZero(), "ValidAndSmallerZero")
			assert.Equal(t, tt.nan, tt.amount.IsNaN(), "IsNaN")
			assert.Equal(t, tt.inf, tt.amount.IsInf(), "IsInf")
			assert.Equal(t, tt.valid, tt.amount.ValidAndHasSign(0), "ValidAndHasSign(0) accepts any sign")
		})
	}

	assert.True(t, Amount(12.34).ValidAndHasSign(1))
	assert.False(t, Amount(12.34).ValidAndHasSign(-1))
	assert.True(t, Amount(-12.34).ValidAndHasSign(-1))
	assert.False(t, Amount(-12.34).ValidAndHasSign(1))
	assert.False(t, Amount(math.Inf(1)).ValidAndHasSign(1))
	assert.False(t, Amount(math.Inf(-1)).ValidAndHasSign(-1))
}

// TestAmount_BigFloat pins that BigFloat carries the exact float64 value,
// so high precision calculations start from the identical amount.
func TestAmount_BigFloat(t *testing.T) {
	f, acc := Amount(1234.56).BigFloat().Float64()
	assert.Equal(t, 1234.56, f)
	assert.Equal(t, big.Exact, acc)
}
