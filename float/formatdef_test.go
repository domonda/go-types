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

// Test_FormatDef_MarshalJSON protects the human readable JSON form of the
// separators. The former `json:",string"` tag on the rune fields wrote the
// code point number as string, so ',' became "44", which nobody editing a
// stored locale configuration can read, and a hand written {"decimalSep":","}
// failed to decode.
func Test_FormatDef_MarshalJSON(t *testing.T) {
	tests := []struct {
		def  FormatDef
		want string
	}{
		{
			def:  FormatDef{ThousandsSep: ',', DecimalSep: '.', Precision: 2, PadPrecision: true},
			want: `{"thousandsSep":",","decimalSep":".","precision":2,"padPrecision":true}`,
		},
		{
			def:  FormatDef{ThousandsSep: '.', DecimalSep: ',', Precision: -1},
			want: `{"thousandsSep":".","decimalSep":",","precision":-1,"padPrecision":false}`,
		},
		{
			def:  FormatDef{ThousandsSep: '\'', DecimalSep: '.'},
			want: `{"thousandsSep":"'","decimalSep":".","precision":0,"padPrecision":false}`,
		},
		{
			def:  FormatDef{ThousandsSep: ' ', DecimalSep: ','},
			want: `{"thousandsSep":" ","decimalSep":",","precision":0,"padPrecision":false}`,
		},
		{
			// No thousands separator is omitted like before
			def:  *NewFormatDef(),
			want: `{"decimalSep":".","precision":-1,"padPrecision":false}`,
		},
		{
			def:  FormatDef{},
			want: `{"decimalSep":"","precision":0,"padPrecision":false}`,
		},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.def)
		if assert.NoError(t, err, "json.Marshal(%#v)", tt.def) {
			assert.JSONEq(t, tt.want, string(data), "json.Marshal(%#v)", tt.def)
		}
		// Through a pointer, like a *FormatDef field
		data, err = json.Marshal(&tt.def)
		if assert.NoError(t, err, "json.Marshal(&%#v)", tt.def) {
			assert.JSONEq(t, tt.want, string(data), "json.Marshal(&%#v)", tt.def)
		}

		var got FormatDef
		if assert.NoError(t, json.Unmarshal([]byte(tt.want), &got), "json.Unmarshal(%s)", tt.want) {
			assert.Equal(t, tt.def, got, "json.Unmarshal(%s)", tt.want)
		}
	}
}

// Test_FormatDef_UnmarshalJSON protects that hand written JSON with one
// character separators decodes, and that FormatDef values stored before
// the human readable form, with the separators as code point strings
// like "44" for ',', still decode to the same separators.
func Test_FormatDef_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json string
		want FormatDef
	}{
		{json: `{"decimalSep":","}`, want: FormatDef{DecimalSep: ','}},
		{json: `{"thousandsSep":"'","decimalSep":"."}`, want: FormatDef{ThousandsSep: '\'', DecimalSep: '.'}},
		{json: `{"thousandsSep":" ","decimalSep":",","precision":3}`, want: FormatDef{ThousandsSep: ' ', DecimalSep: ',', Precision: 3}},
		{json: `{"thousandsSep":"","decimalSep":"."}`, want: FormatDef{DecimalSep: '.'}},
		// Legacy code point strings
		{json: `{"thousandsSep":"44","decimalSep":"46","precision":2,"padPrecision":true}`, want: FormatDef{ThousandsSep: ',', DecimalSep: '.', Precision: 2, PadPrecision: true}},
		{json: `{"thousandsSep":"39","decimalSep":"44"}`, want: FormatDef{ThousandsSep: '\'', DecimalSep: ','}},
		{json: `{"thousandsSep":"32","decimalSep":"44"}`, want: FormatDef{ThousandsSep: ' ', DecimalSep: ','}},
		{json: `{"decimalSep":"0","precision":-1}`, want: FormatDef{Precision: -1}},
	}
	for _, tt := range tests {
		var got FormatDef
		if assert.NoError(t, json.Unmarshal([]byte(tt.json), &got), "json.Unmarshal(%s)", tt.json) {
			assert.Equal(t, tt.want, got, "json.Unmarshal(%s)", tt.json)
		}
	}

	// The legacy encoding written by the former struct tags still decodes
	legacy, err := json.Marshal(struct {
		ThousandsSep rune `json:"thousandsSep,string,omitempty"`
		DecimalSep   rune `json:"decimalSep,string"`
		Precision    int  `json:"precision"`
		PadPrecision bool `json:"padPrecision"`
	}{ThousandsSep: '.', DecimalSep: ',', Precision: 2, PadPrecision: true})
	assert.NoError(t, err)
	var got FormatDef
	if assert.NoError(t, json.Unmarshal(legacy, &got), "json.Unmarshal(%s)", legacy) {
		assert.Equal(t, FormatDef{ThousandsSep: '.', DecimalSep: ',', Precision: 2, PadPrecision: true}, got)
	}

	// Fields missing in the JSON keep their values like with encoding/json
	got = *NewFormatDef()
	assert.NoError(t, json.Unmarshal([]byte(`{"thousandsSep":","}`), &got))
	assert.Equal(t, FormatDef{ThousandsSep: ',', DecimalSep: '.', Precision: -1}, got)

	for _, invalid := range []string{
		`{"decimalSep":".,"}`,
		`{"decimalSep":"ab"}`,
		`{"decimalSep":44}`,
		`{"thousandsSep":"99999999999"}`,
		`[]`,
	} {
		var got FormatDef
		assert.Error(t, json.Unmarshal([]byte(invalid), &got), "json.Unmarshal(%s)", invalid)
	}
}

