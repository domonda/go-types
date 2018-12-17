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

	"EUR1.234": CurrencyAmount{"EUR", 1.234},
	"1.234EUR": CurrencyAmount{"EUR", 1.234},
	"1.234$":   CurrencyAmount{"USD", 1.234},
	"$1.234":   CurrencyAmount{"USD", 1.234},

	"EUR 1.234": CurrencyAmount{"EUR", 1.234},
	"1.234 EUR": CurrencyAmount{"EUR", 1.234},
	"1.234 $":   CurrencyAmount{"USD", 1.234},
	"$ 1.234":   CurrencyAmount{"USD", 1.234},

	"EUR   1.234": CurrencyAmount{"EUR", 1.234},
	"1.234   EUR": CurrencyAmount{"EUR", 1.234},
	"1.234   $":   CurrencyAmount{"USD", 1.234},
	"$   1.234":   CurrencyAmount{"USD", 1.234},

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

	"EUR 1,234": CurrencyAmount{"EUR", 1.234},
	"1,234 EUR": CurrencyAmount{"EUR", 1.234},
	"1,234 $":   CurrencyAmount{"USD", 1.234},
	"$ 1,234":   CurrencyAmount{"USD", 1.234},

	"EUR   1,234": CurrencyAmount{"EUR", 1.234},
	"1,234   EUR": CurrencyAmount{"EUR", 1.234},
	"1,234   $":   CurrencyAmount{"USD", 1.234},
	"$   1,234":   CurrencyAmount{"USD", 1.234},

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
		result, err := ParseCurrencyAmount(str, true)
		assert.NoError(t, err, "ParseCurrencyAmount(%#v, %#v)", str, true)
		assert.Equal(t, expected, result, "ParseCurrencyAmount(%#v, %#v)", str, true)
	}

	for str, expected := range currencyAmountTable {
		result, err := ParseCurrencyAmount(str, true)
		assert.NoError(t, err, "ParseCurrencyAmount(%#v, %#v)", str, true)
		assert.Equal(t, expected, result, "ParseCurrencyAmount(%#v, %#v)", str, true)
	}

	// Don't accept integers
	for str := range currencyIntAmountTable {
		_, err := ParseCurrencyAmount(str, false)
		assert.Error(t, err, "ParseCurrencyAmount(%#v, %#v)", str, true)
	}

	for str, expected := range currencyAmountTable {
		result, err := ParseCurrencyAmount(str, false)
		assert.NoError(t, err, "ParseCurrencyAmount(%#v, %#v)", str, true)
		assert.Equal(t, expected, result, "ParseCurrencyAmount(%#v, %#v)", str, true)
	}
}
