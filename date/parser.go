// Package date provides comprehensive date handling and validation utilities
// for Go applications with support for multiple date formats and internationalization.
//
// The package includes:
// - Date type with ISO 8601 format (YYYY-MM-DD) support
// - Flexible date parsing with language hints
// - Date arithmetic and comparison operations
// - Period range calculations (year, quarter, month, week)
// - Database integration (Scanner/Valuer interfaces)
// - JSON marshalling/unmarshalling
// - Nullable date support
// - Time zone handling
package date

import "github.com/domonda/go-types/language"

// Parser implements the strfmt.Parser interface for dates.
type Parser struct{}

// Parse parses a string as a date and returns the normalized string representation.
func (Parser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	date, err := Normalize(str, langHints...)
	return string(date), err
}
