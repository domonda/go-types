package money

import (
	"database/sql/driver"
	"fmt"
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
func (n *NullableCurrency) GetOrDefault(defaultVal NullableCurrency) NullableCurrency {
	if !n.ValidPtr() {
		return defaultVal
	}
	return *n
}

// IsNull returns true if the NullableCurrency is null
func (n NullableCurrency) IsNull() bool {
	return n == CurrencyNull
}

// Valid returns true if c is an empty string, or a valid 3 character ISO 4217 alphabetic code.
func (n NullableCurrency) Valid() bool {
	return n == CurrencyNull || Currency(n).Valid()
}

// ValidAndNotNull returns if the currency is valid and not Null.
func (n NullableCurrency) ValidAndNotNull() bool {
	return Currency(n).Valid()
}

// Valid returns true if c is nil, an empty string, or a valid 3 character ISO 4217 alphabetic code.
// Safe to call on a nil pointer.
func (n *NullableCurrency) ValidPtr() bool {
	if n == nil || *n == CurrencyNull {
		return true
	}
	return Currency(*n).Valid()
}

// Normalized normalizes a currency string
func (n NullableCurrency) Normalized() (NullableCurrency, error) {
	if n == CurrencyNull {
		return n, nil
	}
	norm, err := Currency(n).Normalized()
	return NullableCurrency(norm), err
}

// NormalizedOrNull returns a normalized currency or CurrencyNull
// if there was an error while normalizing.
func (n NullableCurrency) NormalizedOrNull() NullableCurrency {
	normalized, err := n.Normalized()
	if err != nil {
		return CurrencyNull
	}
	return normalized
}

// Scan implements the database/sql.Scanner interface.
func (n *NullableCurrency) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*n = NullableCurrency(x)
	case []byte:
		*n = NullableCurrency(x)
	case nil:
		*n = CurrencyNull
	default:
		return fmt.Errorf("can't scan SQL value of type %T as NullableCurrency", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n NullableCurrency) Value() (driver.Value, error) {
	if n == CurrencyNull {
		return nil, nil
	}
	return string(n), nil
}

// Symbol returns the currency symbol like € for EUR if available,
// or currency code if no widely recognized symbol is available.
func (n NullableCurrency) Symbol() string {
	if s, ok := currencyCodeToSymbol[Currency(n)]; ok {
		return s
	}
	return string(n)
}

// EnglishName returns the english name of the currency
func (n NullableCurrency) EnglishName() string {
	return currencyCodeToName[Currency(n)]
}

func (n NullableCurrency) Currency() Currency {
	return Currency(n)
}
