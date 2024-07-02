package float

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/domonda/go-types/strutil"
)

// Parse a float string compatible with Format.
// If a separator was not detected, then zero will be returned for thousandsSep or decimalSep.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func Parse(str string) (float64, error) {
	f, _, _, _, err := ParseDetails(str)
	return f, err
}

// ParseDetails parses a float string compatible with Format
// and returns the detected integer thousands separator and decimal separator characters.
// If a separator was not detected, then zero will be returned for thousandsSep or decimalSep.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func ParseDetails(str string) (f float64, thousandsSep, decimalSep rune, decimals int, err error) {
	str = strutil.TrimSpace(str)

	switch str {
	case "":
		return 0, 0, 0, 0, errors.New("empty string can't be parsed as float")
	case "NaN":
		return math.NaN(), 0, 0, 0, nil
	case "+Inf", "Inf":
		return math.Inf(1), 0, 0, 0, nil
	case "-Inf":
		return math.Inf(-1), 0, 0, 0, nil
	}

	var (
		lastDigitIndex    = -1
		lastNonDigitIndex = -1

		pointWritten = false
		eIndex       = -1

		numMinus          int
		numGroupingRunes  int
		lastGroupingRune  rune
		lastGroupingIndex int

		skipFirst int // skip first bytes of str
		skipLast  int // skip last bytes of str

		floatBuilder strings.Builder
	)

	floatBuilder.Grow(len(str))

	// detect the sign, allowed positions are start and end
	for i, r := range str {
		switch {
		case r == 'e', r == 'E':
			eIndex = i

		case r == '-':
			switch {
			case i == 0:
				skipFirst = 1
			case i == len(str)-1:
				skipLast = 1
			case i == eIndex+1:
				continue
			default:
				return 0, 0, 0, 0, fmt.Errorf("minus can only be used as first or last character: %q", str)
			}
			floatBuilder.WriteByte(byte(r))
			numMinus = 1

		case r == '+':
			switch {
			case i == 0:
				skipFirst = 1
			case i == len(str)-1:
				skipLast = 1
			case i == eIndex+1:
				continue
			default:
				return 0, 0, 0, 0, fmt.Errorf("plus can only be used as first or last character: %q", str)
			}
		}
	}

	eIndex = -1

	// remove the sign from the string and trim space in case the removal left one
	trimmedSignsStr := strutil.TrimSpace(str[skipFirst : len(str)-skipLast])
	for i, r := range trimmedSignsStr {
		switch {
		case r >= '0' && r <= '9':
			lastDigitIndex = i

		case r == '.' || r == ',' || r == '\'':
			if pointWritten {
				return 0, 0, 0, 0, fmt.Errorf("no further separators allowed after decimal separator: %q", str)
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
					return 0, 0, 0, 0, fmt.Errorf("thousands separators have to be 3 characters apart: %q", str)
				}
				numGroupingRunes++
				if r == lastGroupingRune {
					if numGroupingRunes == 2 {
						if floatBuilder.Len()-numMinus > 6 {
							return 0, 0, 0, 0, fmt.Errorf("thousands separators have to be 3 characters apart: %q", str)
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
					thousandsSep = lastGroupingRune
					decimalSep = r
				}
			}
			lastNonDigitIndex = i

		case r == ' ':
			if pointWritten {
				return 0, 0, 0, 0, fmt.Errorf("no further separators allowed after decimal separator: %q", str)
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
					return 0, 0, 0, 0, fmt.Errorf("thousands separators have to be 3 characters apart: %q", str)
				}

				numGroupingRunes++
				if r == lastGroupingRune {
					if numGroupingRunes == 2 {
						if floatBuilder.Len()-numMinus > 6 {
							return 0, 0, 0, 0, fmt.Errorf("thousands separators have to be 3 characters apart: %q", str)
						}
					}
					// If it's the same grouping rune, then just save it
					lastGroupingRune = r
					lastGroupingIndex = i
				} else {
					// Spaces only are used as thousands separators.
					// If the the last separator was not a space, something is wrong
					return 0, 0, 0, 0, fmt.Errorf("space can not be used after another thousands separator: %q", str)
				}
			}
			lastNonDigitIndex = i

		case r == 'e', r == 'E':
			if i == 0 || eIndex != -1 {
				return 0, 0, 0, 0, fmt.Errorf("e can't be the first or a repeating character: %q", str)
			}
			if numGroupingRunes > 0 && !pointWritten {
				floatBuilder.WriteByte('.')
				pointWritten = true
				decimalSep = '.'
			}
			floatBuilder.WriteString(trimmedSignsStr[lastNonDigitIndex+1 : i+1]) // i+1 to write including the 'e'
			lastNonDigitIndex = i
			eIndex = i

		case (r == '-' || r == '+') && i == eIndex+1:
			floatBuilder.WriteRune(r)
			lastNonDigitIndex = i

		default:
			return 0, 0, 0, 0, fmt.Errorf("invalid rune '%s' in %q", string(r), str)
		}
	}

	if numGroupingRunes > 0 && !pointWritten {
		if numGroupingRunes > 1 {
			// If more than one grouping rune has been written, but no point
			// then it was pure integer grouping, so the last there
			// have to be 3 bytes since last grouping rune
			if lastDigitIndex-lastGroupingIndex != 3 {
				return 0, 0, 0, 0, fmt.Errorf("thousands separators have to be 3 characters apart: %q", str)
			}
			thousandsSep = lastGroupingRune
		} else {
			floatBuilder.WriteByte('.')
			pointWritten = true
			decimalSep = lastGroupingRune
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
		if eIndex != -1 {
			ePos := strings.LastIndexAny(floatStr, "eE")
			decimals = ePos - (pointPos + 1)
		} else {
			decimals = len(floatStr) - (pointPos + 1)
		}

	}
	return f, thousandsSep, decimalSep, decimals, nil
}
