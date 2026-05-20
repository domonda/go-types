package notnull

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_JSON_MarshalJSON(t *testing.T) {
	b, err := JSON(nil).MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(b), "JSON(nil) marshals to {}")

	b, err = JSON(`{"a":1}`).MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `{"a":1}`, string(b))
}

func Test_JSON_UnmarshalJSON(t *testing.T) {
	var j JSON
	require.NoError(t, j.UnmarshalJSON([]byte(`{"x":42}`)))
	assert.Equal(t, JSON(`{"x":42}`), j)

	// Source mutation must not affect j (UnmarshalJSON copies).
	src := []byte(`{"y":1}`)
	require.NoError(t, j.UnmarshalJSON(src))
	src[1] = 'Z'
	assert.Equal(t, JSON(`{"y":1}`), j)

	var nilJSON *JSON
	assert.Error(t, nilJSON.UnmarshalJSON([]byte("{}")))
}

func Test_JSON_RoundTripViaEncodingJSON(t *testing.T) {
	type wrap struct {
		Data JSON `json:"data"`
	}

	in := wrap{Data: JSON(`{"k":"v"}`)}
	b, err := json.Marshal(in)
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":{"k":"v"}}`, string(b))

	var out wrap
	require.NoError(t, json.Unmarshal(b, &out))
	assert.JSONEq(t, `{"k":"v"}`, string(out.Data))
}

func Test_JSON_MarshalFrom(t *testing.T) {
	var j JSON
	require.NoError(t, j.MarshalFrom(map[string]int{"a": 1}))
	assert.JSONEq(t, `{"a":1}`, string(j))

	require.NoError(t, j.MarshalFrom([]int{1, 2, 3}))
	assert.Equal(t, JSON("[1,2,3]"), j)

	// Error case: channels can't be marshaled. j should be left unchanged.
	before := j
	assert.Error(t, j.MarshalFrom(make(chan int)))
	assert.Equal(t, before, j)
}

func Test_JSON_UnmarshalTo(t *testing.T) {
	var dest map[string]int
	require.NoError(t, JSON(`{"a":1,"b":2}`).UnmarshalTo(&dest))
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, dest)

	// nil JSON acts like "{}".
	dest = nil
	require.NoError(t, JSON(nil).UnmarshalTo(&dest))
	assert.Equal(t, map[string]int{}, dest)
}

func Test_JSON_Valid(t *testing.T) {
	assert.False(t, JSON(nil).Valid(), "nil JSON is not considered valid")
	assert.True(t, JSON(`{"a":1}`).Valid())
	assert.True(t, JSON(`[1,2,3]`).Valid())
	assert.True(t, JSON(`"hello"`).Valid())
	assert.False(t, JSON(`{not json}`).Valid())
}

func Test_JSON_Value(t *testing.T) {
	v, err := JSON(nil).Value()
	require.NoError(t, err)
	assert.Equal(t, []byte("{}"), v)

	v, err = JSON(`{"a":1}`).Value()
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"a":1}`), v)
}

func Test_JSON_IsEmpty(t *testing.T) {
	assert.True(t, JSON(nil).IsEmpty())
	assert.True(t, JSON("").IsEmpty())
	assert.True(t, JSON("{}").IsEmpty())
	assert.True(t, JSON("[]").IsEmpty())
	assert.False(t, JSON(`{"a":1}`).IsEmpty())
	assert.False(t, JSON("[1]").IsEmpty())
}

func Test_JSON_Scan(t *testing.T) {
	var j JSON

	require.NoError(t, j.Scan(nil))
	assert.Equal(t, JSON("{}"), j, "Scan(nil) sets JSON to {}")

	require.NoError(t, j.Scan(`{"a":1}`))
	assert.Equal(t, JSON(`{"a":1}`), j)

	src := []byte(`{"b":2}`)
	require.NoError(t, j.Scan(src))
	assert.Equal(t, JSON(`{"b":2}`), j)
	// Scan must copy the byte slice — mutating src must not change j.
	src[1] = 'Z'
	assert.Equal(t, JSON(`{"b":2}`), j)

	assert.Error(t, j.Scan(42))
}

func Test_JSON_String(t *testing.T) {
	assert.Equal(t, "{}", JSON(nil).String())
	assert.Equal(t, `{"a":1}`, JSON(`{"a":1}`).String())
}

func Test_JSON_GoString(t *testing.T) {
	assert.Equal(t, "notnull.JSON(`{\"a\":1}`)", JSON(`{"a":1}`).GoString())
}

func Test_JSON_PrettyPrint(t *testing.T) {
	var buf bytes.Buffer
	n, err := JSON(`{"a":1}`).PrettyPrint(&buf)
	require.NoError(t, err)
	assert.Equal(t, "`{\"a\":1}`", buf.String())
	assert.Equal(t, buf.Len(), n)
}

func Test_JSON_Clone(t *testing.T) {
	original := JSON(`{"a":1}`)
	clone := original.Clone()
	assert.Equal(t, original, clone)

	// Mutating the clone must not affect the original.
	clone[1] = 'Z'
	assert.NotEqual(t, original, clone)
	assert.Equal(t, JSON(`{"a":1}`), original)
}

func Test_MarshalJSON_Package(t *testing.T) {
	j, err := MarshalJSON(map[string]int{"a": 1})
	require.NoError(t, err)
	assert.JSONEq(t, `{"a":1}`, string(j))

	_, err = MarshalJSON(make(chan int))
	assert.Error(t, err)
}
