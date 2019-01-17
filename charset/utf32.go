package charset

import (
	"encoding/binary"

	"golang.org/x/text/encoding/unicode/utf32"
)

func DecodeUTF32(b []byte, order binary.ByteOrder) ([]byte, error) {
	endian := utf32.LittleEndian
	if order == binary.BigEndian {
		endian = utf32.BigEndian
	}
	result, err := utf32.UTF32(endian, utf32.IgnoreBOM).NewDecoder().Bytes(b)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func DecodeUTF32String(b []byte, order binary.ByteOrder) (string, error) {
	result, err := DecodeUTF32(b, order)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
