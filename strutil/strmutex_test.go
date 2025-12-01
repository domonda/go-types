package strutil

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_StrMutex(t *testing.T) {
	strMutex := NewStrMutex()
	assert.Panics(t, func() { strMutex.Unlock("test") }, "not locked string should panic")

	numParallel := 100
	numAccess := 1000
	wg := sync.WaitGroup{}
	wg.Add(numParallel)

	testFunc := func() {
		for range numAccess {
			strMutex.Lock("test")
			time.Sleep(time.Nanosecond * time.Duration(rand.Intn(100)))
			strMutex.Unlock("test")
			time.Sleep(1 * time.Nanosecond) // Minimal sleep
		}
		wg.Done()
	}

	for range numParallel {
		go testFunc()
	}
	wg.Wait()
}
