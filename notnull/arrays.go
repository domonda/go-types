package notnull

import (
	"github.com/domonda/go-types/internal"
)

// SplitArray splits an SQL or JSON array into its top level elements.
// Returns a non nil empty slice in case of an empty array ("{}" or "[]")
// or when passing "null" or "NULL" as array.
func SplitArray(array string) ([]string, error) {
	s, err := internal.SplitArray(array)
	if err != nil {
		return nil, err
	}
	if s == nil {
		s = []string{}
	}
	return s, nil
}

// SQLArrayLiteral joins the passed strings as an SQL array literal
// A nil slice will produce NULL, pass an empty non nil slice to
// get the empty SQL array literal {}.
func SQLArrayLiteral(s []string) string {
	if len(s) == 0 {
		return `{}`
	}
	return internal.SQLArrayLiteral(s)
}
