package nullable

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ sql.Scanner   = (*StringArray)(nil)
	_ driver.Valuer = StringArray(nil)
)

func Test_StringArray_Value(t *testing.T) {
	val, err := StringArray(nil).Value()
	require.NoError(t, err)
	assert.Nil(t, val, "nil StringArray returns SQL NULL")

	val, err = StringArray{}.Value()
	require.NoError(t, err)
	assert.Equal(t, "{}", val, "empty non-nil StringArray returns {}")

	val, err = StringArray{"a", "b"}.Value()
	require.NoError(t, err)
	assert.Equal(t, `{"a","b"}`, val)

	val, err = StringArray{`a"b`}.Value()
	require.NoError(t, err)
	assert.Equal(t, `{"a\"b"}`, val)
}

func Test_StringArray_Scan(t *testing.T) {
	var a StringArray

	// Only an untyped nil source is SQL NULL.
	require.NoError(t, a.Scan(nil))
	assert.Nil(t, a)

	// The empty array {} scans to a non-nil empty slice.
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a)
	assert.Empty(t, a)

	require.NoError(t, a.Scan(`{"a","b"}`))
	assert.Equal(t, StringArray{"a", "b"}, a)

	require.NoError(t, a.Scan([]byte(`{x,y}`)))
	assert.Equal(t, StringArray{"x", "y"}, a)

	assert.Error(t, a.Scan("not an array"))
	assert.Error(t, a.Scan(123))
}

func Test_StringArray_RoundTrip(t *testing.T) {
	original := StringArray{"hello", "world", `with "quotes"`, ""}
	val, err := original.Value()
	require.NoError(t, err)

	var scanned StringArray
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, original, scanned)
}
