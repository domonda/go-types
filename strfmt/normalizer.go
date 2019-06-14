package strfmt

import "github.com/domonda/go-types/language"

type Normalizer interface {
	// Normalize parses str using optional language hints and
	// returns a normalized version of str or an parsing error.
	Normalize(str string, langHints ...language.Code) (normalized string, err error)
}
