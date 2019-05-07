package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var currencyIntAmountTable = map[string]CurrencyAmount{
	"1": CurrencyAmount{"", 1},

	"EUR1": CurrencyAmount{"EUR", 1},
	"1EUR": CurrencyAmount{"EUR", 1},
	"1$":   CurrencyAmount{"USD", 1},
	"$1":   CurrencyAmount{"USD", 1},

	"EUR 1": CurrencyAmount{"EUR", 1},
	"1 EUR": CurrencyAmount{"EUR", 1},
	"1 $":   CurrencyAmount{"USD", 1},
	"$ 1":   CurrencyAmount{"USD", 1},

	"EUR   1": CurrencyAmount{"EUR", 1},
	"1   EUR": CurrencyAmount{"EUR", 1},
	"1   $":   CurrencyAmount{"USD", 1},
	"$   1":   CurrencyAmount{"USD", 1},
}

var currencyAmountTable = map[string]CurrencyAmount{
	"EUR -12,34": CurrencyAmount{"EUR", -12.34},
	"EUR +12,34": CurrencyAmount{"EUR", 12.34},
	"EUR-12,34":  CurrencyAmount{"EUR", -12.34},
	"EUR+12,34":  CurrencyAmount{"EUR", 12.34},
	"EUR -12.34": CurrencyAmount{"EUR", -12.34},
	"EUR +12.34": CurrencyAmount{"EUR", 12.34},
	"EUR-12.34":  CurrencyAmount{"EUR", -12.34},
	"EUR+12.34":  CurrencyAmount{"EUR", 12.34},

	"EUR1.23": CurrencyAmount{"EUR", 1.23},
	"1.23EUR": CurrencyAmount{"EUR", 1.23},
	"1.23$":   CurrencyAmount{"USD", 1.23},
	"$1.23":   CurrencyAmount{"USD", 1.23},

	"EUR 1.23": CurrencyAmount{"EUR", 1.23},
	"1.23 EUR": CurrencyAmount{"EUR", 1.23},
	"1.23 $":   CurrencyAmount{"USD", 1.23},
	"$ 1.23":   CurrencyAmount{"USD", 1.23},

	"EUR   1.23": CurrencyAmount{"EUR", 1.23},
	"1.23   EUR": CurrencyAmount{"EUR", 1.23},
	"1.23   $":   CurrencyAmount{"USD", 1.23},
	"$   1.23":   CurrencyAmount{"USD", 1.23},

	"EUR1,234,567.89": CurrencyAmount{"EUR", 1234567.89},
	"1,234,567.89EUR": CurrencyAmount{"EUR", 1234567.89},
	"1,234,567.89$":   CurrencyAmount{"USD", 1234567.89},
	"$1,234,567.89":   CurrencyAmount{"USD", 1234567.89},

	"EUR 1,234,567.89": CurrencyAmount{"EUR", 1234567.89},
	"1,234,567.89 EUR": CurrencyAmount{"EUR", 1234567.89},
	"1,234,567.89 $":   CurrencyAmount{"USD", 1234567.89},
	"$ 1,234,567.89":   CurrencyAmount{"USD", 1234567.89},

	"EUR   1,234,567.89": CurrencyAmount{"EUR", 1234567.89},
	"1,234,567.89   EUR": CurrencyAmount{"EUR", 1234567.89},
	"1,234,567.89   $":   CurrencyAmount{"USD", 1234567.89},
	"$   1,234,567.89":   CurrencyAmount{"USD", 1234567.89},

	"EUR 1,23": CurrencyAmount{"EUR", 1.23},
	"1,23 EUR": CurrencyAmount{"EUR", 1.23},
	"1,23 $":   CurrencyAmount{"USD", 1.23},
	"$ 1,23":   CurrencyAmount{"USD", 1.23},

	"EUR   1,23": CurrencyAmount{"EUR", 1.23},
	"1,23   EUR": CurrencyAmount{"EUR", 1.23},
	"1,23   $":   CurrencyAmount{"USD", 1.23},
	"$   1,23":   CurrencyAmount{"USD", 1.23},

	"EUR1.234.567,89": CurrencyAmount{"EUR", 1234567.89},
	"1.234.567,89EUR": CurrencyAmount{"EUR", 1234567.89},
	"1.234.567,89$":   CurrencyAmount{"USD", 1234567.89},
	"$1.234.567,89":   CurrencyAmount{"USD", 1234567.89},

	"EUR 1.234.567,89": CurrencyAmount{"EUR", 1234567.89},
	"1.234.567,89 EUR": CurrencyAmount{"EUR", 1234567.89},
	"1.234.567,89 $":   CurrencyAmount{"USD", 1234567.89},
	"$ 1.234.567,89":   CurrencyAmount{"USD", 1234567.89},

	"EUR   1.234.567,89": CurrencyAmount{"EUR", 1234567.89},
	"1.234.567,89   EUR": CurrencyAmount{"EUR", 1234567.89},
	"1.234.567,89   $":   CurrencyAmount{"USD", 1234567.89},
	"$   1.234.567,89":   CurrencyAmount{"USD", 1234567.89},

	"EUR1 234 567,89": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567,89EUR": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567,89$":   CurrencyAmount{"USD", 1234567.89},
	"$1 234 567,89":   CurrencyAmount{"USD", 1234567.89},

	"EUR 1 234 567,89": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567,89 EUR": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567,89 $":   CurrencyAmount{"USD", 1234567.89},
	"$ 1 234 567,89":   CurrencyAmount{"USD", 1234567.89},

	"EUR   1 234 567,89": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567,89   EUR": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567,89   $":   CurrencyAmount{"USD", 1234567.89},
	"$   1 234 567,89":   CurrencyAmount{"USD", 1234567.89},

	"EUR 1 234 567.89": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567.89 EUR": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567.89 $":   CurrencyAmount{"USD", 1234567.89},
	"$ 1 234 567.89":   CurrencyAmount{"USD", 1234567.89},

	"EUR   1 234 567.89": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567.89   EUR": CurrencyAmount{"EUR", 1234567.89},
	"1 234 567.89   $":   CurrencyAmount{"USD", 1234567.89},
	"$   1 234 567.89":   CurrencyAmount{"USD", 1234567.89},

	"EUR 1'234'567,89": CurrencyAmount{"EUR", 1234567.89},
	"1'234'567,89 EUR": CurrencyAmount{"EUR", 1234567.89},
	"1'234'567,89 $":   CurrencyAmount{"USD", 1234567.89},
	"$ 1'234'567,89":   CurrencyAmount{"USD", 1234567.89},

	"EUR   1'234'567,89": CurrencyAmount{"EUR", 1234567.89},
	"1'234'567,89   EUR": CurrencyAmount{"EUR", 1234567.89},
	"1'234'567,89   $":   CurrencyAmount{"USD", 1234567.89},
	"$   1'234'567,89":   CurrencyAmount{"USD", 1234567.89},
}

func TestParseCurrencyAmount(t *testing.T) {
	// Accept integers
	for str, expected := range currencyIntAmountTable {
		result, err := ParseCurrencyAmount(str)
		assert.NoError(t, err, "ParseCurrencyAmount(%#v)", str)
		assert.Equal(t, expected, result, "ParseCurrencyAmount(%#v)", str)
	}

	for str, expected := range currencyAmountTable {
		result, err := ParseCurrencyAmount(str)
		assert.NoError(t, err, "ParseCurrencyAmount(%#v)", str)
		assert.Equal(t, expected, result, "ParseCurrencyAmount(%#v)", str)
	}

	// Don't accept integers
	for str := range currencyIntAmountTable {
		_, err := ParseCurrencyAmount(str, 2)
		assert.Error(t, err, "ParseCurrencyAmount(%#v, %#v)", str, 2)
	}

	for str, expected := range currencyAmountTable {
		result, err := ParseCurrencyAmount(str, 2)
		assert.NoError(t, err, "ParseCurrencyAmount(%#v, %#v)", str, 2)
		assert.Equal(t, expected, result, "ParseCurrencyAmount(%#v, %#v)", str, 2)
	}
}
