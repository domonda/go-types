package notnull

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FloatArray_Value(t *testing.T) {
	val, err := FloatArray(nil).Value()
	assert.NoError(t, err, "FloatArray.Value")
	assert.Equal(t, "{}", val, "FloatArray(nil).Value() returns empty SQL array")

	val, err = FloatArray([]float64{}).Value()
	assert.NoError(t, err, "FloatArray.Value")
	assert.Equal(t, "{}", val, "FloatArray([]float64{}).Value() returns empty SQL array")

	val, err = FloatArray([]float64{1, 2, 3}).Value()
	assert.NoError(t, err, "FloatArray.Value")
	assert.Equal(t, "{1,2,3}", val, "FloatArray.Value")
}

func Test_FloatArray_MarshalJSON(t *testing.T) {
	val, err := json.Marshal(FloatArray(nil))
	assert.NoError(t, err, "json.Marshal(FloatArray(nil))")
	assert.Equal(t, "[]", string(val), "json.Marshal(FloatArray(nil)) returns empty JSON array")

	val, err = json.Marshal(FloatArray([]float64{}))
	assert.NoError(t, err, "json.Marshal(FloatArray([]float64{}))")
	assert.Equal(t, "[]", string(val), "json.Marshal(FloatArray([]float64{})) returns empty JSON array")

	val, err = json.Marshal(FloatArray([]float64{1, 2, 3}))
	assert.NoError(t, err, "json.Marshal(FloatArray([]float64{1, 2, 3}))")
	assert.Equal(t, "[1,2,3]", string(val), "json.Marshal(FloatArray([]float64{1, 2, 3}))")
}

func Test_FloatArray_UnmarshalJSON(t *testing.T) {
	var a FloatArray

	// JSON null and [] both unmarshal to a non-nil empty slice.
	require.NoError(t, json.Unmarshal([]byte("null"), &a))
	assert.NotNil(t, a, "JSON null unmarshals to a non-nil empty slice")
	assert.Empty(t, a)

	a = FloatArray{9}
	require.NoError(t, json.Unmarshal([]byte("[]"), &a))
	assert.NotNil(t, a, "JSON [] unmarshals to a non-nil empty slice")
	assert.Empty(t, a)

	require.NoError(t, json.Unmarshal([]byte("[1.5,2,3]"), &a))
	assert.Equal(t, FloatArray{1.5, 2, 3}, a)

	// Direct call: encoding/json pre-validates structure, so a
	// malformed document must be passed to UnmarshalJSON directly.
	assert.Error(t, a.UnmarshalJSON([]byte("[1,2,")), "invalid JSON returns an error")
}

func Test_FloatArray_JSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(FloatArray(nil))
	require.NoError(t, err)
	assert.Equal(t, "[]", string(data), "nil notnull array marshals to []")

	var decoded FloatArray
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, FloatArray{}, decoded)

	data, err = json.Marshal(FloatArray{1.5, 2, 3})
	require.NoError(t, err)
	var decoded2 FloatArray
	require.NoError(t, json.Unmarshal(data, &decoded2))
	assert.Equal(t, FloatArray{1.5, 2, 3}, decoded2)
}

func Test_FloatArray_String(t *testing.T) {
	assert.Equal(t, "[]", FloatArray(nil).String(), "FloatArray(nil).String()")
	assert.Equal(t, "[]", FloatArray([]float64{}).String(), "empty FloatArray.String()")
	assert.Equal(t, "[1]", FloatArray([]float64{1}).String(), "single element FloatArray.String()")
	assert.Equal(t, "[1.5, 2, 3]", FloatArray([]float64{1.5, 2, 3}).String(), "FloatArray.String()")
}

func Test_FloatArray_Contains(t *testing.T) {
	a := FloatArray([]float64{1.5, 2, 3})
	assert.True(t, a.Contains(1.5), "FloatArray.Contains existing")
	assert.True(t, a.Contains(3), "FloatArray.Contains existing")
	assert.False(t, a.Contains(4), "FloatArray.Contains missing")
	assert.False(t, FloatArray(nil).Contains(1), "FloatArray(nil).Contains")
}

func Test_FloatArray_Scan(t *testing.T) {
	var a FloatArray

	// A notnull array is never nil: NULL, empty input and {}
	// all scan to a non-nil empty slice.
	require.NoError(t, a.Scan(nil), "FloatArray.Scan(nil)")
	assert.NotNil(t, a, "FloatArray.Scan(nil) results in a non-nil empty slice")
	assert.Empty(t, a)

	a = FloatArray{9}
	require.NoError(t, a.Scan(""), "FloatArray.Scan empty string")
	assert.NotNil(t, a, "FloatArray.Scan empty string results in a non-nil empty slice")
	assert.Empty(t, a)

	a = FloatArray{9}
	require.NoError(t, a.Scan("{}"), "FloatArray.Scan {}")
	assert.NotNil(t, a, "FloatArray.Scan {} results in a non-nil empty slice")
	assert.Empty(t, a)

	a = FloatArray{9}
	require.NoError(t, a.Scan([]byte{}), "FloatArray.Scan empty []byte")
	assert.NotNil(t, a, "FloatArray.Scan empty []byte results in a non-nil empty slice")
	assert.Empty(t, a)

	require.NoError(t, a.Scan("{1.5,2,3}"), "FloatArray.Scan string")
	assert.Equal(t, FloatArray{1.5, 2, 3}, a, "FloatArray.Scan string")

	require.NoError(t, a.Scan([]byte("{4,5.5}")), "FloatArray.Scan []byte")
	assert.Equal(t, FloatArray{4, 5.5}, a, "FloatArray.Scan []byte")

	// Error paths
	assert.Error(t, a.Scan(123), "FloatArray.Scan unsupported type")
	assert.Error(t, a.Scan("1,2,3"), "FloatArray.Scan missing braces")
	assert.Error(t, a.Scan("{1,2,3"), "FloatArray.Scan missing closing brace")
	assert.Error(t, a.Scan("1,2,3}"), "FloatArray.Scan missing opening brace")
	assert.Error(t, a.Scan("{1,abc,3}"), "FloatArray.Scan invalid element")
}

func Test_FloatArray_ScanValueRoundTrip(t *testing.T) {
	original := FloatArray{1.5, 2, 3, -4.25}
	val, err := original.Value()
	require.NoError(t, err, "FloatArray.Value")

	var scanned FloatArray
	require.NoError(t, scanned.Scan(val), "FloatArray.Scan round-trip")
	assert.Equal(t, original, scanned, "FloatArray Value->Scan round-trip")
}

func Test_FloatArray_Sort(t *testing.T) {
	a := FloatArray{3, 1.5, 2, -1}
	assert.Equal(t, 4, a.Len(), "FloatArray.Len")
	assert.True(t, a.Less(1, 0), "FloatArray.Less")
	assert.False(t, a.Less(0, 1), "FloatArray.Less")

	sort.Sort(a)
	assert.Equal(t, FloatArray{-1, 1.5, 2, 3}, a, "sort.Sort(FloatArray)")
	assert.True(t, sort.IsSorted(a), "FloatArray sorted")
}
