package types

import (
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// deadlockGuard is only a safety net so a broken KeyMutex fails the test
// with a clear message instead of hanging until the global test timeout.
// It is not a timing assertion: correct code finishes far below it.
const deadlockGuard = 10 * time.Second

func waitOrFail(t *testing.T, done <-chan struct{}, msg string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(deadlockGuard):
		t.Fatal(msg)
	}
}

// keyLockCount returns the number of goroutines holding or waiting
// for the lock of key, read under the global mutex.
func keyLockCount[T comparable](m *KeyMutex[T], key T) int {
	m.global.Lock()
	defer m.global.Unlock()
	if lock := m.locks[key]; lock != nil {
		return lock.count
	}
	return 0
}

// TestKeyMutex_IsLocked checks the lifecycle of a key:
// unlocked before Lock, locked while held, unlocked again after Unlock.
func TestKeyMutex_IsLocked(t *testing.T) {
	m := NewKeyMutex[string]()
	if m.IsLocked("a") {
		t.Error("IsLocked() of never locked key = true, want false")
	}
	m.Lock("a")
	if !m.IsLocked("a") {
		t.Error("IsLocked() of locked key = false, want true")
	}
	if m.IsLocked("b") {
		t.Error("IsLocked() of other key = true, want false")
	}
	m.Unlock("a")
	if m.IsLocked("a") {
		t.Error("IsLocked() after Unlock = true, want false")
	}
}

// TestKeyMutex_UnlockNotLockedPanics checks that unlocking a key that is not
// locked panics like sync.Mutex does, instead of silently corrupting state,
// and that the panic message names the key for debugging.
func TestKeyMutex_UnlockNotLockedPanics(t *testing.T) {
	assertPanics := func(t *testing.T, m *KeyMutex[string], key string) {
		t.Helper()
		defer func() {
			r := recover()
			if r == nil {
				t.Fatalf("Unlock(%q) of not locked key did not panic", key)
			}
			if msg, _ := r.(string); !strings.Contains(msg, `"`+key+`"`) {
				t.Errorf("panic message %q does not name the key %q", r, key)
			}
		}()
		m.Unlock(key)
	}

	t.Run("never locked", func(t *testing.T) {
		assertPanics(t, NewKeyMutex[string](), "never-locked")
	})
	t.Run("double unlock", func(t *testing.T) {
		m := NewKeyMutex[string]()
		m.Lock("key")
		m.Unlock("key")
		assertPanics(t, m, "key")
	})
}

// TestKeyMutex_SameKeyBlocks checks that a second Lock of the same key
// waits until the first holder unlocks, and that the key stays locked
// while handed over to the waiter.
func TestKeyMutex_SameKeyBlocks(t *testing.T) {
	m := NewKeyMutex[string]()
	m.Lock("a")

	acquired := make(chan struct{})
	released := make(chan struct{})
	go func() {
		m.Lock("a")
		close(acquired)
		<-released
		m.Unlock("a")
	}()

	// Wait until the goroutine registered itself as waiter for "a".
	for keyLockCount(m, "a") != 2 {
		runtime.Gosched()
	}
	select {
	case <-acquired:
		t.Fatal("second Lock of the same key acquired it while the first holder still held it")
	default:
	}

	m.Unlock("a")
	waitOrFail(t, acquired, "waiting Lock did not acquire the key after Unlock")
	if !m.IsLocked("a") {
		t.Error("IsLocked() = false while the second holder holds the key, want true")
	}

	close(released)
	for m.IsLocked("a") {
		runtime.Gosched()
	}
}

// TestKeyMutex_DifferentKeysDoNotBlock checks the point of a per-key mutex:
// holding the lock of one key must not block locking another key.
func TestKeyMutex_DifferentKeysDoNotBlock(t *testing.T) {
	m := NewKeyMutex[string]()
	m.Lock("a")
	defer m.Unlock("a")

	done := make(chan struct{})
	go func() {
		m.Lock("b")
		m.Unlock("b")
		close(done)
	}()
	waitOrFail(t, done, "Lock of key b blocked while only key a was locked")
}

// TestKeyMutex_MutualExclusion checks that the same key is never held by two
// goroutines at once under contention. The counters are plain ints guarded
// only by the KeyMutex, so with -race any broken exclusion is also reported
// as a data race. After all goroutines are done no lock entries may be left
// behind, otherwise the map would grow with every key ever locked.
func TestKeyMutex_MutualExclusion(t *testing.T) {
	const (
		numKeys       = 3
		numGoroutines = 8
		numIterations = 300
	)
	m := NewKeyMutex[int]()
	var (
		counters [numKeys]int
		holders  [numKeys]atomic.Int32
		wg       sync.WaitGroup
	)
	for g := range numGoroutines {
		wg.Go(func() {
			for i := range numIterations {
				key := (g + i) % numKeys
				m.Lock(key)
				if n := holders[key].Add(1); n != 1 {
					t.Errorf("key %d held by %d goroutines at once", key, n)
				}
				counters[key]++
				holders[key].Add(-1)
				m.Unlock(key)
			}
		})
	}
	wg.Wait()

	total := 0
	for _, c := range counters {
		total += c
	}
	if want := numGoroutines * numIterations; total != want {
		t.Errorf("sum of counters = %d, want %d (lost updates)", total, want)
	}
	m.global.Lock()
	leftover := len(m.locks)
	m.global.Unlock()
	if leftover != 0 {
		t.Errorf("%d lock entries left after all keys were unlocked, want 0", leftover)
	}
}
