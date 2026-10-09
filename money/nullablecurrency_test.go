package money

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNullableCurrency_Null pins the NULL state handling: the empty string
// is NULL, Set/SetNull switch between the states, and Get refuses to
// return a NULL currency instead of silently returning an empty Currency.
func TestNullableCurrency_Null(t *testing.T) {
	var n NullableCurrency
	assert.True(t, n.IsNull())
	assert.False(t, n.IsNotNull())
	assert.Panics(t, func() { n.Get() }, "Get must panic on NULL")

	n.Set(EUR)
	assert.True(t, n.IsNotNull())
	assert.False(t, n.IsNull())
	assert.Equal(t, EUR, n.Get())

	n.SetNull()
	assert.True(t, n.IsNull())
	assert.Equal(t, NullableCurrency(CurrencyNull), n)
}

// TestNullableCurrency_GetOrStringOr pins the fallbacks for NULL values:
// GetOr also falls back for invalid currencies so that callers always
// get a usable Currency, while StringOr only falls back for NULL.
func TestNullableCurrency_GetOrStringOr(t *testing.T) {
	assert.Equal(t, EUR, NullableCurrency("").GetOr(EUR))
	assert.Equal(t, EUR, NullableCurrency("XYZ").GetOr(EUR))
	assert.Equal(t, USD, NullableCurrency("USD").GetOr(EUR))

	assert.Equal(t, "none", NullableCurrency("").StringOr("none"))
	assert.Equal(t, "USD", NullableCurrency("USD").StringOr("none"))
	assert.Equal(t, "XYZ", NullableCurrency("XYZ").StringOr("none"))
}

// TestNullableCurrency_Validity pins the main difference to Currency:
// NULL is a valid NullableCurrency, but not a valid not-null currency.
func TestNullableCurrency_Validity(t *testing.T) {
	tests := []struct {
		n                  NullableCurrency
		valid              bool
		validAndNotNull    bool
		validAndNormalized bool
	}{
		{n: "", valid: true, validAndNotNull: false, validAndNormalized: true},
		{n: "EUR", valid: true, validAndNotNull: true, validAndNormalized: true},
		{n: "eur", valid: true, validAndNotNull: true, validAndNormalized: false},
		{n: "€", valid: true, validAndNotNull: true, validAndNormalized: false},
		{n: "XYZ", valid: false, validAndNotNull: false, validAndNormalized: false},
		{n: " ", valid: false, validAndNotNull: false, validAndNormalized: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", string(tt.n)), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.n.Valid(), "Valid")
			assert.Equal(t, tt.valid, tt.n.Validate() == nil, "Validate")
			assert.Equal(t, tt.valid, tt.n.ValidPtr(), "ValidPtr")
			assert.Equal(t, tt.validAndNotNull, tt.n.ValidAndNotNull(), "ValidAndNotNull")
			assert.Equal(t, tt.validAndNormalized, tt.n.ValidAndNormalized(), "ValidAndNormalized")
		})
	}

	var nilPtr *NullableCurrency
	assert.True(t, nilPtr.ValidPtr(), "a nil pointer is a valid NULL")
}

// TestNullableCurrency_Normalized pins that NULL normalizes to NULL
// without error, and that NormalizedOrNull degrades unknown currencies
// to NULL instead of keeping them in their invalid form.
func TestNullableCurrency_Normalized(t *testing.T) {
	tests := []struct {
		n       NullableCurrency
		want    NullableCurrency
		wantErr bool
	}{
		{n: "", want: ""},
		{n: "EUR", want: "EUR"},
		{n: " eur ", want: "EUR"},
		{n: "£", want: "GBP"},
		{n: "XYZ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", string(tt.n)), func(t *testing.T) {
			got, err := tt.n.Normalized()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, NullableCurrency(CurrencyNull), tt.n.NormalizedOrNull())
				assert.Equal(t, string(tt.n), tt.n.String(), "String must return an unnormalizable value unchanged")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tt.n.NormalizedOrNull())
			assert.Equal(t, string(tt.want), tt.n.String(), "String must normalize")
		})
	}
}

