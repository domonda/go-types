package date

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/language"
)

// Formatter must satisfy the strfmt.Parser method signature so it can be
// plugged into generic parsing pipelines (strfmt can't be imported here
// because it depends on this package).
var _ interface {
	Parse(str string, langHints ...language.Code) (string, error)
} = Formatter("")

func TestFormatter_Format(t *testing.T) {
	tests := []struct {
		name   string
		layout Formatter
		date   Date
		want   string
	}{
		{name: "German", layout: "02.01.2006", date: "2023-12-25", want: "25.12.2023"},
		{name: "US", layout: "01/02/2006", date: "2023-12-25", want: "12/25/2023"},
		{name: "long English", layout: "January 2, 2006", date: "2023-03-07", want: "March 7, 2023"},
		{name: "ISO Layout", layout: Layout, date: "2023-12-25", want: "2023-12-25"},
		// Empty inputs must not be turned into "01.01.0001" from the zero time.Time
		{name: "empty date", layout: "02.01.2006", date: "", want: ""},
		{name: "empty layout", layout: "", date: "2023-12-25", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.layout.Format(tt.date))
		})
	}
}

func TestFormatter_Parse(t *testing.T) {
	// Parse normalizes any supported real-world date input
	// and re-formats it with the Formatter's layout.
	f := Formatter("02.01.2006")

	tests := []struct {
		input string
		lang  []language.Code
		want  string
	}{
		{input: "2023-12-25", want: "25.12.2023"},
		{input: "25.12.23", want: "25.12.2023"},
		{input: "Mon Jan 02 2006", want: "02.01.2006"},
		{input: "2014年04月08日", want: "08.04.2014"},
		{input: "2023-12-25T15:04:05Z", want: "25.12.2023"},
		{input: "05/06/2024", lang: []language.Code{language.EN}, want: "06.05.2024"},
		{input: "05/06/2024", lang: []language.Code{language.DE}, want: "05.06.2024"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := f.Parse(tt.input, tt.lang...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	for _, invalid := range []string{"", "2014", "31.02.2023", "hello world"} {
		t.Run("invalid "+invalid, func(t *testing.T) {
			got, err := f.Parse(invalid)
			require.Error(t, err)
			assert.Equal(t, "", got, "no partial result on error")
		})
	}
}
