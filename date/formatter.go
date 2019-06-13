package date

import "github.com/domonda/go-types/language"

type Formatter string

func (f Formatter) Format(date Date) string {
	return date.Format(string(f))
}

// Normalize implements the strfmt.Normalizer interface
func (f Formatter) Normalize(str string, langHints ...language.Code) (normalized string, err error) {
	date, err := Normalize(str, langHints...)
	if err != nil {
		return "", err
	}
	return f.Format(date), nil
}
