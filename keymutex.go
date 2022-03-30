package types

import (
	"fmt"
	"sync"
)

// KeyMutex manages a unique mutex for every locked key.
// The mutex for a key exists as long as there are any locks
// waiting to be unlocked.
// This is equivalent to declaring a mutex variable for every key,
// except that the key and the number of mutexes are dynamic.
type KeyMutex[T comparable] struct {
	global sync.Mutex
	locks  map[T]*lockCount
}

type lockCount struct {
	sync.Mutex
	count int
}

// NewKeyMutex returns a new KeyMutex
func NewKeyMutex[T comparable]() *KeyMutex[T] {
	return &KeyMutex[T]{locks: make(map[T]*lockCount)}
}

// Lock the mutex for a given key
func (m *KeyMutex[T]) Lock(key T) {
	m.global.Lock()
	l := m.locks[key]
	if l == nil {
		l = new(lockCount)
		m.locks[key] = l
	}
	l.count++
	m.global.Unlock()

	l.Lock()
}

// Unlock the mutex for a given key.
func (m *KeyMutex[T]) Unlock(key T) {
	m.global.Lock()
	defer m.global.Unlock()

	l := m.locks[key]
	if l == nil {
		panic(fmt.Sprintf("KeyMutex.Unlock called for non locked key: %#v", key))
	}
	l.count--
	if l.count == 0 {
		delete(m.locks, key)
	}
	l.Unlock()
}

// IsLocked tells wether a key is locked.
func (m *KeyMutex[T]) IsLocked(key T) bool {
	m.global.Lock()
	_, locked := m.locks[key]
	m.global.Unlock()
	return locked
}
