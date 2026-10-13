package money

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestCurrencyAmount_Constructors pins that every convenience constructor
// assigns its own currency, a copy-paste error here would silently
// book amounts in the wrong currency.
func TestCurrencyAmount_Constructors(t *testing.T) {
	assert.Equal(t, CurrencyAmount{Currency: SEK, Amount: 1.5}, NewCurrencyAmount(SEK, 1.5))
	assert.Equal(t, CurrencyAmount{Currency: USD, Amount: 1.5}, CurrencyAmountUSD(1.5))
	assert.Equal(t, CurrencyAmount{Currency: EUR, Amount: 1.5}, CurrencyAmountEUR(1.5))
	assert.Equal(t, CurrencyAmount{Currency: CHF, Amount: 1.5}, CurrencyAmountCHF(1.5))
	assert.Equal(t, CurrencyAmount{Currency: GBP, Amount: 1.5}, CurrencyAmountGBP(1.5))
	assert.Equal(t, CurrencyAmount{Currency: JPY, Amount: 1.5}, CurrencyAmountJPY(1.5))
}

// TestCurrencyAmount_Format pins the String format that is also used
// as SQL value: currency code first, no thousands grouping, a point as
// decimal separator and the decimal places of the minor unit of the
// currency, two for an empty currency, and no currency prefix if the
// currency is empty.
func TestCurrencyAmount_Format(t *testing.T) {
	assert.Equal(t, "EUR 1234.56", CurrencyAmountEUR(1234.56).String())
	assert.Equal(t, "EUR -1234567.89", CurrencyAmountEUR(-1234567.89).String())
	assert.Equal(t, "USD 5.00", CurrencyAmountUSD(5).String())
	assert.Equal(t, "5.00", CurrencyAmount{Amount: 5}.String())

	// Two decimals lost the third decimal of a dinar and wrote a yen
	// fraction that doesn't exist: "KWD 1.23" and "JPY 1235.00" before
	assert.Equal(t, "KWD 1.234", CurrencyAmount{Currency: KWD, Amount: 1.234}.String())
	assert.Equal(t, "BHD 0.500", CurrencyAmount{Currency: BHD, Amount: 0.5}.String())
	assert.Equal(t, "KWD 1.235", CurrencyAmount{Currency: KWD, Amount: 1.2345}.String(), "rounded half away from zero")
	assert.Equal(t, "JPY 1235", CurrencyAmountJPY(1234.5).String())
	assert.Equal(t, "JPY -1000", CurrencyAmountJPY(-1000).String())
	assert.Equal(t, "KRW 50000", CurrencyAmount{Currency: KRW, Amount: 50000}.String())

	ca := CurrencyAmountEUR(1234567.891)
	assert.Equal(t, "1.234.567,89 EUR", ca.FormatSep(false, '.', ',', 2))
	assert.Equal(t, "EUR 1,234,567.89", ca.FormatSep(true, ',', '.', 2))
	assert.Equal(t, "CHF 1'234.50", CurrencyAmountCHF(1234.5).FormatSep(true, '\'', '.', 2))
	assert.Equal(t, "1234567.891 EUR", ca.FormatSep(false, 0, '.', -1))
	assert.Equal(t, "1234.5", CurrencyAmount{Amount: 1234.5}.FormatSep(true, 0, '.', 1))
}

// TestCurrencyAmount_FormatParseRoundTrip pins that every combination of
// currency position and separator convention produced by FormatSep is
// parsed back by ParseCurrencyAmount to the identical value, so amounts
// formatted for one locale can be read again.
func TestCurrencyAmount_FormatParseRoundTrip(t *testing.T) {
	formats := []struct {
		currencyFirst bool
		thousandsSep  rune
		decimalSep    rune
	}{
		{currencyFirst: true, thousandsSep: 0, decimalSep: '.'},
		{currencyFirst: false, thousandsSep: 0, decimalSep: ','},
		{currencyFirst: true, thousandsSep: ',', decimalSep: '.'},  // English
		{currencyFirst: false, thousandsSep: '.', decimalSep: ','}, // German
		{currencyFirst: false, thousandsSep: ' ', decimalSep: ','}, // French
		{currencyFirst: true, thousandsSep: '\'', decimalSep: '.'}, // Swiss
	}
	amounts := []CurrencyAmount{
		CurrencyAmountEUR(0),
		CurrencyAmountEUR(0.01),
		CurrencyAmountUSD(-0.99),
		CurrencyAmountGBP(12.34),
		CurrencyAmountCHF(1234.56),
		CurrencyAmountEUR(-1234567.89),
		CurrencyAmountUSD(999999999.99),
		{Amount: 1234.56}, // no currency
	}
	for _, f := range formats {
		for _, ca := range amounts {
			str := ca.FormatSep(f.currencyFirst, f.thousandsSep, f.decimalSep, 2)
			parsed, err := ParseCurrencyAmount(str, 2)
			if assert.NoError(t, err, "ParseCurrencyAmount(%q)", str) {
				assert.Equal(t, ca, parsed, "ParseCurrencyAmount(%q)", str)
			}
		}
	}
}

