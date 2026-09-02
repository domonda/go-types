package notnull

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SplitArray(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{name: "empty SQL array", input: "{}", want: []string{}},
		{name: "empty JSON array", input: "[]", want: []string{}},
		{name: "null literal", input: "null", want: []string{}},
		{name: "NULL literal", input: "NULL", want: []string{}},
		{name: "SQL ints", input: "{1,2,3}", want: []string{"1", "2", "3"}},
		{name: "JSON ints", input: "[1,2,3]", want: []string{"1", "2", "3"}},
		{name: "SQL quoted strings keep quotes", input: `{"a","b","c"}`, want: []string{`"a"`, `"b"`, `"c"`}},
		{name: "JSON quoted strings keep quotes", input: `["a","b"]`, want: []string{`"a"`, `"b"`}},

		{name: "too short", input: "{", wantErr: true},
		{name: "not an array", input: "abc", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SplitArray(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.NotNil(t, got, "SplitArray must never return nil on success")
		})
	}
}

func Test_SplitArrayValues(t *testing.T) {
	t.Run("empty slice for empty/null", func(t *testing.T) {
		for _, in := range []string{"{}", "[]", "null", "NULL"} {
			values, err := SplitArrayValues(in)
			require.NoError(t, err, in)
			assert.NotNil(t, values, "SplitArrayValues must never return nil on success")
			assert.Empty(t, values, in)
		}
	})

	t.Run("quoted elements are unquoted", func(t *testing.T) {
		values, err := SplitArrayValues(`{a,"b,c","d\\e","f\"g"}`)
		require.NoError(t, err)
		assert.Equal(t, []string{`a`, `b,c`, `d\e`, `f"g`}, values)
	})

	t.Run("errors", func(t *testing.T) {
		_, err := SplitArrayValues("abc")
		assert.Error(t, err)
	})
}

func Test_SQLArrayLiteral(t *testing.T) {
	assert.Equal(t, "{}", SQLArrayLiteral(nil), "nil slice yields empty array, not NULL")
	assert.Equal(t, "{}", SQLArrayLiteral([]string{}), "empty slice yields empty array")
	assert.Equal(t, `{"a","b"}`, SQLArrayLiteral([]string{"a", "b"}))
	assert.Equal(t, `{"a\\b"}`, SQLArrayLiteral([]string{`a\b`}), "a backslash must be escaped too")
}
