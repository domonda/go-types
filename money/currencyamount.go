package money

import (
	"database/sql/driver"
	"fmt"
	"slices"
	"strings"

	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/strutil"
)

// CurrencyAmount combines a Currency code with a monetary Amount.
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

// String returns the currency code as stored followed by the amount
// rounded to two decimal places, or to the three of the minor unit of
// currencies like KWD, see Currency.MinorUnits, like "EUR 1234.50",
// "JPY 1234.50" or "KWD 1.234", with a point as decimal separator and
// without thousands grouping. Never fewer than two decimal places,
// because the amount of a currency without a minor unit, like JPY,
// can still have a fraction. Value writes the same format with the
// normalized currency. An amount without currency is formatted with
// two decimal places.
// String implements the fmt.Stringer interface.
func (ca CurrencyAmount) String() string {
	return ca.format(string(ca.Currency), true, 0, '.', defaultPrecision(ca.Currency.MinorUnits()))
}

// FormatSep formats the CurrencyAmount using the given separators and decimal precision.
// If currencyFirst is true the currency code is placed before the amount,
// otherwise it is placed after. The currency is normalized if valid,
// else written unchanged. See Amount.FormatSep for details about the
// separator and precision arguments.
func (ca CurrencyAmount) FormatSep(currencyFirst bool, thousandsSep, decimalSep rune, precision int) string {
	currency := internal.NormalizedOrRaw(ca.Currency, Currency.Normalized, nil)
	return ca.format(string(currency), currencyFirst, thousandsSep, decimalSep, precision)
}

func (ca CurrencyAmount) format(currency string, currencyFirst bool, thousandsSep, decimalSep rune, precision int) string {
	return internal.JoinCurrencyAmount(currency, ca.Amount.FormatSep(thousandsSep, decimalSep, precision), currencyFirst)
}

// defaultPrecision returns the decimal places that String and Value write
// for a currency with the passed minor units, see Currency.MinorUnits:
// those of the minor unit, but never fewer than two, because the amount
// of a currency without a minor unit, like JPY, can still have a fraction.
func defaultPrecision(minorUnits int) int {
	return max(2, minorUnits)
}

// GoString returns a Go syntax representation of the CurrencyAmount for debugging.
// GoString implements the fmt.GoStringer interface.
func (ca CurrencyAmount) GoString() string {
	return fmt.Sprintf("{Currency: %#v, Amount: %#v}", ca.Currency, ca.Amount)
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// Only amounts with zero or two decimal places or with the decimal places
// of the minor unit of the currency, see Currency.MinorUnits, are accepted,
// so that an ambiguous thousands grouping like "EUR 1.234" is rejected
// while "KWD 1.234" written by String and "KWD 1,234" written by the
// German strfmt money format are scanned. Like ParseAmount reads a single
// separator, both are 1.234 and not 1234 with thousands grouping, which
// String writes as "KWD 1234.000" and the strfmt money formats with
// decimals too, like "KWD 1,234.000".
//
// If validate is true, then the Amount.Valid method is checked
// to return an error if the floating point value is infinite or NaN.
func (ca *CurrencyAmount) ScanString(source string, validate bool) error {
	parsed, err := ParseCurrencyAmount(source)
	if err != nil {
		return err
	}
	// NaN and ±Inf have no decimals to check, validate decides about them
	if parsed.Amount.Valid() {
		acceptedDecimals := []int{0, 2}
		if minorUnits := parsed.Currency.MinorUnits(); !slices.Contains(acceptedDecimals, minorUnits) {
			acceptedDecimals = append(acceptedDecimals, minorUnits)
		}
		parsed, err = ParseCurrencyAmount(source, acceptedDecimals...)
		if err != nil {
			return err
		}
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
// by writing the format of String with the normalized currency,
// or returning the error of Currency.Normalized for a currency
// that can't be normalized.
func (ca CurrencyAmount) Value() (driver.Value, error) {
	currency, err := ca.Currency.normalizedForValue()
	if err != nil {
		return nil, err
	}
	// MinorUnits would normalize the normalized currency again
	return ca.format(string(currency), true, 0, '.', defaultPrecision(currency.minorUnitsOfNormalized())), nil
}
