package nullable

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ sql.Scanner    = (*FloatArray)(nil)
	_ driver.Valuer  = FloatArray(nil)
	_ sort.Interface = FloatArray(nil)
)

func Test_FloatArray_Value(t *testing.T) {
	val, err := FloatArray(nil).Value()
	assert.NoError(t, err, "FloatArray.Value")
	assert.Equal(t, nil, val, "FloatArray(nil).Value() returns empty SQL array")

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
	assert.Equal(t, "null", string(val), "json.Marshal(FloatArray(nil)) returns empty JSON array")

	val, err = json.Marshal(FloatArray([]float64{}))
	assert.NoError(t, err, "json.Marshal(FloatArray([]float64{}))")
	assert.Equal(t, "[]", string(val), "json.Marshal(FloatArray([]float64{})) returns empty JSON array")

	val, err = json.Marshal(FloatArray([]float64{1, 2, 3}))
	assert.NoError(t, err, "json.Marshal(FloatArray([]float64{1, 2, 3}))")
	assert.Equal(t, "[1,2,3]", string(val), "json.Marshal(FloatArray([]float64{1, 2, 3}))")
}

func Test_FloatArray_IsNull(t *testing.T) {
	assert.True(t, FloatArray(nil).IsNull())
	assert.False(t, FloatArray{}.IsNull())
	assert.False(t, FloatArray{1.5}.IsNull())
}

func Test_FloatArray_String(t *testing.T) {
	assert.Equal(t, "NULL", FloatArray(nil).String())
	assert.Equal(t, "[]", FloatArray{}.String())
	assert.Equal(t, "[1.5, 2, 3]", FloatArray{1.5, 2, 3}.String())
}

func Test_FloatArray_StringOr(t *testing.T) {
	assert.Equal(t, "n/a", FloatArray(nil).StringOr("n/a"))
	assert.Equal(t, "[1.5]", FloatArray{1.5}.StringOr("n/a"))
}

func Test_FloatArray_Contains(t *testing.T) {
	a := FloatArray{1.5, 2, 3}
	assert.True(t, a.Contains(2))
	assert.False(t, a.Contains(99))
	assert.False(t, FloatArray(nil).Contains(0))
}

func Test_FloatArray_Scan(t *testing.T) {
	var a FloatArray

	// Only an untyped nil source is SQL NULL.
	require.NoError(t, a.Scan(nil))
	assert.Nil(t, a)

	// The empty array {} scans to a non-nil empty slice.
	a = FloatArray{9}
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, a.Scan("{1.5,2,3}"))
	assert.Equal(t, FloatArray{1.5, 2, 3}, a)

	require.NoError(t, a.Scan([]byte("{-7.25,42}")))
	assert.Equal(t, FloatArray{-7.25, 42}, a)

	assert.Error(t, a.Scan("not an array"))
	assert.Error(t, a.Scan("{notanumber}"))
	assert.Error(t, a.Scan(123))
}

// A Scan error must name the type the caller scans into, not the
// notnull.FloatArray it shares the array format with, so the message
// leads to the right Go field. The parse error must stay wrapped.
func Test_FloatArray_ScanErrorNamesType(t *testing.T) {
	var a FloatArray
	for _, src := range []any{123, "not an array", "{1.5,NULL,3}"} {
		err := a.Scan(src)
		require.Error(t, err, "Scan(%#v)", src)
		assert.Contains(t, err.Error(), "nullable.FloatArray", "Scan(%#v)", src)
		assert.NotContains(t, err.Error(), "notnull", "Scan(%#v)", src)
	}
	assert.ErrorIs(t, a.Scan("{1.5,x}"), strconv.ErrSyntax)
}

// PostgreSQL writes the non-finite float8 values as NaN, Infinity and
// -Infinity, so Scan must read them, and the values written by Value
// must scan back to the same floats. JSON has no NaN or infinity,
// so json.Marshal must fail for them instead of writing invalid JSON.
func Test_FloatArray_NonFinite(t *testing.T) {
	var a FloatArray
	require.NoError(t, a.Scan("{NaN,Infinity,-Infinity}"))
	require.Len(t, a, 3)
	assert.True(t, math.IsNaN(a[0]))
	assert.Equal(t, math.Inf(1), a[1])
	assert.Equal(t, math.Inf(-1), a[2])

	val, err := FloatArray{math.NaN(), math.Inf(1), math.Inf(-1)}.Value()
	require.NoError(t, err)
	var scanned FloatArray
	require.NoError(t, scanned.Scan(val), "Scan(%q)", val)
	require.Len(t, scanned, 3)
	assert.True(t, math.IsNaN(scanned[0]))
	assert.Equal(t, math.Inf(1), scanned[1])
	assert.Equal(t, math.Inf(-1), scanned[2])

	for _, f := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := json.Marshal(FloatArray{1, f})
		var unsupported *json.UnsupportedValueError
		assert.ErrorAs(t, err, &unsupported, "json.Marshal of %v", f)
	}
}

func Test_FloatArray_RoundTrip(t *testing.T) {
	for _, original := range []FloatArray{nil, {}, {1.5, 2, 3}, {-5.5, 0, 5.5}} {
		val, err := original.Value()
		require.NoError(t, err)

		var scanned FloatArray
		require.NoError(t, scanned.Scan(val))
		assert.Equal(t, original, scanned)
	}
}

func Test_FloatArray_SortInterface(t *testing.T) {
	a := FloatArray{3, 1.5, 2}
	assert.Equal(t, 3, a.Len())
	assert.True(t, a.Less(1, 0))
	assert.False(t, a.Less(0, 1))
	a.Swap(0, 2)
	assert.Equal(t, FloatArray{2, 1.5, 3}, a)

	sort.Sort(a)
	assert.Equal(t, FloatArray{1.5, 2, 3}, a)
}
