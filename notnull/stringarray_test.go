package notnull

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StringArray_Value(t *testing.T) {
	val, err := StringArray(nil).Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val, "StringArray(nil).Value() returns empty SQL array")

	val, err = StringArray{}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val, "empty StringArray returns empty SQL array")

	val, err = StringArray{"a", "b", "c"}.Value()
	assert.NoError(t, err)
	assert.Equal(t, `{"a","b","c"}`, val)
}

func Test_StringArray_Scan(t *testing.T) {
	var a StringArray

	// A notnull array is never nil: NULL and {} scan to a
	// non-nil empty slice.
	assert.NoError(t, a.Scan(nil))
	assert.NotNil(t, a, "Scan(nil) results in a non-nil empty slice")
	assert.Empty(t, a)

	a = StringArray{"old"}
	assert.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a, "{} scans to a non-nil empty slice")
	assert.Empty(t, a)

	assert.NoError(t, a.Scan("{x,y,z}"))
	assert.Equal(t, StringArray{"x", "y", "z"}, a)

	assert.NoError(t, a.Scan([]byte(`{"a","b"}`)))
	assert.Equal(t, StringArray{"a", "b"}, a)
}

func Test_StringArray_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(StringArray(nil))
	assert.NoError(t, err)
	assert.Equal(t, "[]", string(b))

	b, err = json.Marshal(StringArray{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, `["a","b"]`, string(b))
}

func Test_StringArray_UnmarshalJSON(t *testing.T) {
	var a StringArray

	// JSON null and [] both unmarshal to a non-nil empty slice.
	assert.NoError(t, json.Unmarshal([]byte("null"), &a))
	assert.NotNil(t, a, "JSON null unmarshals to a non-nil empty slice")
	assert.Empty(t, a)

	a = StringArray{"old"}
	assert.NoError(t, json.Unmarshal([]byte("[]"), &a))
	assert.NotNil(t, a, "JSON [] unmarshals to a non-nil empty slice")
	assert.Empty(t, a)

	assert.NoError(t, json.Unmarshal([]byte(`["a","b"]`), &a))
	assert.Equal(t, StringArray{"a", "b"}, a)

	// Direct call: encoding/json pre-validates structure, so a
	// malformed document must be passed to UnmarshalJSON directly.
	assert.Error(t, a.UnmarshalJSON([]byte(`["a",`)), "invalid JSON returns an error")
}

func Test_StringArray_JSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(StringArray(nil))
	assert.NoError(t, err)
	assert.Equal(t, "[]", string(data), "nil notnull array marshals to []")

	var decoded StringArray
	assert.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, StringArray{}, decoded)

	data, err = json.Marshal(StringArray{"a", "b"})
	assert.NoError(t, err)
	var decoded2 StringArray
	assert.NoError(t, json.Unmarshal(data, &decoded2))
	assert.Equal(t, StringArray{"a", "b"}, decoded2)
}

func Test_StringArray_Contains(t *testing.T) {
	a := StringArray{"a", "b", "c"}
	assert.True(t, a.Contains("b"))
	assert.False(t, a.Contains("z"))
}

func Test_StringArray_Sort(t *testing.T) {
	a := StringArray{"c", "a", "b", "a"}
	assert.Equal(t, 4, a.Len(), "StringArray.Len")
	assert.True(t, a.Less(1, 0), "StringArray.Less")
	assert.False(t, a.Less(0, 1), "StringArray.Less")

	sort.Sort(a)
	assert.Equal(t, StringArray{"a", "a", "b", "c"}, a, "sort.Sort(StringArray)")
	assert.True(t, sort.IsSorted(a), "StringArray sorted")
}
