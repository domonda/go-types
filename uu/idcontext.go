package uu

import (
	"context"
	"fmt"
	"sync"
)

var newIDFuncCtxKey int

// NewID returns a UUID generated with IDvDefault
// if the context does not have a function added
// to it with ContextWithIDFunc,
// else the context function is used.
//
// The context function can be used
// to override random generation with a
// deterministic series for testing.
//
// See also [IDv7Deterministic] and [IDv7DeterministicFunc].
func NewID(ctx context.Context) ID {
	return IDFromContext(ctx, &newIDFuncCtxKey)
}

// ContextWithIDFunc adds an ID generating function to the context
// either for the passed keys, or for the default key
// used by [NewID] if no keys are passed.
func ContextWithIDFunc(ctx context.Context, f func() ID, keys ...any) context.Context {
	if len(keys) == 0 {
		return context.WithValue(ctx, &newIDFuncCtxKey, f)
	}
	for _, key := range keys {
		ctx = context.WithValue(ctx, key, f)
	}
	return ctx
}

// ContextWithID adds an ID generating function to the context
// that will return the passed ID for the given keys
// or for the default key used by [NewID] if no keys are passed.
//
// The function is a shortcut for ContextWithIDFunc(ctx, func() ID { return id }, keys...)
func ContextWithID(ctx context.Context, id ID, keys ...any) context.Context {
	return ContextWithIDFunc(ctx, func() ID { return id }, keys...)
}

// ContextWithIDSequence adds an ID generating function to the context
// that will return the passed IDs in the given order.
//
// This function is intended to be used in tests
// to override random generation with a
// deterministic series using IDFromContext
// to retrieve the IDs with the given key.
//
// The ID generating function is safe for concurrent use
// and panics when called after all passed IDs have been returned,
// because using more IDs than defined is a bug in the test
// and the invalid IDNil must never be returned instead.
func ContextWithIDSequence(ctx context.Context, key any, ids ...ID) context.Context {
	var (
		mutex sync.Mutex
		index int
	)
	return ContextWithIDFunc(
		ctx,
		func() ID {
			mutex.Lock()
			defer mutex.Unlock()

			if index >= len(ids) {
				panic(fmt.Sprintf("uu.ContextWithIDSequence: all %d IDs of the sequence have already been used", len(ids)))
			}
			id := ids[index]
			index++
			return id
		},
		key,
	)
}

// IDFuncFromContext returns the ID generating function from the context
// added with ContextWithIDFunc without keys for the default key used by
// [NewID], or nil if there is none.
// Functions added for other keys are not returned.
func IDFuncFromContext(ctx context.Context) func() ID {
	if f, ok := ctx.Value(&newIDFuncCtxKey).(func() ID); ok {
		return f
	}
	return nil
}

// IDFromContext returns the ID that was added to the context
// with the passed key, or the result of IDvDefault() if no ID was added.
//
// This function enables writing tests with predefined IDs
// that are passed through from the test via the context,
// and falling back to the default ID generation
// for non-testing environments.
func IDFromContext(ctx context.Context, key any) ID {
	if f, ok := ctx.Value(key).(func() ID); ok {
		return f()
	}
	return IDvDefault()
}
