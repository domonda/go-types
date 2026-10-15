package strutil

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStrMutex_ZeroValue checks that the zero value is usable without
// NewStrMutex, like a sync.Mutex, so a StrMutex can be declared as a
// struct field or variable. Unlocking a not locked key must still panic
// with the documented message instead of a nil map error.
// The first Lock calls can race to initialize the zero value,
// run with -race to verify that this is synchronized.
func TestStrMutex_ZeroValue(t *testing.T) {
	var s struct{ m StrMutex }
	require.NotPanics(t, func() { s.m.Lock("a") })
	s.m.Unlock("a")
	// The key is free again, a second Lock must not block
	s.m.Lock("a")
	s.m.Unlock("a")

	var m StrMutex
	assert.PanicsWithValue(t, `Unlock called for non locked key: "a"`, func() { m.Unlock("a") })

	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			key := strconv.Itoa(i)
			m.Lock(key)
			m.Unlock(key)
		})
	}
	wg.Wait()
}

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
