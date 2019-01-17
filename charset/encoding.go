package charset

import (
	"bytes"
	"strings"
	"sync"

	"golang.org/x/text/encoding"

	"github.com/domonda/errors"
)

// Encoding provides threadsafe methods for encoding and decoding text
type Encoding interface {
	Encode(utf8Str []byte) (encodedStr []byte, err error)
	Decode(encodedStr []byte) (utf8Str []byte, err error)
	Name() string
	String() string
}

func GetEncoding(name string) (Encoding, error) {
	if n := strings.ToUpper(name); n == "UTF-8" || n == "UTF8" {
		return UTF8Encoding{}, nil
	}
	enc, name := findEncoding(name)
	if enc == nil {
		return nil, errors.Errorf("encoding not found: '%s'", name)
	}
	return &encodingImpl{name: name, encoding: enc}, nil
}

func MustGetEncoding(name string) Encoding {
	enc, err := GetEncoding(name)
	if err != nil {
		panic(err)
	}
	return enc
}

func AutoDecode(data []byte, keyWords []string) (str []byte, enc string, err error) {
	bom, data := SplitBOM(data)
	if bom != NoBOM {
		str, err = bom.Decode(data)
		if err != nil {
			return nil, "", err
		}
		return str, bom.String(), nil
	}

	var (
		iso8859   = MustGetEncoding("ISO 8859-1")
		macintosh = MustGetEncoding("Macintosh")
	)

	utf8Score := 0
	iso8859Score := 0
	macintoshScore := 0

	iso8859Bytes, _ := iso8859.Decode(data)
	macintoshBytes, _ := macintosh.Decode(data)

	for _, keyWord := range keyWords {
		key := []byte(keyWord)
		utf8Score += bytes.Count(data, key)
		iso8859Score += bytes.Count(iso8859Bytes, key)
		macintoshScore += bytes.Count(macintoshBytes, key)
	}

	// t.Log(docCSV, accountConfig.ConfigName, utf8Score, iso8859Score, macintoshScore)

	switch {
	case iso8859Score > 0 && iso8859Score > utf8Score && iso8859Score > macintoshScore:
		data = iso8859Bytes
		enc = "ISO 8859-1"

	case macintoshScore > 0 && macintoshScore > utf8Score && macintoshScore > iso8859Score:
		data = macintoshBytes
		enc = "Macintosh"

	default:
		enc = "UTF-8"
	}
	return data, enc, nil
}

type encodingImpl struct {
	name       string
	encoding   encoding.Encoding
	encoder    *encoding.Encoder
	encoderMtx sync.Mutex
	decoder    *encoding.Decoder
	decoderMtx sync.Mutex
}

func (e *encodingImpl) Encode(utf8Str []byte) (encodedStr []byte, err error) {
	e.encoderMtx.Lock()
	defer e.encoderMtx.Unlock()

	if e.encoder == nil {
		e.encoder = e.encoding.NewEncoder()
	}
	return e.encoder.Bytes(utf8Str)
}

func (e *encodingImpl) Decode(encodedStr []byte) (utf8Str []byte, err error) {
	e.decoderMtx.Lock()
	defer e.decoderMtx.Unlock()

	if e.decoder == nil {
		e.decoder = e.encoding.NewDecoder()
	}
	return e.decoder.Bytes(utf8Str)
}

func (e *encodingImpl) Name() string {
	return e.name
}

func (e *encodingImpl) String() string {
	return e.Name() + " Encoding"
}
