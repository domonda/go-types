package vat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullID(t *testing.T) {
	var n NullID
	assert.True(t, n.Valid(), "empty NullID is valid")
	assert.NoError(t, n.Validate(), "empty NullID is valid")

	assert.Empty(t, n.NormalizedOrEmpty())
}
