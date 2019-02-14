package date

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullDate(t *testing.T) {
	var n NullDate
	assert.True(t, n.Valid(), "empty NullDate is valid")
	assert.NoError(t, n.Validate(), "empty NullDate is valid")

	n.Date = "0001-01-01"
	assert.True(t, n.Valid(), "empty NullDate is valid")
	assert.NoError(t, n.Validate(), "empty NullDate is valid")

	assert.Empty(t, n.NormalizedOrEmpty())
}
