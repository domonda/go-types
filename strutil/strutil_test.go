package strutil

import (
	"strings"
	"testing"
	"unicode"
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
	"x.y":                            [][]int{[]int{0, 3}},
	"x.y .x":                         [][]int{[]int{0, 3}, []int{5, 6}},
	"HelloWorld":                     [][]int{[]int{0, 10}},
	"HelloWorld ":                    [][]int{[]int{0, 10}},
	"HelloWorld  ":                   [][]int{[]int{0, 10}},
	"HelloWorld. ":                   [][]int{[]int{0, 10}},
	"HelloWorld.. ":                  [][]int{[]int{0, 10}},
	"HelloWorld.  .":                 [][]int{[]int{0, 10}},
	" HelloWorld":                    [][]int{[]int{1, 11}},
	" .HelloWorld":                   [][]int{[]int{2, 12}},
	". .HelloWorld":                  [][]int{[]int{3, 13}},
	"...HelloWorld":                  [][]int{[]int{3, 13}},
	"Hello World":                    [][]int{[]int{0, 5}, []int{6, 11}},
	"Hello.World":                    [][]int{[]int{0, 11}},
	"Hello  World":                   [][]int{[]int{0, 5}, []int{7, 12}},
	"Hello.  World":                  [][]int{[]int{0, 5}, []int{8, 13}},
	"Hello.. World":                  [][]int{[]int{0, 5}, []int{8, 13}},
	"Hello...World":                  [][]int{[]int{0, 13}},
	"Hello. .World":                  [][]int{[]int{0, 5}, []int{8, 13}},
	" Hello. .World":                 [][]int{[]int{1, 6}, []int{9, 14}},
	" Hello. .World ":                [][]int{[]int{1, 6}, []int{9, 14}},
	" Hello. .World  ":               [][]int{[]int{1, 6}, []int{9, 14}},
	" Hello. .World. ":               [][]int{[]int{1, 6}, []int{9, 14}},
	" Hello. .World.. ":              [][]int{[]int{1, 6}, []int{9, 14}},
	"one two three four 5":           [][]int{[]int{0, 3}, []int{4, 7}, []int{8, 13}, []int{14, 18}, []int{19, 20}},
	"one two three four 5.":          [][]int{[]int{0, 3}, []int{4, 7}, []int{8, 13}, []int{14, 18}, []int{19, 20}},
	"one two three four 5  ":         [][]int{[]int{0, 3}, []int{4, 7}, []int{8, 13}, []int{14, 18}, []int{19, 20}},
	".one. .two. .three. .four. .5":  [][]int{[]int{1, 4}, []int{7, 10}, []int{13, 18}, []int{21, 25}, []int{28, 29}},
	".one. .two. .three. .four. .5.": [][]int{[]int{1, 4}, []int{7, 10}, []int{13, 18}, []int{21, 25}, []int{28, 29}},
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
