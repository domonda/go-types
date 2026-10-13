package charset

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf8"

	"golang.org/x/text/encoding/unicode/utf32"
)

// UTF32Encoding returns an UTF-32 Encoding with the passed binary.ByteOrder.
// Its Decode method handles a leading BOM like DecodeUTF32.
func UTF32Encoding(byteOrder binary.ByteOrder) Encoding {
	return utf32Encoding{byteOrder}
}

// utf32Encoding implements Encoding for UTF-32
type utf32Encoding struct {
	byteOrder binary.ByteOrder
}

func (e utf32Encoding) Encode(utf8Str []byte) (encodedStr []byte, err error) {
	return EncodeUTF32(utf8Str, e.byteOrder)
}

func (e utf32Encoding) Decode(encodedStr []byte) (utf8Str []byte, err error) {
	return DecodeUTF32(encodedStr, e.byteOrder)
}

func (e utf32Encoding) Name() string {
	if e.byteOrder == binary.BigEndian {
		return "UTF-32BE"
	}
	return "UTF-32LE"
}

func (e utf32Encoding) String() string {
	return e.Name() + " Encoding"
}

func (e utf32Encoding) BOM() BOM {
	if e.byteOrder == binary.BigEndian {
		return BOMUTF32BE
	}
	return BOMUTF32LE
}

// DecodeUTF32 decodes the UTF-32 encoded bytes b using byteOrder and returns
// the equivalent UTF-8 bytes. Like DecodeUTF16, an optional BOM at the start
// of b is validated against byteOrder and then stripped before decoding.
// Returns an error if b starts with another BOM whose bytes are not a valid
// UTF-32 character, the UTF-16LE BOM FF FE starts valid UTF-32LE text
// like U+FFEFF.
func DecodeUTF32(b []byte, byteOrder binary.ByteOrder) ([]byte, error) {
	if len(b) == 0 {
		return nil, nil
	}
	endian := utf32.LittleEndian
	expected := BOMUTF32LE
	var order binary.ByteOrder = binary.LittleEndian
	if byteOrder == binary.BigEndian {
		endian = utf32.BigEndian
		expected = BOMUTF32BE
		order = binary.BigEndian
	}
	// Match the expected BOM before falling back to BOMOfBytes like DecodeUTF16.
	// Another BOM is only evidence of a wrong encoding if its bytes don't start
	// with a valid UTF-32 character, the UTF-16LE BOM FF FE starts valid
	// UTF-32LE text like U+FFEFF.
	if bytes.HasPrefix(b, []byte(expected)) {
		b = b[len(expected):]
	} else if bom := BOMOfBytes(b); bom != NoBOM && (len(b) < 4 || !utf8.ValidRune(rune(order.Uint32(b)))) { //#nosec G115 -- integer conversion OK, values beyond int32 are invalid runes either way
		return nil, fmt.Errorf("expected %s BOM but got %s", expected, bom)
	}
	return utf32.UTF32(endian, utf32.IgnoreBOM).NewDecoder().Bytes(b)
}

// EncodeUTF32 encodes the UTF-8 bytes b to UTF-32 using byteOrder and returns
// the encoded bytes without a leading BOM.
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

// DecodeUTF32String decodes the UTF-32 encoded bytes b using byteOrder and returns
// the result as a UTF-8 string. Like DecodeUTF32, an optional BOM at the start
// of b is validated against byteOrder and then stripped before decoding.
// Returns an error if b contains an unexpected BOM.
func DecodeUTF32String(b []byte, byteOrder binary.ByteOrder) (string, error) {
	result, err := DecodeUTF32(b, byteOrder)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
