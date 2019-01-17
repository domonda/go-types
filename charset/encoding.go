package charset

import (
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
