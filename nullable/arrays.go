package nullable

import (
	"github.com/domonda/go-types/internal"
)

// SplitArray splits an SQL or JSON array into its top level elements.
// Array elements that are quoted strings will not be unquoted.
// Returns nil in case of an empty array ("{}" or "[]").
// Passing "null" or "NULL" as array will return nil without an error.
func SplitArray(array string) ([]string, error) {
	return internal.SplitArray(array)
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
