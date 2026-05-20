package notnull

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IntArray_Value(t *testing.T) {
	val, err := IntArray(nil).Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val, "IntArray(nil).Value() returns empty SQL array")

	val, err = IntArray{}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val, "empty IntArray returns empty SQL array")

	val, err = IntArray{1, 2, 3}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{1,2,3}", val)

	val, err = IntArray{-1, 0, 42}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{-1,0,42}", val)
}

func Test_IntArray_Scan(t *testing.T) {
	var a IntArray

	// A notnull array is never nil: NULL, {} and empty input
	// all scan to a non-nil empty slice.
	require.NoError(t, a.Scan(nil))
	assert.NotNil(t, a, "Scan(nil) results in a non-nil empty slice")
	assert.Empty(t, a)

	a = IntArray{9}
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a, "{} scans to a non-nil empty slice")
	assert.Empty(t, a)

	a = IntArray{9}
	require.NoError(t, a.Scan(""))
	assert.NotNil(t, a, "empty string scans to a non-nil empty slice")
	assert.Empty(t, a)

	require.NoError(t, a.Scan("{1,2,3}"))
	assert.Equal(t, IntArray{1, 2, 3}, a)

	require.NoError(t, a.Scan([]byte("{-1,0,42}")))
	assert.Equal(t, IntArray{-1, 0, 42}, a)

	assert.Error(t, a.Scan("not an array"))
	assert.Error(t, a.Scan("{not,a,number}"))
	assert.Error(t, a.Scan(123), "unsupported type returns error")
}

func Test_IntArray_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(IntArray(nil))
	assert.NoError(t, err)
	assert.Equal(t, "[]", string(b), "IntArray(nil) marshals to empty JSON array")

	b, err = json.Marshal(IntArray{})
	assert.NoError(t, err)
	assert.Equal(t, "[]", string(b))

	b, err = json.Marshal(IntArray{1, 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, "[1,2,3]", string(b))
}

func Test_IntArray_UnmarshalJSON(t *testing.T) {
	var a IntArray

	// JSON null and [] both unmarshal to a non-nil empty slice.
	require.NoError(t, json.Unmarshal([]byte("null"), &a))
	assert.NotNil(t, a, "JSON null unmarshals to a non-nil empty slice")
	assert.Empty(t, a)

	a = IntArray{9}
	require.NoError(t, json.Unmarshal([]byte("[]"), &a))
	assert.NotNil(t, a, "JSON [] unmarshals to a non-nil empty slice")
	assert.Empty(t, a)

	require.NoError(t, json.Unmarshal([]byte("[1,2,3]"), &a))
	assert.Equal(t, IntArray{1, 2, 3}, a)

	// Direct call: encoding/json pre-validates structure, so a
	// malformed document must be passed to UnmarshalJSON directly.
	assert.Error(t, a.UnmarshalJSON([]byte("[1,2,")), "invalid JSON returns an error")
}

func Test_IntArray_JSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(IntArray(nil))
	require.NoError(t, err)
	assert.Equal(t, "[]", string(data), "nil notnull array marshals to []")

	var decoded IntArray
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, IntArray{}, decoded)

	data, err = json.Marshal(IntArray{1, 2, 3})
	require.NoError(t, err)
	var decoded2 IntArray
	require.NoError(t, json.Unmarshal(data, &decoded2))
	assert.Equal(t, IntArray{1, 2, 3}, decoded2)
}

func Test_IntArray_String(t *testing.T) {
	assert.Equal(t, "[]", IntArray(nil).String())
	assert.Equal(t, "[1]", IntArray{1}.String())
	assert.Equal(t, "[1, 2, 3]", IntArray{1, 2, 3}.String())
}

func Test_IntArray_Contains(t *testing.T) {
	a := IntArray{1, 2, 3}
	assert.True(t, a.Contains(2))
	assert.False(t, a.Contains(99))
	assert.False(t, IntArray(nil).Contains(0))
}

func Test_IntArray_Sort(t *testing.T) {
	a := IntArray{3, 1, 2, -1}
	sort.Sort(a)
	assert.Equal(t, IntArray{-1, 1, 2, 3}, a)
}

func Test_IntArray_RoundTrip(t *testing.T) {
	original := IntArray{-100, 0, 42, 1234567890}
	val, err := original.Value()
	require.NoError(t, err)

	var scanned IntArray
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, original, scanned)
}
