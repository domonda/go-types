package strutil

import (
	"bytes"
	"strings"
	"testing"
	"unicode"

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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		"no_cr":   strings.Repeat("Hello, World!\n", 100),
		"mixed":   strings.Repeat("Hello\r\nWorld\rFoo\n\r", 100),
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
