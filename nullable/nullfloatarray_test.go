package nullable

import (
	"database/sql"
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ sql.Scanner = (*NullFloatArray)(nil)

func Test_NullFloatArray_Value(t *testing.T) {
	val, err := NullFloatArray(nil).Value()
	assert.NoError(t, err)
	assert.Nil(t, val, "nil NullFloatArray returns SQL NULL")

	val, err = NullFloatArray{}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val, "empty non-nil NullFloatArray returns {}")

	val, err = NullFloatArray{TypeFrom(1.5), {}, TypeFrom(3.0)}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{1.5,NULL,3}", val)
}

func Test_NullFloatArray_Scan(t *testing.T) {
	var a NullFloatArray

	// Only an untyped nil source is SQL NULL.
	require.NoError(t, a.Scan(nil))
	assert.Nil(t, a)

	// The empty array {} scans to a non-nil empty slice,
	// distinct from SQL NULL (nil slice).
	a = NullFloatArray{TypeFrom(9.0)}
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	// Empty input must not panic and scans to a non-nil empty slice.
	a = NullFloatArray{TypeFrom(9.0)}
	require.NotPanics(t, func() { _ = a.Scan("") })
	require.NoError(t, a.Scan(""))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	a = NullFloatArray{TypeFrom(9.0)}
	require.NoError(t, a.Scan([]byte{}))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, a.Scan("{1.5,NULL,3}"))
	assert.Equal(t, NullFloatArray{TypeFrom(1.5), {}, TypeFrom(3.0)}, a)

	require.NoError(t, a.Scan([]byte("{1,null,2}")))
	assert.Equal(t, NullFloatArray{TypeFrom(1.0), {}, TypeFrom(2.0)}, a)

	assert.Error(t, a.Scan("not an array"))
	assert.Error(t, a.Scan("{notanumber}"))
	assert.Error(t, a.Scan(123))
}

// PostgreSQL writes the non-finite float8 values as NaN, Infinity and
// -Infinity, so Scan must read them next to NULL elements, and the values
// written by Value must scan back to the same floats. JSON has no NaN or
// infinity, so json.Marshal must fail for them instead of writing invalid JSON.
func Test_NullFloatArray_NonFinite(t *testing.T) {
	var a NullFloatArray
	require.NoError(t, a.Scan("{NaN,NULL,Infinity,-Infinity}"))
	require.Len(t, a, 4)
	assert.True(t, math.IsNaN(a[0].Get()))
	assert.True(t, a[1].IsNull())
	assert.Equal(t, math.Inf(1), a[2].Get())
	assert.Equal(t, math.Inf(-1), a[3].Get())

	val, err := NullFloatArray{TypeFrom(math.NaN()), TypeFrom(math.Inf(1)), TypeFrom(math.Inf(-1))}.Value()
	require.NoError(t, err)
	var scanned NullFloatArray
	require.NoError(t, scanned.Scan(val), "Scan(%q)", val)
	require.Len(t, scanned, 3)
	assert.True(t, math.IsNaN(scanned[0].Get()))
	assert.Equal(t, math.Inf(1), scanned[1].Get())
	assert.Equal(t, math.Inf(-1), scanned[2].Get())

	for _, f := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := json.Marshal(NullFloatArray{TypeFrom(1.0), TypeFrom(f)})
		var unsupported *json.UnsupportedValueError
		assert.ErrorAs(t, err, &unsupported, "json.Marshal of %v", f)
	}
}

func Test_NullFloatArray_RoundTrip(t *testing.T) {
	original := NullFloatArray{TypeFrom(-7.25), {}, TypeFrom(42.0)}
	val, err := original.Value()
	require.NoError(t, err)

	var scanned NullFloatArray
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, original, scanned)
}

func Test_NullFloatArray_Floats(t *testing.T) {
	assert.Nil(t, NullFloatArray(nil).Floats())

	// Null elements project to 0.
	a := NullFloatArray{TypeFrom(1.5), {}, TypeFrom(3.0)}
	assert.Equal(t, []float64{1.5, 0, 3}, a.Floats())
}

func Test_NullFloatArray_IsNull(t *testing.T) {
	assert.True(t, NullFloatArray(nil).IsNull())
	assert.False(t, NullFloatArray{}.IsNull())
	assert.False(t, NullFloatArray{TypeFrom(0.0)}.IsNull())
}

func Test_NullFloatArray_MarshalJSON(t *testing.T) {
	// A nil slice marshals to JSON null (not [] — nullable keeps
	// NULL distinct from an empty array).
	b, err := json.Marshal(NullFloatArray(nil))
	require.NoError(t, err)
	assert.Equal(t, "null", string(b))

	b, err = json.Marshal(NullFloatArray{})
	require.NoError(t, err)
	assert.Equal(t, "[]", string(b), "empty non-nil slice marshals to []")

	// Each Type[float64] element marshals a value as a number and a
	// null element as JSON null — no slice-level method needed.
	b, err = json.Marshal(NullFloatArray{TypeFrom(1.5), {}, TypeFrom(3.0)})
	require.NoError(t, err)
	assert.Equal(t, "[1.5,null,3]", string(b))
}

func Test_NullFloatArray_UnmarshalJSON(t *testing.T) {
	var a NullFloatArray

	// JSON null unmarshals to a nil slice (SQL NULL).
	a = NullFloatArray{TypeFrom(9.0)}
	require.NoError(t, json.Unmarshal([]byte("null"), &a))
	assert.Nil(t, a)

	// [] unmarshals to a non-nil empty slice.
	a = NullFloatArray{TypeFrom(9.0)}
	require.NoError(t, json.Unmarshal([]byte("[]"), &a))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, json.Unmarshal([]byte("[1.5,null,3]"), &a))
	assert.Equal(t, NullFloatArray{TypeFrom(1.5), {}, TypeFrom(3.0)}, a)

	assert.Error(t, json.Unmarshal([]byte("[1,2,"), &a), "invalid JSON returns an error")
}

func Test_NullFloatArray_JSONRoundTrip(t *testing.T) {
	for _, original := range []NullFloatArray{
		nil,
		{},
		{TypeFrom(1.5), {}, TypeFrom(-3.25)},
	} {
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded NullFloatArray
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, original, decoded)
	}
}
