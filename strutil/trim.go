package strutil

import (
	"bytes"
	"strings"
	"unicode"
)

// IsSpace reports whether r is a space character as defined by Unicode
// or zero width space '\u200b'.
func IsSpace(r rune) bool {
	return unicode.IsSpace(r) || r == '\u200b'
}

// TrimSpace returns a slice of the string s, with all leading
// and trailing white space removed including zero width space '\u200b'.
func TrimSpace[S ~string](s S) S {
	return S(strings.TrimFunc(string(s), IsSpace))
}

// TrimSpaceBytes returns a slice of the bytes string s, with all leading
// and trailing white space removed including zero width space '\u200b'.
func TrimSpaceBytes(s []byte) []byte {
	return bytes.TrimFunc(s, IsSpace)
}

// CutTrimSpace slices s around the first instance of sep,
// returning the text before and after sep with all leading
// and trailing white space removed, as defined by Unicode.
// The found result reports whether sep appears in s.
// If sep does not appear in s, cut returns s, "", false.
func CutTrimSpace[S ~string](s, sep S) (before, after S, found bool) {
	if i := strings.Index(string(s), string(sep)); i >= 0 {
		return TrimSpace(s[:i]), TrimSpace(s[i+len(sep):]), true
	}
	return s, "", false
}

// Truncate returns a truncated version of s with at most maxLen runes.
// If s has more than maxLen runes, it returns the first maxLen runes.
// The truncation is UTF-8 aware and will not split multi-byte characters.
func Truncate[S ~string](s S, maxLen int) S {
	numRunes := 0
	for byteIndex := range s {
		numRunes++
		if numRunes > maxLen {
			return s[:byteIndex]
		}
	}
	return s
}

// TruncateTrimSpace truncates s to maxLen runes and trims leading and trailing whitespace.
func TruncateTrimSpace[S ~string](s S, maxLen int) S {
	return TrimSpace(Truncate(s, maxLen))
}

// TruncateWithEllipsis truncates s to maxLenInclEllipsis runes, appending an ellipsis (…) if truncated.
// The returned string will have at most maxLenInclEllipsis runes including the ellipsis.
// If maxLenInclEllipsis is 0 or negative, returns an empty string.
func TruncateWithEllipsis[S ~string](s S, maxLenInclEllipsis int) S {
	if maxLenInclEllipsis <= 0 {
		return ""
	}
	numRunes := 0
	lastByteIndex := 0
	for byteIndex := range s {
		numRunes++
		if numRunes > maxLenInclEllipsis {
			return s[:lastByteIndex] + "…"
		}
		lastByteIndex = byteIndex
	}
	return s
}
