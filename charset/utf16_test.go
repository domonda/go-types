package charset

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUTF16Encoding_String(t *testing.T) {
	assert.Equal(t, "UTF-16LE Encoding", UTF16Encoding(binary.LittleEndian).String())
	assert.Equal(t, "UTF-16BE Encoding", UTF16Encoding(binary.BigEndian).String())
}

func TestEncodeUTF16_ByteLayout(t *testing.T) {
	// "A€😀": U+0041, U+20AC and U+1F600 which needs the surrogate pair
	// D83D DE00. Encoded without a BOM, as documented.
	const text = "A€😀"
	tests := []struct {
		name      string
		byteOrder binary.ByteOrder
		want      []byte
	}{
		{name: "LE", byteOrder: binary.LittleEndian, want: []byte{0x41, 0x00, 0xAC, 0x20, 0x3D, 0xD8, 0x00, 0xDE}},
		{name: "BE", byteOrder: binary.BigEndian, want: []byte{0x00, 0x41, 0x20, 0xAC, 0xD8, 0x3D, 0xDE, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeUTF16([]byte(text), tt.byteOrder)
			require.NoError(t, err)
			assert.Equal(t, tt.want, encoded)

			decoded, err := DecodeUTF16(tt.want, tt.byteOrder)
			require.NoError(t, err)
			assert.Equal(t, text, string(decoded))

			str, err := DecodeUTF16String(tt.want, tt.byteOrder)
			require.NoError(t, err)
			assert.Equal(t, text, str)
		})
	}
}

func TestUTF16Encoding_RoundTrip(t *testing.T) {
	for _, byteOrder := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
		enc := UTF16Encoding(byteOrder)
		// "\uBBEF\uC5BF" and "\uEFBB\uBF41" encode to bytes starting with the UTF-8 BOM
		for _, text := range []string{"a", "Grüße 100 €", "世界 😀 𝄞", "\x00leading NUL", "\uBBEF\uC5BFx", "\uEFBB\uBF41"} {
			encoded, err := enc.Encode([]byte(text))
			require.NoError(t, err)
			assert.Len(t, encoded, 2*len(utf16.Encode([]rune(text))), "%s %q", enc, text)

			decoded, err := enc.Decode(encoded)
			require.NoError(t, err)
			assert.Equal(t, text, string(decoded), "%s", enc)
		}
	}
}

func TestDecodeUTF16_BOM(t *testing.T) {
	// A matching BOM is stripped for both byte orders
	got, err := DecodeUTF16([]byte{0xFE, 0xFF, 0x00, 'h', 0x00, 'i'}, binary.BigEndian)
	require.NoError(t, err)
	assert.Equal(t, "hi", string(got))
	got, err = DecodeUTF16([]byte{0xFF, 0xFE, 'h', 0x00, 'i', 0x00}, binary.LittleEndian)
	require.NoError(t, err)
	assert.Equal(t, "hi", string(got))

	// A BOM of the other byte order is an error,
	// decoding with the wrong byte order would produce garbage
	_, err = DecodeUTF16([]byte{0xFE, 0xFF, 0x00, 'h'}, binary.LittleEndian)
	assert.EqualError(t, err, "expected UTF-16LE BOM but got UTF-16BE")
	_, err = DecodeUTF16([]byte{0xFF, 0xFE, 'h', 0x00}, binary.BigEndian)
	assert.EqualError(t, err, "expected UTF-16BE BOM but got UTF-16LE")
}

