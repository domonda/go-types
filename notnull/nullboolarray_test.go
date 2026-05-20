package notnull

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NullBoolArray_Value(t *testing.T) {
	val, err := NullBoolArray(nil).Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val, "NullBoolArray(nil).Value() returns empty SQL array")

	val, err = NullBoolArray{}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val)

	val, err = NullBoolArray{
		{Valid: true, Bool: true},
		{Valid: true, Bool: false},
		{Valid: false},
	}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{t,f,NULL}", val)
}

func Test_NullBoolArray_Scan(t *testing.T) {
	var a NullBoolArray

	// A notnull array is never nil: NULL, {} and empty input
	// all scan to a non-nil empty slice.
	require.NoError(t, a.Scan(nil))
	assert.NotNil(t, a, "Scan(nil) results in a non-nil empty slice")
	assert.Empty(t, a)

	a = NullBoolArray{{Valid: true}}
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a, "{} scans to a non-nil empty slice")
	assert.Empty(t, a)

	a = NullBoolArray{{Valid: true}}
	require.NoError(t, a.Scan(""))
	assert.NotNil(t, a, "empty string scans to a non-nil empty slice")
	assert.Empty(t, a)

	require.NoError(t, a.Scan("{t,f,NULL}"))
	assert.Equal(t, NullBoolArray{
		{Valid: true, Bool: true},
		{Valid: true, Bool: false},
		{Valid: false},
	}, a)

	require.NoError(t, a.Scan([]byte("{t,t}")))
	assert.Equal(t, NullBoolArray{
		{Valid: true, Bool: true},
		{Valid: true, Bool: true},
	}, a)

	assert.Error(t, a.Scan("malformed"))
	assert.Error(t, a.Scan(123))
}

func Test_NullBoolArray_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(NullBoolArray(nil))
	require.NoError(t, err)
	assert.Equal(t, "[]", string(b), "nil array marshals to []")

	b, err = json.Marshal(NullBoolArray{})
	require.NoError(t, err)
	assert.Equal(t, "[]", string(b), "empty array marshals to []")

	b, err = json.Marshal(NullBoolArray{
		{Valid: true, Bool: true},
		{Valid: true, Bool: false},
		{Valid: false},
	})
	require.NoError(t, err)
	assert.Equal(t, "[true,false,null]", string(b), "invalid elements marshal to JSON null")
}

func Test_NullBoolArray_UnmarshalJSON(t *testing.T) {
	var a NullBoolArray

	// JSON null and [] both unmarshal to a non-nil empty slice.
	require.NoError(t, json.Unmarshal([]byte("null"), &a))
	assert.NotNil(t, a, "JSON null unmarshals to a non-nil empty slice")
	assert.Empty(t, a)

	a = NullBoolArray{{Valid: true}}
	require.NoError(t, json.Unmarshal([]byte("[]"), &a))
	assert.NotNil(t, a, "JSON [] unmarshals to a non-nil empty slice")
	assert.Empty(t, a)

	require.NoError(t, json.Unmarshal([]byte("[true,false,null]"), &a))
	assert.Equal(t, NullBoolArray{
		{Valid: true, Bool: true},
		{Valid: true, Bool: false},
		{Valid: false},
	}, a)

	// Direct call: encoding/json pre-validates structure, so a
	// malformed document must be passed to UnmarshalJSON directly.
	assert.Error(t, a.UnmarshalJSON([]byte("[true,")), "invalid JSON returns an error")
}

func Test_NullBoolArray_JSONRoundTrip(t *testing.T) {
	original := NullBoolArray{
		{Valid: true, Bool: true},
		{Valid: false},
		{Valid: true, Bool: false},
	}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded NullBoolArray
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded)
}

func Test_NullBoolArray_String(t *testing.T) {
	assert.Equal(t, "[]", NullBoolArray(nil).String())
	assert.Equal(t, "[true]", NullBoolArray{{Valid: true, Bool: true}}.String())
	assert.Equal(t, "[true, false, NULL]", NullBoolArray{
		{Valid: true, Bool: true},
		{Valid: true, Bool: false},
		{Valid: false},
	}.String())
}

func Test_NullBoolArray_Bools(t *testing.T) {
	assert.Nil(t, NullBoolArray(nil).Bools())
	assert.Nil(t, NullBoolArray{}.Bools())

	a := NullBoolArray{
		{Valid: true, Bool: true},
		{Valid: true, Bool: false},
		{Valid: false, Bool: true}, // NULL → false in projection
	}
	assert.Equal(t, []bool{true, false, false}, a.Bools())
}

func Test_NullBoolArray_RoundTrip(t *testing.T) {
	original := NullBoolArray{
		{Valid: true, Bool: true},
		{Valid: false},
		{Valid: true, Bool: false},
	}
	val, err := original.Value()
	require.NoError(t, err)

	var scanned NullBoolArray
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, original, scanned)
}

// Sanity check that sql.NullBool comparisons in this test file behave as expected.
func Test_NullBoolArray_ScanZeroElements(t *testing.T) {
	var a NullBoolArray
	require.NoError(t, a.Scan("{NULL}"))
	require.Len(t, a, 1)
	assert.Equal(t, sql.NullBool{}, a[0])
}
