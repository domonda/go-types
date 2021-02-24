package bank

import (
	"database/sql/driver"
	"fmt"
)

// BICNull is an empty string and will be treatet as SQL NULL.
const BICNull NullableBIC = ""

// NullableBIC is a BIC value which can hold an emtpy string ("") as the null value.
type NullableBIC string

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (bic *NullableBIC) ScanString(source string) (normalized bool, err error) {
	err = NullableBIC(source).Validate()
	if err != nil {
		return false, err
	}
	*bic = NullableBIC(source)
	return true, nil
}

// Valid returns if this is a valid SWIFT Business Identifier Code
func (bic NullableBIC) Valid() bool {
	return bic.Validate() == nil
}

// Validate returns an error if this is not a valid SWIFT Business Identifier Code
func (bic NullableBIC) Validate() error {
	if bic == BICNull {
		return nil
	}
	return BIC(bic).Validate()
}

// Scan implements the database/sql.Scanner interface.
func (bic *NullableBIC) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*bic = NullableBIC(x)
	case []byte:
		*bic = NullableBIC(x)
	case nil:
		*bic = BICNull
	default:
		return fmt.Errorf("can't scan SQL value of type %T as BIC", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (bic NullableBIC) Value() (driver.Value, error) {
	if bic == BICNull {
		return nil, nil
	}
	return string(bic), nil
}

// Set sets an BIC for this NullableBIC
func (bic *NullableBIC) Set(id BIC) {
	*bic = NullableBIC(id)
}

// SetNull sets the NullableBIC to null
func (bic *NullableBIC) SetNull() {
	*bic = BICNull
}

// Get returns the non nullable BIC value
// or panics if the NullableBIC is null.
// Note: check with IsNull before using Get!
func (bic *NullableBIC) Get() BIC {
	if bic.IsNull() {
		panic("NULL bank.BIC")
	}
	return BIC(*bic)
}

// IsNull returns true if the NullableBIC is null.
// IsNull implements the nullable.Nullable interface.
func (bic NullableBIC) IsNull() bool {
	return bic == BICNull
}

func (bic NullableBIC) IsNotNull() bool {
	return bic != BICNull
}
