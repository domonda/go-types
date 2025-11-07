package strutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrimSpace(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		s    string
		want string
	}{
		{s: "\u200b\tZERO WIDTH SPACE\r\n", want: "ZERO WIDTH SPACE"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := TrimSpace(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrimSpace(%#v) = %#v, want %#v", tt.s, got, tt.want)
			}
		})
	}
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
