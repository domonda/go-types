package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestJoinCurrencyAmount pins the one text format that the money currency
// amount types and the strfmt money formats share, so that their outputs
// can't drift apart: ParseCurrencyAmount reads the currency from either end,
// and an amount without currency has no dangling space that would make
// "1.50" and "1.50 " different cells.
func TestJoinCurrencyAmount(t *testing.T) {
	assert.Equal(t, "EUR 1.50", JoinCurrencyAmount("EUR", "1.50", true))
	assert.Equal(t, "1,50 EUR", JoinCurrencyAmount("EUR", "1,50", false))
	assert.Equal(t, "1.50", JoinCurrencyAmount("", "1.50", true))
	assert.Equal(t, "1.50", JoinCurrencyAmount("", "1.50", false))
}
