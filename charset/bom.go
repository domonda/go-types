package charset

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// BOM is a Unicode Byte Order Mark
type BOM string

var (
	NoBOM BOM
	// UTF-8, BOM bytes: EF BB BF
	BOMUTF8 = BOM(bomUTF8)
	// UTF-16BE, BOM bytes: FE FF
	BOMUTF16BE = BOM(bomUTF16BE)
	// UTF-16LE, BOM bytes: FF FE
	BOMUTF16LE = BOM(bomUTF16LE)
	// UTF-32BE, BOM bytes: 00 00 FE FF
	BOMUTF32BE = BOM(bomUTF32BE)
	// UTF-32LE, BOM bytes: FF FE 00 00
	BOMUTF32LE = BOM(bomUTF32LE)
)

var (
	// UTF-8, BOM bytes: EF BB BF
	bomUTF8 = []byte{0xEF, 0xBB, 0xBF}
	// UTF-16BE, BOM bytes: FE FF
	bomUTF16BE = []byte{0xFE, 0xFF}
	// UTF-16LE, BOM bytes: FF FE
	bomUTF16LE = []byte{0xFF, 0xFE}
	// UTF-32BE, BOM bytes: 00 00 FE FF
	bomUTF32BE = []byte{0x00, 0x00, 0xFE, 0xFF}
	// UTF-32LE, BOM bytes: FF FE 00 00
	bomUTF32LE = []byte{0xFF, 0xFE, 0x00, 0x00}
)

func BOMOfString(str string) BOM {
	switch {
	case strings.HasPrefix(str, string(BOMUTF8)):
		return BOMUTF8
	case strings.HasPrefix(str, string(BOMUTF16BE)):
		return BOMUTF16BE
	case strings.HasPrefix(str, string(BOMUTF16LE)):
		return BOMUTF16LE
	case strings.HasPrefix(str, string(BOMUTF32BE)):
		return BOMUTF32BE
	case strings.HasPrefix(str, string(BOMUTF32LE)):
		return BOMUTF32LE
	}
	return NoBOM
}

func BOMOfBytes(b []byte) BOM {
	switch {
	case bytes.HasPrefix(b, bomUTF8):
		return BOMUTF8
	case bytes.HasPrefix(b, bomUTF16LE):
		return BOMUTF16LE
	case bytes.HasPrefix(b, bomUTF16BE):
		return BOMUTF16BE
	case bytes.HasPrefix(b, bomUTF32LE):
		return BOMUTF32LE
	case bytes.HasPrefix(b, bomUTF32BE):
		return BOMUTF32BE
	}
	return NoBOM
}

func TrimBOM(b []byte, bom BOM) []byte {
	if bytes.HasPrefix(b, bomUTF8) {
		return b[len(bomUTF8):]
	}
	return b
}

func SplitBOM(b []byte) (BOM, []byte) {
	bom := BOMOfBytes(b)
	return bom, b[len(bom):]
}

func DecodeWithBOM(b []byte) ([]byte, error) {
	bom, data := SplitBOM(b)
	return bom.Decode(data)
}

func DecodeStringWithBOM(b []byte) (string, error) {
	bom, data := SplitBOM(b)
	return bom.DecodeString(data)
}

func (bom BOM) Encoding() (Encoding, error) {
	switch bom {
	case NoBOM, BOMUTF8:
		return UTF8Encoding(), nil

	case BOMUTF16LE, BOMUTF16BE:
		return UTF16Encoding(bom.Endian()), nil

	case BOMUTF32LE, BOMUTF32BE:
		return UTF32Encoding(bom.Endian()), nil
	}

	return nil, fmt.Errorf("unsupported BOM: %v", []byte(bom))
}

func (bom BOM) Decode(data []byte) ([]byte, error) {
	dataBOM, data := SplitBOM(data)
	if dataBOM != NoBOM && dataBOM != bom {
		return nil, fmt.Errorf("wrong BOM in data: %v, expected: %v", []byte(dataBOM), []byte(bom))
	}

	switch bom {
	case NoBOM, BOMUTF8:
		return data, nil

	case BOMUTF16LE, BOMUTF16BE:
		return DecodeUTF16(data, bom.Endian())

	case BOMUTF32LE, BOMUTF32BE:
		return DecodeUTF32(data, bom.Endian())
	}

	return nil, fmt.Errorf("unsupported BOM: %v", []byte(bom))
}

func (bom BOM) DecodeString(data []byte) (string, error) {
	dataBOM, data := SplitBOM(data)
	if dataBOM != NoBOM && dataBOM != bom {
		return "", fmt.Errorf("wrong BOM in data: %v, expected: %v", []byte(dataBOM), []byte(bom))
	}

	switch bom {
	case NoBOM, BOMUTF8:
		return string(data), nil

	case BOMUTF16LE, BOMUTF16BE:
		return DecodeUTF16String(data, bom.Endian())

	case BOMUTF32LE, BOMUTF32BE:
		return DecodeUTF32String(data, bom.Endian())
	}

	return "", fmt.Errorf("unsupported BOM: %v", []byte(bom))
}

func (bom BOM) Endian() binary.ByteOrder {
	switch bom {
	case BOMUTF16LE, BOMUTF32LE:
		return binary.LittleEndian
	case BOMUTF16BE, BOMUTF32BE:
		return binary.BigEndian
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (bom BOM) String() string {
	switch bom {
	case NoBOM:
		return "No BOM"
	case BOMUTF8:
		return "UTF-8"
	case BOMUTF16BE:
		return "UTF-16BE"
	case BOMUTF16LE:
		return "UTF-16LE"
	case BOMUTF32BE:
		return "UTF-32BE"
	case BOMUTF32LE:
		return "UTF-32LE"
	}
	return fmt.Sprintf("Invalid BOM: %v", []byte(bom))
}
