package internal

import (
	"fmt"
	"strings"
)

// SplitArray splits an SQL or JSON array into its top level elements.
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
	inner := strings.TrimSpace(array[1 : len(array)-1])
	if inner == "" {
		return nil, nil
	}
	var (
		elems        []string
		objectDepth  = 0
		bracketDepth = 0
		elemStart    = 0
		rLast        rune
		withinQuote  rune
	)
	for i, r := range inner {
		if withinQuote == 0 {
			switch r {
			case ',':
				if objectDepth == 0 && bracketDepth == 0 {
					elems = append(elems, strings.TrimSpace(inner[elemStart:i]))
					elemStart = i + 1
				}

			case '{':
				objectDepth++

			case '}':
				objectDepth--
				if objectDepth < 0 {
					return nil, fmt.Errorf("array %q has too many '}'", array)
				}

			case '[':
				bracketDepth++

			case ']':
				bracketDepth--
				if bracketDepth < 0 {
					return nil, fmt.Errorf("array %q has too many ']'", array)
				}

			case '"':
				// Begin JSON string
				withinQuote = r

			case '\'':
				// Begin SQL string
				withinQuote = r
			}
		} else {
			// withinQuote != 0
			switch withinQuote {
			case '\'':
				if r == '\'' && rLast != '\'' {
					// End of SQL quote because ' was not escapded as ''
					withinQuote = 0
				}
			case '"':
				if r == '"' && rLast != '\\' {
					// End of JSON quote because " was not escapded as \"
					withinQuote = 0
				}
			}
		}

		rLast = r
	}

	if objectDepth != 0 {
		return nil, fmt.Errorf("array %q has not enough '}'", array)
	}
	if bracketDepth != 0 {
		return nil, fmt.Errorf("array %q has not enough ']'", array)
	}
	if withinQuote != 0 {
		return nil, fmt.Errorf("array %q has an unclosed '%s' quote", array, string(withinQuote))
	}

	// Rameining element after begin and separators
	if elemStart < len(inner) {
		elems = append(elems, strings.TrimSpace(inner[elemStart:]))
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
