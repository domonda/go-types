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
