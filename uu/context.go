package uu

import "context"

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
// See also IDv7Deterministic and IDv7DeterministicFunc.
func NewID(ctx context.Context) ID {
	if f := IDFuncFromContext(ctx); f != nil {
		return f()
	}
	return IDvDefault()
}

// ContextWithIDFunc adds an ID generating function to the context
// that will be used by NewID.
func ContextWithIDFunc(ctx context.Context, f func() ID) context.Context {
	return context.WithValue(ctx, &newIDFuncCtxKey, f)
}

// IDFuncFromContext returns the ID generating function from the context
// added with ContextWithIDFunc.
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
	if id, ok := ctx.Value(key).(ID); ok {
		return id
	}
	return IDvDefault()
}