// TestDecodeUTF16_BOMVariants checks that every UTF-16 decoder handles a
// leading BOM the same way: a matching BOM is stripped instead of ending up
// as U+FEFF at the start of the text, and a BOM of the other byte order is
// an error because decoding with the wrong byte order produces garbage.
// Data without a BOM is decoded with the passed byte order.
func TestDecodeUTF16_BOMVariants(t *testing.T) {
	decoders := []struct {
		name   string
		decode func(b []byte, byteOrder binary.ByteOrder) (string, error)
	}{
		{name: "DecodeUTF16", decode: func(b []byte, byteOrder binary.ByteOrder) (string, error) {
			decoded, err := DecodeUTF16(b, byteOrder)
			return string(decoded), err
		}},
		{name: "DecodeUTF16String", decode: DecodeUTF16String},
		{name: "UTF16Encoding.Decode", decode: func(b []byte, byteOrder binary.ByteOrder) (string, error) {
			decoded, err := UTF16Encoding(byteOrder).Decode(b)
			return string(decoded), err
		}},
	}
	tests := []struct {
		name      string
		b         []byte
		byteOrder binary.ByteOrder
		want      string
		wantErr   string
	}{
		{name: "LE matching BOM", b: []byte{0xFF, 0xFE, 'h', 0, 'i', 0}, byteOrder: binary.LittleEndian, want: "hi"},
		{name: "BE matching BOM", b: []byte{0xFE, 0xFF, 0, 'h', 0, 'i'}, byteOrder: binary.BigEndian, want: "hi"},
		// FF FE 00 00 is also the UTF-32LE BOM, but the passed byte order
		// already says that it is a UTF-16LE BOM followed by U+0000
		{name: "LE matching BOM followed by U+0000", b: []byte{0xFF, 0xFE, 0, 0, 'h', 0}, byteOrder: binary.LittleEndian, want: "\x00h"},
		{name: "LE opposite BOM", b: []byte{0xFE, 0xFF, 0, 'h'}, byteOrder: binary.LittleEndian, wantErr: "expected UTF-16LE BOM but got UTF-16BE"},
		{name: "BE opposite BOM", b: []byte{0xFF, 0xFE, 'h', 0}, byteOrder: binary.BigEndian, wantErr: "expected UTF-16BE BOM but got UTF-16LE"},
		{name: "LE no BOM", b: []byte{'h', 0, 'i', 0}, byteOrder: binary.LittleEndian, want: "hi"},
		{name: "BE no BOM", b: []byte{0, 'h', 0, 'i'}, byteOrder: binary.BigEndian, want: "hi"},
		// Valid UTF-16 text can start with the bytes of the UTF-8 BOM EF BB BF,
		// so they must not be mistaken for one: U+BBEF U+C5BF 'x' in LE
		// and U+EFBB U+BF41 in BE, as EncodeUTF16 writes them.
		{name: "LE text starting with UTF-8 BOM bytes", b: []byte{0xEF, 0xBB, 0xBF, 0xC5, 'x', 0}, byteOrder: binary.LittleEndian, want: "\uBBEF\uC5BFx"},
		{name: "BE text starting with UTF-8 BOM bytes", b: []byte{0xEF, 0xBB, 0xBF, 0x41}, byteOrder: binary.BigEndian, want: "\uEFBB\uBF41"},
	}
	for _, d := range decoders {
		for _, tt := range tests {
			t.Run(d.name+"/"+tt.name, func(t *testing.T) {
				got, err := d.decode(tt.b, tt.byteOrder)
				if tt.wantErr != "" {
					assert.EqualError(t, err, tt.wantErr)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			})
		}
	}
}

func TestDecodeUTF16_EmptyAndOddLength(t *testing.T) {
	got, err := DecodeUTF16(nil, binary.LittleEndian)
	assert.NoError(t, err)
	assert.Nil(t, got)

	str, err := DecodeUTF16String(nil, binary.BigEndian)
	assert.NoError(t, err)
	assert.Equal(t, "", str)

	// An odd number of bytes can't be UTF-16, most likely the data is
	// truncated or in a different encoding
	_, err = DecodeUTF16([]byte{'h', 0, 'i'}, binary.LittleEndian)
	assert.EqualError(t, err, "odd length of UTF-16 string: 3")
}

func TestDecodeUTF16_InvalidSurrogates(t *testing.T) {
	// Unpaired surrogates can't be represented in UTF-8, they must be
	// replaced with U+FFFD so that the result is always valid UTF-8
	tests := []struct {
		name string
		le   []byte
		want string
	}{
		{name: "lone high surrogate", le: []byte{0x3D, 0xD8, 'a', 0}, want: "\ufffda"},
		{name: "lone low surrogate", le: []byte{0x00, 0xDE, 'a', 0}, want: "\ufffda"},
		{name: "high surrogate at end", le: []byte{'a', 0, 0x3D, 0xD8}, want: "a\ufffd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeUTF16(tt.le, binary.LittleEndian)
			require.NoError(t, err)
			assert.True(t, utf8.Valid(got))
			assert.Equal(t, tt.want, string(got))
		})
	}
}

// TestEncodeUTF16_ReplacementCharacter protects that a valid U+FFFD in the
// input is encoded and not mistaken for invalid UTF-8. DecodeUTF16 itself
// produces U+FFFD for unpaired surrogates, so text decoded by this package
// must be encodable again.
func TestEncodeUTF16_ReplacementCharacter(t *testing.T) {
	encoded, err := EncodeUTF16([]byte("a�b"), binary.LittleEndian)
	require.NoError(t, err)
	assert.Equal(t, []byte{'a', 0x00, 0xFD, 0xFF, 'b', 0x00}, encoded)

	decoded, err := DecodeUTF16([]byte{0x3D, 0xD8, 'a', 0}, binary.LittleEndian)
	require.NoError(t, err)
	reencoded, err := EncodeUTF16(decoded, binary.LittleEndian)
	require.NoError(t, err)
	assert.Equal(t, []byte{0xFD, 0xFF, 'a', 0x00}, reencoded)
}

func TestEncodeUTF16_InvalidUTF8(t *testing.T) {
	// Invalid UTF-8 anywhere in the input is an error, not only at the start
	for _, in := range []string{"\xff", "abc\xff", "ab\xc3", "a\xed\xa0\x80b"} {
		_, err := EncodeUTF16([]byte(in), binary.BigEndian)
		assert.Error(t, err, "EncodeUTF16(%q)", in)
	}
}

// TestUTF16_NilByteOrder protects against a nil pointer panic:
// a nil binary.ByteOrder is an error like an invalid one,
// also for empty input, so a misconfigured encoding fails
// on the first call and not only on the first non-empty one.
func TestUTF16_NilByteOrder(t *testing.T) {
	for _, input := range [][]byte{nil, {}, {'a', 0}} {
		_, err := DecodeUTF16(input, nil)
		assert.Error(t, err, "DecodeUTF16(%v, nil)", input)
		_, err = EncodeUTF16(input, nil)
		assert.Error(t, err, "EncodeUTF16(%v, nil)", input)
	}
}
