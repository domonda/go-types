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

func Test_StringArray_RoundTrip_SpecialCharacters(t *testing.T) {
	// Every element is quoted and escaped by Value, so content that has
	// a meaning in the PostgreSQL array literal syntax must survive
	// a Value/Scan round trip unchanged.
	tests := []struct {
		name  string
		array StringArray
	}{
		{name: "NULL literal as text", array: StringArray{"NULL", "null"}},
		{name: "comma", array: StringArray{"a,b", ","}},
		{name: "braces", array: StringArray{"{a}", "}", "{"}},
		{name: "backslash", array: StringArray{`C:\path\`, `\`, `\\`}},
		{name: "double quotes", array: StringArray{`"`, `""`, `say "hi"`}},
		{name: "whitespace", array: StringArray{" leading", "trailing ", " ", "\t\n"}},
		{name: "unicode", array: StringArray{"Grüße", "日本", "€"}},
		{name: "single empty element", array: StringArray{""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.array.Value()
			require.NoError(t, err)

			var scanned StringArray
			require.NoError(t, scanned.Scan(val))
			assert.Equal(t, tt.array, scanned)
		})
	}
}

func Test_StringArray_Scan_PostgresLiterals(t *testing.T) {
	var a StringArray

	// A quoted "NULL" is the text NULL, not an SQL NULL element
	require.NoError(t, a.Scan(`{"NULL",b}`))
	assert.Equal(t, StringArray{"NULL", "b"}, a)

	// StringArray can't represent NULL elements, so they must be an error
	// instead of being silently converted to an empty string.
	assert.Error(t, a.Scan(`{a,NULL}`))

	// Only one-dimensional arrays are supported
	assert.Error(t, a.Scan(`{{a,b},{c,d}}`))

	// Scanning the empty array into a non-nil slice keeps it non-nil
	a = StringArray{"x"}
	require.NoError(t, a.Scan("{}"))
	assert.NotNil(t, a)
	assert.Empty(t, a)
}
