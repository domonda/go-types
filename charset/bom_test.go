package charset

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BOMOfBytes(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want BOM
	}{
		{"empty", []byte{}, NoBOM},
		{"no BOM", []byte("hello"), NoBOM},
		{"UTF-8", append([]byte{0xEF, 0xBB, 0xBF}, []byte("hi")...), BOMUTF8},
		{"UTF-16BE", append([]byte{0xFE, 0xFF}, 0, 'h'), BOMUTF16BE},
		{"UTF-16LE", append([]byte{0xFF, 0xFE}, 'h', 0), BOMUTF16LE},
		{"UTF-32BE", append([]byte{0x00, 0x00, 0xFE, 0xFF}, 0, 0, 0, 'h'), BOMUTF32BE},

		// The UTF-16LE BOM (FF FE) is a prefix of the UTF-32LE BOM (FF FE 00 00),
		// so the longer BOM has to win, otherwise UTF-32LE data is silently
		// decoded as UTF-16LE. Matching longest-first is what ICU and .NET do.
		{"UTF-32LE", append([]byte{0xFF, 0xFE, 0x00, 0x00}, 'h', 0, 0, 0), BOMUTF32LE},

		// Only FF FE 00 00 may win over UTF-16LE, FF FE followed by
		// anything else must still be UTF-16LE.
		{"UTF-16LE with 00 in second byte pair", append([]byte{0xFF, 0xFE}, 'h', 0x00, 0x00, 0x01), BOMUTF16LE},

		// Too short to be the UTF-32LE BOM, so it can only be UTF-16LE.
		{"UTF-16LE BOM only", []byte{0xFF, 0xFE}, BOMUTF16LE},
		{"UTF-16LE BOM plus one byte", []byte{0xFF, 0xFE, 0x00}, BOMUTF16LE},

		// Known and accepted ambiguity: UTF-16LE text starting with U+0000
		// serializes to the same bytes as the UTF-32LE BOM and is reported as
		// UTF-32LE. A leading NUL is not plain text in practice, so UTF-32LE
		// is the conventional reading of these bytes.
		{"UTF-16LE starting with U+0000 reads as UTF-32LE", []byte{0xFF, 0xFE, 0x00, 0x00, 'h', 0}, BOMUTF32LE},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BOMOfBytes(tt.in))
		})
	}
}

func Test_BOMOfString(t *testing.T) {
	assert.Equal(t, BOMUTF8, BOMOfString("\xEF\xBB\xBFhi"))
	assert.Equal(t, NoBOM, BOMOfString("hello"))

	// Same longest-match-first ordering as BOMOfBytes, see Test_BOMOfBytes.
	assert.Equal(t, BOMUTF32LE, BOMOfString("\xFF\xFE\x00\x00h\x00\x00\x00"))
	assert.Equal(t, BOMUTF32BE, BOMOfString("\x00\x00\xFE\xFF\x00\x00\x00h"))
	assert.Equal(t, BOMUTF16LE, BOMOfString("\xFF\xFEh\x00"))
	assert.Equal(t, BOMUTF16BE, BOMOfString("\xFE\xFF\x00h"))
	assert.Equal(t, BOMUTF16LE, BOMOfString("\xFF\xFE"))
}

func Test_SplitBOM(t *testing.T) {
	bom, data := SplitBOM(append([]byte{0xEF, 0xBB, 0xBF}, []byte("hi")...))
	assert.Equal(t, BOMUTF8, bom)
	assert.Equal(t, []byte("hi"), data)

	bom, data = SplitBOM([]byte("plain"))
	assert.Equal(t, NoBOM, bom)
	assert.Equal(t, []byte("plain"), data)

	// All 4 bytes of the UTF-32LE BOM have to be split off, splitting only the
	// first 2 would leave a bogus 00 00 in front of the payload.
	bom, data = SplitBOM([]byte{0xFF, 0xFE, 0x00, 0x00, 'h', 0, 0, 0})
	assert.Equal(t, BOMUTF32LE, bom)
	assert.Equal(t, []byte{'h', 0, 0, 0}, data)
}

func Test_TrimBOM(t *testing.T) {
	// Strips the matching BOM.
	assert.Equal(t, []byte("hi"), TrimBOM([]byte("\xEF\xBB\xBFhi"), BOMUTF8))
	assert.Equal(t, []byte("hi"), TrimBOM([]byte("\xFF\xFEhi"), BOMUTF16LE))

	// Non-matching BOM leaves b unchanged.
	assert.Equal(t, []byte("\xEF\xBB\xBFhi"), TrimBOM([]byte("\xEF\xBB\xBFhi"), BOMUTF16LE))

	// NoBOM and a BOM-less input are both no-ops.
	assert.Equal(t, []byte("\xEF\xBB\xBFhi"), TrimBOM([]byte("\xEF\xBB\xBFhi"), NoBOM))
	assert.Equal(t, []byte("plain"), TrimBOM([]byte("plain"), BOMUTF8))
}

