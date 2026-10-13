package strutil

import (
	"bytes"
	"strings"
	"testing"
	"unicode"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IsWordSeparator(t *testing.T) {
	for _, r := range `.,;:&*|()[]{}<>#$@'"!? =_-+/\~` + "`" {
		if !IsWordSeparator(r) {
			t.Fatalf("isWordSeparator('%s') == false", string(r))
		}
	}
}

var splitAndTrimIndexTable = map[string][][]int{
	"":                               nil,
	".":                              nil,
	". .":                            nil,
	" . ":                            nil,
	" .. ":                           nil,
	"x.y":                            {{0, 3}},
	"x.y .x":                         {{0, 3}, {5, 6}},
	"HelloWorld":                     {{0, 10}},
	"HelloWorld ":                    {{0, 10}},
	"HelloWorld  ":                   {{0, 10}},
	"HelloWorld. ":                   {{0, 10}},
	"HelloWorld.. ":                  {{0, 10}},
	"HelloWorld.  .":                 {{0, 10}},
	" HelloWorld":                    {{1, 11}},
	" .HelloWorld":                   {{2, 12}},
	". .HelloWorld":                  {{3, 13}},
	"...HelloWorld":                  {{3, 13}},
	"Hello World":                    {{0, 5}, {6, 11}},
	"Hello.World":                    {{0, 11}},
	"Hello  World":                   {{0, 5}, {7, 12}},
	"Hello.  World":                  {{0, 5}, {8, 13}},
	"Hello.. World":                  {{0, 5}, {8, 13}},
	"Hello...World":                  {{0, 13}},
	"Hello. .World":                  {{0, 5}, {8, 13}},
	" Hello. .World":                 {{1, 6}, {9, 14}},
	" Hello. .World ":                {{1, 6}, {9, 14}},
	" Hello. .World  ":               {{1, 6}, {9, 14}},
	" Hello. .World. ":               {{1, 6}, {9, 14}},
	" Hello. .World.. ":              {{1, 6}, {9, 14}},
	"one two three four 5":           {{0, 3}, {4, 7}, {8, 13}, {14, 18}, {19, 20}},
	"one two three four 5.":          {{0, 3}, {4, 7}, {8, 13}, {14, 18}, {19, 20}},
	"one two three four 5  ":         {{0, 3}, {4, 7}, {8, 13}, {14, 18}, {19, 20}},
	".one. .two. .three. .four. .5":  {{1, 4}, {7, 10}, {13, 18}, {21, 25}, {28, 29}},
	".one. .two. .three. .four. .5.": {{1, 4}, {7, 10}, {13, 18}, {21, 25}, {28, 29}},
}

func Test_SplitAndTrimIndex(t *testing.T) {
	for str, refIndices := range splitAndTrimIndexTable {
		indices := SplitAndTrimIndex([]byte(str), unicode.IsSpace, unicode.IsPunct)
		if len(indices) != len(refIndices) {
			var words []string
			for i := range indices {
				words = append(words, "'"+str[indices[i][0]:indices[i][1]]+"'")
			}
			t.Errorf("SplitAndTrimIndex('%s') expected %d words but got %d: %s", str, len(refIndices), len(indices), strings.Join(words, " "))
		} else {
			for i := range indices {
				if indices[i][0] != refIndices[i][0] || indices[i][1] != refIndices[i][1] {
					// t.Error(i, indices[i], refIndices[i], len(str))
					t.Errorf("SplitAndTrimIndex('%s') word %d expected %v '%s' but got %v '%s'", str, i, refIndices[i], str[refIndices[i][0]:refIndices[i][1]], indices[i], str[indices[i][0]:indices[i][1]])
				}
			}
		}
	}
}

func Test_SanitizeFileName(t *testing.T) {
	filenameTable := map[string]string{
		"": "_",

		"image.JpG": "image.jpeg",
		"image.Tif": "image.tiff",

		"/var/log/file.txt":  "var-log-file.txt",
		"Hello World!":       "Hello_World-",
		"Hello World!!!":     "Hello_World-",
		"Hello World!!!.jpg": "Hello_World-.jpeg",
		"-500_600x100-":      "-500_600x100-",
		"../Back\\Path":      "Back-Path",
		"Nix__da~!6%+^?.":    "Nix__da-6-",
	}

	for filename, expected := range filenameTable {
		result := SanitizeFileName(filename)
		if result != expected {
			t.Errorf("SanitizeFileName('%s') returned '%s', expected '%s'", filename, result, expected)
		}
	}
}

func Test_MakeValidFileName(t *testing.T) {
	filenameTable := map[string]string{
		"":             "_",
		"image.jpeg":   "image.jpeg",
		"Hello World!": "Hello World!",

		"../Back\\Path":                    ".._Back_Path",
		"\nHello/Darkness<my>old\\Friend:": "Hello_Darkness_my_old_Friend",
		":nix>":                            "nix",
	}

	for filename, expected := range filenameTable {
		result := MakeValidFileName(filename)
		if result != expected {
			t.Errorf("MakeValidFileName('%s') returned '%s', expected '%s'", filename, result, expected)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	testCases := map[string]string{
		"":                     "",
		"_":                    "_",
		" ":                    "_",
		"  ":                   "__",
		"\tX\n":                "_x_",
		"already_snake_case":   "already_snake_case",
		"_already_snake_case_": "_already_snake_case_",
		"HelloWorld":           "hello_world",
		"Hello World":          "hello_world",
		"Hello-World":          "hello_world",
		"*Hello+World*":        "_hello_world_",
		"Hello.World":          "hello_world",
		"Hello/World":          "hello_world",
		"(Hello World!)":       "_hello_world__",
		"DocumentID":           "document_id",
		"HTMLHandler":          "htmlhandler",
		"Straßenadresse":       "straßenadresse",
		"もしもしWorld":            "もしもし_world",
	}
	for str, expected := range testCases {
		t.Run(str, func(t *testing.T) {
			actual := ToSnakeCase(str)
			require.Equal(t, expected, actual, "snake case")
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Common European special characters are transliterated instead
		// of being dropped so that the name stays readable
		{name: "transliteration", args: args{"Grüße an Österreich.PDF"}, want: "Gruesse_an_Oesterreich.pdf"},
		// File names from URLs and multipart uploads are often query
		// escaped, they are unescaped before sanitizing
		{name: "query escaped", args: args{"My%20Invoice%3A2024.pdf"}, want: "My_Invoice-2024.pdf"},
		{name: "plus means space", args: args{"a+b.txt"}, want: "a_b.txt"},
		// An invalid escape sequence keeps the name as is
		{name: "invalid escape kept", args: args{"100%.txt"}, want: "100-.txt"},
		// No path separators or traversal may survive
		{name: "path traversal", args: args{"../../etc/passwd"}, want: "etc-passwd"},
		// Never returns an empty string, even if nothing valid is left
		{name: "only invalid characters", args: args{"***"}, want: "_"},
		{name: "escaped space only", args: args{"%20"}, want: "_"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeFileName(tt.args.name); got != tt.want {
				t.Errorf("SanitizeFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceTransliterationsMaxLen(t *testing.T) {
	tests := []struct {
		str    string
		maxLen int
		want   string
	}{
		{"", -1, ""},
		{"Österreich", -1, "Oesterreich"},
		{"Österreich", 3, "Oes"},
		{"Öster\uFFFDeich", 100, "Oestereich"},
		{"Öster\uFFFDeich", -1, "Oestereich"},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			if got := TransliterateSpecialCharactersMaxLen(tt.str, tt.maxLen); got != tt.want {
				t.Errorf("ReplaceTransliterationsMaxLen(%#v) = %#v, want %#v", tt.str, got, tt.want)
			}
		})
	}
}

func TestTransliterateSpecialCharactersMaxLen(t *testing.T) {
	type args struct {
		str    string
		maxLen int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{"", 10}, want: ""},
		// Zero means zero runes, not "no limit"
		{name: "zero maxLen", args: args{"Österreich", 0}, want: ""},
		{name: "unchanged ASCII", args: args{"Hello", -1}, want: "Hello"},
		{name: "sharp s", args: args{"Straße", -1}, want: "Strasse"},
		{name: "Scandinavian", args: args{"Ærøskøbing", -1}, want: "AEroeskoebing"},
		// Characters missing from the transliteration table are kept
		{name: "unknown special character kept", args: args{"Łódź", -1}, want: "Lodź"},
		// maxLen counts runes, not bytes, so multi-byte text is not cut
		// in the middle of a rune
		{name: "maxLen counts runes", args: args{"もしもし", 2}, want: "もし"},
		{name: "maxLen with transliteration that fits", args: args{"Öl", 2}, want: "Oe"},
		{name: "maxLen larger than result", args: args{"Öl", 10}, want: "Oel"},
		// The result must be valid UTF-8, invalid bytes are dropped
		{name: "invalid byte in the middle", args: args{"ab\xffc", -1}, want: "abc"},
		{name: "invalid byte at the end", args: args{"abc\xff", -1}, want: "abc"},
		// A dropped first byte writes nothing, the result must still
		// be the cleaned string and not the invalid input unchanged
		{name: "invalid byte at the start", args: args{"\xe4pfel", -1}, want: "pfel"},
		{name: "only invalid bytes before valid ones", args: args{"\xff\xffxy", -1}, want: "xy"},
		{name: "invalid byte at the start with maxLen", args: args{"\xffabc", 2}, want: "ab"},
		// Dropped invalid bytes don't count towards maxLen
		{name: "invalid byte with maxLen", args: args{"ab\xffcd", 3}, want: "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TransliterateSpecialCharactersMaxLen(tt.args.str, tt.args.maxLen); got != tt.want {
				t.Errorf("TransliterateSpecialCharactersMaxLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

// sanitizeLineEndingsReference is the original three-pass implementation,
// used as a behavioral oracle for the optimized single-pass versions.
func sanitizeLineEndingsReference(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n\r", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

var sanitizeLineEndingsCases = []struct {
	name string
	in   string
	want string
}{
	{name: "empty", in: "", want: ""},
	{name: "no_line_endings", in: "ABC", want: "ABC"},
	{name: "only_lf", in: "\n", want: "\n"},
	{name: "only_cr", in: "\r", want: "\n"},
	{name: "crlf", in: "\r\n", want: "\n"},
	{name: "lfcr", in: "\n\r", want: "\n"},
	{name: "crlfcr", in: "\r\n\r", want: "\n"},
	{name: "crcrlf", in: "\r\r\n", want: "\n\n"},
	{name: "crlfcrlf", in: "\r\n\r\n", want: "\n\n"},
	{name: "lfcrlf", in: "\n\r\n", want: "\n\n"},
	{name: "lfcrcr", in: "\n\r\r", want: "\n\n"},
	{name: "lfcrlfcr", in: "\n\r\n\r", want: "\n\n"},
	{name: "lfcrcrlf", in: "\n\r\r\n", want: "\n\n"},
	{name: "crcrlfcr", in: "\r\r\n\r", want: "\n\n"},
	{name: "crlfcrcr", in: "\r\n\r\r", want: "\n\n"},
	{name: "crlflfcr", in: "\r\n\n\r", want: "\n\n"},
	{name: "crlflfcrlf", in: "\r\n\n\r\n", want: "\n\n\n"},
	{name: "crlfcrlfcr", in: "\r\n\r\n\r", want: "\n\n"},
	{name: "lflfcr", in: "\n\n\r", want: "\n\n"},
	{name: "lfcrlfcr_text", in: "ABC\r\nDEF\rGHI\nJKL\n\r\r\nMNO\r\n\rPQR", want: "ABC\nDEF\nGHI\nJKL\n\nMNO\nPQR"},
	{name: "only_three_cr", in: "\r\r\r", want: "\n\n\n"},
	{name: "leading_cr_text", in: "\rABC", want: "\nABC"},
	{name: "trailing_crlf", in: "ABC\r\n", want: "ABC\n"},
	{name: "interspersed_cr", in: "\rA\r", want: "\nA\n"},
	{name: "no_cr_with_lf", in: "ABC\nDEF\n", want: "ABC\nDEF\n"},
	{name: "long_no_cr", in: strings.Repeat("Hello\n", 10), want: strings.Repeat("Hello\n", 10)},
	{name: "long_mixed", in: strings.Repeat("Hello\r\nWorld\r", 5), want: strings.Repeat("Hello\nWorld\n", 5)},
	{name: "utf8_with_cr", in: "Größe\r\nstraße\rÜber", want: "Größe\nstraße\nÜber"},
}

func TestSanitizeLineEndings(t *testing.T) {
	for _, tc := range sanitizeLineEndingsCases {
		t.Run(tc.name, func(t *testing.T) {
			// Sanity check that the expected value matches the original
			// three-pass semantics — guards against typos in the table.
			require.Equal(t, tc.want, sanitizeLineEndingsReference(tc.in), "reference oracle mismatch")
			require.Equal(t, tc.want, SanitizeLineEndings(tc.in))
		})
	}
}

func TestSanitizeLineEndingsBytes(t *testing.T) {
	for _, tc := range sanitizeLineEndingsCases {
		t.Run(tc.name, func(t *testing.T) {
			in := []byte(tc.in)
			got := SanitizeLineEndingsBytes(in)
			require.Equal(t, tc.want, string(got))
			// Ensure the input slice is never mutated.
			require.Equal(t, tc.in, string(in), "input slice was mutated")
		})
	}
}

func TestSanitizeLineEndingsBytes_NoCRReturnsSameBacking(t *testing.T) {
	in := []byte("no carriage returns here\n")
	got := SanitizeLineEndingsBytes(in)
	// Fast path returns the original slice unchanged.
	require.True(t, bytes.Equal(in, got))
	require.Equal(t, &in[0], &got[0], "fast path should return original backing array")
}

// FuzzSanitizeLineEndings differentially tests the optimized implementation
// against the original three-pass replacement so that any divergence on any
// byte sequence is caught.
func FuzzSanitizeLineEndings(f *testing.F) {
	for _, tc := range sanitizeLineEndingsCases {
		f.Add(tc.in)
	}
	f.Fuzz(func(t *testing.T, s string) {
		want := sanitizeLineEndingsReference(s)
		if got := SanitizeLineEndings(s); got != want {
			t.Fatalf("SanitizeLineEndings(%q) = %q, want %q", s, got, want)
		}
		if got := SanitizeLineEndingsBytes([]byte(s)); string(got) != want {
			t.Fatalf("SanitizeLineEndingsBytes(%q) = %q, want %q", s, string(got), want)
		}
	})
}

func BenchmarkSanitizeLineEndings(b *testing.B) {
	inputs := map[string]string{
		"no_cr":    strings.Repeat("Hello, World!\n", 100),
		"mixed":    strings.Repeat("Hello\r\nWorld\rFoo\n\r", 100),
		"all_crlf": strings.Repeat("line\r\n", 200),
	}
	for name, in := range inputs {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(in)))
			for i := 0; i < b.N; i++ {
				_ = SanitizeLineEndings(in)
			}
		})
	}
}

func TestIsRuneFuncs(t *testing.T) {
	isAOrB := IsRune('a', 'b')
	assert.True(t, isAOrB('a'))
	assert.True(t, isAOrB('b'))
	assert.False(t, isAOrB('c'))
	assert.False(t, IsRune()('a'), "IsRune without runes matches nothing")
	assert.True(t, IsRune('ü', '世')('世'), "multi-byte runes")

	notAOrB := IsRuneInverse(isAOrB)
	assert.False(t, notAOrB('a'))
	assert.True(t, notAOrB('c'))

	isLowerVowel := IsRuneAll(unicode.IsLower, IsRune('a', 'e', 'i', 'o', 'u', 'A'))
	assert.True(t, isLowerVowel('e'))
	assert.False(t, isLowerVowel('A'), "upper case fails IsLower")
	assert.False(t, isLowerVowel('x'), "consonant fails IsRune")

	isDigitOrDot := IsRuneAny(unicode.IsDigit, IsRune('.'))
	assert.True(t, isDigitOrDot('7'))
	assert.True(t, isDigitOrDot('.'))
	assert.False(t, isDigitOrDot(','))

	// Without functions the combinators return the identity of their
	// logical operation so that they can be built up from dynamic lists:
	// an empty AND is true, an empty OR is false
	assert.True(t, IsRuneAll()('x'))
	assert.False(t, IsRuneAny()('x'))
}

func TestIsNorLetterOrDigit(t *testing.T) {
	// Letters and digits of all scripts count, not only ASCII
	for _, r := range []rune{'a', 'Z', '0', '9', 'ü', 'ß', '世', 'Ω', '٣'} {
		assert.False(t, IsNorLetterOrDigit(r), "%q is a letter or digit", r)
	}
	for _, r := range []rune{' ', '\t', '.', '-', '_', '€', '\u00a0', '\u200b'} {
		assert.True(t, IsNorLetterOrDigit(r), "%q is neither letter nor digit", r)
	}
}

// removeKeepRunesCases only uses valid UTF-8 without U+FFFD,
// see TestRemoveKeepRunes_InvalidUTF8AndReplacementChar for those.
var removeKeepRunesCases = []struct {
	name       string
	str        string
	funcs      []IsRuneFunc
	wantRemove string
	wantKeep   string
}{
	{
		name:       "empty",
		str:        "",
		funcs:      []IsRuneFunc{unicode.IsDigit},
		wantRemove: "",
		wantKeep:   "",
	},
	{
		name:       "no funcs",
		str:        "abc",
		funcs:      nil,
		wantRemove: "abc",
		wantKeep:   "",
	},
	{
		name:       "digits from phone number",
		str:        "Tel: +43 (1) 234-56",
		funcs:      []IsRuneFunc{unicode.IsDigit},
		wantRemove: "Tel: + () -",
		wantKeep:   "43123456",
	},
	{
		name:       "any of multiple funcs matches",
		str:        "a-b_c d",
		funcs:      []IsRuneFunc{IsRune('-', '_'), unicode.IsSpace},
		wantRemove: "abcd",
		wantKeep:   "-_ ",
	},
	{
		name:       "multi-byte runes",
		str:        "Grüße aus 世界",
		funcs:      []IsRuneFunc{IsRune('ü', '世')},
		wantRemove: "Grße aus 界",
		wantKeep:   "ü世",
	},
	{
		name:       "nothing matches",
		str:        "Grüße",
		funcs:      []IsRuneFunc{unicode.IsDigit},
		wantRemove: "Grüße",
		wantKeep:   "",
	},
}

func TestRemoveRunes(t *testing.T) {
	for _, tt := range removeKeepRunesCases {
		t.Run(tt.name, func(t *testing.T) {
			in := []byte(tt.str)
			assert.Equal(t, tt.wantRemove, string(RemoveRunes(in, tt.funcs...)))
			assert.Equal(t, tt.str, string(in), "input must not be modified")
			// The string variant must produce the same result
			assert.Equal(t, tt.wantRemove, RemoveRunesString(tt.str, tt.funcs...))
		})
	}
}

func TestKeepRunes(t *testing.T) {
	for _, tt := range removeKeepRunesCases {
		t.Run(tt.name, func(t *testing.T) {
			in := []byte(tt.str)
			assert.Equal(t, tt.wantKeep, string(KeepRunes(in, tt.funcs...)))
			assert.Equal(t, tt.str, string(in), "input must not be modified")
			// The string variant must produce the same result
			assert.Equal(t, tt.wantKeep, KeepRunesString(tt.str, tt.funcs...))
		})
	}
}

// RemoveRunes and KeepRunes must process the whole input like their
// string variants: they used utf8.RuneError as end of input marker,
// so an invalid UTF-8 byte or a legitimate U+FFFD silently dropped
// everything after it. Like ranging over a string, an invalid byte
// is passed to the functions as utf8.RuneError and written as U+FFFD.
func TestRemoveKeepRunes_InvalidUTF8AndReplacementChar(t *testing.T) {
	isReplacementChar := IsRune(unicode.ReplacementChar)
	tests := []struct {
		name       string
		str        string
		funcs      []IsRuneFunc
		wantRemove string
		wantKeep   string
	}{
		{
			name:       "invalid byte in the middle",
			str:        "1a\xffb2",
			funcs:      []IsRuneFunc{unicode.IsDigit},
			wantRemove: "a\uFFFDb",
			wantKeep:   "12",
		},
		{
			name:       "invalid byte at the end",
			str:        "1a\xff",
			funcs:      []IsRuneFunc{unicode.IsDigit},
			wantRemove: "a\uFFFD",
			wantKeep:   "1",
		},
		{
			name:       "truncated multi-byte rune",
			str:        "1a\xe2\x82b2",
			funcs:      []IsRuneFunc{unicode.IsDigit},
			wantRemove: "a\uFFFD\uFFFDb",
			wantKeep:   "12",
		},
		{
			name:       "legitimate U+FFFD",
			str:        "1a\uFFFDb2\uFFFD",
			funcs:      []IsRuneFunc{unicode.IsDigit},
			wantRemove: "a\uFFFDb\uFFFD",
			wantKeep:   "12",
		},
		{
			name:       "invalid byte and U+FFFD both match utf8.RuneError",
			str:        "a\xffb\uFFFDc",
			funcs:      []IsRuneFunc{isReplacementChar},
			wantRemove: "abc",
			wantKeep:   "\uFFFD\uFFFD",
		},
		// RemoveRunesString copied the unchanged prefix before the first
		// removed rune verbatim, so an invalid byte in that prefix stayed
		// raw instead of becoming U+FFFD like in RemoveRunes, and whether
		// it was replaced depended on a later rune being removed.
		{
			name:       "invalid byte before the first removed rune",
			str:        "\xff1a",
			funcs:      []IsRuneFunc{unicode.IsDigit},
			wantRemove: "\uFFFDa",
			wantKeep:   "1",
		},
		{
			name:       "truncated multi-byte rune before the first removed rune",
			str:        "a\xe2\x82b1",
			funcs:      []IsRuneFunc{unicode.IsDigit},
			wantRemove: "a\uFFFD\uFFFDb",
			wantKeep:   "1",
		},
		{
			name:       "invalid byte and no rune removed",
			str:        "a\xffb",
			funcs:      []IsRuneFunc{unicode.IsDigit},
			wantRemove: "a\uFFFDb",
			wantKeep:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := []byte(tt.str)
			assert.Equal(t, tt.wantRemove, string(RemoveRunes(in, tt.funcs...)), "RemoveRunes")
			assert.Equal(t, tt.wantKeep, string(KeepRunes(in, tt.funcs...)), "KeepRunes")
			assert.Equal(t, tt.str, string(in), "input must not be modified")

			// The string variants must produce the same result
			assert.Equal(t, tt.wantRemove, RemoveRunesString(tt.str, tt.funcs...), "RemoveRunesString")
			assert.Equal(t, tt.wantKeep, KeepRunesString(tt.str, tt.funcs...), "KeepRunesString")
		})
	}
}

// SplitAndTrimIndex must split the whole input: it used utf8.RuneError
// as end of input marker, so an invalid UTF-8 byte or a legitimate
// U+FFFD dropped all following words and the last word was extended
// to the end of the input. Like ranging over a string, an invalid byte
// is a single utf8.RuneError rune that is part of its word.
func Test_SplitAndTrimIndex_InvalidUTF8AndReplacementChar(t *testing.T) {
	tests := []struct {
		str  string
		want [][]int
	}{
		{str: "ab\xffcd ef", want: [][]int{{0, 5}, {6, 8}}},
		{str: "\xff ab", want: [][]int{{0, 1}, {2, 4}}},
		{str: "ab \xff", want: [][]int{{0, 2}, {3, 4}}},
		{str: "ab. \xe2\x82. cd", want: [][]int{{0, 2}, {4, 6}, {8, 10}}},
		{str: "ab \uFFFD ef", want: [][]int{{0, 2}, {3, 6}, {7, 9}}},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got := SplitAndTrimIndex([]byte(tt.str), unicode.IsSpace, unicode.IsPunct)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemoveRunesString_NoCopyIfUnchanged(t *testing.T) {
	// Documented: if no rune was removed, the string is not copied
	str := strings.Repeat("Grüße ", 10)
	got := RemoveRunesString(str, unicode.IsDigit)
	require.Equal(t, str, got)
	assert.Same(t, unsafe.StringData(str), unsafe.StringData(got), "unchanged string must not be copied")
	// And not allocate either: it runs on the Normalized path of IBAN,
	// BIC and vat.ID, which Value, MarshalJSON and String call
	allocs := testing.AllocsPerRun(100, func() { _ = RemoveRunesString(str, unicode.IsDigit) })
	assert.Zero(t, allocs, "unchanged string must not allocate")

	// Removing the very first rune must not keep any of the original prefix
	assert.Equal(t, "bc", RemoveRunesString("abc", IsRune('a')))
	// Removing the very last rune
	assert.Equal(t, "ab", RemoveRunesString("abc", IsRune('c')))
}

func TestMapRuneIsAfterWordSeparator(t *testing.T) {
	// The result has one entry per byte (not per rune) so that it can be
	// indexed with byte offsets, for example from regexp matches.
	// The start of the string counts as being after a separator
	// so that the first word is treated like every other word.
	tests := []struct {
		str  string
		want []bool
	}{
		{str: "", want: []bool{}},
		{str: "a", want: []bool{true}},
		{str: "ab", want: []bool{true, false}},
		{str: "a b", want: []bool{true, false, true}},
		{str: "a.b", want: []bool{true, false, true}},
		{str: "a  b", want: []bool{true, false, true, true}},
		// Continuation bytes of a multi-byte rune get the value of the
		// rune's first byte
		{str: "ä b", want: []bool{true, true, false, true}},
		{str: "x äb", want: []bool{true, false, true, true, false}},
		{str: "世x", want: []bool{true, true, true, false}},
		// Also for the last rune of the string: its continuation bytes
		// were left false, so a byte offset into the last rune
		// (for example the end of a regexp match minus one)
		// got a different value than the rune's first byte
		{str: "ä", want: []bool{true, true}},
		{str: "a ä", want: []bool{true, false, true, true}},
		{str: "x世", want: []bool{true, false, false, false}},
		{str: "a 世", want: []bool{true, false, true, true, true}},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got := MapRuneIsAfterWordSeparator(tt.str)
			require.Len(t, got, len(tt.str))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEqualJSON(t *testing.T) {
	type pair struct {
		A int `json:"a"`
	}
	tests := []struct {
		name string
		a, b any
		want bool
	}{
		{name: "both nil", a: nil, b: nil, want: true},
		{name: "nil vs value", a: nil, b: 0, want: false},
		{name: "value vs nil", a: "", b: nil, want: false},
		{name: "equal strings", a: "a", b: "a", want: true},
		{name: "different strings", a: "a", b: "b", want: false},
		// Values of different Go types are equal if their JSON is equal
		{name: "int vs float", a: 1, b: 1.0, want: true},
		{name: "struct vs map", a: pair{A: 1}, b: map[string]int{"a": 1}, want: true},
		{name: "struct pointer vs struct", a: &pair{A: 1}, b: pair{A: 1}, want: true},
		// but the JSON type still matters
		{name: "number vs string", a: 1, b: "1", want: false},
		// Values that can't be marshalled are never equal to anything else
		{name: "unmarshallable", a: func() {}, b: 1, want: false},
		{name: "unmarshallable second", a: 1, b: make(chan int), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, EqualJSON(tt.a, tt.b))
		})
	}
}

// EqualJSON is meant for any values that can be marshalled as JSON,
// and slices and maps are the most common of those. Comparing two
// values of the same uncomparable type with == panics, so EqualJSON
// must compare them by their JSON instead of crashing the caller.
func TestEqualJSON_Uncomparable(t *testing.T) {
	type withSlice struct {
		S []int `json:"s"`
	}
	type withAny struct {
		A any `json:"a"`
	}
	tests := []struct {
		name string
		a, b any
		want bool
	}{
		{name: "equal slices", a: []int{1, 2}, b: []int{1, 2}, want: true},
		{name: "different slices", a: []int{1, 2}, b: []int{2, 1}, want: false},
		{name: "nil slices", a: []int(nil), b: []int(nil), want: true},
		{name: "equal maps", a: map[string]int{"a": 1, "b": 2}, b: map[string]int{"b": 2, "a": 1}, want: true},
		{name: "different maps", a: map[string]int{"a": 1}, b: map[string]int{"a": 2}, want: false},
		{name: "struct with slice", a: withSlice{S: []int{1}}, b: withSlice{S: []int{1}}, want: true},
		// The struct type is comparable, but its dynamic field value is not
		{name: "struct with slice in any field", a: withAny{A: []int{1}}, b: withAny{A: []int{1}}, want: true},
		{name: "slice vs array", a: []int{1}, b: [1]int{1}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				assert.Equal(t, tt.want, EqualJSON(tt.a, tt.b))
			})
		})
	}
}

func TestTransliterateSpecialCharacters(t *testing.T) {
	// Without a length limit
	assert.Equal(t, "Gruesse aus Oesterreich", TransliterateSpecialCharacters("Grüße aus Österreich"))
	assert.Equal(t, "", TransliterateSpecialCharacters(""))

	// Unchanged input is returned without copying
	str := strings.Repeat("plain ASCII ", 10)
	got := TransliterateSpecialCharacters(str)
	require.Equal(t, str, got)
	assert.Same(t, unsafe.StringData(str), unsafe.StringData(got), "unchanged string must not be copied")
}

func TestSubStringInAndStringContainsAny(t *testing.T) {
	// The two functions are mirror images: SubStringIn looks for a needle
	// in many haystacks, StringContainsAny for many needles in a haystack.
	strs := []string{"Hello World", "Grüße"}
	assert.True(t, SubStringIn("World", strs))
	assert.True(t, SubStringIn("üß", strs))
	assert.True(t, SubStringIn("Grüße", strs), "equal string counts as substring")
	assert.False(t, SubStringIn("world", strs), "case sensitive")
	assert.False(t, SubStringIn("Hello World and more", strs), "haystack must contain needle, not vice versa")
	assert.False(t, SubStringIn("", nil))

	assert.True(t, StringContainsAny("Hello World", []string{"x", "World"}))
	assert.True(t, StringContainsAny("Grüße", []string{"üß"}))
	assert.False(t, StringContainsAny("Hello", []string{"Hello World"}), "needle must be in haystack, not vice versa")
	assert.False(t, StringContainsAny("Hello", nil))
	// Like strings.Contains, the empty string is contained in every string
	assert.True(t, StringContainsAny("Hello", []string{""}))
	assert.True(t, SubStringIn("", strs))
}
