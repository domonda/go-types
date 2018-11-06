package money

import (
	"database/sql/driver"

	"github.com/guregu/null"
)

// NullCurrency holds a 3 character ISO 4217 alphabetic code,
// or an empty string as valid value representing NULL in SQL databases.
// The main difference between Currency and NullCurrency is:
// Currency("").Valid() == false
// NullCurrency("").Valid() == true
type NullCurrency string

// GetOrDefault returns the value c references if it is valid and c is not nil.
// Safe to call on a nil pointer.
func (c *NullCurrency) GetOrDefault(defaultVal NullCurrency) NullCurrency {
	if !c.ValidPtr() {
		return defaultVal
	}
	return *c
}

// Valid returns true if c is an empty string, or a valid 3 character ISO 4217 alphabetic code.
// Safe to call on a nil pointer.
func (c NullCurrency) Valid() bool {
	return c == "" || Currency(c).Valid()
}

// Valid returns true if c is nil, an empty string, or a valid 3 character ISO 4217 alphabetic code.
// Safe to call on a nil pointer.
func (c *NullCurrency) ValidPtr() bool {
	if c == nil || *c == "" {
		return true
	}
	return Currency(*c).Valid()
}

// Normalized normalizes a currency string
func (c NullCurrency) Normalized() (NullCurrency, error) {
	if c == "" {
		return c, nil
	}
	norm, err := Currency(c).Normalized()
	return NullCurrency(norm), err
}

// Scan implements the database/sql.Scanner interface.
func (c *NullCurrency) Scan(value interface{}) error {
	var ns null.String
	err := ns.Scan(value)
	if err != nil {
		return err
	}
	if ns.Valid {
		*c = NullCurrency(ns.String)
	} else {
		*c = ""
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (c NullCurrency) Value() (driver.Value, error) {
	if c == "" {
		return nil, nil
	}
	return string(c), nil
}

// Symbol returns the currency symbol like € for EUR if available,
// or currency code if no widely recognized symbol is available.
func (c NullCurrency) Symbol() string {
	if s, ok := currencyCodeToSymbol[Currency(c)]; ok {
		return s
	}
	return string(c)
}

// EnglishName returns the english name of the currency
func (c NullCurrency) EnglishName() string {
	return currencyCodeToName[Currency(c)]
}
