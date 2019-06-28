package money

import "github.com/domonda/go-types/language"

// CurrencyAmountParser implements the strfmt.Parser interface for money amounts.
type CurrencyAmountParser struct {
	acceptedDecimals []int
}

func NewCurrencyAmountParser(acceptedDecimals ...int) *CurrencyAmountParser {
	return &CurrencyAmountParser{acceptedDecimals}
}

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
	return amount.Format(true, 0, '.', decimals), nil
}
