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

func Test_SQLArrayLiteral(t *testing.T) {
	assert.Equal(t, "{}", SQLArrayLiteral(nil), "nil slice yields empty array, not NULL")
	assert.Equal(t, "{}", SQLArrayLiteral([]string{}), "empty slice yields empty array")
	assert.Equal(t, `{"a","b"}`, SQLArrayLiteral([]string{"a", "b"}))
}
