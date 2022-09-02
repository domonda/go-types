package uu

import (
	"sync"
)

// IDMutex manages a unique mutex for every locked UUID key.
// The mutex for a key exists as long as there are any locks
// waiting to be unlocked.
// This is equivalent to declaring a mutex variable for every key,
// except that the key and the number of mutexes are dynamic.
type IDMutex struct {
	global sync.Mutex
	locks  map[ID]*countedLock

	onLock   func(ID)
	onUnlock func(ID)
}

type countedLock struct {
	sync.Mutex
	count int
}

// NewIDMutex returns a new IDMutex
func NewIDMutex() *IDMutex {
	return &IDMutex{locks: make(map[ID]*countedLock)}
}

// NewIDMutexWithCallbacks returns a new IDMutex
// where the passed functions are called back
// as first operation from Lock and Unlock
// before any any locking or unlocking.
// Usable for debugging and logging.
func NewIDMutexWithCallbacks(onLock, onUnlock func(ID)) *IDMutex {
	return &IDMutex{
		locks:    make(map[ID]*countedLock),
		onLock:   onLock,
		onUnlock: onUnlock,
	}
}

// Lock the mutex for a given ID.
func (m *IDMutex) Lock(id ID) {
	if m.onLock != nil {
		m.onLock(id)
	}

	m.global.Lock()
	lock := m.locks[id]
	if lock == nil {
		lock = new(countedLock)
		m.locks[id] = lock
	}
	lock.count++
	m.global.Unlock()

	lock.Lock()
}

// Unlock the mutex for a given ID.
func (m *IDMutex) Unlock(id ID) {
	if m.onUnlock != nil {
		m.onUnlock(id)
	}

	m.global.Lock()
	defer m.global.Unlock()

	lock := m.locks[id]
	if lock == nil {
		panic("uu.IDMutex.Unlock called for non locked key " + id.String())
	}
	lock.count--
	if lock.count == 0 {
		delete(m.locks, id)
	}
	lock.Unlock()
}

// IsLocked tells wether an ID is locked.
func (m *IDMutex) IsLocked(id ID) bool {
	m.global.Lock()
	_, locked := m.locks[id]
	m.global.Unlock()
	return locked
}
