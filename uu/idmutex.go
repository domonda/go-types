package uu

import (
	"fmt"
	"sync"
)

// IDMutex manages a unique mutex for every locked UUID key.
// The mutex for a key exists as long as there are any locks
// waiting to be unlocked.
// This is equivalent to declaring a mutex variable for every key,
// except that the key and the number of mutexes are dynamic.
type IDMutex struct {
	global sync.Mutex
	locks  map[ID]*locker
}

type locker struct {
	sync.Mutex
	count int
}

// NewIDMutex returns a new IDMutex
func NewIDMutex() *IDMutex {
	return &IDMutex{locks: make(map[ID]*locker)}
}

// Lock the mutex for a given ID.
func (m *IDMutex) Lock(id ID) {
	m.global.Lock()
	l := m.locks[id]
	if l == nil {
		l = new(locker)
		m.locks[id] = l
	}
	l.count++
	m.global.Unlock()

	l.Lock()
}

// Unlock the mutex for a given ID.
func (m *IDMutex) Unlock(id ID) {
	m.global.Lock()
	defer m.global.Unlock()

	l := m.locks[id]
	if l == nil {
		panic(fmt.Sprintf("uu.IDMutex.Unlock called for non locked key: %s", id))
	}
	l.count--
	if l.count == 0 {
		delete(m.locks, id)
	}
	l.Unlock()
}

// IsLocked tells wether an ID is locked.
func (m *IDMutex) IsLocked(id ID) bool {
	m.global.Lock()
	_, locked := m.locks[id]
	m.global.Unlock()
	return locked
}
