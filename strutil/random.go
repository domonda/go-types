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
func RandomStringBytes(length int) []byte {
	if length < 0 {
		panic("invalid length for RandomStringBytes")
	}
	numRandomBytes := (length*6 + 7) / 8
	encodedLen := base64.RawURLEncoding.EncodedLen(numRandomBytes)
	randomBytes := make([]byte, numRandomBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	result := make([]byte, encodedLen)
	base64.RawURLEncoding.Encode(result, randomBytes)
	return result[:length]
}
