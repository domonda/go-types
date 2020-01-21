package charset

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/domonda/errors"
)

type BOM string

var (
	NoBOM BOM
	// UTF-8, BOM bytes: EF BB BF
	BOMUTF8 = BOM(bomUTF8)
	// UTF-16, big-endian, BOM bytes: FE FF
	BOMUTF16BigEndian = BOM(bomUTF16BigEndian)
	// UTF-16, little-endian, BOM bytes: FF FE
	BOMUTF16LittleEndian = BOM(bomUTF16LittleEndian)
	// UTF-32, big-endian, BOM bytes: 00 00 FE FF
	BOMUTF32BigEndian = BOM(bomUTF32BigEndian)
	// UTF-32, little-endian, BOM bytes: FF FE 00 00
	BOMUTF32LittleEndian = BOM(bomUTF32LittleEndian)
)

var (
	// UTF-8, BOM bytes: EF BB BF
	bomUTF8 = []byte{0xEF, 0xBB, 0xBF}
	// UTF-16, big-endian, BOM bytes: FE FF
	bomUTF16BigEndian = []byte{0xFE, 0xFF}
	// UTF-16, little-endian, BOM bytes: FF FE
	bomUTF16LittleEndian = []byte{0xFF, 0xFE}
	// UTF-32, big-endian, BOM bytes: 00 00 FE FF
	bomUTF32BigEndian = []byte{0x00, 0x00, 0xFE, 0xFF}
	// UTF-32, little-endian, BOM bytes: FF FE 00 00
	bomUTF32LittleEndian = []byte{0xFF, 0xFE, 0x00, 0x00}
)

func BOMOfString(str string) BOM {
	switch {
	case strings.HasPrefix(str, string(BOMUTF8)):
		return BOMUTF8
	case strings.HasPrefix(str, string(BOMUTF16BigEndian)):
		return BOMUTF16BigEndian
	case strings.HasPrefix(str, string(BOMUTF16LittleEndian)):
		return BOMUTF16LittleEndian
	case strings.HasPrefix(str, string(BOMUTF32BigEndian)):
		return BOMUTF32BigEndian
	case strings.HasPrefix(str, string(BOMUTF32LittleEndian)):
		return BOMUTF32LittleEndian
	}
	return NoBOM
}

func BOMOfBytes(b []byte) BOM {
	switch {
	case bytes.HasPrefix(b, bomUTF8):
		return BOMUTF8
	case bytes.HasPrefix(b, bomUTF16BigEndian):
		return BOMUTF16BigEndian
	case bytes.HasPrefix(b, bomUTF16LittleEndian):
		return BOMUTF16LittleEndian
	case bytes.HasPrefix(b, bomUTF32BigEndian):
		return BOMUTF32BigEndian
	case bytes.HasPrefix(b, bomUTF32LittleEndian):
		return BOMUTF32LittleEndian
	}
	return NoBOM
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
		return UTF8Encoding{}, nil

	case BOMUTF16LittleEndian, BOMUTF16BigEndian:
		return NewUTF16Encoding(bom.Endian()), nil

	case BOMUTF32LittleEndian, BOMUTF32BigEndian:
		return NewUTF32Encoding(bom.Endian()), nil
	}

	return nil, errors.Errorf("unsupported BOM: %v", []byte(bom))
}

func (bom BOM) Decode(data []byte) ([]byte, error) {
	dataBOM, data := SplitBOM(data)
	if dataBOM != NoBOM && dataBOM != bom {
		return nil, errors.Errorf("wrong BOM in data: %s, expected: %s", dataBOM, bom)
	}

	switch bom {
	case NoBOM, BOMUTF8:
		return data, nil

	case BOMUTF16LittleEndian, BOMUTF16BigEndian:
		return DecodeUTF16(data, bom.Endian())

	case BOMUTF32LittleEndian, BOMUTF32BigEndian:
		return DecodeUTF32(data, bom.Endian())
	}

	return nil, errors.Errorf("unsupported BOM: %v", []byte(bom))
}

func (bom BOM) DecodeString(data []byte) (string, error) {
	dataBOM, data := SplitBOM(data)
	if dataBOM != NoBOM && dataBOM != bom {
		return "", errors.Errorf("wrong BOM in data: %s, expected: %s", dataBOM, bom)
	}

	switch bom {
	case NoBOM, BOMUTF8:
		return string(data), nil

	case BOMUTF16LittleEndian, BOMUTF16BigEndian:
		return DecodeUTF16String(data, bom.Endian())

	case BOMUTF32LittleEndian, BOMUTF32BigEndian:
		return DecodeUTF32String(data, bom.Endian())
	}

	return "", errors.Errorf("unsupported BOM: %v", []byte(bom))
}

func (bom BOM) Endian() binary.ByteOrder {
	switch bom {
	case BOMUTF16LittleEndian, BOMUTF32LittleEndian:
		return binary.LittleEndian
	case BOMUTF16BigEndian, BOMUTF32BigEndian:
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
	case BOMUTF16BigEndian:
		return "UTF-16, big-endian"
	case BOMUTF16LittleEndian:
		return "UTF-16, little-endian"
	case BOMUTF32BigEndian:
		return "UTF-32, big-endian"
	case BOMUTF32LittleEndian:
		return "UTF-32, little-endian"
	}
	return fmt.Sprintf("Invalid BOM: %#v", string(bom))
}
