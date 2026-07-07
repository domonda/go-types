package money

import "github.com/domonda/go-types/language"

// CurrencyAmountParser implements the strfmt.Parser interface for money amounts.
type CurrencyAmountParser struct {
	acceptedDecimals []int
}

// NewCurrencyAmountParser returns a CurrencyAmountParser that only accepts amounts
// with the given numbers of decimal digits.
// If no acceptedDecimals are passed, any decimal digit count is accepted.
func NewCurrencyAmountParser(acceptedDecimals ...int) *CurrencyAmountParser {
	return &CurrencyAmountParser{acceptedDecimals}
}

// Parse parses str as a CurrencyAmount and returns it formatted with the currency
// code first, a dot as decimal separator, and the maximum number of accepted
// decimal digits. langHints are accepted but not used for parsing.
// Parse implements the strfmt.Parser interface.
func (p *CurrencyAmountParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	amount, err := ParseCurrencyAmount(str, p.acceptedDecimals...)
	if err != nil {
		return "", err
	}
	decimals := 2
	if len(p.acceptedDecimals) > 0 {
		decimals = -1
		for _, accepted := range p.acceptedDecimals {
			if accepted > decimals {
				decimals = accepted
			}
		}
	}
	return amount.FormatSep(true, 0, '.', decimals), nil
}
