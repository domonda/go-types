package charset

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/unicode"
)

// AutoDecode tries to automatically decode the passed data as text.
// If data begins with an UTF BOM, then the BOM information will be used for decoding.
// If there is no BOM, then data will be decoded with all passed encodings
// and the passed keyWords will be counted in the error free decoded texts.
// The decoded text and encoding name will be returned for the encoding with
// the most key-word matches.
// If no key-word was found for any of the encodings,
// then data will be returned unchanged with an empty string as encoding name.
func AutoDecode(data []byte, encodings []Encoding, keyWords []string) (text []byte, encName string, err error) {
	if len(data) == 0 {
		return nil, "", nil
	}

	if bom := BOMOfBytes(data); bom != NoBOM {
		// BOM.Decode strips the one BOM itself
		text, err = bom.Decode(data)
		if err != nil {
			return nil, "", err
		}
		return text, bom.String(), nil
	}

	keyWordsBytes := make([][]byte, len(keyWords))
	for i, keyWord := range keyWords {
		keyWordsBytes[i] = []byte(keyWord)
	}

	type candidate struct {
		encoding string
		decoded  []byte
		score    int
	}
	var candidates []candidate

	for _, enc := range encodings {
		c := candidate{encoding: enc.Name()}
		c.decoded, err = enc.Decode(data)
		if err != nil {
			continue
		}
		for _, keyWord := range keyWordsBytes {
			c.score += bytes.Count(c.decoded, keyWord)
		}
		if c.score > 0 {
			candidates = append(candidates, c)
		}
	}

	if len(candidates) == 0 {
		return data, "", nil
	}

	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })

	return candidates[0].decoded, candidates[0].encoding, nil
}

// GetEncoding returns an Encoding for a name or an error.
//
// The name can be one of EncodingNames, matched case-insensitively
// and with "UTF" also accepted without the dash, or a charset label
// registered by IANA as used in HTTP and MIME headers,
// like "windows-1252", "latin1", "US-ASCII", or "UTF-16".
//
// A label of an encoding listed by EncodingNames returns the Encoding
// with that name, for example "Windows 1252" for "windows-1252".
// Other IANA charsets supported by golang.org/x/text/encoding,
// including multi-byte ones like "Shift_JIS", are wrapped as Encoding
// named by their MIME name, their BOM method returns NoBOM.
// "UTF-16" decodes with the byte order of a leading BOM
// (big endian without BOM) and encodes big endian with a BOM.
//
// An empty name returns an error.
func GetEncoding(name string) (Encoding, error) {
	if name == "" {
		return nil, errors.New("empty encoding name")
	}
	nameUpper := strings.ToUpper(name)
	if strings.HasPrefix(nameUpper, "UTF") && !strings.HasPrefix(nameUpper, "UTF-") {
		nameUpper = "UTF-" + nameUpper[3:]
	}
	if enc, ok := utfEncodings[nameUpper]; ok {
		return enc, nil
	}
	for _, e := range charmap.All {
		if cm, ok := e.(*charmap.Charmap); ok {
			if strings.ToUpper(cm.String()) == nameUpper {
				return &encodingImpl{name: cm.String(), encoding: e}, nil
			}
		}
	}
	if e, ok := sharedEncodings[nameUpper]; ok {
		return &encodingImpl{name: nameUpper, encoding: e}, nil
	}

	// Resolve charset labels registered by IANA.
	// The returned encoding is nil for charsets registered by IANA
	// that are not supported by golang.org/x/text/encoding.
	e, err := ianaindex.IANA.Encoding(nameUpper)
	if err != nil || e == nil {
		return nil, fmt.Errorf("encoding not found: %q", name)
	}
	switch e {
	case unicode.UTF8:
		return utfEncodings["UTF-8"], nil
	case unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM):
		return utfEncodings["UTF-16BE"], nil
	case unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM):
		return utfEncodings["UTF-16LE"], nil
	}
	for _, c := range charmap.All {
		if c == e {
			// The String method of the *charmap.Charmap values
			// and of the ISO-8859-6E/6I/8E/8I encodings
			// returns the name listed by EncodingNames.
			return &encodingImpl{name: fmt.Sprint(c), encoding: c}, nil
		}
	}
	mimeName, err := ianaindex.MIME.Name(e)
	if err != nil {
		return nil, fmt.Errorf("encoding not found: %q", name)
	}
	return &encodingImpl{name: mimeName, encoding: e}, nil
}

// MustGetEncoding returns an Encoding for a name or panics.
// Common variants of the passed name will be considered.
func MustGetEncoding(name string) Encoding {
	enc, err := GetEncoding(name)
	if err != nil {
		panic(err)
	}
	return enc
}

// EncodingNames returns the sorted names of all supported encodings
func EncodingNames() []string {
	var names []string
	for name := range utfEncodings {
		names = append(names, name)
	}
	for _, enc := range charmap.All {
		if cm, ok := enc.(*charmap.Charmap); ok {
			names = append(names, cm.String())
		}
	}
	for name := range sharedEncodings {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var utfEncodings = map[string]Encoding{
	"UTF-8":    UTF8Encoding(),
	"UTF-16LE": UTF16Encoding(binary.LittleEndian),
	"UTF-16BE": UTF16Encoding(binary.BigEndian),
	"UTF-32LE": UTF32Encoding(binary.LittleEndian),
	"UTF-32BE": UTF32Encoding(binary.BigEndian),
}

var sharedEncodings = map[string]encoding.Encoding{
	"ISO-8859-6E": charmap.ISO8859_6,
	"ISO-8859-6I": charmap.ISO8859_6,
	"ISO-8859-8E": charmap.ISO8859_8,
	"ISO-8859-8I": charmap.ISO8859_8,
}
