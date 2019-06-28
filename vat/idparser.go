package vat

import "github.com/domonda/go-types/language"

// IDParser implements the strfmt.Parser interface for VAT IDs.
type IDParser struct{}

func (IDParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	vatID, err := NormalizeVATID(str)
	return string(vatID), err
}
