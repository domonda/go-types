package nullable

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ sql.Scanner   = (*BoolArray)(nil)
	_ driver.Valuer = BoolArray(nil)
)

func Test_BoolArray_Value(t *testing.T) {
	val, err := BoolArray(nil).Value()
	require.NoError(t, err)
	assert.Nil(t, val, "nil BoolArray returns SQL NULL")

	val, err = BoolArray{}.Value()
	require.NoError(t, err)
	assert.Equal(t, "{}", val, "empty non-nil BoolArray returns {}")

	val, err = BoolArray{true, false, true}.Value()
	require.NoError(t, err)
	assert.Equal(t, "{t,f,t}", val)
}

func Test_BoolArray_Scan(t *testing.T) {
	var a BoolArray

	// Only an untyped nil source is SQL NULL.
	require.NoError(t, a.Scan(nil))
	assert.Nil(t, a)

	// The empty array {} scans to a non-nil empty slice.
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, a.Scan("{t,f,t}"))
	assert.Equal(t, BoolArray{true, false, true}, a)

	require.NoError(t, a.Scan([]byte("{f,t}")))
	assert.Equal(t, BoolArray{false, true}, a)

	assert.Error(t, a.Scan("{x}"))
	assert.Error(t, a.Scan("not an array"))
	assert.Error(t, a.Scan(123))
}

func Test_BoolArray_RoundTrip(t *testing.T) {
	original := BoolArray{true, false, false, true}
	val, err := original.Value()
	require.NoError(t, err)

	var scanned BoolArray
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, original, scanned)
}

func Test_BoolArray_Scan_PostgresLiterals(t *testing.T) {
	var a BoolArray

	// BoolArray can't represent NULL elements, so they must be an error
	// instead of being silently converted to false.
	assert.Error(t, a.Scan(`{t,NULL}`))

	// PostgreSQL emits only t and f, other spellings are rejected
	assert.Error(t, a.Scan(`{true}`))

	// Only one-dimensional arrays are supported
	assert.Error(t, a.Scan(`{{t,f},{f,t}}`))

	// Scanning the empty array into a non-nil slice keeps it non-nil
	a = BoolArray{true}
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a)
	assert.Empty(t, a)
}

func Test_BoolArray_RoundTrip_EdgeCases(t *testing.T) {
	// nil must round trip as SQL NULL and empty as {},
	// so that NULL and empty arrays stay distinguishable.
	for _, original := range []BoolArray{nil, {}, {false}, {true}} {
		val, err := original.Value()
		require.NoError(t, err)

		var scanned BoolArray
		require.NoError(t, scanned.Scan(val))
		assert.Equal(t, original, scanned)
		assert.Equal(t, original == nil, scanned == nil, "nil-ness is preserved")
	}
}
