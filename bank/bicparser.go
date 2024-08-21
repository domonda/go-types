package bank

import "github.com/domonda/go-types/language"

// BICParser implements the strfmt.Parser interface for BICs.
type BICParser struct{}

func (BICParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	// Don't normalize by removing spaces and appending "XXX" in case of a valid length of 8 charaters.
	return str, BIC(str).Validate()
}
