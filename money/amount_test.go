package money

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
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
		{amount: 1, numAmounts: 17}:      {0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.06, 0.04},

		{amount: -100, numAmounts: 0}:     nil,
		{amount: -100.005, numAmounts: 1}: {-100.005}, // Single amount won't be rounded to cents
		{amount: -100.005, numAmounts: 2}: {-50, -50.01},
		{amount: -100, numAmounts: 3}:     {-33.33, -33.33, -33.34},
		{amount: -0.01, numAmounts: 5}:    {-0, -0, -0, -0, -0.01},
		{amount: -0.05, numAmounts: 5}:    {-0.01, -0.01, -0.01, -0.01, -0.01},
		{amount: -1, numAmounts: 17}:      {-0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.06, -0.04},
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
		{100, []Amount{11, 17, 37}, []Amount{16.92, 26.15, 56.93}},
		{1, []Amount{11, 17, 37}, []Amount{0.17, 0.26, 0.57}},
		{100, []Amount{1, 0, 1}, []Amount{50, 0, 50}},

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
		{-100, []Amount{-11, 17, 37}, []Amount{-16.92, -26.15, -56.93}},
		{-1, []Amount{-11, 17, 37}, []Amount{-0.17, -0.26, -0.57}},
		{-100, []Amount{1, 0, 1}, []Amount{-50, 0, -50}},
	}
	for _, test := range data {
		result := test.amount.SplitProportionally(test.weights)
		assert.Equal(t, test.expected, result)
	}
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
	tests := []struct {
		a Amount
		s string
	}{
		{a: 0, s: "0"},
		{a: -0, s: "0"},
		{a: 0.001, s: "0.001000000000000000020816681711721685132943093776702880859375"},
		{a: 0.001000000000000000020816681711721685132943093776702880859375, s: "0.001000000000000000020816681711721685132943093776702880859375"},
		{a: -100000.99, s: "-100000.990000000005238689482212066650390625"},
	}
	for _, tt := range tests {
		if got := tt.a.GoString(); got != tt.s {
			t.Errorf("Amount.GoString() = %v, want %v", got, tt.s)
		}
	}
}
