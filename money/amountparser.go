package money

import "github.com/domonda/go-types/language"

// AmountParser implements the strfmt.Parser interface for money amounts.
type AmountParser struct {
	acceptedDecimals []int
}

// NewAmountParser returns an AmountParser that only accepts amounts
// with the given numbers of decimal digits.
// If no acceptedDecimals are passed, any decimal digit count is accepted.
func NewAmountParser(acceptedDecimals ...int) *AmountParser {
	return &AmountParser{acceptedDecimals}
}

// Parse parses str as an Amount and returns it formatted with a dot as decimal
// separator and the maximum number of accepted decimal digits.
// langHints are accepted but not used for amount parsing.
// Parse implements the strfmt.Parser interface.
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
