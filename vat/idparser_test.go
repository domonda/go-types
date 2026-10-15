package vat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/language"
	"github.com/domonda/go-types/strfmt"
)

// IDParser is registered as a strfmt.Parser, so it has to keep
// implementing that interface.
var _ strfmt.Parser = IDParser{}

func TestIDParser_Parse(t *testing.T) {
	t.Run("valid input is normalized", func(t *testing.T) {
		got, err := IDParser{}.Parse("at u 1022 3006")
		require.NoError(t, err)
		assert.Equal(t, "ATU10223006", got)
	})
	t.Run("invalid input wraps ErrInvalidID", func(t *testing.T) {
		_, err := IDParser{}.Parse("ATU12345678")
		assert.ErrorIs(t, err, ErrInvalidID)

		_, err = IDParser{}.Parse("")
		assert.ErrorIs(t, err, ErrInvalidID)
	})
	t.Run("language hints do not change the result", func(t *testing.T) {
		// VAT IDs are language independent: the country prefix
		// alone decides the format, not the language of the document.
		got, err := IDParser{}.Parse("de 167 015 661", language.EN, language.FR)
		require.NoError(t, err)
		assert.Equal(t, "DE167015661", got)
	})
}
