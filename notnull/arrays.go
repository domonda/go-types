package notnull

import (
	"github.com/domonda/go-types/internal"
)

// SplitArray splits an SQL or JSON array into its top level elements.
// Array elements that are quoted strings will not be unquoted,
// use SplitArrayValues to get the values of quoted string elements.
// Returns a non nil empty slice in case of an empty array ("{}" or "[]")
// or when passing "null" or "NULL" as array.
//
// An unquoted NULL element of an SQL array (null in a JSON array)
// is returned as the unquoted string "NULL" ("null").
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

// SplitArrayValues splits an SQL or JSON array into its top level elements
// like SplitArray and returns the value of every element that is a
// double quoted string with the quotes removed and the escape sequences
// of the parsed array syntax undone.
// Returns a non nil empty slice in case of an empty array ("{}" or "[]")
// or when passing "null" or "NULL" as array.
//
// Elements that are not double quoted strings are returned unchanged,
// like the objects of a JSON array of objects.
//
// An unquoted NULL element of an SQL array (null in a JSON array) is
// returned as the string "NULL" ("null") and is indistinguishable from a
// quoted "NULL" ("null") string element. Use SplitArray to tell them
// apart, its elements correspond by index and a quoted element still has
// its quotes there.
//
// A quoted element of a JSON array that is not a valid JSON string, like
// one with an invalid escape sequence, is returned unchanged with its
// quotes, so that it fails visibly downstream instead of being half
// unescaped.
func SplitArrayValues(array string) ([]string, error) {
	s, err := internal.SplitArrayValues(array)
	if err != nil {
		return nil, err
	}
	if s == nil {
		s = []string{}
	}
	return s, nil
}

// SQLArrayLiteral joins the passed strings as an SQL array literal.
// Both a nil and an empty slice produce the empty array literal {}
// (use nullable.SQLArrayLiteral if a nil slice should produce NULL).
//
// The result uses the PostgreSQL array text format ({"a","b"}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
func SQLArrayLiteral(s []string) string {
	if len(s) == 0 {
		return `{}`
	}
	return internal.SQLArrayLiteral(s)
}
