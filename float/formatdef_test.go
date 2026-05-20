package float

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewFormatDef(t *testing.T) {
	ff := NewFormatDef()
	assert.Equal(t, rune(0), ff.ThousandsSep)
	assert.Equal(t, rune('.'), ff.DecimalSep)
	assert.Equal(t, -1, ff.Precision)
	assert.False(t, ff.PadPrecision)
}

func Test_FormatDef_Format(t *testing.T) {
	ff := &FormatDef{ThousandsSep: '.', DecimalSep: ',', Precision: 2, PadPrecision: true}
	assert.Equal(t, "1.234.567,80", ff.Format(1234567.8))
	assert.Equal(t, "0,50", ff.Format(0.5))
	assert.Equal(t, "-1.234.567,80", ff.Format(-1234567.8))

	// NewFormatDef: dot decimal separator, no grouping, no precision limit.
	assert.Equal(t, "1234567.8", NewFormatDef().Format(1234567.8))
	assert.Equal(t, "0.5", NewFormatDef().Format(0.5))
}

func Test_FormatDef_Parse(t *testing.T) {
	ff := &FormatDef{ThousandsSep: ',', DecimalSep: '.', Precision: 2, PadPrecision: true}

	// Parse normalizes any recognized input into ff's own format.
	normalized, err := ff.Parse("1.234.567,89")
	assert.NoError(t, err)
	assert.Equal(t, "1,234,567.89", normalized)

	normalized, err = ff.Parse("1 234 567,8")
	assert.NoError(t, err)
	assert.Equal(t, "1,234,567.80", normalized)

	_, err = ff.Parse("not a number")
	assert.Error(t, err)
}

func Test_FormatDef_JSON(t *testing.T) {
	ff := &FormatDef{ThousandsSep: ',', DecimalSep: '.', Precision: 2, PadPrecision: true}
	data, err := json.Marshal(ff)
	assert.NoError(t, err)

	var got FormatDef
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)
	assert.Equal(t, *ff, got)
}
