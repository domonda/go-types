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
//
// A trailing comma announces an element that is not there and yields an
// empty string as last element. PostgreSQL rejects such a literal, this
// parser reports the announced element instead of dropping it silently.
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
		inElemString bool
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
				// A backslash escapes the following rune here as well as
				// within the element, else {\,b} would be split at a
				// comma that PostgreSQL reads as part of the value
				escaped = r == '\\' && isSQL
				state = inElem
			}

		case inElem:
			switch {
			case escaped:
				// An escaped rune is part of the element value
				// and can never be array syntax
				escaped = false
			case r == '\\' && (isSQL || inElemString):
				// A backslash escapes the following rune in a PostgreSQL
				// array literal, in an unquoted element as well as in a
				// quoted one, so {a\,b} is the single element a,b, and
				// within a string of a nested JSON object or array
				escaped = true
			case inElemString:
				// Within a nested object or array the structural runes
				// are part of the string value and must not be counted
				if r == '"' {
					inElemString = false
				}
			case r == '"' && (objectDepth > 0 || bracketDepth > 0):
				// Only a double quote within a nested object or array
				// starts a string. At the top level of an element it is
				// literal text of an unquoted SQL element that begins
				// with something else: {'double " quote'}
				inElemString = true
			case r == ',':
				if objectDepth == 0 && bracketDepth == 0 {
					elems = append(elems, inner[elemStart:i])
					elemStart = -1
					state = beforeElem
				}
			case r == '{':
				// Only nested within an already open object or array,
				// an unmatched '{' at the top level of an element is
				// tolerated as literal text: {in{o@example.com}
				if objectDepth > 0 || bracketDepth > 0 {
					objectDepth++
				}
			case r == '[':
				if objectDepth > 0 || bracketDepth > 0 {
					bracketDepth++
				}
			case r == '}':
				objectDepth--
				if objectDepth < 0 {
					return nil, fmt.Errorf("array %q has too many '}'", array)
				}
			case r == ']':
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

	// Checked before the depth, an unclosed string within a nested
	// object or array is what leaves that object or array open
	if inElemString {
		return nil, fmt.Errorf(`array %q has an unclosed " quote`, array)
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
	if state == beforeElem {
		// A trailing comma. TrimSpace has removed every rune that
		// beforeElem skips, so a non empty array can only end in this
		// state after a comma announced another element.
		elems = append(elems, "")
	}

	return elems, nil
}

// SplitArrayValues splits an SQL or JSON array into its top level elements
// like SplitArray and returns the value of every element that is a
// double quoted string with the quotes removed and the escape sequences
// of the parsed array syntax undone.
//
// A backslash escapes the following character in a PostgreSQL array
// element whether it is quoted or not, so every element of an SQL array
// is unescaped. In a JSON array only double quoted strings are, every
// other element, like the objects of a JSON array of objects, is
// returned unchanged.
//
// Returns an error for a quoted element of a JSON array that is not a
// valid JSON string, like one with an invalid escape sequence.
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
		value, err := unquoteArrayElem(elem, jsonSyntax)
		if err != nil {
			return nil, fmt.Errorf("array %q: %w", array, err)
		}
		elems[i] = value
	}
	return elems, nil
}

// unquoteArrayElem returns the value of an array element, with the quotes
// of a double quoted string removed and the escape sequences of the
// parsed array syntax undone.
//
// SQL and JSON elements have to be unescaped differently:
// a backslash in a PostgreSQL array literal escapes the following
// character whatever it is, so `\n` are the two characters backslash
// and n, while JSON interprets \n, \t, \uXXXX and friends. The
// backslash escaping applies to an unquoted PostgreSQL element too,
// while a JSON element that is not a double quoted string, like an
// object or a number, is returned unchanged.
func unquoteArrayElem(elem string, jsonSyntax bool) (string, error) {
	quoted := len(elem) >= 2 && elem[0] == '"' && elem[len(elem)-1] == '"'
	if !jsonSyntax {
		if quoted {
			elem = elem[1 : len(elem)-1]
		}
		return unescapeSQLArrayElem(elem), nil
	}
	if !quoted {
		return elem, nil
	}
	var s string
	err := json.Unmarshal([]byte(elem), &s)
	if err != nil {
		// Returning the element unchanged would hand out the value with
		// the quotes that this function exists to remove, which is not
		// distinguishable from a value that really has them. Note that
		// json.Unmarshal accepts a lone surrogate and yields U+FFFD for
		// it, so that one is a value here and not an error.
		return "", fmt.Errorf("invalid JSON string element %s: %w", elem, err)
	}
	return s, nil
}

// unescapeSQLArrayElem undoes the backslash escaping of a PostgreSQL
// array element, where a backslash escapes the following character
// whatever that character is.
//
// Implemented here instead of delegating to the vendored lib/pq parser
// in internal/pq because that parser can only parse a complete
// PostgreSQL array literal, not a single element, and neither the JSON
// arrays nor the unquoted [a,b] form that SplitArray also accepts.
// The rule is the one of the PostgreSQL array_in parser, verified
// against PostgreSQL 16, for a quoted and for an unquoted element. The
// vendored parser implements it for a quoted element only and reads
// {a\,b} as the two elements `a\` and `b`, where PostgreSQL and
// SplitArrayValues read the single value `a,b`. They also still differ
// on the literals that parser rejects and this one passes through as
// text, like one with a NULL element or more than one dimension, and
// this one keeps the trailing space of an unquoted element where
// PostgreSQL strips it. PostgreSQL quotes any element that needs the
// space, so only a hand written literal can tell that one apart.
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
