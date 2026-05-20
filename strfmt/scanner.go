package strfmt

import (
	"reflect"
)

// Scanner parses a string and assigns the result to a reflect.Value
// using the settings from the provided ScanConfig.
// It is the core extension point for registering custom per-type
// scanning logic in ScanConfig.TypeScanners.
type Scanner interface {
	ScanString(dest reflect.Value, str string, config *ScanConfig) error
}

// ScannerFunc is a function adapter that implements the Scanner interface,
// allowing an ordinary function to be used wherever a Scanner is expected.
type ScannerFunc func(dest reflect.Value, str string, config *ScanConfig) error

// ScanString calls f(dest, str, config), implementing the Scanner interface.
func (f ScannerFunc) ScanString(dest reflect.Value, str string, config *ScanConfig) error {
	return f(dest, str, config)
}
