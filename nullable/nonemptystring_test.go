package nullable

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ sql.Scanner    = (*NonEmptyString)(nil)
	_ driver.Valuer  = NonEmptyString("")
	_ json.Marshaler = NonEmptyString("")
)

func Test_NonEmptyStringf(t *testing.T) {
	assert.Equal(t, NonEmptyString("a=1"), NonEmptyStringf("a=%d", 1))
	assert.Equal(t, NonEmptyStringNull, NonEmptyStringf("%s", ""))
}

func Test_NonEmptyStringFromPtr(t *testing.T) {
	assert.Equal(t, NonEmptyStringNull, NonEmptyStringFromPtr(nil))
	s := "hello"
	assert.Equal(t, NonEmptyString("hello"), NonEmptyStringFromPtr(&s))
}

func Test_NonEmptyStringFromError(t *testing.T) {
	assert.Equal(t, NonEmptyStringNull, NonEmptyStringFromError(nil))
	assert.Equal(t, NonEmptyString("boom"), NonEmptyStringFromError(errors.New("boom")))
}

func Test_NonEmptyStringTrimSpace(t *testing.T) {
	assert.Equal(t, NonEmptyString("hello"), NonEmptyStringTrimSpace("  hello \n"))
	assert.Equal(t, NonEmptyStringNull, NonEmptyStringTrimSpace("   "))
}

func Test_JoinNonEmptyStrings(t *testing.T) {
	assert.Equal(t, NonEmptyString("a,b,c"), JoinNonEmptyStrings(",", "a", "b", "c"))
	assert.Equal(t, NonEmptyString("a,c"), JoinNonEmptyStrings(",", "a", "", "c"))
	assert.Equal(t, NonEmptyStringNull, JoinNonEmptyStrings(",", "", ""))
}

func Test_NonEmptyString_Ptr(t *testing.T) {
	assert.Nil(t, NonEmptyStringNull.Ptr())
	p := NonEmptyString("x").Ptr()
	require.NotNil(t, p)
	assert.Equal(t, "x", *p)
}

func Test_NonEmptyString_IsNull(t *testing.T) {
	assert.True(t, NonEmptyString("").IsNull())
	assert.False(t, NonEmptyString("x").IsNull())
	assert.False(t, NonEmptyString("").IsNotNull())
	assert.True(t, NonEmptyString("x").IsNotNull())
}

func Test_NonEmptyString_TrimSpace(t *testing.T) {
	assert.Equal(t, NonEmptyString("x"), NonEmptyString("  x ").TrimSpace())
	assert.Equal(t, NonEmptyStringNull, NonEmptyString("  ").TrimSpace())
}

func Test_NonEmptyString_StringOr(t *testing.T) {
	assert.Equal(t, "fallback", NonEmptyString("").StringOr("fallback"))
	assert.Equal(t, "x", NonEmptyString("x").StringOr("fallback"))
}

func Test_NonEmptyString_Get(t *testing.T) {
	assert.Panics(t, func() { NonEmptyString("").Get() })
	assert.Equal(t, "x", NonEmptyString("x").Get())
}

func Test_NonEmptyString_SetAndSetNull(t *testing.T) {
	var n NonEmptyString
	n.Set("hello")
	assert.Equal(t, NonEmptyString("hello"), n)
	n.SetNull()
	assert.True(t, n.IsNull())
}

func Test_NonEmptyString_Scan(t *testing.T) {
	var n NonEmptyString

	require.NoError(t, n.Scan(nil))
	assert.True(t, n.IsNull())

	require.NoError(t, n.Scan("hello"))
	assert.Equal(t, NonEmptyString("hello"), n)

	require.NoError(t, n.Scan([]byte("bytes")))
	assert.Equal(t, NonEmptyString("bytes"), n)

	assert.Error(t, n.Scan(""))
	assert.Error(t, n.Scan([]byte{}))
	assert.Error(t, n.Scan(123))
}

func Test_NonEmptyString_Value(t *testing.T) {
	val, err := NonEmptyString("").Value()
	require.NoError(t, err)
	assert.Nil(t, val, "empty NonEmptyString returns SQL NULL")

	val, err = NonEmptyString("x").Value()
	require.NoError(t, err)
	assert.Equal(t, "x", val)
}

func Test_NonEmptyString_UnmarshalText(t *testing.T) {
	var n NonEmptyString
	require.NoError(t, n.UnmarshalText([]byte("hello")))
	assert.Equal(t, NonEmptyString("hello"), n)
}

func Test_NonEmptyString_RoundTrip_SQL(t *testing.T) {
	original := NonEmptyString("round-trip")
	val, err := original.Value()
	require.NoError(t, err)
	var scanned NonEmptyString
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, original, scanned)
}

func TestNonEmptyString_MarshalJSON(t *testing.T) {
	type Struct struct {
		Null     NonEmptyString
		NullOmit NonEmptyString `json:",omitempty"`
		NonNull  NonEmptyString
	}
	s := Struct{NonNull: "Hello"}

	j, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.Equal(t, `{"Null":null,"NonNull":"Hello"}`, string(j))
}

func TestNonEmptyString_UnmarshalJSON(t *testing.T) {
	type Struct struct {
		Null     NonEmptyString
		NullOmit NonEmptyString `json:",omitempty"`
		Empty    NonEmptyString
		NonNull  NonEmptyString
	}
	input := `{
		"Null": null,
		"NullOmit": "here",
		"Empty": "",
		"NonNull": "Hello"
	}`
	expected := Struct{NullOmit: "here", NonNull: "Hello"}
	var result Struct

	err := json.Unmarshal([]byte(input), &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_NonEmptyString_UnmarshalJSON_Direct(t *testing.T) {
	// JSON null resets a reused (non-null) value to the null state.
	n := NonEmptyString("old")
	assert.NoError(t, n.UnmarshalJSON([]byte("null")))
	assert.Equal(t, NonEmptyString(""), n)
	assert.True(t, n.IsNull())

	assert.NoError(t, n.UnmarshalJSON([]byte(`"hello"`)))
	assert.Equal(t, NonEmptyString("hello"), n)

	// Direct call: encoding/json pre-validates structure, so a
	// malformed document must be passed to UnmarshalJSON directly.
	assert.Error(t, n.UnmarshalJSON([]byte(`"unterminated`)))
}
