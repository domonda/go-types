package nullable

import (
	"database/sql"
	"database/sql/driver"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ sql.Scanner    = (*IntArray)(nil)
	_ driver.Valuer  = IntArray(nil)
	_ sort.Interface = IntArray(nil)
)

func Test_IntArray_IsNull(t *testing.T) {
	assert.True(t, IntArray(nil).IsNull())
	assert.False(t, IntArray{}.IsNull())
	assert.False(t, IntArray{1}.IsNull())
}

func Test_IntArray_String(t *testing.T) {
	assert.Equal(t, "IntArray<nil>", IntArray(nil).String())
	assert.Equal(t, "IntArray{}", IntArray{}.String())
	assert.Equal(t, "IntArray{1,2,3}", IntArray{1, 2, 3}.String())
}

func Test_IntArray_Contains(t *testing.T) {
	a := IntArray{1, 2, 3}
	assert.True(t, a.Contains(2))
	assert.False(t, a.Contains(99))
	assert.False(t, IntArray(nil).Contains(0))
}

func Test_IntArray_Value(t *testing.T) {
	val, err := IntArray(nil).Value()
	require.NoError(t, err)
	assert.Nil(t, val, "nil IntArray returns SQL NULL")

	val, err = IntArray{}.Value()
	require.NoError(t, err)
	assert.Equal(t, "{}", val, "empty non-nil IntArray returns {}")

	val, err = IntArray{1, 2, 3}.Value()
	require.NoError(t, err)
	assert.Equal(t, "{1,2,3}", val)
}

func Test_IntArray_Scan(t *testing.T) {
	var a IntArray

	// Only an untyped nil source is SQL NULL.
	require.NoError(t, a.Scan(nil))
	assert.Nil(t, a)

	// The empty array {} scans to a non-nil empty slice.
	a = IntArray{9}
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, a.Scan("{1,2,3}"))
	assert.Equal(t, IntArray{1, 2, 3}, a)

	require.NoError(t, a.Scan([]byte("{-7,42}")))
	assert.Equal(t, IntArray{-7, 42}, a)

	assert.Error(t, a.Scan("not an array"))
	assert.Error(t, a.Scan("{notanumber}"))
	assert.Error(t, a.Scan(123))
}

func Test_IntArray_RoundTrip(t *testing.T) {
	for _, original := range []IntArray{nil, {}, {1, 2, 3}, {-5, 0, 5}} {
		val, err := original.Value()
		require.NoError(t, err)

		var scanned IntArray
		require.NoError(t, scanned.Scan(val))
		assert.Equal(t, original, scanned)
	}
}

func Test_IntArray_SortInterface(t *testing.T) {
	a := IntArray{3, 1, 2}
	assert.Equal(t, 3, a.Len())
	assert.True(t, a.Less(1, 0))
	assert.False(t, a.Less(0, 1))
	a.Swap(0, 2)
	assert.Equal(t, IntArray{2, 1, 3}, a)

	sort.Sort(a)
	assert.Equal(t, IntArray{1, 2, 3}, a)
}
