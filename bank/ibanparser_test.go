package bank

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIBANParser_Parse(t *testing.T) {
	// The parser returns the normalized IBAN without spaces
	normalized, err := IBANParser{}.Parse("DE89 3704 0044 0532 0130 00")
	require.NoError(t, err)
	assert.Equal(t, "DE89370400440532013000", normalized)

	_, err = IBANParser{}.Parse("DE89 3704 0044 0532 0130 01")
	assert.Error(t, err, "invalid check sum")
	_, err = IBANParser{}.Parse("")
	assert.Error(t, err)
}
