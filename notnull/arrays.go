package notnull

import (
	"github.com/domonda/go-types/internal"
)

// SplitArray splits an SQL or JSON array into its top level elements.
// Array elements that are quoted strings will not be unquoted.
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
