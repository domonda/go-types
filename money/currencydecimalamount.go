package money

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/domonda/go-types/strutil"
)

// Implemented interfaces
var (
	_ fmt.Stringer   = CurrencyDecimalAmount{}
	_ fmt.GoStringer = CurrencyDecimalAmount{}
	_ driver.Valuer  = CurrencyDecimalAmount{}
	_ sql.Scanner    = (*CurrencyDecimalAmount)(nil)
)

// CurrencyDecimalAmount combines a Currency code with an exact fixed-point
// DecimalAmount. It is the exact counterpart of CurrencyAmount, which pairs a
// Currency with a float64-based Amount.
type CurrencyDecimalAmount struct {
	Currency Currency
	Amount   DecimalAmount
}

// NewCurrencyDecimalAmount returns a CurrencyDecimalAmount with the given
// currency and amount.
func NewCurrencyDecimalAmount(currency Currency, amount DecimalAmount) CurrencyDecimalAmount {
	return CurrencyDecimalAmount{Currency: currency, Amount: amount}
}

// CurrencyDecimalAmountUSD returns a CurrencyDecimalAmount with the amount in USD (US Dollar).
func CurrencyDecimalAmountUSD(amount DecimalAmount) CurrencyDecimalAmount {
	return CurrencyDecimalAmount{Currency: USD, Amount: amount}
}

// CurrencyDecimalAmountEUR returns a CurrencyDecimalAmount with the amount in EUR (Euro).
func CurrencyDecimalAmountEUR(amount DecimalAmount) CurrencyDecimalAmount {
	return CurrencyDecimalAmount{Currency: EUR, Amount: amount}
}

// CurrencyDecimalAmountCHF returns a CurrencyDecimalAmount with the amount in CHF (Swiss Franc).
func CurrencyDecimalAmountCHF(amount DecimalAmount) CurrencyDecimalAmount {
	return CurrencyDecimalAmount{Currency: CHF, Amount: amount}
}

// CurrencyDecimalAmountGBP returns a CurrencyDecimalAmount with the amount in GBP (British Pound).
func CurrencyDecimalAmountGBP(amount DecimalAmount) CurrencyDecimalAmount {
	return CurrencyDecimalAmount{Currency: GBP, Amount: amount}
}

// CurrencyDecimalAmountJPY returns a CurrencyDecimalAmount with the amount in JPY (Japanese Yen).
func CurrencyDecimalAmountJPY(amount DecimalAmount) CurrencyDecimalAmount {
	return CurrencyDecimalAmount{Currency: JPY, Amount: amount}
}

// ParseCurrencyDecimalAmount parses a currency and an exact amount from str with
// acceptedDecimals. If acceptedDecimals is empty, then any number of decimal
// places up to MaxDecimalAmountScale is accepted. The amount is parsed exactly
// via ParseDecimalAmount.
func ParseCurrencyDecimalAmount(str string, acceptedDecimals ...int) (result CurrencyDecimalAmount, err error) {
	str = strutil.TrimSpace(str)

	// Find first separator between currency and amount
	if pos := strings.IndexAny(str, " .,'-+0123456789"); pos != -1 {
		// Try parsing string until separator as currency
		result.Currency, err = NormalizeCurrency(str[:pos])
		if err == nil {
			// Set str to remaining amount part
			str = strings.TrimLeft(str[pos:], " ")
		} else {
			// If currency was not at string start, try from end
			pos = strings.LastIndexAny(str, " .,'-+0123456789")
			if pos != -1 && len(str)-pos > 0 {
				result.Currency, err = NormalizeCurrency(str[pos+1:])
				if err == nil {
					// Set str to remaining amount part
					str = strings.TrimRight(str[:pos+1], " ")
				}
			}
		}
	}

	result.Amount, err = ParseDecimalAmount(str, acceptedDecimals...)
	if err != nil {
		return CurrencyDecimalAmount{}, err
	}

	return result, nil
}

// String implements the fmt.Stringer interface, placing the currency before the
// amount and using the amount's own scale (unlike CurrencyAmount.String, which
// forces two decimal places).
func (ca CurrencyDecimalAmount) String() string {
	return ca.Format(true, 0, '.')
}

// Format formats the CurrencyDecimalAmount using the given separators. If
// currencyFirst is true the currency code is placed before the amount,
// otherwise it is placed after. The number of decimal places is the amount's
// own Scale; see DecimalAmount.FormatSep for the separator arguments.
func (ca CurrencyDecimalAmount) Format(currencyFirst bool, thousandsSep, decimalSep rune) string {
	amountStr := ca.Amount.FormatSep(thousandsSep, decimalSep)
	if ca.Currency == "" {
		return amountStr
	}
	if currencyFirst {
		return string(ca.Currency) + " " + amountStr
	}
	return amountStr + " " + string(ca.Currency)
}

// GoString returns a Go syntax representation of the CurrencyDecimalAmount for
// debugging. GoString implements the fmt.GoStringer interface.
func (ca CurrencyDecimalAmount) GoString() string {
	return fmt.Sprintf("{Currency: %#v, Amount: %#v}", ca.Currency, ca.Amount)
}

// CurrencyAmount returns the value as a float64-based CurrencyAmount,
// which may lose precision. See DecimalAmount.Amount.
func (ca CurrencyDecimalAmount) CurrencyAmount() CurrencyAmount {
	return CurrencyAmount{Currency: ca.Currency, Amount: ca.Amount.Amount()}
}

// ScanString tries to parse and assign the passed source string as value of the
// implementing type. If validate is true, a non-finite amount (NaN or ±Inf) is
// rejected.
func (ca *CurrencyDecimalAmount) ScanString(source string, validate bool) error {
	parsed, err := ParseCurrencyDecimalAmount(source)
	if err != nil {
		return err
	}
	if validate && !parsed.Amount.Valid() {
		return fmt.Errorf("invalid money.CurrencyDecimalAmount: %q", source)
	}
	*ca = parsed
	return nil
}

// Scan implements the database/sql.Scanner interface using
// ParseCurrencyDecimalAmount.
func (ca *CurrencyDecimalAmount) Scan(value any) (err error) {
	var parsed CurrencyDecimalAmount
	switch x := value.(type) {
	case string:
		parsed, err = ParseCurrencyDecimalAmount(x)
	case []byte:
		parsed, err = ParseCurrencyDecimalAmount(string(x))
	case float64:
		err = parsed.Amount.Scan(x)
	default:
		return fmt.Errorf("can't scan SQL value of type %T as money.CurrencyDecimalAmount", value)
	}
	if err != nil {
		return err
	}
	*ca = parsed
	return nil
}

// Value implements the database/sql/driver.Valuer interface by returning the
// result of the String method.
func (ca CurrencyDecimalAmount) Value() (driver.Value, error) {
	return ca.String(), nil
}
