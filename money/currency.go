package money

import (
	"database/sql/driver"
	"strings"

	"github.com/domonda/errors"
	"github.com/guregu/null"
)

// StringIsCurrency returns if a string can be parsed as Currency.
func StringIsCurrency(str string) bool {
	_, err := NormalizeCurrency(str)
	return err == nil
}

// Currency is holds a 3 character ISO 4217 alphabetic code.
// The main difference between Currency and NullCurrency is:
// Currency("").Valid() == false
// NullCurrency("").Valid() == true
type Currency string

// NormalizeCurrency returns str as normalized Currency or an error.
func NormalizeCurrency(str string) (Currency, error) {
	return Currency(str).Normalized()
}

// AssignString implements strfmt.StringAssignable
func (c *Currency) AssignString(str string) error {
	normalized, err := Currency(str).Normalized()
	if err != nil {
		return err
	}
	*c = normalized
	return nil
}

// GetOrDefault returns the value c references if it is valid and c is not nil.
// Safe to call on a nil pointer.
func (c *Currency) GetOrDefault(defaultVal Currency) Currency {
	if !c.ValidPtr() {
		return defaultVal
	}
	return *c
}

// Valid returns true if c is a valid 3 character ISO 4217 alphabetic code.
func (c Currency) Valid() bool {
	_, valid := currencyCodeToName[c]
	return valid
}

// ValidPtr returns if c is not nil and references a valid currency.
// Safe to call on a nil pointer.
func (c *Currency) ValidPtr() bool {
	if c == nil {
		return false
	}
	_, valid := currencyCodeToName[*c]
	return valid
}

// Normalized normalizes a currency string
func (c Currency) Normalized() (Currency, error) {
	str := strings.TrimSpace(string(c))

	result, found := currencySymbolToCode[str]
	if found {
		return result, nil
	}

	str = strings.ToUpper(str)

	if len(str) > 3 {
		switch {
		case strings.Contains(str, "EUR"):
			return "EUR", nil
		case strings.Contains(str, "SWISS") || strings.Contains(str, "SCHWEITZ") || strings.Contains(str, "FRANC"):
			return "CHF", nil
		case strings.Contains(str, "POUND") || strings.Contains(str, "BRITISH") || strings.Contains(str, "STERLING"):
			return "GBP", nil
		case strings.Contains(str, "U.S.") || strings.Contains(str, "DOLLAR"):
			return "USD", nil
		}
	}

	if _, ok := currencyCodeToName[Currency(str)]; !ok {
		return "", errors.New("invalid currency")
	}

	return Currency(str), nil
}

// Scan implements the database/sql.Scanner interface.
func (c *Currency) Scan(value interface{}) error {
	var ns null.String
	err := ns.Scan(value)
	if err != nil {
		return err
	}
	if ns.Valid {
		*c = Currency(ns.String)
	} else {
		*c = ""
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (c Currency) Value() (driver.Value, error) {
	if c == "" {
		return nil, nil
	}
	return string(c), nil
}

// Symbol returns the currency symbol like € for EUR if available,
// or currency code if no widely recognized symbol is available.
func (c Currency) Symbol() string {
	if s, ok := currencyCodeToSymbol[c]; ok {
		return s
	}
	return string(c)
}

// EnglishName returns the english name of the currency
func (c Currency) EnglishName() string {
	return currencyCodeToName[c]
}