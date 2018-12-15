package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	for length := 1; length <= 100; length++ {
		str := RandomString(length)
		assert.Len(t, str, length, "expected RandomString length")
	}
}
