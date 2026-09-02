package nullable

import (
	"github.com/domonda/go-types/internal"
)

// SplitArray splits an SQL or JSON array into its top level elements.
// Array elements that are quoted strings will not be unquoted,
// use SplitArrayValues to get the values of quoted string elements.
// Returns nil in case of an empty array ("{}" or "[]").
// Passing "null" or "NULL" as array will return nil without an error.
//
// An unquoted NULL element of an SQL array (null in a JSON array)
// is returned as the unquoted string "NULL" ("null").
//
// A trailing comma announces an element that is not there and yields an
// empty string as last element. PostgreSQL rejects such a literal, this
// parser reports the announced element instead of dropping it silently.
func SplitArray(array string) ([]string, error) {
	return internal.SplitArray(array)
}

// SplitArrayValues splits an SQL or JSON array into its top level elements
// like SplitArray and returns the value of every element that is a
// double quoted string with the quotes removed and the escape sequences
// of the parsed array syntax undone.
// Returns nil in case of an empty array ("{}" or "[]").
// Passing "null" or "NULL" as array will return nil without an error.
//
// A backslash escapes the following character in a PostgreSQL array
// element whether it is quoted or not, so every element of an SQL array
// is unescaped. In a JSON array only double quoted strings are, every
// other element, like the objects of a JSON array of objects, is
// returned unchanged.
//
// An unquoted NULL element of an SQL array (null in a JSON array) is
// returned as the string "NULL" ("null") and is indistinguishable from a
// quoted "NULL" ("null") string element. Use SplitArray to tell them
// apart, its elements correspond by index and a quoted element still has
// its quotes there.
//
// Returns an error for a quoted element of a JSON array that is not a
// valid JSON string, like one with an invalid escape sequence.
func SplitArrayValues(array string) ([]string, error) {
	return internal.SplitArrayValues(array)
}

// SQLArrayLiteral joins the passed strings as an SQL array literal.
// A nil slice will produce NULL, pass an empty non nil slice to
// get the empty SQL array literal {}
// (use notnull.SQLArrayLiteral if a nil slice should also produce {}).
//
// The result uses the PostgreSQL array text format ({"a","b"}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
func SQLArrayLiteral(s []string) string {
	return internal.SQLArrayLiteral(s)
}
