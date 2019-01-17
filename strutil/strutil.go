package strutil

import (
	"bytes"
	"encoding/json"
	"net/url"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"
)

// panicIfNot panics if any assumption is false
func panicIfNot(assumptions ...bool) {
	for _, assumption := range assumptions {
		if !assumption {
			panic("assert false")
		}
	}
}

// func toUpperCaseLettersAndDigits(str string) string {
// 	var b strings.Builder
// 	for _, r := range str {
// 		if unicode.IsDigit(r) || unicode.IsLetter(r) {
// 			b.WriteRune(unicode.ToUpper(r))
// 		}
// 	}
// 	return b.String()
// }

// IsRuneFunc is function pionter for specifying if a rune matches a criteria
type IsRuneFunc func(rune) bool

// SplitAndTrimIndex first splits str into words not containing isSplitRune,
// then trims the words with isTrimRune.
func SplitAndTrimIndex(str []byte, isSplitRune, isTrimRune IsRuneFunc) (indices [][]int) {
	inWord := false
	inTrimmedWord := false
	lastWasTrim := false
	onlyTrimsSince := -1
	wordStart := -1
	r, n := utf8.DecodeRune(str)

	for i := 0; r != utf8.RuneError; {
		isSplit := isSplitRune(r)
		isTrim := isTrimRune(r)

		if isTrim && !lastWasTrim {
			onlyTrimsSince = i
		}

		if inWord {
			if inTrimmedWord {
				if isSplit {
					wordEnd := i
					if onlyTrimsSince != -1 {
						// if there have been only trim rune since the last word rune
						// backtrack wordEnd to first trim character
						wordEnd = onlyTrimsSince
					}
					if wordEnd > wordStart {
						// panicIfNot(wordStart >= 0)
						indices = append(indices, []int{wordStart, wordEnd})
					}
					inTrimmedWord = false
					inWord = false
					wordStart = -2
				}
			} else {
				if isSplit {
					inTrimmedWord = false
					inWord = false
					wordStart = -3
				} else if !isTrim {
					inWord = true
					inTrimmedWord = true
					wordStart = i
				}
			}
		} else {
			// panicIfNot(inTrimmedWord == false)
			if !isSplit {
				inWord = true
				if !isTrim {
					inTrimmedWord = true
					wordStart = i
				}
			}

		}

		lastWasTrim = isTrim
		if !isTrim {
			onlyTrimsSince = -1
		}

		i += n
		r, n = utf8.DecodeRune(str[i:])
	}
	if inTrimmedWord {
		wordEnd := len(str)
		if onlyTrimsSince != -1 {
			wordEnd = onlyTrimsSince
		}
		if wordEnd > wordStart {
			panicIfNot(wordStart >= 0)
			indices = append(indices, []int{wordStart, wordEnd})
		}
	}
	return indices
}

func IsRune(runes ...rune) IsRuneFunc {
	return func(testR rune) bool {
		for _, r := range runes {
			if testR == r {
				return true
			}
		}
		return false
	}
}

func IsWordSeparator(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsMark(r) || unicode.IsSymbol(r)
}

func IsNorLetterOrDigit(r rune) bool {
	return !(unicode.IsDigit(r) || unicode.IsLetter(r))
}

func IsRuneInverse(isRune IsRuneFunc) IsRuneFunc {
	return func(r rune) bool {
		return !isRune(r)
	}
}

func IsRuneAll(isRune ...IsRuneFunc) IsRuneFunc {
	return func(r rune) bool {
		for _, is := range isRune {
			if !is(r) {
				return false
			}
		}
		return true
	}
}

func IsRuneAny(isRune ...IsRuneFunc) IsRuneFunc {
	return func(r rune) bool {
		for _, is := range isRune {
			if is(r) {
				return true
			}
		}
		return false
	}
}

