package bank

import (
	"database/sql/driver"
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

// Valid returns if this is a valid SWIFT Business Identifier Code
func (iban NullableIBAN) Valid() bool {
	_, err := iban.Normalized()
	return err == nil
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
func (iban *NullableIBAN) Scan(value interface{}) error {
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
func (n *NullableIBAN) Set(id IBAN) {
	*n = NullableIBAN(id)
}

// SetNull sets the NullableIBAN to null
func (n *NullableIBAN) SetNull() {
	*n = IBANNull
}

// Get returns the non nullable IBAN value
// or panics if the NullableIBAN is null.
// Note: check with IsNull before using Get!
func (n *NullableIBAN) Get() IBAN {
	if n.IsNull() {
		panic("NULL bank.IBAN")
	}
	return IBAN(*n)
}

// IsNull returns true if the NullableIBAN is null.
// IsNull implements the nullable.Nullable interface.
func (n NullableIBAN) IsNull() bool {
	return n == IBANNull
}
