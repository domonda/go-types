package strfmt

import "github.com/domonda/go-types/language"

type Parser interface {
	// Parse str using optional language hints and
	// returns a normalized version of str or an parsing error.
	Parse(str string, langHints ...language.Code) (normalized string, err error)
}
