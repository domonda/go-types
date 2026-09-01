package charset

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
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
