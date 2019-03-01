package vat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullableID(t *testing.T) {
	var n NullableID
	assert.True(t, n.Valid(), "empty NullableID is valid")
	assert.NoError(t, n.Validate(), "empty NullableID is valid")

	assert.Empty(t, n.NormalizedOrNull())
}
