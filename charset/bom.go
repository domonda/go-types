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
	// NoBOM is the zero value of BOM, indicating that no byte order mark is present.
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

// BOMOfString returns the BOM found at the beginning of str,
// or NoBOM if no byte order mark is present.
//
// See [BOMOfBytes] for the detection order and the
// UTF-16LE/UTF-32LE ambiguity that comes with it.
func BOMOfString(str string) BOM {
	switch {
	case strings.HasPrefix(str, string(BOMUTF8)):
		return BOMUTF8
	case strings.HasPrefix(str, string(BOMUTF32LE)):
		return BOMUTF32LE
	case strings.HasPrefix(str, string(BOMUTF32BE)):
		return BOMUTF32BE
	case strings.HasPrefix(str, string(BOMUTF16LE)):
		return BOMUTF16LE
	case strings.HasPrefix(str, string(BOMUTF16BE)):
		return BOMUTF16BE
	}
	return NoBOM
}

// BOMOfBytes returns the BOM found at the beginning of b,
// or NoBOM if no byte order mark is present.
//
// The 4 byte BOMs are checked before the 2 byte ones because the
// UTF-16LE BOM (FF FE) is a prefix of the UTF-32LE BOM (FF FE 00 00).
//
// Caveat: those two are not distinguishable from the bytes alone.
// UTF-16LE text whose first character is U+0000 also serializes to
// FF FE 00 00 and is therefore reported as BOMUTF32LE. Preferring the
// longer BOM is the conventional resolution, used by ICU and .NET among
// others, because a leading NUL is not plain text while real UTF-32LE
// data is. The Unicode standard lists the signatures in table 23-6 but
// does not say how to resolve the overlap.
//
// The ambiguity only exists when detecting an unknown encoding.
// Callers that already know the encoding should skip detection and pass
// the known BOM to [BOM.Decode] or [BOM.DecodeString], where FF FE can
// only mean UTF-16LE.
func BOMOfBytes(b []byte) BOM {
	switch {
	case bytes.HasPrefix(b, bomUTF8):
		return BOMUTF8
	case bytes.HasPrefix(b, bomUTF32LE):
		return BOMUTF32LE
	case bytes.HasPrefix(b, bomUTF32BE):
		return BOMUTF32BE
	case bytes.HasPrefix(b, bomUTF16LE):
		return BOMUTF16LE
	case bytes.HasPrefix(b, bomUTF16BE):
		return BOMUTF16BE
	}
	return NoBOM
}

// TrimBOM removes a leading bom byte order mark from b and returns the
// remaining bytes. If b does not start with bom, or bom is NoBOM, b is
// returned unchanged.
func TrimBOM(b []byte, bom BOM) []byte {
	if bom != NoBOM && bytes.HasPrefix(b, []byte(bom)) {
		return b[len(bom):]
	}
	return b
}

// SplitBOM detects and returns the BOM at the beginning of b together with
// the remaining bytes after the BOM. If no BOM is present, NoBOM and b are returned.
func SplitBOM(b []byte) (BOM, []byte) {
	bom := BOMOfBytes(b)
	return bom, b[len(bom):]
}

// DecodeWithBOM detects the BOM at the beginning of b and decodes the remaining
// bytes to UTF-8. Returns an error if the BOM indicates an unsupported encoding.
func DecodeWithBOM(b []byte) ([]byte, error) {
	bom, data := SplitBOM(b)
	return bom.Decode(data)
}

// DecodeStringWithBOM detects the BOM at the beginning of b and decodes the remaining
// bytes to a UTF-8 string. Returns an error if the BOM indicates an unsupported encoding.
func DecodeStringWithBOM(b []byte) (string, error) {
	bom, data := SplitBOM(b)
	return bom.DecodeString(data)
}

// trimExpectedBOM strips a leading bom from data and returns the rest.
//
// bom is what the caller already knows the encoding to be, so it is matched
// before falling back to BOMOfBytes detection. That order matters for the
// UTF-16LE/UTF-32LE overlap: the valid UTF-16LE sequence FF FE 00 00 (BOM
// followed by U+0000) is indistinguishable from the UTF-32LE BOM, and
// detection resolves it to UTF-32LE. A caller passing BOMUTF16LE has already
// resolved that ambiguity and must not be second-guessed.
//
// Returns an error if data carries a different BOM than the expected one.
func trimExpectedBOM(data []byte, bom BOM) ([]byte, error) {
	if bom != NoBOM && bytes.HasPrefix(data, []byte(bom)) {
		return data[len(bom):], nil
	}
	if dataBOM := BOMOfBytes(data); dataBOM != NoBOM && dataBOM != bom {
		return nil, fmt.Errorf("wrong BOM in data: %v, expected: %v", []byte(dataBOM), []byte(bom))
	}
	return data, nil
}

// Encoding returns the Encoding corresponding to the BOM.
// NoBOM and BOMUTF8 both map to the UTF-8 encoding.
// Returns an error for unrecognised BOM values.
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

// Decode decodes data from the encoding indicated by bom to UTF-8 bytes.
// An optional leading BOM in data is validated against bom and then stripped.
// Returns an error if the BOM does not match or the encoding is unsupported.
func (bom BOM) Decode(data []byte) ([]byte, error) {
	data, err := trimExpectedBOM(data, bom)
	if err != nil {
		return nil, err
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

// DecodeString decodes data from the encoding indicated by bom to a UTF-8 string.
// An optional leading BOM in data is validated against bom and then stripped.
// Returns an error if the BOM does not match or the encoding is unsupported.
func (bom BOM) DecodeString(data []byte) (string, error) {
	data, err := trimExpectedBOM(data, bom)
	if err != nil {
		return "", err
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

// Endian returns the binary.ByteOrder implied by the BOM:
// binary.LittleEndian for UTF-16LE/UTF-32LE, binary.BigEndian for UTF-16BE/UTF-32BE,
// and nil for NoBOM or BOMUTF8.
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
