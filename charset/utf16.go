package charset

import (
	"bytes"
	"encoding/binary"
	"errors"
	"unicode/utf16"
	"unicode/utf8"
)

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
		return nil, errors.New("odd length of UTF-16 string")
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
		return "", errors.New("odd length of UTF-16 string")
	}
	return string(decodeUTF16Runes(b, byteOrder)), nil
}

// UTF16Encoding implements Encoding for UTF-16
type UTF16Encoding struct {
	byteOrder binary.ByteOrder
}

func NewUTF16Encoding(byteOrder binary.ByteOrder) Encoding {
	return &UTF16Encoding{byteOrder}
}

func (e *UTF16Encoding) Encode(utf8Str []byte) (encodedStr []byte, err error) {
	return EncodeUTF16(utf8Str, e.byteOrder)
}

func (e *UTF16Encoding) Decode(encodedStr []byte) (utf8Str []byte, err error) {
	return DecodeUTF16(encodedStr, e.byteOrder)
}

func (e *UTF16Encoding) Name() string {
	if e.byteOrder == binary.BigEndian {
		return "UTF-16, big-endian"
	}
	return "UTF-16, little-endian"
}

// String implements the fmt.Stringer interface.
func (e *UTF16Encoding) String() string {
	return e.Name() + " Encoding"
}
