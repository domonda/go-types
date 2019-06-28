package bank

import "github.com/domonda/go-types/language"

// BICParser implements the strfmt.Parser interface for BICs.
type BICParser struct{}

func (BICParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	bic, err := ValidateBIC(str)
	return string(bic), err
}
