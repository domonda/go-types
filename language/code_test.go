package language

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeNormalized(t *testing.T) {
	c := Code("")
	n, err := c.Normalized()
	assert.Empty(t, n)
	assert.Error(t, err)
}
