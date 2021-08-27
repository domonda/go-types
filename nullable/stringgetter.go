package nullable

// StringGetter is an interface that types that represent
// a string value or null can implement.
type StringGetter interface {
	// IsNull indicates if the implementing value represents null.
	// IsNull implements the Nullable interface.
	IsNull() bool

	// IsNotNull indicates if the implementing value represents not null.
	IsNotNull() bool

	// StringOr returns the implementing value as string
	// or the passed nullString if the implementing value represents null.
	StringOr(nullString string) string
}
