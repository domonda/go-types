package money

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/domonda/go-types/strutil"
)

type CurrencyAmount struct {
	Currency Currency
	Amount   Amount
}

// NewCurrencyAmount returns a CurrencyAmount with the given currency and amount.
func NewCurrencyAmount(currency Currency, amount Amount) CurrencyAmount {
	return CurrencyAmount{Currency: currency, Amount: amount}
}

// CurrencyAmountUSD returns a CurrencyAmount with the amount in USD (US Dollar).
func CurrencyAmountUSD(amount Amount) CurrencyAmount {
	return CurrencyAmount{Currency: USD, Amount: amount}
}

// CurrencyAmountEUR returns a CurrencyAmount with the amount in EUR (Euro).
func CurrencyAmountEUR(amount Amount) CurrencyAmount {
	return CurrencyAmount{Currency: EUR, Amount: amount}
}

// CurrencyAmountCHF returns a CurrencyAmount with the amount in CHF (Swiss Franc).
func CurrencyAmountCHF(amount Amount) CurrencyAmount {
	return CurrencyAmount{Currency: CHF, Amount: amount}
}

// CurrencyAmountGBP returns a CurrencyAmount with the amount in GBP (British Pound).
func CurrencyAmountGBP(amount Amount) CurrencyAmount {
	return CurrencyAmount{Currency: GBP, Amount: amount}
}

// CurrencyAmountJPY returns a CurrencyAmount with the amount in JPY (Japanese Yen).
func CurrencyAmountJPY(amount Amount) CurrencyAmount {
	return CurrencyAmount{Currency: JPY, Amount: amount}
}

// ParseCurrencyAmount parses a currency and an amount from str with acceptedDecimals.
// If acceptedDecimals is empty, then any decimal number is accepted.
func ParseCurrencyAmount(str string, acceptedDecimals ...int) (result CurrencyAmount, err error) {
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

	result.Amount, err = ParseAmount(str, acceptedDecimals...)
	if err != nil {
		return CurrencyAmount{}, err
	}

	return result, nil
}

// String implements the fmt.Stringer interface.
func (ca CurrencyAmount) String() string {
	return ca.Format(true, 0, '.', 2)
}

func (ca CurrencyAmount) Format(currencyFirst bool, thousandsSep, decimalSep rune, precision int) string {
	amountStr := ca.Amount.Format(thousandsSep, decimalSep, precision)
	if ca.Currency == "" {
		return amountStr
	}
	if currencyFirst {
		return string(ca.Currency) + " " + amountStr
	}
	return amountStr + " " + string(ca.Currency)
}

func (ca CurrencyAmount) GoString() string {
	return fmt.Sprintf("{Currency: %#v, Amount: %#v}", ca.Currency, ca.Amount)
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// If validate is true, then the Amount.Valid method is checked
// to return an error if the floating point value is infinite or NaN.
func (ca *CurrencyAmount) ScanString(source string, validate bool) error {
	parsed, err := ParseCurrencyAmount(source, 0, 2)
	if err != nil {
		return err
	}
	if validate && !parsed.Amount.Valid() {
		return fmt.Errorf("invalid amount: %q", source)
	}
	*ca = parsed
	return nil
}

// Scan implements the database/sql.Scanner interface
// using ParseCurrencyAmount.
func (ca *CurrencyAmount) Scan(value any) (err error) {
	var parsed CurrencyAmount
	switch x := value.(type) {
	case string:
		parsed, err = ParseCurrencyAmount(x)
	case []byte:
		parsed, err = ParseCurrencyAmount(string(x))
	case float64:
		parsed.Amount = Amount(x)
	default:
		return fmt.Errorf("can't scan SQL value of type %T as money.CurrencyAmount", value)
	}
	if err != nil {
		return err
	}
	*ca = parsed
	return nil
}

// Value implements the database/sql/driver.Valuer interface
// by returning the result of the String method.
func (ca CurrencyAmount) Value() (driver.Value, error) {
	return ca.String(), nil
}
