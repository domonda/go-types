// Package xmlattrtest checks how encoding/xml writes a type as attribute.
//
// It is only imported from _test.go files.
package xmlattrtest

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ptrDoc[T any] struct {
	XMLName xml.Name `xml:"doc"`
	Attr    *T       `xml:"attr,attr"`
}

type valueDoc[T any] struct {
	XMLName xml.Name `xml:"doc"`
	Attr    T        `xml:"attr,attr"`
}

// NilPointer checks that encoding/xml leaves out a nil *T attribute field
// without omitempty, and that a non-nil *T attribute field writes each of
// values like a T field does. encoding/xml calls a value receiver
// MarshalText through a nil pointer attribute field and panics, unless *T
// implements xml.MarshalerAttr, which encoding/xml checks first.
func NilPointer[T any](t *testing.T, values ...T) {
	t.Helper()
	var data []byte
	var err error
	require.NotPanics(t, func() { data, err = xml.Marshal(ptrDoc[T]{}) }, "nil *%T attribute", *new(T))
	require.NoError(t, err)
	assert.Equal(t, `<doc></doc>`, string(data), "nil *%T attribute", *new(T))

	for _, v := range values {
		want, err := xml.Marshal(valueDoc[T]{Attr: v})
		require.NoError(t, err)
		got, err := xml.Marshal(ptrDoc[T]{Attr: &v})
		require.NoError(t, err)
		assert.Equal(t, string(want), string(got), "*%T attribute %#v", v, v)
	}
}
