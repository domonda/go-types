package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var currencyTestTable = []string{
	"€", "EUR",
	"EUR", "EUR",
	"eur", "EUR",
	"Euro", "EUR",
	"euro", "EUR",
	"EURO", "EUR",
	" EURO ", "EUR",

	"$", "USD",
	"USD", "USD",
	"usd", "USD",
	"US Dollar", "USD",
	"US$", "USD",
	"dollar", "USD",

	"£", "GBP",
	"GB£", "GBP",
	"Pound Sterling", "GBP",

	"₣", "CHF",
	"Swiss Franc", "CHF",
}

func Test_NormalizeCurrency(t *testing.T) {
	for i := 0; i < len(currencyTestTable); i += 2 {
		testVal, refVal := currencyTestTable[i], currencyTestTable[i+1]
		result, err := NormalizeCurrency(testVal)
		if err != nil {
			t.Errorf("NormalizeCurrency(%s): %s", testVal, err)
		}
		if string(result) != refVal {
			t.Errorf("NormalizeCurrency(%s): %s != %s", testVal, result, refVal)
		}
	}
}

func Test_NullCurrency(t *testing.T) {
	assert.False(t, Currency("").Valid())
	assert.True(t, NullCurrency("").Valid())
}
