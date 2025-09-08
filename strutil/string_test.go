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

func TestTruncate(t *testing.T) {
	tests := []struct {
		str    string
		maxLen int
		want   string
	}{
		{str: "", maxLen: -1, want: ""},
		{str: "", maxLen: 0, want: ""},
		{str: "", maxLen: 1, want: ""},
		{str: "", maxLen: 2, want: ""},

		{str: "abc", maxLen: -1, want: ""},
		{str: "abc", maxLen: 0, want: ""},
		{str: "abc", maxLen: 1, want: "a"},
		{str: "abc", maxLen: 2, want: "ab"},
		{str: "abc", maxLen: 3, want: "abc"},
		{str: "abc", maxLen: 4, want: "abc"},

		{str: "もしもし World", maxLen: -1, want: ""},
		{str: "もしもし World", maxLen: 0, want: ""},
		{str: "もしもし World", maxLen: 1, want: "も"},
		{str: "もしもし World", maxLen: 2, want: "もし"},
		{str: "もしもし World", maxLen: 3, want: "もしも"},
		{str: "もしもし World", maxLen: 4, want: "もしもし"},
		{str: "もしもし World", maxLen: 5, want: "もしもし "},
		{str: "もしもし World", maxLen: 10, want: "もしもし World"},
		{str: "もしもし World", maxLen: 100, want: "もしもし World"},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got := Truncate(tt.str, tt.maxLen)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	tests := []struct {
		str                string
		maxLenInclEllipsis int
		want               string
	}{
		{str: "", maxLenInclEllipsis: -1, want: ""},
		{str: "", maxLenInclEllipsis: 0, want: ""},
		{str: "", maxLenInclEllipsis: 1, want: ""},
		{str: "", maxLenInclEllipsis: 2, want: ""},

		{str: "abc", maxLenInclEllipsis: -1, want: ""},
		{str: "abc", maxLenInclEllipsis: 0, want: ""},
		{str: "abc", maxLenInclEllipsis: 1, want: "…"},
		{str: "abc", maxLenInclEllipsis: 2, want: "a…"},
		{str: "abc", maxLenInclEllipsis: 3, want: "abc"},
		{str: "abc", maxLenInclEllipsis: 4, want: "abc"},

		{str: "もしもし World", maxLenInclEllipsis: -1, want: ""},
		{str: "もしもし World", maxLenInclEllipsis: 0, want: ""},
		{str: "もしもし World", maxLenInclEllipsis: 1, want: "…"},
		{str: "もしもし World", maxLenInclEllipsis: 2, want: "も…"},
		{str: "もしもし World", maxLenInclEllipsis: 3, want: "もし…"},
		{str: "もしもし World", maxLenInclEllipsis: 4, want: "もしも…"},
		{str: "もしもし World", maxLenInclEllipsis: 5, want: "もしもし…"},
		{str: "もしもし World", maxLenInclEllipsis: 10, want: "もしもし World"},
		{str: "もしもし World", maxLenInclEllipsis: 100, want: "もしもし World"},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got := TruncateWithEllipsis(tt.str, tt.maxLenInclEllipsis)
			require.Equal(t, tt.want, got)
		})
	}
}
