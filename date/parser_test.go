package date

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/language"
)

// Parser must satisfy the strfmt.Parser method signature so it can be
// plugged into generic parsing pipelines (strfmt can't be imported here
// because it depends on this package).
var _ interface {
	Parse(str string, langHints ...language.Code) (string, error)
} = Parser{}

func TestParser_Parse(t *testing.T) {
	// Parser always returns the canonical ISO 8601 form,
	// whatever real-world shape the input had.
	tests := []struct {
		input string
		lang  []language.Code
		want  string
	}{
		{input: "2023-12-25", want: "2023-12-25"},
		{input: "25.12.2023", want: "2023-12-25"},
		{input: "  25.12.2023  ", want: "2023-12-25"},
		{input: "1. Dezember 2016", want: "2016-12-01"},
		{input: "September 17th, 2012", want: "2012-09-17"},
		{input: "Fri, 03 Jul 2015", want: "2015-07-03"},
		{input: "30.11.2023 00:00:00", want: "2023-11-30"},
		{input: "2 janv. 2019", want: "2019-01-02"},
		// The language hint only matters for ambiguous numeric day/month order
		{input: "05/06/2024", want: "2024-06-05"},
		{input: "05/06/2024", lang: []language.Code{language.EN}, want: "2024-05-06"},
		{input: "05/06/2024", lang: []language.Code{"en-US"}, want: "2024-05-06"},
		{input: "05/06/2024", lang: []language.Code{language.DE}, want: "2024-06-05"},
		// Unambiguous input ignores a conflicting language hint
		{input: "12/25/2023", lang: []language.Code{language.DE}, want: "2023-12-25"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parser{}.Parse(tt.input, tt.lang...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	for _, invalid := range []string{"", "2014-04", "2023-13-01", "32.01.2023", "3:28:00"} {
		t.Run("invalid "+invalid, func(t *testing.T) {
			got, err := Parser{}.Parse(invalid)
			require.Error(t, err)
			assert.Equal(t, "", got, "no partial result on error")
		})
	}
}
