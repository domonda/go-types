package float

import "github.com/domonda/go-types/language"

// FormatDef holds the definition for a float format
type FormatDef struct {
	ThousandsSep byte `json:"thousandsSep,string,omitempty"`
	DecimalSep   byte `json:"decimalSep,string"`
	Precision    int  `json:"precision"`
	PadPrecision bool `json:"padPrecision"`
}

// NewFormatDef returns a format definition with a dot
// as decimal separator and no precision limit.
func NewFormatDef() *FormatDef {
	return &FormatDef{
		DecimalSep: '.',
		Precision:  -1,
	}
}

func (ff *FormatDef) Format(f float64) string {
	return Format(f, ff.ThousandsSep, ff.DecimalSep, ff.Precision, ff.PadPrecision)
}

// Parse implements the Parser interface
func (ff *FormatDef) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	f, err := Parse(str)
	if err != nil {
		return "", err
	}
	return ff.Format(f), nil
}