// RemoveRunes returns a copy of str with all runes removed
// where any call to a removeRunes function reeturns true.
func RemoveRunes(str []byte, removeRunes ...IsRuneFunc) []byte {
	var buf bytes.Buffer
	r, n := utf8.DecodeRune(str)
	for r != utf8.RuneError {
		doRemove := false
		for _, remove := range removeRunes {
			if remove(r) {
				doRemove = true
				break
			}
		}
		if !doRemove {
			buf.WriteRune(r)
		}
		str = str[n:]
		r, n = utf8.DecodeRune(str)
	}
	return buf.Bytes()
}

// KeepRunes returns a copy of str where with all runes kept in it
// where any call to a keepRunes function reeturns true.
func KeepRunes(str []byte, keepRunes ...IsRuneFunc) []byte {
	var buf bytes.Buffer
	r, n := utf8.DecodeRune(str)
	for r != utf8.RuneError {
		for _, keep := range keepRunes {
			if keep(r) {
				buf.WriteRune(r)
				break
			}
		}
		str = str[n:]
		r, n = utf8.DecodeRune(str)
	}
	return buf.Bytes()
}

// RemoveRunesString returns a copy of str with all runes removed
// where any call to a removeRunes function reeturns true.
// If no rune was removed, the string is not copied.
func RemoveRunesString(str string, removeRunes ...IsRuneFunc) string {
	var b strings.Builder
	changed := false
	for i, r := range str {
		doRemove := false
		for _, remove := range removeRunes {
			if remove(r) {
				doRemove = true
				break
			}
		}
		if changed {
			if !doRemove {
				b.WriteRune(r)
			}
		} else {
			if doRemove {
				changed = true
				b.WriteString(str[:i])
			}
		}
	}
	if changed {
		return b.String()
	}
	return str
}

// func RemoveRunesString(str string, removeRunes ...IsRuneFunc) string {
// 	var buf bytes.Buffer
// 	for _, r := range str {
// 		doRemove := false
// 		for _, remove := range removeRunes {
// 			if remove(r) {
// 				doRemove = true
// 				break
// 			}
// 		}
// 		if !doRemove {
// 			buf.WriteRune(r)
// 		}
// 	}
// 	return buf.String()
// }

// KeepRunesString returns a copy of str where with all runes kept in it
// where any call to a keepRunes function reeturns true.
func KeepRunesString(str string, keepRunes ...IsRuneFunc) string {
	var b strings.Builder
	for _, r := range str {
		for _, keep := range keepRunes {
			if keep(r) {
				b.WriteRune(r)
				break
			}
		}
	}
	return b.String()
}

func MapRuneIsAfterWordSeparator(str string) []bool {
	result := make([]bool, len(str))
	i0 := -1
	r0 := rune(' ')
	for i, r := range str {
		lastWasSep := IsWordSeparator(r0)
		result[i] = lastWasSep
		// If rune has more than one byte,
		// then copy mapping to further bytes
		for b := i0 + 1; b < i; b++ {
			result[b] = result[b-1]
		}
		i0 = i
		r0 = r
	}
	return result
}

// EqualJSON returns true a and b are equal on a value basis,
// or if their marshalled JSON representation is equal.
func EqualJSON(a, b interface{}) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(aJSON, bJSON)

}

func writeSafeFileNameRune(buf *bytes.Buffer, i int, r rune, lastWasPlaceholder bool) (wrotePlaceholder bool) {
	if unicode.IsSpace(r) {
		if i > 0 && !lastWasPlaceholder {
			buf.WriteByte('_')
		}
		return true
	}

	trans, ok := transliterations[r]
	if ok {
		buf.WriteString(trans)
		return false
	}

	if r == '-' || r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
		buf.WriteRune(r)
		return false
	} else {
		if i > 0 && !lastWasPlaceholder {
			buf.WriteByte('-')
		}
		return true
	}
}

func SanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if strings.IndexAny(name, "%+") != -1 {
		n, err := url.QueryUnescape(name)
		if err == nil && n != "" {
			name = n
		}
	}
	ext := path.Ext(name)
	if ext != "" {
		name = name[:len(name)-len(ext)]
		ext = NormalizeExt(ext)
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(name)))
	lastWasPlaceholder := false
	for i, r := range name {
		lastWasPlaceholder = writeSafeFileNameRune(buf, i, r, lastWasPlaceholder)
	}

	return buf.String() + ext
}

