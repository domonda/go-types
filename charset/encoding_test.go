package charset

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UTF8Encoding(t *testing.T) {
	enc := UTF8Encoding()
	assert.Equal(t, "UTF-8", enc.Name())
	assert.Equal(t, BOMUTF8, enc.BOM())

	// Encode is a passthrough for UTF-8.
	out, err := enc.Encode([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), out)
}

func Test_UTF16_RoundTrip(t *testing.T) {
	enc := UTF16Encoding(binary.LittleEndian)
	assert.Equal(t, "UTF-16LE", enc.Name())
	assert.Equal(t, BOMUTF16LE, enc.BOM())

	original := []byte("Hello, 世界")
	encoded, err := enc.Encode(original)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded, err := enc.Decode(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func Test_UTF16_BigEndian(t *testing.T) {
	enc := UTF16Encoding(binary.BigEndian)
	assert.Equal(t, "UTF-16BE", enc.Name())
	assert.Equal(t, BOMUTF16BE, enc.BOM())
}

func Test_UTF32_RoundTrip(t *testing.T) {
	enc := UTF32Encoding(binary.LittleEndian)
	assert.Equal(t, "UTF-32LE", enc.Name())
	assert.Equal(t, BOMUTF32LE, enc.BOM())

	original := []byte("ASCII + €")
	encoded, err := enc.Encode(original)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded, err := enc.Decode(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func Test_DecodeUTF16String(t *testing.T) {
	// "hi" in UTF-16LE without BOM
	got, err := DecodeUTF16String([]byte{'h', 0, 'i', 0}, binary.LittleEndian)
	assert.NoError(t, err)
	assert.Equal(t, "hi", got)

	// Odd byte length is rejected.
	_, err = DecodeUTF16String([]byte{'h', 0, 'i'}, binary.LittleEndian)
	assert.Error(t, err)
}

func Test_EncodeUTF16_EmptyAndError(t *testing.T) {
	got, err := EncodeUTF16(nil, binary.LittleEndian)
	assert.NoError(t, err)
	assert.Nil(t, got)

	// Invalid UTF-8 input must surface as an error.
	_, err = EncodeUTF16([]byte{0xFF, 0xFE}, binary.LittleEndian)
	assert.Error(t, err)
}
