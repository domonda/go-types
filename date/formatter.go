package date

import "github.com/domonda/go-types/language"

type Formatter string

func (f Formatter) Format(date Date) string {
	return date.Format(string(f))
}

// Parse implements the strfmt.Parser interface
func (f Formatter) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	date, err := Normalize(str, langHints...)
	if err != nil {
		return "", err
	}
	return f.Format(date), nil
}
