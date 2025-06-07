package bank

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types/country"
)

// IBANNull is an empty string and will be treatet as SQL NULL.
const IBANNull NullableIBAN = ""

// NullableIBAN is a IBAN value which can hold an emtpy string ("") as the null value.
type NullableIBAN string

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// If validate is true, the source string is checked
// for validity before it is assigned to the type.
//
// If validate is false and the source string
// can still be assigned in some non-normalized way
// it will be assigned without returning an error.
func (iban *NullableIBAN) ScanString(source string, validate bool) error {
	switch source {
	case "", "NULL", "null", "nil":
		iban.SetNull()
		return nil
	}
	newIBAN, err := NullableIBAN(source).Normalized()
	if err != nil {
		if validate {
			return err
		}
		newIBAN = NullableIBAN(source)
	}
	*iban = newIBAN
	return nil
}

// Valid returns true if iban is null or a valid International Bank Account Number
func (iban NullableIBAN) Valid() bool {
	return iban.Validate() == nil
}

// ValidAndNotNull returns true if iban is not null and a valid International Bank Account Number
func (iban NullableIBAN) ValidAndNotNull() bool {
	return iban.IsNotNull() && iban.Valid()
}

// Validate returns an error if this is not null and not a valid International Bank Account Number
func (iban NullableIBAN) Validate() error {
	_, err := iban.Normalized()
	return err
}

func (iban NullableIBAN) ValidAndNormalized() bool {
	norm, err := iban.Normalized()
	return err == nil && iban == norm
}

// CountryCode returns the country code of the IBAN
func (iban NullableIBAN) CountryCode() country.Code {
	if iban.IsNull() || !iban.Valid() {
		return ""
	}
	return country.Code(iban[:2])
}

// Normalized returns the iban in normalized form,
// or an error if the format can't be detected.
// Returns the NullableIBAN unchanged in case of an error.
func (iban NullableIBAN) Normalized() (NullableIBAN, error) {
	if iban.IsNull() {
		return iban, nil
	}
	normalized, err := IBAN(iban).Normalized()
	if err != nil {
		return iban, err
	}
	return NullableIBAN(normalized), nil
}

func (iban NullableIBAN) NormalizedOrNull() NullableIBAN {
	normalized, err := iban.Normalized()
	if err != nil {
		return IBANNull
	}
	return normalized
}

// NormalizedWithSpaces returns the iban in normalized form with spaces every 4 characters,
// or an error if the format can't be detected.
// Returns the NullableIBAN unchanged in case of an error.
func (iban NullableIBAN) NormalizedWithSpaces() (NullableIBAN, error) {
	if iban.IsNull() {
		return iban, nil
	}
	normalized, err := IBAN(iban).NormalizedWithSpaces()
	if err != nil {
		return iban, err
	}
	return NullableIBAN(normalized), nil
}

// Scan implements the database/sql.Scanner interface.
func (iban *NullableIBAN) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*iban = NullableIBAN(x)
	case []byte:
		*iban = NullableIBAN(x)
	case nil:
		*iban = IBANNull
	default:
		return fmt.Errorf("can't scan SQL value of type %T as NullableIBAN", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (iban NullableIBAN) Value() (driver.Value, error) {
	if iban.IsNull() {
		return nil, nil
	}
	return string(iban), nil
}

// Set sets an IBAN for this NullableIBAN
func (iban *NullableIBAN) Set(id IBAN) {
	*iban = NullableIBAN(id)
}

// SetNull sets the NullableIBAN to null
func (iban *NullableIBAN) SetNull() {
	*iban = IBANNull
}

// Get returns the non nullable IBAN value
// or panics if the NullableIBAN is null.
// Note: check with IsNull before using Get!
func (iban NullableIBAN) Get() IBAN {
	if iban.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", iban))
	}
	return IBAN(iban)
}

// GetOr returns the non nullable IBAN value
// or the passed defaultIBAN if the NullableIBAN is null.
func (iban NullableIBAN) GetOr(defaultIBAN IBAN) IBAN {
	if iban.IsNull() {
		return defaultIBAN
	}
	return IBAN(iban)
}

// StringOr returns the NullableIBAN as string
// or the passed defaultString if the NullableIBAN is null.
func (iban NullableIBAN) StringOr(defaultString string) string {
	if iban.IsNull() {
		return defaultString
	}
	return string(iban)
}

// IsNull returns true if the NullableIBAN is null.
// IsNull implements the nullable.Nullable interface.
func (iban NullableIBAN) IsNull() bool {
	return iban == IBANNull
}

func (iban NullableIBAN) IsNotNull() bool {
	return iban != IBANNull
}

// String returns the normalized IBAN string if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (iban NullableIBAN) String() string {
	norm, err := iban.Normalized()
	if err != nil {
		return string(iban)
	}
	return string(norm)
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value for an empty (null) string.
func (iban NullableIBAN) MarshalJSON() ([]byte, error) {
	if iban.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(string(iban))
}

func (NullableIBAN) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable IBAN",
		OneOf: []*jsonschema.Schema{
			{
				Type:    "string",
				Pattern: IBANRegex,
			},
			{Type: "null"},
		},
		Default: IBANNull,
	}
}
