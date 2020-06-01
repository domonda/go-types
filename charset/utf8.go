package charset

func UTF8Encoding() Encoding {
	return utf8Encoding{}
}

// utf8Encoding passes strings through
type utf8Encoding struct{}

func (utf8Encoding) Encode(utf8Str []byte) (encodedStr []byte, err error) {
	return utf8Str, nil
}

func (utf8Encoding) Decode(encodedStr []byte) (utf8Str []byte, err error) {
	return BOMUTF8.Decode(encodedStr)
}

func (utf8Encoding) Name() string {
	return "UTF-8"
}

// String implements the fmt.Stringer interface.
func (e utf8Encoding) String() string {
	return e.Name() + " Encoding"
}

func (utf8Encoding) BOM() BOM {
	return BOMUTF8
}