// TestParseCurrencyAmount_currencyWord pins that a currency written as
// one word before or after the amount is normalized like NormalizeCurrency
// does it, so "Franken" on a Swiss receipt is CHF. ParseCurrencyAmount cuts
// the currency at the first or last separator, so a qualified name like
// "Canadian Dollar" can't be passed as a whole and must not turn into USD
// through its last word.
func TestParseCurrencyAmount_currencyWord(t *testing.T) {
	for str, want := range map[string]CurrencyAmount{
		"12,50 Franken": {Currency: CHF, Amount: 12.5},
		"Dollar 12.00":  {Currency: USD, Amount: 12},
		"12.00 Euro":    {Currency: EUR, Amount: 12},
		"Pounds 3.50":   {Currency: GBP, Amount: 3.5},
	} {
		got, err := ParseCurrencyAmount(str)
		if assert.NoError(t, err, "ParseCurrencyAmount(%q)", str) {
			assert.Equal(t, want, got, "ParseCurrencyAmount(%q)", str)
		}
	}
	for _, str := range []string{"12.00 Canadian Dollar", "Canadian Dollar 12.00", "12.00 Burundi-Franken"} {
		got, err := ParseCurrencyAmount(str)
		assert.Error(t, err, "ParseCurrencyAmount(%q) = %v", str, got)
	}
}

// TestParseCurrencyAmount_invalid pins that an unknown currency or a
// malformed amount is reported as an error instead of being silently
// dropped and returning just the amount.
func TestParseCurrencyAmount_invalid(t *testing.T) {
	for _, str := range []string{
		"",
		"EUR",
		"XYZ 12.34",
		"12.34 XYZ",
		"EUR 12.34 USD",
		"EUR 1,23,45",
	} {
		got, err := ParseCurrencyAmount(str)
		assert.Error(t, err, "ParseCurrencyAmount(%q)", str)
		assert.Equal(t, CurrencyAmount{}, got, "ParseCurrencyAmount(%q) must return the zero value on error", str)
	}
}

// TestCurrencyAmount_GoString pins that the debug representation shows
// the exact binary amount value via Amount.GoString.
func TestCurrencyAmount_GoString(t *testing.T) {
	ca := CurrencyAmountEUR(1234.5)
	want := `{Currency: "EUR", Amount: money.Amount(1234.5)}`
	assert.Equal(t, want, ca.GoString())
	assert.Equal(t, want, fmt.Sprintf("%#v", ca))
}

// TestCurrencyAmount_ScanString pins that ScanString accepts the
// currency at either side in any separator convention, but only whole
// or cent amounts or the minor units of the currency, like the three
// decimals of KWD that String writes, and that validate rejects NaN and ±Inf.
func TestCurrencyAmount_ScanString(t *testing.T) {
	tests := []struct {
		source   string
		validate bool
		want     CurrencyAmount
		wantErr  bool
	}{
		{source: "EUR 1.234,56", validate: true, want: CurrencyAmountEUR(1234.56)},
		{source: "1,234.56 $", validate: true, want: CurrencyAmountUSD(1234.56)},
		{source: "€ 12", validate: true, want: CurrencyAmountEUR(12)},
		{source: "12,34", validate: true, want: CurrencyAmount{Amount: 12.34}},
		{source: "EUR NaN", validate: false, want: CurrencyAmountEUR(Amount(math.NaN()))},
		{source: "EUR 12.345", validate: false, wantErr: true}, // sub-cent amount
		{source: "EUR 12.3", validate: false, wantErr: true},
		{source: "KWD 12.345", validate: true, want: CurrencyAmount{Currency: KWD, Amount: 12.345}},
		{source: "12.345 BHD", validate: true, want: CurrencyAmount{Currency: BHD, Amount: 12.345}},
		{source: "KWD 12.34", validate: true, want: CurrencyAmount{Currency: KWD, Amount: 12.34}},
		{source: "KWD 12.3456", validate: true, wantErr: true},
		{source: "JPY 1000", validate: true, want: CurrencyAmountJPY(1000)},
		{source: "JPY 1.000", validate: true, wantErr: true}, // ambiguous grouping
		{source: "EUR NaN", validate: true, wantErr: true},
		{source: "EUR -Inf", validate: true, wantErr: true},
		{source: "XYZ 12.34", validate: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q validate=%t", tt.source, tt.validate), func(t *testing.T) {
			ca := CurrencyAmountCHF(666)
			err := ca.ScanString(tt.source, tt.validate)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, CurrencyAmountCHF(666), ca, "a failed scan must not modify the value")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want.Currency, ca.Currency)
			assert.True(t, equalInclNaN(float64(tt.want.Amount), float64(ca.Amount)), "got %v, want %v", ca.Amount, tt.want.Amount)
		})
	}
}

