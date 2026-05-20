package bank

import "github.com/domonda/go-types/language"

// BICParser implements the strfmt.Parser interface for BICs.
type BICParser struct{}

// Parse validates str as a BIC and returns it unchanged if valid, or an error if not.
// Unlike full normalization, it does not remove spaces or append "XXX" for 8-character BICs.
func (BICParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	// Don't normalize by removing spaces and appending "XXX" in case of a valid length of 8 charaters.
	return str, BIC(str).Validate()
}
