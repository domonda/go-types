package types

import "iter"

// Yield returns an iterator that yields a single value.
func Yield[V any](value V) iter.Seq[V] {
	return func(yield func(V) bool) {
		yield(value)
	}
}

// Yield2 returns an iterator that yields a single key-value pair.
func Yield2[K, V any](key K, value V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		yield(key, value)
	}
}

// YieldErr returns an iterator that yields a single key-value pair with
// the default value for the key type K and the passed error as the value.
func YieldErr[K any](err error) iter.Seq2[K, error] {
	return func(yield func(K, error) bool) {
		yield(*new(K), err)
	}
}

// Seq2NilError converts a sequence of values to a sequence of key-value pairs with
// the passed sequence values as keys and nil errors as values.
func Seq2NilError[K any](seq iter.Seq[K]) iter.Seq2[K, error] {
	return func(yield func(K, error) bool) {
		for v := range seq {
			if !yield(v, nil) {
				return
			}
		}
	}
}