// TestFormatDef_UnmarshalJSONRejectsSeparatorsFormatPanicsOn protects
// Format from a panic: a field missing in the JSON keeps its value,
// so {"decimalSep":"."} into a FormatDef with '.' as thousandsSep gave
// the same rune for both separators, which Format panics on, like it
// panics on a precision below -1.
func TestFormatDef_UnmarshalJSONRejectsSeparatorsFormatPanicsOn(t *testing.T) {
	german := FormatDef{ThousandsSep: '.', DecimalSep: ',', Precision: 2}
	for _, j := range []string{`{"decimalSep":"."}`, `{"decimalSep":"x"}`, `{"thousandsSep":"x"}`, `{"precision":-2}`, `{"precision":-5}`} {
		ff := german
		assert.Error(t, json.Unmarshal([]byte(j), &ff), j)
		assert.Equal(t, german, ff, "unchanged on error for %s", j)
	}

	ff := german
	assert.NoError(t, json.Unmarshal([]byte(`{"thousandsSep":",","decimalSep":"."}`), &ff))
	assert.Equal(t, "1,234.5", ff.Format(1234.5))
	var zero FormatDef
	assert.NoError(t, json.Unmarshal([]byte(`{}`), &zero), "a zero FormatDef stays accepted")

	// A zero decimalSep formats a point, the thousandsSep of german
	ff = german
	assert.Error(t, json.Unmarshal([]byte(`{"decimalSep":""}`), &ff), "a zero decimalSep with '.' as thousandsSep")
	assert.Equal(t, german, ff, "unchanged on error")
}

// TestFormatDef_ZeroDecimalSep protects that a FormatDef without a decimal
// separator formats instead of panicking. The zero FormatDef, like an unset
// field of a config struct, and {"decimalSep":""}, which MarshalJSON writes
// for it and UnmarshalJSON accepts, are valid values, and the package
// function Format panics for a zero decimalSep. Format and Parse use a
// point for it alike, so that what Parse normalizes is what Format writes.
func TestFormatDef_ZeroDecimalSep(t *testing.T) {
	var zero FormatDef
	assert.Equal(t, "1235", zero.Format(1234.5), "rounded to its precision 0")
	normalized, err := zero.Parse("1.234,5")
	assert.NoError(t, err)
	assert.Equal(t, "1235", normalized)
	shortest := FormatDef{Precision: -1}
	assert.Equal(t, "1234.5", shortest.Format(1234.5))
	normalized, err = shortest.Parse("1.234,5")
	assert.NoError(t, err)
	assert.Equal(t, "1234.5", normalized)

	var fromJSON FormatDef
	assert.NoError(t, json.Unmarshal([]byte(`{"thousandsSep":",","decimalSep":"","precision":2,"padPrecision":true}`), &fromJSON))
	assert.Equal(t, rune(0), fromJSON.DecimalSep, "stays zero, so it marshals back unchanged")
	assert.Equal(t, "1,234.50", fromJSON.Format(1234.5))
	normalized, err = fromJSON.Parse("1.234,5")
	assert.NoError(t, err)
	assert.Equal(t, "1,234.50", normalized)

	assert.Panics(t, func() { Format(1234.5, 0, 0, -1, false) }, "the package function stays strict")
}
