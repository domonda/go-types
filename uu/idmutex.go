package uu

import (
	"sync"
)

// IDMutex allows mutex locking per UUID
type IDMutex struct {
	mapMutex  sync.RWMutex
	idMutexes map[ID]*sync.Mutex
}

func NewIDMutex() *IDMutex {
	return &IDMutex{idMutexes: make(map[ID]*sync.Mutex)}
}

func (m *IDMutex) Lock(id ID) {
	m.mapMutex.Lock()
	idMutex, ok := m.idMutexes[id]
	if !ok {
		idMutex = new(sync.Mutex)
		m.idMutexes[id] = idMutex
	}
	m.mapMutex.Unlock()

	idMutex.Lock()
}

func (m *IDMutex) Unlock(id ID) {
	m.mapMutex.RLock()
	idMutex, ok := m.idMutexes[id]
	m.mapMutex.RUnlock()

	if !ok {
		panic("Unlock called for non locked UUID: " + id.String())
	}

	idMutex.Unlock()
}