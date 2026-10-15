// Package set provides generic helper functions for sets represented as
// plain map[T]struct{} values with T constrained to cmp.Ordered.
//
// The package defines no set type. [New] creates a set, [Add] adds values
// and allocates a nil set first, and [Contains], [ContainsAll],
// [ContainsAny], [ContainsAllOther] and [ContainsAnyOther] test membership.
// All functions accept a nil set as the empty set. Values of defined map
// types like the root package's types.Set can be passed directly.
//
// New code should use package github.com/domonda/go-types/mapset, which
// has the full set algebra, supports map[K]bool sets and any comparable
// element type, and mirrors the container/mapset package proposed for Go.
// This package is kept unchanged for its existing callers.
//
// See the [README] for the function table and a migration table to mapset.
//
// [README]: https://github.com/domonda/go-types/blob/master/set/README.md
package set
