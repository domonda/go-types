package money

import (
	"database/sql/driver"

	"github.com/domonda/errors"
)

// NullableCurrency holds a 3 character ISO 4217 alphabetic code,
// or an empty string as valid value representing NULL in SQL databases.
// NullableCurrency implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty NullableCurrency string as SQL NULL value.
// The main difference between Currency and NullableCurrency is:
// Currency("").Valid() == false
// NullableCurrency("").Valid() == true
type NullableCurrency string

// GetOrDefault returns the value c references if it is valid and c is not nil.
// Safe to call on a nil pointer.
func (c *NullableCurrency) GetOrDefault(defaultVal NullableCurrency) NullableCurrency {
	if !c.ValidPtr() {
		return defaultVal
	}
	return *c
}

// Valid returns true if c is an empty string, or a valid 3 character ISO 4217 alphabetic code.
func (c NullableCurrency) Valid() bool {
	return c == CurrencyNull || Currency(c).Valid()
}

// ValidAndNotNull returns if the currency is valid and not Null.
func (c NullableCurrency) ValidAndNotNull() bool {
	return Currency(c).Valid()
}

// Valid returns true if c is nil, an empty string, or a valid 3 character ISO 4217 alphabetic code.
// Safe to call on a nil pointer.
func (c *NullableCurrency) ValidPtr() bool {
	if c == nil || *c == CurrencyNull {
		return true
	}
	return Currency(*c).Valid()
}

// Normalized normalizes a currency string
func (c NullableCurrency) Normalized() (NullableCurrency, error) {
	if c == CurrencyNull {
		return c, nil
	}
	norm, err := Currency(c).Normalized()
	return NullableCurrency(norm), err
}

// NormalizedOrNull returns a normalized currency or CurrencyNull
// if there was an error while normalizing.
func (c NullableCurrency) NormalizedOrNull() NullableCurrency {
	normalized, err := c.Normalized()
	if err != nil {
		return CurrencyNull
	}
	return normalized
}

// Scan implements the database/sql.Scanner interface.
func (c *NullableCurrency) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*c = NullableCurrency(x)
	case []byte:
		*c = NullableCurrency(x)
	case nil:
		*c = CurrencyNull
	default:
		return errors.Errorf("can't scan SQL value of type %T as NullableCurrency", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (c NullableCurrency) Value() (driver.Value, error) {
	if c == CurrencyNull {
		return nil, nil
	}
	return string(c), nil
}

// Symbol returns the currency symbol like € for EUR if available,
// or currency code if no widely recognized symbol is available.
func (c NullableCurrency) Symbol() string {
	if s, ok := currencyCodeToSymbol[Currency(c)]; ok {
		return s
	}
	return string(c)
}

// EnglishName returns the english name of the currency
func (c NullableCurrency) EnglishName() string {
	return currencyCodeToName[Currency(c)]
}

func (c NullableCurrency) Currency() Currency {
	return Currency(c)
}
