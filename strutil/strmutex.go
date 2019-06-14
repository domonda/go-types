package strutil

import (
	"fmt"
	"sync"
)

// StrMutex allows mutex locking per UUID
type StrMutex struct {
	mapMutex   sync.Mutex
	strMutexes map[string]struct {
		mtx *sync.Mutex
		wg  *sync.WaitGroup
	}
}

func NewStrMutex() *StrMutex {
	return &StrMutex{strMutexes: make(map[string]struct {
		mtx *sync.Mutex
		wg  *sync.WaitGroup
	})}
}

func (m *StrMutex) Lock(str string) {
	m.mapMutex.Lock()
	strMutex, ok := m.strMutexes[str]
	if !ok {
		strMutex.mtx = new(sync.Mutex)
		strMutex.wg = new(sync.WaitGroup)
		m.strMutexes[str] = strMutex
	}
	m.mapMutex.Unlock()

	strMutex.wg.Add(1)

	// if this is the first Lock ever called, create a delete waiter
	if !ok {
		go func() {
			strMutex.wg.Wait()

			m.mapMutex.Lock()
			delete(m.strMutexes, str)
			m.mapMutex.Unlock()
		}()
	}

	strMutex.mtx.Lock()
}

func (m *StrMutex) Unlock(str string) {
	m.mapMutex.Lock()
	strMutex, ok := m.strMutexes[str]
	m.mapMutex.Unlock()

	if !ok {
		panic(fmt.Sprintf("Unlock called for non locked string: %q", str))
	}

	strMutex.mtx.Unlock()
	strMutex.wg.Done()
}
