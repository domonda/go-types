package internal

import (
	"fmt"
	"strings"

	"github.com/domonda/go-types/strutil"
)

// SplitArray splits an SQL or JSON array into its top level elements.
// Array elements that are quoted strings will not be unquoted.
// Returns nil in case of an empty array ("{}" or "[]").
// Passing "null" or "NULL" as array will return nil without an error.
func SplitArray(array string) ([]string, error) {
	if len(array) < 2 {
		return nil, fmt.Errorf("%q is too short for an array", array)
	}
	first := array[0]
	last := array[len(array)-1]
	isJSON := first == '[' && last == ']'
	isSQL := first == '{' && last == '}'
	if !isJSON && !isSQL {
		if array == "null" || array == "NULL" {
			return nil, nil
		}
		return nil, fmt.Errorf("%q is not a SQL or JSON array", array)
	}
	inner := strutil.TrimSpace(array[1 : len(array)-1])
	if inner == "" {
		return nil, nil
	}
	const (
		beforeElem = iota
		afterElem
		inElem
		inQuotedElem
	)
	var (
		state        = beforeElem
		objectDepth  = 0
		bracketDepth = 0
		elemStart    = -1
		lastRune     rune
		quoteRune    rune
		elems        []string
	)
	for i, r := range inner {
		switch state {
		case beforeElem:
			switch r {
			case ' ', '\t', '\n', '\r':
				// skip
			case ',':
				return nil, fmt.Errorf("invalid comma before array element in %q", array)
			case '{':
				objectDepth++
				elemStart = i
				state = inElem
			case '[':
				bracketDepth++
				elemStart = i
				state = inElem
			case '"':
				quoteRune = r
				elemStart = i
				state = inQuotedElem
			case '\'':
				if isSQL {
					// Leading single quote is normal text: {'A,B}
					elemStart = i
					state = inElem
				} else {
					quoteRune = r
					elemStart = i
					state = inQuotedElem
				}
			default:
				elemStart = i
				state = inElem
			}

		case inElem:
			switch r {
			case ',':
				if objectDepth == 0 && bracketDepth == 0 {
					elems = append(elems, inner[elemStart:i])
					elemStart = -1
					state = beforeElem
				}
			case '}':
				objectDepth--
				if objectDepth < 0 {
					return nil, fmt.Errorf("array %q has too many '}'", array)
				}
			case ']':
				bracketDepth--
				if bracketDepth < 0 {
					return nil, fmt.Errorf("array %q has too many ']'", array)
				}
			}

		case inQuotedElem:
			if r == quoteRune && (r != '"' || lastRune != '\\') {
				elems = append(elems, inner[elemStart:i+1])
				elemStart = -1
				quoteRune = 0
				state = afterElem
			}

		case afterElem:
			switch r {
			case ' ', '\t', '\n', '\r':
				// skip
			case ',':
				state = beforeElem
			default:
				return nil, fmt.Errorf("invalid rune %q after array element in %q", r, array)
			}
		}

		lastRune = r
	}

	if objectDepth != 0 {
		return nil, fmt.Errorf("array %q has not enough '}'", array)
	}
	if bracketDepth != 0 {
		return nil, fmt.Errorf("array %q has not enough ']'", array)
	}
	if state == inQuotedElem {
		return nil, fmt.Errorf("array %q has an unclosed %s quote", array, string(quoteRune))
	}

	if state == inElem {
		elems = append(elems, inner[elemStart:])
	}

	return elems, nil
}

// SQLArrayLiteral joins the passed strings as an SQL array literal
// A nil slice will produce NULL, pass an empty non nil slice to
// get the empty SQL array literal {}.
func SQLArrayLiteral(s []string) string {
	if s == nil {
		return `NULL`
	}
	if len(s) == 0 {
		return `{}`
	}
	b := strings.Builder{}
	b.Grow(2 - 1 + len(s)*3 + len(s[0]))
	b.WriteString(`{"`)
	b.WriteString(escapeQuoted(s[0]))
	for i := 1; i < len(s); i++ {
		b.WriteString(`","`)
		b.WriteString(escapeQuoted(s[i]))
	}
	b.WriteString(`"}`)
	return b.String()
}

func escapeQuoted(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`) // ?
}
