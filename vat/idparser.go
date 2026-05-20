package vat

import "github.com/domonda/go-types/language"

// IDParser implements the strfmt.Parser interface for VAT IDs.
type IDParser struct{}

// Parse normalizes str as a VAT ID and returns it as a string.
// It returns an error if str is not a valid VAT ID.
// The optional langHints are accepted for interface compatibility but are not used.
func (IDParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	vatID, err := NormalizeVATID(str)
	return string(vatID), err
}
