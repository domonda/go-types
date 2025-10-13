package types

// Normalizable is a generic interface for types that can be normalized to a standard form.
// Normalization typically involves converting input data to a consistent, validated format.
// This interface is commonly implemented by types that handle user input or external data
// that may come in various formats but need to be standardized.
//
// The Normalized method should:
// - Return the normalized value in its standard form
// - Return an error if the value cannot be normalized
// - Be idempotent (calling Normalized on an already normalized value should return the same result)
type Normalizable[T any] interface {
	// Normalized returns the normalized value of type T or an error if normalization fails.
	Normalized() (T, error)
}

// NormalizableValidator is a generic interface that combines validation and normalization capabilities.
// It extends both the Validator interface and Normalizable interface, providing a comprehensive
// solution for types that need both validation and normalization functionality.
//
// This interface is particularly useful for types that:
// - Accept input in various formats
// - Need to validate the input before normalization
// - Should provide a quick check for both validity and normalization status
type NormalizableValidator[T any] interface {
	Validator
	Normalizable[T]

	// ValidAndNormalized returns true if the value is both valid and already in normalized form.
	// This is a performance optimization that avoids the need to call both Valid() and Normalized()
	// separately when checking if a value is ready for use.
	ValidAndNormalized() bool
}
