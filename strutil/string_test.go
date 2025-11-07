package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertSlice(t *testing.T) {
	type StringType string

	out := ConvertSlice[StringType]([]string{"a", "b", "c"})
	require.Equal(t, []StringType{"a", "b", "c"}, out)

	out = ConvertSlice[StringType]([]string(nil))
	require.Nil(t, out)
}
