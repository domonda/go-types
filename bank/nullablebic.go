package bank

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// BICNull is an empty string and will be treatet as SQL NULL.
const BICNull NullableBIC = ""

// NullableBIC is a BIC value which can hold an emtpy string ("") as the null value.
type NullableBIC string

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// If validate is true, the source string is checked
// for validity before it is assigned to the type.
//
// If validate is false and the source string
// can still be assigned in some non-normalized way
// it will be assigned without returning an error.
func (bic *NullableBIC) ScanString(source string, validate bool) error {
	switch source {
	case "", "NULL", "null", "nil":
		bic.SetNull()
		return nil
	}
	if validate && NullableBIC(source).Validate() != nil {
		return NullableBIC(source).Validate()
	}
	*bic = NullableBIC(source)
	return nil
}

// Valid returns true if bic is null or a valid SWIFT Business Identifier Code
func (bic NullableBIC) Valid() bool {
	return bic.Validate() == nil
}

// ValidAndNotNull returns true if bic is not null and a valid SWIFT Business Identifier Code
func (bic NullableBIC) ValidAndNotNull() bool {
	return bic.IsNotNull() && bic.Valid()
}

// Validate returns an error if this is not a valid SWIFT Business Identifier Code
func (bic NullableBIC) Validate() error {
	if bic == BICNull {
		return nil
	}
	return BIC(bic).Validate()
}

// Scan implements the database/sql.Scanner interface.
func (bic *NullableBIC) Scan(value any) error {
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

// StringOr returns the NullableBIC as string
// or the passed nullString if the NullableBIC is null.
func (bic NullableBIC) StringOr(nullString string) string {
	if bic.IsNull() {
		return nullString
	}
	return string(bic)
}

// IsNull returns true if the NullableBIC is null.
// IsNull implements the nullable.Nullable interface.
func (bic NullableBIC) IsNull() bool {
	return bic == BICNull
}

func (bic NullableBIC) IsNotNull() bool {
	return bic != BICNull
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value for an empty (null) string.
func (bic NullableBIC) MarshalJSON() ([]byte, error) {
	if bic.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(string(bic))
}
