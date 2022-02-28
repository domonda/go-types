package nullable

import (
	"github.com/domonda/go-types/internal"
)

// SplitArray splits an SQL or JSON array into its top level elements.
// Returns nil in case of an empty array ("{}" or "[]").
// Passing "null" or "NULL" as array will return nil without an error.
func SplitArray(array string) ([]string, error) {
	return internal.SplitArray(array)
}

// SQLArrayLiteral joins the passed strings as an SQL array literal
// A nil slice will produce NULL, pass an empty non nil slice to
// get the empty SQL array literal {}.
func SQLArrayLiteral(s []string) string {
	return internal.SQLArrayLiteral(s)
}
