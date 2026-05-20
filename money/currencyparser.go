package money

import "github.com/domonda/go-types/language"

// CurrencyParser implements the strfmt.Parser interface for dates.
type CurrencyParser struct{}

// Parse normalizes str as a currency code and returns it as a string.
// langHints are accepted but not used for currency parsing.
// Parse implements the strfmt.Parser interface.
func (CurrencyParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	currency, err := NormalizeCurrency(str)
	return string(currency), err
}
