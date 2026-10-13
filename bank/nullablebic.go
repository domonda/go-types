package bank

import (
	"database/sql/driver"
	"fmt"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// BICNull is an empty string and will be treated as SQL NULL.
const BICNull NullableBIC = ""

// NullableBIC is a BIC value which can hold an empty string ("") as the null value.
// A whitespace only NullableBIC is also null, like for country.NullableCode.
type NullableBIC string

// Compile-time check that NullableBIC implements nullable.NullSetable[BIC]
var _ nullable.NullSetable[BIC] = (*NullableBIC)(nil)

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

// Validate returns an error if this is not null
// and not a valid SWIFT Business Identifier Code
func (bic NullableBIC) Validate() error {
	if bic.IsNull() {
		return nil
	}
	return BIC(bic).Validate()
}

// ValidAndNormalized returns true if the BIC is valid and already normalized,
// or if it is the empty BICNull. A whitespace-only value is null and valid,
// but not normalized.
func (bic NullableBIC) ValidAndNormalized() bool {
	norm, err := bic.Normalized()
	return err == nil && bic == norm
}

// Normalized returns the BIC normalized to 11 characters,
// or BICNull if the value is null (empty or whitespace only), see IsNull.
// Returns an error if the BIC is non-null but invalid.
func (bic NullableBIC) Normalized() (NullableBIC, error) {
	if bic.IsNull() {
		return BICNull, nil
	}
	normalized, err := BIC(bic).Normalized()
	if err != nil {
		return bic, err
	}
	return NullableBIC(normalized), nil
}

// NormalizedOrNull returns the NullableBIC in normalized form,
// or BICNull if the value is null or normalization fails.
func (bic NullableBIC) NormalizedOrNull() NullableBIC {
	normalized, err := bic.Normalized()
	if err != nil {
		return BICNull
	}
	return normalized
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
// Returns nil for SQL NULL if the NullableBIC is null, see IsNull,
// else the normalized BIC like BIC.Value.
func (bic NullableBIC) Value() (driver.Value, error) {
	if bic.IsNull() {
		return nil, nil
	}
	return BIC(bic).Value()
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
func (bic NullableBIC) Get() BIC {
	if bic.IsNull() {
		panic(fmt.Sprintf("Get() called on NULL %T", bic))
	}
	return BIC(bic)
}

// GetOr returns the non nullable BIC value
// or the passed defaultBIC if the NullableBIC is null.
func (bic NullableBIC) GetOr(defaultBIC BIC) BIC {
	if bic.IsNull() {
		return defaultBIC
	}
	return BIC(bic)
}

// String returns the normalized NullableBIC string if possible,
// else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
func (bic NullableBIC) String() string {
	norm, err := bic.Normalized()
	if err != nil {
		return string(bic)
	}
	return string(norm)
}

// StringOr returns the normalized NullableBIC as string,
// or the passed defaultString if the NullableBIC is null
// or can't be normalized.
func (bic NullableBIC) StringOr(defaultString string) string {
	norm, err := bic.Normalized()
	if err != nil || norm.IsNull() {
		return defaultString
	}
	return string(norm)
}

// IsNull returns true if the NullableBIC is null,
// which includes whitespace only like for Normalized.
// IsNull implements the nullable.Nullable interface.
func (bic NullableBIC) IsNull() bool {
	return strutil.TrimSpace(bic) == BICNull
}

// IsNotNull returns true if the NullableBIC is not null,
// see IsNull.
func (bic NullableBIC) IsNotNull() bool {
	return !bic.IsNull()
}

// MarshalJSON implements encoding/json.Marshaler
// by returning the JSON null value if the NullableBIC is null, see IsNull,
// else the normalized BIC like BIC.MarshalJSON.
func (bic NullableBIC) MarshalJSON() ([]byte, error) {
	if bic.IsNull() {
		return []byte(`null`), nil
	}
	return BIC(bic).MarshalJSON()
}

// UnmarshalJSON implements encoding/json.Unmarshaler
// by setting the NullableBIC to null for the JSON null value,
// so that decoding null into a field that holds a value resets it.
// A JSON string is assigned unchanged, without normalization or validation,
// like encoding/json does for a string type, any other JSON value is an error.
func (bic *NullableBIC) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONStringOrNull(bic, data)
}

// JSONSchema returns the JSON schema definition for the NullableBIC type,
// accepting either a BIC string matching BICRegex or JSON null.
func (NullableBIC) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title: "Nullable BIC/SWIFT-Code",
		OneOf: []*jsonschema.Schema{
			{
				Type:    "string",
				Pattern: BICRegex,
			},
			{Type: "null"},
		},
		Default: BICNull,
	}
}
