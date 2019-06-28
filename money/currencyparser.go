package money

import "github.com/domonda/go-types/language"

// CurrencyParser implements the strfmt.Parser interface for dates.
type CurrencyParser struct{}

func (CurrencyParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	currency, err := NormalizeCurrency(str)
	return string(currency), err
}
