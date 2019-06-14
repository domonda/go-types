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
	strMutex := m.strMutexes[str]
	if strMutex == nil {
		strMutex = new(sync.Mutex)
		m.strMutexes[str] = strMutex
	}
	m.mapMutex.Unlock()

	strMutex.Lock()
}

func (m *StrMutex) Unlock(str string) {
	m.mapMutex.Lock()
	strMutex := m.strMutexes[str]
	// TODO: delete is not thread safe!
	// delete is only safe when no thread is waiting for an unlock anymore
	// delete(m.strMutexes, str)
	m.mapMutex.Unlock()

	if strMutex == nil {
		panic(fmt.Sprintf("Unlock called for non locked string: %q", str))
	}
	strMutex.Unlock()
}
