package charset

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

// UTF16Encoding returns an UTF-16 Encoding with the passed binary.ByteOrder
func UTF16Encoding(byteOrder binary.ByteOrder) Encoding {
	return utf16Encoding{byteOrder}
}

// utf16Encoding implements Encoding for UTF-16
type utf16Encoding struct {
	byteOrder binary.ByteOrder
}

func (e utf16Encoding) Encode(utf8Str []byte) (encodedStr []byte, err error) {
	return EncodeUTF16(utf8Str, e.byteOrder)
}

func (e utf16Encoding) Decode(encodedStr []byte) (utf8Str []byte, err error) {
	return DecodeUTF16(encodedStr, e.byteOrder)
}

func (e utf16Encoding) Name() string {
	if e.byteOrder == binary.BigEndian {
		return "UTF-16BE"
	}
	return "UTF-16LE"
}

func (e utf16Encoding) String() string {
	return e.Name() + " Encoding"
}

func (e utf16Encoding) BOM() BOM {
	if e.byteOrder == binary.BigEndian {
		return BOMUTF16BE
	}
	return BOMUTF16LE
}

func decodeUTF16Runes(b []byte, byteOrder binary.ByteOrder) []rune {
	numRunes := len(b) / 2
	u16s := make([]uint16, numRunes)
	for i := 0; i < numRunes; i++ {
		u16s[i] = byteOrder.Uint16(b[i*2:])
	}
	return utf16.Decode(u16s)
}

func DecodeUTF16(b []byte, byteOrder binary.ByteOrder) ([]byte, error) {
	if len(b) == 0 {
		return nil, nil
	}
	if len(b)&1 != 0 {
		return nil, fmt.Errorf("odd length of UTF-16 string: %d", len(b))
	}

	// Check for BOM and remove it before decoding
	if bom := BOMOfBytes(b); bom != NoBOM {
		switch byteOrder {
		case binary.LittleEndian:
			if bom != BOMUTF16LE {
				return nil, fmt.Errorf("expected %s BOM but got %s", BOMUTF16LE, bom)
			}
		case binary.BigEndian:
			if bom != BOMUTF16BE {
				return nil, fmt.Errorf("expected %s BOM but got %s", BOMUTF16BE, bom)
			}
		default:
			return nil, fmt.Errorf("invalid binary.ByteOrder: %v", byteOrder)
		}
		b = b[len(bom):]
	}

	runes := decodeUTF16Runes(b, byteOrder)
	buf := bytes.Buffer{}
	buf.Grow(len(runes))
	for _, r := range runes {
		buf.WriteRune(r)
	}
	return buf.Bytes(), nil
}

func EncodeUTF16(b []byte, byteOrder binary.ByteOrder) ([]byte, error) {
	if len(b) == 0 {
		return nil, nil
	}
	buf := bytes.Buffer{}
	buf.Grow(len(b) * 2)
	u16Bytes := make([]byte, 2)
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r == utf8.RuneError {
			return nil, errors.New("invalid UTF-8 rune")
		}
		for _, u16 := range utf16.Encode([]rune{r}) {
			byteOrder.PutUint16(u16Bytes, u16)
			_, err := buf.Write(u16Bytes)
			if err != nil {
				return nil, err
			}
		}
		b = b[size:]
	}
	return buf.Bytes(), nil
}

func DecodeUTF16String(b []byte, byteOrder binary.ByteOrder) (string, error) {
	if len(b)&1 != 0 {
		return "", fmt.Errorf("odd length of UTF-16 string: %d", len(b))
	}
	return string(decodeUTF16Runes(b, byteOrder)), nil
}