func Test_BOM_Endian(t *testing.T) {
	assert.Equal(t, binary.LittleEndian, BOMUTF16LE.Endian())
	assert.Equal(t, binary.BigEndian, BOMUTF16BE.Endian())
	assert.Equal(t, binary.LittleEndian, BOMUTF32LE.Endian())
	assert.Equal(t, binary.BigEndian, BOMUTF32BE.Endian())
	assert.Nil(t, NoBOM.Endian())
}

func Test_BOM_String(t *testing.T) {
	assert.Equal(t, "No BOM", NoBOM.String())
	assert.Equal(t, "UTF-8", BOMUTF8.String())
	assert.Equal(t, "UTF-16BE", BOMUTF16BE.String())
	assert.Equal(t, "UTF-16LE", BOMUTF16LE.String())
}

func Test_DecodeWithBOM(t *testing.T) {
	// UTF-8 BOM passthrough
	got, err := DecodeWithBOM(append([]byte{0xEF, 0xBB, 0xBF}, []byte("hello")...))
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), got)

	// No BOM → returned as-is
	got, err = DecodeWithBOM([]byte("plain"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("plain"), got)

	// UTF-32LE was decoded as UTF-16LE before the BOM ordering fix, which
	// silently produced "\x00h\x00i\x00" instead of an error.
	got, err = DecodeWithBOM([]byte{0xFF, 0xFE, 0x00, 0x00, 'h', 0, 0, 0, 'i', 0, 0, 0})
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), got)

	str, err := DecodeStringWithBOM([]byte{0xFF, 0xFE, 0x00, 0x00, 'h', 0, 0, 0, 'i', 0, 0, 0})
	assert.NoError(t, err)
	assert.Equal(t, "hi", str)
}

// allBOMs are the byte order marks that have an Encoding
var allBOMs = []BOM{BOMUTF8, BOMUTF16LE, BOMUTF16BE, BOMUTF32LE, BOMUTF32BE}

func Test_BOM_Encoding(t *testing.T) {
	for _, bom := range allBOMs {
		t.Run(bom.String(), func(t *testing.T) {
			enc, err := bom.Encoding()
			require.NoError(t, err)
			// The Encoding must be the one the BOM names and report the
			// same BOM back, so that encoding text and writing enc.BOM()
			// in front of it produces data that is detected again.
			assert.Equal(t, bom.String(), enc.Name())
			assert.Equal(t, bom, enc.BOM())
		})
	}

	// Data without BOM is assumed to be UTF-8
	enc, err := NoBOM.Encoding()
	require.NoError(t, err)
	assert.Equal(t, "UTF-8", enc.Name())

	_, err = BOM("xy").Encoding()
	assert.EqualError(t, err, "unsupported BOM: [120 121]")
}

// Test_BOM_RoundTrip checks that text encoded with the Encoding of a BOM
// and prefixed with that BOM decodes back to the original text through
// every BOM aware decoding function.
func Test_BOM_RoundTrip(t *testing.T) {
	// Must not start with U+0000 because of the documented
	// UTF-16LE/UTF-32LE ambiguity, see Test_BOMOfBytes
	const text = "Grüße 100 € 世界 😀"
	for _, bom := range allBOMs {
		t.Run(bom.String(), func(t *testing.T) {
			enc, err := bom.Encoding()
			require.NoError(t, err)
			encoded, err := enc.Encode([]byte(text))
			require.NoError(t, err)
			withBOM := append([]byte(bom), encoded...)

			assert.Equal(t, bom, BOMOfBytes(withBOM))
			assert.Equal(t, bom, BOMOfString(string(withBOM)))

			decoded, err := DecodeWithBOM(withBOM)
			require.NoError(t, err)
			assert.Equal(t, text, string(decoded), "DecodeWithBOM")

			str, err := DecodeStringWithBOM(withBOM)
			require.NoError(t, err)
			assert.Equal(t, text, str, "DecodeStringWithBOM")

			// With a known BOM, the BOM in the data is optional
			for _, data := range [][]byte{withBOM, encoded} {
				decoded, err = bom.Decode(data)
				require.NoError(t, err)
				assert.Equal(t, text, string(decoded), "BOM.Decode")

				str, err = bom.DecodeString(data)
				require.NoError(t, err)
				assert.Equal(t, text, str, "BOM.DecodeString")
			}
		})
	}
}

func Test_BOM_Decode_WrongBOM(t *testing.T) {
	// Data that announces a different encoding than the caller expects
	// must be rejected instead of being decoded into garbage
	data := []byte{0xFE, 0xFF, 0x00, 'h'} // UTF-16BE
	for _, bom := range []BOM{BOMUTF8, BOMUTF16LE, BOMUTF32LE, BOMUTF32BE} {
		_, err := bom.Decode(data)
		assert.Error(t, err, "%s.Decode", bom)
		_, err = bom.DecodeString(data)
		assert.Error(t, err, "%s.DecodeString", bom)
	}
}

