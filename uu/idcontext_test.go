package uu

import (
	"context"
	"testing"

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
	// IDNil is returned when the sequence is exhausted
	require.Equal(t, IDNil, IDFromContext(ctx, &testIDFromContextCtxKeys.c))
	require.Equal(t, IDNil, IDFromContext(ctx, &testIDFromContextCtxKeys.c))
}
