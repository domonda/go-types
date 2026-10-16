package date

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/internal/normtest"
)

func TestNullableYearQuarter_UnmarshalJSON(t *testing.T) {
	type doc struct {
		Period NullableYearQuarter `json:"period"`
	}
	// JSON null must reset a field that holds a value, like for
	// uu.NullableID and email.NullableAddress, else decoding a partial
	// update with null into an existing struct keeps the old period.
	d := doc{Period: "2024-Q1"}
	require.NoError(t, json.Unmarshal([]byte(`{"period":null}`), &d))
	require.Equal(t, YearQuarterNull, d.Period)

	// A JSON string is trimmed and uppercased. A blank one is null.
	// A zero quarter is kept. An invalid one is kept after that
	// prettification, so one bad period doesn't fail a whole document.
	for str, want := range map[string]NullableYearQuarter{
		"2023-Q3":   "2023-Q3",
		" 2023-q3 ": "2023-Q3",
		"2023-Q5":   "2023-Q5",
		"bad":       "BAD",
		"":          YearQuarterNull,
		"  ":        YearQuarterNull,
	} {
		d := doc{Period: "2024-Q1"}
		data, err := json.Marshal(map[string]string{"period": str})
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, &d))
		require.Equal(t, want, d.Period, "json.Unmarshal(%s)", data)
	}

	// Any other JSON value is an error and leaves the value unchanged.
	for _, data := range []string{`20233`, `true`, `{}`, `["2023-Q3"]`} {
		yq := NullableYearQuarter("2024-Q1")
		require.Error(t, json.Unmarshal([]byte(data), &yq), "json.Unmarshal(%s)", data)
		require.Equal(t, NullableYearQuarter("2024-Q1"), yq, "json.Unmarshal(%s)", data)
	}
}

// A blank value is null, valid and normalizes to YearQuarterNull,
// like NullableDate, but it is not normalized itself.
func TestNullableYearQuarter_Normalized(t *testing.T) {
	for input, want := range map[NullableYearQuarter]NullableYearQuarter{
		"":          YearQuarterNull,
		" \t":       YearQuarterNull,
		"2023-Q3":   "2023-Q3",
		" 2023-q3 ": "2023-Q3",
	} {
		got, err := input.Normalized()
		require.NoError(t, err, "%q", input)
		require.Equal(t, want, got, "%q", input)
		require.True(t, input.Valid(), "%q", input)
		require.True(t, input.IsNull() == (want == YearQuarterNull), "IsNull of %q", input)
		require.Equal(t, input == want, input.ValidAndNormalized(), "%q", input)
	}
	got, err := NullableYearQuarter("2023-Q5").Normalized()
	require.Error(t, err)
	require.Equal(t, NullableYearQuarter("2023-Q5"), got, "unchanged")
	require.Equal(t, 0, NullableYearQuarter("  ").Compare(YearQuarterNull), "whitespace is null")
}

// TestNullableYearQuarter_ReadingAndWriting checks the reading and writing
// contract shared by the normalizing types, with a blank or zero
// NullableYearQuarter as null like for NullableDate.
func TestNullableYearQuarter_ReadingAndWriting(t *testing.T) {
	normtest.Run(t, normtest.Spec[NullableYearQuarter]{
		Canonical:    "2023-Q3",
		Variants:     []NullableYearQuarter{"2023-q3", " 2023-Q3 ", "\t2023-q3\n"},
		Invalid:      []NullableYearQuarter{"2023-Q5", "2023-Q", "0000-Q2", "Q3/2023", "bad"},
		Blank:        []NullableYearQuarter{"0001-Q1", "0001-q1", "0000-Q0", " 0001-Q1 "},
		Nullable:     true,
		StoredString: true,
		Prettify:     normtest.TrimUpper[NullableYearQuarter],
	})
}

// A date or timestamp column scans as its year-quarter, like
// NullableDate.Scan scans it as its date, and the zero time as null.
func TestNullableYearQuarter_ScanTime(t *testing.T) {
	var yq NullableYearQuarter
	require.NoError(t, yq.Scan(time.Date(2023, 8, 15, 10, 0, 0, 0, time.UTC)))
	require.Equal(t, NullableYearQuarter("2023-Q3"), yq)
	require.NoError(t, yq.Scan(time.Time{}))
	require.Equal(t, YearQuarterNull, yq)
}
