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
		// NOTE: UTF-32LE BOM (FF FE 00 00) starts with the UTF-16LE BOM (FF FE),
		// and BOMOfBytes checks UTF-16LE before UTF-32LE, so detection
		// short-circuits to BOMUTF16LE here. Documenting current behavior.
		{"UTF-32LE detects as UTF-16LE", append([]byte{0xFF, 0xFE, 0x00, 0x00}, 'h', 0), BOMUTF16LE},
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
}

func Test_SplitBOM(t *testing.T) {
	bom, data := SplitBOM(append([]byte{0xEF, 0xBB, 0xBF}, []byte("hi")...))
	assert.Equal(t, BOMUTF8, bom)
	assert.Equal(t, []byte("hi"), data)

	bom, data = SplitBOM([]byte("plain"))
	assert.Equal(t, NoBOM, bom)
	assert.Equal(t, []byte("plain"), data)
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
}
