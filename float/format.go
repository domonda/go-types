package float

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// Format a float similar to strconv.Format with the 'f' format option,
// but with decimalSep as decimal separator instead of a point
// and optional thousands grouping of the integer part.
// Valid values for decimalSep are '.' and ','.
// If thousandsSep is not zero, then the integer part of the number is grouped
// with thousandsSep between every group of 3 digits from right to left.
// Valid rune values for thousandsSep are 0, ',', '.', ' ', "'"
// and thousandsSep must be different from decimalSep.
// The precision argument is the number of fractional digits the value is
// rounded to half away from zero. Like RoundToDecimals it rounds the
// shortest decimal representation of f, which is the decimal literal that
// parses back to f, instead of its binary value, so 1.005 formats as "1.01"
// and 2.5 as "3" with precisions 2 and 0, where strconv.FormatFloat rounds
// the binary value 1.00499999999999989342 to "1.00" and the tie 2.5 to even.
// The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return f exactly.
// Unlike by strconv.FormatFloat, negative zero and negative values that
// round to zero are formatted without a minus sign, like "0.00".
// If padPrecision is true, the fractional part is kept at exactly precision
// digits, padding the end with '0' characters when needed.
// If padPrecision is false, trailing fractional zeros are trimmed to return
// the shortest representation of the value rounded to precision.
// See: https://en.wikipedia.org/wiki/Decimal_separator
func Format[T ~float32 | ~float64](f T, thousandsSep, decimalSep rune, precision int, padPrecision bool) string {
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

	bitSize := reflect.TypeFor[T]().Bits()
	str := strconv.FormatFloat(float64(f), 'f', -1, bitSize)
	// NaN and ±Inf have no integer part to group or fractional part
	// to trim, so return strconv's representation unchanged.
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return str
	}
	if precision >= 0 {
		str = roundFormatted(str, precision)
	}
	// Negative zero and negative values rounded to zero are formatted
	// without a sign, because "-0.00" is not a different number than "0.00"
	if str[0] == '-' && strings.Trim(str[1:], "0.") == "" {
		str = str[1:]
	}

	// With a non-negative precision str has exactly precision fractional
	// digits. When padPrecision is false those trailing zeros are trimmed
	// (along with a then-dangling point) to return the shortest
	// representation of the value rounded to precision.
	if !padPrecision && strings.IndexByte(str, '.') != -1 {
		str = strings.TrimRight(str, "0")
		str = strings.TrimSuffix(str, ".")
	}

	pointPos := strings.IndexByte(str, '.')
	if pointPos == -1 {
		pointPos = len(str)
	}
	prefixLen := 0
	if str[0] == '-' {
		prefixLen = 1
	}
	integerLen := pointPos - prefixLen
	// Checking the formatted integer digits instead of f also groups
	// a value like 999.995 that rounding carried to 1000
	if thousandsSep != 0 && integerLen > 3 {
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
		for i := range numGroupSeps {
			b.WriteRune(thousandsSep)
			start := firstGroupLen + i*3
			b.WriteString(str[start : start+3])
		}

		if pointPos != len(str) {
			b.WriteRune(decimalSep)
			b.WriteString(str[pointPos+1:])
		}

		return b.String()
	}

	if decimalSep != '.' {
		if dot := strings.IndexByte(str, '.'); dot != -1 {
			var b strings.Builder
			b.Grow(len(str))
			b.WriteString(str[:dot])
			// decimalSep is validated to be '.' or ',' above — always ASCII.
			b.WriteByte(byte(decimalSep))
			b.WriteString(str[dot+1:])
			return b.String()
		}
	}

	return str
}

// roundFormatted rounds the decimal number str, as formatted by
// strconv.FormatFloat with the 'f' format, half away from zero
// to exactly precision fractional digits.
func roundFormatted(str string, precision int) string {
	_, fracDigits, hasPoint := strings.Cut(str, ".")
	if len(fracDigits) <= precision {
		// Nothing to round, only zeros to pad
		if precision == 0 {
			return str
		}
		var b strings.Builder
		b.Grow(len(str) + 1 + precision - len(fracDigits))
		b.WriteString(str)
		if !hasPoint {
			b.WriteByte('.')
		}
		for range precision - len(fracDigits) {
			b.WriteByte('0')
		}
		return b.String()
	}
	sign := ""
	if str[0] == '-' {
		sign, str = "-", str[1:]
	}
	intDigits, fracDigits, _ := strings.Cut(str, ".")
	// The digits of the rounded number times 10^precision, which are at
	// least len(intDigits)+precision digits because intDigits is never empty
	digits := roundDigitsHalfAwayFromZero(intDigits, fracDigits, precision)
	if precision == 0 {
		return sign + digits
	}
	intLen := len(digits) - precision
	return sign + digits[:intLen] + "." + digits[intLen:]
}
