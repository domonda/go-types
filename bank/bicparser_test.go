package bank

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBICParser_Parse(t *testing.T) {
	// Unlike BIC.Normalized, the parser must return a valid BIC unchanged
	// and must not append "XXX" to 8 character BICs.
	for _, str := range []string{"BELADEBE", "BELADEBEXXX", "GENODEF1JEV"} {
		normalized, err := BICParser{}.Parse(str)
		require.NoError(t, err, str)
		assert.Equal(t, str, normalized)
	}

	// Spaces are not removed, so a BIC with spaces is invalid
	_, err := BICParser{}.Parse("BKAU AT WW")
	assert.Error(t, err)
	_, err = BICParser{}.Parse("DEUTZZFF")
	assert.Error(t, err, "invalid country code")
	_, err = BICParser{}.Parse("")
	assert.Error(t, err)
}
