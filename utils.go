package types

import (
	"strings"
	"unicode"
)

func toUpperCaseLettersAndDigits(str string) string {
	var b strings.Builder
	for _, r := range str {
		if unicode.IsDigit(r) || unicode.IsLetter(r) {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}
