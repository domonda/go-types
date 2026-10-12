package date

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
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

	// A JSON string is assigned unchanged, without validation,
	// like encoding/json did before NullableYearQuarter had UnmarshalJSON.
	for _, str := range []string{"2023-Q3", "2023-q3", "bad", ""} {
		d := doc{Period: "2024-Q1"}
		data, err := json.Marshal(map[string]string{"period": str})
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, &d))
		require.Equal(t, NullableYearQuarter(str), d.Period, "json.Unmarshal(%s)", data)
	}

	// Any other JSON value is an error and leaves the value unchanged.
	for _, data := range []string{`20233`, `true`, `{}`, `["2023-Q3"]`} {
		yq := NullableYearQuarter("2024-Q1")
		require.Error(t, json.Unmarshal([]byte(data), &yq), "json.Unmarshal(%s)", data)
		require.Equal(t, NullableYearQuarter("2024-Q1"), yq, "json.Unmarshal(%s)", data)
	}
}
