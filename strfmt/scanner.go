package strfmt

import (
	"reflect"
)

type Scanner interface {
	ScanString(dest reflect.Value, str string, config *ScanConfig) error
}

type ScannerFunc func(dest reflect.Value, str string, config *ScanConfig) error

func (f ScannerFunc) ScanString(dest reflect.Value, str string, config *ScanConfig) error {
	return f(dest, str, config)
}