func NormalizeExt(ext string) string {
	ext = strings.ToLower(ext)
	if norm, ok := normalizedExt[ext]; ok {
		return norm
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(ext)))
	lastWasPlaceholder := false
	for i, r := range ext {
		if i == 0 && r == '.' {
			buf.WriteRune(r)
		} else {
			lastWasPlaceholder = writeSafeFileNameRune(buf, i, r, lastWasPlaceholder)
		}
	}

	return buf.String()
}

var normalizedExt = map[string]string{
	".":    "",
	".tif": ".tiff",
	".jpg": ".jpeg",
}

// A very limited list of transliterations to catch common european names translated to urls.
// This set could be expanded with at least caps and many more characters.
var transliterations = map[rune]string{
	'À': "A",
	'Á': "A",
	'Â': "A",
	'Ã': "A",
	'Ä': "Ae",
	'Å': "AA",
	'Æ': "AE",
	'Ç': "C",
	'È': "E",
	'É': "E",
	'Ê': "E",
	'Ë': "E",
	'Ì': "I",
	'Í': "I",
	'Î': "I",
	'Ï': "I",
	'Ð': "D",
	'Ł': "L",
	'Ñ': "N",
	'Ò': "O",
	'Ó': "O",
	'Ô': "O",
	'Õ': "O",
	'Ö': "Oe",
	'Ø': "OE",
	'Ù': "U",
	'Ú': "U",
	'Ü': "Ue",
	'Û': "U",
	'Ý': "Y",
	'Þ': "Th",
	'ß': "ss",
	'à': "a",
	'á': "a",
	'â': "a",
	'ã': "a",
	'ä': "ae",
	'å': "aa",
	'æ': "ae",
	'ç': "c",
	'è': "e",
	'é': "e",
	'ê': "e",
	'ë': "e",
	'ì': "i",
	'í': "i",
	'î': "i",
	'ï': "i",
	'ð': "d",
	'ł': "l",
	'ñ': "n",
	'ń': "n",
	'ò': "o",
	'ó': "o",
	'ô': "o",
	'õ': "o",
	'ō': "o",
	'ö': "oe",
	'ø': "oe",
	'ś': "s",
	'ù': "u",
	'ú': "u",
	'û': "u",
	'ū': "u",
	'ü': "ue",
	'ý': "y",
	'þ': "th",
	'ÿ': "y",
	'ż': "z",
	'Œ': "OE",
	'œ': "oe",
}

// SanitizeLineEndings converts all line endings to just '\n'
func SanitizeLineEndings(text string) string {
	// var (
	// 	needsCopy = false
	// 	buf bytes.Buffer
	// 	lastByte byte
	// )

	// for i, b := range text {

	// TODO optimized version

	// 	lastByte = b
	// }

	// if needsCopy {
	// 	return buf.Bytes()
	// }
	// return text

	text = strings.Replace(text, "\r\n", "\n", -1)
	text = strings.Replace(text, "\n\r", "\n", -1)
	text = strings.Replace(text, "\r", "\n", -1)

	return text
}

// SanitizeLineEndingsBytes converts all line endings to just '\n'
func SanitizeLineEndingsBytes(text []byte) []byte {
	// var (
	// 	needsCopy = false
	// 	buf bytes.Buffer
	// 	lastByte byte
	// )

	// for i, b := range text {

	// TODO optimized version

	// 	lastByte = b
	// }

	// if needsCopy {
	// 	return buf.Bytes()
	// }
	// return text

	text = bytes.Replace(text, []byte{'\r', '\n'}, []byte{'\n'}, -1)
	text = bytes.Replace(text, []byte{'\n', '\r'}, []byte{'\n'}, -1)
	text = bytes.Replace(text, []byte{'\r'}, []byte{'\n'}, -1)

	return text
}
