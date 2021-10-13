package float

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Format a float similar to strconv.Format with the 'f' format option,
// but with decimalSep as decimal separator instead of a point
// and optional thousands grouping of the integer part.
// Valid values for decimalSep are '.' and ','.
// If thousandsSep is not zero, then the integer part of the number is grouped
// with thousandsSep between every group of 3 digits from right to left.
// Valid values for thousandsSep are [0, ',', '.', '\'']
// and thousandsSep must be different from decimalSep.
// The precision argument controls the number of digits (excluding the exponent).
// The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return f exactly.
// If padPrecision is true and precision is greater zero,
// then the end of the fractional part will be padded with
// '0' characters to reach the length of precision.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func Format(f float64, thousandsSep, decimalSep rune, precision int, padPrecision bool) string {
	if thousandsSep != 0 && thousandsSep != '.' && thousandsSep != ',' && thousandsSep != ' ' && thousandsSep != '\'' {
		panic(fmt.Errorf("invalid thousandsSep: '%s'", string(thousandsSep)))
	}
	if decimalSep != '.' && decimalSep != ',' {
		panic(fmt.Errorf("invalid decimalSep: '%s'", string(decimalSep)))
	}
	if thousandsSep == decimalSep {
		panic(fmt.Errorf("thousandsSep == decimalSep: '%s'", string(thousandsSep)))
	}
	if precision < -1 {
		panic(fmt.Errorf("precision < -1: %d", precision))
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
			b.WriteRune(thousandsSep)
			start := firstGroupLen + i*3
			b.WriteString(str[start : start+3])
		}

		if pointPos != len(str) {
			b.WriteRune(decimalSep)
			fraction := str[pointPos+1:]
			b.WriteString(fraction)
			if padPrecision {
				for i := len(fraction); i < precision; i++ {
					b.WriteByte('0')
				}
			}
		} else if padPrecision && precision > 0 {
			b.WriteRune(decimalSep)
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
