package strutil

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{s: "", want: ""},
		{s: "   ", want: ""},
		{s: "hello", want: "hello"},
		{s: "  hello  ", want: "hello"},
		{s: "\u200b\tZERO WIDTH SPACE\r\n", want: "ZERO WIDTH SPACE"},
		// Non-printable control characters
		{s: "\x00\x01hello\x7f", want: "hello"},
		// BOM, zero width joiner / non-joiner, word joiner
		{s: "\ufeff\u200chello\u200d\u2060", want: "hello"},
		// Non-breaking space and line/paragraph separators
		{s: "\u00a0\u2028hello\u2029\u00a0", want: "hello"},
		// Invalid UTF-8 bytes at both ends
		{s: "\xff\xfehello\xff\xc0", want: "hello"},
		// Mixed whitespace, invalid UTF-8 and non-visual runes
		{s: "\xff \u200b\t hello world\u200b\xff\n", want: "hello world"},
		// Only trimmable runes
		{s: "\u200b\u200c\u200d\ufeff\xff", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := TrimSpace(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrimSpace(%#v) = %#v, want %#v", tt.s, got, tt.want)
			}
			if got := string(TrimSpaceBytes([]byte(tt.s))); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrimSpaceBytes(%#v) = %#v, want %#v", tt.s, got, tt.want)
			}
		})
	}
}

// TestTrimSpaceEqualsTrimFunc checks that the ASCII fast path of TrimSpace
// and TrimSpaceBytes trims exactly what the isTrimRune predicate reports,
// which is what they did before the fast path was added. The fast path
// duplicates the predicate for ASCII bytes, so nothing but a test keeps
// the two in sync.
func TestTrimSpaceEqualsTrimFunc(t *testing.T) {
	check := func(s string) {
		t.Helper()
		want := strings.TrimFunc(s, isTrimRune)
		if got := TrimSpace(s); got != want {
			t.Errorf("TrimSpace(%q) = %q, want %q", s, got, want)
		}
		want = string(bytes.TrimFunc([]byte(s), isTrimRune))
		if got := string(TrimSpaceBytes([]byte(s))); got != want {
			t.Errorf("TrimSpaceBytes(%q) = %q, want %q", s, got, want)
		}
	}

	// Every single byte, alone and around a rune that is kept
	for b := range 256 {
		c := string([]byte{byte(b)})
		check(c)
		check(c + "x")
		check("x" + c)
		check(c + "x" + c)
		check(c + c + "x y" + c + c)
	}

	// Random strings mixing ASCII, multi-byte runes, zero-width runes
	// and invalid UTF-8 bytes
	alphabet := []string{
		"a", "Z", "0", "-", " ", "\t", "\n", "\r", "\v", "\f", "\x00", "\x1f", "\x7f",
		"\u00e9", "\u4e16", "\u00a0", "\u200b", "\u200c", "\u200d", "\u2060", "\ufeff", "\u2028",
		"\ufffd", "\xff", "\xc0", "\xfe", "\xed\xa0\x80",
	}
	rng := rand.New(rand.NewSource(42))
	var b strings.Builder
	for range 20000 {
		b.Reset()
		for range rng.Intn(8) {
			b.WriteString(alphabet[rng.Intn(len(alphabet))])
		}
		check(b.String())
	}
}

func BenchmarkTrimSpace(b *testing.B) {
	for _, s := range []string{"12345", "  12345  ", "\t\n 12345 \r\n", "\u200b\t 12345 \ufeff"} {
		b.Run(fmt.Sprintf("%q", s), func(b *testing.B) {
			for b.Loop() {
				_ = TrimSpace(s)
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
