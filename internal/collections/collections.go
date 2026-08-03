// Package collections reproduces the abstract collection interfaces
// documented by the Go 1.28 collections proposal, https://go.dev/issue/80590.
//
// The proposal declares them unexported (_AbstractCollection, _AbstractMap,
// _AbstractSet) purely as documentation of the shape that container types
// like *hash.Map, *hash.Set, *ordered.Map and set.Set have in common.
// They are reproduced here, exported within this module but not part of its
// public API, so that compile-time assertions in the packages of this module
// can prove that their container types implement them.
//
// Keep the method sets in sync with the proposal verbatim. They are the
// specification the container types of this module are checked against.
//
// The proposal constrains the collection type parameter to the interface
// itself, as in Collection[E any, C Collection[E, C]]. Such a self-referential
// declaration is only accepted by the Go 1.26 type checker; Go 1.25, which
// this module still targets, rejects it with "invalid recursive type". The
// collection type parameter is therefore declared as any here. That is a
// weaker constraint, but it does not weaken what these interfaces are used
// for: every instantiation names a concrete type, and the compile-time
// assertions still check the full method set. Restore the self-reference
// once this module requires Go 1.26.
package collections

import "iter"

// Collection models a collection C of elements E,
// such as *hash.Map, *hash.Set, *ordered.Map, or set.Set.
type Collection[E, C any] interface {
	Clear()
	Clone() C
	Contains(E) bool
	ContainsAll(iter.Seq[E]) bool
	Len() int
	String() string
}

// Map models a mapping M from keys K to values V,
// such as *hash.Map or *ordered.Map.
//
// No type of this module implements Map; it is kept here as the reference
// shape for any future key/value container type.
type Map[K, V, M any] interface {
	Collection[K, M]

	All() iter.Seq2[K, V]
	At(K) V
	Delete(K) (V, bool)
	DeleteAll(iter.Seq[K]) bool
	DeleteFunc(func(K, V) bool) bool
	Get(K) (V, bool)
	Keys() iter.Seq[K]
	Set(K, V) (V, bool)
	SetAll(iter.Seq2[K, V]) bool
	Values() iter.Seq[V]
}

// Set models a set S of elements E,
// such as *hash.Set, or set.Set.
type Set[E, S any] interface {
	Collection[E, S]

	All() iter.Seq[E]
	Delete(E) bool
	DeleteAll(iter.Seq[E]) bool
	DeleteFunc(func(E) bool) bool
	Difference(S) S
	DifferenceWith(S)
	Equal(S) bool
	Insert(E) bool
	InsertAll(iter.Seq[E]) bool
	Intersection(S) S
	IntersectionWith(S)
	Intersects(S) bool
	SymmetricDifference(S) S
	SymmetricDifferenceWith(S)
	Union(S) S
	UnionWith(S)
}
