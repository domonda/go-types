package money

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

// StringIsCurrency returns if a string can be parsed as Currency.
func StringIsCurrency(str string) bool {
	_, err := NormalizeCurrency(str)
	return err == nil
}

// Currency is holds a 3 character ISO 4217 alphabetic code.
// Currency implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty Currency string as SQL NULL value.
// The main difference between Currency and NullableCurrency is:
// Currency("").Valid() == false
// NullableCurrency("").Valid() == true
type Currency string

// NormalizeCurrency returns str as normalized Currency or an error.
func NormalizeCurrency(str string) (Currency, error) {
	return Currency(str).Normalized()
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for wasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (c *Currency) ScanString(source string) (wasNormalized bool, err error) {
	newC, err := Currency(source).Normalized()
	if err != nil {
		return false, err
	}
	*c = newC
	return newC == Currency(source), nil
}

// GetOrDefault returns the value c references if it is valid and c is not nil.
// Safe to call on a nil pointer.
func (c *Currency) GetOrDefault(defaultVal Currency) Currency {
	if !c.ValidPtr() {
		return defaultVal
	}
	return *c
}

// NullableCurrency returns c as NullableCurrency where
// NullableCurrency("").Valid() == true
func (c Currency) NullableCurrency() NullableCurrency {
	return NullableCurrency(c)
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

// NormalizedOrNull returns the normalized currency or CurrencyNull
// if it was not a valid currency.
func (c Currency) NormalizedOrNull() NullableCurrency {
	norm, err := c.Normalized()
	if err != nil {
		return CurrencyNull
	}
	return NullableCurrency(norm)
}

// IsEUR returns if the currency can be normalized as Euro.
func (c Currency) IsEUR() bool {
	norm, _ := c.Normalized()
	return norm == EUR
}

// Scan implements the database/sql.Scanner interface.
func (c *Currency) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*c = Currency(x)
	case []byte:
		*c = Currency(x)
	case nil:
		*c = CurrencyNull
	default:
		return fmt.Errorf("can't scan SQL value of type %T as Currency", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (c Currency) Value() (driver.Value, error) {
	if c == CurrencyNull {
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

// String returns the normalized currency as string if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (c Currency) String() string {
	norm, err := c.Normalized()
	if err != nil {
		return string(c)
	}
	return string(norm)
}
