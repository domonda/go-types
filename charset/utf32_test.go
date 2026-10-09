package charset

import (
	"encoding/binary"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUTF32Encoding_NameBOMString(t *testing.T) {
	be := UTF32Encoding(binary.BigEndian)
	assert.Equal(t, "UTF-32BE", be.Name())
	assert.Equal(t, "UTF-32BE Encoding", be.String())
	assert.Equal(t, BOMUTF32BE, be.BOM())

	le := UTF32Encoding(binary.LittleEndian)
	assert.Equal(t, "UTF-32LE", le.Name())
	assert.Equal(t, "UTF-32LE Encoding", le.String())
	assert.Equal(t, BOMUTF32LE, le.BOM())
}

func TestEncodeUTF32_ByteLayout(t *testing.T) {
	// "A€😀": every code point is one 4 byte unit, no surrogates.
	// Encoded without a BOM, as documented.
	const text = "A€😀"
	tests := []struct {
		name      string
		byteOrder binary.ByteOrder
		want      []byte
	}{
		{name: "LE", byteOrder: binary.LittleEndian, want: []byte{0x41, 0, 0, 0, 0xAC, 0x20, 0, 0, 0x00, 0xF6, 0x01, 0x00}},
		{name: "BE", byteOrder: binary.BigEndian, want: []byte{0, 0, 0, 0x41, 0, 0, 0x20, 0xAC, 0x00, 0x01, 0xF6, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeUTF32([]byte(text), tt.byteOrder)
			require.NoError(t, err)
			assert.Equal(t, tt.want, encoded)

			// The Encoding returned by UTF32Encoding must produce the same
			// bytes as the function
			encoded, err = UTF32Encoding(tt.byteOrder).Encode([]byte(text))
			require.NoError(t, err)
			assert.Equal(t, tt.want, encoded)

			decoded, err := DecodeUTF32(tt.want, tt.byteOrder)
			require.NoError(t, err)
			assert.Equal(t, text, string(decoded))

			str, err := DecodeUTF32String(tt.want, tt.byteOrder)
			require.NoError(t, err)
			assert.Equal(t, text, str)
		})
	}
}

func TestUTF32_RoundTrip(t *testing.T) {
	for _, byteOrder := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
		enc := UTF32Encoding(byteOrder)
		for _, text := range []string{"a", "Grüße 100 €", "世界 😀 𝄞", "\x00leading NUL"} {
			encoded, err := EncodeUTF32([]byte(text), byteOrder)
			require.NoError(t, err)
			assert.Len(t, encoded, 4*utf8.RuneCountInString(text), "%s %q", enc, text)

			decoded, err := DecodeUTF32(encoded, byteOrder)
			require.NoError(t, err)
			assert.Equal(t, text, string(decoded), "%s", enc)

			decoded, err = enc.Decode(encoded)
			require.NoError(t, err)
			assert.Equal(t, text, string(decoded), "%s", enc)
		}
	}
}

func TestUTF32_Empty(t *testing.T) {
	encoded, err := EncodeUTF32(nil, binary.LittleEndian)
	assert.NoError(t, err)
	assert.Nil(t, encoded)

	decoded, err := DecodeUTF32(nil, binary.BigEndian)
	assert.NoError(t, err)
	assert.Nil(t, decoded)

	str, err := DecodeUTF32String(nil, binary.LittleEndian)
	assert.NoError(t, err)
	assert.Equal(t, "", str)
}

func TestDecodeUTF32_InvalidCodePoints(t *testing.T) {
	// Values that are not Unicode scalar values must be replaced with
	// U+FFFD so that the result is always valid UTF-8
	tests := []struct {
		name string
		le   []byte
		want string
	}{
		{name: "beyond U+10FFFF", le: []byte{0x00, 0x00, 0x11, 0x00, 'a', 0, 0, 0}, want: "\ufffda"},
		{name: "surrogate", le: []byte{0x00, 0xD8, 0x00, 0x00, 'a', 0, 0, 0}, want: "\ufffda"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeUTF32(tt.le, binary.LittleEndian)
			require.NoError(t, err)
			assert.True(t, utf8.Valid(got))
			assert.Equal(t, tt.want, string(got))

			str, err := DecodeUTF32String(tt.le, binary.LittleEndian)
			require.NoError(t, err)
			assert.Equal(t, tt.want, str)
		})
	}
}
