package notnull

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TrimmedString_Set(t *testing.T) {
	var s TrimmedString
	s.Set("  hello  ")
	assert.Equal(t, TrimmedString("hello"), s)
}

func Test_TrimmedString_FromAndFmt(t *testing.T) {
	assert.Equal(t, TrimmedString("hi"), TrimmedStringFrom("  hi\t"))
	assert.Equal(t, TrimmedString("a=1"), TrimmedStringf(" a=%d ", 1))
}

func Test_TrimmedString_IsEmpty(t *testing.T) {
	assert.True(t, TrimmedString("").IsEmpty())
	assert.True(t, TrimmedString("   ").IsEmpty())
	assert.False(t, TrimmedString("x").IsEmpty())
	assert.False(t, TrimmedString("  x  ").IsEmpty())
}

func Test_TrimmedString_Value(t *testing.T) {
	v, err := TrimmedString("  abc  ").Value()
	assert.NoError(t, err)
	assert.Equal(t, "abc", v)
}

func Test_TrimmedString_Scan(t *testing.T) {
	var s TrimmedString
	assert.NoError(t, s.Scan("  hi  "))
	assert.Equal(t, TrimmedString("hi"), s)

	assert.NoError(t, s.Scan([]byte("  bye  ")))
	assert.Equal(t, TrimmedString("bye"), s)
}

func Test_TrimmedString_JSON(t *testing.T) {
	b, err := json.Marshal(TrimmedString(" hello "))
	assert.NoError(t, err)
	assert.Equal(t, `"hello"`, string(b))

	var s TrimmedString
	assert.NoError(t, json.Unmarshal([]byte(`" world "`), &s))
	assert.Equal(t, TrimmedString("world"), s)
}

func Test_TrimmedString_Join(t *testing.T) {
	got := TrimmedStringJoin(", ", " a ", "b", "  c  ")
	assert.Equal(t, TrimmedString("a, b, c"), got)
}

func Test_TrimmedString_StringMethods(t *testing.T) {
	s := TrimmedString("  Hello World  ")
	assert.Equal(t, TrimmedString("HELLO WORLD"), s.ToUpper())
	assert.Equal(t, TrimmedString("hello world"), s.ToLower())
	assert.True(t, s.Contains("World"))
	assert.True(t, s.HasPrefix("Hello"))
	assert.True(t, s.HasSuffix("World"))
}
