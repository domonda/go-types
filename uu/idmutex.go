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
	locksMtx sync.Mutex
	locks    map[ID]*locker
}

type locker struct {
	mutex sync.Mutex
	count int
}

// NewIDMutex returns a new IDMutex
func NewIDMutex() *IDMutex {
	return &IDMutex{locks: make(map[ID]*locker)}
}

// Lock the mutex for a given key.
func (m *IDMutex) Lock(key ID) {
	m.locksMtx.Lock()
	lock := m.locks[key]
	if lock == nil {
		lock = new(locker)
		m.locks[key] = lock
	}
	lock.count++
	m.locksMtx.Unlock()

	lock.mutex.Lock()
}

// Unlock the mutex for a given key.
func (m *IDMutex) Unlock(key ID) {
	m.locksMtx.Lock()
	defer m.locksMtx.Unlock()

	lock := m.locks[key]
	if lock == nil {
		panic(fmt.Sprintf("uu.IDMutex.Unlock called for non locked key: %s", key))
	}
	lock.count--
	if lock.count == 0 {
		delete(m.locks, key)
	}
	lock.mutex.Unlock()
}
