package strutil

import (
	"math"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
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
		"": "",

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
		"もしもしWorld":            "もしもし_world",
	}
	for str, expected := range testCases {
		t.Run(str, func(t *testing.T) {
			actual := ToSnakeCase(str)
			assert.Equal(t, expected, actual, "snake case")
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
		{"", math.MaxInt, ""},
		{"Österreich", math.MaxInt, "Oesterreich"},
		{"Österreich", 3, "Oes"},
		{"Öster\uFFFDeich", 100, "Oestereich"},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			if got := ReplaceTransliterationsMaxLen(tt.str, tt.maxLen); got != tt.want {
				t.Errorf("ReplaceTransliterationsMaxLen(%#v) = %#v, want %#v", tt.str, got, tt.want)
			}
		})
	}
}
