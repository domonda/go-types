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
// rounded to. The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return f exactly.
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
	str := strconv.FormatFloat(float64(f), 'f', precision, bitSize)
	// NaN and ±Inf have no integer part to group or fractional part
	// to trim, so return strconv's representation unchanged.
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return str
	}

	// strconv.FormatFloat with a non-negative precision always emits exactly
	// precision fractional digits. When padPrecision is false those trailing
	// zeros are trimmed (along with a then-dangling point) to return the
	// shortest representation of the value rounded to precision.
	if !padPrecision && strings.IndexByte(str, '.') != -1 {
		str = strings.TrimRight(str, "0")
		str = strings.TrimSuffix(str, ".")
	}

	if thousandsSep != 0 && math.Abs(float64(f)) >= 1000 {
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
