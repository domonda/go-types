package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCurrencyParser_Parse pins that CurrencyParser normalizes the
// currency notations found on invoices to the ISO 4217 code for strfmt,
// and reports unknown currencies as an error.
func TestCurrencyParser_Parse(t *testing.T) {
	for str, want := range map[string]string{
		"EUR":            "EUR",
		" eur ":          "EUR",
		"€":              "EUR",
		"Euro":           "EUR",
		"$":              "USD",
		"US$":            "USD",
		"£":              "GBP",
		"Ft":             "HUF",
		"sek":            "SEK",
		"Swiss Franc":    "CHF",
		"Pound Sterling": "GBP",
	} {
		got, err := CurrencyParser{}.Parse(str)
		assert.NoError(t, err, "Parse(%q)", str)
		assert.Equal(t, want, got, "Parse(%q)", str)
	}

	for _, str := range []string{"", " ", "XYZ", "BTC", "12.34"} {
		_, err := CurrencyParser{}.Parse(str)
		assert.Error(t, err, "Parse(%q)", str)
	}
}
