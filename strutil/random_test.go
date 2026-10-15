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

func TestRandomString_URLSafe(t *testing.T) {
	// The string is documented as URL compatible, so it must only
	// contain characters of the unpadded base64 URL alphabet.
	for range 100 {
		str := RandomString(64)
		for _, r := range str {
			ok := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_'
			if !assert.True(t, ok, "rune %q in %q is not URL safe", r, str) {
				return
			}
		}
	}
}

func TestRandomString_Unique(t *testing.T) {
	// Used for tokens, so repeated calls must not return the same value
	seen := make(map[string]bool)
	for range 1000 {
		str := RandomString(16)
		assert.False(t, seen[str], "duplicate random string %q", str)
		seen[str] = true
	}
}

func TestRandomStringBytes_Length(t *testing.T) {
	assert.Empty(t, RandomStringBytes(0))
	assert.Equal(t, "", RandomString(0))
	assert.Len(t, RandomStringBytes(33), 33)
	// A negative length is a programming error
	assert.Panics(t, func() { RandomStringBytes(-1) })
	assert.Panics(t, func() { RandomString(-1) })
}
