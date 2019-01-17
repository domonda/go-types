package charset

import (
	"bytes"
	"encoding/binary"
	"errors"
	"unicode/utf16"
)

func decodeUTF16Runes(b []byte, order binary.ByteOrder) []rune {
	numRunes := len(b) / 2
	u16s := make([]uint16, numRunes)
	for i := 0; i < numRunes; i++ {
		u16s[i] = order.Uint16(b[i*2:])
	}
	return utf16.Decode(u16s)
}

func DecodeUTF16(b []byte, order binary.ByteOrder) ([]byte, error) {
	if len(b)&1 != 0 {
		return nil, errors.New("odd length of UTF-16 string")
	}
	runes := decodeUTF16Runes(b, order)
	buf := bytes.Buffer{}
	buf.Grow(len(runes))
	for _, r := range runes {
		buf.WriteRune(r)
	}
	return buf.Bytes(), nil
}

func DecodeUTF16String(b []byte, order binary.ByteOrder) (string, error) {
	if len(b)&1 != 0 {
		return "", errors.New("odd length of UTF-16 string")
	}
	return string(decodeUTF16Runes(b, order)), nil
}
