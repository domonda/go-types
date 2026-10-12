package strutil

import (
	"crypto/rand"
	"encoding/base64"
)

// RandomString returns an URL compatible random string
// with the requested length.
func RandomString(length int) string {
	return string(RandomStringBytes(length))
}

// RandomStringBytes returns an URL compatible random string
// with the requested length as []byte slice,
// saving a string copy compared to RandomString.
// It panics if length is negative.
func RandomStringBytes(length int) []byte {
	if length < 0 {
		panic("invalid length for RandomStringBytes")
	}
	randomBytesLen := (length*6 + 7) / 8
	randomBytes := make([]byte, randomBytesLen)
	_, _ = rand.Read(randomBytes) //#nosec G104 -- crypto/rand.Read never returns an error since Go 1.24
	encodedLen := base64.RawURLEncoding.EncodedLen(randomBytesLen)
	encoded := make([]byte, encodedLen)
	base64.RawURLEncoding.Encode(encoded, randomBytes)
	return encoded[:length]
}
