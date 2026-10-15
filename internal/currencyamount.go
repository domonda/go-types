package internal

// JoinCurrencyAmount returns the formatted amount with the currency code
// and a space before it if currencyFirst is true, else after it,
// or the amount alone for an empty currency. It is the text format of
// money.CurrencyAmount and money.CurrencyDecimalAmount, which the
// strfmt money formats write too.
func JoinCurrencyAmount(currency, amount string, currencyFirst bool) string {
	switch {
	case currency == "":
		return amount
	case currencyFirst:
		return currency + " " + amount
	default:
		return amount + " " + currency
	}
}
