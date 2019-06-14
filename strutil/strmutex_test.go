package strutil

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func Test_StrMutex(t *testing.T) {
	numParallel := 100
	numAccess := 1000
	str := "test"
	strMutext := NewStrMutex()

	var wg sync.WaitGroup
	wg.Add(numParallel)

	testFunc := func() {
		for i := 0; i < numAccess; i++ {
			strMutext.Lock(str)
			time.Sleep(time.Nanosecond * time.Duration(rand.Intn(100)))
			strMutext.Unlock(str)
		}
		wg.Done()
	}

	for i := 0; i < numParallel; i++ {
		go testFunc()
	}
	wg.Wait()
}