// TestCurrencyAmount_SQL pins that the SQL value is the String format,
// that it scans back to the identical CurrencyAmount, and that a bare
// float64 from a numeric column is accepted without currency.
// CurrencyAmount is not nullable, so SQL NULL must be an error.
func TestCurrencyAmount_SQL(t *testing.T) {
	for _, ca := range []CurrencyAmount{
		CurrencyAmountEUR(1234.56),
		CurrencyAmountUSD(-0.99),
		CurrencyAmountJPY(1000),
		{Currency: KWD, Amount: 1.234},
		{Amount: 12.34},
		{},
	} {
		v, err := ca.Value()
		require.NoError(t, err)
		assert.Equal(t, ca.String(), v)

		// String and ScanString round-trip for every minor unit
		var scanned CurrencyAmount
		if assert.NoError(t, scanned.ScanString(v.(string), true), "ScanString(%q)", v) {
			assert.Equal(t, ca, scanned, "ScanString(%q)", v)
		}

		var fromString, fromBytes CurrencyAmount
		require.NoError(t, fromString.Scan(v))
		assert.Equal(t, ca, fromString, "SQL round-trip of %q", v)
		require.NoError(t, fromBytes.Scan([]byte(v.(string))))
		assert.Equal(t, ca, fromBytes, "SQL round-trip of %q as []byte", v)
	}

	var ca CurrencyAmount
	require.NoError(t, ca.Scan(float64(12.5)))
	assert.Equal(t, CurrencyAmount{Amount: 12.5}, ca)

	ca = CurrencyAmountEUR(1)
	assert.Error(t, ca.Scan(nil), "CurrencyAmount can't represent SQL NULL")
	assert.Error(t, ca.Scan(int64(1)))
	assert.Error(t, ca.Scan("EUR abc"))
	assert.Equal(t, CurrencyAmountEUR(1), ca, "a failed scan must not modify the value")
}

// Value, String and JSON write the currency of a CurrencyAmount normalized
// like Currency.Value, because ParseCurrencyAmount, Scan and JSON decoding
// keep a currency like "eur" as given: without normalization the same
// amount would reach SQL and JSON in different spellings. A currency that
// can't be normalized is written unchanged.
func TestCurrencyAmount_NormalizedCurrencyOutput(t *testing.T) {
	tests := []struct {
		ca       CurrencyAmount
		wantSQL  string
		wantJSON string
	}{
		{ca: CurrencyAmount{Currency: "eur", Amount: 1.5}, wantSQL: "EUR 1.50", wantJSON: `{"Currency":"EUR","Amount":1.5}`},
		{ca: CurrencyAmount{Currency: "€", Amount: 1.5}, wantSQL: "EUR 1.50", wantJSON: `{"Currency":"EUR","Amount":1.5}`},
		{ca: CurrencyAmount{Currency: "jpy", Amount: 1234.5}, wantSQL: "JPY 1235", wantJSON: `{"Currency":"JPY","Amount":1234.5}`},
		{ca: CurrencyAmount{Currency: "XYZ", Amount: 1.5}, wantSQL: "XYZ 1.50", wantJSON: `{"Currency":"XYZ","Amount":1.5}`},
	}
	for _, tt := range tests {
		v, err := tt.ca.Value()
		require.NoError(t, err)
		assert.Equal(t, tt.wantSQL, v, "Value of %#v", tt.ca)
		assert.Equal(t, tt.wantSQL, tt.ca.String(), "String of %#v", tt.ca)

		j, err := json.Marshal(tt.ca)
		require.NoError(t, err)
		assert.JSONEq(t, tt.wantJSON, string(j), "json.Marshal of %#v", tt.ca)
	}
}
