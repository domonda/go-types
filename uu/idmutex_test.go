package uu

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_IDMutex(t *testing.T) {
	uuid := IDMust("946521d6-9aef-4bb1-8a19-e0fc0c7e3665")

	idMutex := NewIDMutex()
	assert.Panics(t, func() { idMutex.Unlock(uuid) }, "not locked string should panic")

	numParallel := 100
	numAccess := 1000
	wg := sync.WaitGroup{}
	wg.Add(numParallel)

	testFunc := func() {
		for range numAccess {
			idMutex.Lock(uuid)
			time.Sleep(time.Nanosecond * time.Duration(rand.Intn(100)))
			idMutex.Unlock(uuid)
			time.Sleep(1 * time.Nanosecond) // Minimal sleep
		}
		wg.Done()
	}

	for range numParallel {
		go testFunc()
	}
	wg.Wait()
}
