package bank

import "github.com/domonda/go-types/language"

// IBANParser implements the strfmt.Parser interface for IBANs.
type IBANParser struct{}

func (IBANParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	iban, err := NormalizeIBAN(str)
	return string(iban), err
}
