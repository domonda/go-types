package strfmt

import (
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

// FormatFloat formats a float similar to strconv.FormatFloat with the 'f' format option,
// but with decimalSep as decemal separator instead of a point
// and optional grouping of the integer part.
// Valid values for decimalSep are '.' and ','.
// If groupSep is not zero, then the integer part of the number is grouped
// with groupSep between every group of 3 digits.
// Valid values for groupSep are [0, ',', '.'] and groupSep must be different from  decimalSep.
// precision controls the number of digits (excluding the exponent).
// The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return f exactly.
func FormatFloat(f float64, groupSep, decimalSep byte, precision int) string {
	if groupSep != 0 && groupSep != '.' && groupSep != ',' && groupSep != ' ' {
		panic("invalid groupSep")
	}
	if decimalSep != '.' && decimalSep != ',' {
		panic("invalid decimalSep")
	}
	if groupSep == decimalSep {
		panic("groupSep == decimalSep")
	}

	str := strconv.FormatFloat(f, 'f', precision, 64)
	if groupSep != 0 {
		pointPos := strings.IndexByte(str, '.')
		if pointPos == -1 {
			pointPos = len(str)
		}
		prefixLen := 0
		if f < 0 {
			prefixLen = 1
		}
		integerLen := pointPos - prefixLen
		firstGroupLen := prefixLen
		if integerLen%3 == 0 {
			firstGroupLen += 3
		} else {
			firstGroupLen += integerLen % 3
		}
		numGroupSeps := (integerLen - 1) / 3

		b := strings.Builder{}
		b.Grow(len(str) + numGroupSeps)

		b.WriteString(str[:firstGroupLen])
		for i := 0; i < numGroupSeps; i++ {
			b.WriteByte(groupSep)
			start := firstGroupLen + i*3
			b.WriteString(str[start : start+3])
		}

		if pointPos != len(str) {
			b.WriteByte(decimalSep)
			b.WriteString(str[pointPos+1:])
		}

		return b.String()
	}

	if decimalSep != '.' {
		return strings.Replace(str, ".", string(decimalSep), 1)
	}

	return str
}

// ParseFloat parses float string compatible with FormatFloat.
func ParseFloat(str string) (float64, error) {
	f, _, _, err := ParseFloatInfo(str)
	return f, err
}

// ParseFloatInfo parses float string compatible with FormatFloat
// and returns the detected integer group separator and decimal separator characters.
func ParseFloatInfo(str string) (f float64, groupSep, decimalSep byte, err error) {
	var (
		lastDigitIndex    = -1
		lastNonDigitIndex = -1

		pointWritten = false
		eWritten     = false

		hasGrouping       = false
		lastGroupingRune  rune
		lastGroupingIndex int
	)

	var floatBuilder strings.Builder
	floatBuilder.Grow(len(str))

	for i, r := range str {
		switch {
		case r == ' ':
			if pointWritten {
				// No further separator runes allowed after point
				return 0, 0, 0, errors.Errorf("Can't parse %#v as float", str)
			}

			// Write everything after the lastNonDigitIndex and before current index
			floatBuilder.WriteString(str[lastNonDigitIndex+1 : i])
			lastNonDigitIndex = i

			if !hasGrouping {
				// This is the first decimal rune, just save it
				hasGrouping = true
				lastGroupingRune = r
				lastGroupingIndex = i
			} else {
				// It's a further decimal rune, has to be 3 bytes since last decimal rune
				if i-(lastGroupingIndex+1) != 3 {
					return 0, 0, 0, errors.Errorf("Can't parse %#v as float", str)
				}
				if r == lastGroupingRune {
					// If it's the same decimal rune, then just save it
					lastGroupingRune = r
					lastGroupingIndex = i
				} else {
					// Spaces only are used as separators,
					// the the last separator was not a space, something is wrong
					return 0, 0, 0, errors.Errorf("Can't parse %#v as float", str)
				}
			}

		case r == '.' || r == ',':
			if pointWritten {
				// No further separator runes allowed after point
				return 0, 0, 0, errors.Errorf("Can't parse %#v as float", str)
			}

			// Write everything after the lastNonDigitIndex and before current index
			floatBuilder.WriteString(str[lastNonDigitIndex+1 : i])
			lastNonDigitIndex = i

			if !hasGrouping {
				// This is the first decimal rune, just save it
				hasGrouping = true
				lastGroupingRune = r
				lastGroupingIndex = i
			} else {
				// It's a further decimal rune, has to be 3 bytes since last decimal rune
				if i-(lastGroupingIndex+1) != 3 {
					return 0, 0, 0, errors.Errorf("Can't parse %#v as float", str)
				}
				if r == lastGroupingRune {
					// If it's the same decimal rune, then just save it
					lastGroupingRune = r
					lastGroupingIndex = i
				} else {
					// If it's a different decimal rune, then we have
					// reached the decimal separator
					floatBuilder.WriteByte('.')
					pointWritten = true
					groupSep = byte(lastGroupingRune)
					decimalSep = byte(r)
				}
			}

		case r >= '0' && r <= '9':
			// floatBuilder.WriteByte(byte(r))
			lastDigitIndex = i

			// case unicode.IsSpace(r):
			// 	continue

		case r == '-' || r == '+':
			if i > 0 {
				return 0, 0, 0, errors.Errorf("Can't parse %#v as float", str)
			}
			floatBuilder.WriteByte(byte(r))
			lastNonDigitIndex = i

		case r == 'e':
			if i == 0 || eWritten {
				// e can't be the first rune or be repeated
				return 0, 0, 0, errors.Errorf("Can't parse %#v as float", str)
			}
			if hasGrouping && !pointWritten {
				floatBuilder.WriteByte('.')
				pointWritten = true
				decimalSep = '.'
			}
			floatBuilder.WriteString(str[lastNonDigitIndex+1 : i])
			lastNonDigitIndex = i

			floatBuilder.WriteByte('e')
			eWritten = true

		default:
			return 0, 0, 0, errors.Errorf("Can't parse %#v as float", str)
		}
	}

	if hasGrouping && !pointWritten {
		floatBuilder.WriteByte('.')
		pointWritten = true
		decimalSep = byte(lastGroupingRune)
	}
	if lastDigitIndex >= lastNonDigitIndex {
		floatBuilder.WriteString(str[lastNonDigitIndex+1 : lastDigitIndex+1])
	}

	// floatStr := floatBuilder.String()
	f, err = strconv.ParseFloat(floatBuilder.String(), 64)
	if err != nil {
		return 0, 0, 0, err
	}
	return f, groupSep, decimalSep, nil
}
