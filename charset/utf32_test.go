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

// TestDecodeUTF32_BOMVariants checks that every UTF-32 decoder handles a
// leading BOM like the UTF-16 decoders do: a matching BOM is stripped
// instead of ending up as U+FEFF at the start of the text, and a BOM of
// the other byte order is an error instead of silently decoding to U+FFFD,
// because decoding with the wrong byte order produces garbage.
// Data without a BOM is decoded with the passed byte order.
func TestDecodeUTF32_BOMVariants(t *testing.T) {
	decoders := []struct {
		name   string
		decode func(b []byte, byteOrder binary.ByteOrder) (string, error)
	}{
		{name: "DecodeUTF32", decode: func(b []byte, byteOrder binary.ByteOrder) (string, error) {
			decoded, err := DecodeUTF32(b, byteOrder)
			return string(decoded), err
		}},
		{name: "DecodeUTF32String", decode: DecodeUTF32String},
		{name: "UTF32Encoding.Decode", decode: func(b []byte, byteOrder binary.ByteOrder) (string, error) {
			decoded, err := UTF32Encoding(byteOrder).Decode(b)
			return string(decoded), err
		}},
	}
	var (
		bomLE = []byte{0xFF, 0xFE, 0, 0}
		bomBE = []byte{0, 0, 0xFE, 0xFF}
		hiLE  = []byte{'h', 0, 0, 0, 'i', 0, 0, 0}
		hiBE  = []byte{0, 0, 0, 'h', 0, 0, 0, 'i'}
	)
	tests := []struct {
		name      string
		b         []byte
		byteOrder binary.ByteOrder
		want      string
		wantErr   string
	}{
		{name: "LE matching BOM", b: append(bomLE, hiLE...), byteOrder: binary.LittleEndian, want: "hi"},
		{name: "BE matching BOM", b: append(bomBE, hiBE...), byteOrder: binary.BigEndian, want: "hi"},
		{name: "LE only BOM", b: bomLE, byteOrder: binary.LittleEndian, want: ""},
		{name: "LE opposite BOM", b: append(bomBE, hiBE...), byteOrder: binary.LittleEndian, wantErr: "expected UTF-32LE BOM but got UTF-32BE"},
		{name: "BE opposite BOM", b: append(bomLE, hiLE...), byteOrder: binary.BigEndian, wantErr: "expected UTF-32BE BOM but got UTF-32LE"},
		{name: "LE no BOM", b: hiLE, byteOrder: binary.LittleEndian, want: "hi"},
		{name: "BE no BOM", b: hiBE, byteOrder: binary.BigEndian, want: "hi"},
		// Valid UTF-32LE text can start with the UTF-16LE BOM FF FE,
		// like U+FFEFF, so it must not be mistaken for one
		{name: "LE text starting with UTF-16LE BOM bytes", b: append([]byte{0xFF, 0xFE, 0x0F, 0}, hiLE...), byteOrder: binary.LittleEndian, want: "\U000FFEFFhi"},
		// The same BOM followed by UTF-16LE text is no valid UTF-32LE character
		{name: "LE UTF-16LE text", b: []byte{0xFF, 0xFE, 'h', 0, 'i', 0, 'x', 0}, byteOrder: binary.LittleEndian, wantErr: "expected UTF-32LE BOM but got UTF-16LE"},
		{name: "BE UTF-8 BOM", b: []byte{0xEF, 0xBB, 0xBF, 'h'}, byteOrder: binary.BigEndian, wantErr: "expected UTF-32BE BOM but got UTF-8"},
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
