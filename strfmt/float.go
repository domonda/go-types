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
// If padPrecision is true and precision is greater zero,
// then the end of the fractional part will be padded with
// '0' characters to reach the length of precision.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func FormatFloat(f float64, thousandsSep, decimalSep byte, precision int, padPrecision bool) string {
	if thousandsSep != 0 && thousandsSep != '.' && thousandsSep != ',' && thousandsSep != ' ' && thousandsSep != '\'' {
		panic(errors.Errorf("invalid thousandsSep: '%s'", string(thousandsSep)))
	}
	if decimalSep != '.' && decimalSep != ',' {
		panic(errors.Errorf("invalid decimalSep: '%s'", string(decimalSep)))
	}
	if thousandsSep == decimalSep {
		panic(errors.Errorf("thousandsSep == decimalSep: '%s'", string(thousandsSep)))
	}
	if precision < -1 {
		panic(errors.Errorf("precision < -1: %#v", precision))
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
			fraction := str[pointPos+1:]
			b.WriteString(fraction)
			if padPrecision {
				for i := len(fraction); i < precision; i++ {
					b.WriteByte('0')
				}
			}
		} else if padPrecision && precision > 0 {
			b.WriteByte(decimalSep)
			for i := 0; i < precision; i++ {
				b.WriteByte('0')
			}
		}

		return b.String()
	}

	if decimalSep != '.' {
		for i, c := range str {
			if c == '.' {
				// TODO optimize
				fraction := str[i+1:]
				if padPrecision {
					for i := len(fraction); i < precision; i++ {
						fraction += "0"
					}
				}
				return str[:i] + string(decimalSep) + fraction
			}
		}
	}

	if padPrecision && precision > 0 {
		pointPos := strings.IndexByte(str, '.')
		if pointPos == -1 {
			var b strings.Builder
			b.Grow(len(str) + 1 + precision)
			b.WriteString(str)
			b.WriteByte('.')
			for i := 0; i < precision; i++ {
				b.WriteByte('0')
			}
			return b.String()
		}

		numMissingZeros := precision - (len(str) - (pointPos + 1))
		if numMissingZeros > 0 {
			var b strings.Builder
			b.Grow(len(str) + numMissingZeros)
			b.WriteString(str)
			for i := 0; i < numMissingZeros; i++ {
				b.WriteByte('0')
			}
			return b.String()
		}
	}

	return str
}

// ParseFloat parses float string compatible with FormatFloat.
// If a separator was not detected, then zero will be returned for thousandsSep or decimalSep.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func ParseFloat(str string) (float64, error) {
	f, _, _, _, err := ParseFloatDetails(str)
	return f, err
}

