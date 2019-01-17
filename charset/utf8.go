package charset

// UTF8Encoding passes strings through
type UTF8Encoding struct{}

func (UTF8Encoding) Encode(utf8Str []byte) (encodedStr []byte, err error) {
	return utf8Str, nil
}

func (UTF8Encoding) Decode(encodedStr []byte) (utf8Str []byte, err error) {
	return encodedStr, nil
}

func (UTF8Encoding) Name() string {
	return "UTF-8"
}

func (e UTF8Encoding) String() string {
	return e.Name() + " Encoding"
}
