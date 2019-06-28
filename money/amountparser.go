package money

import "github.com/domonda/go-types/language"

// AmountParser implements the strfmt.Parser interface for money amounts.
type AmountParser struct {
	acceptedDecimals []int
}

func NewAmountParser(acceptedDecimals ...int) *AmountParser {
	return &AmountParser{acceptedDecimals}
}

func (p *AmountParser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	amount, err := ParseAmount(str, p.acceptedDecimals...)
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
	return amount.Format(0, '.', decimals), nil
}