// ParseFloatDetails parses float string compatible with FormatFloat
// and returns the detected integer thousands separator and decimal separator characters.
// If a separator was not detected, then zero will be returned for thousandsSep or decimalSep.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func ParseFloatDetails(str string) (f float64, thousandsSep, decimalSep byte, decimals int, err error) {

	var (
		lastDigitIndex    = -1
		lastNonDigitIndex = -1

		pointWritten = false
		eWritten     = false

		numMinus          int
		numGroupingRunes  int
		lastGroupingRune  rune
		lastGroupingIndex int

		skipFirst int // skip first bytes of str
		skipLast  int // skip last bytes of str

		floatBuilder strings.Builder
	)

	str = strings.TrimSpace(str)

	floatBuilder.Grow(len(str))

	// detect the sign, allowed positions are start and end
	for i, r := range str {
		switch {
		case r == '-':
			switch {
			case i == 0:
				skipFirst = 1
			case i == len(str)-1:
				skipLast = 1
			default:
				return 0, 0, 0, 0, errors.Errorf("minus can only be used as first or last character: %q", str)
			}
			floatBuilder.WriteByte(byte(r))
			numMinus = 1

		case r == '+':
			switch {
			case i == 0:
				skipFirst = 1
			case i == len(str)-1:
				skipLast = 1
			default:
				return 0, 0, 0, 0, errors.Errorf("plus can only be used as first or last character: %q", str)
			}
		}
	}

	// remove the sign from the string and trim space in case the removal left one
	trimmedSignsStr := strings.TrimSpace(str[skipFirst : len(str)-skipLast])
	for i, r := range trimmedSignsStr {
		switch {
		case r >= '0' && r <= '9':
			lastDigitIndex = i

		case r == '.' || r == ',' || r == '\'':
			if pointWritten {
				return 0, 0, 0, 0, errors.Errorf("no further separators allowed after decimal separator: %q", str)
			}

			// Write everything after the lastNonDigitIndex and before current index
			floatBuilder.WriteString(trimmedSignsStr[lastNonDigitIndex+1 : i])

			if numGroupingRunes == 0 {
				// This is the first grouping rune, just save it
				numGroupingRunes = 1
				lastGroupingRune = r
				lastGroupingIndex = i
			} else {
				// It's a further grouping rune, has to be 3 bytes since last grouping rune
				if i-(lastGroupingIndex+1) != 3 {
					return 0, 0, 0, 0, errors.Errorf("thousands separators have to be 3 characters apart: %q", str)
				}
				numGroupingRunes++
				if r == lastGroupingRune {
					if numGroupingRunes == 2 {
						if floatBuilder.Len()-numMinus > 6 {
							return 0, 0, 0, 0, errors.Errorf("thousands separators have to be 3 characters apart: %q", str)
						}
					}
					// If it's the same grouping rune, then just save it
					lastGroupingRune = r
					lastGroupingIndex = i
				} else {
					// If it's a different grouping rune, then we have
					// reached the decimal separator
					floatBuilder.WriteByte('.')
					pointWritten = true
					thousandsSep = byte(lastGroupingRune)
					decimalSep = byte(r)
				}
			}
			lastNonDigitIndex = i

		case r == ' ':
			if pointWritten {
				return 0, 0, 0, 0, errors.Errorf("no further separators allowed after decimal separator: %q", str)
			}

			// Write everything after the lastNonDigitIndex and before current index
			floatBuilder.WriteString(trimmedSignsStr[lastNonDigitIndex+1 : i])

			if numGroupingRunes == 0 {

				// This is the first grouping rune, just save it
				numGroupingRunes = 1
				lastGroupingRune = r
				lastGroupingIndex = i
			} else {
				// It's a further grouping rune, has to be 3 bytes since last grouping rune
				if i-(lastGroupingIndex+1) != 3 {
					return 0, 0, 0, 0, errors.Errorf("thousands separators have to be 3 characters apart: %q", str)
				}

				numGroupingRunes++
				if r == lastGroupingRune {
					if numGroupingRunes == 2 {
						if floatBuilder.Len()-numMinus > 6 {
							return 0, 0, 0, 0, errors.Errorf("thousands separators have to be 3 characters apart: %q", str)
						}
					}
					// If it's the same grouping rune, then just save it
					lastGroupingRune = r
					lastGroupingIndex = i
				} else {
					// Spaces only are used as thousands separators.
					// If the the last separator was not a space, something is wrong
					return 0, 0, 0, 0, errors.Errorf("space can not be used after another thousands separator: %q", str)
				}
			}
			lastNonDigitIndex = i

		case r == 'e':
			if i == 0 || eWritten {
				return 0, 0, 0, 0, errors.Errorf("e can't be the first or a repeating character: %q", str)
			}
			if numGroupingRunes > 0 && !pointWritten {
				floatBuilder.WriteByte('.')
				pointWritten = true
				decimalSep = '.'
			}
			floatBuilder.WriteString(trimmedSignsStr[lastNonDigitIndex+1 : i])
			lastNonDigitIndex = i

			floatBuilder.WriteByte('e')
			eWritten = true

		default:
			return 0, 0, 0, 0, errors.Errorf("invalid rune '%s' in %#v", string(r), str)
		}
	}

	if numGroupingRunes > 0 && !pointWritten {
		if numGroupingRunes > 1 {
			// If more than one grouping rune has been written, but no point
			// then it was pure integer grouping, so the last there
			// have to be 3 bytes since last grouping rune
			if lastDigitIndex-lastGroupingIndex != 3 {
				return 0, 0, 0, 0, errors.Errorf("thousands separators have to be 3 characters apart: %q", str)
			}
			thousandsSep = byte(lastGroupingRune)
		} else {
			floatBuilder.WriteByte('.')
			pointWritten = true
			decimalSep = byte(lastGroupingRune)
		}
	}
	if lastDigitIndex >= lastNonDigitIndex {
		floatBuilder.WriteString(trimmedSignsStr[lastNonDigitIndex+1 : lastDigitIndex+1])
	}

	floatStr := floatBuilder.String()
	f, err = strconv.ParseFloat(floatStr, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	pointPos := strings.IndexByte(floatStr, '.')
	if pointPos != -1 {
		if eWritten {
			ePos := strings.IndexByte(floatStr, 'e')
			decimals = ePos - (pointPos + 1)
		} else {
			decimals = len(floatStr) - (pointPos + 1)
		}

	}
	return f, thousandsSep, decimalSep, decimals, nil
}
