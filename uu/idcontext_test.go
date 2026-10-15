package uu

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Using a struct as namespace for the context keys
// used by TestIDFromContext.
// Each struct field has a unique address in memory
// and is thus usable as a unique context key.
// Using byte for minimal memory footprint,
// struct{} does not create unique addresses.
// Note that because of struct padding and alignment,
// int could be used as well.
var testIDFromContextCtxKeys struct {
	a byte
	b byte
	c byte
}

func TestIDFromContext(t *testing.T) {
	// Fixed IDs for test
	a := IDFrom("e6eca3db-1f94-41ad-8a5c-80960f22b0de")
	b := IDFrom("de8482f2-164a-40f9-b4e0-c349b78c22f6")
	c := []ID{
		IDFrom("825a4e93-af00-48b4-8c07-a726cb701a3f"),
		IDFrom("6fb0f2bc-cba7-4cdb-ac36-87404e7c94cc"),
		IDFrom("edb4b646-2989-4128-bcdc-daf48a56c7fa"),
	}

	ctx := context.Background()
	ctx = ContextWithID(ctx, a, &testIDFromContextCtxKeys.a)
	ctx = ContextWithID(ctx, b, &testIDFromContextCtxKeys.b)
	ctx = ContextWithIDSequence(ctx, &testIDFromContextCtxKeys.c, c...)

	// Defined single IDs
	require.Equal(t, a, IDFromContext(ctx, &testIDFromContextCtxKeys.a))
	require.Equal(t, b, IDFromContext(ctx, &testIDFromContextCtxKeys.b))

	// Sequence of IDs retrieved by sequential IDFromContext calls
	require.Equal(t, c[0], IDFromContext(ctx, &testIDFromContextCtxKeys.c))
	require.Equal(t, c[1], IDFromContext(ctx, &testIDFromContextCtxKeys.c))
	require.Equal(t, c[2], IDFromContext(ctx, &testIDFromContextCtxKeys.c))
	// An exhausted sequence panics instead of returning IDNil,
	// because IDFromContext must never yield the invalid IDNil
	// and a test using more IDs than it defined is wrong.
	require.Panics(t, func() { IDFromContext(ctx, &testIDFromContextCtxKeys.c) })
	require.Panics(t, func() { IDFromContext(ctx, &testIDFromContextCtxKeys.c) }, "stays exhausted")
}

func TestContextWithIDSequence_Concurrent(t *testing.T) {
	// The context is shared by all goroutines handling a request,
	// so concurrent IDFromContext calls must hand out every ID
	// of the sequence exactly once (go test -race detects
	// an unsynchronized sequence index).
	const goroutines, perGoroutine = 8, 50
	ids := make(IDSlice, goroutines*perGoroutine)
	for i := range ids {
		ids[i] = IDv7Deterministic(int64(i))
	}
	ctx := ContextWithIDSequence(context.Background(), &testIDFromContextCtxKeys.c, ids...)

	results := make(chan ID, len(ids))
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perGoroutine {
				results <- IDFromContext(ctx, &testIDFromContextCtxKeys.c)
			}
		})
	}
	wg.Wait()
	close(results)

	seen := make(IDSet)
	for id := range results {
		require.True(t, seen.Insert(id), "ID %s handed out twice", id)
	}
	require.True(t, seen.Equal(ids.AsSet()), "every ID of the sequence handed out")

	require.Panics(t, func() { IDFromContext(ctx, &testIDFromContextCtxKeys.c) }, "exhausted")
}

func TestNewID(t *testing.T) {
	t.Run("without context function generates default IDs", func(t *testing.T) {
		// Production code calls NewID with contexts that have no
		// ID function, which must fall back to random IDvDefault IDs.
		ctx := context.Background()
		a := NewID(ctx)
		b := NewID(ctx)
		require.NoError(t, a.Validate())
		require.Equal(t, IDvDefault().Version(), a.Version(), "uses IDvDefault")
		require.NotEqual(t, a, b, "IDs are unique")
	})

	t.Run("uses function added without keys", func(t *testing.T) {
		// Tests inject a deterministic series to get reproducible IDs
		ctx := ContextWithIDFunc(context.Background(), IDv7DeterministicFunc(1))
		require.Equal(t, IDv7Deterministic(1), NewID(ctx))
		require.Equal(t, IDv7Deterministic(2), NewID(ctx))
	})

	t.Run("uses fixed ID added without keys", func(t *testing.T) {
		id := IDMust("e6eca3db-1f94-41ad-8a5c-80960f22b0de")
		ctx := ContextWithID(context.Background(), id)
		require.Equal(t, id, NewID(ctx))
		require.Equal(t, id, NewID(ctx), "same ID on every call")
	})

	t.Run("is not affected by functions added for other keys", func(t *testing.T) {
		id := IDMust("e6eca3db-1f94-41ad-8a5c-80960f22b0de")
		ctx := ContextWithID(context.Background(), id, &testIDFromContextCtxKeys.a)
		require.NotEqual(t, id, NewID(ctx))
	})
}

func TestContextWithIDFunc_MultipleKeys(t *testing.T) {
	// All passed keys share the same function,
	// so a sequence is consumed across the keys.
	ctx := ContextWithIDFunc(
		context.Background(),
		IDv7DeterministicFunc(100),
		&testIDFromContextCtxKeys.a,
		&testIDFromContextCtxKeys.b,
	)
	require.Equal(t, IDv7Deterministic(100), IDFromContext(ctx, &testIDFromContextCtxKeys.a))
	require.Equal(t, IDv7Deterministic(101), IDFromContext(ctx, &testIDFromContextCtxKeys.b))

	// The default NewID key is only set when no keys are passed
	assert.Nil(t, IDFuncFromContext(ctx))
	require.NotEqual(t, IDv7Deterministic(102), NewID(ctx))
}

func TestIDFromContext_FallbackForUnknownKey(t *testing.T) {
	// A key without a function must fall back to a generated ID,
	// never to IDNil which would be an invalid primary key.
	id := IDFromContext(context.Background(), &testIDFromContextCtxKeys.c)
	require.NoError(t, id.Validate())

	// A value of another type under the key is ignored
	ctx := context.WithValue(context.Background(), &testIDFromContextCtxKeys.c, "not a func")
	id = IDFromContext(ctx, &testIDFromContextCtxKeys.c)
	require.NoError(t, id.Validate())
}

func TestIDFuncFromContext(t *testing.T) {
	require.Nil(t, IDFuncFromContext(context.Background()), "no function added")

	id := IDMust("e6eca3db-1f94-41ad-8a5c-80960f22b0de")
	f := IDFuncFromContext(ContextWithID(context.Background(), id))
	require.NotNil(t, f)
	require.Equal(t, id, f())

	// Functions added for explicit keys are not returned
	ctx := ContextWithID(context.Background(), id, &testIDFromContextCtxKeys.a)
	require.Nil(t, IDFuncFromContext(ctx))
}