// Test_BOM_Decode_UTF32TextStartingWithUTF16BOM checks that the BOM aware
// functions decode valid UTF-32LE text like DecodeUTF32 does, even if its
// first bytes FF FE look like the UTF-16LE BOM.
func Test_BOM_Decode_UTF32TextStartingWithUTF16BOM(t *testing.T) {
	const text = "\U000FFEFFabc"
	encoded, err := EncodeUTF32([]byte(text), binary.LittleEndian)
	require.NoError(t, err)
	require.Equal(t, BOMUTF16LE, BOMOfBytes(encoded))

	decoded, err := BOMUTF32LE.Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, text, string(decoded), "BOM.Decode")
	str, err := BOMUTF32LE.DecodeString(encoded)
	require.NoError(t, err)
	assert.Equal(t, text, str, "BOM.DecodeString")
	decoded, err = DecodeWithBOM(append([]byte(BOMUTF32LE), encoded...))
	require.NoError(t, err)
	assert.Equal(t, text, string(decoded), "DecodeWithBOM")
}

func Test_BOM_Decode_NoBOM(t *testing.T) {
	// NoBOM means UTF-8 without BOM, the bytes are passed through
	decoded, err := NoBOM.Decode([]byte("Grüße"))
	require.NoError(t, err)
	assert.Equal(t, "Grüße", string(decoded))

	str, err := NoBOM.DecodeString([]byte("Grüße"))
	require.NoError(t, err)
	assert.Equal(t, "Grüße", str)

	// but data that starts with a BOM can't be BOM-less
	_, err = NoBOM.Decode(append([]byte(BOMUTF8), 'a'))
	assert.Error(t, err)
}

func Test_BOM_Decode_Unsupported(t *testing.T) {
	invalid := BOM("xy")
	_, err := invalid.Decode([]byte("xyz"))
	assert.EqualError(t, err, "unsupported BOM: [120 121]")
	_, err = invalid.DecodeString([]byte("z"))
	assert.EqualError(t, err, "unsupported BOM: [120 121]")
}

func Test_DecodeWithBOM_Errors(t *testing.T) {
	// Decoding errors after the BOM are returned, not swallowed
	_, err := DecodeWithBOM([]byte{0xFE, 0xFF, 0x00, 'h', 'i'}) // odd UTF-16 length
	assert.Error(t, err)
	_, err = DecodeStringWithBOM([]byte{0xFE, 0xFF, 0x00, 'h', 'i'})
	assert.Error(t, err)

	// Empty input has no BOM and decodes to nothing
	decoded, err := DecodeWithBOM(nil)
	assert.NoError(t, err)
	assert.Empty(t, decoded)
}

func Test_BOM_String_UTF32AndInvalid(t *testing.T) {
	assert.Equal(t, "UTF-32BE", BOMUTF32BE.String())
	assert.Equal(t, "UTF-32LE", BOMUTF32LE.String())
	// Invalid values show their bytes for debugging
	assert.Equal(t, "Invalid BOM: [120 121]", BOM("xy").String())
}

// TestBOM_DecodeKeepsTextBOM protects that only the one leading BOM is
// removed: a U+FEFF after it is a character of the text, like a zero
// width no-break space, and must be decoded instead of stripped again.
func TestBOM_DecodeKeepsTextBOM(t *testing.T) {
	tests := []struct {
		bom  BOM
		data []byte
	}{
		{bom: BOMUTF32LE, data: []byte{0xFF, 0xFE, 0, 0, 0xFF, 0xFE, 0, 0, 'a', 0, 0, 0}},
		{bom: BOMUTF32BE, data: []byte{0, 0, 0xFE, 0xFF, 0, 0, 0xFE, 0xFF, 0, 0, 0, 'a'}},
		{bom: BOMUTF16LE, data: []byte{0xFF, 0xFE, 0xFF, 0xFE, 'a', 0}},
		{bom: BOMUTF16BE, data: []byte{0xFE, 0xFF, 0xFE, 0xFF, 0, 'a'}},
	}
	for _, tt := range tests {
		decoded, err := tt.bom.Decode(tt.data)
		require.NoError(t, err, tt.bom.String())
		assert.Equal(t, "\ufeffa", string(decoded), tt.bom.String())
		str, err := tt.bom.DecodeString(tt.data)
		require.NoError(t, err, tt.bom.String())
		assert.Equal(t, "\ufeffa", str, tt.bom.String())
	}
}

// TestDecodeWithBOM_KeepsTextBOM protects that the BOM detecting functions
// strip one BOM like BOM.Decode: they split the BOM off and then let
// BOM.Decode strip a second one, a U+FEFF of the text.
func TestDecodeWithBOM_KeepsTextBOM(t *testing.T) {
	data := []byte{0xFF, 0xFE, 0xFF, 0xFE, 'A', 0}
	decoded, err := DecodeWithBOM(data)
	require.NoError(t, err)
	assert.Equal(t, "\ufeffA", string(decoded))
	str, err := DecodeStringWithBOM(data)
	require.NoError(t, err)
	assert.Equal(t, "\ufeffA", str)
	text, encName, err := AutoDecode(data, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "\ufeffA", string(text), encName)
}
