package nullable

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ sql.Scanner = (*NullIntArray)(nil)

func Test_NullIntArray_Value(t *testing.T) {
	val, err := NullIntArray(nil).Value()
	assert.NoError(t, err)
	assert.Nil(t, val, "nil NullIntArray returns SQL NULL")

	val, err = NullIntArray{}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val, "empty non-nil NullIntArray returns {}")

	val, err = NullIntArray{TypeFrom[int64](1), {}, TypeFrom[int64](3)}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{1,NULL,3}", val)
}

func Test_NullIntArray_Scan(t *testing.T) {
	var a NullIntArray

	// Only an untyped nil source is SQL NULL.
	require.NoError(t, a.Scan(nil))
	assert.Nil(t, a)

	// The empty array {} scans to a non-nil empty slice,
	// distinct from SQL NULL (nil slice).
	a = NullIntArray{TypeFrom[int64](9)}
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	// Empty input must not panic and scans to a non-nil empty slice.
	a = NullIntArray{TypeFrom[int64](9)}
	require.NotPanics(t, func() { _ = a.Scan("") })
	require.NoError(t, a.Scan(""))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	a = NullIntArray{TypeFrom[int64](9)}
	require.NoError(t, a.Scan([]byte{}))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, a.Scan("{1,NULL,3}"))
	assert.Equal(t, NullIntArray{TypeFrom[int64](1), {}, TypeFrom[int64](3)}, a)

	require.NoError(t, a.Scan([]byte("{1,null,2}")))
	assert.Equal(t, NullIntArray{TypeFrom[int64](1), {}, TypeFrom[int64](2)}, a)

	assert.Error(t, a.Scan("not an array"))
	assert.Error(t, a.Scan("{notanumber}"))
	assert.Error(t, a.Scan(123))
}

func Test_NullIntArray_RoundTrip(t *testing.T) {
	original := NullIntArray{TypeFrom[int64](-7), {}, TypeFrom[int64](42)}
	val, err := original.Value()
	require.NoError(t, err)

	var scanned NullIntArray
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, original, scanned)
}

func Test_NullIntArray_Ints(t *testing.T) {
	assert.Nil(t, NullIntArray(nil).Ints())

	// Null elements project to 0.
	a := NullIntArray{TypeFrom[int64](1), {}, TypeFrom[int64](3)}
	assert.Equal(t, []int64{1, 0, 3}, a.Ints())
}

func Test_NullIntArray_IsNull(t *testing.T) {
	assert.True(t, NullIntArray(nil).IsNull())
	assert.False(t, NullIntArray{}.IsNull())
	assert.False(t, NullIntArray{TypeFrom[int64](0)}.IsNull())
}

func Test_NullIntArray_MarshalJSON(t *testing.T) {
	// A nil slice marshals to JSON null (not [] — nullable keeps
	// NULL distinct from an empty array).
	b, err := json.Marshal(NullIntArray(nil))
	require.NoError(t, err)
	assert.Equal(t, "null", string(b))

	b, err = json.Marshal(NullIntArray{})
	require.NoError(t, err)
	assert.Equal(t, "[]", string(b), "empty non-nil slice marshals to []")

	// Each Type[int64] element marshals a value as a number and a
	// null element as JSON null — no slice-level method needed.
	b, err = json.Marshal(NullIntArray{TypeFrom[int64](1), {}, TypeFrom[int64](3)})
	require.NoError(t, err)
	assert.Equal(t, "[1,null,3]", string(b))
}

func Test_NullIntArray_UnmarshalJSON(t *testing.T) {
	var a NullIntArray

	// JSON null unmarshals to a nil slice (SQL NULL).
	a = NullIntArray{TypeFrom[int64](9)}
	require.NoError(t, json.Unmarshal([]byte("null"), &a))
	assert.Nil(t, a)

	// [] unmarshals to a non-nil empty slice.
	a = NullIntArray{TypeFrom[int64](9)}
	require.NoError(t, json.Unmarshal([]byte("[]"), &a))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, json.Unmarshal([]byte("[1,null,3]"), &a))
	assert.Equal(t, NullIntArray{TypeFrom[int64](1), {}, TypeFrom[int64](3)}, a)

	assert.Error(t, json.Unmarshal([]byte("[1,2,"), &a), "invalid JSON returns an error")
}

func Test_NullIntArray_JSONRoundTrip(t *testing.T) {
	for _, original := range []NullIntArray{
		nil,
		{},
		{TypeFrom[int64](1), {}, TypeFrom[int64](-3)},
	} {
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded NullIntArray
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, original, decoded)
	}
}
