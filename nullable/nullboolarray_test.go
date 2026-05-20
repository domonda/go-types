package nullable

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ sql.Scanner   = (*NullBoolArray)(nil)
	_ driver.Valuer = NullBoolArray(nil)
)

func Test_NullBoolArray_IsNull(t *testing.T) {
	assert.True(t, NullBoolArray(nil).IsNull())
	assert.False(t, NullBoolArray{}.IsNull())
	assert.False(t, NullBoolArray{TypeFrom(false)}.IsNull())
}

func Test_NullBoolArray_Bools(t *testing.T) {
	assert.Nil(t, NullBoolArray(nil).Bools())

	// Null elements project to false.
	a := NullBoolArray{TypeFrom(true), {}, TypeFrom(false)}
	assert.Equal(t, []bool{true, false, false}, a.Bools())
}

func Test_NullBoolArray_String(t *testing.T) {
	assert.Equal(t, "NullBoolArray<nil>", NullBoolArray(nil).String())
	assert.Equal(t, "NullBoolArray{}", NullBoolArray{}.String())
	assert.Equal(t, "NullBoolArray{t,NULL,f}", NullBoolArray{
		TypeFrom(true), {}, TypeFrom(false),
	}.String())
}

func Test_NullBoolArray_Value(t *testing.T) {
	val, err := NullBoolArray(nil).Value()
	require.NoError(t, err)
	assert.Nil(t, val, "nil NullBoolArray returns SQL NULL")

	val, err = NullBoolArray{}.Value()
	require.NoError(t, err)
	assert.Equal(t, "{}", val, "empty non-nil NullBoolArray returns {}")

	val, err = NullBoolArray{TypeFrom(true), {}, TypeFrom(false)}.Value()
	require.NoError(t, err)
	assert.Equal(t, "{t,NULL,f}", val)
}

func Test_NullBoolArray_Scan(t *testing.T) {
	var a NullBoolArray

	// Only an untyped nil source is SQL NULL.
	require.NoError(t, a.Scan(nil))
	assert.Nil(t, a)

	// The empty array {} scans to a non-nil empty slice.
	a = NullBoolArray{TypeFrom(true)}
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	// Empty input must not panic and scans to a non-nil empty slice.
	a = NullBoolArray{TypeFrom(true)}
	require.NotPanics(t, func() { _ = a.Scan("") })
	require.NoError(t, a.Scan(""))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	a = NullBoolArray{TypeFrom(true)}
	require.NoError(t, a.Scan([]byte{}))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, a.Scan("{t,NULL,f}"))
	assert.Equal(t, NullBoolArray{TypeFrom(true), {}, TypeFrom(false)}, a)

	require.NoError(t, a.Scan([]byte("{f,t}")))
	assert.Equal(t, NullBoolArray{TypeFrom(false), TypeFrom(true)}, a)

	// Unrecognized elements are treated as null (not an error).
	require.NoError(t, a.Scan("{x}"))
	assert.Equal(t, NullBoolArray{Type[bool]{}}, a)

	assert.Error(t, a.Scan("not an array"))
	assert.Error(t, a.Scan(123))
}

func Test_NullBoolArray_RoundTrip(t *testing.T) {
	original := NullBoolArray{TypeFrom(true), {}, TypeFrom(false)}
	val, err := original.Value()
	require.NoError(t, err)

	var scanned NullBoolArray
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, original, scanned)
}

func Test_NullBoolArray_MarshalJSON(t *testing.T) {
	// A nil slice marshals to JSON null (not [] — nullable keeps
	// NULL distinct from an empty array).
	b, err := json.Marshal(NullBoolArray(nil))
	require.NoError(t, err)
	assert.Equal(t, "null", string(b))

	b, err = json.Marshal(NullBoolArray{})
	require.NoError(t, err)
	assert.Equal(t, "[]", string(b), "empty non-nil slice marshals to []")

	// Each Type[bool] element marshals a value as true/false and a
	// null element as JSON null — no slice-level method needed.
	b, err = json.Marshal(NullBoolArray{TypeFrom(true), {}, TypeFrom(false)})
	require.NoError(t, err)
	assert.Equal(t, "[true,null,false]", string(b))
}

func Test_NullBoolArray_UnmarshalJSON(t *testing.T) {
	var a NullBoolArray

	// JSON null unmarshals to a nil slice (SQL NULL).
	a = NullBoolArray{TypeFrom(true)}
	require.NoError(t, json.Unmarshal([]byte("null"), &a))
	assert.Nil(t, a)

	// [] unmarshals to a non-nil empty slice.
	a = NullBoolArray{TypeFrom(true)}
	require.NoError(t, json.Unmarshal([]byte("[]"), &a))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, json.Unmarshal([]byte("[true,null,false]"), &a))
	assert.Equal(t, NullBoolArray{TypeFrom(true), {}, TypeFrom(false)}, a)

	assert.Error(t, json.Unmarshal([]byte("[true,"), &a), "invalid JSON returns an error")
}

func Test_NullBoolArray_JSONRoundTrip(t *testing.T) {
	for _, original := range []NullBoolArray{
		nil,
		{},
		{TypeFrom(true), {}, TypeFrom(false)},
	} {
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded NullBoolArray
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, original, decoded)
	}
}
