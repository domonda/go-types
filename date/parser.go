package date

import "github.com/domonda/go-types/language"

// Parser implements the strfmt.Parser interface for dates.
type Parser struct{}

// Parse parses a string as a date and returns the normalized string representation.
func (Parser) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	date, err := Normalize(str, langHints...)
	return string(date), err
}
