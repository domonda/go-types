package strfmt

import "github.com/domonda/go-types/language"

type FloatFormat struct {
	ThousandsSep byte `json:"thousandsSep"`
	DecimalSep   byte `json:"decimalSep"`
	Precision    int  `json:"precision"`
	PadPrecision bool `json:"padPrecision"`
}

func NewDefaultFloatFormat() *FloatFormat {
	return &FloatFormat{
		DecimalSep: '.',
		Precision:  -1,
	}
}

func (ff *FloatFormat) Format(f float64) string {
	return FormatFloat(f, ff.ThousandsSep, ff.DecimalSep, ff.Precision, ff.PadPrecision)
}

// Parse implements the Parser interface
func (ff *FloatFormat) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	f, err := ParseFloat(str)
	if err != nil {
		return "", err
	}
	return ff.Format(f), nil
}
