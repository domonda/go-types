package nullable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SplitArray(t *testing.T) {
	t.Run("nil for empty/null", func(t *testing.T) {
		for _, in := range []string{"{}", "[]", "null", "NULL", "{  }", "[ ]"} {
			elems, err := SplitArray(in)
			require.NoError(t, err, in)
			assert.Nil(t, elems, in)
		}
	})

	t.Run("SQL array", func(t *testing.T) {
		elems, err := SplitArray("{1,2,3}")
		require.NoError(t, err)
		assert.Equal(t, []string{"1", "2", "3"}, elems)
	})

	t.Run("JSON array", func(t *testing.T) {
		elems, err := SplitArray("[1, 2, 3]")
		require.NoError(t, err)
		assert.Equal(t, []string{"1", "2", "3"}, elems)
	})

	t.Run("quoted elements not unquoted", func(t *testing.T) {
		elems, err := SplitArray(`{"a","b,c"}`)
		require.NoError(t, err)
		assert.Equal(t, []string{`"a"`, `"b,c"`}, elems)
	})

	t.Run("nested arrays", func(t *testing.T) {
		elems, err := SplitArray("[[1,2],[3,4]]")
		require.NoError(t, err)
		assert.Equal(t, []string{"[1,2]", "[3,4]"}, elems)
	})

	t.Run("errors", func(t *testing.T) {
		for _, in := range []string{"", "x", "not an array", "{1,2", "{1,,2}", "[1,2"} {
			_, err := SplitArray(in)
			assert.Error(t, err, in)
		}
	})
}

func Test_SplitArrayValues(t *testing.T) {
	t.Run("nil for empty/null", func(t *testing.T) {
		for _, in := range []string{"{}", "[]", "null", "NULL", "{  }", "[ ]"} {
			values, err := SplitArrayValues(in)
			require.NoError(t, err, in)
			assert.Nil(t, values, in)
		}
	})

	t.Run("quoted elements are unquoted", func(t *testing.T) {
		values, err := SplitArrayValues(`{a,"b,c","d\\e","f\"g"}`)
		require.NoError(t, err)
		assert.Equal(t, []string{`a`, `b,c`, `d\e`, `f"g`}, values)
	})

	t.Run("JSON escapes", func(t *testing.T) {
		values, err := SplitArrayValues(`["a\nb", "\u00e4"]`)
		require.NoError(t, err)
		assert.Equal(t, []string{"a\nb", "ä"}, values)
	})

	t.Run("errors", func(t *testing.T) {
		for _, in := range []string{"", "x", "not an array", "{1,2", "{1,,2}"} {
			_, err := SplitArrayValues(in)
			assert.Error(t, err, in)
		}
	})
}

func Test_SQLArrayLiteral(t *testing.T) {
	assert.Equal(t, "NULL", SQLArrayLiteral(nil))
	assert.Equal(t, "{}", SQLArrayLiteral([]string{}))
	assert.Equal(t, `{"a"}`, SQLArrayLiteral([]string{"a"}))
	assert.Equal(t, `{"a","b"}`, SQLArrayLiteral([]string{"a", "b"}))
	assert.Equal(t, `{"a\"b"}`, SQLArrayLiteral([]string{`a"b`}))
	assert.Equal(t, `{"a\\b"}`, SQLArrayLiteral([]string{`a\b`}), "a backslash must be escaped too")
}