// TestNullableCurrency_ScanString pins that the textual NULL markers found
// in CSV and other text sources scan as NULL instead of as an invalid
// currency, that valid input is normalized, and that invalid input
// is only rejected when validating.
func TestNullableCurrency_ScanString(t *testing.T) {
	tests := []struct {
		source   string
		validate bool
		want     NullableCurrency
		wantErr  bool
	}{
		{source: "", validate: true, want: CurrencyNull},
		{source: "NULL", validate: true, want: CurrencyNull},
		{source: "null", validate: true, want: CurrencyNull},
		{source: "nil", validate: true, want: CurrencyNull},
		{source: " eur ", validate: true, want: "EUR"},
		{source: "€", validate: true, want: "EUR"},
		{source: "XYZ", validate: true, wantErr: true},
		{source: "XYZ", validate: false, want: "XYZ"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q validate=%t", tt.source, tt.validate), func(t *testing.T) {
			n := NullableCurrency("CHF")
			err := n.ScanString(tt.source, tt.validate)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, NullableCurrency("CHF"), n, "a failed scan must not modify the currency")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, n)
		})
	}
}

// TestNullableCurrency_SQL pins that NULL is written to SQL as nil
// instead of an empty string, and that SQL NULL and both driver
// text representations scan back to the same value.
func TestNullableCurrency_SQL(t *testing.T) {
	v, err := NullableCurrency(CurrencyNull).Value()
	require.NoError(t, err)
	assert.Nil(t, v)

	v, err = NullableCurrency("EUR").Value()
	require.NoError(t, err)
	assert.Equal(t, "EUR", v)

	for _, n := range []NullableCurrency{CurrencyNull, "EUR", "USD"} {
		v, err := n.Value()
		require.NoError(t, err)
		var got NullableCurrency = "CHF"
		require.NoError(t, got.Scan(v))
		assert.Equal(t, n, got, "SQL round-trip of %q", string(n))
	}

	var n NullableCurrency
	src := []byte("GBP")
	require.NoError(t, n.Scan(src))
	src[0] = 'X' // drivers may reuse the buffer after Scan returns
	assert.Equal(t, NullableCurrency("GBP"), n)

	assert.Error(t, n.Scan(int64(1)))
	assert.Equal(t, NullableCurrency("GBP"), n, "a failed scan must not modify the currency")
}

// TestNullableCurrency_JSON pins that NULL is marshalled as JSON null
// instead of "", and that both forms decode back into a zero value.
func TestNullableCurrency_JSON(t *testing.T) {
	type doc struct {
		Currency NullableCurrency `json:"currency"`
	}
	tests := []struct {
		n    NullableCurrency
		json string
	}{
		{n: CurrencyNull, json: `{"currency":null}`},
		{n: "EUR", json: `{"currency":"EUR"}`},
	}
	for _, tt := range tests {
		data, err := json.Marshal(doc{Currency: tt.n})
		require.NoError(t, err)
		assert.Equal(t, tt.json, string(data))

		var got doc
		require.NoError(t, json.Unmarshal(data, &got))
		assert.Equal(t, tt.n, got.Currency, "round-trip of %s", data)
	}
}

// TestNullableCurrency_Display pins the display helpers, which must not
// panic on NULL like Get does, and the raw Currency conversion.
func TestNullableCurrency_Display(t *testing.T) {
	assert.Equal(t, "€", NullableCurrency("EUR").Symbol())
	assert.Equal(t, "SEK", NullableCurrency("SEK").Symbol(), "falls back to the code")
	assert.Equal(t, "", NullableCurrency(CurrencyNull).Symbol())

	assert.Equal(t, "United States Dollar", NullableCurrency("USD").EnglishName())
	assert.Equal(t, "", NullableCurrency(CurrencyNull).EnglishName())

	assert.Equal(t, CHF, NullableCurrency("CHF").Currency())
	assert.Equal(t, Currency(CurrencyNull), NullableCurrency(CurrencyNull).Currency(), "Currency must not panic on NULL")
}
