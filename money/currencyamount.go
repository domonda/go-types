package money

import (
	"fmt"
	"strings"
)

type CurrencyAmount struct {
	Currency Currency
	Amount   Amount
}

func NewCurrencyAmount(currency Currency, amount Amount) *CurrencyAmount {
	return &CurrencyAmount{Currency: currency, Amount: amount}
}

func (ca *CurrencyAmount) String() string {
	return fmt.Sprintf("%s %.2f", ca.Currency, ca.Amount)
}

func (ca *CurrencyAmount) GermanString() string {
	return strings.Replace(ca.String(), ".", ",", 1)
}

func (ca *CurrencyAmount) StringCurrencyAfterAmount() string {
	return fmt.Sprintf("%.2f %s", ca.Amount, ca.Currency)
}

func (ca *CurrencyAmount) GermanStringCurrencyAfterAmount() string {
	return strings.Replace(ca.StringCurrencyAfterAmount(), ".", ",", 1)
}

func ParseCurrencyAmount(currency, amount string, acceptInt bool) (result CurrencyAmount, err error) {
	result.Currency, err = NormalizeCurrency(currency)
	if err != nil {
		return CurrencyAmount{}, err
	}

	result.Amount, err = ParseAmount(amount, acceptInt)
	if err != nil {
		return CurrencyAmount{}, err
	}

	return result, nil
}
