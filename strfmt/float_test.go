package strfmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Variations with leading + and - are created automatically, don't put them here
var validDecimalFloats = map[string]float64{
	"100":              100,
	"1e6":              1e6,
	"1.2e6":            1.2e6,
	",1":               0.1,
	".1":               0.1,
	"1,":               1.0,
	"1.":               1.0,
	"123.456":          123.456,
	"123,456":          123.456,
	"100 200 300.1234": 100200300.1234,
	"100 200 300,1234": 100200300.1234,
	"100,200,300.1234": 100200300.1234,
	"100.200.300,1234": 100200300.1234,
	"1,200,300.1234":   1200300.1234,
	"1.200.300,1234":   1200300.1234,
}

var invalidDecimalFloats = []string{
	"xxx",
	"e3",
	"1ee6",
	",1,1",
	"9,1,1",
	"10.000.00,00",
}

func Test_ParseFloat(t *testing.T) {
	for s, f := range validDecimalFloats {
		parsed, err := ParseFloat(s)
		if err != nil {
			assert.NoError(t, err)
		}
		assert.Equal(t, f, parsed)

		parsed, err = ParseFloat("+" + s)
		if err != nil {
			assert.NoError(t, err)
		}
		assert.Equal(t, +f, parsed)

		parsed, err = ParseFloat("-" + s)
		if err != nil {
			assert.NoError(t, err)
		}
		assert.Equal(t, -f, parsed)
	}

	for _, s := range invalidDecimalFloats {
		_, err := ParseFloat(s)
		if err != nil {
			assert.Error(t, err)
		}
	}
}
