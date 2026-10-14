package money

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"

	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// NullableCurrency holds a 3 character ISO 4217 alphabetic code,
// or an empty or whitespace only string as valid value representing NULL.
// The main difference between Currency and NullableCurrency is:
// Currency("").Valid() == false
// NullableCurrency("").Valid() == true
//
// NullableCurrency reads and writes like Currency, except that null is
// SQL NULL and JSON null. See "Reading and writing values" in the README
// of the module.
type NullableCurrency string

// Compile-time check that NullableCurrency implements nullable.NullSetable[Currency]
var _ nullable.NullSetable[Currency] = (*NullableCurrency)(nil)

// IsNull returns true if the NullableCurrency is null,
// which includes a string of only the runes strutil.TrimSpace removes:
// whitespace, control and other non-printable runes, and invalid UTF-8,
// like for country.NullableCode.
// IsNull implements the nullable.Nullable interface.
func (n NullableCurrency) IsNull() bool {
	return strutil.TrimSpace(n) == CurrencyNull
}

// IsNotNull returns true if the NullableCurrency is not null,
// see IsNull.
func (n NullableCurrency) IsNotNull() bool {
	return !n.IsNull()
}

// Set sets the passed currency as value of the NullableCurrency
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

// StringOr returns the NullableCurrency as stored,
// or the passed defaultString if the NullableCurrency is null.
func (n NullableCurrency) StringOr(defaultString string) string {
	if n.IsNull() {
		return defaultString
	}
	return string(n)
}

// Valid returns true if n is null (empty or whitespace only),
// or can be normalized to a valid currency.
func (n NullableCurrency) Valid() bool {
	return n.IsNull() || Currency(n).Valid()
}

// Validate returns an error if n is not null (empty or whitespace only)
// and can not be normalized to a valid currency.
func (n NullableCurrency) Validate() error {
	if n.IsNull() {
		return nil
	}
	return Currency(n).Validate()
}

// ValidAndNotNull returns if the currency is valid and not Null.
func (n NullableCurrency) ValidAndNotNull() bool {
	return Currency(n).Valid()
}

// ValidAndNormalized returns true if the NullableCurrency is valid and already normalized,
// or if it is the empty CurrencyNull. A whitespace-only value is null and valid,
// but not normalized.
func (n NullableCurrency) ValidAndNormalized() bool {
	norm, err := n.Normalized()
	return err == nil && n == norm
}

// ValidPtr returns true if n is nil, null (empty or whitespace only),
// or can be normalized to a valid currency.
// Safe to call on a nil pointer.
func (n *NullableCurrency) ValidPtr() bool {
	if n == nil || n.IsNull() {
		return true
	}
	return Currency(*n).Valid()
}

// Normalized normalizes a currency string.
// Returns CurrencyNull for a null (empty or whitespace only) currency, see IsNull.
// Returns the NullableCurrency unchanged in case of an error.
func (n NullableCurrency) Normalized() (NullableCurrency, error) {
	if n.IsNull() {
		return CurrencyNull, nil
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

// ScanString implements the strfmt.Scannable interface.
// The strings "NULL", "null" and "nil" set the NullableCurrency to null.
// Without validate it stores source like Scan, with validate it stores
// the normalized currency or returns the error of Normalized.
func (n *NullableCurrency) ScanString(source string, validate bool) error {
	switch source {
	case "NULL", "null", "nil":
		n.SetNull()
		return nil
	}
	return internal.ScanStringNormalized(n, source, validate, NullableCurrency.Normalized, nil)
}

// Scan implements the database/sql.Scanner interface.
// A valid currency is stored normalized, SQL NULL and a blank string as
// CurrencyNull, and a currency that can't be normalized unchanged without
// an error, so that one invalid value doesn't fail a whole query;
// Validate reports it.
func (n *NullableCurrency) Scan(value any) error {
	return internal.ScanNormalized(n, value, NullableCurrency.Normalized, nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the normalized currency, SQL NULL if the NullableCurrency
// is null, or the error of Normalized for a currency that can't be normalized.
func (n NullableCurrency) Value() (driver.Value, error) {
	return internal.ValueNormalized(n, NullableCurrency.Normalized, nil)
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

// Currency returns the NullableCurrency as a Currency value,
// regardless of whether it is null or not.
func (n NullableCurrency) Currency() Currency {
	return Currency(n)
}

// MarshalJSON implements encoding/json.Marshaler by writing the JSON null
// value if the NullableCurrency is null, else the normalized currency,
// or the currency unchanged if it can't be normalized.
func (n NullableCurrency) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte(`null`), nil
	}
	return internal.MarshalJSONNormalized(n, NullableCurrency.Normalized, nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null sets the
// NullableCurrency to null, so that decoding null into a field that holds
// a value resets it, and any other JSON value is an error.
func (n *NullableCurrency) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(n, data, true, NullableCurrency.Normalized, nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// but with "" for null, so that encoding/xml and strfmt write the
// normalized currency. encoding/json writes map keys of string types
// as they are and doesn't call MarshalText for them.
func (n NullableCurrency) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(n, NullableCurrency.Normalized, nil)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (n *NullableCurrency) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(n, d, start, NullableCurrency.Normalized, nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (n *NullableCurrency) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(n, attr, NullableCurrency.Normalized, nil)
}
