package strutil

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"
)

// IsSpace reports whether r is a space character as defined by Unicode
// or zero width space '\u200b'.
func IsSpace(r rune) bool {
	return unicode.IsSpace(r) || r == '\u200b'
}

// isTrimRune reports whether r should be trimmed by TrimSpace and
// TrimSpaceBytes: any Unicode whitespace, any non-printable rune
// (control, format, surrogate, private-use, unassigned \u2014 including
// zero-width characters like '\u200b', '\u200c', '\u200d', '\u2060'
// and the BOM '\ufeff'), or the replacement rune utf8.RuneError that
// strings.TrimFunc and bytes.TrimFunc emit for invalid UTF-8 bytes.
// A legitimate '\ufffd' rune in the input is indistinguishable from
// invalid UTF-8 at this layer and is treated as trimmable.
func isTrimRune(r rune) bool {
	// The zero-width runes and the BOM are the most common non-ASCII ones
	// to trim and unicode.IsPrint needs a range table search to report them
	switch r {
	case '\u200b', '\u200c', '\u200d', '\u2060', '\ufeff':
		return true
	}
	return r == utf8.RuneError || unicode.IsSpace(r) || !unicode.IsPrint(r)
}

// isTrimByte is the ASCII fast path of isTrimRune: every byte up to and
// including the space is whitespace or a control character, and so is
// the delete character. All other ASCII bytes are printable.
func isTrimByte(c byte) bool {
	return c <= ' ' || c == 0x7f
}

// TrimSpace returns a slice of the string s with all leading and trailing
// trimmable runes removed: any Unicode whitespace, any non-printable rune
// (including zero-width characters and the BOM), and any byte that does
// not decode as valid UTF-8.
func TrimSpace[S ~string](s S) S {
	// Fast path for ASCII like the one of strings.TrimSpace, falling back
	// to the rune based isTrimRune as soon as a non-ASCII byte is found.
	// Decoding every leading and trailing rune to call isTrimRune through
	// a function value is several times slower than these byte compares.
	start := 0
	for ; start < len(s); start++ {
		c := s[start]
		if c >= utf8.RuneSelf {
			return S(strings.TrimFunc(string(s[start:]), isTrimRune))
		}
		if !isTrimByte(c) {
			break
		}
	}
	stop := len(s)
	for ; stop > start; stop-- {
		c := s[stop-1]
		if c >= utf8.RuneSelf {
			return S(strings.TrimFunc(string(s[start:stop]), isTrimRune))
		}
		if !isTrimByte(c) {
			break
		}
	}
	return s[start:stop]
}

// TrimSpaceBytes returns a slice of the bytes string s with all leading
// and trailing trimmable runes removed: any Unicode whitespace, any
// non-printable rune (including zero-width characters and the BOM), and
// any byte that does not decode as valid UTF-8.
func TrimSpaceBytes(s []byte) []byte {
	// Fast path for ASCII, see TrimSpace
	start := 0
	for ; start < len(s); start++ {
		c := s[start]
		if c >= utf8.RuneSelf {
			return bytes.TrimFunc(s[start:], isTrimRune)
		}
		if !isTrimByte(c) {
			break
		}
	}
	stop := len(s)
	for ; stop > start; stop-- {
		c := s[stop-1]
		if c >= utf8.RuneSelf {
			return bytes.TrimFunc(s[start:stop], isTrimRune)
		}
		if !isTrimByte(c) {
			break
		}
	}
	return s[start:stop]
}

// CutTrimSpace slices s around the first instance of sep,
// returning the text before and after sep with all leading and
// trailing trimmable runes removed (see TrimSpace for the predicate).
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

// TruncateTrimSpace truncates s to maxLen runes and trims leading and
// trailing trimmable runes (see TrimSpace for the predicate).
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
