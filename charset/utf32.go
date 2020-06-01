package charset

import (
	"encoding/binary"

	"golang.org/x/text/encoding/unicode/utf32"
)

// UTF32Encoding returns an UTF-32 Encoding with the passed binary.ByteOrder
func UTF32Encoding(byteOrder binary.ByteOrder) Encoding {
	name := "UTF-32LE"
	endian := utf32.LittleEndian
	bom := BOMUTF32LE
	if byteOrder == binary.BigEndian {
		name = "UTF-32BE"
		endian = utf32.BigEndian
		bom = BOMUTF32BE
	}
	return &encodingImpl{
		name:    name,
		bom:     bom,
		encoder: utf32.UTF32(endian, utf32.IgnoreBOM).NewEncoder(),
		decoder: utf32.UTF32(endian, utf32.IgnoreBOM).NewDecoder(),
	}
}

func DecodeUTF32(b []byte, byteOrder binary.ByteOrder) ([]byte, error) {
	if len(b) == 0 {
		return nil, nil
	}
	endian := utf32.LittleEndian
	if byteOrder == binary.BigEndian {
		endian = utf32.BigEndian
	}
	return utf32.UTF32(endian, utf32.IgnoreBOM).NewDecoder().Bytes(b)
}

func EncodeUTF32(b []byte, byteOrder binary.ByteOrder) ([]byte, error) {
	if len(b) == 0 {
		return nil, nil
	}
	endian := utf32.LittleEndian
	if byteOrder == binary.BigEndian {
		endian = utf32.BigEndian
	}
	return utf32.UTF32(endian, utf32.IgnoreBOM).NewEncoder().Bytes(b)
}

func DecodeUTF32String(b []byte, byteOrder binary.ByteOrder) (string, error) {
	result, err := DecodeUTF32(b, byteOrder)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
