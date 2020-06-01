package charset

import (
	"sync"

	"golang.org/x/text/encoding"
)

// Encoding provides threadsafe methods for encoding and decoding text
type Encoding interface {
	Encode(utf8Str []byte) (encodedStr []byte, err error)
	Decode(encodedStr []byte) (utf8Str []byte, err error)
	Name() string
	String() string

	// BOM returns the Unicode Byte Order Mark of the encoding
	// or NoBOM if the encoding has no BOM.
	BOM() BOM
}

type encodingImpl struct {
	name       string
	bom        BOM
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
	return e.decoder.Bytes(encodedStr)
}

func (e *encodingImpl) Name() string {
	return e.name
}

// String implements the fmt.Stringer interface.
func (e *encodingImpl) String() string {
	return e.Name() + " Encoding"
}

func (e *encodingImpl) BOM() BOM {
	return e.bom
}
