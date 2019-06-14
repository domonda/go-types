package strutil

import (
	"fmt"
	"sync"
)

// StrMutex allows mutex locking per UUID
type StrMutex struct {
	mapMutex   sync.Mutex
	strMutexes map[string]*sync.Mutex
}

func NewStrMutex() *StrMutex {
	return &StrMutex{strMutexes: make(map[string]*sync.Mutex)}
}

func (m *StrMutex) Lock(str string) {
	m.mapMutex.Lock()
	strMutex, ok := m.strMutexes[str]
	if !ok {
		strMutex = new(sync.Mutex)
		m.strMutexes[str] = strMutex
	}
	m.mapMutex.Unlock()

	strMutex.Lock()
}

func (m *StrMutex) Unlock(str string) {
	m.mapMutex.Lock()
	strMutex, ok := m.strMutexes[str]
	// delete(m.strMutexes, str) // TODO not thread safe!
	m.mapMutex.Unlock()

	if !ok {
		panic(fmt.Sprintf("Unlock called for non locked string: %q", str))
	}

	strMutex.Unlock()
}
