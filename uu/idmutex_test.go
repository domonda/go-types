package uu

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// waitOrFail waits for ch to be closed and fails the test instead of
// hanging forever if that does not happen. The timeout is only a
// safety net against deadlocks, not a timing assertion.
func waitOrFail(t *testing.T, ch <-chan struct{}, msg string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		t.Fatal(msg)
	}
}

func TestIDMutex_MutualExclusion(t *testing.T) {
	// The counter is only protected by the IDMutex.
	// If two goroutines could hold the lock for the same ID at once,
	// the race detector reports the unsynchronized access
	// and increments get lost.
	id := IDMust("946521d6-9aef-4bb1-8a19-e0fc0c7e3665")
	m := NewIDMutex()

	const goroutines, iterations = 20, 200
	counter := 0
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range iterations {
				m.Lock(id)
				counter++
				m.Unlock(id)
			}
		})
	}
	wg.Wait()

	assert.Equal(t, goroutines*iterations, counter, "no lost increments")
	assert.False(t, m.IsLocked(id), "unlocked after all goroutines are done")
	assert.Empty(t, m.locks, "no per-ID mutex is leaked after the last Unlock")
}

func TestIDMutex_LockBlocksUntilUnlock(t *testing.T) {
	id := IDMust("946521d6-9aef-4bb1-8a19-e0fc0c7e3665")
	m := NewIDMutex()

	m.Lock(id)

	acquired := make(chan struct{})
	released := make(chan struct{})
	go func() {
		m.Lock(id)
		close(acquired)
		m.Unlock(id)
		close(released)
	}()

	// Wait until the goroutine is registered as waiter for the ID
	require.Eventually(t, func() bool {
		m.global.Lock()
		defer m.global.Unlock()
		return m.locks[id] != nil && m.locks[id].count == 2
	}, 10*time.Second, time.Millisecond)

	// The first lock is still held, so the waiter can't have acquired it
	select {
	case <-acquired:
		t.Fatal("second Lock acquired the ID while the first lock was still held")
	default:
	}
	assert.True(t, m.IsLocked(id))

	m.Unlock(id)
	waitOrFail(t, acquired, "waiter did not acquire the lock after Unlock")
	waitOrFail(t, released, "waiter did not release the lock")
	assert.False(t, m.IsLocked(id))
}

func TestIDMutex_DifferentIDsDoNotBlock(t *testing.T) {
	// The point of IDMutex over a single mutex is that
	// unrelated IDs can be locked concurrently.
	a := IDMust("946521d6-9aef-4bb1-8a19-e0fc0c7e3665")
	b := IDMust("de8482f2-164a-40f9-b4e0-c349b78c22f6")
	m := NewIDMutex()

	m.Lock(a)
	defer m.Unlock(a)

	done := make(chan struct{})
	go func() {
		m.Lock(b)
		m.Unlock(b)
		close(done)
	}()
	waitOrFail(t, done, "locking another ID blocked on a held lock")

	assert.True(t, m.IsLocked(a))
	assert.False(t, m.IsLocked(b))
}

func TestIDMutex_IsLocked(t *testing.T) {
	id := IDMust("946521d6-9aef-4bb1-8a19-e0fc0c7e3665")
	m := NewIDMutex()

	assert.False(t, m.IsLocked(id))
	m.Lock(id)
	assert.True(t, m.IsLocked(id))
	assert.False(t, m.IsLocked(IDNil), "other IDs are not affected")
	m.Unlock(id)
	assert.False(t, m.IsLocked(id))

	// Unlocking more often than locking is a programming error
	assert.Panics(t, func() { m.Unlock(id) })
}

func TestNewIDMutexWithCallbacks(t *testing.T) {
	a := IDMust("946521d6-9aef-4bb1-8a19-e0fc0c7e3665")
	b := IDMust("de8482f2-164a-40f9-b4e0-c349b78c22f6")

	var (
		m      *IDMutex
		events []string
	)
	// The callbacks are documented to be called before any locking
	// or unlocking, so the lock state they observe is the one before
	// the operation.
	m = NewIDMutexWithCallbacks(
		func(id ID) {
			events = append(events, fmt.Sprintf("lock %s locked=%t", id, m.IsLocked(id)))
		},
		func(id ID) {
			events = append(events, fmt.Sprintf("unlock %s locked=%t", id, m.IsLocked(id)))
		},
	)

	m.Lock(a)
	m.Lock(b)
	m.Unlock(a)
	m.Unlock(b)

	assert.Equal(t, []string{
		"lock " + a.String() + " locked=false",
		"lock " + b.String() + " locked=false",
		"unlock " + a.String() + " locked=true",
		"unlock " + b.String() + " locked=true",
	}, events)

	// Nil callbacks are allowed
	m = NewIDMutexWithCallbacks(nil, nil)
	m.Lock(a)
	assert.True(t, m.IsLocked(a))
	m.Unlock(a)
	assert.False(t, m.IsLocked(a))
}
