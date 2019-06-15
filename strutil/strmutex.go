package strutil

import (
	"fmt"
	"sync"
)

// StrMutex manages a unique mutex for every locked string key.
// The mutex for a key exists as long as there are any locks
// waiting to be unlocked.
// This is equivalent to declaring a mutex variable for every key,
// except that the key and the number of mutexes are dynamic.
type StrMutex struct {
	locksMtx sync.Mutex
	locks    map[string]*locker
}

type locker struct {
	mutex sync.Mutex
	count int
}

// NewStrMutex returns a new StrMutex
func NewStrMutex() *StrMutex {
	return &StrMutex{locks: make(map[string]*locker)}
}

// Lock the mutex for a given key.
func (m *StrMutex) Lock(key string) {
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
func (m *StrMutex) Unlock(key string) {
	m.locksMtx.Lock()
	defer m.locksMtx.Unlock()

	lock := m.locks[key]
	if lock == nil {
		panic(fmt.Sprintf("Unlock called for non locked key: %q", key))
	}
	lock.count--
	if lock.count == 0 {
		delete(m.locks, key)
	}
	lock.mutex.Unlock()
}
