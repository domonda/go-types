package utf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"unicode/utf16"
)

func DecodeUTF16Runes(b []byte, order binary.ByteOrder) ([]rune, error) {
	if len(b)%2 == 1 {
		return nil, errors.New("odd length of UTF-16 string")
	}
	u16s := make([]uint16, 0, len(b)/2)
	for i, j := 0, len(b); i < j; i += 2 {
		u16s = append(u16s, order.Uint16(b[i:]))
	}
	return utf16.Decode(u16s), nil
}

func DecodeUTF16(b []byte, order binary.ByteOrder) ([]byte, error) {
	runes, err := DecodeUTF16Runes(b, order)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.Grow(len(runes))
	for _, r := range runes {
		buf.WriteRune(r)
	}
	return buf.Bytes(), nil
}

func DecodeUTF16String(b []byte, order binary.ByteOrder) (string, error) {
	runes, err := DecodeUTF16Runes(b, order)
	if err != nil {
		return "", err
	}
	return string(runes), nil
}
