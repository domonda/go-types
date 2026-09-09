package money

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrencyDecimalAmount_String(t *testing.T) {
	ca := CurrencyDecimalAmountEUR(NewDecimalAmount(123456, 2))
	assert.Equal(t, "EUR 1234.56", ca.String())

	// The amount's own scale is preserved, unlike CurrencyAmount which forces 2.
	assert.Equal(t, "EUR 1234.5678", CurrencyDecimalAmountEUR(NewDecimalAmount(12345678, 4)).String())
	assert.Equal(t, "JPY 1000", CurrencyDecimalAmountJPY(NewDecimalAmount(1000, 0)).String())

	// Empty currency omits the code.
	assert.Equal(t, "1234.56", NewCurrencyDecimalAmount("", NewDecimalAmount(123456, 2)).String())
}

func TestCurrencyDecimalAmount_constructors(t *testing.T) {
	amount := NewDecimalAmount(100, 2)
	assert.Equal(t, Currency("USD"), CurrencyDecimalAmountUSD(amount).Currency)
	assert.Equal(t, Currency("EUR"), CurrencyDecimalAmountEUR(amount).Currency)
	assert.Equal(t, Currency("CHF"), CurrencyDecimalAmountCHF(amount).Currency)
	assert.Equal(t, Currency("GBP"), CurrencyDecimalAmountGBP(amount).Currency)
	assert.Equal(t, Currency("JPY"), CurrencyDecimalAmountJPY(amount).Currency)
	assert.Equal(t, amount, NewCurrencyDecimalAmount(EUR, amount).Amount)
}

func TestCurrencyDecimalAmount_Format(t *testing.T) {
	ca := CurrencyDecimalAmountEUR(NewDecimalAmount(123456789, 2)) // 1234567.89
	assert.Equal(t, "EUR 1,234,567.89", ca.FormatSep(true, ',', '.'))
	assert.Equal(t, "1.234.567,89 EUR", ca.FormatSep(false, '.', ','))
}

func TestParseCurrencyDecimalAmount(t *testing.T) {
	cases := []struct {
		str      string
		currency Currency
		amount   string
	}{
		{"EUR 1234.56", EUR, "1234.56"},
		{"1234.56 EUR", EUR, "1234.56"},
		{"USD 1,234.56", USD, "1234.56"},
		{"1.234,56 EUR", EUR, "1234.56"},
		{"CHF 1'234.56", CHF, "1234.56"},
		{"-99.99 USD", USD, "-99.99"},
		// Exactness preserved through the currency+amount parse.
		{"EUR 99999999999999.99", EUR, "99999999999999.99"},
	}
	for _, c := range cases {
		ca, err := ParseCurrencyDecimalAmount(c.str)
		require.NoError(t, err, "ParseCurrencyDecimalAmount(%q)", c.str)
		assert.Equal(t, c.currency, ca.Currency, "currency of %q", c.str)
		assert.Equal(t, c.amount, ca.Amount.String(), "amount of %q", c.str)
	}
}

func TestParseCurrencyDecimalAmount_acceptedDecimals(t *testing.T) {
	_, err := ParseCurrencyDecimalAmount("EUR 1.234", 0, 2)
	assert.Error(t, err)

	ca, err := ParseCurrencyDecimalAmount("EUR 1.23", 0, 2)
	require.NoError(t, err)
	assert.Equal(t, "1.23", ca.Amount.String())
}

func TestCurrencyDecimalAmount_GoString(t *testing.T) {
	ca := CurrencyDecimalAmountEUR(NewDecimalAmount(123456, 2))
	assert.Equal(t, "{Currency: \"EUR\", Amount: money.NewDecimalAmount(123456, 2)}", ca.GoString())
	assert.Equal(t, "{Currency: \"EUR\", Amount: money.NewDecimalAmount(123456, 2)}", fmt.Sprintf("%#v", ca))
}

func TestCurrencyDecimalAmount_SQL(t *testing.T) {
	ca := CurrencyDecimalAmountEUR(NewDecimalAmount(123456, 2))
	v, err := ca.Value()
	require.NoError(t, err)
	assert.Equal(t, "EUR 1234.56", v)
	assert.IsType(t, driver.Value(""), v)

	var scanned CurrencyDecimalAmount
	require.NoError(t, scanned.Scan("EUR 1234.56"))
	assert.Equal(t, ca, scanned)

	require.NoError(t, scanned.Scan([]byte("USD 1.234,56")))
	assert.Equal(t, Currency("USD"), scanned.Currency)
	assert.Equal(t, "1234.56", scanned.Amount.String())

	// Bare float64 from the driver has no currency.
	require.NoError(t, scanned.Scan(float64(12.5)))
	assert.Equal(t, Currency(""), scanned.Currency)
	assert.Equal(t, "12.5", scanned.Amount.String())

	assert.Error(t, scanned.Scan(true))
}

func TestCurrencyDecimalAmount_JSON(t *testing.T) {
	ca := CurrencyDecimalAmountEUR(NewDecimalAmount(99999, 2))
	data, err := json.Marshal(ca)
	require.NoError(t, err)
	assert.JSONEq(t, `{"Currency":"EUR","Amount":999.99}`, string(data))

	var got CurrencyDecimalAmount
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, ca, got)
}

func TestCurrencyDecimalAmount_CurrencyAmount(t *testing.T) {
	ca := CurrencyDecimalAmountEUR(NewDecimalAmount(123456, 2))
	fa := ca.CurrencyAmount()
	assert.Equal(t, Currency("EUR"), fa.Currency)
	assert.InDelta(t, 1234.56, float64(fa.Amount), 1e-9)
}

func TestCurrencyDecimalAmount_ScanString(t *testing.T) {
	var ca CurrencyDecimalAmount
	require.NoError(t, ca.ScanString("EUR 1.234,56", true))
	assert.Equal(t, Currency("EUR"), ca.Currency)
	assert.Equal(t, "1234.56", ca.Amount.String())
}
