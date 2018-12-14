package strutil

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"

	"golang.org/x/text/encoding/unicode/utf32"

	"github.com/domonda/errors"
)

type BOM string

var (
	NoBOM BOM = ""

	// UTF-8, BOM bytes: EF BB BF
	BOMUTF8 BOM = "UTF-8"

	// UTF-16, big-endian, BOM bytes: FE FF
	BOMUTF16BigEndian BOM = "UTF-16, big-endian"

	// UTF-16, little-endian, BOM bytes: FF FE
	BOMUTF16LittleEndian BOM = "UTF-16, little-endian"

	// UTF-32, big-endian, BOM bytes: 00 00 FE FF
	BOMUTF32BigEndian BOM = "UTF-32, big-endian"

	// UTF-32, little-endian, BOM bytes: FF FE 00 00
	BOMUTF32LittleEndian BOM = "UTF-32, little-endian"
)

var (
	bomUTF8              = []byte{0xEF, 0xBB, 0xBF}
	bomUTF16BigEndian    = []byte{0xFE, 0xFF}
	bomUTF16LittleEndian = []byte{0xFF, 0xFE}
	bomUTF32BigEndian    = []byte{0x00, 0x00, 0xFE, 0xFF}
	bomUTF32LittleEndian = []byte{0xFF, 0xFE, 0x00, 0x00}
)

func StringBOM(str string) (bom BOM, length int) {
	return BytesBOM([]byte(str))
}

func BytesBOM(b []byte) (bom BOM, length int) {
	if bytes.HasPrefix(b, bomUTF8) {
		return BOMUTF8, len(bomUTF8)
	}
	if bytes.HasPrefix(b, bomUTF16BigEndian) {
		return BOMUTF16BigEndian, len(bomUTF16BigEndian)
	}
	if bytes.HasPrefix(b, bomUTF16LittleEndian) {
		return BOMUTF16LittleEndian, len(bomUTF16LittleEndian)
	}
	if bytes.HasPrefix(b, bomUTF32BigEndian) {
		return BOMUTF32BigEndian, len(bomUTF32BigEndian)
	}
	if bytes.HasPrefix(b, bomUTF32LittleEndian) {
		return BOMUTF32LittleEndian, len(bomUTF32LittleEndian)
	}
	return NoBOM, 0
}

func SplitBOM(b []byte) (BOM, []byte) {
	bom, length := BytesBOM(b)
	return bom, b[length:]
}

func DecodeUTF16(b []byte, order binary.ByteOrder) (string, error) {
	u16s := make([]uint16, 0, len(b)/2)

	for i, j := 0, len(b); i < j; i += 2 {
		u16s = append(u16s, order.Uint16(b[i:]))
	}

	runes := utf16.Decode(u16s)
	return string(runes), nil
}

func (bom BOM) DecodeRemaining(data []byte) (str string, err error) {
	switch bom {
	case NoBOM, BOMUTF8:
		return string(data), nil

	case BOMUTF16LittleEndian:
		return DecodeUTF16(data, binary.LittleEndian)

	case BOMUTF16BigEndian:
		return DecodeUTF16(data, binary.BigEndian)

	case BOMUTF32LittleEndian:
		data, err = utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM).NewDecoder().Bytes(data)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case BOMUTF32BigEndian:
		data, err = utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM).NewDecoder().Bytes(data)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return "", errors.Errorf("Unsupported BOM: %s", bom)
}

func (bom BOM) Bytes() []byte {
	switch bom {
	case BOMUTF8:
		return bomUTF8

	case BOMUTF16LittleEndian:
		return bomUTF16LittleEndian

	case BOMUTF16BigEndian:
		return bomUTF16BigEndian

	case BOMUTF32LittleEndian:
		return bomUTF32LittleEndian

	case BOMUTF32BigEndian:
		return bomUTF32BigEndian
	}

	return nil
}

func DecodeUTFWithOptionalBOM(b []byte) (str string, err error) {
	bom, data := SplitBOM(b)
	return bom.DecodeRemaining(data)
}
