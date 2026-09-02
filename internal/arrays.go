package internal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/domonda/go-types/strutil"
)

// SplitArray splits an SQL or JSON array into its top level elements.
// Array elements that are quoted strings will not be unquoted,
// use SplitArrayValues to get the values of quoted string elements.
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
		escaped      bool
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
			switch {
			case escaped:
				// An escaped character can't end the element
				escaped = false
			case r == '\\' && quoteRune == '"':
				escaped = true
			case r == quoteRune:
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

// SplitArrayValues splits an SQL or JSON array into its top level elements
// like SplitArray and returns the value of every element that is a
// double quoted string with the quotes removed and the escape sequences
// of the parsed array syntax undone.
//
// Elements that are not double quoted strings are returned unchanged,
// like the objects of a JSON array of objects, or elements quoted with
// single quotes, which are neither valid SQL nor valid JSON array syntax.
//
// An unquoted NULL element of an SQL array (null in a JSON array) is
// returned as the string "NULL" ("null") and is indistinguishable from a
// quoted "NULL" ("null") string element. Use SplitArray to tell them
// apart, its elements correspond by index and a quoted element still has
// its quotes there.
func SplitArrayValues(array string) ([]string, error) {
	elems, err := SplitArray(array)
	if err != nil || elems == nil {
		return nil, err
	}
	// SplitArray has validated the array syntax,
	// so a leading '[' means JSON and a leading '{' SQL
	jsonSyntax := array[0] == '['
	for i, elem := range elems {
		elems[i] = unquoteArrayElem(elem, jsonSyntax)
	}
	return elems, nil
}

// unquoteArrayElem returns the value of an array element that is a double
// quoted string, or the element unchanged if it is not a quoted string.
//
// SQL and JSON string elements have to be unescaped differently:
// a backslash in a PostgreSQL array literal escapes the following
// character whatever it is, so `\n` are the two characters backslash
// and n, while JSON interprets \n, \t, \uXXXX and friends.
func unquoteArrayElem(elem string, jsonSyntax bool) string {
	if len(elem) < 2 || elem[0] != '"' || elem[len(elem)-1] != '"' {
		return elem
	}
	if jsonSyntax {
		var s string
		err := json.Unmarshal([]byte(elem), &s)
		if err != nil {
			// Return a syntactically invalid JSON string unchanged
			// so it fails visibly downstream instead of being half
			// unescaped. Note that json.Unmarshal accepts a lone
			// surrogate and yields U+FFFD for it, so that one is
			// not returned unchanged.
			return elem
		}
		return s
	}
	return unescapeSQLArrayElem(elem[1 : len(elem)-1])
}

// unescapeSQLArrayElem undoes the backslash escaping within a quoted
// PostgreSQL array element, where a backslash escapes the following
// character whatever that character is.
//
// Implemented here instead of delegating to the vendored lib/pq parser
// in internal/pq because that parser can only parse a complete
// PostgreSQL array literal, not a single element, and neither the JSON
// arrays nor the unquoted [a,b] form that SplitArray also accepts.
// The escaping rule is the one of that parser, so SplitArrayValues and
// notnull.StringArray.Scan unescape a quoted element identically. They
// still differ on the literals that parser rejects and this one passes
// through as text, like one with a NULL element or more than one
// dimension.
func unescapeSQLArrayElem(elem string) string {
	if !strings.Contains(elem, `\`) {
		return elem
	}
	var b strings.Builder
	b.Grow(len(elem))
	for i := 0; i < len(elem); i++ {
		if elem[i] == '\\' && i+1 < len(elem) {
			i++
		}
		b.WriteByte(elem[i])
	}
	return b.String()
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

// escapeQuotedReplacer escapes the only two characters that are special
// within a quoted PostgreSQL array element. A replacer is used instead of
// two strings.ReplaceAll calls because it does a single left to right pass
// and therefore can't escape the backslashes it has inserted itself.
var escapeQuotedReplacer = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

func escapeQuoted(s string) string {
	return escapeQuotedReplacer.Replace(s)
}
