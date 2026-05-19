package notnull

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StringArray_Value(t *testing.T) {
	val, err := StringArray(nil).Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val, "StringArray(nil).Value() returns empty SQL array")

	val, err = StringArray{}.Value()
	assert.NoError(t, err)
	assert.Equal(t, "{}", val, "empty StringArray returns empty SQL array")

	val, err = StringArray{"a", "b", "c"}.Value()
	assert.NoError(t, err)
	assert.Equal(t, `{"a","b","c"}`, val)
}

func Test_StringArray_Scan(t *testing.T) {
	var a StringArray
	assert.NoError(t, a.Scan(nil))
	assert.Empty(t, a)

	assert.NoError(t, a.Scan("{x,y,z}"))
	assert.Equal(t, StringArray{"x", "y", "z"}, a)

	assert.NoError(t, a.Scan([]byte(`{"a","b"}`)))
	assert.Equal(t, StringArray{"a", "b"}, a)
}

func Test_StringArray_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(StringArray(nil))
	assert.NoError(t, err)
	assert.Equal(t, "[]", string(b))

	b, err = json.Marshal(StringArray{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, `["a","b"]`, string(b))
}

func Test_StringArray_Contains(t *testing.T) {
	a := StringArray{"a", "b", "c"}
	assert.True(t, a.Contains("b"))
	assert.False(t, a.Contains("z"))
}
