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

func Test_DecodeUTF16_ExplicitByteOrderBeatsBOMAmbiguity(t *testing.T) {
	// FF FE 00 00 is the UTF-32LE BOM, but it is equally a valid UTF-16LE BOM
	// followed by U+0000. BOMOfBytes has to resolve that to UTF-32LE because it
	// is guessing, but a caller passing binary.LittleEndian is not guessing.
	// The explicit byte order must win, otherwise asking for the encoding you
	// already know turns valid input into an error.
	got, err := DecodeUTF16([]byte{0xFF, 0xFE, 0x00, 0x00, 'h', 0, 0x00, 0x00}, binary.LittleEndian)
	assert.NoError(t, err)
	assert.Equal(t, []byte("\x00h\x00"), got)

	// A UTF-16LE BOM not followed by 00 00 still decodes normally.
	got, err = DecodeUTF16([]byte{0xFF, 0xFE, 'h', 0, 'i', 0}, binary.LittleEndian)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), got)

	// A genuinely wrong BOM is still rejected: UTF-32BE cannot be UTF-16LE.
	_, err = DecodeUTF16([]byte{0x00, 0x00, 0xFE, 0xFF, 0, 0, 0, 'h'}, binary.LittleEndian)
	assert.EqualError(t, err, "expected UTF-16LE BOM but got UTF-32BE")
}

func Test_DecodeUTF16_NativeEndian(t *testing.T) {
	// binary.NativeEndian is a distinct binary.ByteOrder value, it is not equal
	// to binary.LittleEndian even on a little-endian machine. BOM-less data has
	// no BOM to validate against, so it must decode with any byte order rather
	// than being rejected for not being one of the two that have a BOM.
	// Encode with the same byte order the decoder is given so the test is not
	// hardcoded to a little-endian host.
	hi := make([]byte, 4)
	binary.NativeEndian.PutUint16(hi[0:], 'h')
	binary.NativeEndian.PutUint16(hi[2:], 'i')
	got, err := DecodeUTF16(hi, binary.NativeEndian)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), got)

	// With a BOM present there is nothing to validate it against, so the
	// unusable byte order becomes fatal.
	_, err = DecodeUTF16([]byte{0xFF, 0xFE, 'h', 0}, binary.NativeEndian)
	assert.EqualError(t, err, "invalid binary.ByteOrder: NativeEndian")
}

func Test_BOMDecode_ExplicitBOMBeatsAmbiguity(t *testing.T) {
	// Same rule one level up: BOM.Decode is the known-encoding entry point, so
	// the BOM the caller names wins over what BOMOfBytes would guess.
	utf16le, err := BOMUTF16LE.DecodeString([]byte{0xFF, 0xFE, 0x00, 0x00, 'h', 0, 0x00, 0x00})
	assert.NoError(t, err)
	assert.Equal(t, "\x00h\x00", utf16le)

	// The very same bytes read as UTF-32LE when that is what the caller names.
	utf32le, err := BOMUTF32LE.DecodeString([]byte{0xFF, 0xFE, 0x00, 0x00, 'h', 0, 0x00, 0x00})
	assert.NoError(t, err)
	assert.Equal(t, "h", utf32le)

	// A mismatched BOM is still an error.
	_, err = BOMUTF8.DecodeString([]byte{0xFE, 0xFF, 0x00, 'h'})
	assert.Error(t, err)
}

func Test_AutoDecode_UTF32LEBOM(t *testing.T) {
	// AutoDecode routes through SplitBOM, so it inherits the BOM detection
	// order. It has to report UTF-32LE and decode all 4 byte code units,
	// not stop at the 2 byte UTF-16LE BOM prefix.
	text, encName, err := AutoDecode([]byte{0xFF, 0xFE, 0x00, 0x00, 'h', 0, 0, 0, 'i', 0, 0, 0}, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), text)
	assert.Equal(t, "UTF-32LE", encName)
}
