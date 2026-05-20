package date

import "github.com/domonda/go-types/language"

// Formatter is a date layout string (compatible with time.Time.Format) that can
// format Date values and parse date strings. It implements the strfmt.Formatter
// interface so it can be used with generic formatting/parsing pipelines.
type Formatter string

// Format formats date using f as the layout string (see time.Time.Format).
// Returns an empty string if f or date are empty.
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
