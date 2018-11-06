package uuidmtx

import (
	"sync"

	uuid "github.com/ungerik/go-uuid"
)

type Mutex struct {
	mapMutex  sync.RWMutex
	idMutexes map[uuid.UUID]*sync.Mutex
}

func New() *Mutex {
	return &Mutex{idMutexes: make(map[uuid.UUID]*sync.Mutex)}
}

func (m *Mutex) Lock(id uuid.UUID) {
	m.mapMutex.Lock()
	idMutex, ok := m.idMutexes[id]
	if !ok {
		idMutex = new(sync.Mutex)
		m.idMutexes[id] = idMutex
	}
	m.mapMutex.Unlock()

	idMutex.Lock()
}

func (m *Mutex) Unlock(id uuid.UUID) {
	m.mapMutex.RLock()
	idMutex, ok := m.idMutexes[id]
	m.mapMutex.RUnlock()

	if !ok {
		panic("Unlock called for non locked UUID: " + id.String())
	}

	idMutex.Unlock()
}
