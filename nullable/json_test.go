package nullable

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ sql.Scanner    = (*JSON)(nil)
	_ driver.Valuer  = JSON(nil)
	_ json.Marshaler = JSON(nil)
)

func Test_MarshalJSON(t *testing.T) {
	j, err := MarshalJSON(map[string]int{"a": 1})
	require.NoError(t, err)
	assert.Equal(t, JSON(`{"a":1}`), j)

	j, err = MarshalJSON(nil)
	require.NoError(t, err)
	assert.Equal(t, JSON("null"), j)

	_, err = MarshalJSON(make(chan int))
	assert.Error(t, err)
}

func Test_JSON_IsNull(t *testing.T) {
	assert.True(t, JSON(nil).IsNull())
	assert.False(t, JSON("").IsNull(), "non-nil empty JSON is not null")
	assert.False(t, JSON("{}").IsNull())
}

func Test_JSON_MarshalFrom(t *testing.T) {
	var j JSON
	require.NoError(t, j.MarshalFrom(map[string]int{"x": 2}))
	assert.Equal(t, JSON(`{"x":2}`), j)

	// nil marshals to JSON "null" which becomes a nil JSON value.
	j = JSON(`{"old":1}`)
	require.NoError(t, j.MarshalFrom(nil))
	assert.Nil(t, j)

	assert.Error(t, j.MarshalFrom(make(chan int)))
}

func Test_JSON_UnmarshalTo(t *testing.T) {
	var dest map[string]int
	require.NoError(t, JSON(`{"a":3}`).UnmarshalTo(&dest))
	assert.Equal(t, map[string]int{"a": 3}, dest)

	assert.Error(t, JSON(`not json`).UnmarshalTo(&dest))
}

func Test_JSON_MarshalJSON(t *testing.T) {
	b, err := JSON(nil).MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, "null", string(b))

	b, err = JSON(`{"a":1}`).MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, `{"a":1}`, string(b))

	// Round-trip through encoding/json.
	type wrapper struct {
		J JSON
	}
	out, err := json.Marshal(wrapper{J: JSON(`[1,2]`)})
	require.NoError(t, err)
	assert.Equal(t, `{"J":[1,2]}`, string(out))

	out, err = json.Marshal(wrapper{})
	require.NoError(t, err)
	assert.Equal(t, `{"J":null}`, string(out))
}

func Test_JSON_UnmarshalJSON(t *testing.T) {
	var j JSON
	require.NoError(t, j.UnmarshalJSON([]byte("null")))
	assert.Nil(t, j)

	require.NoError(t, j.UnmarshalJSON(nil))
	assert.Nil(t, j)

	require.NoError(t, j.UnmarshalJSON([]byte(`{"a":1}`)))
	assert.Equal(t, JSON(`{"a":1}`), j)

	// UnmarshalJSON must copy the source bytes.
	src := []byte(`[9]`)
	require.NoError(t, j.UnmarshalJSON(src))
	src[1] = '0'
	assert.Equal(t, JSON(`[9]`), j, "UnmarshalJSON must copy the source")

	require.EqualError(t, (*JSON)(nil).UnmarshalJSON([]byte("1")), "UnmarshalJSON on nil pointer")
}

func Test_JSON_Valid(t *testing.T) {
	assert.True(t, JSON(nil).Valid(), "nil JSON is valid")
	assert.True(t, JSON(`{"a":1}`).Valid())
	assert.False(t, JSON(`{not json`).Valid())
}

func Test_JSON_Value(t *testing.T) {
	val, err := JSON(nil).Value()
	require.NoError(t, err)
	assert.Nil(t, val, "nil JSON returns SQL NULL")

	val, err = JSON(`{"a":1}`).Value()
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"a":1}`), val)
}

func Test_JSON_IsEmpty(t *testing.T) {
	assert.True(t, JSON(nil).IsEmpty())
	assert.True(t, JSON("").IsEmpty())
	assert.True(t, JSON("{}").IsEmpty())
	assert.True(t, JSON("[]").IsEmpty())
	assert.False(t, JSON(`{"a":1}`).IsEmpty())
	assert.False(t, JSON("null").IsEmpty())
}

func Test_JSON_Scan(t *testing.T) {
	var j JSON

	require.NoError(t, j.Scan(nil))
	assert.Nil(t, j)

	require.NoError(t, j.Scan("null"))
	assert.Nil(t, j)

	require.NoError(t, j.Scan(`{"a":1}`))
	assert.Equal(t, JSON(`{"a":1}`), j)

	require.NoError(t, j.Scan([]byte("null")))
	assert.Nil(t, j)

	// Scan from []byte must copy.
	src := []byte(`[7]`)
	require.NoError(t, j.Scan(src))
	src[1] = '0'
	assert.Equal(t, JSON(`[7]`), j, "Scan must copy the source bytes")

	assert.Error(t, j.Scan(123))
}

func Test_JSON_String(t *testing.T) {
	assert.Equal(t, "null", JSON(nil).String())
	assert.Equal(t, `{"a":1}`, JSON(`{"a":1}`).String())
}

func Test_JSON_GoString(t *testing.T) {
	assert.Equal(t, "nullable.JSON(`{\"a\":1}`)", JSON(`{"a":1}`).GoString())
}

func Test_JSON_PrettyPrint(t *testing.T) {
	var buf bytes.Buffer
	n, err := JSON(`{"a":1}`).PrettyPrint(&buf)
	require.NoError(t, err)
	assert.Equal(t, "`{\"a\":1}`", buf.String())
	assert.Equal(t, buf.Len(), n)

	buf.Reset()
	_, err = JSON(nil).PrettyPrint(&buf)
	require.NoError(t, err)
	assert.Equal(t, "`null`", buf.String())
}

func Test_JSON_Clone(t *testing.T) {
	assert.Nil(t, JSON(nil).Clone())

	orig := JSON(`{"a":1}`)
	clone := orig.Clone()
	assert.Equal(t, orig, clone)
	clone[1] = 'X'
	assert.NotEqual(t, orig, clone, "Clone must be independent of original")
}
