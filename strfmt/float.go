package strfmt

import (
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

func ParseFloat(str string) (float64, error) {
	var (
		lastDigitIndex    = -1
		lastNonDigitIndex = -1

		hasSeparator = false

		pointWritten = false
		eWritten     = false

		lastSeparatorRune  rune
		lastSeparatorIndex int
	)

	var floatBuilder strings.Builder
	floatBuilder.Grow(len(str))

	for i, r := range str {
		switch {
		case r == ' ':
			if pointWritten {
				// No further separator runes allowed after point
				return 0, errors.Errorf("Can't parse %#v as float", str)
			}

			// Write everything after the lastNonDigitIndex and before current index
			floatBuilder.WriteString(str[lastNonDigitIndex+1 : i])
			lastNonDigitIndex = i

			if !hasSeparator {
				// This is the first decimal rune, just save it
				hasSeparator = true
				lastSeparatorRune = r
				lastSeparatorIndex = i
			} else {
				// It's a further decimal rune, has to be 3 bytes since last decimal rune
				if i-(lastSeparatorIndex+1) != 3 {
					return 0, errors.Errorf("Can't parse %#v as float", str)
				}
				if r == lastSeparatorRune {
					// If it's the same decimal rune, then just save it
					lastSeparatorRune = r
					lastSeparatorIndex = i
				} else {
					// Spaces only are used as separators,
					// the the last separator was not a space, something is wrong
					return 0, errors.Errorf("Can't parse %#v as float", str)
				}
			}

		case r == '.' || r == ',':
			if pointWritten {
				// No further separator runes allowed after point
				return 0, errors.Errorf("Can't parse %#v as float", str)
			}

			// Write everything after the lastNonDigitIndex and before current index
			floatBuilder.WriteString(str[lastNonDigitIndex+1 : i])
			lastNonDigitIndex = i

			if !hasSeparator {
				// This is the first decimal rune, just save it
				hasSeparator = true
				lastSeparatorRune = r
				lastSeparatorIndex = i
			} else {
				// It's a further decimal rune, has to be 3 bytes since last decimal rune
				if i-(lastSeparatorIndex+1) != 3 {
					return 0, errors.Errorf("Can't parse %#v as float", str)
				}
				if r == lastSeparatorRune {
					// If it's the same decimal rune, then just save it
					lastSeparatorRune = r
					lastSeparatorIndex = i
				} else {
					// If it's a different decimal rune, then we have
					// reached the decimal separator
					floatBuilder.WriteByte('.')
					pointWritten = true
				}
			}

		case r >= '0' && r <= '9':
			// floatBuilder.WriteByte(byte(r))
			lastDigitIndex = i

			// case unicode.IsSpace(r):
			// 	continue

		case r == '-' || r == '+':
			if i > 0 {
				return 0, errors.Errorf("Can't parse %#v as float", str)
			}
			floatBuilder.WriteByte(byte(r))
			lastNonDigitIndex = i

		case r == 'e':
			if i == 0 || eWritten {
				// e can't be the first rune or be repeated
				return 0, errors.Errorf("Can't parse %#v as float", str)
			}
			if hasSeparator && !pointWritten {
				floatBuilder.WriteByte('.')
				pointWritten = true
			}
			floatBuilder.WriteString(str[lastNonDigitIndex+1 : i])
			lastNonDigitIndex = i

			floatBuilder.WriteByte('e')
			eWritten = true

		default:
			return 0, errors.Errorf("Can't parse %#v as float", str)
		}
	}

	if hasSeparator && !pointWritten {
		floatBuilder.WriteByte('.')
		pointWritten = true
	}
	if lastDigitIndex >= lastNonDigitIndex {
		floatBuilder.WriteString(str[lastNonDigitIndex+1 : lastDigitIndex+1])
	}

	floatStr := floatBuilder.String()
	return strconv.ParseFloat(floatStr, 64)
}
