package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var currencyIntAmountTable = map[string]CurrencyAmount{
	"1": {"", 1},

	"EUR1": {"EUR", 1},
	"1EUR": {"EUR", 1},
	"1$":   {"USD", 1},
	"$1":   {"USD", 1},

	"EUR 1": {"EUR", 1},
	"1 EUR": {"EUR", 1},
	"1 $":   {"USD", 1},
	"$ 1":   {"USD", 1},

	"EUR   1": {"EUR", 1},
	"1   EUR": {"EUR", 1},
	"1   $":   {"USD", 1},
	"$   1":   {"USD", 1},
}

var currencyAmountTable = map[string]CurrencyAmount{
	"EUR -12,34": {"EUR", -12.34},
	"EUR +12,34": {"EUR", 12.34},
	"EUR-12,34":  {"EUR", -12.34},
	"EUR+12,34":  {"EUR", 12.34},
	"EUR -12.34": {"EUR", -12.34},
	"EUR +12.34": {"EUR", 12.34},
	"EUR-12.34":  {"EUR", -12.34},
	"EUR+12.34":  {"EUR", 12.34},

	"EUR1.23": {"EUR", 1.23},
	"1.23EUR": {"EUR", 1.23},
	"1.23$":   {"USD", 1.23},
	"$1.23":   {"USD", 1.23},

	"EUR 1.23": {"EUR", 1.23},
	"1.23 EUR": {"EUR", 1.23},
	"1.23 $":   {"USD", 1.23},
	"$ 1.23":   {"USD", 1.23},

	"EUR   1.23": {"EUR", 1.23},
	"1.23   EUR": {"EUR", 1.23},
	"1.23   $":   {"USD", 1.23},
	"$   1.23":   {"USD", 1.23},

	"EUR1,234,567.89": {"EUR", 1234567.89},
	"1,234,567.89EUR": {"EUR", 1234567.89},
	"1,234,567.89$":   {"USD", 1234567.89},
	"$1,234,567.89":   {"USD", 1234567.89},

	"EUR 1,234,567.89": {"EUR", 1234567.89},
	"1,234,567.89 EUR": {"EUR", 1234567.89},
	"1,234,567.89 $":   {"USD", 1234567.89},
	"$ 1,234,567.89":   {"USD", 1234567.89},

	"EUR   1,234,567.89": {"EUR", 1234567.89},
	"1,234,567.89   EUR": {"EUR", 1234567.89},
	"1,234,567.89   $":   {"USD", 1234567.89},
	"$   1,234,567.89":   {"USD", 1234567.89},

	"EUR 1,23": {"EUR", 1.23},
	"1,23 EUR": {"EUR", 1.23},
	"1,23 $":   {"USD", 1.23},
	"$ 1,23":   {"USD", 1.23},

	"EUR   1,23": {"EUR", 1.23},
	"1,23   EUR": {"EUR", 1.23},
	"1,23   $":   {"USD", 1.23},
	"$   1,23":   {"USD", 1.23},

	"EUR1.234.567,89": {"EUR", 1234567.89},
	"1.234.567,89EUR": {"EUR", 1234567.89},
	"1.234.567,89$":   {"USD", 1234567.89},
	"$1.234.567,89":   {"USD", 1234567.89},

	"EUR 1.234.567,89": {"EUR", 1234567.89},
	"1.234.567,89 EUR": {"EUR", 1234567.89},
	"1.234.567,89 $":   {"USD", 1234567.89},
	"$ 1.234.567,89":   {"USD", 1234567.89},

	"EUR   1.234.567,89": {"EUR", 1234567.89},
	"1.234.567,89   EUR": {"EUR", 1234567.89},
	"1.234.567,89   $":   {"USD", 1234567.89},
	"$   1.234.567,89":   {"USD", 1234567.89},

	"EUR1 234 567,89": {"EUR", 1234567.89},
	"1 234 567,89EUR": {"EUR", 1234567.89},
	"1 234 567,89$":   {"USD", 1234567.89},
	"$1 234 567,89":   {"USD", 1234567.89},

	"EUR 1 234 567,89": {"EUR", 1234567.89},
	"1 234 567,89 EUR": {"EUR", 1234567.89},
	"1 234 567,89 $":   {"USD", 1234567.89},
	"$ 1 234 567,89":   {"USD", 1234567.89},

	"EUR   1 234 567,89": {"EUR", 1234567.89},
	"1 234 567,89   EUR": {"EUR", 1234567.89},
	"1 234 567,89   $":   {"USD", 1234567.89},
	"$   1 234 567,89":   {"USD", 1234567.89},

	"EUR 1 234 567.89": {"EUR", 1234567.89},
	"1 234 567.89 EUR": {"EUR", 1234567.89},
	"1 234 567.89 $":   {"USD", 1234567.89},
	"$ 1 234 567.89":   {"USD", 1234567.89},

	"EUR   1 234 567.89": {"EUR", 1234567.89},
	"1 234 567.89   EUR": {"EUR", 1234567.89},
	"1 234 567.89   $":   {"USD", 1234567.89},
	"$   1 234 567.89":   {"USD", 1234567.89},

	"EUR 1'234'567,89": {"EUR", 1234567.89},
	"1'234'567,89 EUR": {"EUR", 1234567.89},
	"1'234'567,89 $":   {"USD", 1234567.89},
	"$ 1'234'567,89":   {"USD", 1234567.89},

	"EUR   1'234'567,89": {"EUR", 1234567.89},
	"1'234'567,89   EUR": {"EUR", 1234567.89},
	"1'234'567,89   $":   {"USD", 1234567.89},
	"$   1'234'567,89":   {"USD", 1234567.89},
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
