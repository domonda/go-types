package float

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/domonda/go-types/language"
)

// FormatDef holds the definition for a float format.
//
// FormatDef implements encoding/json.Marshaler and encoding/json.Unmarshaler
// representing the separators as JSON strings of one character like ","
// or as empty string for a zero separator, see FormatDef.UnmarshalJSON.
// The json tags of the fields name the keys of that representation,
// but the separators are not written as numbers like their rune type.
type FormatDef struct {
	ThousandsSep rune `json:"thousandsSep,omitempty"`
	DecimalSep   rune `json:"decimalSep"`
	Precision    int  `json:"precision"`
	PadPrecision bool `json:"padPrecision"`
}

// formatDefJSON is the JSON representation of FormatDef
// with the separators as strings.
type formatDefJSON struct {
	ThousandsSep string `json:"thousandsSep,omitempty"`
	DecimalSep   string `json:"decimalSep"`
	Precision    int    `json:"precision"`
	PadPrecision bool   `json:"padPrecision"`
}

// NewFormatDef returns a format definition with a dot
// as decimal separator and no precision limit.
func NewFormatDef() *FormatDef {
	return &FormatDef{
		DecimalSep: '.',
		Precision:  -1,
	}
}

// Format formats the float64 value f using the thousands separator, decimal
// separator, precision, and padding settings defined in ff.
func (ff *FormatDef) Format(f float64) string {
	return Format(f, ff.ThousandsSep, ff.DecimalSep, ff.Precision, ff.PadPrecision)
}

// Parse implements the strfmt.Parser interface.
// It parses str with the package function Parse
// and returns the value formatted with ff.
// The langHints are ignored.
func (ff *FormatDef) Parse(str string, langHints ...language.Code) (normalized string, err error) {
	f, err := Parse(str)
	if err != nil {
		return "", err
	}
	return ff.Format(f), nil
}

// MarshalJSON implements encoding/json.Marshaler
// writing the separators as JSON strings of one character like ","
// or as empty string for a zero separator,
// where a zero ThousandsSep is omitted.
func (ff FormatDef) MarshalJSON() ([]byte, error) {
	return json.Marshal(formatDefJSON{
		ThousandsSep: separatorString(ff.ThousandsSep),
		DecimalSep:   separatorString(ff.DecimalSep),
		Precision:    ff.Precision,
		PadPrecision: ff.PadPrecision,
	})
}

// UnmarshalJSON implements encoding/json.Unmarshaler
// reading the separators from JSON strings of one character like ","
// or from an empty string for a zero separator.
// Strings of decimal digits like "44" for ',' are read as the code point
// of the separator, which is how FormatDef was written before it
// implemented encoding/json.Marshaler.
// Fields missing in the JSON object keep their value.
func (ff *FormatDef) UnmarshalJSON(j []byte) error {
	v := formatDefJSON{
		ThousandsSep: separatorString(ff.ThousandsSep),
		DecimalSep:   separatorString(ff.DecimalSep),
		Precision:    ff.Precision,
		PadPrecision: ff.PadPrecision,
	}
	err := json.Unmarshal(j, &v)
	if err != nil {
		return err
	}
	thousandsSep, err := parseSeparator(v.ThousandsSep)
	if err != nil {
		return fmt.Errorf("can't unmarshal float.FormatDef thousandsSep: %w", err)
	}
	decimalSep, err := parseSeparator(v.DecimalSep)
	if err != nil {
		return fmt.Errorf("can't unmarshal float.FormatDef decimalSep: %w", err)
	}
	*ff = FormatDef{
		ThousandsSep: thousandsSep,
		DecimalSep:   decimalSep,
		Precision:    v.Precision,
		PadPrecision: v.PadPrecision,
	}
	return nil
}

// separatorString returns sep as string or an empty string for zero.
func separatorString(sep rune) string {
	if sep == 0 {
		return ""
	}
	return string(sep)
}

// parseSeparator parses a separator written by separatorString
// or as string of the decimal code point, which no digit separator
// can be confused with, because digits are no valid separators.
func parseSeparator(str string) (rune, error) {
	if str == "" {
		return 0, nil
	}
	if strings.Trim(str, "0123456789") == "" {
		codePoint, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid separator code point %q: %w", str, err)
		}
		return rune(codePoint), nil
	}
	sep, size := utf8.DecodeRuneInString(str)
	if sep == utf8.RuneError || size != len(str) {
		return 0, fmt.Errorf("separator must be a single character: %q", str)
	}
	return sep, nil
}
