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
// It is ParseDetails without the detected separators and decimals,
// see ParseDetails for the separator rules.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func Parse(str string) (float64, error) {
	f, _, _, _, err := ParseDetails(str)
	return f, err
}

// ParseDetails parses a float string compatible with Format
// and returns the detected integer thousands separator and decimal separator characters.
// If a separator was not detected, then zero will be returned for thousandsSep or decimalSep.
//
// One sign, '+' or '-', may be the first or the last character,
// but not both, so "+1-" and "-1-" are errors.
// A '-' right after the decimal separator means no cents,
// so "12,-" and "12.-" are 12.
//
// A single '.' or ',' is the decimal separator, so "1,234" parses as 1.234.
// The apostrophe and the space are only ever thousands separators,
// so "1'234" and "1 234" parse as 1234, and one that can't separate
// thousands, like in "1'5", is an error.
// Thousands separators have to be followed by exactly 3 digits
// and the group in front of the first one must have 1 to 3 digits.
// The returned decimalSep is always zero, '.' or ','.
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

	// A dash right after the decimal separator means no cents,
	// like the German "12,-" or the Swiss "12.-", and is no minus sign
	if strings.HasSuffix(str, ",-") || strings.HasSuffix(str, ".-") {
		str = str[:len(str)-1]
	}

	var (
		lastDigitIndex    = -1
		lastNonDigitIndex = -1

		pointWritten = false
		eIndex       = -1

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
			floatBuilder.WriteByte(byte(r)) //#nosec G115

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
	if skipFirst != 0 && skipLast != 0 {
		return 0, 0, 0, 0, fmt.Errorf("only one sign allowed, either as first or last character: %q", str)
	}

	eIndex = -1

	// groupsThousands reports whether the grouping runes found in front of the
	// index end without a decimal separator separate thousands, which more than
	// one of them and every apostrophe or space do, or whether a single '.' or
	// ',' is the decimal separator. It returns an error if they separate
	// thousands but the groups don't have the required number of digits.
	groupsThousands := func(end int) (bool, error) {
		if numGroupingRunes == 1 && (lastGroupingRune == '.' || lastGroupingRune == ',') {
			return false, nil
		}
		if end-(lastGroupingIndex+1) != 3 {
			return false, fmt.Errorf("thousands separators have to be 3 characters apart: %q", str)
		}
		if numGroupingRunes == 1 && (lastGroupingIndex < 1 || lastGroupingIndex > 3) {
			return false, fmt.Errorf("first group before a thousands separator must have 1 to 3 digits: %q", str)
		}
		return true, nil
	}

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
			if eIndex != -1 {
				return 0, 0, 0, 0, fmt.Errorf("no separators allowed in the exponent: %q", str)
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
				// The first grouping rune is a thousands separator now,
				// so the group in front of it must have 1 to 3 digits
				if numGroupingRunes == 2 && (lastGroupingIndex < 1 || lastGroupingIndex > 3) {
					return 0, 0, 0, 0, fmt.Errorf("first group before a thousands separator must have 1 to 3 digits: %q", str)
				}
				if r == lastGroupingRune {
					// If it's the same grouping rune, then just save it
					lastGroupingRune = r
					lastGroupingIndex = i
				} else {
					// If it's a different grouping rune, then we have
					// reached the decimal separator
					if r == '\'' {
						return 0, 0, 0, 0, fmt.Errorf("apostrophe can only be used as thousands separator: %q", str)
					}
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
			if eIndex != -1 {
				return 0, 0, 0, 0, fmt.Errorf("no separators allowed in the exponent: %q", str)
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
					// The first space is a thousands separator now,
					// so the group in front of it must have 1 to 3 digits
					if numGroupingRunes == 2 && (lastGroupingIndex < 1 || lastGroupingIndex > 3) {
						return 0, 0, 0, 0, fmt.Errorf("first group before a thousands separator must have 1 to 3 digits: %q", str)
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
				thousands, err := groupsThousands(i)
				if err != nil {
					return 0, 0, 0, 0, err
				}
				if thousands {
					thousandsSep = lastGroupingRune
				} else {
					floatBuilder.WriteByte('.')
					pointWritten = true
					decimalSep = lastGroupingRune
				}
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

	// Grouping runes in front of an exponent were already resolved there
	if numGroupingRunes > 0 && !pointWritten && eIndex == -1 {
		thousands, err := groupsThousands(lastDigitIndex + 1)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		if thousands {
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
