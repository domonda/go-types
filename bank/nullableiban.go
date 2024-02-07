package bank

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/domonda/go-types/country"
)

// IBANNull is an empty string and will be treatet as SQL NULL.
const IBANNull NullableIBAN = ""

// NullableIBAN is a IBAN value which can hold an emtpy string ("") as the null value.
type NullableIBAN string

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (iban *NullableIBAN) ScanString(source string) (normalized bool, err error) {
	newIBAN, err := NullableIBAN(source).Normalized()
	if err != nil {
		return false, err
	}
	*iban = newIBAN
	return newIBAN == NullableIBAN(source), nil
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
func (iban NullableIBAN) Normalized() (NullableIBAN, error) {
	if iban.IsNull() {
		return "", nil
	}
	normalized, err := IBAN(iban).Normalized()
	if err != nil {
		return "", err
	}
	return NullableIBAN(normalized), nil
}

// NormalizedOrUnchanged returns the iban in normalized form,
// or unchanged if the format has an error.
func (iban NullableIBAN) NormalizedOrUnchanged() NullableIBAN {
	normalized, err := iban.Normalized()
	if err != nil {
		return iban
	}
	return normalized
}

// NormalizedWithSpaces returns the iban in normalized form with spaces every 4 characters,
// or an error if the format can't be detected.
func (iban NullableIBAN) NormalizedWithSpaces() (NullableIBAN, error) {
	if iban.IsNull() {
		return "", nil
	}
	normalized, err := IBAN(iban).NormalizedWithSpaces()
	if err != nil {
		return "", err
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
func (iban *NullableIBAN) Get() IBAN {
	if iban.IsNull() {
		panic("NULL bank.IBAN")
	}
	return IBAN(*iban)
}

// StringOr returns the NullableIBAN as string
// or the passed nullString if the NullableIBAN is null.
func (iban NullableIBAN) StringOr(nullString string) string {
	if iban.IsNull() {
		return nullString
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
