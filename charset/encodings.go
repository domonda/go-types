package charset

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
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

	bom, rest := SplitBOM(data)
	if bom != NoBOM {
		text, err = bom.Decode(rest)
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
// Common variants of the passed name will be considered.
func GetEncoding(name string) (Encoding, error) {
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
			if _, other := cm.ID(); other == name {
				return &encodingImpl{name: name, encoding: e}, nil
			}
		}
	}
	if e, ok := sharedEncodings[nameUpper]; ok {
		return &encodingImpl{name: nameUpper, encoding: e}, nil
	}
	return nil, fmt.Errorf("encoding not found: %q", name)
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
