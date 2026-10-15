package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) {
	t.Run("default layouts", func(t *testing.T) {
		// Without explicit layouts the common machine formats
		// (RFC3339 with and without fractions, SQL timestamp) are accepted.
		tests := []struct {
			input string
			want  time.Time
		}{
			{input: "2023-12-25T13:14:15Z", want: time.Date(2023, 12, 25, 13, 14, 15, 0, time.UTC)},
			{input: "2023-12-25T13:14:15.123456789Z", want: time.Date(2023, 12, 25, 13, 14, 15, 123456789, time.UTC)},
			{input: "2023-12-25T13:14:15+02:00", want: time.Date(2023, 12, 25, 11, 14, 15, 0, time.UTC)},
			{input: "2023-12-25 13:14:15", want: time.Date(2023, 12, 25, 13, 14, 15, 0, time.UTC)},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				got, err := ParseTime(tt.input)
				require.NoError(t, err)
				assert.True(t, tt.want.Equal(got), "got %s, want %s", got, tt.want)
			})
		}
	})

	t.Run("custom layouts are tried in order", func(t *testing.T) {
		got, err := ParseTime("25.12.2023 13:14", time.RFC3339, "02.01.2006 15:04")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2023, 12, 25, 13, 14, 0, 0, time.UTC), got)

		// Passing layouts replaces the defaults instead of extending them
		_, err = ParseTime("2023-12-25T13:14:15Z", "02.01.2006 15:04")
		assert.Error(t, err)
	})

	t.Run("errors", func(t *testing.T) {
		for _, invalid := range []string{"", "2023-12-25", "not a time", "2023-12-25T25:00:00Z"} {
			got, err := ParseTime(invalid)
			require.Error(t, err, "input %q", invalid)
			assert.True(t, got.IsZero(), "zero time on error")
			assert.Contains(t, err.Error(), invalid, "error mentions the input")
		}
	})
}
