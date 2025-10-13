package money

import (
	"database/sql/driver"
	"encoding/json"
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

// IsNull returns true if the NullableCurrency is null.
// IsNull implements the nullable.Nullable interface.
func (n NullableCurrency) IsNull() bool {
	return n == CurrencyNull
}

// IsNotNull returns true if the NullableCurrency is not null.
func (n NullableCurrency) IsNotNull() bool {
	return n != CurrencyNull
}

// Set sets an ID for this NullableCurrency
func (n *NullableCurrency) Set(currency Currency) {
	*n = NullableCurrency(currency)
}

// SetNull sets the NullableCurrency to null
func (n *NullableCurrency) SetNull() {
	*n = CurrencyNull
}

// Get returns the non nullable Currency value
// or panics if the NullableCurrency is null.
// Note: check with IsNull before using Get!
func (n NullableCurrency) Get() Currency {
	if n.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", n))
	}
	return Currency(n)
}

// GetOr returns the Currency if n valid and not null,
// or else the passed defaultVal is returned.
func (n NullableCurrency) GetOr(defaultVal Currency) Currency {
	if !n.ValidAndNotNull() {
		return defaultVal
	}
	return Currency(n)
}

// StringOr returns the NullableCurrency as string
// or the passed defaultString if the NullableCurrency is null.
func (n NullableCurrency) StringOr(defaultString string) string {
	if n.IsNull() {
		return defaultString
	}
	return string(n)
}

// Valid returns true if n is an empty string, or can be normalized to a valid currency.
func (n NullableCurrency) Valid() bool {
	return n == CurrencyNull || Currency(n).Valid()
}

// Validate returns an error if n is not an empty string and
// can not be normalized to a valid currency.
func (n NullableCurrency) Validate() error {
	if n == CurrencyNull {
		return nil
	}
	return Currency(n).Validate()
}

// ValidAndNotNull returns if the currency is valid and not Null.
func (n NullableCurrency) ValidAndNotNull() bool {
	return Currency(n).Valid()
}

// ValidAndNormalized returns true if the NullableCurrency is valid and already normalized.
func (n NullableCurrency) ValidAndNormalized() bool {
	norm, err := n.Normalized()
	return err == nil && n == norm
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

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// If validate is true, the source string is checked
// for validity before it is assigned to the type.
//
// If validate is false and the source string
// can still be assigned in some non-normalized way
// it will be assigned without returning an error.
func (n *NullableCurrency) ScanString(source string, validate bool) error {
	switch source {
	case "", "NULL", "null", "nil":
		n.SetNull()
		return nil
	}
	newCurrency, err := NullableCurrency(source).Normalized()
	if err != nil {
		if validate {
			return err
		}
		newCurrency = NullableCurrency(source)
	}
	*n = newCurrency
	return nil
}

// Scan implements the database/sql.Scanner interface.
func (n *NullableCurrency) Scan(value any) error {
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

// String returns the normalized currency as string if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (n NullableCurrency) String() string {
	norm, err := n.Normalized()
	if err != nil {
		return string(n)
	}
	return string(norm)
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value for an empty (null) string.
func (n NullableCurrency) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(string(n))
}
