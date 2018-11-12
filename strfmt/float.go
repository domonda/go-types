package strfmt

import (
	"math"
	"strconv"
	"strings"

	"github.com/domonda/errors"
)

// FormatFloat formats a float similar to strconv.FormatFloat with the 'f' format option,
// but with decimalSep as decimal separator instead of a point
// and optional thousands grouping of the integer part.
// Valid values for decimalSep are '.' and ','.
// If thousandsSep is not zero, then the integer part of the number is grouped
// with thousandsSep between every group of 3 digits from right to left.
// Valid values for thousandsSep are [0, ',', '.', '\''] and thousandsSep must be different from  decimalSep.
// precision controls the number of digits (excluding the exponent).
// The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return f exactly.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func FormatFloat(f float64, thousandsSep, decimalSep byte, precision int) string {
	if thousandsSep != 0 && thousandsSep != '.' && thousandsSep != ',' && thousandsSep != ' ' && thousandsSep != '\'' {
		panic(errors.Errorf("invalid thousandsSep: %#v", string(thousandsSep)))
	}
	if decimalSep != '.' && decimalSep != ',' {
		panic(errors.Errorf("invalid decimalSep: %#v", string(decimalSep)))
	}
	if thousandsSep == decimalSep {
		panic(errors.Errorf("thousandsSep == decimalSep: %#v", string(thousandsSep)))
	}

	str := strconv.FormatFloat(f, 'f', precision, 64)
	if thousandsSep != 0 && math.Abs(f) >= 1000 {
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
			b.WriteByte(thousandsSep)
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
		for i, c := range str {
			if c == '.' {
				return str[:i] + string(decimalSep) + str[i+1:]
			}
		}
	}

	return str
}

// ParseFloat parses float string compatible with FormatFloat.
// If a separator was not detected, then zero will be returned for thousandsSep or decimalSep.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func ParseFloat(str string) (float64, error) {
	f, _, _, err := ParseFloatInfo(str)
	return f, err
}

// ParseFloatInfo parses float string compatible with FormatFloat
// and returns the detected integer thousands separator and decimal separator characters.
// If a separator was not detected, then zero will be returned for thousandsSep or decimalSep.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func ParseFloatInfo(str string) (f float64, thousandsSep, decimalSep byte, err error) {
	var (
		lastDigitIndex    = -1
		lastNonDigitIndex = -1

		pointWritten = false
		eWritten     = false

		hasGrouping       = false
		lastGroupingRune  rune
		lastGroupingIndex int
	)

	floatBuilder := strings.Builder{}
	floatBuilder.Grow(len(str))

	for i, r := range str {
		switch {
		case r >= '0' && r <= '9':
			lastDigitIndex = i

		case r == '.' || r == ',' || r == '\'':
			if pointWritten {
				return 0, 0, 0, errors.Errorf("No further separators allowed after decimal separator: %#v", str)
			}

			// Write everything after the lastNonDigitIndex and before current index
			floatBuilder.WriteString(str[lastNonDigitIndex+1 : i])
			lastNonDigitIndex = i

			if !hasGrouping {
				// This is the first separator rune, just save it
				hasGrouping = true
				lastGroupingRune = r
				lastGroupingIndex = i
			} else {
				// It's a further separator rune, has to be 3 bytes since last separator rune
				if i-(lastGroupingIndex+1) != 3 {
					return 0, 0, 0, errors.Errorf("Separators have to be 3 characters apart: %#v", str)
				}
				if r == lastGroupingRune {
					// If it's the same separator rune, then just save it
					lastGroupingRune = r
					lastGroupingIndex = i
				} else {
					// If it's a different separator rune, then we have
					// reached the decimal separator
					floatBuilder.WriteByte('.')
					pointWritten = true
					thousandsSep = byte(lastGroupingRune)
					decimalSep = byte(r)
				}
			}

		case r == ' ':
			if pointWritten {
				return 0, 0, 0, errors.Errorf("No further separators allowed after decimal separator: %#v", str)
			}

			// Write everything after the lastNonDigitIndex and before current index
			floatBuilder.WriteString(str[lastNonDigitIndex+1 : i])
			lastNonDigitIndex = i

			if !hasGrouping {
				// This is the first separator rune, just save it
				hasGrouping = true
				lastGroupingRune = r
				lastGroupingIndex = i
			} else {
				// It's a further separator rune, has to be 3 bytes since last separator rune
				if i-(lastGroupingIndex+1) != 3 {
					return 0, 0, 0, errors.Errorf("Separators have to be 3 characters apart: %#v", str)
				}
				if r == lastGroupingRune {
					// If it's the same separator rune, then just save it
					lastGroupingRune = r
					lastGroupingIndex = i
				} else {
					// Spaces only are used as thousands separators.
					// If the the last separator was not a space, something is wrong
					return 0, 0, 0, errors.Errorf("Space can not be used after another thousands separator: %#v", str)
				}
			}

		case r == '-' || r == '+':
			if i > 0 {
				return 0, 0, 0, errors.Errorf("Sign can only be used as first character: %#v", str)
			}
			floatBuilder.WriteByte(byte(r))
			lastNonDigitIndex = i

		case r == 'e':
			if i == 0 || eWritten {
				return 0, 0, 0, errors.Errorf("e can't be the first or a repeating character: %#v", str)
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
			return 0, 0, 0, errors.Errorf("Invalid rune '%s' in %#v", string(r), str)
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

	f, err = strconv.ParseFloat(floatBuilder.String(), 64)
	if err != nil {
		return 0, 0, 0, err
	}
	return f, thousandsSep, decimalSep, nil
}
